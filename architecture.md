# Architecture — ShikaGari Auction Engine

**Status:** Draft for review · **Last updated:** 2026-10-13
**Companion docs:** `prd.md` (what and why), `design.md` (what it looks like)

This document is the implementation blueprint requested in `prompt.md` §4: system architecture, database schema, real-time state machine, API contracts and SSO flow. It is annotated with the **phase** in which each component is built.

---

## 1. Guiding constraint

> **A bid is a financial transaction. It must be accepted exactly once, or not at all.**

Everything below follows from that. We optimise for *provable correctness first, latency second, throughput third*. Throughput is easy to add; a double-awarded lot cannot be undone.

---

## 2. Topology

```
                    ┌──────────────────────────────────────────┐
                    │  shikagari-web  (Next.js, existing)      │
                    │  /auctions/*  ·  navbar "Auctions" button │
                    └───────────┬──────────────────┬───────────┘
                                │ SSO token       │ REST/WS
                                │ (short-lived)   │
                    ┌───────────▼──────────────────▼───────────┐
                    │  shikagari-auction  (Go, new service)     │
                    │  ┌────────────┐  ┌──────────┐  ┌────────┐ │
                    │  │ REST API   │  │ WS hub   │  │ Auction│ │
                    │  │ (control)  │  │ (fan-out)│  │ runner │ │
                    │  └─────┬──────┘  └────┬─────┘  └───┬────┘ │
                    │        │              │            │      │
                    │  ┌─────▼──────────────▼────────────▼────┐ │
                    │  │ Redis — locks, streams, pub/sub, cache │ │
                    │  └──────────────┬────────────────────────┘ │
                    └─────────────────┼──────────────────────────┘
                                      │ write
                    ┌─────────────────▼──────────────────────────┐
                    │  PostgreSQL — auction DB (own database)    │
                    │  lots · bids · ledger · invoices · audit    │
                    └──────────────────────────────────────────┘
                                      │ read identity/listing
                    ┌─────────────────▼──────────────────────────┐
                    │  shikagari/api  (existing Go marketplace)  │
                    │  users · dealers · listings · verification │
                    └──────────────────────────────────────────┘
```

### 2.1 Why a separate service

The hot path (a bid accepted in <150ms, fanned out to thousands) has different scaling, failure, and deployment characteristics from the marketplace API. Coupling a lane runner's release cadence to catalogue browsing is how auction platforms get flaky during their busiest hour.

**A separate Go service** (`shikagari-auction`) is also the path of least resistance: the team already writes Go with Gin, GORM, and Postgres. Redis and a WebSocket library are new dependencies, but nothing else changes.

### 2.2 Why the UI stays in the existing Next app

`prompt.md` suggests a subdomain. We keep the auction UI in `shikagari-web` under `/auctions/*`:

- The navbar "Auctions" button is already a client-side transition — no subdomain, no DNS/TLS/cookie work, no second deployment.
- Auth context, `cn()`, `formatKES()`, design tokens and components are reused.
- It is a routing choice, not an architectural one: the auction API can move to a subdomain later behind the same client, since all traffic is token-authenticated rather than cookie-scoped.

Trade-off we accept: a heavy auction room adds JS to the main bundle. Mitigation — route-level code splitting, which `shikagari-web` already does per route.

---

## 3. Technology choices

| Concern | Choice | Phase | Rationale |
| --- | --- | --- | --- |
| Language | Go 1.25 | 1 | Same as `shikagari/api`; goroutines suit many idle sockets |
| HTTP | Gin | 1 | Already used; keeps the mental model identical |
| Database | PostgreSQL 16 | 1 | ACID, `SELECT … FOR UPDATE`, money-safe integers |
| Hot coordination | Redis 7 | 2 | Streams give ordered, durable, replayable bid intake; Pub/Sub does fan-out |
| WebSocket | `nhooyr.io/websocket` (or `gorilla/websocket`) | 2 | Light, stdlib-adjacent, no framework lock-in |
| ORM | GORM for CRUD; **raw SQL for the bid transaction** | 1 | Hot path stays explicit and reviewable |
| Payments | M-Pesa Daraja STK push | 1 | The only rail a Kenyan buyer will use at auction speed |
| Cache | Redis | 4 | Snapshot payloads per lane; absorb spectator traffic |
| Observability | OpenTelemetry + Prometheus + structured logs | 6 | Required before opening to the public |

