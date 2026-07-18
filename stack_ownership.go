package boardgame

import (
	"fmt"
	"sort"
)

// stackOwner records the canonical persisted location of one physical stack.
// current deliberately re-reads that location so a captured stack becomes
// stale as soon as its field or board is replaced.
type stackOwner struct {
	path    string
	current func() (Stack, error)
}

type mergedStackLocation struct {
	path  string
	stack MergedStack
}

func sortedPropertyNames(reader PropertyReader) []string {
	names := make([]string, 0, len(reader.Props()))
	for name := range reader.Props() {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *state) readersByPath() []struct {
	path   string
	reader PropertyReadSetter
} {
	result := []struct {
		path   string
		reader PropertyReadSetter
	}{{"Game", s.gameState.ReadSetter()}}

	for i, player := range s.playerStates {
		result = append(result, struct {
			path   string
			reader PropertyReadSetter
		}{fmt.Sprintf("Players[%d]", i), player.ReadSetter()})
	}

	deckNames := make([]string, 0, len(s.dynamicComponentValues))
	for deckName := range s.dynamicComponentValues {
		deckNames = append(deckNames, deckName)
	}
	sort.Strings(deckNames)
	for _, deckName := range deckNames {
		for i, value := range s.dynamicComponentValues[deckName] {
			result = append(result, struct {
				path   string
				reader PropertyReadSetter
			}{fmt.Sprintf("DynamicComponentValues[%s][%d]", deckName, i), value.ReadSetter()})
		}
	}
	return result
}

func (s *state) initializeStackOwners() error {
	owners := make(map[Stack]stackOwner)
	var merged []mergedStackLocation

	addOwner := func(stack Stack, owner stackOwner) error {
		if stack == nil {
			return fmt.Errorf("%s is nil", owner.path)
		}
		if existing, ok := owners[stack]; ok {
			return fmt.Errorf("physical stack is declared at both %s and %s", existing.path, owner.path)
		}
		if stack.state() != s {
			return fmt.Errorf("%s belongs to a different state", owner.path)
		}
		owners[stack] = owner
		return nil
	}

	for _, item := range s.readersByPath() {
		reader := item.reader
		for _, propName := range sortedPropertyNames(reader) {
			path := item.path + "." + propName
			switch reader.Props()[propName] {
			case TypeStack:
				value, err := reader.ImmutableStackProp(propName)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				if value == nil {
					return fmt.Errorf("%s is nil", path)
				}
				if view := value.MergedStack(); view != nil {
					merged = append(merged, mergedStackLocation{path, view})
					continue
				}
				if !reader.PropMutable(propName) {
					return fmt.Errorf("%s is a physical stack behind an immutable property; immutable stack properties must be views", path)
				}
				stack, ok := value.(Stack)
				if !ok {
					return fmt.Errorf("%s is not a physical mutable stack", path)
				}
				capturedReader, capturedName := reader, propName
				if err := addOwner(stack, stackOwner{path: path, current: func() (Stack, error) {
					current, err := capturedReader.ImmutableStackProp(capturedName)
					if err != nil {
						return nil, err
					}
					physical, ok := current.(Stack)
					if !ok || current.MergedStack() != nil {
						return nil, fmt.Errorf("property no longer contains a physical stack")
					}
					return physical, nil
				}}); err != nil {
					return err
				}

			case TypeBoard:
				value, err := reader.ImmutableBoardProp(propName)
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				if value == nil {
					return fmt.Errorf("%s is nil", path)
				}
				if !reader.PropMutable(propName) {
					return fmt.Errorf("%s is a physical board behind an immutable property; immutable board properties cannot own components", path)
				}
				board, ok := value.(Board)
				if !ok {
					return fmt.Errorf("%s is not a physical mutable board", path)
				}
				for i, stack := range board.Spaces() {
					spacePath := fmt.Sprintf("%s[%d]", path, i)
					capturedReader, capturedName, capturedIndex := reader, propName, i
					if err := addOwner(stack, stackOwner{path: spacePath, current: func() (Stack, error) {
						current, err := capturedReader.ImmutableBoardProp(capturedName)
						if err != nil {
							return nil, err
						}
						physical, ok := current.(Board)
						if !ok {
							return nil, fmt.Errorf("property no longer contains a physical board")
						}
						if capturedIndex < 0 || capturedIndex >= physical.Len() {
							return nil, fmt.Errorf("board no longer contains space %d", capturedIndex)
						}
						return physical.SpaceAt(capturedIndex), nil
					}}); err != nil {
						return err
					}
				}
			}
		}
	}

	s.stackOwners = owners
	for _, location := range merged {
		if err := location.stack.Valid(); err != nil {
			s.stackOwners = nil
			return fmt.Errorf("%s is an invalid merged stack: %w", location.path, err)
		}
		if err := s.validateMergedOwnerLeaves(location.path, location.stack, make(map[MergedStack]bool), make(map[Stack]string)); err != nil {
			s.stackOwners = nil
			return err
		}
	}
	return nil
}

func (s *state) validateMergedOwnerLeaves(path string, view MergedStack, visiting map[MergedStack]bool, leaves map[Stack]string) error {
	if visiting[view] {
		return fmt.Errorf("%s contains a cycle of merged stacks", path)
	}
	visiting[view] = true
	defer delete(visiting, view)

	for i, child := range view.ImmutableStacks() {
		childPath := fmt.Sprintf("%s[%d]", path, i)
		if child == nil {
			return fmt.Errorf("%s is nil", childPath)
		}
		if nested := child.MergedStack(); nested != nil {
			if err := s.validateMergedOwnerLeaves(childPath, nested, visiting, leaves); err != nil {
				return err
			}
			continue
		}
		physical, ok := child.(Stack)
		if !ok {
			return fmt.Errorf("%s is not a physical stack or merged view", childPath)
		}
		owner, ok := s.stackOwners[physical]
		if !ok {
			return fmt.Errorf("%s uses a backing stack that is not a declared attached owner", childPath)
		}
		if previous, duplicate := leaves[physical]; duplicate {
			return fmt.Errorf("%s repeats physical stack %s (first used at %s)", path, owner.path, previous)
		}
		leaves[physical] = childPath
	}
	return nil
}

func (s *state) validateStackAttachment(stack ImmutableStack) error {
	physical, err := requirePhysicalStack(stack, "stack")
	if err != nil {
		return err
	}
	if physical.state() != s {
		return fmt.Errorf("stack belongs to a different state")
	}
	// Some package-level primitive tests construct an internal *state directly.
	// Real states always initialize this registry in emptyState.
	if s.stackOwners == nil {
		return nil
	}
	owner, ok := s.stackOwners[physical]
	if !ok {
		return fmt.Errorf("stack is not an attached owner in this state")
	}
	current, err := owner.current()
	if err != nil {
		return fmt.Errorf("stack owner %s cannot be read: %w", owner.path, err)
	}
	if current != physical {
		return fmt.Errorf("stack is stale: owner location %s now contains a different stack", owner.path)
	}
	return nil
}

// ValidateStackAttachment reports whether stack is a current physical owner in
// state. It is primarily useful to ConfigurationValidator implementations in
// subpackages; it exposes no mutation authority or internal state pointer.
func ValidateStackAttachment(input ImmutableState, stack ImmutableStack) error {
	st, ok := input.(*state)
	if !ok || st == nil {
		return fmt.Errorf("state is not a framework state")
	}
	return st.validateStackAttachment(stack)
}

func rawStackIndexes(stack Stack) (deckName string, deck *Deck, indexes []int, sized bool, err error) {
	switch concrete := stack.(type) {
	case *growableStack:
		return concrete.deckName, concrete.deckPtr, concrete.indexes, false, nil
	case *sizedStack:
		return concrete.deckName, concrete.deckPtr, concrete.indexes, true, nil
	default:
		return "", nil, nil, false, fmt.Errorf("unsupported physical stack type %T", stack)
	}
}

func (s *state) validateCurrentMergedViews() error {
	for _, item := range s.readersByPath() {
		for _, propName := range sortedPropertyNames(item.reader) {
			if item.reader.Props()[propName] != TypeStack {
				continue
			}
			value, err := item.reader.ImmutableStackProp(propName)
			if err != nil {
				return fmt.Errorf("%s.%s: %w", item.path, propName, err)
			}
			if value == nil {
				return fmt.Errorf("%s.%s is nil", item.path, propName)
			}
			view := value.MergedStack()
			if view == nil {
				if !item.reader.PropMutable(propName) {
					return fmt.Errorf("%s.%s is a physical stack behind an immutable property", item.path, propName)
				}
				continue
			}
			if err := view.Valid(); err != nil {
				return fmt.Errorf("%s.%s is an invalid merged stack: %w", item.path, propName, err)
			}
			if err := s.validateMergedOwnerLeaves(item.path+"."+propName, view, make(map[MergedStack]bool), make(map[Stack]string)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *state) validateComponentConservation() error {
	if s.sanitized {
		return nil
	}
	if s.stackOwners == nil {
		return fmt.Errorf("state has no initialized stack-owner registry")
	}
	if err := s.validateCurrentMergedViews(); err != nil {
		return err
	}

	deckNames := s.manager.Chest().DeckNames()
	sort.Strings(deckNames)
	found := make(map[*Deck][]string, len(deckNames))
	for _, deckName := range deckNames {
		deck := s.manager.Chest().Deck(deckName)
		found[deck] = make([]string, len(deck.Components()))
	}

	owners := make([]struct {
		stack Stack
		owner stackOwner
	}, 0, len(s.stackOwners))
	for stack, owner := range s.stackOwners {
		owners = append(owners, struct {
			stack Stack
			owner stackOwner
		}{stack, owner})
	}
	sort.Slice(owners, func(i, j int) bool { return owners[i].owner.path < owners[j].owner.path })

	for _, item := range owners {
		if err := s.validateStackAttachment(item.stack); err != nil {
			return fmt.Errorf("%s: %w", item.owner.path, err)
		}
		deckName, deck, indexes, sized, err := rawStackIndexes(item.stack)
		if err != nil {
			return fmt.Errorf("%s: %w", item.owner.path, err)
		}
		expectedDeck := s.manager.Chest().Deck(deckName)
		if expectedDeck == nil {
			return fmt.Errorf("%s names unknown deck %q", item.owner.path, deckName)
		}
		if deck != expectedDeck {
			return fmt.Errorf("%s deck pointer does not match manager chest deck %q", item.owner.path, deckName)
		}

		for slot, index := range indexes {
			location := fmt.Sprintf("%s[%d]", item.owner.path, slot)
			if index == emptyIndexSentinel {
				if sized {
					continue
				}
				return fmt.Errorf("%s contains the empty-slot sentinel in a growable stack", location)
			}
			if index == genericComponentSentinel {
				return fmt.Errorf("%s contains a generic component in an unsanitized state", location)
			}
			if index < 0 || index >= len(found[deck]) {
				return fmt.Errorf("%s contains invalid component index %d for deck %q", location, index, deckName)
			}
			if previous := found[deck][index]; previous != "" {
				return fmt.Errorf("deck %q component %d appears at both %s and %s", deckName, index, previous, location)
			}
			found[deck][index] = location
		}
	}

	for _, deckName := range deckNames {
		deck := s.manager.Chest().Deck(deckName)
		for index, location := range found[deck] {
			if location == "" {
				return fmt.Errorf("deck %q component %d is missing from every attached stack", deckName, index)
			}
		}
	}
	return nil
}
