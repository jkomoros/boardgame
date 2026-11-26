# An audience named by a cell, for the custodian-owner case

A design study, written to be compared with "viewer-bound release"
(`patterns/examples/cfc-exchange-rules/MODULE-NAMED-AUDIENCE-PROPOSAL.md` in the
`viewer-bound-rebased` worktree). Nothing here is built, ruled, filed or posted.
No repository file was changed.

Every claim is tagged with how it is known:

- **[spec]** read in `commonfabric/specs` at `main` (`f549c796`), or in an open
  specs pull request, which is named.
- **[code]** read in labs code or labs documents, on `main` (`8c6f09c5fc`) or in
  the worktree, which is named.
- **[design]** my design or my inference. None of it ran.

## 1. The short answer

- **The best audience-named-by-data design is a "co-written audience":** the
  blessed output holds its own audience at a path, the module rule releases the
  output to `Members` of *that path of the same stamped document*, and the
  runtime resolves the path live at each display and each delivery. No
  principal is written into any stored label, `TransformedBy` gains no field,
  and scope carries no evidence. [design]
- **It can be made to work by argument, and it is not close to working in
  labs.** It stacks on specs#59, which is open and unruled, and adds three
  things to it. Viewer-bound release is built behind a flag with passing tests.
  [spec] [code] [design]
- **Every other by-reference design I tried reduces to this one or fails on the
  same point.** A parameterized projection has no trustworthy link from "this
  value" to "that audience record" other than the two being written by one
  attempt. Section 3 shows this option by option. [design]
- **My choice for the custodian-owner case is the co-written audience in its
  per-viewer layout (section 4.2), as the specification direction,** with
  viewer-bound release kept behind its flag as the working stopgap. The two are
  rivals on one question only, where the principal comes from. They share
  everything else, and the pattern an author writes differs by one field and one
  line of the rule. [design]
- **If specs#59 is not going to be ruled, viewer-bound release is the right
  answer,** and section 6.3 says what it should borrow. [design]

## 2. What any design here has to get right

