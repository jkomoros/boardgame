package legal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/enum"
)

const (
	// TemplateComponentPropEquals is the default ComponentPropEquals failure.
	TemplateComponentPropEquals = "legal.component_prop_equals"
	// TemplateComponentPropNotEquals is the default ComponentPropNotEquals failure.
	TemplateComponentPropNotEquals = "legal.component_prop_not_equals"
	// TemplateComponentPropsEqual is the default ComponentPropsEqual failure.
	TemplateComponentPropsEqual = "legal.component_props_equal"
	// TemplateComponentPropsNotEqual is the default ComponentPropsNotEqual failure.
	TemplateComponentPropsNotEqual = "legal.component_props_not_equal"
)

type componentSelectorKind string

const (
	selectorFirstOccupied   componentSelectorKind = "first-occupied"
	selectorOccupiedOrdinal componentSelectorKind = "occupied-ordinal"
	selectorMoveIndex       componentSelectorKind = "move-index"
)

// ComponentSelector is a closed description of one component in a stack.
// Its fields are private so callers can only construct supported selectors.
// It serializes into LegalSpec.Args; it does not widen the legal path grammar.
type ComponentSelector struct {
	stackPath string
	wire      string
}

// FirstOccupied selects the first non-empty slot in stackPath.
func FirstOccupied(stackPath string) ComponentSelector {
	return ComponentSelector{stackPath: stackPath, wire: string(selectorFirstOccupied)}
}

// OccupiedOrdinal selects the ordinal-th occupied slot in stackPath, ignoring
// holes. ordinal is zero based.
func OccupiedOrdinal(stackPath string, ordinal int) ComponentSelector {
	return ComponentSelector{stackPath: stackPath, wire: string(selectorOccupiedOrdinal) + ":" + strconv.Itoa(ordinal)}
}

// MoveIndex selects the slot whose zero-based index is stored at moveIndexPath.
// The path must be a move.* int property and is validated at boot.
func MoveIndex(stackPath, moveIndexPath string) ComponentSelector {
	return ComponentSelector{stackPath: stackPath, wire: string(selectorMoveIndex) + ":" + moveIndexPath}
}

// ComponentPropEquals passes when selector's static string or enum component
// field equals value.
func ComponentPropEquals(selector ComponentSelector, field, value string) Spec {
	return Spec{Name: "componentPropEquals", Args: []string{selector.stackPath, selector.wire, field, value}}
}

// ComponentPropNotEquals is the definite inverse of ComponentPropEquals.
func ComponentPropNotEquals(selector ComponentSelector, field, value string) Spec {
	return Spec{Name: "componentPropNotEquals", Args: []string{selector.stackPath, selector.wire, field, value}}
}

// ComponentPropsEqual passes when the selected components expose equal values
// for the same static string or enum field.
func ComponentPropsEqual(left, right ComponentSelector, field string) Spec {
	return Spec{Name: "componentPropsEqual", Args: []string{left.stackPath, left.wire, right.stackPath, right.wire, field}}
}

// ComponentPropsNotEqual is the definite inverse of ComponentPropsEqual.
func ComponentPropsNotEqual(left, right ComponentSelector, field string) Spec {
	return Spec{Name: "componentPropsNotEqual", Args: []string{left.stackPath, left.wire, right.stackPath, right.wire, field}}
}

type parsedComponentSelector struct {
	stackPath string
	kind      componentSelectorKind
	ordinal   int
	indexPath string
}

