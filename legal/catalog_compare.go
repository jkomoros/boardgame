package legal

import (
	"fmt"
	"strconv"

	"github.com/jkomoros/boardgame"
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
	// TemplatePlayerBool is the default Fail template key for PlayerBool.
	// Bindings: "prop" (the property name checked).
	TemplatePlayerBool = "legal.player_bool"
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
// quantified over.)
func PlayerBool(prop string) Spec {
	return Spec{Name: "playerBool", Args: []string{prop}}
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
				Name:             "propAtLeast",
				Args:             spec.Args,
				Reads:            []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
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
				Name:             "propCompare",
				Args:             spec.Args,
				Reads:            []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				Cost:             boardgame.LegalCostCheap,
				EmittedTemplates: []string{template},
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
func playerBoolConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "playerBool",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 1 {
				return nil, fmt.Errorf("legal: playerBool requires 1 arg (prop), got %d", len(spec.Args))
			}
			prop := spec.Args[0]
			path := "player." + prop

			template := spec.Message
			if template == "" {
				template = TemplatePlayerBool
			}

			return &Predicate{
				Name:             "playerBool",
				Args:             spec.Args,
				Reads:            []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				Cost:             boardgame.LegalCostTrivial,
				EmittedTemplates: []string{template},
				Evaluate: func(ctx Context) Verdict {
					value, err := resolveBoolPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if value {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"prop": String(prop),
					})
				},
			}, nil
		},
	}
}
