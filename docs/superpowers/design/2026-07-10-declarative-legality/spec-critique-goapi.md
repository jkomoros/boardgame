# Spec Critique — Go API Quality & Idiom

**Spec:** `docs/superpowers/specs/2026-07-10-declarative-legality-design.md`
**Lens:** Go API quality and idiom — reviewed as a Go practitioner would in PR review.
**Reference points read:** `stack_constraint.go`, `constraints/` (constructors.go, same.go, prop_path.go, doc.go), `errors/main.go`, `moves/default.go` (Legal path), `moves/with.go`, `server/api/main.go` (form legality).

The headline: the *architecture* is sound and the anti-tarpit discipline is the best thing in the doc. But the *Go surface* the spec draws is looser than the repo's own precedent (`constraints`), and several of the value types have zero-value and `any`-typed hazards that will bite exactly the authors this system is meant to protect. Findings ordered by severity.

---

## BLOCKING

### B1. The zero `Verdict` is an accidental legal-move bug waiting to happen

> ```go
> const (
>     Pass Outcome = iota
>     Fail
>     Unknown
> )
> ```
> ```go
> type Verdict struct {
>     Outcome Outcome
>     Message *Message
>     Reason  string
> }
> ```

`Pass = 0` means the zero `Verdict{}` is a **silent pass**. In a legality engine that is the single most dangerous default you can pick: any code path that constructs a `Verdict` and forgets to set `Outcome`, any `map[…]Verdict` miss that returns the zero value, any partially-initialized struct in a test or a cache, *reads as "this move is legal."* Legality is precisely the domain where "fail closed" is the rule — an accidental zero should deny, not permit.

The spec even flags Unknown as "load-bearing … how a predicate that cannot decide stays honest instead of guessing," but then makes the *guessing* value (Pass) the zero. That is backwards.

**Better:** make the zero value non-committal and non-permissive. Either:

```go
const (
    Unknown Outcome = iota // zero value: "I did not decide" — fail-closed at the gate
    Pass
    Fail
)
```

so a zero `Verdict` is `Unknown`, which the engine already treats as "not evaluable / fall back," never as "legal." Or, if you want the enum order to read Pass/Fail/Unknown for docs reasons, keep it but make the zero *invalid*:

```go
const (
    outcomeInvalid Outcome = iota // unexported; the true zero, never legal
    Pass
    Fail
    Unknown
)
```

and have the evaluator treat `outcomeInvalid` identically to a panic-worthy programmer error (or at minimum, identically to `Unknown`). Add a `String()` and a `Valid()` method. The cost is one line; the bug it prevents is a move that silently becomes legal in production and is nearly impossible to spot in review because *nothing was written*.

This interacts with B2: constructor functions (`PassVerdict()`, `FailT()`) reduce but do not eliminate the exposure, because the struct is exported and hand-constructable.

---

### B2. `Predicate` as a 5-method interface diverges from the repo's own precedent without justifying the divergence

> ```go
> type Predicate interface {
>     Name() string
>     Args() []string
>     Reads() []PropPath
>     Cost() Cost
>     Evaluate(ctx Context) Verdict
> }
> ```

The repo already faced this exact modeling choice and chose the opposite. `StackConstraint` is **a func type**, not an interface:

```go
type StackConstraint func(destination ImmutableStack, proposed []ImmutableComponentInstance, state ImmutableState) error
```

The *identity/metadata* (Name + arg-parsing) lives in a separate small **struct**, `StackConstraintConstructor`, and constraint funcs are produced by constructors. So the precedent is: **behavior is a func; identity is data; the two are not welded into one interface.** The spec's `Predicate` welds four data-ish accessors (`Name`, `Args`, `Reads`, `Cost`) onto one behavior method (`Evaluate`) and calls the bundle an interface.