func parseComponentSelector(stackPath, wire string) (parsedComponentSelector, error) {
	if stackPath == "" {
		return parsedComponentSelector{}, fmt.Errorf("stack path is empty")
	}
	switch {
	case wire == string(selectorFirstOccupied):
		return parsedComponentSelector{stackPath: stackPath, kind: selectorFirstOccupied}, nil
	case strings.HasPrefix(wire, string(selectorOccupiedOrdinal)+":"):
		raw := strings.TrimPrefix(wire, string(selectorOccupiedOrdinal)+":")
		ordinal, err := strconv.Atoi(raw)
		if err != nil || ordinal < 0 {
			return parsedComponentSelector{}, fmt.Errorf("occupied ordinal must be a non-negative integer, got %q", raw)
		}
		return parsedComponentSelector{stackPath: stackPath, kind: selectorOccupiedOrdinal, ordinal: ordinal}, nil
	case strings.HasPrefix(wire, string(selectorMoveIndex)+":"):
		path := strings.TrimPrefix(wire, string(selectorMoveIndex)+":")
		if !strings.HasPrefix(path, "move.") || len(path) == len("move.") {
			return parsedComponentSelector{}, fmt.Errorf("move index selector requires a move.* path, got %q", path)
		}
		return parsedComponentSelector{stackPath: stackPath, kind: selectorMoveIndex, indexPath: path}, nil
	default:
		return parsedComponentSelector{}, fmt.Errorf("unknown component selector %q", wire)
	}
}

func (s parsedComponentSelector) reads() []Read {
	reads := []Read{{Path: PropPath(s.stackPath), Facet: boardgame.LegalFacetValues}}
	if s.kind == selectorMoveIndex {
		reads = append(reads, Read{Path: PropPath(s.indexPath), Facet: boardgame.LegalFacetValues})
	}
	return reads
}

func (s parsedComponentSelector) requiredTypes() map[PropPath]boardgame.PropertyType {
	result := map[PropPath]boardgame.PropertyType{PropPath(s.stackPath): boardgame.TypeStack}
	if s.kind == selectorMoveIndex {
		result[PropPath(s.indexPath)] = boardgame.TypeInt
	}
	return result
}

func (s parsedComponentSelector) resolve(ctx Context) (boardgame.ImmutableComponentInstance, error) {
	stack, err := resolveStackPath(s.stackPath, ctx)
	if err != nil {
		return nil, err
	}
	if stack == nil {
		return nil, fmt.Errorf("stack %q was nil", s.stackPath)
	}
	switch s.kind {
	case selectorFirstOccupied:
		return stack.ImmutableFirst(), nil
	case selectorOccupiedOrdinal:
		seen := 0
		for i := 0; i < stack.Len(); i++ {
			component := stack.ImmutableComponentAt(i)
			if component == nil {
				continue
			}
			if seen == s.ordinal {
				return component, nil
			}
			seen++
		}
		return nil, nil
	case selectorMoveIndex:
		index, err := resolveIntPath(s.indexPath, ctx)
		if err != nil {
			return nil, err
		}
		if index < 0 || index >= stack.Len() {
			return nil, nil
		}
		return stack.ImmutableComponentAt(index), nil
	default:
		return nil, fmt.Errorf("unsupported component selector %q", s.kind)
	}
}

type componentScalar struct {
	typeOf boardgame.PropertyType
	text   string
	enum   enum.ImmutableVal
}

func readComponentScalar(component boardgame.ImmutableComponentInstance, field string) (componentScalar, error) {
	if component == nil || component.Values() == nil || component.Values().Reader() == nil {
		return componentScalar{}, fmt.Errorf("selected component has no values reader")
	}
	reader := component.Values().Reader()
	typeOf, ok := reader.Props()[field]
	if !ok {
		return componentScalar{}, fmt.Errorf("component field %q does not exist", field)
	}
	switch typeOf {
	case boardgame.TypeString:
		value, err := reader.StringProp(field)
		return componentScalar{typeOf: typeOf, text: value}, err
	case boardgame.TypeEnum:
		value, err := reader.ImmutableEnumProp(field)
		return componentScalar{typeOf: typeOf, enum: value}, err
	default:
		return componentScalar{}, fmt.Errorf("component field %q has unsupported PropertyType %v", field, typeOf)
	}
}

func (s componentScalar) display() string {
	if s.typeOf == boardgame.TypeEnum && s.enum != nil {
		return s.enum.String()
	}
	return s.text
}

