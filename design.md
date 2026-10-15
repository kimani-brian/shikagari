# Design — ShikaGari Auction Experience

**Status:** Draft for review · **Last updated:** 2026-10-13
**Companion docs:** `prd.md` (what and why), `architecture.md` (how it is built)

Every screen is annotated with the **phase** it ships in. All screens live under `/auctions/*` inside the existing `shikagari-web` Next.js app and reuse its design system: neutral/`ink` palette, Inter + Material Symbols, `Button`, `Input`, `Badge`, `EmptyState`, `Pagination`, `CarCard`, `cn()`, `formatKES()`, `toast`, and the `Icon` wrapper.

---

## 1. Design principles

1. **One decision per screen.** A bidder should never be unsure what to press next. The auction room has exactly one primary action.
2. **Money is always visible.** Buying power, held amount, and the price you are bidding are on screen at all times — never behind a tap.
3. **The clock is the loudest thing.** When a lot is live, nothing else competes with the countdown visually.
4. **Honest about outcomes.** We never imply a pre-bid is "guaranteed". See §8 for the exact wording rules.
5. **Thumbs first, then desktop.** The bidder is on a phone, often on mobile data, sometimes one-handed in a queue.
6. **Say who is conducting the sale.** ShikaGari is the auctioneer. The dealership is the vendor. That distinction is visible on lot pages, terms, and invoices — not buried in a footer.
7. **Recoverable, never stuck.** Every error state has a next action. A dead "Bid" button is a failure state, not a design.

---

## 2. Information architecture

```
/                         Main navbar  ──►  [ Auctions ]  (new button)
/listings                 existing marketplace browse (unchanged)
/dashboard                existing seller/buyer dashboard (unchanged)

/auctions                                  PUBLIC — Phase 1
  ├── calendar            upcoming + live sessions
  ├── session/[id]        one auction day: lanes, schedule, live strip
  ├── lot/[id]            lot detail, pre-bid, photos, terms
  └── room/[id]           LIVE auction room                 Phase 2
        └── ?lane=A        lane focus for multi-lane         Phase 4

/auctions/me                               buyer — Phase 1
  ├── bids                my bids, leading / outbid, results
  ├── power               buying power, deposit, top-up       Phase 1
  └── invoices/[id]       invoice, 72h countdown, pay         Phase 5

/dealer/auctions                           dealer — Phase 1
  ├── new                 convert a listing → lot             Phase 1
  └── [id]                lot editor, schedule, results

/admin/auctions                           internal — Phase 1
  ├── sessions            create sessions, configure lanes
  ├── schedule            drag lots into lanes
  ├── approvals           reserve/on-approval queue           Phase 5
  └── reconciliation      ledger vs M-Pesa                   Phase 6
```

**Nav integration (Phase 1).** One new navbar item, `Auctions`, placed after `Vehicles` and before `Showrooms`. It is a plain route link — no auth check, because guests can spectate. The dashboard sidebar gains `Auctions` under `New Listing` for logged-in users.

---

## 3. Screen: Auction calendar `/auctions`

**Purpose:** answer "is anything on now, and what can I bid on?"
**Phase:** 1

| Element | Spec |
| --- | --- |
| Page header | "Auctions" + one-line plain subtitle: *"Live and upcoming vehicle auctions."* |
| Status filter | All · Live now · Today · Upcoming · Finished |
| Session cards | Date, start time (`EAT`), lane count, lot count, live badge if running |
| Live session card | Pulsing red dot + "LIVE" + count of lots currently on block |
| Empty state | `EmptyState variant="generic"` — "No auctions scheduled." + "Browse vehicles" link |
| Loading | 3 skeleton cards, matching the existing skeleton style |

Cards are the existing `CarCard` visual language: `rounded-2xl`, `border-neutral-200`, `shadow-card`. A live session card gets a `ring-1 ring-red-500` treatment so it is findable while scrolling.

---

## 4. Screen: Auction session `/auctions/session/[id]`

**Purpose:** orient a visitor before committing to the live room.
**Phase:** 1 (structure), 4 (live strip)

**Layout**