**No Kafka.** Redis Streams plus a Postgres outbox cover our throughput. Kafka is a Phase 6 revisit if we add regions.

---

## 4. Data model

All money columns are `BIGINT` KES. All timestamps are `TIMESTAMPTZ` in UTC.

### 4.1 Core tables (Phase 1)

```sql
-- Platform-level settings, single row. license_enabled gates public auctions
-- until ShikaGari's auctioneer's licence is granted; the number is published
-- once it is.
CREATE TABLE auction_settings (
  id                SMALLSERIAL PRIMARY KEY DEFAULT 1,
  license_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
  license_number    VARCHAR(64),
  licensee_name     VARCHAR(160) NOT NULL DEFAULT 'ShikaGari',
  max_concurrent_lanes INT,
  bid_rate_limit_per_minute INT NOT NULL DEFAULT 30,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT settings_singleton CHECK (id = 1),
  CONSTRAINT licence_number_required_when_enabled
    CHECK (NOT license_enabled OR license_number IS NOT NULL)
);

CREATE TABLE auction_sessions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name            VARCHAR(160) NOT NULL,
  starts_at       TIMESTAMPTZ NOT NULL,
  ends_at         TIMESTAMPTZ NOT NULL,
  status          VARCHAR(16) NOT NULL DEFAULT 'scheduled'
                  CHECK (status IN ('scheduled','live','closed','cancelled')),
  -- Accountability: a named auctioneer operator must be recorded per session
  auctioneer_user_id UUID,
  settings_snapshot  JSONB NOT NULL,   -- tunables frozen at open (see §5.4)
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_start ON auction_sessions (starts_at DESC);

CREATE TABLE auction_lanes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         VARCHAR(16) NOT NULL UNIQUE,          -- "A", "B"
  name         VARCHAR(64) NOT NULL,
  sort_order   INT NOT NULL DEFAULT 0,
  is_active    BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auction_lots (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  -- Bridge to the marketplace
  listing_id        UUID NOT NULL,                    -- logical FK to shikagari/api listings
  dealer_user_id    UUID NOT NULL,                    -- logical FK to shikagari/api users

  -- Frozen snapshot of the vehicle at schedule time, so the catalogue can be
  -- deleted from the marketplace without destroying auction history.
  snapshot          JSONB NOT NULL,

  sale_type         VARCHAR(20) NOT NULL
                    CHECK (sale_type IN ('pure_sale','reserve','on_approval')),
  reserve_price_kes BIGINT,                            -- NULL for pure_sale
  opening_bid_kes   BIGINT NOT NULL,
  increment_kes     BIGINT NOT NULL CHECK (increment_kes > 0),

  status            VARCHAR(24) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','scheduled','open','on_block',
                                     'sold','pending_approval','unsold','withdrawn')),

  scheduled_start   TIMESTAMPTZ NOT NULL,
  lane_id           UUID REFERENCES auction_lanes(id),
  session_id        UUID REFERENCES auction_sessions(id),

  -- Live state, owned exclusively by the lane runner for this lot
  current_price_kes BIGINT NOT NULL DEFAULT 0,
  leading_bid_id    UUID,
  leader_user_id    UUID,
  bid_count         INT  NOT NULL DEFAULT 0,
  version           BIGINT NOT NULL DEFAULT 0,        -- monotonic; client resync token

  closed_at         TIMESTAMPTZ,
  withdrawal_reason TEXT,

  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lots_status_start ON auction_lots (status, scheduled_start);
CREATE INDEX idx_lots_lane_status   ON auction_lots (lane_id, status);
CREATE INDEX idx_lots_leader        ON auction_lots (leader_user_id) WHERE status = 'on_block';
```

