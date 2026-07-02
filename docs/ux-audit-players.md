# UX Audit: Players (Companion / Table+Hand Mode)

*Written 2026-07-02, based on hands-on journey testing against the
`per-person-mobile-ui` branch (Playwright at 390×844 for phones, desktop for
the Table, plus API-level probing). Each journey below was actually driven,
not desk-reviewed. Items marked ✅ FIXED were fixed during the audit passes;
the rest are the ranked backlog.*

## The journeys

1. **J1 — Host:** create a companion game, get the room on the shared screen,
   manage the room (lock, skip an absent player, downgrade to solo).
2. **J2 — Player joins & plays:** scan the QR (or type the code), pick an
   identity, land on the hand, play a full round.
3. **J3 — Interruptions:** step away (hide hand), phone dies / browser
   closes / reload → get back to your seat.
4. **J4 — Game night (asymmetric):** werewolf — seat picker, secret role on
   the phone, public table, voting.
5. **J5 — Stranger in the room:** wrong codes, full rooms, locked rooms,
   nosy observers (security-adjacent UX).
6. **J6 — Solo player unaffected:** the existing single-device flow keeps
   working (regression journey).

## What's genuinely good

- **Zero-typing join.** QR scan → guest → "Looks good — join!" → seated.
  Two taps, no account, no typing unless you *want* to customize. The
  randomized avatar+name front door is the single best decision in the flow.
- **The room reacts.** Your avatar pops onto the projector the moment you
  join; the current-player highlight moves live; deals animate deck → your
  name on the table and fly into your hand on the phone.
- **Privacy is a first-class citizen.** Own cards on your phone only
  (verified at the API level: opponents' hidden cards are placeholder ids,
  not data); 🙈 Hide-my-hand covers the whole screen for bathroom breaks.
- **Recovery is invisible.** Reload/reopen mid-game and you're back in your
  seat with your hand — cookies do the work, no re-join ceremony.
- **Room security matches the vibe.** Codes are rate-limited (10/min/IP),
  locked/finished rooms 404 identically to bad codes (no existence leak),
  host controls 403 for non-hosts, host-takeover has a 30s staleness gate.

## Fixed during this audit

| Journey | Friction | Fix |
|---|---|---|
| J2 | QR scan still required a "Join" tap on a prefilled code | ✅ auto-submit `?code=` |
| J2 | Enter/Go on the phone keyboard didn't submit the code | ✅ keydown handler |
| J1 | Table never updated when someone joined ("Nobody" until reload) | ✅ `presence-changed` broadcast on seat claim |
| J2 | Hand/Table wrapped in desktop chrome (roster, chat, admin, "I'm in!") | ✅ chrome hidden on companion surfaces |
| J2 | No feedback pressing Hit/Stand out of turn (silently swallowed) | ✅ buttons disabled via move legality + turn banner |
| J2 | Phone never showed *who you are* or *whose turn it is* | ✅ `renderHandHeader()`: "🐲 SlyFalcon · Waiting for AnimBot…" |
| J2 | Waiting banner offered solo-flow "Copy invite link" on companion surfaces | ✅ replaced with "Join code: XXXX" |
| J3 | No way to step away safely | ✅ hide-my-hand shield |
| J4 | Werewolf vote buttons said "Player 1/2/3" instead of who people are | ✅ labeled "🐺 WolfBot2" etc. via seatPresentations |
| J4 | Table declared "Villagers Win!" from Day 1 (computed from sanitized roles) | ✅ removed impossible client-side banner; server-side game-end filed |
| J2 | Moves silently dropped on companion surfaces (admin-controls hidden but load-bearing) | ✅ always mounted |

## Backlog (ranked)

### P1 — hurts every game night

1. **Game-over is a shrug.** Nothing special happens on any surface when a
   game finishes. The projector should celebrate (winners by avatar/name);
   phones should say you won/lost and offer "Play again". Werewolf can't even
   end (no `CheckGameFinished` — task filed). The "play again with the same
   room" loop doesn't exist; a new game means new code, everyone re-joins.
2. **Room code banner never shrinks.** The spec (§12) says big code during
   lobby, small corner code during play. It's projector-dominating forever.
   `renderRoomCodeBanner()` should collapse once the game starts (state
   phase ≠ gathering), with the QR available behind a tap.
3. **Absent-player UX is untested theater.** Heartbeat → absent badge →
   host SkipTurn exists and is unit-tested, but the real journey (phone
   sleeps → projector shows "Waiting for Alice…" → host skips) hasn't been
   driven end-to-end with real timing. Do a deliberate session; tune the
   30/60s thresholds to phone-lock reality (screens sleep in ~30s).
4. **Identity is split-brained where chrome remains.** The app drawer still
   shows the *account* ("Guest") while the game shows your picked identity
   ("🐲 SlyFalcon"). Roster (solo view) and chat tabs show "Player N" /
   account names. Companion identity (seatPresentation) should be the one
   name everywhere a companion game is on screen.

### P2 — polish that earns the "mobile UI" name

5. **Projector fullscreen mode.** The left app drawer (sign-in, admin
   toggle, menu) is still visible around the table view on desktop. A
   "present" affordance (fullscreen + hide drawer) — or hiding app chrome
   whenever surface=table — would make the projector feel intentional.
6. **Phone top bar.** The hamburger + "Boardgame App" banner survive on the
   hand. Fine, but a slimmer game-branded bar (game name + room code) would
   reclaim 56px on every phone.
7. **Join-flow dead ends.** "Room is full" leaves you staring at the avatar
   step with a toast; there's no "watch instead?" or "notify me" path.
   Mid-flow reload falls back to the code step (good) but forgets the code.
8. **Seat picker shows no faces.** Filled seats render as generic filled
   slots; they should show who's sitting there (avatar + name are already in
   the payload).
9. **Avatar name/emoji mismatch.** "QuickWhale" paired with 🐯. Either pair
   the noun list to the glyph list or leans into the absurdity deliberately.
10. **Hide-my-hand doesn't survive reload** (per-tab state). Arguably
    correct (reload = you're holding the phone), but decide on purpose.

### P3 — nice-to-haves observed

11. Blackjack table shows no visible cards (MVP renderer renders only deck +
    avatars) — players' public cards belong on the projector.
12. Chat is hidden on companion surfaces (right call for phones at a shared
    table) — but observers on the solo view can still chat; decide whether
    table-mode games want any chat at all.
13. No sound/haptics on "your turn" — a phone buzz is the natural cue when
    your eyes are on the projector. (Spec defers cross-surface audio; a
    local vibration is cheap.)
14. The `?display=` dev override and per-game surface cookies make one
    browser profile confusing when testing (documented in
    utils/companion-surface.ts, but worth a dev-tools note).

## J5/J6 notes (verified, no action)

- Brute-forcing `/api/join` reveals nothing about game structure (seat
  metadata requires auth); rate limiting engages after 10 tries.
- Non-companion games are untouched: solo flow, debug auto-seating in dev,
  and all 5 sibling-repo games' tests pass against the branch.