```
┌───────────────────────────────────────────────────────┐
│ SAT 14 MAR · From 10:00 EAT              [ Join room ] │
│ 4 lanes · 26 lots                                      │
├───────────────────────────────────────────────────────┤
│ LIVE NOW  (only when running — Phase 4)                │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐                 │
│ │ Lane A   │ │ Lane B   │ │ Lane C   │  horizontally    │
│ │ Hilux    │ │ Note     │ │ GLE 63   │  scrollable      │
│ │ KES 2.4M │ │ KES 890K │ │ KES 11.2M│  cards           │
│ │ 0:42    │ │ 0:17    │ │ —        │  live countdown  │
│ └──────────┘ └──────────┘ └──────────┘                 │
├───────────────────────────────────────────────────────┤
│ SCHEDULE                                                │
│ 10:00  Lane A  2019 Toyota Hilux    KES 2,400,000      │
│ 10:02  Lane A  2015 Nissan Note     KES 890,000        │
│ …                                                      │
├───────────────────────────────────────────────────────┤
│ Terms strip — static, honest, before anyone bids:      │
│ Sold by ShikaGari as auctioneer for <dealer>           │
│ Buyer premium 5% · Payment within 72 hours ·            │
│ Storage fee from day 3 · Licence no. <if granted>       │
└───────────────────────────────────────────────────────┘
```

The **terms strip** is not decoration. A bidder must know the premium and payment window *before* the hammer, not on the invoice.

---

## 5. Screen: Lot detail `/auctions/lot/[id]`

**Purpose:** everything needed to decide to bid.
**Phase:** 1

- Vehicle block reuses `CarCard` media + the existing spec grid (year, mileage, drive, fuel), plus the full spec table from the catalogue.
- Dealer byline badge, matching the new card convention: dealership name, or "Private listing".
- County of the vehicle, from the marketplace's existing county list.
- **Reserve disclosure:** for `reserve` lots, show "Reserve: not disclosed". For `pure_sale`, "No reserve — highest bid wins". Never expose the reserve number.
- **Bid panel** — state-dependent, this is the core component:
  | Lot state | Panel |
  | --- | --- |
  | `scheduled`, no pre-bid | Max-bid input + "Place pre-bid" |
  | `scheduled`, pre-bid placed | Your max, with edit / withdraw |
  | `open`, not on block | Pre-bid panel + "Join the live room" |
  | `on_block` | Redirect affordance: "This lot is on the block — open the room" |
  | `sold` / `unsold` | Result panel: hammer price, winner (or "no sale"), next steps |
- Deposit reminder inline if buying power < opening bid: *"Your deposit covers bids up to KES X. Top up to bid higher."* with a direct CTA.

---

## 6. Screen: The auction room `/auctions/room/[id]`

The highest-stakes screen in the product. Every second of attention is designed.

**Phase:** 2 (single lane) · 4 (lane switcher)

### 6.1 Desktop layout

```
┌──────────────────────────────────────────────────────────────────┐
│ ● LIVE   Lane A                          🔊   [ Session details ] │
├────────────────────────────────────┬─────────────────────────────┤
│                                    │  YOUR BID                  │
│  2021 Toyota Hilux Double Cab       │  ┌────────────────────────┐ │
│  ┌────────────────────────────────┐  │  │ Status: Leading        │ │
│  │                                │  │  └────────────────────────┘ │
│  │         vehicle photo          │  │  Bid: KES 2,400,000       │
│  │                                │  │  Next: KES 2,450,000      │
│  └────────────────────────────────┘  │                            │
│  KES 2,400,000        3 bids        │  [  BID KES 2,450,000  ]   │
│                                    │  [  Custom amount      ]   │
│        ╭─────────────╮              │  Your max: KES 3,000,000   │
│        │    0:37     │              │  Power:  KES 1,200,000 of  │
│        │  on  block  │              │           KES 3,000,000     │
│        ╰─────────────╯              │  [  + Top up deposit   ]   │
├────────────────────────────────────┴─────────────────────────────┤
│  Up next   2015 Nissan Note · Lane A · in ~2 min                  │
└──────────────────────────────────────────────────────────────────┘
```

### 6.2 Mobile layout (primary)

Single column, **sticky bid bar pinned to the bottom** so the thumb never travels:

```
┌──────────────────────────────┐
│ ● LIVE  Lane A      0:37     │  ← compact header + countdown
├──────────────────────────────┤
│  [photo, 16:10 ]             │
│  2021 Toyota Hilux           │
│  KES 2,400,000 · 3 bids      │
│  ── spec strip (scroll x) ── │
│  Year 2021 · 84,000 km · 4WD │
├──────────────────────────────┤
│  Leading · your max 3.0M     │  ← status line, one sentence
│  Power 1.2M of 3.0M           │
├──────────────────────────────┤
│ ┌──────────────────────────┐ │
│ │  BID  KES 2,450,000      │ │  ← sticky, ≥56px tall
│ └──────────────────────────┘ │
│  Custom amount · Top up      │
└──────────────────────────────┘
```

Rules: bid button is full-width, minimum 56px tall, safe-area padded for iPhone home indicator. Nothing else is fixed.

### 6.3 Timer design

A circular countdown with three visual states. Colour is the primary signal, reinforced by number size so it survives colour-blindness:

| Remaining | Ring | Number | Motion |
| --- | --- | --- | --- |
| 60–16s | `emerald-500` | 2xl, regular weight | Slow pulse, 2s cycle |
| 15–6s | `amber-500` | 2xl, semibold | Faster pulse, 1s |
| 5–0s | `red-500` | 3xl, bold | Pulse 0.5s + subtle scale nudge |

**Soft close:** when a bid extends the clock, the ring animates one full sweep and a toast-style line reads *"Bid placed — clock extended"*. The user must understand why they got more time.

The countdown is rendered from the server's `deadline_at`, interpolated locally, corrected on every tick. It never counts down from a client-side start value.

### 6.4 Bid pad

- **Primary button** = next increment. Always pre-filled with the exact amount. Never a slider: precision beats feel at this moment.
- **Custom amount** = a separate, visually secondary control. Label: *"Bid more"*. Placeholder shows the minimum allowed. Validation is local first (must exceed current + increment), authoritative second (the server decides).
- **Monster bid** needs no separate concept. A typed amount is a monster bid.
- Both controls are **disabled with a reason** when: not logged in, deposit unsettled, already leading, lot closed, or engine busy. A disabled button always shows why underneath — never a dead control.

### 6.5 Your-bid states

| State | Presentation |
| --- | --- |
| **Leading** | Green status dot, "You're highest — KES X". Bid button replaced by "You're leading". |
| **Outbid** | Red status line, "Outbid at KES X". Bid button becomes urgent (dark, pulse once). Auto-offer stays armed if a max was set. |
| **Winning via proxy** | "Auto-bidding to your max of KES Y". Distinct from a manual win so the user knows their ceiling is being spent. |
| **Lost** | Final, muted: "Sold for KES X to another bidder". |
| **Won, reserve met** | Strong confirmation + invoice CTA + 72h countdown. |
| **Won, needs approval** | Distinct state: "Sold subject to dealer approval. Decision by <time>." Never the same green as a clean win. |

### 6.6 Spectator vs bidder

Guests see the full live room with the bid panel replaced by a single clear CTA: *"Register to bid"* — and after login they land back on the same lot. Never a modal wall on first paint.

### 6.7 Connection states

| State | UI |
| --- | --- |
| Connecting | Skeleton for live figures, dimmed panel, "Connecting to the auction…" |
| Live | Everything normal |
| Reconnecting | Persistent amber strip under the header: *"Reconnecting — your last bid is safe."* Bid button disabled. |
| Offline | Bid button disabled with reason *"You're offline. Reconnect to bid."* |
| Resynced | Brief "Prices updated" flash when the client's `version` was stale. |

The reassurance line "your last bid is safe" is deliberate: the #1 fear during a live auction is that a bid "didn't go through".

---

## 7. States and error copy

Server error codes map to fixed UI copy. No generic "Something went wrong" on money or bidding paths.