**Bid table is append-only.** Never updated, never deleted:

```sql
CREATE TABLE auction_bids (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  lot_id              UUID NOT NULL REFERENCES auction_lots(id),
  bidder_user_id      UUID NOT NULL,
  source              VARCHAR(12) NOT NULL
                      CHECK (source IN ('live','proxy','admin')),
  amount_kes          BIGINT NOT NULL CHECK (amount_kes > 0),
  max_amount_kes      BIGINT,                        -- pre-bid ceiling
  is_pre_bid          BOOLEAN NOT NULL DEFAULT FALSE,
  idempotency_key     UUID,                          -- client-supplied
  ip_address          INET,
  user_agent          TEXT,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT bids_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_bids_lot_time   ON auction_bids (lot_id, created_at DESC);
CREATE INDEX idx_bids_lot_amount  ON auction_bids (lot_id, amount_kes DESC);
CREATE INDEX idx_bids_bidder_time ON auction_bids (bidder_user_id, created_at DESC);
```

### 4.2 Money ledger (Phase 1) — the most important table in the system

Double-entry. Every movement is a pair of rows that always sum to zero. There is no mutable "balance" column anywhere; balances are a query over the ledger.

```sql
CREATE TABLE ledger_accounts (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID REFERENCES users(id),
  kind        VARCHAR(24) NOT NULL
              CHECK (kind IN ('user_deposit','escrow_pool','buyer_premium',
                              'seller_payout','platform_fee','refundable')),
  currency    CHAR(3) NOT NULL DEFAULT 'KES',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, kind)
);

CREATE TABLE ledger_entries (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  txn_id        UUID NOT NULL,
  account_id    UUID NOT NULL REFERENCES ledger_accounts(id),
  direction     VARCHAR(4) NOT NULL CHECK (direction IN ('debit','credit')),
  amount_kes    BIGINT NOT NULL CHECK (amount_kes > 0),
  kind          VARCHAR(32) NOT NULL              -- deposit_funded, hold_placed,
                                                  -- hold_released, hold_applied,
                                                  -- payment_received, refund_sent
  reference_type VARCHAR(24),                      -- lot | bid | invoice
  reference_id   UUID,
  memo          TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_account_time ON ledger_entries (account_id, created_at DESC);
CREATE INDEX idx_ledger_txn           ON ledger_entries (txn_id);
CREATE INDEX idx_ledger_ref           ON ledger_entries (reference_type, reference_id);
```

**Invariants (enforced by tests, asserted nightly):**

1. Every `txn_id` sums to zero across debit and credit.
2. `available_buying_power(user) = deposits_held − active_holds`, where `active_holds` = holds on bids where that user is not the current leader and the lot is not closed.
3. No negative available power for any user, ever. Checked before any bid is accepted.
4. Deposits are never deleted; only moved.

### 4.3 Holds (Phase 1)

```sql
CREATE TABLE auction_bid_holds (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL,
  lot_id      UUID NOT NULL REFERENCES auction_lots(id),
  bid_id      UUID REFERENCES auction_bids(id),
  amount_kes  BIGINT NOT NULL CHECK (amount_kes > 0),
  status      VARCHAR(16) NOT NULL DEFAULT 'active'
              CHECK (status IN ('active','released','applied','forfeited')),
  released_at TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_holds_user_status ON auction_bid_holds (user_id, status);
CREATE INDEX idx_holds_lot_status  ON auction_bid_holds (lot_id, status);
```

### 4.4 Settlement (Phase 5)

