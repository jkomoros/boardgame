# Viewer-bound release and list release, by use case

An evaluation of two candidate CFC mechanisms for releasing data to people,
worked through eleven use cases. Written 2026-11-25. Read-only: no repository
file was changed, nothing was run, and nothing was posted.

## How to read the evidence marks

- **[R]** read: the claim comes from a document or source file I read, named
  where it matters.
- **[R-ran]** read, and the source says a test pins it. I did not run the test.
- **[I]** inferred: my conclusion from the read material. Nothing marked [I] was
  run.

I ran no tests and no runtime. Every "works" below is either the source
documents' report of a run, or my inference.

What I read:

- `MODULE-NAMED-AUDIENCE-PROPOSAL.md`, `SEAT-VIEWS-FINDINGS.md`, and the head of
  `card-table.tsx` in the worktree `viewer-bound-rebased`; that worktree's
  `docs/features/current-principal.md` ("In a computation") and the minting
  site in `runtime/runner/src/cfc/prepare.ts`.
- Specs checkout at `main` (`f549c796`): §3.6.1, §3.6.4, §4.3.5, §4.9, §5.3
  with §5.3.2 and §5.3.3, §8.15.8, §15's `TransformedBy` row, decision D5.
  §8.15.12, §4.9.5, §8.7.5 and §8.12.5.1 are not on `main`; I read them only as
  their pull requests describe them.
- Specs pull requests #59, #50, #95, #90 (bodies and comments), #72 (summary
  only), #51 (title and dependency only).
- `docs/specs/server-side-execution/scopes.md` §1, §2 (the fan-out rulings), §4,
  §5.
- Labs patterns: `direct-release.tsx`, `apps/redactor/policy.tsx`,
  `custody-answer-room.tsx` (header), `showcase/inductor/pool` (policy and
  README), `system/private-inbox`, `cfc-row-label-mailbox`,
  `cfc-row-label-records`, `apps/recommend-a-book/privacy.tsx`,
  `examples/fair-share`, the FabriChat manager's `TODO(direct-room-label)`.
- Loom (`/Users/jkomoros/Code/commonfabric/loom`): `cf-location-findmy.tsx`,
  `questions-policy.ts`, and parts of the Questions design document.

