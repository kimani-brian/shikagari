# PRD — ShikaGari Live Vehicle Auctions

**Status:** Draft for review · **Owner:** Product · **Last updated:** 2026-10-13
**Source material:** `auction-brief.md` (Copart mechanics), `prompt.md` (engine spec)
**Companion docs:** `architecture.md` (how it is built), `design.md` (how it looks and behaves)

---

## 1. What this is

A live, timed vehicle auction engine bolted onto the existing ShikaGari marketplace. Today the platform supports static listings and buyer inquiries only. This product adds **bidding**: dealers convert a listing into an auction lot, buyers deposit funds to gain buying power, and a live auction room streams the sale lot by lot in real time.

The reference model is Copart's VB3 — pre-bidding, then a fast live floor with a countdown clock that resets on every bid. We copy the mechanics, not the letter of the law: everything is denominated in **KES**, settled through **M-Pesa**, and the licensing rules, salvage-title restrictions and card rails in the brief do not apply here.

### 1.1 Why now

- Kenya's used-vehicle trade is fragmented: WhatsApp groups, newspaper listings, and informal brokers. Live auctions concentrate supply and create an auditable price.
- Our existing buyer-verification work (NTSA e-logbook check, admin review) is already the trust primitive auctions need — the winning bidder is an identified, verified person.
- We already own the inventory, the dealer relationships, and the inquiry funnel. Auctions monetise assets we have.

### 1.2 Product principles

1. **The clock is the product.** Everything else serves the 60 seconds a lot is on the block.
2. **Never let a race decide a sale.** A bid either wins cleanly or is rejected with a reason. No ambiguous outcomes, ever.
3. **Money is held transparently.** Every shilling of a buyer's deposit maps to a visible, releasable hold. Kenyan buyers are M-Pesa users; distrust of held funds kills conversion.
4. **Mobile is the primary client.** Most bidding will happen on a phone on mobile data, sometimes on poor signal.
5. **No dark patterns.** Reserve prices and buyer premiums are disclosed before bidding starts, not revealed in the invoice.

---

## 2. Users

| Persona | Needs | Success looks like |
| --- | --- | --- |
| **Individual buyer** (existing `buyer` account) | Buy a specific car without haggling; trust the seller; pay with M-Pesa | Wins a lot, pays within 72h |
| **Dealer / exporter** (existing `dealer` account) | Clear a slow-moving inventory fast; reach many bidders at once | Sells 3+ lots per auction day |
| **Auction admin** (internal ops) | Schedule lanes, approve dealers, release payments, resolve disputes | Runs an auction day with zero manual SQL |
| **Spectator** (logged out) | Watch a live lane | Follows the auction, prompted to register to bid |

We do **not** build a membership-tier business in v1. Copart's Basic/Premier tiers exist to gate deposits; we start with a flat deposit rule and revisit after we have real data.

---

## 2.1 Licence status — in progress

ShikaGari is **in the process of acquiring an auctioneer's licence** in Kenya. This is a stated part of the product plan, not an open question, and it shapes three things:

1. **Who conducts the sale.** ShikaGari conducts the auction as auctioneer; the dealer is the vendor. Terms of sale, invoice headers, and the auction calendar all state this. Buyers are told at registration time whose sale they are entering.
2. **What gates public auctions.** The licence is *in process*, not granted. Until it is, we may build and run the full system in staging and in invite-only internal sessions, but **public live auctions stay switched off by a single feature flag**. Nothing about the code changes when the licence lands — the flag flips, and the licence number appears everywhere it must.
3. **What the licence obliges us to keep.** A licensed auctioneer keeps a record of every sale. We assume that means a permanent, tamper-evident register of lots, bids, hammer results, buyers, and settlements — so `auction_audit_log` is a **compliance artifact with a retention policy**, not just a debugging aid. Counsel confirms the exact obligations and retention period in Phase 0.

**Dependency:** `auction_settings.license_enabled` (default `false`) gates public sessions, and `auction_settings.license_number` is displayed wherever the law requires it once granted.

---

## 3. Scope

### 3.1 In scope (v1, Phases 1–4)

