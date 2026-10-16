# Phase 7 Closeout — Launch Readiness (2026-10-16)

PRD DoD: **go/no-go checklist signed off.** This file is the checklist.
`GO` = verified this session · `MANUAL` = needs a human · `DEFERRED` = known gap, tracked.

## Engineering verification (GO)

| Item | Evidence |
| --- | --- |
| Web `tsc --noEmit` clean, `npm run build` green, all `/auctions/*` routes emit | Build log 2026-10-16 |
| ESLint: 4 touched files clean, zero new problems repo-wide | Per-file eslint run |
| Auction engine `go build ./...` + `go test ./...` all packages ok | Engine test run |
| Terms of sale acknowledged once before first bid (LIC-6) | New `TermsGate` component; CDP browser test: gate opens, no `BID_SUBMIT` before accept, `BID_SUBMIT` fires after, ack persisted per user id, skipped on revisit |
| Every money/bid form submits with Enter (room custom bid, lot pre-bid, deposit top-up, invoice pay) | `<form onSubmit>` wraps; room custom bid keeps `type="button"` increment separate so Enter never double-bids |
| Bid buttons meet the 56px touch target (design §11) | `min-h-[56px]` on desktop + sticky bar bid buttons (`Button lg` alone is 44px) |
| Sticky mobile bid bar no longer covers page content | `h-24 lg:hidden` spacer in normal flow |
| Custom bid input labelled for screen readers | `aria-label="Custom bid amount"` (`Input label=""` rendered no label) |
| Bad lot id renders a recoverable error, not a dead room | CDP test: unknown UUID → "This auction room does not exist." + back link |
| Guest room: spectator CTA, zero JS errors | CDP test on `/auctions/room/:bad-id` and `/auctions` |
| Reduced-motion kills all auction pulses (design §10) | Pre-existing global `@media (prefers-reduced-motion: reduce)` in `globals.css`; covers `animate-pulse` on timer ring, LIVE dot, skeletons |
| Live regions: countdown polite, outbid persistent polite region, hammer assertive | Pre-existing in room page + `CountdownRing` |
| Edge states: session/lot/invoice load failures, empty lanes, empty bids/invoices lists | Pre-existing; swept this session, no dead ends found |
| Metrics, Prometheus alerts (`prometheus/alerts.yml`), runbooks (`runbooks.md`) present | Engine repo inspection |
| Licence gating: `license_enabled=false` blocks public sessions; licence number surfaces only on fees page when granted (LIC-1, LIC-2) | Pre-existing; no licence number shown anywhere pre-grant |

## Manual sign-off required (MANUAL)

- [ ] Penetration test findings triaged (external).
- [ ] Load test at 3× forecast peak; alerting wired to on-call (alerts defined, not wired).
- [ ] Legal sign-off on published terms + licence application status; flip `license_enabled` only when granted.
- [ ] M-Pesa production credentials + callback verification in staging.
- [ ] Fraud-mitigation rehearsal with ops (runbooks §3).
- [ ] Staged rollout by invitation.

## Known gaps (DEFERRED, not launch-blocking for invite-only staging)

1. **Vendor name not shown on lot/room pages.** Design §8.1 wants "Auctioneer: ShikaGari. Vendor: X." The engine `AuctionLot` snapshot carries no dealer name, so pages show only the auctioneer line. Needs an engine change (expose dealer name on lot responses) + frontend line. Tracked, not built.
2. **Dev passwords are weak** (`Admin123`, `Normal123`) — rotate before any shared staging.
3. **`.env` with real secrets is tracked in git history** (backend repo) — rotate `DB_PASSWORD`/`JWT_SECRET` and purge history before public work.
4. **Leftover engine data:** withdrawn test lot from this session (correct per audit-trail policy) plus pre-existing `P7 load` session with 3 sold lots. Harmless in staging; purge before launch.

## Files changed this session (uncommitted)

- `shikagari-web/src/components/auctions/TermsGate.tsx` (new)
- `shikagari-web/src/app/auctions/room/[lotId]/page.tsx` (gate wiring, custom-bid form + label, 56px buttons, sticky spacer)
- `shikagari-web/src/app/auctions/lot/[id]/page.tsx` (pre-bid form)
- `shikagari-web/src/app/auctions/me/power/page.tsx` (top-up form)
- `shikagari-web/src/app/auctions/invoices/[id]/page.tsx` (pay form)