Four of the five methods are pure getters that return values fixed at construction time. That is a struct with a function field, not an interface. The interface shape forces every predicate — including every trivial catalog leaf — to be a named type with five method definitions, boilerplate the func-type precedent specifically avoids. It also makes the common `Evaluate` the *only* interesting method carry the same syntactic weight as `Cost()`.

**Better — mirror `constraints` exactly:**

```go
// Predicate is one legality question. Metadata is data; behavior is a func.
type Predicate struct {
    Name     string
    Args     []string
    Reads    []PropPath
    Cost     Cost
    Evaluate func(ctx Context) Verdict
}
```

Composites (`any`) become a `Predicate` whose `Evaluate` closes over its children and whose `Reads` is the union of theirs — computed once at construction, which is where you *want* it (the interface version recomputes `Reads()` on every call unless each type caches it by hand). This:

- matches `StackConstraint`'s "func + constructor-struct" idiom the repo already teaches;
- makes `Reads`/`Cost`/`Name` obviously immutable identity, not recomputable behavior;
- kills five-method boilerplate for every leaf in the catalog;
- keeps the constructor registry (`PredicateConstructor`) unchanged — it already returns `(Predicate, error)`.

If there is a genuine reason to keep the interface — e.g. you expect predicates that lazily compute `Reads` from state, or you want nominal typing so `revealableCardAt` is its own type for testing — the spec must *say so*, because the burden is on the diverging design to justify departing from `StackConstraint`. As written it doesn't, and I don't see the reason: every example predicate in §8 has fixed metadata. **Recommend the struct.** (If kept as interface, at minimum drop `Name()`/`Args()` from it — those are registry identity and belong on the constructor/Spec, exactly as `StackConstraintConstructor.Name` is not on the `StackConstraint` func.)

---

### B3. The `.WithMessage(...)` builder chain contradicts the `WithPreconditions(specs ...legal.Spec)` signature

> ```go
> moves.WithPreconditions(
>     legal.PropAtLeast("player.CardsLeftToReveal", 1).
>         WithMessage("reveal.no_cards_left"),
>     legal.RevealableCardAt(...),
> )
> ```
> ```go
> moves.WithPreconditions(specs ...legal.Spec)
> ```

`WithPreconditions` is typed `...legal.Spec` (a plain serializable struct with `Name/Args/Sub`). But the builders `legal.PropAtLeast(...).WithMessage(...)` are shown as a fluent chain — `WithMessage` must be a method returning something. If `PropAtLeast` returns `Spec` and `WithMessage` is a method on `Spec` returning `Spec`, fine — but then `Spec` needs a `Message`/`Template` field the wire-format `Spec` (`{name, args, sub}`) doesn't show. There's an **undeclared field on the serialized type**, or a type mismatch between what builders return and what `WithPreconditions` accepts.

This is not a nit because it determines the whole authoring ergonomics and the JSON shape. Three coherent options; the spec must pick one:

1. **`Spec` carries the message.** Add `Message *Message` (or `Template string` + `Bindings`) to `Spec`, wire-serialized. Builders return `Spec`; `WithMessage` mutates and returns it. Clean, but now `Spec` is not purely `{name,args,sub}` and the wire examples in §1 are incomplete — fix them.
2. **A distinct builder type.** `PropAtLeast` returns a `*specBuilder` with `WithMessage`, and `WithPreconditions` takes `...SpecBuilder` (an interface with `Spec() Spec`). More Go-idiomatic (builders aren't the wire type), but adds a type.
3. **Message is a Spec arg.** No `WithMessage`; the template is positional/trailing in `Args`. Ugly, loses the fluent readability the DX lens wanted.

I'd take (1) and make `Spec` the one authoring+wire type with an optional message, because it keeps a single type users learn. But *say it*, and update the §1 `Spec` struct and JSON examples so `Message` appears. As written, the spec's own examples don't type-check against its own signature.

---

## IMPORTANT

### I1. `Message.Bindings map[string]any` — `any` in a core wire/API type

> ```go
> type Message struct {
>     Template string
>     Bindings map[string]any // {"left": 0}
> }
> ```

`map[string]any` in a type that (a) lives in core, (b) crosses the serialization boundary, and (c) feeds template rendering is a smell, though a defensible one. The repo's own precedent is *already* `map[string]interface{}` — `errors.Fields` is exactly that. So this is consistent with the codebase, which is why it's Important not Blocking. But two concrete hazards the spec should address:

- **JSON round-trip lossiness.** `{"left": 0}` marshals fine, but unmarshals `0` as `float64`, not `int`. A template or a Go-side consumer that type-asserts `bindings["left"].(int)` will panic or mis-render after a wire round-trip. Since the whole point is server-renders-and-client-re-renders, values *will* cross JSON both ways. The spec should constrain binding values to a small JSON-safe closed set (string, number-as-`json.Number` or float64, bool) and *document the number type*, or the "re-renderable on server or client" promise has a latent panic in it.
- **Unrenderable-value silent failure.** Nothing stops a binding value being a `*boardgame.Stack`. Since `Message` is constructed by catalog authors, add a validation/lint pass (you already ship one for `Reads()` conservativeness — piggyback) that rejects non-JSON-scalar binding values at construction/boot.

**Recommend:** keep `map[string]any` for consistency with `errors.Fields` (fighting it would be inconsistent), but (1) document that numbers round-trip as `float64`/`json.Number`, (2) add boot/lint validation of scalar-only values, and (3) mirror `errors.Fields` as the type alias (`type Bindings = errors.Fields`) so the `Verdict.Error()` → `errors.Friendly` adaptation in I3 is a straight handoff, not a re-map.

### I2. Stringly-typed paths and template keys — a real regression vs. what Go can catch, only partly redeemed by boot validation

> `legal.PropAtLeast("player.CardsLeftToReveal", 1)` … `"reveal.no_cards_left"`

Two different stringly-typed surfaces with **very different safety stories**, which the spec conflates under "the constraints precedent makes this OK":

- **Paths (`"player.CardsLeftToReveal"`)** — *acceptable*, and genuinely on-precedent. `constraints` uses string propPaths and validates them at construction against the chest (`validatePropPath`), returning "does not exist on any deck's components; check for typos." The spec promises the same: "Paths are validated against the reader hierarchy at NewGameManager — a typo'd path fails at boot with the move name and path." That is the right bar and matches `constraints`. **Caught at boot. Good.** One gap: the constraints validator only checks the *leaf property name* exists on *some* deck; the new grammar has structural prefixes (`game.`/`player.`/`players[*].`/`move.`) that must validate against *the specific reader for this move/game*, and `move.X` must validate against *this move type's* fields. The spec says this happens; the risk (correctly flagged in §10 as "the biggest net-new component") is that boot validation is only as good as its coverage. Make it a hard requirement that *every* prefix+leaf is resolved at boot, and that a `move.X` referencing a nonexistent move field fails with the move name — otherwise it's runtime-or-never.

- **Template keys (`"reveal.no_cards_left"`)** — *this* is the type-safety regression Go users will resent, and the spec gives it no boot check. There is no registry of valid template keys analogous to the reader hierarchy for paths, so a typo'd template key is caught **never** — it silently renders as the raw key or a "missing template" fallback at display time, in production, to an end user. Unlike paths (validated against a concrete schema) and predicate names (validated against the constructor registry), template keys float free.

  **Better:** give templates the same "resolved against a registry at boot" treatment as predicate names. Require a template catalog (map key→format string, per locale) registered on the delegate or the `legal` package, and validate at `NewGameManager` that every `WithMessage(key)` and every `legal.FailT(key)`/`legal.Errorf(key)` references a known key, with the move name in the error. Now all three stringly surfaces (name, path, template) fail fast at boot — a coherent story instead of two-out-of-three. Without this, "greppable" is the *only* safety property templates have, and grep is not a type system.

Summary of what's caught where, which the spec should state explicitly:

| Surface | Boot | Runtime | Never |
|---|---|---|---|
| predicate name (registry) | ✅ | | |
| predicate args count/parse | ✅ | | |
| path (`player.X`) | ✅ *(if coverage complete)* | | |
| **template key** | ❌ | ❌ | **✅ (silent)** ← fix this |
| `Bindings` value types | ❌ | ⚠️ panic on assert | |

### I3. `Verdict.Error()` → `errors.Friendly` adaptation — sound in shape, but the mapping is under-specified and drops secure/friendly split

> "the structured `Verdict` … is what crosses the core boundary … `Verdict.Error()` adapts to `error`/`errors.Friendly`."

Adapting to `errors.Friendly` is the right call — the server API already does `err.Error()` on the Legal() return (`main.go:1614`), and `errors.Friendly` gives you the insecure-friendly-string / secure-string / Fields triad that legality explanations *need* (a "not your turn" is friendly-safe; a residue leaking hidden state is not). But the spec says "adapts" and stops. Three things to pin down:

- **Which string is which.** `errors.Friendly` has `Error()` (ok-in-insecure-but-maybe-confusing), `FriendlyError()` (end-user-safe), `SecureError()` (server-only). A rendered template like "You have no cards left to reveal this turn" is a `FriendlyError`. What populates `Error()` vs `FriendlyError()` vs `SecureError()` from a `Verdict`? The spec must map: rendered template → `friendlyMsg`; `Verdict.Reason` (e.g. "reads hidden property HiddenCards") → **`secureMsg`**, because Reason can name hidden properties and must not reach the client. Right now Reason's audience is unspecified and it's the field most likely to leak.
- **Bindings → Fields.** `Message.Bindings` should map straight to `errors.Fields` (see I1's alias suggestion). Fields are "secure contexts only" in this repo — consistent, since raw bindings (`{"left": 0}`) may reveal counts of hidden things.
- **nil-Verdict / Pass path.** `Verdict.Error()` on a Pass must return a **nil `error` interface**, not a non-nil `*Friendly` wrapping "". This is the classic Go nil-interface trap and the errors package even documents it on `NewWrapped`. Spec must guarantee `Verdict{Outcome: Pass}.Error() == nil` (true untyped nil), or every `if err := move.Legal(...); err != nil` in the codebase breaks silently. This is arguably Blocking; I keep it Important only because it's a well-known trap the implementer will likely catch — but it must be an explicit test.

### I4. `Context.Move boardgame.Move` nil during field-independent eval — the nil-check burden falls on every predicate author

> ```go
> type Context struct {
>     ...
>     Move boardgame.Move // nil during field-independent evaluation
> }
> ```

A nilable field in the one struct every predicate receives means **every predicate author who touches `ctx.Move` must remember to nil-check**, or panic mid-game. The spec's own example doesn't:

```go
func (p *revealableCardAt) Evaluate(ctx legal.Context) legal.Verdict {
    idx := intField(ctx.Move, p.field)   // ctx.Move may be nil
    ...
```

The engine's bucketing is *supposed* to guarantee field-dependent predicates only run with a bound move — but that guarantee is invisible at the call site and relies on `Reads()` being correctly declared (which §10 admits is "by-convention" for custom predicates). A predicate author who forgets to list `move.CardIndex` in `Reads()` gets sorted into the field-independent bucket and then panics on `ctx.Move.Field(...)`. So a `Reads()` mistake (a metadata bug) becomes a nil-panic (a crash), far from its cause.

**Better, two options:**

1. **Make it structurally impossible to hold the nil.** Split evaluation into two method families keyed by the bucket — field-independent predicates get a `Context` with **no `Move` field at all** (a `StaticContext`), field-dependent ones get a `MoveContext` with a guaranteed-non-nil `Move`. The type system then enforces what the bucketing intends; a field-independent predicate *cannot* reference `move.X` because the field isn't there. This is more types but removes an entire class of author error and makes the `Reads()`→bucket contract checkable.
2. **If keeping one `Context`,** at minimum ship `ctx.Move` access through a helper that panics with a *diagnostic* ("predicate %q read move field but declared no move.* path in Reads()") rather than a bare nil deref, and make `intField(nil, …)`-style helpers in the catalog defensive. This keeps the burden but makes the failure legible.

I'd push for (1): it converts the spec's prose invariant ("nil during field-independent evaluation") into a compile-time one, and the two contexts fall naturally out of the two buckets the plan already maintains.

### I5. Optional-interface upgrade pattern (`PreconditionsProvider`, `CustomLegaler`) — idiomatic, with one real embedding hazard

> ```go
> type PreconditionsProvider interface { PreconditionPlan() *PreconditionPlan }
> type CustomLegaler interface { LegalCustom(state ImmutableState, proposer PlayerIndex) error }
> ```

The optional-interface / interface-upgrade pattern (`if p, ok := move.(PreconditionsProvider); ok`) is thoroughly idiomatic Go and consistent with how the repo already probes moves (`interfaces.GatheringStartMover`, `SourceStacker`, etc. — `main.go:1607`). No objection to the pattern. Two hazards to nail down:

- **Embedding silently confers the interface — including the *wrong* plan.** Because `moves.Default` implements `PreconditionsProvider`, *every* move that embeds `Default` (i.e. essentially all of them) satisfies the interface via promotion. That's the intent. But it means a hand-rolled move that embeds `Default` **for convenience** but overrides `Legal()` wholesale now *also* advertises a `PreconditionPlan()` that the engine may trust — while the overridden `Legal()` does something else entirely. The "opaque fallback for wholesale Legal() overrides" guarantee (§2 tests) hinges on the engine detecting the override, but Go gives you **no way to detect that a method was overridden** — `move.Legal` is just a method value; you can't tell if it's `Default.Legal` or a shadow. The `opaque bool` in `PreconditionPlan` has to be set by *something*, and there is no reflowable signal. Spec must specify the detection mechanism: likely an explicit opt-out (`moves.WithOpaqueLegal()` / a marker method), because "we notice you overrode Legal" is not mechanizable in Go. As written this is hand-waved and it's load-bearing for the prime guarantee.

- **Method-set breakage on pointer vs value embedding.** If `PreconditionPlan()` is defined on `*Default` but a move embeds `Default` by value and is handled as a value, the interface isn't satisfied. The repo mixes these. Pin the receiver convention (everything on pointer receivers, moves always handled as pointers — which is already the norm here) and add a test that a value-embedded move still upgrades, or someone will lose an afternoon.

### I6. `Cost` as an unbounded exported `int` enum invites out-of-range values

> ```go
> type Cost int
> const ( CostTrivial Cost = iota; CostCheap; CostModerate; CostExpensive )
> ```

Exported `int` named type with four consts and no validation: an author can write `Cost(99)` or leave it zero. Zero here is `CostTrivial`, which is benign (a mis-costed predicate just runs early — a perf nit, not a correctness bug, unlike B1). So this is minor. But since the engine stable-sorts on `Cost`, an out-of-range value silently sorts somewhere. Add a `String()` and clamp/validate at plan-build. Not worth more than a line, hence Important-low.

---

## NIT

### N1. Package name `legal` — collision and readability

`legal` is fine and reads well next to `constraints`. `legal.Any`, `legal.PropAtLeast`, `legal.Errorf` all scan cleanly. Minor worries:

- `legal.Any` collides conceptually with the builtin `any` (Go 1.18+ alias for `interface{}`). Reading `legal.Any(...)` next to `map[string]any` in the same file is momentarily confusing. Not fatal — it's clearly a call vs. a type — but consider `legal.AnyOf(...)` which also reads more like the disjunction it is.
- `legal` as an identifier can shadow: an author with a local `legal := ...` bool ("is this legal?") shadows the package. Unlikely given the domain, but `legality` or `rules` avoid it. I'd keep `legal` — the rhyme with `constraints` and the brevity win.
- `legal.Errorf` reads slightly oddly (it doesn't format like `fmt.Errorf`; it takes a template *key* + bindings, not a format string). The `f` suffix in Go connotes printf-style formatting. Consider `legal.TemplateErr(key, bindings)` or `legal.Failf`. Minor.

### N2. `Args() []string` everywhere loses type information

Every predicate's args are `[]string` even when they're semantically ints (`PropAtLeast("...", 1)` — the `1` becomes `"1"`). This mirrors `StackConstraintConstructor`'s `args []string` (struct-tag origin), so it's on-precedent and defensible for the *serialized* form. But the *typed builders* (`PropAtLeast(path, min int)`) should keep the int typed in the Go API and only stringify at the `Spec` boundary — don't make Go authors write `PropAtLeast("path", "1")`. The §8 examples correctly show `1` (int); just make sure the builder signatures are typed and the stringification is internal. Nit because it's probably intended, but unstated.

### N3. `PropPath` type is used but never defined in the spec

`Reads() []PropPath` and `allReadPaths []legal.PropPath` reference a `PropPath` type that §1 never declares (only `Spec`, `Message`, `Verdict`, `Outcome`, `Cost`, `Context`, `Predicate` are shown). Is it `type PropPath string`? A struct with prefix+leaf? This matters for I2 (boot validation) and for the dirty-tracking that §5 defers. Define it. If it's `type PropPath string`, say so; if structured (`{Scope, Leaf}`), that's better for validation and should be shown.

### N4. `Reason string` on `Verdict` overlaps `Message` and lacks an audience

`Verdict` has both `Message *Message` (structured, localizable) and `Reason string` (freeform, "reads hidden property HiddenCards"). Two explanation channels with unclear precedence. Per I3, `Reason` is server-secure; `Message` is client-safe. Make that explicit in the doc comment, and consider whether `Reason` should itself be a `*Message` (template + bindings) for consistency — a hidden-property name is greppable/localizable too. Minor.

---

## VERDICT

**Is this API good enough to build? Yes, after three fixes — the architecture is right and on-precedent; the value-type details are where it's currently unsafe.** The anti-tarpit rules, the leaf≡node wire format, the constructor-registry mirror of `constraints`, and the "purely sugar / opaque fallback" layering are all well-judged and worth building on. But two of the flaws below are correctness hazards (a zero value that means "legal," an unmechanizable "we detect Legal() overrides" claim), not polish.

**The 3 changes that most improve it:**

1. **Fix the zero `Verdict` to fail closed (B1).** Make `Unknown` (or an unexported invalid) the zero `Outcome`, so a forgotten/defaulted/map-miss `Verdict` never reads as a legal move. One-line change, closes a silent-legalization class of bug in the exact domain where fail-closed is mandatory.

2. **Model `Predicate` as a struct with an `Evaluate` func, mirroring `StackConstraint` (B2).** The repo already made this choice; the 5-method interface diverges without stated cause, adds per-leaf boilerplate, and recomputes immutable metadata. Struct-with-func-field keeps the constructor registry intact and computes `Reads`/`Cost` once at construction.

3. **Give template keys the same boot-time registry validation that predicate names and paths get (I2), and specify the `opaque`-detection mechanism for wholesale `Legal()` overrides (I5).** These are the two places the spec's own guarantees ("fail-fast at boot," "purely sugar / opaque fallback") are currently unenforceable in Go — templates fail never, and "we notice you overrode Legal()" isn't mechanizable without an explicit opt-out marker. Both need a concrete mechanism before build, or the prime guarantee and the fail-fast story are prose, not code.