- **Lot lifecycle:** convert an existing listing → auction lot with reserve, sale type, start time, lane.
- **Two-stage bidding:** preliminary (proxy) bidding → live auction.
- **Live room:** multi-lane, countdown clock, soft-close anti-sniping, increment + custom ("monster") bids, live outbid alerts.
- **Buying power:** deposit ledger, deposit locking, per-bid holds, automatic release.
- **Sale statuses:** pure sale, reserve-not-met (pending seller approval), sold.
- **Settlement:** invoice with buyer premium, 72-hour payment window, M-Pesa payment, storage-fee accrual after day 3.
- **SSO from the main site** via an "Auctions" navbar button; guests may spectate.
- **Licence gating:** public auctions are behind a feature flag until ShikaGari's auctioneer's licence is granted.

### 3.2 Explicitly out of scope (v1)

| Excluded | Why | Revisit |
| --- | --- | --- |
| Membership subscription tiers | Adds billing complexity before we know demand | Phase 6 |
| Import duty / customs calculator | Needs a licensed partner and current statutory rates | Phase 5 |
| Transport / delivery booking | Logistically heavy; third parties already do this | Phase 6 |
| Financing / hire purchase | Regulated product, needs a licensed lender | Never without a partner |
| Live video streaming of the yard | Bandwidth cost on 3G | Phase 6 |

---

## 4. Requirements

### 4.1 Auction lots

| ID | Requirement | Phase |
| --- | --- | --- |
| LOT-1 | A dealer can convert any of their own listings into a lot, reusing its photos, price and county | 1 |
| LOT-2 | A lot declares: reserve price (KES), sale type, open time, close time, lane | 1 |
| LOT-3 | Lots move through `draft → scheduled → open → on_block → sold / unsold / withdrawn` | 1 |
| LOT-4 | Sale types: `pure_sale`, `reserve` (hidden minimum), `on_approval` (seller discretion) | 1 |
| LOT-5 | Once a lot is `open`, vehicle details (year, make, mileage) are frozen; only price mechanics change | 1 |
| LOT-6 | An admin can withdraw a lot mid-auction with a recorded reason | 2 |
| LOT-7 | Every state change is written to an immutable audit table | 1 |

### 4.2 Preliminary (proxy) bidding

| ID | Requirement | Phase |
| --- | --- | --- |
| PRE-1 | A buyer submits a **maximum** bid on any `scheduled` lot up to 60s before the lot goes on block | 1 |
| PRE-2 | The system proxies the bid on the buyer's behalf during the live phase, bidding only as much as needed to stay high, never above their maximum | 1 (engine), 3 (proxy automation) |
| PRE-3 | A buyer may revise or withdraw their pre-bid any time before the lot opens | 1 |
| PRE-4 | The opening bid for a lot = second-highest pre-bid maximum + one increment, unless a reserve applies | 1 |
| PRE-5 | A pre-bid must be covered by available buying power at the moment it is placed | 1 |

### 4.3 Live auction room

| ID | Requirement | Phase |
| --- | --- | --- |
| LIVE-1 | Multiple lots may be on block simultaneously, one per lane; lanes are admin-scheduled | 2 (single) → 4 (multi) |
| LIVE-2 | Each lot shows a countdown from 60s, coloured green → amber → red | 2 |
| LIVE-3 | **Soft close:** any bid inside the final 10s resets the clock to 15s | 2 |
| LIVE-4 | Bidders can place the next increment with one tap, or type a custom higher amount ("monster bid") | 2 |
| LIVE-5 | Current price, leader, and your position update on every bid for every connected client within 500ms | 2 |
| LIVE-6 | Clients that disconnect mid-lot can reconnect and are resynchronised to authoritative state | 2 |
| LIVE-7 | A lot is hammered when the clock expires with no further bid | 2 |
| LIVE-8 | Bidding is rejected — never partially applied — when buying power is insufficient or the lot has closed | 2 |

### 4.4 Buying power and deposits

