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
7. **J7 — Move the projector:** transfer an active shared Table to a fresh,
   accountless screen without pausing play or exposing a seated player's hand.
8. **J8 — Play again:** finish a game, create one successor, and carry the
   Table plus every Hand and identity into it without rejoining.

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
  seat with your hand — no re-join ceremony. Reconnect and tab-resume refresh
  non-version room state too, so roster, lock, presence, and Table ownership
  cannot remain stale merely because no move happened during the outage.
- **Room security matches the vibe.** Codes are rate-limited (10/min/IP),
  locked/finished rooms 404 identically to bad codes (no existence leak),
  host controls reject non-lease devices; Table recovery uses an expiring
  socket-renewed lease and a storage-backed first-winner takeover.
- **Projector changes are transactional.** The old Table keeps working while a
  five-minute QR/manual offer is pending; explicit confirmation atomically
  fences it only after the replacement is ready. Lost responses are safe to
  retry on the same device, other devices cannot replay the offer, and the old
  screen says where the game went instead of failing mysteriously.
- **Tabs stay independent.** Hand/Table renderer intent is per-tab rather than
  origin-wide; opening or restoring one surface cannot silently change another
  tab. Server-declared solo mode always overrides stale browser intent.

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
| J4 | Table declared "Villagers Win!" from Day 1 (computed from sanitized roles) | ✅ server-owned team win conditions and winner lists |
| J2 | Moves silently dropped on companion surfaces (admin-controls hidden but load-bearing) | ✅ always mounted |

## Backlog (ranked)

### Fixed in the second pass (2026-07-02)

- ✅ **Game-over display**: `Finished`/`Winners` plumbed to every renderer;
  the projector celebrates ("Game over! 🦊 Ada wins!" with avatars), the
  phone header shows 🎉 You won! / you lost / draw. Verified with a real
  finished game. Also: the connection-lost dim no longer engages on
  finished games (the socket closing at game end is expected).
- ✅ **Room code banner shrinks** to a corner "Room XXXX" badge once every
  seat is claimed (the moment the code stops mattering). The old
  version-based check read a field that doesn't exist and never fired.
- ✅ **Absent-player journey driven with real timing** — and it found a
  hole: heartbeats were only tracked after a phone connected, so a
  no-show who claimed a seat and never arrived looked present forever
  (the exact case SkipTurn exists for). Seat claims now seed the absence
  clock; verified live: claim → Absent after ~35s → host SkipTurn
  advances the game.
- ✅ **Projector/phone fullscreen**: the app drawer collapses to a
  hamburger overlay at every width on companion surfaces.
- ✅ **Join-flow dead ends**: "Room is full" now offers "Watch this game
  instead" (Table view); the validated code survives a mid-flow reload
  via sessionStorage.
- ✅ **Seat picker shows faces** (avatar + name on filled slots, "open"
  on empty ones).
- ✅ **Avatar name matches glyph** ("🦊 CleverFox", never "🐯 QuickWhale")
  via a glyph-aligned noun list; free-form nouns remain for Customize.
- ✅ **Blackjack table shows each player's public cards + score**, with a
  gold highlight on the current player.
- ✅ **Your-turn haptic**: the hand base buzzes the phone
  (`navigator.vibrate`, progressive enhancement) when it becomes your turn.
- ✅ **Play-again loop**: a finished room offers one idempotent successor;
  owner/Table races converge, the Table capability and exact human seat/name/
  avatar bindings carry forward, every open companion surface follows once
  setup is published, and the new room receives a fresh join code.

### Remaining backlog

The Werewolf follow-up is complete: its delegate now owns team win conditions;
day and night votes are separate public/private fields; and each wolf receives
an owner-only, server-computed teammate list. API-level tests pin all three
privacy views (self, other player, and observer).

1. **Identity unification in solo-view chrome.** On companion surfaces the
   account name is now hidden with the drawer, but the solo view of a
   companion game still shows account names / "Player N" in roster and
   chat tabs.
2. **Phone top bar** could become game-branded (name + room code) and
   reclaim 56px; hamburger + "Boardgame App" survive today.
4. **Hide-my-hand is per-tab** and doesn't survive reload — decided:
   acceptable (a reload means you're holding the phone).
5. Chat on table-mode games: observers can still chat from the solo view;
   decide whether that's wanted.
6. Absent thresholds (30s) worked well in testing; revisit against real
   phone-lock behavior at a live game night.

### Frame-by-frame animation review (2026-07-02, adversarial agent)

An instrumented reviewer sampled every card's position per frame during
real deals on three simultaneous pages. Round 1 verdict: the cross-screen
animation feature **did not exist** — the phone fly-in never fired (a Lit
`id` property shadowed `Element.id` without reflecting, so no id lookup
ever matched a card, silently), the table's chip flight played at
opacity 0 from a wrong origin, and the sync estimator had no consumers.
After fixes, round 2 confirmed with frame evidence: fly-in real (top-edge
entry, 600ms monotonic ease-out), chip flight visible from the deck, no
regressions, hand wrap works. Remaining (tracked in the sync follow-up
task): a ≤2-frame FLIP/animateBetween handoff blip with ~40px origin
skew, and deal-phase cross-screen drift (up to ~1s on the 4th card of a
deal) until serverPlayAt is actually consumed. Investigated afterward and RESOLVED as a test-rig artifact, not a
product bug: a real UI join authenticates before its socket connects and
stays present past the 30s threshold (verified live); the reviewer's
localStorage-injected pages never fully signed in, so their heartbeats
credited the observer index.

## J5/J6 notes (verified, no action)

- Brute-forcing `/api/join` reveals nothing about game structure (seat
  metadata requires auth); rate limiting engages after 10 tries.
- Non-companion games are untouched: solo flow, debug auto-seating in dev,
  and all 5 sibling-repo games' tests pass against the branch.
