package legal

import (
	"fmt"
	"strconv"

	"github.com/jkomoros/boardgame"
)

// Template keys for the count/emptiness predicates in this file.
const (
	// TemplateStackCount is the default Fail template key for StackCount.
	// Bindings: "count" (the stack's actual NumComponents()), "op" (the
	// comparison operator), "n" (the threshold).
	TemplateStackCount = "legal.stack_count"
	// TemplateStackEmpty is the default Fail template key for StackEmpty.
	// No bindings.
	TemplateStackEmpty = "legal.stack_empty"
	// TemplateStackNotEmpty is the default Fail template key for
	// StackNotEmpty. No bindings.
	TemplateStackNotEmpty = "legal.stack_not_empty"
)

// StackCount returns a Spec for the "stackCount" predicate: Passes if the
// stack at path's NumComponents() compares to n via op, which must be one
// of ==, !=, <, <=, >, >= (the same operator set PropCompare uses — see
// legalCompareOps in catalog_compare.go, shared rather than duplicated
// here).
func StackCount(path string, op string, n int) Spec {
	return Spec{Name: "stackCount", Args: []string{path, op, strconv.Itoa(n)}}
}

// StackEmpty returns a Spec for the "stackEmpty" predicate: Passes if the
// stack at path has zero components (NumComponents() == 0).
func StackEmpty(path string) Spec {
	return Spec{Name: "stackEmpty", Args: []string{path}}
}

// StackNotEmpty returns a Spec for the "stackNotEmpty" predicate: Passes if
// the stack at path has at least one component (NumComponents() > 0). The
// logical negation of StackEmpty, registered as its own named predicate
// (v1 has no general `not` compositor — see doc.go's v1 limits) rather than
// composed from it.
func StackNotEmpty(path string) Spec {
	return Spec{Name: "stackNotEmpty", Args: []string{path}}
}

// stackCountConstructor returns the registry entry for "stackCount".
func stackCountConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "stackCount",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 3 {
				return nil, fmt.Errorf("legal: stackCount requires 3 args (path, op, n), got %d", len(spec.Args))
			}
			path := spec.Args[0]
			op := spec.Args[1]
			cmp, ok := legalCompareOps[op]
			if !ok {
				return nil, fmt.Errorf("legal: stackCount: unknown op %q (expected one of ==, !=, <, <=, >, >=)", op)
			}
			n, err := strconv.Atoi(spec.Args[2])
			if err != nil {
				return nil, fmt.Errorf("legal: stackCount: arg 3 (n) must be an integer: %w", err)
			}

			template := spec.Message
			if template == "" {
				template = TemplateStackCount
			}

			return &Predicate{
				Name:              "stackCount",
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetCount}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeStack},
				Cost:              boardgame.LegalCostCheap,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"count", "op", "n"}},
				Evaluate: func(ctx Context) Verdict {
					stack, err := resolveStackPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if stack == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", path))
					}
					count := stack.NumComponents()
					if cmp(count, n) {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"count": Int(count),
						"op":    String(op),
						"n":     Int(n),
					})
				},
			}, nil
		},
	}
}

// stackEmptinessConstructor builds either "stackEmpty" or "stackNotEmpty",
// depending on wantEmpty: both share arg shape (path) and Reads, differing
// only in which side of NumComponents() == 0 counts as Pass.
func stackEmptinessConstructor(name string, template string, wantEmpty bool) *PredicateConstructor {
	return &PredicateConstructor{
		Name: name,
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 1 {
				return nil, fmt.Errorf("legal: %s requires 1 arg (path), got %d", name, len(spec.Args))
			}
			path := spec.Args[0]

			msgTemplate := spec.Message
			if msgTemplate == "" {
				msgTemplate = template
			}

			return &Predicate{
				Name:              name,
				ClientEvaluable:   true,
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: FacetNonEmpty}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeStack},
				Cost:              boardgame.LegalCostCheap,
				EmittedTemplates:  []string{msgTemplate},
				EmittedBindings:   map[string][]string{msgTemplate: nil},
				Evaluate: func(ctx Context) Verdict {
					stack, err := resolveStackPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if stack == nil {
						return UnknownVerdict(fmt.Sprintf("legal: stack path %q resolved to nil", path))
					}
					empty := stack.NumComponents() == 0
					if empty == wantEmpty {
						return PassVerdict()
					}
					return FailT(msgTemplate)
				},
			}, nil
		},
	}
}

func stackEmptyConstructor() *PredicateConstructor {
	return stackEmptinessConstructor("stackEmpty", TemplateStackEmpty, true)
}

func stackNotEmptyConstructor() *PredicateConstructor {
	return stackEmptinessConstructor("stackNotEmpty", TemplateStackNotEmpty, false)
}