| ID | Requirement | Phase |
| --- | --- | --- |
| MONEY-1 | A buyer funds a deposit via M-Pesa before bidding; funds settle to a platform escrow ledger | 1 |
| MONEY-2 | v1 buying-power rule: **10% of your maximum active bid**. A KES 4,000,000 bid requires KES 400,000 held | 1 |
| MONEY-3 | Placing a bid places a **hold** on the deposit equal to the bid amount; the remaining deposit stays locked for the lot's duration | 1 |
| MONEY-4 | Being outbid releases the hold immediately; winning converts it to applied balance; losing at hammer releases it | 1 |
| MONEY-5 | Every ledger movement is a double-entry record with an immutable audit trail | 1 |
| MONEY-6 | Refunds return to the original M-Pesa number within 24h of release | 1 |
| MONEY-7 | Deposit and holds are computed from the ledger, never from a mutable counter column | 1 |
| MONEY-8 | v1 permits one active lot per buyer; v2 lifts this for verified dealers | 3 |

### 4.5 Sale outcome

| ID | Requirement | Phase |
| --- | --- | --- |
| SALE-1 | `pure_sale`: the high bidder wins the moment the clock expires | 2 |
| SALE-2 | `reserve` not met: the lot becomes **pending seller approval**; the seller may accept, counter, or decline | 2 |
| SALE-3 | A counter-offer has a 24h expiry and creates a binding commitment if accepted | 5 |
| SALE-4 | Winning locks the listing's logbook verification; the buyer's NTSA e-logbook check is reused from the existing platform | 5 |

### 4.6 Settlement

| ID | Requirement | Phase |
| --- | --- | --- |
| PAY-1 | On hammer, generate an invoice: hammer price + buyer premium + platform fee, itemised | 5 |
| PAY-2 | The buyer has 72 hours (Kenyan business days) to pay in full via M-Pesa | 5 |
| PAY-3 | Unpaid lots default after the window; the deposit is applied as a penalty and the lot re-enters the next auction | 5 |
| PAY-4 | Storage fee accrues daily from day 3 after hammer until pickup is confirmed | 5 |
| PAY-5 | Default fee and storage rate are admin-configurable, not hard-coded | 5 |

### 4.7 Access and identity

| ID | Requirement | Phase |
| --- | --- | --- |
| ACC-1 | An "Auctions" button on the main navbar links logged-in users into the engine with an existing session | 1 |
| ACC-2 | Guests can spectate live lanes; the bid button prompts them to register | 1 |
| ACC-3 | Only `dealer` accounts with an approved profile may create lots | 1 |
| ACC-4 | Only `buyer` and `dealer` accounts with a settled identity check may bid | 1 |
| ACC-5 | Every bid records the bidding account id and the IP/user-agent fingerprint for fraud review | 2 |

### 4.8 Auctioneer licence

| ID | Requirement | Phase |
| --- | --- | --- |
| LIC-1 | No public auction session may be created while `license_enabled` is false | 1 |
| LIC-2 | Once granted, the licence number appears in the auction footer, on every lot page, and on every invoice | 2 |
| LIC-3 | Terms of sale name ShikaGari as auctioneer and the dealer as vendor | 1 |
| LIC-4 | Every lot, bid, hammer, settlement, and default is retained in the audit log for the licence's retention period | 1 |
| LIC-5 | A named auctioneer operator account is required to open a session (accountability, not just permissions) | 1 |
| LIC-6 | Buyers are asked to acknowledge the terms of sale before their first bid | 2 |

### 4.9 Non-functional requirements

| ID | Requirement |
| --- | --- |
| NFR-1 | **Bid acceptance p99 < 150ms** from request to authoritative accept, engine-local |
| NFR-2 | **Fan-out p99 < 500ms** from accept to every connected client in the same region |
| NFR-3 | **Zero** accepted-but-lost bids. Total bid correctness over throughput |
| NFR-4 | Horizontal scale to 20+ concurrent lanes and 10,000 connected spectators |
| NFR-5 | Auction room usable at 3G latency (400ms RTT) and on a 360px-wide phone |
| NFR-6 | Engine recoverable from a crash with ≤ 5s of in-flight bid ambiguity, resolved from the durable log |
| NFR-7 | Every KES amount stored as integer shillings — never float |
| NFR-8 | All times stored UTC; rendered in `Africa/Nairobi` |

---

## 5. Copart → ShikaGari adaptations