```sql
CREATE TABLE auction_invoices (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  lot_id            UUID NOT NULL REFERENCES auction_lots(id),
  buyer_user_id     UUID NOT NULL,
  hammer_price_kes  BIGINT NOT NULL,
  buyer_premium_kes BIGINT NOT NULL,
  platform_fee_kes  BIGINT NOT NULL,
  total_kes         BIGINT NOT NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'issued'
                    CHECK (status IN ('issued','part_paid','paid','defaulted',
                                     'refunded','disputed')),
  payment_due_at    TIMESTAMPTZ NOT NULL,
  paid_at           TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auction_storage_charges (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  lot_id       UUID NOT NULL REFERENCES auction_lots(id),
  charge_date  DATE NOT NULL,
  amount_kes   BIGINT NOT NULL,
  waived       BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE (lot_id, charge_date)
);

CREATE TABLE auction_audit_log (
  id          BIGSERIAL PRIMARY KEY,
  actor_id    UUID,
  actor_role  VARCHAR(16),
  entity_type VARCHAR(24) NOT NULL,
  entity_id   UUID NOT NULL,
  action      VARCHAR(48) NOT NULL,
  before      JSONB,
  after       JSONB,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_entity ON auction_audit_log (entity_type, entity_id, created_at DESC);
-- Phase 1 (LIC-4): retention period set by counsel. Never hard-delete rows;
-- archive to cold storage instead so the sales register survives indefinitely.
-- CREATE INDEX idx_audit_time ON auction_audit_log (created_at);
```

### 4.5 Cross-database references

`listing_id` and `dealer_user_id` are **logical, not physical, foreign keys**. The auction database must not block on the marketplace DB, and a deleted listing must never cascade-delete auction history. Integrity is enforced by the bridge sync in §7.2.

---

## 5. Concurrency model — how we make a bid atomic

This is the core of the design. **Phase 2.**

### 5.1 The problem

Ten thousand people, 60 seconds, and bids arriving in the same millisecond. Two things must never happen: two bids accepted as leaders on one lot, or an accepted bid that vanishes.

A naive `UPDATE auction_lots SET current_price_kes = $1 WHERE id = $2` under high concurrency will lose updates. A naive Redis `GET/SET` will accept bids that fail the money check.

### 5.2 The solution: single-writer lane runner over an ordered log

```
   HTTP POST /lots/:id/bids
              │
              ▼
   ┌──────────────────────┐   XADD  (atomic, ordered, durable)
   │  Bid intake (any     │──────────────►  stream  auction:lane:{lane_id}:bids
   │  engine replica)     │
   └──────────────────────┘
              │
              ▼
   ┌──────────────────────────────────────────────┐
   │  Lane runner  (EXACTLY ONE active per lane,  │
   │  leader-elected in Redis)                     │
   │                                               │
   │  XREADGROUP → validate → Postgres txn →       │
   │  publish → XACK                                │
   └───────────────┬──────────────────────────────┘
                   │ PUBLISH auction:lane:{lane_id}:events
                   ▼
            WS hub → every connected client
```

**Why this is correct:**

- **Ingest is atomic and ordered.** `XADD` to a single Redis stream gives a total order per lane with no lost writes. Two simultaneous bids become two ordered entries.
- **Validation is single-threaded per lot.** One goroutine decides the winner. No compare-and-set race is possible because there is no concurrency inside a lot.
- **The write path is one Postgres transaction:**
  ```sql
  BEGIN;
    SELECT * FROM auction_lots WHERE id = $1 FOR UPDATE;   -- defence in depth
    -- (the runner already guarantees exclusivity; this guards against
    --  an admin edit racing the runner)
    -- 1. validate: lot open, amount >= next_increment, not self-outbidding
    -- 2. validate: available buying power >= amount
    -- 3. INSERT INTO auction_bids (idempotency_key unique)
    -- 4. UPDATE auction_lots SET current_price_kes, leader_user_id,
    --        bid_count = bid_count + 1, version = version + 1
    -- 5. ledger: release loser's hold, place winner's hold
  COMMIT;
  ```
- **Idempotency.** A client retry with the same `idempotency_key` hits the unique constraint and returns the original result instead of double-charging.
- **The stream is the replay log.** After a crash, the new leader resumes the consumer group from the last acked entry. In-flight ambiguity is bounded and recoverable — this satisfies NFR-6.
- **Locking is advisory, not load-bearing.** We use `SET NX PX` per lot for admin-vs-runner edits and for cross-lane balance checks. The invariant does not depend on Redis being up.