| Code | UI treatment | Copy |
| --- | --- | --- |
| `BID_TOO_LOW` | Inline under the bid pad | "Too low. The next bid is KES X." |
| `LOT_CLOSED` | Modal + auto-navigate to result | "Bidding closed on this lot." |
| `SELF_OUTBID` | Inline, non-blocking | "You're already the highest bidder." |
| `INSUFFICIENT_BUYING_POWER` | Panel with top-up CTA | "Your deposit covers bids up to KES X." |
| `ALREADY_HOLDING_LOT` | Inline | "You're bidding on another lot. Finish that one first." |
| `NOT_FUNDS_SETTLED` | Pending state + retry | "Waiting for your M-Pesa payment to confirm." |
| `ACCOUNT_NOT_VERIFIED` | Blocking panel linking to verification | "Verify your account before bidding." |
| `ENGINE_BUSY` | Toast, auto-retry in 2s | "Busy — trying again." |
| `RATE_LIMITED` | Inline under the bid pad, clears on next allowed bid | "Slow down — too many bids. Try again in a few seconds." |
| `NETWORK` | Reconnecting strip | "Reconnecting — your last bid is safe." |

---

## 8. Copy rules

The product is financial; vague or padded copy reads as untrustworthy. House rules:

| Do not write | Write instead |
| --- | --- |
| "You're guaranteed to win!" | "Your maximum bid is KES X." |
| "Cars you have hearted. Tap the heart again to remove one." | *(nothing — just the heading)* |
| "Experience the thrill of live bidding" | "Bidding closes in 0:37" |
| "Your deposit is safe with us" | "Your deposit is KES 400,000. It is released when you are outbid." |
| "Something went wrong" | The specific reason and the next action |
| "Licence no. 12345" before we hold one | *(omit the line entirely)* |
| "Reserve price: confidential" | "No reserve — highest bid wins" |

Also: no exclamation marks in money and bidding surfaces; no emoji; sentence case throughout; never describe a pre-bid as an offer that guarantees anything. Match the existing `formatKES()` output everywhere so figures are formatted identically across the platform.

---

## 8.1 Compliance and trust surfaces

Because ShikaGari conducts the sales as auctioneer, three things must be visible without hunting:

| Surface | Shows | When |
| --- | --- | --- |
| Auction footer (every page) | "ShikaGari · Auctioneer licence no. <number>" | Only once granted; before that the footer omits the number rather than showing a placeholder |
| Lot page, above the bid panel | "Auctioneer: ShikaGari. Vendor: <dealer name>." | Always |
| Invoice | Licence number in the header block; "Auction fee" and "Buyer premium" as separate lines | Always |
| Terms of sale | Full terms, acknowledged once before a user's first bid (LIC-6) | Always |

Before the licence is granted, the **pre-launch state** is designed rather than hidden: the navbar entry shows "Auctions — coming soon", the calendar renders, and no session can be opened. We never show a licence number we do not hold.

**Phase:** terms and vendor/auctioneer line in Phase 1; licence number surfaces Phase 2 (LIC-2) once granted; first-bid acknowledgement Phase 2 (LIC-6).

---

## 9. Component inventory

**Reuse from `shikagari-web`:** `Button`, `Input`/`SelectField`/`Textarea`, `Badge`/`ApprovalBadge`/`ListingStatusBadge`, `EmptyState`, `Pagination`, `Skeleton`/`PageLoader`, `CarCard`, `SearchBar`, `Modal` pattern, `Icon`, `cn`, `formatKES`, `formatMileage`, `timeAgo`, `react-hot-toast`.

**New, auction-specific:**

| Component | Purpose | Phase |
| --- | --- | --- |
| `AuctionNavbarButton` | Navbar entry | 1 |
| `SessionCard` | Calendar card, live variant | 1 |
| `LiveDot` | Pulsing LIVE indicator | 1 |
| `LotSpecGrid` | Full spec table | 1 |
| `ReserveNotice` | Reserve disclosure block | 1 |
| `PreBidPanel` | Max-bid form with edit/withdraw | 1 |
| `BuyingPowerMeter` | Available vs held vs max | 1 |
| `DepositTopUp` | M-Pesa STK flow | 1 |
| `AuctionRoom` | WS-connected container + state machine | 2 |
| `CountdownRing` | Circular timer, 3 states | 2 |
| `BidPad` | Increment + custom bid | 2 |
| `BidStatusLine` | Leading / outbid / proxy states | 2 |
| `ConnectionStrip` | Reconnecting / offline banner | 2 |
| `SoftCloseToast` | Clock-extension notice | 2 |
| `LaneSwitcher` | Multi-lane navigation | 4 |
| `InvoicePanel` + `PaymentCountdown` | 72h window | 5 |
| `CounterOfferPanel` | Seller accept/counter/decline | 5 |
| `AdminLaneBoard` | Live ops board, one column per lane | 4 |
| `LicenseBadge` | Licence number in footer/invoice; hidden until granted | 2 |
| `TermsOfSaleGate` | One-time acknowledgement before a user's first bid | 2 |
| `VendorAuctioneerLine` | "Auctioneer: ShikaGari. Vendor: X." on lot pages | 1 |
| `LedgerReconciliationTable` | M-Pesa vs ledger | 6 |

