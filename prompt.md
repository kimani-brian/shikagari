# SYSTEM SPECIFICATION & ARCHITECTURE PROMPT: REAL-TIME VEHICLE AUCTION ENGINE

## 1. PROJECT OVERVIEW & GOAL
I operate an existing online automotive marketplace where dealerships and private sellers list vehicles and receive buyer inquiries. 

I need you to act as a **Lead Systems Architect and Principal Full-Stack Engineer**. Your task is to design and provide the implementation blueprint for a **standalone, high-concurrency Live Auction Engine** modeled after enterprise auto-auction platforms (like Copart VB3). 

This engine must operate as an **independent service** while seamlessly integrating with our existing web platform via a top navigation bar button labeled **"Auctions"**.

---

## 2. INTEGRATION & USER ENTRY POINT
1. **Entry Point:** 
   - A prominent button named **"Auctions"** on the main website’s top navbar.
   - When clicked, users are routed to the auction portal (e.g., `auction.ourdomain.com` or `/auctions`).
2. **Authentication / Single Sign-On (SSO):**
   - Users should not need to register twice. Implement a secure token-based handoff (e.g., JWT / OAuth2 / Shared Session Cookie).
   - If a logged-in user clicks "Auctions", they must enter with an authenticated session. If an unauthenticated guest clicks it, they can view live lanes as spectators but must be prompted to log in/register to place bids.
3. **Vehicle Listing Bridge:**
   - Existing sellers should be able to select any existing vehicle listing from their dealer dashboard and click **"Convert to Auction"** or **"Submit to Auction"**, setting auction-specific parameters (reserve price, auction date/time, sale type).

---

## 3. CORE MODULES & FUNCTIONAL REQUIREMENTS

### A. Real-Time Live Auction Room (WebSockets / PubSub)
- **Multi-Lane System:** Ability to run single or simultaneous auction lanes.
- **Dynamic Countdown Clock:** 
  - A fast-paced countdown timer (e.g., 30–60 seconds per lot).
  - **Anti-sniping / Soft-close logic:** Placing a bid in the final 10 seconds resets the timer back to 15 seconds to allow counter-bids.
- **Bid Placement:** 
  - Next increment bidding (pre-calculated increments based on current price).
  - "Monster Bid" (custom bid higher than the increment).
- **Concurrency & Race Conditions:**
  - Redis-backed distributed locks or atomic queues to handle simultaneous bids within milliseconds without race conditions or dirty writes.

### B. Two-Stage Bidding Lifecycle
1. **Preliminary Bidding (Pre-Bids):**
   - Proxy bidding open prior to the live sale. Users set their maximum ceiling; system auto-bids up to that ceiling.
2. **Live Floor Transition:**
   - Pre-bids determine the opening floor price of the live auction.
3. **Sale Status Types:**
   - **Pure Sale:** Lot automatically awarded to highest bidder when timer expires.
   - **On Approval / Reserve:** If highest bid is below reserve, marked as "Pending Seller Approval" with an automated counter-offer/approval workflow.

### C. Buyer Verification, Deposit & Buying Power Engine
- **Buying Power Ledger:**
  - Track user deposits and available bidding limits before allowing bid submission.
  - Basic Tier: 10% cash deposit rule (e.g., $500 held in deposit unlocks $5,000 maximum active bidding power).
  - Premier / Verified Dealer Tier: Unlimited/flat-rate deposit bypass upon business verification.
- **Deposit Locking:** Deposits must be locked during active high-bids and unlocked automatically if outbid or won & cleared.

### D. Billing, Fees & Post-Auction Workflow
- Automatic invoice generation upon sale completion:
  - Hammer price + dynamic buyer premium fee + platform processing fees.
- Payment countdown window (e.g., 72-hour timer for full balance settlement).

---

## 4. WHAT YOU NEED TO PRODUCE

Please produce a comprehensive, structured implementation plan covering:

1. **High-Level System Architecture:**
   - Recommended tech stack for high-throughput, low-latency live operations (e.g., Node.js/Go, Redis Pub/Sub, WebSockets/Socket.io, PostgreSQL).
   - Decoupling strategy between the existing website and this new auction engine.
2. **Database Schema (Relational/PostgreSQL):**
   - Tables for: `AuctionLots`, `LiveBids`, `UserWallets/Deposits`, `AuctionLanes`, `AuctionSettings`. Include data types, indices, and foreign keys.
3. **Real-Time WebSocket State Machine:**
   - Describe the exact WebSocket event flow (`LOT_STARTED`, `BID_SUBMITTED`, `BID_ACCEPTED`, `TIMER_TICK`, `LOT_HAMMERED`).
   - Strategy for handling network lag, disconnections, and server reconciliation.
4. **API Contracts & SSO Flow:**
   - Sequence diagram / step-by-step explanation of how a user transfers from the main site to the auction engine.
   - Key REST endpoints for inventory sync and administrative lot scheduling.
5. **Phase-by-Phase MVP Roadmap:**
   - Clear development phases from MVP (Single lane, pre-bidding + simple live timer) to Full Scale (Multi-lane, dynamic anti-sniping, dealer approval dashboard).

   we should build our version of this