func (s componentScalar) equalsLiteral(value string) (bool, error) {
	switch s.typeOf {
	case boardgame.TypeString:
		return s.text == value, nil
	case boardgame.TypeEnum:
		if s.enum == nil || s.enum.Enum() == nil {
			return false, fmt.Errorf("enum component field has no enum")
		}
		key := s.enum.Enum().ValueFromString(value)
		if key == enum.IllegalValue {
			return false, fmt.Errorf("%q is not a value in enum %q", value, s.enum.Enum().Name())
		}
		return s.enum.Value() == key, nil
	default:
		return false, fmt.Errorf("unsupported component scalar type %v", s.typeOf)
	}
}

func (s componentScalar) equals(other componentScalar) (bool, error) {
	if s.typeOf != other.typeOf {
		return false, fmt.Errorf("component scalar types differ (%v and %v)", s.typeOf, other.typeOf)
	}
	switch s.typeOf {
	case boardgame.TypeString:
		return s.text == other.text, nil
	case boardgame.TypeEnum:
		if s.enum == nil || other.enum == nil {
			return false, fmt.Errorf("enum component field was nil")
		}
		return s.enum.Equals(other.enum), nil
	default:
		return false, fmt.Errorf("unsupported component scalar type %v", s.typeOf)
	}
}

func mergeSelectorMetadata(selectors ...parsedComponentSelector) ([]Read, map[PropPath]boardgame.PropertyType) {
	var reads []Read
	seenReads := make(map[Read]bool)
	required := make(map[PropPath]boardgame.PropertyType)
	for _, selector := range selectors {
		for _, read := range selector.reads() {
			if !seenReads[read] {
				seenReads[read] = true
				reads = append(reads, read)
			}
		}
		for path, typeOf := range selector.requiredTypes() {
			required[path] = typeOf
		}
	}
	return reads, required
}

func componentPropLiteralConstructor(name, template string, negate bool) *PredicateConstructor {
	return &PredicateConstructor{Name: name, Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
		if len(spec.Args) != 4 {
			return nil, fmt.Errorf("legal: %s requires 4 args (stackPath, selector, field, value), got %d", name, len(spec.Args))
		}
		selector, err := parseComponentSelector(spec.Args[0], spec.Args[1])
		if err != nil {
			return nil, fmt.Errorf("legal: %s: %w", name, err)
		}
		field, want := spec.Args[2], spec.Args[3]
		if field == "" {
			return nil, fmt.Errorf("legal: %s requires a non-empty component field", name)
		}
		effectiveTemplate := spec.Message
		if effectiveTemplate == "" {
			effectiveTemplate = template
		}
		reads, required := mergeSelectorMetadata(selector)
		requiredValue := want
		return &Predicate{
			Name: name, Args: append([]string(nil), spec.Args...), Reads: reads, Cost: boardgame.LegalCostTrivial,
			RequiredReadTypes:       required,
			RequiredComponentFields: []boardgame.LegalComponentFieldRequirement{{StackPath: PropPath(selector.stackPath), Field: field, AllowedTypes: []boardgame.PropertyType{boardgame.TypeString, boardgame.TypeEnum}, RequiredEnumValue: &requiredValue}},
			ClientEvaluable:         false,
			EmittedTemplates:        []string{effectiveTemplate}, EmittedBindings: map[string][]string{effectiveTemplate: {"prop", "value", "want"}},
			Evaluate: func(ctx Context) Verdict {
				component, err := selector.resolve(ctx)
				if err != nil || component == nil {
					return UnknownVerdict(fmt.Sprintf("legal: %s could not select a component: %v", name, err))
				}
				actual, err := readComponentScalar(component, field)
				if err != nil {
					return UnknownVerdict("legal: " + name + ": " + err.Error())
				}
				matches, err := actual.equalsLiteral(want)
				if err != nil {
					return UnknownVerdict("legal: " + name + ": " + err.Error())
				}
				if matches != negate {
					return PassVerdict()
				}
				return FailT(effectiveTemplate, map[string]boardgame.LegalBindingValue{"prop": String(field), "value": String(actual.display()), "want": String(want)})
			},
		}, nil
	}}
}