### 5.3 Buying power check under concurrency

The check is only meaningful inside the runner, where lots are serialised:

```
available = SUM(deposits_held) − SUM(active holds for this user)
if (bid_amount > available) → reject with `insufficient_buying_power`
```

A buyer bidding in two lanes is serialised only by luck — the runner for lane A and lane B are different goroutines. Therefore:

- **Per-user mutex:** the runner acquires a short Redis lock `auction:user:{user_id}:lock` (50ms TTL, spin-on-contention with 5ms backoff) before the balance check and hold placement, releasing after commit.
- This is the one place Redis correctness *is* load-bearing, because it spans lanes. Failure mode is explicit: if Redis is unavailable, **bids are rejected** (fail closed), never accepted optimistically. Money outranks availability.

### 5.4 Tunable constants

Every number below is admin-configurable at runtime (Phase 1) and stored per session, not hard-coded. Defaults are chosen for a 60-second lot; the Copart reference uses 60–90s and the original spec suggested 30–60s — we start at 60s, which is long enough to react on mobile data without dragging.

| Constant | Default | Phase | Notes |
| --- | --- | --- | --- |
| `lot_duration_seconds` | 60 | 2 | Time on block |
| `soft_close_threshold_seconds` | 10 | 2 | Inside this window a bid extends the clock |
| `soft_close_extension_seconds` | 15 | 2 | Clock resets to this |
| `bid_increment_schedule` | tiered by price | 1 | Smaller increments low, larger high; admin-editable per session |
| `pre_bid_cutoff_seconds` | 60 | 1 | Before scheduled start |
| `one_active_lot_per_buyer` | true | 1 | Relaxed for dealers in Phase 3 |
| `payment_window_hours` | 72 | 5 | From hammer |
| `storage_free_days` | 3 | 5 | Then daily charge accrues |
| `approval_decision_hours` | 24 | 5 | Seller accept / counter / decline |
| `bid_rate_limit_per_minute` | 30 | 3 | Anti-bot guard |

### 5.5 Timer and soft close (Phase 2)

The runner owns the clock for its lot:

```
on_lot_open:     timer = 60s
on_accepted_bid: if timer_remaining <= 10s → timer = 15s   # soft close
on_timer_zero:   hammer
```

Server-authoritative: the client never decides the outcome. Clients receive `deadline_at` (absolute UTC) and render a locally-interpolated countdown, correcting drift on each tick. A client that thinks it has 3 seconds when the server says 8 will be told by the server when it is too late — and its bid is rejected with `lot_closed`.

---

## 6. Real-time protocol

### 6.1 Connection and subscription

`GET /ws?token=<jwt>` — token in query string (WebSocket cannot set headers), validated on upgrade, never logged. Re-auth on reconnect.

Client sends `SUBSCRIBE` frames:

```json
{"type":"SUBSCRIBE","channels":["auction:all","lane:A","lot:<uuid>","me"]}
```

- `auction:all` — lot schedule changes, hammer results
- `lane:A` — the live floor for one lane
- `lot:<uuid>` — fine-grained view for the lot page
- `me` — private: your holds, your outbid notices, your buying power

Private channel data is **never** written to shared pub/sub; the WS hub resolves `me` from the connection's authenticated user id.

### 6.2 Event catalogue

| Event | Direction | Payload |
| --- | --- | --- |
| `SNAPSHOT` | S→C | Full authoritative state on subscribe/reconnect: lots, current price, leader display, `version` |
| `LOT_STARTED` | S→C | `{lot_id, lane_id, opening_bid_kes, increment_kes, deadline_at, version}` |
| `BID_ACCEPTED` | S→C | `{lot_id, amount_kes, bidder_display, is_you, version, deadline_at}` |
| `BID_REJECTED` | S→C | `{client_bid_id, code, message}` — sent to the bidder only |
| `TIMER_TICK` | S→C | `{lot_id, seconds_remaining, deadline_at}` — 1/s, only last 15s |
| `TIMER_EXTENDED` | S→C | `{lot_id, deadline_at, reason:"soft_close"}` |
| `YOU_WERE_OUTBID` | S→C | `{lot_id, new_amount_kes, your_max_kes, seconds_to_act}` |
| `LOT_HAMMERED` | S→C | `{lot_id, hammer_price_kes, winner_display, sale_type, reserve_met}` |
| `LOT_APPROVAL_REQUIRED` | S→C | `{lot_id, hammer_price_kes, counter_deadline}` |
| `POWER_UPDATE` | S→C | `{available_kes, held_kes}` — private |
| `HEARTBEAT` | S→C | every 20s; client replies `PONG` |

