package boardgame

import (
	"math"
	"math/rand"
	"strconv"
)

//StatePolicy defines a sanitization policy for a State object. In particular,
//it defines a policy for the Game state, and a single, fixed policy for all
//Player states, and a policy for each deck whose components have Dynamic
//State. Each string returns the policy for the property with that name in
//that sub-state object. Properties with no corresponding policy are
//effectively PolicyNoOp for all groups.
type StatePolicy struct {
	Game                   SubStatePolicy
	Player                 SubStatePolicy
	DynamicComponentValues map[string]SubStatePolicy
}

//Policies apply to Groups of players. Groups with numbers 0 or above are
//defined in State.GroupMembership. There are two special groups: Self and
//Other.
const (
	//GroupSelf applies if the player the state is being prepared for is the
	//current PlayerState being transformed.
	GroupSelf = -1
	//GroupOther applies if the player the state is being prepared for is NOT
	//the current PlayerState being transformed.
	GroupOther = -2
	//GroupAll matches all players. It's useful for setting a restrictive
	//policy by default, that then some sub-groups relax by applying a less
	//restrictive policy.
	GroupAll = -3
)

//SubStatePolicy is a sanitization policy for a sub-part of a State, for
//example a Game or Player.
type SubStatePolicy map[string]GroupPolicy

//A group Santization policy represents all of the various policies that apply
//depending on whether the player we're preparing the state for is a member of
//the given group. To calculate the effective policy, we first collect all
//Policies that apply to the given player, based on their group membership,
//and then applied the *least* restrictive one.
type GroupPolicy map[int]Policy

//A sanitization policy reflects how to tranform a given State property when
//presenting to someone outside of the target group.
type Policy int

const (
	//Non sanitized. For non-group properties (e.g. strings, ints, bools), any
	//policy other than PolicyVisible or PolicyRandom is effectively
	//PolicyHidden.
	PolicyVisible Policy = iota
	//For groups (e.g. stacks, int slices), return a group that has the same
	//length. For all else, it's effectively PolicyHidden. In practice, stacks
	//will be set so that their NumComponents() is the same, but every
	//component that exists returns the GenericComponent.
	PolicyLen

	//For groups, PolicyNonEmpty will allow it to be observed that the stack's
	//NumComponents is either Empty (0 components) or non-empty (1
	//components). So for GrowableStacks, it will either have no components or
	//1 component. And for SizedStack, either all of the slots will be empty,
	//or the first slot will be non-empty. In all cases, the Component
	//present, if there is one, will be the deck's GenericComponent.
	PolicyNonEmpty

	//PolicyHidden returns effectively the zero value for the type. For
	//stacks, the deck it is, and the Size (for SizedStack) is set, but
	//nothing else is.
	PolicyHidden

	//PolicyRandom sets the property to a random but legal value to obscure
	//it. Not practically useful often, but is used in ComputedProperties to
	//ensure that people do not take an accidental dependency on a property
	//they didn't explicitly list in their dependencies.
	PolicyRandom

	//TODO: implement the other policies.
)

func (s *state) SanitizedForPlayer(playerIndex int) State {

	//If the playerIndex isn't an actuall player's index, just return self.
	if playerIndex < 0 || playerIndex >= len(s.players) {
		return s
	}

	policy := s.delegate.StateSanitizationPolicy()

	if policy == nil {
		policy = &StatePolicy{}
	}

	return s.sanitizedWithDefault(policy, playerIndex, PolicyVisible)
}

//sanitizedWithExceptions will return a Sanitized() State where properties
//that are not in the passed policy are treated as PolicyRandom. Useful in
//computing properties.
func (s *state) sanitizedWithDefault(policy *StatePolicy, playerIndex int, defaultPolicy Policy) State {

	sanitized := s.copy(true)

	//We need to figure out which components that have dynamicvalues are
	//visible after sanitizing game and player states. We'll have
	//sanitizeStateObj tell us which ones are visible, and which player's
	//state they're visible through, by accumulating the information in
	//visibleDyanmicComponents.
	visibleDynamicComponents := make(map[string]map[int]int)

	for deckName, _ := range s.dynamicComponentValues {
		visibleDynamicComponents[deckName] = make(map[int]int)
	}

	sanitizeStateObj(sanitized.game.ReadSetter(), policy.Game, -1, playerIndex, defaultPolicy, visibleDynamicComponents)

	playerStates := sanitized.players

	for i := 0; i < len(playerStates); i++ {
		sanitizeStateObj(playerStates[i].ReadSetter(), policy.Player, i, playerIndex, defaultPolicy, visibleDynamicComponents)
	}

	//Some of the DynamicComponentValues that were marked as visible might
	//have their own stacks with dynamic values that are visible, so we need
	//to go through and mark those, too..
	transativelyMarkDynamicComponentsAsVisible(sanitized.dynamicComponentValues, visibleDynamicComponents)

	//Now that all dynamic components are marked, we need to go through and
	//sanitize all of those objects according to the policy.
	sanitizeDynamicComponentValues(sanitized.dynamicComponentValues, visibleDynamicComponents, policy.DynamicComponentValues, playerIndex)

	return sanitized

}

