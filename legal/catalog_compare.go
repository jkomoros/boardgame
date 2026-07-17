package legal

import (
	"fmt"
	"strconv"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

// Template keys for the comparison/property predicates in this file.
// Task 6's legal.DefaultTemplates() consumes these (via defaultTemplateKeys,
// declared at the bottom of catalog_stack.go) to build the default template
// table; see that file for the full list and handoff note.
const (
	// TemplatePropAtLeast is the default Fail template key for PropAtLeast.
	// Bindings: "value" (the actual int prop value), "min" (the threshold).
	TemplatePropAtLeast = "legal.prop_at_least"
	// TemplatePropCompare is the default Fail template key for PropCompare.
	// Bindings: "value", "op", "n".
	TemplatePropCompare = "legal.prop_compare"
	// TemplatePlayerBool is the default Fail template key for PlayerBool and
	// PlayerBoolIs (spec §4: both share the "playerBool" registry entry, and
	// share this one template key — see playerBoolConstructor's doc comment
	// for why a second key was not warranted). Bindings: "prop" (the
	// property name checked), "want" (the string "true" or "false" — what
	// the property was required to equal).
	TemplatePlayerBool = "legal.player_bool"
	// TemplatePropEquals is the default Fail template key for PropEquals.
	// Bindings: "value" (the actual resolved value, stringified per its
	// type), "want" (the value arg exactly as given).
	TemplatePropEquals = "legal.prop_equals"
	// TemplatePropNotEquals is the default Fail template key for
	// PropNotEquals. Bindings: same as TemplatePropEquals.
	TemplatePropNotEquals = "legal.prop_not_equals"
)

// legalCompareOps maps a PropCompare operator string to the int comparison
// it performs. This is also the authoritative set of valid operators: a
// PropCompare spec whose op isn't a key here fails at construction time.
var legalCompareOps = map[string]func(value, n int) bool{
	"==": func(value, n int) bool { return value == n },
	"!=": func(value, n int) bool { return value != n },
	"<":  func(value, n int) bool { return value < n },
	"<=": func(value, n int) bool { return value <= n },
	">":  func(value, n int) bool { return value > n },
	">=": func(value, n int) bool { return value >= n },
}

// Comparison operators accepted by PropCompare and StackCount. String
// literals remain supported; these constants improve discoverability and
// avoid spelling mistakes without breaking existing callers.
const (
	OpEqual              = "=="
	OpNotEqual           = "!="
	OpLessThan           = "<"
	OpLessThanOrEqual    = "<="
	OpGreaterThan        = ">"
	OpGreaterThanOrEqual = ">="
)

// PropAtLeast returns a Spec for the "propAtLeast" predicate: Passes if the
// int property at path is >= n. Works for any int prop reachable via the
// catalog's path grammar (game.X, player.X, move.X).
func PropAtLeast(path string, n int) Spec {
	return Spec{Name: "propAtLeast", Args: []string{path, strconv.Itoa(n)}}
}

// PropCompare returns a Spec for the "propCompare" predicate: Passes if the
// int property at path compares to n via op, which must be one of ==, !=,
// <, <=, >, >=.
func PropCompare(path, op string, n int) Spec {
	return Spec{Name: "propCompare", Args: []string{path, op, strconv.Itoa(n)}}
}

// PlayerBool returns a Spec for the "playerBool" predicate: Passes if the
// current player's bool property named prop is true. ("Current player" is
// what the catalog's player.* path grammar resolves against; in a future
// players[*] quantifier context, that will mean the player being
// quantified over.) This is the 1-arg, backward-compatible spelling —
// existing specs/corpus built with PlayerBool keep their exact meaning
// (want=true) unchanged; see PlayerBoolIs for the explicit-want form (spec
// §4).
func PlayerBool(prop string) Spec {
	return Spec{Name: "playerBool", Args: []string{prop}}
}

// PlayerBoolIs returns a Spec for the "playerBool" predicate with an
// explicit want: Passes if the current player's bool property named prop
// equals want. Both PlayerBool(prop) (want=true, 1-arg spelling) and
// PlayerBoolIs(prop, want) (2-arg spelling) resolve through the SAME
// registry entry ("playerBool") — see playerBoolConstructor, which accepts
// either 1 or 2 args. AllActivePlayers' inner grammar also accepts this
// 2-arg form (buildAllActivePlayersLeaf, catalog_players.go), so
// AllActivePlayers(PlayerBoolIs("DoneWithPhase", false)) is a valid
// per-player negation leaf.
func PlayerBoolIs(prop string, want bool) Spec {
	return Spec{Name: "playerBool", Args: []string{prop, strconv.FormatBool(want)}}
}

// playerBoolWant parses playerBool's optional second arg ("true"/"false")
// into a bool. Shared by playerBoolConstructor (below) and
// AllActivePlayers' inner-grammar compiler (buildAllActivePlayersLeaf,
// catalog_players.go) so both accept identical 2-arg syntax by construction
// — one parser, not two independently-maintained copies.
func playerBoolWant(arg string) (bool, error) {
	switch arg {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("must be \"true\" or \"false\", got %q", arg)
	}
}

// PropEquals returns a Spec for the "propEquals" predicate: Passes if the
// property at path equals value. What "equals" means depends on the
// property's RESOLVED type (int/bool/enum/PlayerIndex — see
// propEqualsConstructor's doc comment for the judgment call this package
// had to make about WHEN that type is discovered):
//
//   - int: value parses as an int, compared numerically.
//   - bool: value must be exactly "true" or "false".
//   - enum: value is an enum value NAME (e.g. "Black"), compared against
//     the resolved enum.ImmutableVal via its own Enum().ValueFromString.
//   - PlayerIndex: value is an int, or one of the specials "observer"
//     (boardgame.ObserverPlayerIndex) or "admin" (boardgame.AdminPlayerIndex).
//
// Reads {path} on LegalFacetValues; CostTrivial; default Fail template
// TemplatePropEquals with {value, want} bindings.
func PropEquals(path, value string) Spec {
	return Spec{Name: "propEquals", Args: []string{path, value}}
}

// PropNotEquals returns a Spec for the "propNotEquals" predicate: the exact
// negation of PropEquals (same type dispatch, same Reads/Cost, default Fail
// template TemplatePropNotEquals). Negation only flips a definite
// match/no-match verdict — an Unknown (unparseable value for the resolved
// type, unresolvable path, unknown enum name) is never flipped to a Pass.
func PropNotEquals(path, value string) Spec {
	return Spec{Name: "propNotEquals", Args: []string{path, value}}
}

// propEqualsPlayerIndexValue parses value as PropEquals/PropNotEquals'
// PlayerIndex arm: the specials "observer"/"admin", or a plain int. This
// never needs the chest or an example state (unlike the enum arm), so it
// COULD be parsed eagerly at construction; per propEqualsFamilyConstructor's
// doc comment, it is deliberately parsed lazily, at Evaluate time, together
// with the enum arm, since a *LegalPredicateConstructor* cannot know which
// arm applies until it sees the resolved PropertyType.
func propEqualsPlayerIndexValue(value string) (boardgame.PlayerIndex, error) {
	switch value {
	case "observer":
		return boardgame.ObserverPlayerIndex, nil
	case "admin":
		return boardgame.AdminPlayerIndex, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("value %q is not a valid PlayerIndex (want an integer, \"observer\", or \"admin\")", value)
	}
	return boardgame.PlayerIndex(n), nil
}

// propEqualsPlayerIndexString renders p the same way propEqualsPlayerIndexValue
// parses it, for the "value" Fail binding.
func propEqualsPlayerIndexString(p boardgame.PlayerIndex) string {
	switch p {
	case boardgame.ObserverPlayerIndex:
		return "observer"
	case boardgame.AdminPlayerIndex:
		return "admin"
	}
	return strconv.Itoa(int(p))
}

// valueMatchesAnyEnumInChest checks if value is a valid value name in ANY
// enum within the chest. Returns true if found, false otherwise. If chest is
// nil, returns false.
func valueMatchesAnyEnumInChest(value string, chest *boardgame.ComponentChest) bool {
	if chest == nil {
		return false
	}
	enumSet := chest.Enums()
	if enumSet == nil {
		return false
	}
	for _, enumName := range enumSet.EnumNames() {
		e := enumSet.Enum(enumName)
		if e == nil {
			continue
		}
		if e.ValueFromString(value) != enum.IllegalValue {
			return true
		}
	}
	return false
}

// propEqualsFamilyConstructor builds the shared registry entry for
// "propEquals" (negate=false) and "propNotEquals" (negate=true).
//
// THE JUDGMENT CALL (design spec §2 says type dispatch "happens at boot: the
// constructor resolves the path's PropertyType from the example state and
// bakes the right comparator" — a runtime type surprise would then be
// impossible). That is not achievable as written: LegalPredicateConstructor.
// Constructor (legal_predicate.go) has the signature
// (spec, chest, resolve) — no example state is threaded through. Widening
// that signature to also carry an example state was considered and
// rejected: it would change the constructor contract for every registered
// predicate mid-round (catalog and game-registered alike), which is exactly
// the cost the pre-spec critique warned about (~15 files) for a single
// predicate pair. So this constructor takes the fallback branch the brief
// names: type dispatch is deferred to Evaluate, where ctx.ResolvePath's
// second return value IS the resolved PropertyType (see legal_path.go's
// ResolvePath). Two of the four arms (int, bool) have no dependency on the
// chest or example state to validate value's shape, so THOSE two are
// eagerly parsed once here at construction time (intWant/intErr, boolWant/
// boolErr below) rather than being re-parsed on every Evaluate call. The
// other two arms (enum, PlayerIndex) are validated lazily, inside Evaluate:
// enum because an unknown value NAME can only be checked against the
// property's actual enum.Enum, which is only reachable once a real
// enum.ImmutableVal has been resolved off a real state; PlayerIndex because,
// per the brief, it is grouped with enum's lazy-validation arm even though
// its "observer"/"admin"/int grammar is self-contained and technically
// eager-parseable — kept lazy here to match the brief's prescribed split
// rather than inventing an inconsistent third eager arm.
//
// Consequence: an unknown enum value name, or a value that doesn't parse
// for whichever type the path resolves to, is a fail-closed LegalUnknown at
// Evaluate time, NOT a boot-time construction error — exactly the "unknown
// enum name = boot error" line of the design spec is not delivered by this
// implementation. That aspiration is recorded for the spec's impl-notes
// appendix (Task 4's report) rather than silently dropped: closing it for
// real requires either the wider constructor signature above, or a
// dedicated boot-validation hook constructors can register into (neither
// exists yet). No predicate here EVER panics: every unparseable/mismatched/
// wrong-typed case returns LegalUnknown.
func propEqualsFamilyConstructor(name, defaultTemplate string, negate bool) *PredicateConstructor {
	return &PredicateConstructor{
		Name: name,
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: %s requires 2 args (path, value), got %d", name, len(spec.Args))
			}
			path := spec.Args[0]
			value := spec.Args[1]

			template := spec.Message
			if template == "" {
				template = defaultTemplate
			}

			// Eager parse of the two arms that need no chest/example-state
			// knowledge to validate — see the doc comment above. A parse
			// failure here is NOT a construction error: value's actual
			// arm is only known once the path resolves against a real
			// state, so e.g. a value of "Black" (a perfectly good enum
			// name) legitimately fails both of these eager parses.
			intWant, intErr := strconv.Atoi(value)
			var boolWant bool
			var boolErr error
			switch value {
			case "true":
				boolWant = true
			case "false":
				boolWant = false
			default:
				boolErr = fmt.Errorf("legal: %s: value %q is not \"true\" or \"false\"", name, value)
			}

			// Construction-time typo guard: if chest is provided and value
			// doesn't match any int/bool/PlayerIndex-special or any enum value
			// name, it's likely a typo. This guard is gated on chest != nil
			// because test harnesses may pass nil chest; calling chest.Enums()
			// on nil would panic. If value is a valid enum name or matches one
			// of the other arms, we construct and defer the final dispatch to
			// Evaluate time (when the path's PropertyType is known).
			if chest != nil && intErr != nil && boolErr != nil && value != "observer" && value != "admin" && !valueMatchesAnyEnumInChest(value, chest) {
				return nil, fmt.Errorf("legal: %s: value %q matches no int/bool/playerindex-special and no enum value name in the chest — likely a typo", name, value)
			}

			return &Predicate{
				Name:             name,
				ClientEvaluable:  true,
				Args:             spec.Args,
				Reads:            []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				AllowedReadTypes: map[PropPath][]boardgame.PropertyType{PropPath(path): {boardgame.TypeInt, boardgame.TypeBool, boardgame.TypeEnum, boardgame.TypePlayerIndex}},
				Cost:             boardgame.LegalCostTrivial,
				EmittedTemplates: []string{template},
				EmittedBindings:  map[string][]string{template: {"value", "want"}},
				Evaluate: func(ctx Context) Verdict {
					val, propType, err := ctx.ResolvePath(PropPath(path))
					if err != nil {
						return UnknownVerdict(err.Error())
					}

					var match bool
					var actual string

					switch propType {
					case boardgame.TypeInt:
						if intErr != nil {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q is an int property but value %q does not parse as an int", name, path, value))
						}
						i, ok := val.(int)
						if !ok {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q resolved to an int-typed property but its value was not an int (%T)", name, path, val))
						}
						actual = strconv.Itoa(i)
						match = i == intWant
					case boardgame.TypeBool:
						if boolErr != nil {
							return UnknownVerdict(boolErr.Error())
						}
						b, ok := val.(bool)
						if !ok {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q resolved to a bool-typed property but its value was not a bool (%T)", name, path, val))
						}
						actual = strconv.FormatBool(b)
						match = b == boolWant
					case boardgame.TypeEnum:
						v, ok := val.(enum.ImmutableVal)
						if !ok || v == nil {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q resolved to an enum-typed property with no value", name, path))
						}
						key := v.Enum().ValueFromString(value)
						if key == enum.IllegalValue {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q is an enum property (enum %q) but %q is not a valid value name", name, path, v.Enum().Name(), value))
						}
						actual = v.String()
						match = v.Value() == key
					case boardgame.TypePlayerIndex:
						p, ok := val.(boardgame.PlayerIndex)
						if !ok {
							return UnknownVerdict(fmt.Sprintf("legal: %s: path %q resolved to a PlayerIndex-typed property but its value was not a PlayerIndex (%T)", name, path, val))
						}
						want, err := propEqualsPlayerIndexValue(value)
						if err != nil {
							return UnknownVerdict(fmt.Sprintf("legal: %s: %v", name, err))
						}
						actual = propEqualsPlayerIndexString(p)
						match = p == want
					default:
						return UnknownVerdict(fmt.Sprintf("legal: %s: path %q resolved to property type %v, which typed equality does not support (want int, bool, enum, or PlayerIndex)", name, path, propType))
					}

					if negate {
						match = !match
					}
					if match {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"value": String(actual),
						"want":  String(value),
					})
				},
			}, nil
		},
	}
}