`version` on every lot-scoped event is the client-side ordering key. A client receiving version 42 after 43 discards 42.

### 6.3 Reconnection and reconciliation

1. Client loses connection. It retries with exponential backoff (1s, 2s, 4s… cap 10s) and a jitter.
2. On reconnect it sends `SUBSCRIBE` with its **last known `version`**.
3. Server replies `SNAPSHOT` with authoritative state, always — we do not attempt delta replay for the socket path. The snapshot is one Redis-cached payload per lane, so it is cheap.
4. Pending bids are the client's responsibility: each gets a client-generated `idempotency_key`, and on reconnect the client asks `/bids/{key}` for the outcome before deciding whether to retry. **Never blind-retry a bid without its key.**

This satisfies LIVE-6 and NFR-5 without building a delta protocol we do not need yet.

---

## 7. API contracts

### 7.1 SSO handoff (Phase 1)

Sequence for a logged-in user clicking **Auctions**:

```
Browser            shikagari-web          shikagari/api        auction-engine
   │                     │                      │                     │
   │ click "Auctions"    │                      │                     │
   ├────────────────────►│                      │                     │
   │                     │ POST /internal/auction/sso   (server-to-server,
   │                     ├─────────────────────►│        signed with
   │                     │                      │        SHARED_SSO_SECRET)
   │                     │◄─────────────────────┤                     │
   │                     │ { handoff_token, exp=120s, one_time=true }   │
   │                     │                      │                     │
   │◄────────────────────┤ /auctions/sso?ticket=…                     │
   │ GET /auctions/auth/handoff?ticket=…                             │
   ├───────────────────────────────────────────────────────────────►│
   │◄───────────────────────────────────────────────────────────────┤
   │ 302 → /auctions  + auction_session cookie (HttpOnly, Secure)   │
   │                     │                      │                     │
   │ GET /auctions (subsequent requests carry the cookie)            │
   ├───────────────────────────────────────────────────────────────►│
```

Properties:

- **No second registration.** Same user id, same account.
- **Ticket is one-time and 120 seconds** — a stolen ticket is near-useless, and single-use is recorded so replay is detectable.
- The main app mints nothing the engine trusts blindly: the engine calls back or validates the HMAC over `{user_id, role, exp, jti}` against the shared secret. **Phase 1: HMAC. Phase 4: switch to asymmetric signing so the engine can validate offline.**
- Role and dealer-approved status are carried in the ticket, but **the engine re-reads authoritative role from the marketplace API for any privileged action** (creating a lot). Never trust a long-lived ticket for authorisation.

Guests: clicking Auctions with no session still renders the public schedule and live lanes (read-only WS channels). Any bid attempt returns `401` with a `login_url`, and the UI prompts.

### 7.2 Listing bridge (Phase 1)

```
GET  /internal/listings/{listing_id}      → snapshot source (dealer-scoped)
POST /internal/sync/listings              → batch reconciliation nightly
```

When a dealer converts a listing, the engine pulls the listing, snapshots the fields it needs, and stores them in `auction_lots.snapshot`. After that the auction is self-contained.

### 7.3 REST surface

**Gate (Phase 1, LIC-1).** Every session-creating endpoint returns
`403 LICENSE_NOT_GRANTED` while `auction_settings.license_enabled` is false, except the explicitly internal `POST /v1/internal/sessions` used for staging auction days. Read-only endpoints (calendar, lot detail, live spectator channels) stay open.