//statePlayerIndex is the index of the PlayerState that we're working on (-1
//for Game). preparingForPlayerIndex is the index that we're preparing the
//overall santiized state for, as provied to
//GameManager.SanitizedStateForPlayer()
func sanitizeStateObj(readSetter PropertyReadSetter, policy SubStatePolicy, statePlayerIndex int, preparingForPlayerIndex int, defaultPolicy Policy, visibleDynamic map[string]map[int]int) {

	for propName, propType := range readSetter.Props() {
		prop, err := readSetter.Prop(propName)

		if err != nil {
			//TODO: shouldn't we return an error or something?
			continue
		}

		effectivePolicy := calculateEffectivePolicy(prop, propType, policy[propName], statePlayerIndex, preparingForPlayerIndex, defaultPolicy)

		if visibleDynamic != nil {
			if propType == TypeGrowableStack || propType == TypeSizedStack {
				if effectivePolicy == PolicyVisible {
					stackProp := prop.(Stack)
					if _, ok := visibleDynamic[stackProp.deck().Name()]; ok {
						for _, c := range stackProp.Components() {
							if c == nil {
								continue
							}
							visibleDynamic[c.Deck.Name()][c.DeckIndex] = statePlayerIndex
						}
					}

				}
			}
		}

		readSetter.SetProp(propName, applyPolicy(effectivePolicy, prop, propType))
	}

}

func transativelyMarkDynamicComponentsAsVisible(dynamicComponentValues map[string][]MutableDynamicComponentValues, visibleComponents map[string]map[int]int) {

	//All dynamic component values are hidden, except for ones that currently
	//reside in stacks that have resolved to being Visible based on this
	//current sanitization configuration. However, DynamicComponents may
	//themselves have stacks that reference other dynamic components. This
	//method effectively "spreads out" the visibility from visible dynamic
	//compoonents to other ones they point to.

	//TODO: TEST THIS!

	type workItem struct {
		deckName  string
		deckIndex int
	}

	var workItems []workItem

	//Fill the list of items to work through with all visible items.

	for deckName, visibleItems := range visibleComponents {
		for index, _ := range visibleItems {
			workItems = append(workItems, workItem{deckName, index})
		}
	}

	//We can't use range because we will be adding more items to it as we go.

	for i := 0; i < len(workItems); i++ {
		item := workItems[i]

		playerIndex := visibleComponents[item.deckName][item.deckIndex]
		values := dynamicComponentValues[item.deckName][item.deckIndex]

		reader := values.Reader()

		for propName, propType := range reader.Props() {
			if propType != TypeGrowableStack && propType != TypeSizedStack {
				continue
			}
			prop, err := reader.Prop(propName)

			if err != nil {
				continue
			}

			stackProp := prop.(Stack)

			if _, ok := dynamicComponentValues[stackProp.deck().Name()]; !ok {
				//This stack is for a deck that has no dynamic values, can skip.
				continue
			}

			//Ok, if we get to here then we have a stack with items in a deck that does have dynamic values.
			for _, c := range stackProp.Components() {
				if c == nil {
					continue
				}
				//There can't possibly be a collision because each component may only be in a single stack at a time.
				visibleComponents[c.Deck.Name()][c.DeckIndex] = playerIndex
				//Take note that there's another item to add to the queue to explore.
				workItems = append(workItems, workItem{c.Deck.Name(), c.DeckIndex})
			}
		}

	}
}

func sanitizeDynamicComponentValues(dynamicComponentValues map[string][]MutableDynamicComponentValues, visibleComponents map[string]map[int]int, dynamicPolicy map[string]SubStatePolicy, preparingForPlayerIndex int) {
	for name, slice := range dynamicComponentValues {

		visibleDynamicDeck := visibleComponents[name]

		for i, value := range slice {

			readSetter := value.ReadSetter()

			if player, visible := visibleDynamicDeck[i]; visible {

				sanitizeStateObj(readSetter, dynamicPolicy[name], player, preparingForPlayerIndex, PolicyVisible, nil)

			} else {
				//Make it a hidden

				for propName, propType := range readSetter.Props() {
					prop, err := readSetter.Prop(propName)

					if err != nil {
						continue
					}

					readSetter.SetProp(propName, applyPolicy(PolicyHidden, prop, propType))
				}
			}
		}
	}
}