The label is the whole protection (the task's premise, and specs#95's "Delivery
is a boundary": "A deployment delivers a stored value only to a principal who
satisfies its label" [spec, open PR #95, the maintainer's takeover comment]).
So a design is a stored label on the view, plus facts the runtime mints at the
boundary for the acting principal, such that the label fits exactly the intended
principal.

There are two families.

- **Literal.** The label, after exchange, holds `User(q)`. Viewer-bound release
  is this.
- **By reference.** The label, after exchange, holds an atom naming a record,
  and the runtime mints a fact for the acting principal by a point query against
  that record. §4.9.3's `Space`/`HasRole` and specs#59's `Members`/`ListedIn`
  are this. [spec]

A by-reference design in the custodian case owes four things. I use these
letters throughout.

- **(S1) The record is authentic.** Its content is a function of the sealed
  state as the module's own writer wrote it. The owner, and a second binding of
  any handler, cannot produce an accepted record with other content.
- **(S2) The value is bound to the right record.** A real hand cannot be paired
  with an audience record of the attacker's choosing.
- **(S3) The membership bit is released by the policy.** "Am I listed" is a fact
  derived from sealed state, so the policy that seals the state has to be what
  authorizes telling a principal about themselves.
- **(S4) Value and record move together.** A stale hand must not meet a fresh
  audience.

The list release of specs#59, as tried in `SEAT-VIEWS-FINDINGS.md`, fails S1
(worry c, the second binding, pinned by a test) and has no answer to S3 (finding
1, an entry the sealed writer wrote lists nobody). [code, worktree]

## 3. The options explored

### 3.1 A computed audience cell, separate from the view

**Shape.** `audienceOf(state, i)` is a blessed lift that writes seat `i`'s
occupant list to a cell. `seatView(state, i)` writes the view, whose label names
that cell through specs#59's `members` pointer. Resolution is live. [design]

**What it needs.**

- An entry rule for S3. Entries carry the policy clause, and specs#59 counts an
  entry only if "its own label admits `P` at this boundary", judged without
  exchange. specs#59 defers exactly this: "Computed lists. A list a trusted
  computation writes carries its inputs' labels in its entries … Decide whether
  such a list needs its own release". [spec, open PR #59]
- A binding for S2, and this is where it fails. The owner starts a second run in
  which `seatView` reads the live state and `audienceOf` reads a state of the
  owner's own, written by `move` bound elsewhere, in which the owner holds seat
  1. Both outputs carry honest stamps. specs#59's capture check asks that "the
  output and the list position are both rooted in the result of one run of the
  module", which the second run satisfies. [spec, open PR #59] [design]
- The capture check's first item needs an acting principal who solely owns the
  subject. A served derivation "carries its demander's snapshot for its reads
  but acts for no one" [code, `docs/features/current-principal.md`], and a
  player's move, not the owner's, triggers most recomputations.

**Closing S2** needs the two stamps to name the same input. The spec's
`TransformedBy` has `inputs: Array<{ ref, witnesses }>` [spec, §8.7.1, §15], so
a rule could in principle compare references. Labs chose the summary form and
records why: an address list "would put every read path — data-dependent, and so
itself a channel — into the label", and the rule calculus has no quantifier over
inputs [code, `docs/specs/cfc-transformed-by-input-witnesses.md`,
"Representation choices"]. I would not reopen that.

**Verdict.** Not sound without a same-input binding. The cheapest same-input
binding is "the same attempt wrote both", which is option 3.6.

### 3.2 A seat that holds a profile cell

**Shape.** A seat holds a reference to the occupant's profile cell. The
declassifier resolves the reference to the one principal the cell's label
attests through `represents-principal`. This is the maintainer's remark of
2026-10-01 as quoted in specs#59's body: "you pass in the profile cell and then
something needs to work out this is actually this user and declassify". [spec,
open PR #59]

**What exists.** The custody seal's seats are this already: "a seat may be a
reference to a cell whose stored label attests one principal", read in the
strict form a runtime mints, and "a pattern holds a member's profile, never the
member's DID" [code, `docs/specs/cfc-custody-seal.md`, "Seats"]. `principalOf()`
gives pattern code the same read [code, `docs/features/principal-of.md`].

**What it is not.** It is an entry kind, not a release mechanism. It says how an
audience record names a person. It says nothing about S1, S2 or S4, so on its
own it has the binding problem of 3.1. specs#59's Q3 also chose against
resolving a profile at every evaluation ("a profile whose claims change …
redirects every resolution") and pins the principal when the entry is written.
[spec, open PR #59]

**Verdict.** Keep it as the entry form of whichever mechanism is chosen: an
entry is `{ principal }`, pinned by blessed code with `principalOf(profile)`, or
`{ profile }`, resolved by the trusted resolver in the seal's strict form. It
fits 3.6 without change. It gives no revocation of its own, since a runtime
refuses to rebind an attested principal. [design]

### 3.3 An instance-bound writer for lists

**Shape.** Keep specs#59 and make a declared writer mean "this handler as bound
in the run that owns this document". The custody seal gets the same effect by a
host-derived address for the instance `(P, D)` and by writing the link itself
[code, `docs/specs/cfc-custody-seal.md`]. I did not find a conclave pattern in
labs; `custody-answer-room.tsx`, `custody-projector.tsx` and
`patterns/showcase/inductor/lofi/` are what I read.

**What it would specify.** §8.15.1 and §8.15.6 gain an instance qualifier on a
write-authority entry, checked at commit against the run that owns the target
document. §8.15.8 says the opposite today, on purpose: "the same verified code
carries its authority into any pattern that loads it" [spec]. The verified
binding identity in labs already carries a `bindingPath` [code,
`runtime/runner/src/cfc/implementation-identity.ts`], so the runtime half is
plausible.

**What it leaves.** It closes worry (c) for the list. It leaves every other
contortion in `SEAT-VIEWS-FINDINGS.md`: the second handler, the public roster
(finding 2), self-seating only (finding 3), a list per seat fixed in the types
(finding 4), the capture check's owner position (finding 7). Seating stays
public, so a secret-santa draw has no shape. [code, worktree]

**Verdict.** Worth ruling for every custodian list an owner curates. Not an
answer for an audience that is a fact inside sealed state. It complements 3.6
and does not replace it.

### 3.4 Seats as roles: an access fact attested by sealed state

**Shape.** Extend §4.9.3 so the runtime mints `HasRole(P, table, "seat/1")` from
a roster the module's writer attests, where today it mints "from **verified
space membership**" only [spec]. The nearest existing hook is §4.3.5: a grant
record is "a durable *input* to exchange evaluation … generalizing the space ACL
document", consulted by `guard.policyState`. [spec]

**Where it fails.**

- A grant is "written only by a trusted policy writer", and "the writer MUST
  verify the granting principal's release authority" [spec, §4.3.5]. A module's
  `move` is not that identity, and in the custodian case no principal has the
  release authority.
- A grant-guarded result "is never carried to derived values" [spec, §5.3], and
  a derived value has lost the projection's `TransformedBy`, so a rule that
  still requires it goes dark and one that does not would release the raw
  sealed state. This is the cost specs#59 records against its own option C.
- The fact says "`P` holds seat 1". Nothing trustworthy says "this value is seat
  1's view". The seat index is a parameter of the projection, and a public
  parameter "does not constrain the witness" [code,
  `cfc-transformed-by-input-witnesses.md`, "Public parameters"]. So S2 is
  unmet, again.

**Verdict.** Does not stand alone. The fact a custodian release needs is not
"`P` has role `r`" but "`P` is named by this value's own audience", which is
3.6.

### 3.5 Field-level labels and label-checked delivery, with no projection

This is the maintainer's remark on labs#3750: "we have field-level
opaque/not-opaque. why is that not sufficient?" [code, labs#3750 review
comments, 2026-05]. The author's reply there conceded it for multi-field
secrecy and kept one exception, a coarser value derived from one field (a
hand's count).

**Shape.** `move` stores each hand in a field or document labeled for its
occupant. Delivery by label does the rest.

**What blocks it, in order.**

1. **A label that names a principal chosen by data.** A pattern "cannot write a
   label that names another principal" [code, worktree, finding 3]. The spec
   does have ruled precedent for a reader set computed from a value's own
   content: §3.1.8's "data-derived reader sets", worked in §13.12 for SQLite
   rows, where "the label is a pure function of stored content" [spec]. So this
   block is a missing cell-side feature, not a missing principle.
2. **The conjunctive-join rule.** "All confidentiality carried by the inputs …
   joins conjunctively with the authored clause" [spec, §3.1.8 rule 1]. A hand
   dealt from a sealed deck is `[Policy] ∧ [User(occupant)]`, which nobody
   satisfies. Something must still add an alternative to the policy clause, on
   evidence. That is an exchange rule, and the question of what it binds the
   principal from is the original question.
3. **The audience of a card changes.** In the deck nobody reads it, dealt its
   holder does, played everybody does. A stored label is not widened by an
   ordinary write [spec, §3.1.8 rule 4, §8.12.7]. So the field's label has to
   name its audience by reference, and we are back at a list in a cell written
   by the sealed writer, with S1 and S3 to answer.
4. **One confidentiality join per transaction.** A `move` that reads two hands
   to compare them taints everything it writes with both [code, worktree,
   "gap F"].

**Verdict.** The remark is right about what the mechanism should be, and it is
not a separate mechanism. With the release rule of 3.6, where the "projection"
may be the identity, `move` can write one self-describing document per seat and
no projection lift runs at all. That is layout 4.3 below. [design]

### 3.6 The output names its own audience (the design)

All four failures above are the same failure: the audience record and the value
are two things, and nothing binds them. Put the audience in the value. Section 4
specifies it.

### 3.7 Hybrids

- **Per-viewer computation with the audience resolved at each release.** This is
  layout 4.2, and it is what I recommend for the card table.
- **Viewer-bound release that adds a reference instead of `User(q)`.** Possible:
  the rule's postcondition could add `Members` of a one-entry cell the runtime
  writes beside the instance. It keeps the stamp field and the scope-as-source
  objection and adds a resolver, so it has the costs of both designs. I do not
  recommend it. [design]

## 4. The design: a co-written audience

### 4.1 Specification

**The idea.** A blessed computation's output is one document, written whole by
one attempt, and it holds its audience at a path. The module's rule releases the
document to the members of that path. "The list in that cell over there" is a
field of the cell being released. [design]

**The rule.** A new placeholder in the template grammar of §4.3.6, `SELF`,
legal only as the `list` of a `Members` postcondition:

```json
{
  "name": "releaseToOwnAudience",
  "preCondition": {
    "confidentiality": [{ "thisPolicy": true }],
    "integrity": [{
      "type": "https://commonfabric.org/cfc/atom/TransformedBy",
      "identity": { "symbol": "viewFor" },
      "inputWitness": { "type": ".../TransformedBy", "identity": { "symbol": "move" } }
    }]
  },
  "postCondition": {
    "confidentiality": [{
      "type": "https://commonfabric.org/cfc/atom/Members",
      "list": { "self": "/audience" },
      "subject": { "thisPolicyField": "subject" }
    }]
  }
}
```

**How `SELF` resolves.** When the rule is evaluated at a location, the integrity
precondition is satisfied by a flow stamp held by one label entry at some root
`r` at or above the location. `{ "self": "/audience" }` is the address `r` +
`/audience`. If no single stamped entry covers the location, the rule does not
fire. The pointer is in the policy template, which is digest-bound, and not in
the document's schema, so whoever stores the document cannot point it at another
field. [design]

**Why the rule is value-intrinsic.** Its only integrity precondition is "bound
to the exact current value and present in the consumed label" [spec, §5.3]. The
address it writes is a coordinate of the consumed label. So derived values carry
`[Policy ∨ Members(r/audience, S)]`, a reference and not a principal. [design,
needs ruling]

**Resolution.** specs#59's `ListedIn{principal, list}` and `ListMemberAccess`,
unchanged [spec, open PR #59], with one new clause in "which entries count":

> An entry also counts for `P` when it names `P` and the list is
> **self-attested**: evaluating the value-intrinsic rules at the list's own
> location yields `Members` of that same list as an alternative of every clause
> `P` does not otherwise satisfy.

In words, a list counts when the policy that seals it releases it to itself.
The check is one step and assumes membership in no other list. [design]

**What this answers.**

- **S1.** The list is accepted only while the stamp of the blessed code covers
  it. An overwrite of `/audience` by any other code makes a more specific label
  entry without that stamp, so the rule stops firing there and the list counts
  for nobody. A second binding of the blessed code elsewhere writes its own
  document, never this one.
- **S2.** The value and its list are under one stamp entry. There is no second
  document to substitute.
- **S3.** The membership bit is released by the same rule of the same policy.
  A principal learns only whether they are themselves listed. This is specs#59's
  deferred "computed lists" question, answered "yes, it needs its own release,
  and it is this rule".
- **S4.** One document, one transaction.

**Where facts are minted.** `ListedIn` at display boundaries, as specs#59 has
it, and also at the delivery boundary specs#95 adds. specs#59 says it is "never
minted at a sink-request, send, or model-input boundary", which stays. [spec,
open PRs] [design]

**No capture check.** specs#59's capture check guards an *authored* clause. This
design authors none: the only widening is the module's rule, on evidence. So it
needs no acting principal and no sole owner, and it works for a served
derivation that acts for no one. Template validation changes instead: a
`Members` postcondition is admitted in specs#59's form, or with `list: SELF` and
`subject: THIS_POLICY.subject`. [design]

**Sections that change.** §4.3.6 (the placeholder and its validation), §4.9.5
(the entry clause, the delivery mint), §8.7.5 (a second release shape, with no
capture), §5.3 (one sentence that a self address is part of the consumed label),
§18.4.5 (the watch already drafted covers it). §15 gains nothing beyond
specs#59's two atoms. `TransformedBy` is untouched. [design]

### 4.2 Layout A: one instance per viewer (recommended for the card table)

The pattern is `card-table.tsx` [code, worktree] with two edits.

```ts
// Shown for illustration only. Not compiled.
export const viewFor = lift((
  { state, me }: { state: Game | null; me: Profile },
): TableView => {
  const viewer = principalOf(me, "represents-principal");
  const game = gameOf(state);
  const seat = viewer === undefined ? -1 : game.occupants.indexOf(viewer);
  return {
    audience: viewer === undefined ? [] : [{ principal: viewer }],
    seat,
    hand: seat === -1 ? [] : game.hands[seat] ?? [],
    // ... the public fields, as today
  };
});
```

- **What the author writes.** One sealed record, one writer, one lift, one
  rule. The lift returns an `audience` field, and the rule names `/audience`
  where the viewer rule names `viewer`. Seats are a number in the state.
- **Where the viewer comes from.** An input: the viewer's own profile cell, or
  `currentPrincipal()` if that is allowed in a computation. Either is a public
  parameter. That is safe here, and it is the important difference from
  viewer-bound release: the audience is derived from the same parameter as the
  content, so a caller who passes Bob's profile computes Bob's view, released to
  Bob, and can read none of it. Soundness does not depend on which instance the
  run writes. Scope "keys INSTANCES; it NEVER keys authority" [code,
  `docs/specs/server-side-execution/scopes.md`] holds with no exception to
  explain.
- **Stored.** One instance per principal who looks, as in viewer-bound release.
  Every stored label is the policy clause and a stamp. No label names a
  principal. The audience entry is sealed data inside the instance.
- **Recomputed.** `k` views per move for `k` principals looking.
- **Leaks.** A derived value's label holds the instance's address, and a
  user-scoped address contains the scope key `user:<principal>` [code,
  `scopes.md`, "Instance addressing"]. So a derivation stored in a shared
  document would carry the viewer's identifier as an address. Viewer-bound
  release has the same exposure as a `User` atom. Seating is not visible:
  watchers have instances too.
- **Owner, second binding.** The owner's instance names the owner and holds no
  hand. A second run of the module over the live state computes the caller's
  own view. `move` bound elsewhere applies the same rules to the same record.
  These are viewer-bound release's arguments and they carry over unchanged,
  because both designs rest on the same author discipline: the blessed
  functions take the state and nothing that selects among secrets. [design]
- **Revocation.** Emptying the list revokes the instance and everything derived
  from it, at the next evaluation. But an instance is recomputed on its
  viewer's demand [code, `scopes.md` §2], so the list of a principal who
  stopped looking is not emptied. In practice this layout revokes little more
  than viewer-bound release does. It does make revocation *expressible*: a
  later ruling that recomputes demanded-once instances, or an entry that
  delegates to a shared list (specs#59's deferred "delegation"), turns it on
  without touching stored labels.

### 4.3 Layout B: one document per seat, written by `move`, and no projection

This is the field-level reading of 3.5.

- **Shape.** `move` writes a deck document with no audience, a table document
  whose audience is the space, and one seat document
  `{ audience: [occupant], hand }` each. The rule names `move` as the stamped
  code. Everything `move` writes is sealed, and each document says who may open
  it. That removes "the one writer can write nothing public" as an obstacle:
  the public facts are a sealed document released to the space. (A space entry
  is specs#59's deferred "space entries"; a second rule adding
  `Space(THIS_POLICY.subject)` under the table document's stamp avoids needing
  it.) [design]
- **Stored and recomputed.** `n` seat documents whoever is looking. A move
  rewrites the documents it changes, usually one seat and the table.
- **Revocation.** Real and immediate, since the documents are shared and
  written on every move. A leaver's list is emptied in the transaction that
  removes them.
- **The transfer hazard.** A value Alice derived from seat 1 carries
  `Members(seat1/audience)`. When Bob takes seat 1, he satisfies it. Releasing
  "to the seat" is transferable by construction. The fix is for `move` to
  create a fresh document for each occupancy and empty the old one, so an old
  reference is dead and never rebound. A reviewer has to check that.
- **The author discipline is harder.** With the state split across documents, a
  second binding of `move` can be handed the live seat documents and a state of
  the attacker's own. It is safe only if each seat document is the sole record
  of its own occupant and `move` changes a document's audience by rules that
  read that document and the actor alone. Layout A has one record and no such
  hazard.
- **Authoring.** There is no "my view". A pattern renders every seat and lets
  the gate withhold the others, and a computation over "all seats" is readable
  by nobody.
- **Leaks.** Labels are uniform and addresses are seat indexes, so who sits
  where is in no label and no address.

Layout B is the right shape for a ballot box (one receipt document per voter)
and for the maintainer's field-level model. For a card table I would use A.

### 4.4 Rulings this needs, none of which exists

1. specs#59's core: `Members`, `ListedIn`, `ListMemberAccess`, the §5.3 carry of
   a `Members` clause. Open since 2026-10-08, with no review from the maintainer
   on the thread. It stacks on #49, #50 and #51. [spec, open PR]
2. The `SELF` placeholder, and that a rule using it is value-intrinsic.
3. Self-attested entries (the one-step clause in 4.1).
4. `ListedIn` minted at a delivery boundary.
5. For layout A, either a computation may read a principal from a profile
   cell's label (which `principalOf()` does today in pattern code [code]), or
   `currentPrincipal()` is defined in a computation. The second is a smaller
   version of the viewer proposal's §18 text, with lower stakes: a wrong answer
   cannot misdirect a release.

### 4.5 Lean obligations

specs#59 adds `Cfc/ListMembership.lean`; it is not on `main` [spec]. On top of
it:

1. **The rule adds only the value's own list.** If the rule adds
   `Members(L, S)` at a location, one stamp entry of the named code covers both
   the location and `L`.
2. **Self-attested resolution is sound.** `P` is listed at `L` only through an
   entry naming `P` under that stamp. Fail closed on an overwritten list, an
   unreadable document, a missing manifest.
3. **Noninterference per principal.** For two states on which the blessed
   function's outputs agree wherever they name `P`, `P`'s observations are
   equal. This replaces the viewer proposal's theorem 2 and has the same shape.
4. **No pairing.** No reachable label pairs content from one attempt with a
   list another attempt wrote. In the viewer proposal this is the "two atoms in
   place of one" counterexample; here it is positional.
5. **The carry.** specs#59 already owes "the §5.3 carry of the exchanged
   `Members` clause onto derived values" as a follow-up.
6. **Rejected variants, `decide`-checked:** a pointer taken from the schema; a
   list in a second document (3.1's attack); counting entries without the
   self-attestation; layout B without a fresh document per occupancy.

The transfer property of layout B is a statement about a pattern, not about the
calculus, and would be prose.

One hypothesis leaves the model. The viewer proposal needs "the principal the
runtime supplies to an attempt is the principal whose instance the attempt
writes". This design needs nothing of the kind. [design]

### 4.6 Distance from labs today

- `list-membership.ts`, `members-capture.ts` and `runtime-list-membership.ts`
  exist only in the unmerged draft labs#8608 and its port in the worktree. They
  are not on `main`. [code]
- The resolver needs the self-attestation path, and the exchange evaluator the
  `SELF` binding, which needs the stamp's entry root at evaluation. The
  per-location evaluation labs#10130 merged is the place. [code] [design]
- The transformer refuses `cfcPattern.members()` in the draft [code, worktree,
  port note 3], and would need the `self` form.
- Delivery by label is not built (Topics top/3173, cited in
  `docs/specs/cfc-conformance-statement.md`). Viewer-bound release needs it
  equally. [code]
- The three conforming fixes on the viewer branch (the writer's stamp on a
  record it reads and sets, a result stamped as the run that wrote it, exchange
  at observation, the last merged as labs#10130) are prerequisites of this
  design too. [code]

My estimate is that this is several times the work of viewer-bound release,
most of it specs#59's. [design]

## 5. What both designs leave open, equally

- **An endorsed entry point fed secrets.** Unendorsed code may read the sealed
  state, since a run "reads freely and is tainted by what it read" [spec, open
  PR #95], and then send `move` events whose small integers encode what it read,
  to a table of its own. `move` stamps what it writes, the view of that table is
  released to the attacker, and a seat index leaks per move. Labs records the
  class as "Stance stuffing through an endorsed entry point" and "The bottom of
  a chain trusts its caller" [code, `cfc-transformed-by-input-witnesses.md`].
  It applies to viewer-bound release and to this design identically, and what
  closes it is an event-integrity requirement on `move` or labeled event
  payloads. I inferred this; nothing ran. It should be checked before either
  design ships.
- **A derivation first computed before the first move**, and **a list mapped
  from the view** (`DERIVED-RELEASE-FINDING.md`, causes 2 and 3). Layout B
  depends on a container of seat documents and meets cause 3 head on. [code,
  worktree]
- **Before delivery checks labels,** the sealed state is in every reader's
  replica in both designs.

## 6. Head to head

### 6.1 The comparison

| | Viewer-bound release | Co-written audience, layout A |
| --- | --- | --- |
| **Soundness** | Sound by argument, with tests. Rests on a mint invariant: the stamp's `viewer` is what the runtime answered, once per attempt, never retained. Four rejected variants show how narrow that is. | Sound by argument, untested. Rests on a resolver invariant: one stamp entry covers value and list. The viewer may be any public input. |
| **Trusted assumption** | The runtime's answer names the principal whose instance the run writes. | None about instances. |
| **Revocation** | None. `User(q)` is on the instance and, since labs#10130, on each derivation. | Expressible, by emptying a list. Weak in practice in layout A, because an idle viewer's instance is not recomputed. Real in layout B. |
| **What is baked into labels** | A principal. | An address. In layout A the address contains the viewer's scope key. |
| **What the author writes** | One lift that asks who it is for; `viewer: v("viewer")` in the rule. | The same lift returning an `audience` field; `/audience` in the rule. |
| **Runtime cost** | `k` views per move. No resolution at the gate. | `k` views per move. One list resolution per gated read, in the same document, and a watch on it. |
| **Spec surface** | Four short changes: §15, §8.9.3, a §8.7 subsection, §18. One acknowledged tension with §5.3's witness sentence. | specs#59 whole, plus a placeholder, an entry clause and a delivery mint. Nothing in §15 beyond #59, nothing in §8.9.3. |
| **Proof burden** | Five theorems, one hypothesis, four counterexamples. Self-contained. | Six items on top of specs#59's module, which is unmerged. No hypothesis. |
| **Built** | Yes, behind `viewerBoundRelease`, with a three-principal display test and a served case. | No. |
| **Depends on** | Nothing unruled. | specs#59, #49, #50, #51. |

### 6.2 Fit with the maintainer's recorded positions

Each is quoted from a pull request, and I have not spoken to him.

- "the owner is the list [in that] cell over there … the group membership can
  change from the time you put the label on" (2026-10-07, quoted in specs#59).
  The co-written audience is this, with the cell being the released document's
  own field. Viewer-bound release has no cell. The relayed addition that the
  cell "may be computed" is second-hand, per the task; this design's list is
  computed.
- "you pass in the profile cell and then something needs to work out this is
  actually this user and declassify" (2026-10-01, same source). Layout A passes
  in the profile cell.
- "`user` and `session` scope say which instance a document is, not who may read
  it. … Audience is the label's job." (specs#95 takeover comment.) Viewer-bound
  release answers that the release "does not rest on where an instance is
  stored", which is true of the gate. But the principal it releases to *is* the
  instance key, read back by the runtime. In this design the instance key is
  not consulted.
- §15: `TransformedBy` is "minted by the runtime for one attempt, from the
  identity that ran and the witnesses it consumed, and by nothing else" [spec].
  This design leaves the row as it is.
- "we have field-level opaque/not-opaque. why is that not sufficient?"
  (labs#3750). Layout B is that model, with one rule added.

There is one position this design fits *worse*. §4.9.3's candidate rule, as
specs#59 restates it, is that "the records a decision depends on must not be a
function of the decision" [spec, open PR #59]. A list found from the location's
own address is fixed before evaluation, so I think it holds, but it is a new
way to find a candidate and he may not accept it.

### 6.3 Which I would choose, and whether they are rivals

**They are rivals on one question and partners on the rest.** The question is
where the released principal comes from: a field the runtime stamps on the
evidence, or a cell the blessed code writes. Everything else is shared: the
sealed record, the single writer, the per-viewer instance, the witnessed
`TransformedBy`, the three conforming fixes, and the open items of section 5.

**They are also complements across cases.** specs#59 as drafted is for an
audience an entitled owner curates (Find My). An instance-bound writer (3.3)
extends it to a custodian who curates. The co-written audience is for an
audience that is a fact inside sealed state. Viewer-bound release covers the
subset of that last case where each principal's view is "what the state says
about me".

**For the custodian-owner case I would choose the co-written audience, layout
A, as what the specification should say,** for three reasons.

1. It removes all three structural objections to viewer-bound release (a
   literal principal in labels, scope as the source of a declassification, a
   third ingredient in `TransformedBy`) without changing the pattern's shape.
2. It is one mechanism with the list release, and it closes two of specs#59's
   own open items (computed lists, and entries the policy's writer wrote),
   where viewer-bound release is a second, parallel way to name an audience.
3. It needs one less trusted assumption.

**I hold that choice with real reservations.**

- It is unbuilt, and the self-attestation step is exactly the kind of clause
  that looks right on paper. Viewer-bound release was run against three
  principals.
- Its headline advantage, revocation, is mostly theoretical in layout A, and
  for "your own view" a baked `User(q)` is arguably the *correct* meaning: what
  a player was shown is theirs.
- It waits on four unruled pull requests.

**So, in order:** land the conforming fixes regardless; keep viewer-bound
release behind its flag as the stopgap that works; file the co-written audience
as the follow-up that answers specs#59's "computed lists", with the card table
as its worked case; and ask the maintainer the one question that decides
between them, which is whether a rule may bind a list from the location it is
evaluated at.

**If the answer is that specs#59 will not be ruled soon, or that `SELF` is not
acceptable, viewer-bound release is the right design,** and it should borrow
three things:

1. **Make the viewer an output fact, not an instance fact.** Mint the field
   only when the blessed function *returned* the principal at a path the rule
   names, and the runtime's answer equals it. Then the proposal's rejected
   "viewer as a parameter" variant and its accepted form become the same
   check, and the text no longer needs "the principal whose scope instance the
   run reads and writes".
2. **Say in the text what is not revocable and why that is the intended
   meaning** for a principal's own view, so the absence of revocation is a
   ruling and not an omission, and say that a case needing revocation uses a
   list.
3. **State the event-integrity requirement on the writer** (section 5), which
   the proposal's "No instance binding is needed for this shape" currently
   passes over.

## 7. What I did not do

- Nothing was run, compiled or type-checked. The code and JSON above are
  sketches.
- I did not read the Lean in specs#59 line by line, or `formal/Cfc/Exchange.lean`
  to confirm that a location-relative binding fits its model of a rule.
- I did not read labs#8608 itself, only its port and the findings.
- I did not verify that labs resolves a location's stamp to a single entry root
  in the form `SELF` needs; `cfc-transformed-by-input-witnesses.md` describes
  per-location resolution "per origin component, the most specific entry at or
  above it", which suggests it can.
- The maintainer's "may be computed" remark is second-hand and I could not find
  it in writing.