**Public / buyer (Phases 1–5)**

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/v1/auctions` | Upcoming + live sessions, county/status filters |
| `GET` | `/v1/auctions/{id}` | Session with lanes and scheduled lots |
| `GET` | `/v1/lots/{id}` | Lot detail, including snapshot and live state |
| `GET` | `/v1/lots/{id}/bids` | Recent bid history (public amounts, anonymised bidders) |
| `POST` | `/v1/lots/{id}/pre-bids` | `{max_amount_kes, idempotency_key}` |
| `DELETE` | `/v1/lots/{id}/pre-bids/{id}` | Withdraw a pre-bid before open |
| `POST` | `/v1/lots/{id}/bids` | Live bid; also available over WS |
| `GET` | `/v1/me/buying-power` | Available / held / total |
| `POST` | `/v1/me/deposits` | M-Pesa STK push initiation |
| `GET` | `/v1/invoices/{id}` | Invoice and payment state |

**Dealer (Phases 1, 6)**

| Method | Path | Notes |
| --- | --- | --- |
| `POST` | `/v1/lots` | Convert a listing; reserve, sale type, schedule |
| `PATCH` | `/v1/lots/{id}` | Edit only while `draft` or `scheduled` |
| `GET` | `/v1/dealer/auctions` | My lots across sessions |

**Admin (Phases 1–6)**

| Method | Path | Notes |
| --- | --- | --- |
| `POST` | `/v1/admin/sessions` | Create an auction session; requires `auctioneer_user_id` |
| `PATCH` | `/v1/admin/settings/license` | Record the granted licence number and enable public auctions |
| `POST` | `/v1/admin/sessions/{id}/lanes` | Add/configure lanes |
| `POST` | `/v1/admin/lots/{id}/schedule` | Assign lot to lane + time |
| `POST` | `/v1/admin/lots/{id}/open` | Open for bidding (also triggers pre-bid resolution) |
| `POST` | `/v1/admin/lots/{id}/withdraw` | Mid-auction withdrawal with reason |
| `POST` | `/v1/admin/approvals/{lot_id}` | Accept / counter / decline on approval lots |
| `GET` | `/v1/admin/reconciliation` | Ledger vs M-Pesa statement |

### 7.4 Error contract

Every rejection returns a machine-readable code the UI maps to a message (see `design.md` §7):

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_BUYING_POWER",
    "message": "You can bid up to KES 3,200,000 with your current deposit.",
    "retryable": false,
    "client_bid_id": "…"
  }
}
```

| Code | Meaning | Retryable |
| --- | --- | --- |
| `LOT_CLOSED` | Timer expired; hammer already processed | no |
| `BID_TOO_LOW` | Below next increment | yes, with the new minimum |
| `SELF_OUTBID` | You are already leading | no |
| `INSUFFICIENT_BUYING_POWER` | Deposit does not cover it | yes, after topping up |
| `ALREADY_HOLDING_LOT` | v1 one-active-lot rule | no |
| `NOT_FUNDS_SETTLED` | M-Pesa deposit not yet confirmed | yes, later |
| `ACCOUNT_NOT_VERIFIED` | Identity check incomplete | no |
| `ENGINE_BUSY` | Redis coordination unavailable; fail-closed | yes, in 2s |

---

## 8. Leader election and failover (Phase 4)

- One engine instance per lane runner, elected via `SET auction:lane:{id}:runner <instance> NX PX 5000`, renewed every 2s.
- Loss of the lease → stop consuming, drain in-flight work, release per-user locks, exit.
- On takeover, the new runner resumes the consumer group from `XREADGROUP` with `>` — unacked entries are re-delivered. Because bids are idempotent, re-delivery is safe.
- Postgres `FOR UPDATE` on the lot row is the backstop for the brief window between two runners.

---

## 9. Observability (Phase 6, stubbed earlier)