Not read: the Lean development, `DERIVED-RELEASE-FINDING.md` beyond its
headings, `seat-views.tsx`, the labs#8608 diff, conclave's own sources, and
`cfc/notes/FUTURE-SPEC-WORK.md` beyond its headings (the "List Release
Follow-Ups" section lives in specs#59's diff, not on `main`).

## The two mechanisms in one paragraph each

**Viewer-bound release** [R]. A blessed reactive computation calls
`currentPrincipal()`. The runtime answers with the principal the run is
computed for, the run narrows to `user` scope so each demanding principal gets
an instance, and the flow stage mints `TransformedBy{identity, inputWitness,
viewer}` on what the run writes. One value-intrinsic module rule binds a
variable from `viewer` and adds `User(viewer)` to the policy clause. The
function decides from the sealed state what that principal's output holds.
Nothing is stored per recipient; an instance exists per principal who looks.

**List release** [R]. Specs#59 option E. A `Members{list, owner}` atom names a
list held in a cell. An authored clause `[User(Q) ∨ Members(L, S)]` is admitted
by a capture check (the actor solely owns the subject, the output and list are
rooted in one run of the policy's module, the list is owner-bound and declares
its writers). A module rule moves `Members(L, S)` into the policy clause under
the transformation's witness. At display the runtime mints `ListedIn{p, L}` by
point query, so removing an entry revokes.

## What actually differs

Before the cases, the properties that turned out to decide them. All [I] unless
marked.

| Property | Viewer-bound | List |
| --- | --- | --- |
| When the audience is bound | At computation, by code, once per run | At each access, by reference to a cell |
| Who decides the audience | The blessed function, from sealed state | Whoever the list's declared writer admits |
| What the label holds | A literal `User(q)` [R] | A reference `Members(L, S)` [R] |
| Copies of the value | One per principal who looks [R] | One, shared |
| Content per reader | Any function of the reader | The same bytes for every member |
| Label-level revocation | None. The next computation holds less; what was derived keeps `User(q)` | Yes, at the next access, and on derived values too [R] |
| Trust in whoever edits the audience | Not needed; edits are moves the rules check | Required. A custodian owner can list themselves [R-ran] |
| Audience reusable across patterns | No. The audience must be inside this module's witnessed state | Yes. `Members(L, S)` is a principal any clause can name |
| Recipient must be looking | Yes, instances are demand-paced [R] | No |
| Rulings needed | One atom field, one §8.7 subsection [R] | Six items in `SEAT-VIEWS-FINDINGS.md`, instance binding first [R] |

Two rows matter more than the others.

**Binding time.** The list is a late-bound audience: the label says "whoever
this cell names when you ask". The viewer stamp is an early-bound audience: the
label says "q, forever". Everything else follows from that.

**Whether the editor of the audience is entitled to the data.** The list
release is sound only when the list's writer may be trusted with the content,
because a declared writer names a handler and not a run, so the subject's owner
can bind that handler elsewhere and list themselves [R-ran:
`SEAT-VIEWS-FINDINGS.md`, worry c, "against a second binding of the list's
writer"]. Viewer-bound release has no such editor.

## The hypothesis, tested

The hypothesis given: list fits "audience curated by someone entitled to the
data, same content for all"; viewer fits "content is a function of the reader,
decided by sealed state, custodian not entitled".

**Verdict: right at the two poles, but it bundles three independent axes, and
the cases separate them.** [I]

1. *Uniform or per-reader content.* This alone decides whether list release is
   workable. Per-reader content needs one list, one key and one stored view per
   recipient, with the count fixed in the types [R: finding 4].
2. *Is the audience's editor entitled to the content.* This alone decides
   whether list release is sound today.
3. *Must the grant be late-bound.* Does revocation have to reach values already
   derived, and must one value be shared and reused by other code and other
   patterns.

Cases where the bundled hypothesis mispredicts:

- **Case 8b, my row in a table the owner maintains** (grades, payroll). The
  owner is entitled, so the hypothesis says list. Content is per reader, so
  list is unworkable. Viewer-bound, or data-derived row labels, fits.
- **Case 9, moderators see more.** Same content for all moderators, curated by
  an entitled owner, so the hypothesis says list. But the contributors who
  write moderator-only content are not the subject's sole owner, and the
  capture check refuses their authored `Members` clause [R: specs#59, capture
  item 1]. It needs a sealed policy state either way.
- **Case 10, a leaderboard.** The owner may well be entitled to every score.
  Viewer-bound still fits, because content is per reader.
- **Cases 3 (tally), 7 (direct message), and the conclave** fit neither pole.

A sharper statement of the distinction: **list release names one audience per
value, by reference, and trusts its editor; viewer-bound release names one
principal per instance, by computation, and trusts its function.**

## The cases

Each case states the facts asked for, then each mechanism, then a verdict.
"Sound" means under the given threat model: the serving runtime is trusted,
clients are not, and delivery checks labels.

### 1. Card table with hidden hands, owner as custodian

- **Owner:** the host, sole `OWNER` of the table's space. **Entitled:** each
  player to their own hand; everyone to counts and phase; nobody to the deck.
  **Adversaries:** other players, watchers, and the host.
- **Audience changes:** yes, per seat, during seating. **Revocation:** weak
  need; a player who leaves has seen their hand. **Audience secret:** not in
  poker, but could be in other games. **Recipients:** 2 to 12. **Content:**
  different per recipient.

**Viewer-bound** [R]. One sealed record, one handler `move`, one lift
`viewFor(state)`. Stored: the record, and one instance of the view per
looker. A move recomputes k views for k lookers. The host's own instance holds
no hand; a player's instance is refused to the host [R-ran]. Leaving a seat
empties the instance at its next computation [R-ran]. The label of q's
instance names q, in a document already keyed by q.

**List** [R]. Two handlers, a public roster and phase, a sealed key and an
occupant list per seat, seat count fixed in the types. A move recomputes n
views whoever is looking. **Not sound against the host** [R-ran]. Seating
cannot be secret: "a game where seating is secret has no shape that runs" [R].

**Verdict: viewer-bound.** One caveat [I]: `move` shuffles with `Math.random`
in the handler [R: `card-table.tsx` line 303]. With server execution off, the
dealing client computes the deck. The custodian guarantee needs the handler to
run on the serving runtime only. That is true of both mechanisms.

### 2. Sealed-bid auction

- **Owner:** the auctioneer, a custodian. **Entitled:** each bidder to their
  own bid; after close, whatever the rules publish (winner, price, or for
  second-price, the winner learns the second bid). **Adversaries:** bidders,
  the auctioneer (who may also be the seller or a shill).
- **Audience changes:** bidders join until close. **Revocation:** not needed.
  **Audience secret:** often yes, who bid is itself sensitive. **Recipients:**
  unbounded. **Content:** per recipient before close; uniform after.

**Viewer-bound** [I]. `bidView(state)` returns the viewer's own bid and, once
`state.phase` is closed, their outcome. The public result is a second blessed
function released by `dropClause`, as `blessed-computation.tsx` and the
inductor pool release a tally [R], returning nothing before close. Who bid
stays inside the sealed record.

**List** [I]. One list per bidder. Finding 4 says the pointer is in the schema,
"an auction cannot have one list per bidder" [R]. Not expressible, and unsound
against the auctioneer if it were.

**Verdict: viewer-bound for the per-bidder part, an ordinary public release
for the result.** Open [I]: closing on a clock needs a time the pattern can
trust. Loom's Questions design records that patterns cannot read the server's
time today [R]. Closing by the auctioneer's move is expressible now.

### 3. Secret ballot with a public tally

- **Owner:** the election's organizer, a custodian. **Entitled:** each voter to
  their own ballot; everyone to the tally; nobody to another's ballot.
  **Adversaries:** the organizer, voters, observers of the space.
- **Audience changes:** the roll may close. **Revocation:** not needed.
  **Audience secret:** who voted is often public, how is not. **Recipients:**
  the tally to all; a receipt to one. **Content:** tally uniform, receipt per
  voter.

**The tally is neither mechanism** [I]. It is `TransformedBy{tally}` with
`dropClause`, with the writer as input witness. That exists [R].

**The receipt, viewer-bound** [I]. `myBallot(state)` returns the viewer's own
recorded vote. **The receipt, list:** one list per voter, same failure as the
auction.

**What exists today instead** [R]. Loom's Questions writes the sheet twice from
one event: once sealed, by `submitSheet`, and once as a private copy by
`recordMine` "a separate transaction", because one flow label covers every
document a transaction writes. The two copies can disagree. A viewer-bound
readback is the authoritative one.

**Verdict: both parts are needed and neither is list release.** The leak that
neither mechanism touches [R: inductor pool README, cut 2]: a live tally moves
with the commit that adds a ballot, and that commit is signed by the voter. An
observer of the space links voter to vote by timing. A real ballot needs the
tally released at close, or batched.

### 4. Secret-santa assignment

- **Owner:** the organizer, who is also a participant, so a custodian with a
  motive. **Entitled:** each participant to their own giftee. **Adversaries:**
  every participant, the organizer most of all.
- **Audience changes:** before the draw, yes. After, a change is a redraw.
  **Revocation:** not needed. **Audience secret:** the participant set is
  known; the mapping is the secret. **Recipients:** n. **Content:** per
  recipient.

**Viewer-bound** [I]. `myGiftee(state)` returns `state.assignment[viewer]`.
The draw is a move. The organizer's instance holds only the organizer's own
giftee.

**List** [I]. n lists and n keys, and the organizer reads any of them through
the second binding.

**Verdict: viewer-bound.** This is the cleanest instance of the custodian
case. Same caveat as the card table: the draw's randomness must be the serving
runtime's.

### 5. Find-my-friends, location shared to a curated group

- **Owner:** the person whose location it is, in their own space.
  **Entitled:** the owner, and whoever the owner lists. **Adversaries:**
  ex-friends, other members who should not learn the roster, code that reads
  the fix.
- **Audience changes:** often. **Revocation:** must bite, promptly; this is a
  safety property. **Audience secret:** often, from other members.
  **Recipients:** a handful. **Content:** the same for all, or two tiers
  (precise and coarse), each uniform.

**List** [R]. This is the case specs#59 was written for. One list, one rule,
one policy digest. Removal revokes at the next access, and a member's own
derivation (an arrival estimate) stays visible exactly while they stay listed.
Each entry is labeled for its own principal, so a member learns their own
membership and not the roster. Two tiers are two lists and two lifts. The
owner being able to list themselves is harmless: it is their data.

**Viewer-bound** [I]. Expressible: the friend list goes into the sealed state,
edited by a handler, and `fixFor(state)` returns the fix when the viewer is in
it. It hides the roster completely. It costs a copy per looker, and it turns
downstream consumers per-user: Loom's "Arriving" view, which consumes Find My
[R: specs#59 comment], would compute once per viewer instead of once. A removed
friend keeps `User(q)` on anything already derived.

**What exists today** [R]. `cf-location-findmy.tsx` writes
`cfcPattern.user("__FINDMY_VIEWER_DID__")` under `TODO(#7444): audience from a
cell`, and a daemon splices the DID into the program text, so there is one
policy digest per viewer.

**Verdict: list.** The hypothesis holds exactly here. One nuance on revocation
[I]: against an untrusted client, revocation of either kind protects only what
has not been delivered. A viewer instance is computed on demand, so little of
it is ever undelivered. The real gap between the two is derived values held on
the server and copies a handler wrote elsewhere, which the list revokes and the
viewer stamp does not.

### 6. A shared document visible to a team whose membership changes

- **Owner:** the team or its lead. **Entitled:** current members.
  **Adversaries:** ex-members, outsiders. **Audience changes:** yes.
  **Revocation:** must bite. **Audience secret:** rarely. **Recipients:** tens
  to hundreds. **Content:** the same for all, and members also write it.

**Neither mechanism, first** [I]. If the team is the space, the space's access
list is the audience: `HasRole(reader, THIS_POLICY.subject)` adding
`User(reader)` at the boundary, which `direct-release.tsx` and the redactor
already do [R].

**List** [I]. Right when the team is a subset of a larger space, or is a group
that several documents and patterns share. `Members(L, S)` is a principal any
clause can name, so one curated group serves many values.

**Viewer-bound** [I]. Wrong. It computes one copy per member of identical
content, freezes `User(q)` on every derivation, and cannot reuse a group that
lives outside the module's own witnessed state.

**Verdict: space membership, then list. Not viewer-bound.**

### 7. A direct message, or a per-recipient notification

- **Owner:** depends on where it is stored. **Entitled:** sender and recipient.
  **Adversaries:** everyone else in the space; for an anonymous notification,
  also the recipient learning the sender. **Audience changes:** no.
  **Revocation:** unsend, sometimes. **Audience secret:** sometimes (that A
  messaged B). **Recipients:** one. **Content:** one value, one named reader.

**Neither mechanism is the natural one** [I]. The content is the sender's own
and the reader is known when it is written, so this is §3.1.8's "choosing the
audience of one's own content": a clause `[User(sender) ∨ User(recipient)]`
with the recipient pinned at the gesture (specs#59 Q3 B). Labs cannot write
that today: "There is no label a pattern can write that names another
principal" [R: finding 3]. What exists is the mailbox: a per-identity inbox
space whose label the recipient bound with `CurrentPrincipal` when they created
it, into which senders deliver [R: `private-inbox.tsx`].

**List** [I]. A list of two works and gives unsend, at the cost of a list per
conversation. Reasonable for a group thread; heavy for one message.

**Viewer-bound** [I]. Works as an inbox view: a sealed log and
`myMessages(state)` filtering by the viewer. It hides who messaged whom
completely, which the others do not. It recomputes every looker's view on every
append, over the whole log, which is the wrong cost curve for a mailbox.

**Verdict: an authored recipient clause (the unbuilt gesture helper), or the
existing mailbox. Viewer-bound only when the existence of the message must be
hidden**, for example "you have been drawn" in case 4.

### 8. "My own view" of shared data: my vote, my balance, my row

Two different cases.

**8a. Data I wrote.** Owner: me. No release is needed. A `PerUser` cell labeled
`User(CurrentPrincipal)`, as `ReaderPrivate` in recommend-a-book and
`OwnerPrivate` in the private inbox [R]. It stops working the moment an
aggregate must read everyone's value: a computation runs per viewer over the
viewer's own instance. Then the value moves into a sealed shared record and the
readback is case 3's receipt, viewer-bound.

**8b. My row in a table someone else maintains** (a grade, a balance).

- **Owner:** the maintainer, who is entitled to every row. **Entitled:** each
  person to their own row. **Adversaries:** other people in the table.
  **Audience changes:** rows are added. **Revocation:** rarely.
  **Recipients:** n. **Content:** per recipient.

**Viewer-bound** [I]: `myRow(state)`. Fits, though the owner is entitled.

**List** [I]: sound here, since the owner may read everything, and still
unworkable: n lists.

**A third mechanism already in labs** [R]: `cfc-row-label-records` derives each
row's label from the row's own column (`principal("mailto", match(f.patient_email, …))`
joined with the database owner). That is the proposal's option C, a data-named
audience, and it is sound here for the reason C is not sound for the card
table: the label's author is entitled to the data.

**What exists today in ordinary patterns** [R]: `fair-share` has
`myName: PerUser<string>`, "each viewer picks which person they are". Identity
is self-asserted because a computation cannot ask who it runs for.
`currentPrincipal()` in a computation has value there apart from any release.

**Verdict: 8a needs nothing new. 8b is viewer-bound or row labels, and it is
the clearest counterexample to "custodian not entitled" as the deciding
property.**

### 9. A moderated room where moderators see more

- **Owner:** the room's owner. **Entitled:** moderators to reports, flags and
  removed content; members to the rest. **Adversaries:** members, ex-moderators.
  **Audience changes:** moderators come and go. **Revocation:** should bite.
  **Audience secret:** the moderator list sometimes is. **Recipients:** a few
  moderators. **Content:** uniform across moderators; different from members'.

**List** [I]. Natural for content the owner authors. But a report is written
by a member, who is not the subject's sole owner, so the capture check refuses
a `Members(mods)` clause they author [R: capture item 1]. The report has to go
into a sealed policy state, and the moderator view becomes a blessed lift whose
output the policy releases to `Members(mods)`, with the clause carried from a
key the owner's setup wrote. That is the seat-views shape with one list. It is
sound, because the owner is entitled.

**Viewer-bound** [I]. `roomView(state)` returns more when the viewer is in
`state.moderators`. Trivial to write. A demoted moderator keeps what they
derived.

**Verdict: both are defensible; this is where composition earns its place.**
A member's own-view (their reports, their strikes) is viewer-bound; the
moderator tier is list-released. One policy, two rules.

### 10. An audit or leaderboard that shows each person only their rank

- **Owner:** the organizer. **Entitled:** each person to their own rank.
  **Adversaries:** other participants. **Audience changes:** people join.
  **Revocation:** no. **Audience secret:** participation may be.
  **Recipients:** n, unbounded. **Content:** per recipient.

**Viewer-bound** [I]: `myRank(state)`. A non-participant who looks gets "not
ranked". **List** [I]: n lists; not expressible.

**Verdict: viewer-bound.** What leaks is in the function, not the label: a
rank is a statement about everyone else's scores, and a participant with
several identities can difference their ranks. No release mechanism fixes a
function that says too much.

### 11. Existing patterns that already struggle to name an audience

All [R] unless marked.

| Where | The workaround | What it wants |
| --- | --- | --- |
| Loom `cf-location-findmy.tsx` | Literal `User("__FINDMY_VIEWER_DID__")`, spliced into program text per viewer; `TODO(#7444): audience from a cell` | List |
| Loom `questions-policy.ts` | `releaseCard` is `dropClause` on `comparePair`'s output: the card is released to no one in particular because nothing can name "the viewer and the owner" | Viewer-bound, on both policy clauses [I] |
| Loom `questions-policy.ts` | Identity-only rule, because a witnessed one refuses the honest card; the cost is the pinned "crafted sheet" attack. The file names the fix: "a wish result the runtime attests as the RENDERING principal's own element" | A viewer-attested *input*, which neither mechanism provides |
| Loom Questions design §4 item 5 | Private copy written by a second transaction; per-viewer documents stored in the owner's space "labelled for the viewer alone", relying on a wildcard grant | Viewer-bound readback [I] |
| Labs `direct-release.tsx`, `apps/redactor/policy.tsx` | `HasRole(reader, subject)` adds `User(reader)`. Not a workaround: it is the existing rule that binds `User` from a runtime-minted fact | Nothing; it is the precedent for the viewer rule |
| Labs `system/private-inbox` | A space per identity as a mailbox; `WRITE` to all; the label "does not stop that code reading them" | Delivery by label (specs#95), then an authored recipient clause |
| Labs `cfc-row-label-mailbox`, `cfc-row-label-records` | Labels derived from row data | Nothing; a third mechanism that already works where the owner is entitled |
| Labs `examples/fair-share` | `myName: PerUser<string>`, self-asserted identity | `currentPrincipal()` in a computation |
| Labs `custody-answer-room.tsx`, conclave | Sealed stances, released to the seal's own reader, declassified once into a create-only slot | Neither. It is consent release with an instance-named witness |
| Labs `showcase/inductor/pool` | Public release by `dropClause`; contributor linkable by commit signature | Neither; a public release with a timing leak |
| FabriChat (specs#90) | A manager result in a home space whose membership entries carry ten other principals' clauses, which route 2c will not release | A warning for viewer-bound: literal principals accumulate in labels |

The last row matters for the recommendation. Specs#90 exists because literal
principal clauses accumulated on a document that "holds none of their data"
and nobody could remove them. Viewer-bound release manufactures literal
`User(q)` clauses and D5 carries them onto derived values.

## Cases that fit neither, or need both

**Fit neither** [I]:

- *A public aggregate* (tally, pool consensus, auction result). `dropClause`
  under `TransformedBy` with a writer witness. Exists.
- *A direct message.* An authored recipient clause, pinned at a gesture. Not
  built.
- *Multi-party consent with a one-time answer* (conclave). The custody seal.
- *A two-party comparison across two spaces* (Questions' `comparePair`). The
  output is a function of the viewer and the owner, over sealed data in two
  spaces. The release half is viewer-bound. The input half, proving that the
  sheet fed in is the viewer's own, is a different need: a viewer-bound
  *input* attestation.
- *Row-level audiences in a table whose owner is entitled.* Data-derived row
  labels. Exists for SQLite.

**Need both, composed** [I]. A per-viewer view that should also require
current membership: a hand that an ejected player must stop receiving, a
moderator's own queue, a paid subscriber's own page.

The composition I would propose is **grant by viewer, restrict by membership**,
as two conjunctive clauses:

```
[ Policy ∨ User(viewer) ]  ∧  [ User(owner) ∨ Members(L, S) ]
```

- The first clause is the viewer rule's result.
- The second is not added by any rule. It is carried in from the sealed
  state's own label: the owner's setup move authors it once, under specs#59's
  capture check, and every later value inherits it by the join. Specs#59 already
  says "Intersection is free. Two `Members` clauses are conjunctive", and
  `SEAT-VIEWS-FINDINGS.md` reports that a carried `Members` clause needs no
  capture in the run that recomputes [R-ran].
- A reader must be the viewer *and* currently listed. Removal from L revokes
  everything, derivations included, at the next access.
- **The second-binding attack becomes harmless.** An owner who lists
  themselves in L satisfies the second clause and still fails the first for
  every instance but their own. Used this way the list can only restrict. So
  this composition does not wait on instance binding.

Three cheaper variants:

- **`Space(S)` as the conjunct.** If stored data carries its space as a clause
  (§3.6.1 [R]; specs#95's comment reads it that way [R]) and delivery resolves
  it, then removal from the space's access list already revokes every viewer
  instance in it. Whether labs labels carry that clause today I did not check.
- **Recomputation.** A membership change inside the sealed state is a state
  change, so a looker's instance recomputes and holds less [R-ran for the card
  table]. This is the default behavior and it is not revocation: derived values
  keep `User(q)`.
- **Expiry.** An `Expires` atom (§15 [R]) on the view would bound how long a
  stale instance lives. I did not check whether a module rule may add one.

What does **not** work as a composition [I]: adding a `HasRole` or `ListedIn`
precondition to the viewer rule. That makes the rule boundary-scoped under
§5.3, so its result is "never carried to derived values", and a viewer's own
derivations go blank. That is option D of specs#59, rejected there for this
reason.

## Coherence with the rest of the specification

### The acting principal, §8.15.12 and specs#50

**The viewer is not a new principal. It is the run's acting principal.** [I]

- Specs#50, SC-38 option A, says a run's actor on a serving host is "its
  trusted event record's stamp, the demanded instance's principal, or the
  delegation's actor" [R]. "The demanded instance's principal" is the viewer.
- The implementation already agrees: the API is `currentPrincipal()`, the same
  call a handler makes, and on a serving runtime it reads "the principal whose
  demand supplied the run's identity" [R].
- SC-53's reading carries over unchanged: the claim attests that a run acting
  for the principal wrote the value, not that the principal asked for it.
- SC-53 refuses a current-principal claim at a position that declares no
  writer, "since any code could then mint it" [R]. The viewer stamp answers
  that objection differently and correctly: it sits inside `TransformedBy`,
  which names the code, and the rule pins the code.
- An actor-less run mints no viewer atom, so nothing is released. That matches
  SC-53's fail-closed direction.

One real asymmetry: the proposal mints the stamp for computations and not for
handlers, to avoid publishing who triggered each shared write [R]. Under
specs#50 both kinds of run have an actor. The text should say the field is the
acting principal and that only a reactive computation's stamp carries it, and
give that reason, so that it reads as a restriction and not as a second
concept.

### The access facts of §4.9

Coherent, and it is worth saying how [I]. §4.9's facts (`$actingUser`,
`HasRole`) describe the reader at this access, and §5.3 classes a rule guarded
by them as boundary-scoped. The viewer stamp describes the producer at the time
of production. `direct-release.tsx` shows the existing access-side rule:
runtime-minted `HasRole(reader, …)` binds `reader`, and the postcondition adds
`User(reader)` [R]. The viewer rule is the production-side twin of that rule.
Nothing in §4.9 has to change, but §4.9.1 defines the acting context as "the
authenticated principal making the request", which reads as boundary-only;
specs#50 already extends it to runs.

### §5.3 and D5

The rule meets the letter of value-intrinsic: no sink or path, no grant guard,
and an integrity precondition "bound to the exact current value and present in
the consumed label" [R]. It also meets the spirit [I]: the principal is fixed
when the value is produced, so "this value is f(state, q)" is a property of the
value, like "coarsened from the same measurement".

One open tension, which the proposal itself names [R]: §5.3 says the
satisfying integrity "is recorded as a witness on the derived value's
`TransformedBy`", and the proposal drops the viewer-bearing atom from retained
witnesses. Its reason is that a nested `viewer` could be matched by a rule
pinning two levels. A cleaner resolution [I]: keep the atom as a witness, where
it is a true statement ("an input was computed for q"), and have template
validation refuse a `User` postcondition bound from anywhere but the top-level
atom. Specs#59 already restricts `Members` postconditions by template
validation in the same way [R]. That keeps D5's auditability and removes the
exception.

### "Scope keys instances; it never keys authority"

The mechanism is coherent with this; one sentence of the proposed text is not
[I]. The proposed §18 text defines the viewer as "the principal whose scope
instance the run reads and writes" [R]. That defines an authority-relevant fact
by scope. The order should be the other way, as `scopes.md` has it (M1: "scope
is discovered by running AS (principal, session)" [R]): the run has an acting
principal; reading it is a user-scoped read; the output therefore narrows to
`user`. The proposal's own section "Scope is not the protection" says the right
thing, and specs#95 says the same from the other side: a store kept per
principal "is not for that reason confidential to that principal" [R].

A point to check [I]: `scopes.md` says an undemanded node that narrows "writes
the service identity's inert instance" [R], while the proposal says the runtime
"MUST NOT answer with its own service identity" [R]. These are compatible only
if that run is answered with no principal and still narrows. Worth a test.

### A cleaner statement

> §15: `TransformedBy` gains `actor?: DID`, the acting principal (§8.15.12) of
> the attempt that minted it, present only on the stamp of a reactive
> computation that observed its acting principal.
>
> §8.7: a module policy MAY release such an output with a rule whose
> precondition pins the computation's identity and a literal input witness and
> binds a variable from the top-level `actor`, and whose postcondition adds
> `User` of that variable.

This removes three objections [I]: there is no new kind of principal; the
definition does not pass through scope; and the witness exception to §5.3 goes
away if template validation carries the restriction. It keeps the proposal's
one-atom argument, which I think is correct: a separate `ComputedFor{q}` atom
could pair one attempt's code with another attempt's principal where labels
join.

I looked for a restatement that needs no new field and did not find a sound
one [I]. A rule using `$actingUser` at the boundary releases to whoever reads.
A `User` clause supplied by the runtime on the output is the proposal's option
E, unsound in general. A principal passed as a parameter is steerable.

## Recommendation

**Both are needed. They are not alternatives, and neither subsumes the other.**

| Class of case | Mechanism |
| --- | --- |
| Everyone in the space | Space membership (`HasRole`), as today |
| My own data | A `CurrentPrincipal` label, as today |
| A public aggregate | `TransformedBy` with `dropClause`, as today |
| One audience, same content, curated by someone entitled to it, revocable, reusable (cases 5, 6, the moderator tier of 9) | **List release** |
| Content that depends on the reader, decided by sealed state (cases 1, 2, 3's receipt, 4, 8b, 10, Questions' card) | **Viewer-bound release** |
| Per-reader content that must also be revocable | Viewer grants, list or space restricts |
| One message to one named reader | An authored recipient clause (unbuilt); the mailbox meanwhile |

**Adopt viewer-bound release first** [I]. It needs one field on an existing
atom and no change to the rule calculus or any gate, and it unblocks the
custodian cases, which nothing else can do soundly today. State it as the
acting principal of §8.15.12, not as a new "viewer".

**Adopt list release for the entitled-curator class, and say so in the text.**
The proposal's suggested sentence for §8.7.5 is right. With that scope stated,
the list release does not have to wait on instance binding: the second-binding
attack only hands the owner their own data. The other five items in
`SEAT-VIEWS-FINDINGS.md` exist mostly because the card table pushed the list
into a per-recipient shape it was not designed for. Find My needs the
imported-policy answer and the capture position, and little else.

**The minimal coherent combined design:**

1. `TransformedBy.actor`, minted for a computation that observed its acting
   principal, and the §8.7 rule shape above.
2. `Members`, `ListedIn` and the capture check as specs#59 has them, scoped by
   text to an audience its owner curates over data the owner is entitled to.
3. One composition rule, stated once: a release granted by `actor` may be
   narrowed by any conjunctive clause the sealed state carries, `Members` or
   `Space`; a list is never the grant in a custodian case.
4. Template validation: a `User` postcondition binds only from `HasRole` at a
   boundary or from a top-level `actor`.

**The single most important open risk** [I]: **viewer-bound release moves the
audience decision out of the label and into ordinary code, runs it once, and
freezes the answer as a literal grant.**

- The label `[Policy ∨ User(q)]` does not say q was entitled. It says the
  function was computed for q. Any principal who can demand an instance gets
  one, so every blessed viewer function is a declassifier to all comers and
  must be safe for strangers, for every reachable state, and for every public
  input. A mistake produces a perfectly valid label, and no gate can notice.
- The grant is early-bound. D5 carries `User(q)` onto everything derived, and
  nothing at the label level takes it back. Specs#90 shows what accumulated
  literal principals cost later.

The mitigation is item 3: a conjunctive `Space` or `Members` clause puts a
label-level floor under who can ever be a viewer and makes the grant
revocable. I would not ship viewer-bound release for anything beyond a game
without it.

A second risk, which both mechanisms share and which I rank below the first
only because it is already known and filed: nothing here holds until delivery
is a boundary (specs#95; labs documents itself as non-conforming [R]), and
until the handlers that write sealed state run only on the serving runtime.
Today a client replica holds the sealed record.