---

## 10. Motion

Motion explains state changes; it never delays an action.

| Element | Motion | Duration |
| --- | --- | --- |
| Timer ring | Continuous pulse, faster as time runs out | 2s → 0.5s |
| Soft close | Ring sweeps once, then resumes | 400ms |
| New bid lands | Price flashes neutral-900 → default, leader row slides in | 180ms |
| You are outbid | Status line shake, once, no bounce | 250ms |
| Bid accepted (you) | Button → check mark, brief, then reverts to "You're leading" | 300ms |
| Hammer | Full-width flash, result panel slides up | 500ms |
| `prefers-reduced-motion` | All of the above become instant state changes | — |

Existing `animate-fade-up` / `animate-fade-in` classes are reused, keeping the marketplace and auctions visually continuous.

---

## 11. Accessibility

- Timer conveys urgency through number size and weight as well as colour (NFR: not colour alone).
- Bid button ≥56×56px touch target; all controls reachable one-handed at the bottom of the screen.
- Live region: price, leader, and your bid status announced via `aria-live="polite"`; hammer via `assertive`.
- Full keyboard path through the room: `Tab` to the bid pad, bid with `Enter`, custom amount reachable without a mouse.
- Focus is trapped in modals and restored on close; the bid pad keeps focus after each bid so repeated bidding does not require re-tapping.
- Contrast ≥4.5:1 for all price and status text, including on the red and green timer rings.

---

## 12. Responsive breakpoints

| Width | Treatment |
| --- | --- |
| ≤360px | Single column; sticky bid bar; specs horizontally scrollable; no lane switcher (Phase 2) |
| 361–767px | Single column, larger touch targets |
| 768–1279px | Room: single column with a two-up bid panel |
| ≥1280px | Two-pane room as in §6.1; lane switcher rail (Phase 4) |

The auction room is the only screen in the product designed 360px-first. Everything else follows the marketplace's existing patterns.

---

## 13. Phase → screen checklist

| Phase | Screens shipped |
| --- | --- |
| 0 | Nothing. Wireframes reviewed by dealers and pilot bidders; copy and premium rate card signed off |
| 1 | Calendar, session, lot detail, pre-bid panel, buying power, deposit top-up, "my bids", navbar button, dealer lot creation, admin session + scheduling, vendor/auctioneer line, pre-launch "coming soon" state |
| 2 | Auction room: countdown ring, bid pad, bid status line, connection strip, soft-close notice, spectator mode; licence badge and terms-of-sale gate (licence granted) |
| 3 | Auto-bid status line, reserve notices, dealer power tier, top-up prompts inline in the room |
| 4 | Live session strip, lane switcher, admin lane board |
| 5 | Invoice panel, 72h payment countdown, counter-offer panel, default notices |
| 6 | Reconciliation table, dispute tooling, notifications preferences, audit explorer |
| 7 | Accessibility pass, reduced-motion pass, empty and edge-state sweep, load-test UX validation at 3G |

---

## 14. Open design questions

1. Should the bid pad show the previous bidder's name, or stay anonymised? (Recommendation: anonymised as "Bidder 7" until hammer — reduces harassment and collusion.)
2. Does a spectator need audio cues, or is the visual ring sufficient on mobile where audio is often muted? (Recommendation: visual-only in v1; audio is opt-in in Phase 6.)
3. How prominently should the storage fee be surfaced before bidding? (Recommendation: in the session terms strip and on the invoice, never on the bid pad.)
4. Should the lane switcher be swipeable on mobile, or a dropdown? (Recommendation: dropdown — swipe conflicts with photo gallery swiping.)