| Metric | Type | Why |
| --- | --- | --- |
| `bid_accepted_total{result}` | counter | Accepted vs rejected, by reason |
| `bid_accept_duration_ms` | histogram | NFR-1 |
| `fanout_duration_ms` | histogram | NFR-2 |
| `active_lane_runners` | gauge | Should equal configured lanes |
| `stream_pending_entries` | gauge | Backlog; rising means the runner is stuck |
| `buyer_power_check_latency_ms` | histogram | Redis lock contention |
| `ledger_imbalance_kes` | gauge | Must always be 0 (alarm) |
| `ws_connections` / `ws_reconnects` | gauge / counter | Spectator health |
| `mpesa_callback_failures` | counter | Payment rail health |

**Traces** span: HTTP intake → XADD → runner consume → Postgres commit → publish → WS send. A slow bid must be attributable to one of those five hops.

---

## 10. Security

| Concern | Control |
| --- | --- |
| Bid forgery | Server-authoritative; client never sets price outcome |
| Replay | `idempotency_key` unique constraint; one-time SSO tickets |
| Race manipulation | Per-user lock across lanes; soft close prevents last-second snipe |
| Self-dealing | v1: dealers cannot bid on their own lots (enforced in runner) |
| Card/identity data | Engine never stores NIDA numbers; only a verification status flag from the marketplace |
| Money movement | Ledger only; no direct balance writes; every path tested for double-entry |
| M-Pesa callbacks | Signature verification, idempotent processing, replay window |
| Secrets | HMAC secret in env/secret manager; never logged. WS tokens never logged |
| Licence compliance | `auction_settings` gates public sessions; audit log is append-only and retained; licence number surfaced on lot pages and invoices |

---

## 11. Deployment (Phase 1 skeleton, hardened Phase 7)

- **Auction engine:** containerised Go binary, stateless except for the Redis lease. Horizontal replicas behind a load balancer with WS support.
- **Redis:** managed, single region, persistence enabled (AOF). Losing Redis stops bidding — fail-closed, by design.
- **Postgres:** separate database on the existing instance initially; managed HA by Phase 7.
- **Migrations:** own `migrations/` directory, numbered independently from the marketplace (which is at `018`). Never share a migration series across services.
- **Environments:** local (docker-compose), staging, production — each with its own M-Pesa sandbox credentials.

---

## 12. Build order mapped to phases

| Phase | Backend work | Frontend work |
| --- | --- | --- |
| 0 | — | — |
| 1 | Service skeleton, lot model, sessions, licence gate, SSO, M-Pesa deposits, ledger + holds, pre-bids, calendar API | `/auctions` shell, navbar button, calendar, lot page, deposit top-up, pre-bid form |
| 2 | WS hub, lane runner, timer + soft close, increment/custom bids, hammer, reconnect | Auction room UI, bid pad, outbid alerts, timer visualisation |
| 3 | Proxy bidding engine, reserve enforcement, dealer tier, throttles | "Your max bid" UX, buying-power meter, top-up prompts |
| 4 | Leader election, multi-lane, Redis snapshot cache, load tuning | Lane view switcher, spectator scale |
| 5 | Invoices, M-Pesa payment, defaults, storage accrual, counter-offers | Invoice page, payment countdown, default notices |
| 6 | Admin dashboard, reconciliation, dispute tooling, notifications, audit explorer | Admin console, dispute UI, reports |
| 7 | Load/chaos/DR drills, pen test fixes, alerting | Polish, empty/edge states, accessibility pass |

---

## 13. Decisions deliberately deferred

| Decision | Deferred to | Why not now |
| --- | --- | --- |
| Kafka instead of Redis Streams | Phase 6 | Unnecessary at our volume; Streams + Postgres outbox is sufficient and operable |
| Asymmetric SSO signing | Phase 4 | HMAC is fine for a single trust domain |
| Multi-region | Phase 7+ | Kenyan traffic is single-region |
| WebRTC/SFU yard video | Not planned | Bandwidth cost against M-Pesa-margin economics |
| Event sourcing for the ledger | Never | The ledger *is* append-only already; full event sourcing adds replay complexity for no gain |