// propAtLeastConstructor returns the registry entry for "propAtLeast".
func propAtLeastConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "propAtLeast",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: propAtLeast requires 2 args (path, n), got %d", len(spec.Args))
			}
			path := spec.Args[0]
			n, err := strconv.Atoi(spec.Args[1])
			if err != nil {
				return nil, fmt.Errorf("legal: propAtLeast: arg 2 (n) must be an integer: %w", err)
			}

			template := spec.Message
			if template == "" {
				template = TemplatePropAtLeast
			}

			return &Predicate{
				Name:              "propAtLeast",
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeInt},
				Cost:              boardgame.LegalCostCheap,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"value", "min"}},
				Evaluate: func(ctx Context) Verdict {
					value, err := resolveIntPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if value >= n {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"value": Int(value),
						"min":   Int(n),
					})
				},
			}, nil
		},
	}
}

// propCompareConstructor returns the registry entry for "propCompare".
func propCompareConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "propCompare",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: propCompare requires 3 args (path, op, n), got %d", len(spec.Args))
			}
			path := spec.Args[0]
			op := spec.Args[1]
			cmp, ok := legalCompareOps[op]
			if !ok {
				return nil, fmt.Errorf("legal: propCompare: unknown op %q (expected one of ==, !=, <, <=, >, >=)", op)
			}
			n, err := strconv.Atoi(spec.Args[2])
			if err != nil {
				return nil, fmt.Errorf("legal: propCompare: arg 3 (n) must be an integer: %w", err)
			}

			template := spec.Message
			if template == "" {
				template = TemplatePropCompare
			}

			return &Predicate{
				Name:              "propCompare",
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeInt},
				Cost:              boardgame.LegalCostCheap,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"value", "op", "n"}},
				Evaluate: func(ctx Context) Verdict {
					value, err := resolveIntPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if cmp(value, n) {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"value": Int(value),
						"op":    String(op),
						"n":     Int(n),
					})
				},
			}, nil
		},
	}
}