| Copart concept | Our adaptation | Notes |
| --- | --- | --- |
| Basic membership, 10% deposit | Flat 10% of maximum active bid | No annual fee in v1 |
| Premier tier, flat deposit, unlimited power | Verified dealers get a flat, lower deposit multiplier | Phase 3; requires dealer verification |
| Bid 1 lot at a time (Basic) | v1: one active lot per buyer | Phase 3 relaxes for dealers |
| Card rails, wire transfer, cashier's check | **M-Pesa STK push** to a platform paybill | Deposit, balance, and refunds all on M-Pesa |
| Salvage title restrictions, broker bidding | **Not applicable** — Kenyan titles are NTSA logbooks | We reuse the existing e-logbook verification |
| On Approval, 8:00 PM next business day | Seller accepts / counters / declines within 24h | Aligns to Kenyan business hours |
| 3 business days to pay | 72 hours to pay via M-Pesa | Same shape, local rail |
| 2–3 days free storage, then daily fees | Day 3 onward, admin-configurable KES/day | Storage partner integration is Phase 6 |
| USD, US yards | **KES**, per-county vehicle location | Reuse `internal/domain/location.go` counties |

---

## 6. Phased delivery

Seven phases. Each has a goal, a hard boundary, and a definition of done. Effort is relative, not a commitment.

### Phase 0 — Foundations & commercial decisions
**Goal:** remove the unknowns that would force rework later.
**Work:** track the auctioneer's licence application to grant, with a dated plan; confirm with counsel the record-keeping and retention the licence obliges (feeding LIC-4); M-Pesa merchant account and callback URLs; decide escrow handling — segregated account vs internal ledger; define the buyer-premium rate card; decide lane-based vs single-stream scheduling; spike Redis availability and cost.
**DoD:** written money-flow diagram and rate card; M-Pesa sandbox working; counsel's list of licence obligations mapped to product requirements; staging can host a full internal auction day.
**Risk:** the licence is the only external dependency we do not control. Engineering proceeds regardless — the engine is fully buildable and testable behind `license_enabled = false`, so a licence delay delays *revenue*, not *delivery*.

### Phase 1 — Catalogue, deposits & pre-bidding (no live bidding)
**Goal:** a dealer can schedule a lot and a funded buyer can pre-bid on it, with money fully accounted for.
**In scope:** lot model and lifecycle; SSO handoff; M-Pesa deposit funding; double-entry ledger with holds; pre-bid placement/revision/withdrawal; buying-power gate; public auction calendar; admin lot scheduler.
**Out of scope:** any live streaming, timers, or hammer.
**DoD:** a dealer schedules a real lot; a funded buyer places, revises and withdraws a pre-bid; every shilling of deposit is visible in the ledger and reconciles to the cent; unauthenticated visitor sees the calendar but cannot bid.
**Effort:** large — this is the money-handling foundation and the hardest thing to retrofit.

### Phase 2 — Single-lane live auction MVP
**Goal:** one lane, one lot at a time, a real clock, real bids, real hammer.
**In scope:** WebSocket room; countdown with soft close; increment and custom bids; atomic bid acceptance; hammer; `pure_sale` and reserve-not-met outcomes; spectator mode; reconnect and resync.
**Out of scope:** multi-lane, proxy auto-bidding, invoices.
**DoD:** an internal auction can run end-to-end with real accounts and real KES; two simultaneous bidders never produce a split result under a load test; a dropped connection mid-lot rejoins with correct state.
**Effort:** medium on top of Phase 1. This is where correctness is proven.

### Phase 3 — Proxy bidding & buying power hardening
**Goal:** pre-bids start doing real work, and the money rules stop being provisional.
**In scope:** proxy bidding to the maximum; reserve enforcement; deposit locking across multiple lots; dealer tier with flat deposit; one-lot-per-buyer limit (buyers) and lifting it for dealers; bidding throttle and fraud signals.
**Out of scope:** subscriptions.
**DoD:** a pre-bid wins a lot against a live bidder without ever exceeding the buyer's stated maximum; no buyer can win a lot they cannot pay for.