func componentPropsPairConstructor(name, template string, negate bool) *PredicateConstructor {
	return &PredicateConstructor{Name: name, Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
		if len(spec.Args) != 5 {
			return nil, fmt.Errorf("legal: %s requires 5 args (leftStack, leftSelector, rightStack, rightSelector, field), got %d", name, len(spec.Args))
		}
		left, err := parseComponentSelector(spec.Args[0], spec.Args[1])
		if err != nil {
			return nil, fmt.Errorf("legal: %s left selector: %w", name, err)
		}
		right, err := parseComponentSelector(spec.Args[2], spec.Args[3])
		if err != nil {
			return nil, fmt.Errorf("legal: %s right selector: %w", name, err)
		}
		field := spec.Args[4]
		if field == "" {
			return nil, fmt.Errorf("legal: %s requires a non-empty component field", name)
		}
		effectiveTemplate := spec.Message
		if effectiveTemplate == "" {
			effectiveTemplate = template
		}
		reads, required := mergeSelectorMetadata(left, right)
		comparableTypeGroup := name + ":" + field
		requirements := []boardgame.LegalComponentFieldRequirement{
			{StackPath: PropPath(left.stackPath), Field: field, AllowedTypes: []boardgame.PropertyType{boardgame.TypeString, boardgame.TypeEnum}, ComparableTypeGroup: comparableTypeGroup},
		}
		if right.stackPath != left.stackPath {
			requirements = append(requirements, boardgame.LegalComponentFieldRequirement{StackPath: PropPath(right.stackPath), Field: field, AllowedTypes: []boardgame.PropertyType{boardgame.TypeString, boardgame.TypeEnum}, ComparableTypeGroup: comparableTypeGroup})
		}
		return &Predicate{
			Name: name, Args: append([]string(nil), spec.Args...), Reads: reads, Cost: boardgame.LegalCostTrivial,
			RequiredReadTypes: required, RequiredComponentFields: requirements, ClientEvaluable: false,
			EmittedTemplates: []string{effectiveTemplate}, EmittedBindings: map[string][]string{effectiveTemplate: {"prop", "left", "right"}},
			Evaluate: func(ctx Context) Verdict {
				leftComponent, err := left.resolve(ctx)
				if err != nil || leftComponent == nil {
					return UnknownVerdict(fmt.Sprintf("legal: %s could not select left component: %v", name, err))
				}
				rightComponent, err := right.resolve(ctx)
				if err != nil || rightComponent == nil {
					return UnknownVerdict(fmt.Sprintf("legal: %s could not select right component: %v", name, err))
				}
				leftValue, err := readComponentScalar(leftComponent, field)
				if err != nil {
					return UnknownVerdict("legal: " + name + ": " + err.Error())
				}
				rightValue, err := readComponentScalar(rightComponent, field)
				if err != nil {
					return UnknownVerdict("legal: " + name + ": " + err.Error())
				}
				matches, err := leftValue.equals(rightValue)
				if err != nil {
					return UnknownVerdict("legal: " + name + ": " + err.Error())
				}
				if matches != negate {
					return PassVerdict()
				}
				return FailT(effectiveTemplate, map[string]boardgame.LegalBindingValue{"prop": String(field), "left": String(leftValue.display()), "right": String(rightValue.display())})
			},
		}, nil
	}}
}

func componentPropEqualsConstructor() *PredicateConstructor {
	return componentPropLiteralConstructor("componentPropEquals", TemplateComponentPropEquals, false)
}

func componentPropNotEqualsConstructor() *PredicateConstructor {
	return componentPropLiteralConstructor("componentPropNotEquals", TemplateComponentPropNotEquals, true)
}

func componentPropsEqualConstructor() *PredicateConstructor {
	return componentPropsPairConstructor("componentPropsEqual", TemplateComponentPropsEqual, false)
}

func componentPropsNotEqualConstructor() *PredicateConstructor {
	return componentPropsPairConstructor("componentPropsNotEqual", TemplateComponentPropsNotEqual, true)
}