// playerBoolConstructor returns the registry entry for "playerBool".
//
// Want-false template decision (spec §4 leaves this a judgment call: "either
// the same template with a want binding, or a legal.player_bool_want
// template — pick one"): this reuses TemplatePlayerBool for both want=true
// and want=false rather than minting a second key. A second key would only
// earn its keep if the two cases needed structurally different bindings or
// wildly different phrasing; they don't — "requires {prop} to be {want}"
// covers both honestly, and it keeps the registered-template surface (and
// the corpus-completeness/template-coverage meta-tests) from growing for a
// predicate that didn't grow in registry-name terms either (still one
// "playerBool" entry). The "want" binding is new (added unconditionally,
// including for the 1-arg/want=true case) so a caller rendering an existing
// want=true Fail Verdict gets "requires X to be true" exactly as before,
// word for word — the OLD literal template body was "requires {prop} to be
// true"; the new body substitutes the literal "true" for "{want}", which
// renders identically when want=true. No existing corpus case pins the
// rendered STRING (only the template KEY, "legal.player_bool" — see
// testdata/conformance/playerBool.json), so this is not a breaking change
// under the corpus's own conformance contract.
func playerBoolConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "playerBool",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 1 && len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: playerBool requires 1 or 2 args (prop, optional want), got %d", len(spec.Args))
			}
			prop := spec.Args[0]
			want := true
			if len(spec.Args) == 2 {
				w, err := playerBoolWant(spec.Args[1])
				if err != nil {
					return nil, fmt.Errorf("legal: playerBool: arg 2 (want) %s", err)
				}
				want = w
			}
			path := "player." + prop

			template := spec.Message
			if template == "" {
				template = TemplatePlayerBool
			}

			return &Predicate{
				Name:              "playerBool",
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeBool},
				Cost:              boardgame.LegalCostTrivial,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"prop", "want"}},
				Evaluate: func(ctx Context) Verdict {
					value, err := resolveBoolPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if value == want {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"prop": String(prop),
						"want": String(strconv.FormatBool(want)),
					})
				},
			}, nil
		},
	}
}

// propEqualsConstructor returns the registry entry for "propEquals".
func propEqualsConstructor() *PredicateConstructor {
	return propEqualsFamilyConstructor("propEquals", TemplatePropEquals, false)
}

// propNotEqualsConstructor returns the registry entry for "propNotEquals".
func propNotEqualsConstructor() *PredicateConstructor {
	return propEqualsFamilyConstructor("propNotEquals", TemplatePropNotEquals, true)
}