### Phase 4 — Multi-lane concurrency & scale
**Goal:** several lots on the block at once, in front of a real audience.
**In scope:** lane scheduling and admin control; lane runner leader election; spectator scaling; regional fan-out; chaos testing of runner failover.
**Out of scope:** international regions.
**DoD:** 5 concurrent lanes sustain target latency under 10k connected clients; killing a runner mid-lot loses no accepted bid.

### Phase 5 — Settlement: invoices, M-Pesa payment, defaults, storage
**Goal:** a winner becomes a completed, paid-for sale.
**In scope:** invoice generation with itemised fees; 72-hour payment window via M-Pesa; default and penalty handling; seller counter-offer flow; storage-fee accrual; lot re-entry after default; buyer/seller notifications.
**Out of scope:** import-duty calculations, transport booking.
**DoD:** a hammer → M-Pesa payment → released-deposit ledger loop completes with no manual intervention; an unpaid lot defaults correctly on schedule.

### Phase 6 — Operations, trust & growth
**Goal:** run auctions repeatedly without the founder in the room.
**In scope:** dealer approval dashboard; dispute and refund tooling; reporting and reconciliation; SMS/email/WhatsApp notification service; fraud review queue; published buyer-premium and fee schedule; load and disaster-recovery drills; runbooks.
**Out of scope:** membership subscriptions (deliberately), transport marketplace.
**DoD:** an ops user runs a full auction day, reconciles the ledger against the M-Pesa statement, and resolves a test dispute without engineering help.

### Phase 7 — Launch readiness
**Goal:** safe to open to the public.
**Work:** penetration test, load test at 3× forecast peak, alerting and on-call, legal sign-off on published terms, fraud-mitigation rehearsal, staged rollout by invitation.
**DoD:** go/no-go checklist signed off.

### Phase dependency map

```
Phase 0 ──► Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6 ──► Phase 7
           money+model   clock+WS   proxy+tiers  scale      settlement   ops         launch
```

Phases 1 and 2 are strictly sequential — you cannot build a hammer before the money model is trustworthy. Phases 4 and 5 can partially overlap: multi-lane is an infrastructure concern, settlement is a business-logic concern, and they touch different tables.

---

## 7. Success metrics

| Metric | Target (first 90 days) |
| --- | --- |
| Lots scheduled per auction day | 20+ |
| Registered bidders | 500+ |
| Live-room concurrent spectators at peak | 2,000+ |
| Pre-bids per live lot | 3+ median |
| Bids accepted per lot | 5+ median |
| Hammer → paid settlement within 72h | 90%+ |
| Default rate (unpaid wins) | < 5% |
| Bid rejection due to engine fault | 0 |
| p99 bid acceptance latency | < 150ms |

---

## 8. Risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Engine bug awards a lot twice | Total reputational loss in a trust business | Single-writer lane runner, idempotent bids, immutable audit log, replay tooling. Never trust the client |
| Deposit/refund dispute | Regulatory and reputational exposure | Double-entry ledger, M-Pesa reconciliation daily, published refund policy, ops refund tooling in Phase 6 |
| Licence application delayed or refused | Cannot run public auctions; revenue slips | Build proceeds behind a feature flag. Licence is tracked as the single external dependency with a fallback: list-and-inquire only |
| Buyers on poor networks lose a lot they were winning | Complaints, chargeback pressure | Reconnect + resync, bid confirmation, 15s soft close, receipt for every accepted bid |
| Pre-bid proxy logic misunderstood as "guaranteed win" | Trust erosion | Copy that says maximum bid, never "guaranteed". See `design.md` |
| M-Pesa downtime during an auction | Bidders cannot fund or pay | Pre-funding window before lot opens; auction cannot start without sufficient funded buyers |

---

## 9. Open questions

1. What retention period and register format does the licence require? (Phase 0, counsel.)
2. Are deposits held in a segregated bank account or an internal ledger? Affects MONEY-1.
3. Can a lot be re-auctioned after default in the same session, or the next one? (Recommendation: next session.)
4. Do we allow dealers to bid on other dealers' lots? (Recommendation: no in v1.)
5. Is the 72-hour payment window business days or rolling hours, and how is a public holiday handled?
6. What is the published buyer-premium percentage, and is it per-lot or tiered by hammer price?
7. Does the licence restrict how many concurrent lanes we may operate, or require a named auctioneer physically present per session?