func calculateEffectivePolicy(prop interface{}, propType PropertyType, policyGroup GroupPolicy, statePlayerIndex int, preparingForPlayerIndex int, defaultPolicy Policy) Policy {

	//We're going to collect all of the policies that apply.
	var applicablePolicies []Policy

	for group, policy := range policyGroup {
		policyApplies := false
		switch group {
		case GroupSelf:
			policyApplies = (statePlayerIndex == preparingForPlayerIndex)
		case GroupOther:
			policyApplies = (statePlayerIndex != preparingForPlayerIndex)
		case GroupAll:
			policyApplies = true
		default:
			//In the future we'll interrogate whether the given group index is
			//in the specified property at this point.
			panic("Unsupported policy group")
		}
		if policyApplies {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	//Now calculate the LEAST restrictive of the policies that apply.
	effectivePolicy := PolicyVisible
	if len(applicablePolicies) > 0 {
		effectivePolicy = Policy(math.MaxInt64)
		for _, policy := range applicablePolicies {
			if policy < effectivePolicy {
				effectivePolicy = policy
			}
		}
	} else {
		effectivePolicy = defaultPolicy
	}

	return effectivePolicy
}

func randomBool() bool {
	r := rand.Intn(2)
	if r == 0 {
		return false
	}
	return true
}

func randomInt() int {
	return rand.Int()
}

func randomGrowableStack(stack *GrowableStack) *GrowableStack {
	result := stack.Copy()

	indexes := make([]int, rand.Intn(16))

	for i, _ := range indexes {
		indexes[i] = emptyIndexSentinel
	}

	result.indexes = indexes

	return result
}

func randomSizedStack(stack *SizedStack) *SizedStack {
	result := stack.Copy()

	indexes := make([]int, len(result.indexes))

	for i, _ := range indexes {
		if randomBool() {
			indexes[i] = emptyIndexSentinel
		} else {
			indexes[i] = genericComponentSentinel
		}
	}

	result.indexes = indexes

	return result
}

func applyPolicy(policy Policy, input interface{}, propType PropertyType) interface{} {
	if policy == PolicyVisible {
		return input
	}

	if policy == PolicyRandom {
		switch propType {
		case TypeBool:
			return randomBool()
		case TypeInt:
			return randomInt()
		case TypeString:
			//Note: unlike the other random*() functions, this is defined in
			//game for the purposes of creating an ID. That's sufficient for
			//this use.
			return randomString(16)
		case TypeGrowableStack:
			return randomGrowableStack(input.(*GrowableStack))
		case TypeSizedStack:
			return randomSizedStack(input.(*SizedStack))
		default:
			panic("Unknown property type for policy random")
		}
	}

	//Go through the propTypes where everythign that's not PolicyVisible is
	//effectively PolicyHidden.

	switch propType {
	case TypeBool:
		return false
	case TypeInt:
		return 0
	case TypeString:
		return ""
	}

	//Now we're left with len-properties.

	stack := input.(Stack)

	stack.applySanitizationPolicy(policy)

	return input

}

func (g *GrowableStack) applySanitizationPolicy(policy Policy) {

	if policy == PolicyVisible {
		return
	}

	if policy == PolicyLen {

		indexes := make([]int, len(g.indexes))

		for i := 0; i < len(indexes); i++ {
			indexes[i] = genericComponentSentinel
		}

		g.indexes = indexes
		return
	}

	if policy == PolicyHidden {
		g.indexes = make([]int, 0)
		return
	}

	if policy == PolicyNonEmpty {
		if g.NumComponents() == 0 {
			g.indexes = make([]int, 0)
		} else {
			g.indexes = []int{genericComponentSentinel}
		}

		return
	}

	panic("Unknown sanitization policy" + strconv.Itoa(int(policy)))

}

func (s *SizedStack) applySanitizationPolicy(policy Policy) {

	if policy == PolicyVisible {
		return
	}

	if policy == PolicyLen {

		indexes := make([]int, len(s.indexes))

		for i := 0; i < len(indexes); i++ {
			if s.indexes[i] == emptyIndexSentinel {
				indexes[i] = emptyIndexSentinel
			} else {
				indexes[i] = genericComponentSentinel
			}
		}

		s.indexes = indexes

		return
	}

	if policy == PolicyHidden || policy == PolicyNonEmpty {

		hasComponents := s.NumComponents() > 0

		indexes := make([]int, len(s.indexes))
		for i := 0; i < len(indexes); i++ {
			indexes[i] = -1
		}
		s.indexes = indexes

		if policy == PolicyNonEmpty && hasComponents {
			s.indexes[0] = genericComponentSentinel
		}

		return
	}

	panic("Unknown sanitization policy" + strconv.Itoa(int(policy)))
}
