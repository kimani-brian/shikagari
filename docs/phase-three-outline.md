# Phase 3 Follow-ups

## Follow-up #2 – End-to-End validation
- [ ] Run `npm run lint && npm run typecheck && npm run build` for both API and frontend before every push.
- [ ] Capture a fresh `scripts/account-dashboard.png` after UI adjustments and attach it to the verification notes.
- [ ] Execute `node scripts/verify-account-security.mjs` with real credentials to assert: login success, hero metrics match API counts, security events + sessions render newest-first, and the revocation CTA works.
- [ ] Expand scripted coverage to include MFA/login edge cases once the backend exposes the feature flag.
- [ ] Store raw API payloads from `/auth/security-events` and `/auth/sessions` beside the Playwright artifacts for quick diffs if regressions appear.

## Follow-up #3 – Launch readiness
- [ ] Confirm env files for API + frontend list the same required variables (JWT secret, upload bucket, CDN URL, etc.).
- [ ] Document the QA sign-off procedure (Playwright script + manual smoke of listing flows) in `README.md`.
- [ ] Add uptime/latency monitors for `/auth/security-events`, `/auth/sessions`, and `/listings` endpoints in the preferred APM.
- [ ] Ensure dealer/buyer roles have updated onboarding copy and that marketing links are wired up in the dashboard hero.
- [ ] Prepare a rollback plan (database backup + feature flag) before announcing the account security revamp.
