package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"hash/fnv"
	"math"
	"math/rand"
	"strconv"
	"strings"
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

//sanitizationTransformation contains which policy to apply for every property
//in the state. Missing properties will be treated as PolicyVisible.
type sanitizationTransformation struct {
	Game                   subStateSanitizationTransformation
	Players                []subStateSanitizationTransformation
	DynamicComponentValues map[string]subStateSanitizationTransformation
}

//Map of policy to apply for each propname in this sub-state
type subStateSanitizationTransformation map[string]Policy

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
	//length, and whose Ids() represents the identity of the items. In
	//practice, stacks will be set so that their NumComponents() is the same,
	//but every component that exists returns the GenericComponent. This
	//policy is similar to Len, but allows observers to keep track of the
	//identity of cards as they are reordered in the stack.
	PolicyOrder

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

	//PolicyInvalid is not a valid Policy. It can be provided to signal an
	//illegal policy, which will cause the sanitization policy pipeline to
	//error.
	PolicyInvalid

	//TODO: implement the other policies.
)

func groupFromString(groupName string) int {

	//TODO: this will have to do somethign different when we support user-
	//defined groups.

	groupName = strings.ToLower(groupName)
	groupName = strings.TrimSpace(groupName)
	switch groupName {
	case "all":
		return GroupAll
	case "other":
		return GroupOther
	case "self":
		return GroupSelf
	}
	return 0
}

func policyFromString(policyName string) Policy {
	policyName = strings.ToLower(policyName)
	policyName = strings.TrimSpace(policyName)

	switch policyName {
	case "visible":
		return PolicyVisible
	case "order":
		return PolicyOrder
	case "len":
		return PolicyLen
	case "nonempty":
		return PolicyNonEmpty
	case "hidden":
		return PolicyHidden
	case "random":
		return PolicyRandom
	}
	return PolicyInvalid
}

func (s *state) SanitizedForPlayer(player PlayerIndex) State {

	//If the playerIndex isn't an actuall player's index, just return self.
	if player < -1 || int(player) >= len(s.playerStates) {
		return s
	}

	transformation := s.generateSanitizationTransformation(player)

	sanitized, err := s.applySanitizationTransformation(transformation)

	if err != nil {
		s.game.manager.Logger().Error("Couldn't sanitize for player: " + err.Error())
		return nil
	}

	return sanitized
}

func (s *state) generateSanitizationTransformation(player PlayerIndex) *sanitizationTransformation {

	delegate := s.game.manager.delegate

	result := &sanitizationTransformation{}

	result.Game = generateSubStateSanitizationTransformation(s.GameState(),
		StatePropertyRef{
			Group: StateGroupGame,
		}, delegate, player, -1)

	result.Players = make([]subStateSanitizationTransformation, len(s.PlayerStates()))

	for i, playerState := range s.PlayerStates() {
		result.Players[i] = generateSubStateSanitizationTransformation(playerState, StatePropertyRef{
			Group: StateGroupPlayer,
		}, delegate, player, PlayerIndex(i))
	}

	result.DynamicComponentValues = make(map[string]subStateSanitizationTransformation)

	for deckName, deckValues := range s.DynamicComponentValues() {
		if len(deckValues) == 0 {
			return nil
		}
		result.DynamicComponentValues[deckName] = generateSubStateSanitizationTransformation(deckValues[0], StatePropertyRef{
			Group:    StateGroupDynamicComponentValues,
			DeckName: deckName,
		}, delegate, player, -1)
	}

	return result

}

func generateSubStateSanitizationTransformation(subState SubState, propertyRef StatePropertyRef, delegate GameDelegate, generatingForPlayer PlayerIndex, index PlayerIndex) subStateSanitizationTransformation {

	//Since propertyRef is passed in by value we can modify it locally without a problem

	result := make(subStateSanitizationTransformation)

	for propName, _ := range subState.Reader().Props() {
		propertyRef.PropName = propName

		//Initalize it for GroupAll, and either GroupSelf or GroupOther
		groupMembership := make(map[int]bool, 2)

		groupMembership[GroupAll] = true

		if propertyRef.Group == StateGroupPlayer {
			if generatingForPlayer == index {
				groupMembership[GroupSelf] = true
			} else {
				groupMembership[GroupOther] = true
			}
		}

		result[propName] = delegate.SanitizationPolicy(propertyRef, groupMembership)

	}

	return result

}

//sanitizedWithExceptions will return a Sanitized() State where properties
//that are not in the passed policy are treated as PolicyRandom. Useful in
//computing properties.
func (s *state) applySanitizationTransformation(transformation *sanitizationTransformation) (State, error) {

	sanitized := s.copy(true)

	if len(transformation.Players) != len(s.PlayerStates()) {
		return nil, errors.New("The transformation did not have a record for each player state.")
	}

	//We need to figure out which components that have dynamicvalues are
	//visible after sanitizing game and player states. We'll have
	//sanitizeStateObj tell us which ones are visible, and which player's
	//state they're visible through, by accumulating the information in
	//visibleDyanmicComponents.
	visibleDynamicComponents := make(map[string]map[int]bool)

	for deckName, _ := range s.dynamicComponentValues {
		visibleDynamicComponents[deckName] = make(map[int]bool)
	}

	err := sanitizeStateObj(sanitized.gameState.ReadSetter(), transformation.Game, visibleDynamicComponents)

	if err != nil {
		return nil, errors.Extend(err, "Couldn't sanitize game state")
	}

	playerStates := sanitized.playerStates

	for i := 0; i < len(playerStates); i++ {
		err = sanitizeStateObj(playerStates[i].ReadSetter(), transformation.Players[i], visibleDynamicComponents)
		if err != nil {
			return nil, errors.Extend(err, "Couldn't sanitize player state number "+strconv.Itoa(i))
		}
	}

	//Some of the DynamicComponentValues that were marked as visible might
	//have their own stacks with dynamic values that are visible, so we need
	//to go through and mark those, too..
	transativelyMarkDynamicComponentsAsVisible(sanitized.dynamicComponentValues, visibleDynamicComponents)

	//Now that all dynamic components are marked, we need to go through and
	//sanitize all of those objects according to the policy.

	if err := sanitizeDynamicComponentValues(sanitized.dynamicComponentValues, visibleDynamicComponents, transformation.DynamicComponentValues); err != nil {
		return nil, errors.Extend(err, "Couldn't sanitize dyanmic component values")
	}

	return sanitized, nil

}

//sanitizedWithExceptions will return a Sanitized() State where properties
//that are not in the passed policy are treated as PolicyRandom. Useful in
//computing properties.
func (s *state) deprecatedSanitizedWithDefault(policy *StatePolicy, playerIndex PlayerIndex, defaultPolicy Policy) (State, error) {

	sanitized := s.copy(true)

	//We need to figure out which components that have dynamicvalues are
	//visible after sanitizing game and player states. We'll have
	//sanitizeStateObj tell us which ones are visible, and which player's
	//state they're visible through, by accumulating the information in
	//visibleDyanmicComponents.
	visibleDynamicComponents := make(map[string]map[int]PlayerIndex)

	for deckName, _ := range s.dynamicComponentValues {
		visibleDynamicComponents[deckName] = make(map[int]PlayerIndex)
	}

	err := deprecatedSanitizeStateObj(sanitized.gameState.ReadSetter(), policy.Game, AdminPlayerIndex, playerIndex, defaultPolicy, visibleDynamicComponents)

	if err != nil {
		return nil, errors.Extend(err, "Couldn't sanitize game state")
	}

	playerStates := sanitized.playerStates

	for i := 0; i < len(playerStates); i++ {
		err = deprecatedSanitizeStateObj(playerStates[i].ReadSetter(), policy.Player, PlayerIndex(i), playerIndex, defaultPolicy, visibleDynamicComponents)
		if err != nil {
			return nil, errors.Extend(err, "Couldn't sanitize player state number "+strconv.Itoa(i))
		}
	}

	//Some of the DynamicComponentValues that were marked as visible might
	//have their own stacks with dynamic values that are visible, so we need
	//to go through and mark those, too..
	deprecatedTransativelyMarkDynamicComponentsAsVisible(sanitized.dynamicComponentValues, visibleDynamicComponents)

	//Now that all dynamic components are marked, we need to go through and
	//sanitize all of those objects according to the policy.

	shouldRandomizeCompontentValues := false

	if defaultPolicy == PolicyRandom {
		shouldRandomizeCompontentValues = true
	}

	if err := deprecatedSanitizeDynamicComponentValues(sanitized.dynamicComponentValues, visibleDynamicComponents, policy.DynamicComponentValues, playerIndex, shouldRandomizeCompontentValues); err != nil {
		return nil, errors.Extend(err, "Couldn't sanitize dyanmic component values")
	}

	return sanitized, nil

}

func (s *StatePolicy) valid() error {

	if s == nil {
		return nil
	}

	//Sholud these helpers just be methods on StatePolicy and its children?

	if s.Game != nil {
		if err := s.Game.valid(); err != nil {
			return errors.Extend(err, "Game policy was invalid")
		}
	}

	if s.Player != nil {
		if err := s.Player.valid(); err != nil {
			return errors.Extend(err, "Player policy was invalid")
		}
	}

	if s.DynamicComponentValues != nil {
		for name, subPolicy := range s.DynamicComponentValues {
			if err := subPolicy.valid(); err != nil {
				return errors.Extend(err, "DynamicComponentValues policy for "+name+"was invalid")
			}
		}
	}

	return nil

}

func (s SubStatePolicy) valid() error {
	if s == nil {
		return nil
	}

	for prop, group := range s {
		if err := group.valid(); err != nil {
			return errors.Extend(err, "Property "+prop+" was invalid")
		}
	}

	return nil
}

func (g GroupPolicy) valid() error {
	if g == nil {
		return nil
	}

	for groupNumber, policy := range g {

		switch groupNumber {
		case GroupAll:
			//Fine
		case GroupSelf:
			//Fine
		case GroupOther:
			//Fine
		default:
			return errors.New("GroupNumber was not legal: " + strconv.Itoa(groupNumber))
		}

		switch policy {
		case PolicyVisible:
			//Fine
		case PolicyOrder:
			//Fine
		case PolicyLen:
			//Fine
		case PolicyNonEmpty:
		//Fine
		case PolicyHidden:
			//Fine
		case PolicyRandom:
			//Fine
		default:
			return errors.New("Policy was not legal: " + strconv.Itoa(int(policy)))
		}

	}

	return nil
}

//statePlayerIndex is the index of the PlayerState that we're working on (-1
//for Game). preparingForPlayerIndex is the index that we're preparing the
//overall santiized state for, as provied to
//GameManager.SanitizedStateForPlayer()
func sanitizeStateObj(readSetter PropertyReadSetter, transformation subStateSanitizationTransformation, visibleDynamic map[string]map[int]bool) error {

	for propName, propType := range readSetter.Props() {
		prop, err := readSetter.Prop(propName)

		if err != nil {
			return errors.Extend(err, propName+" had an error")
		}

		policy := transformation[propName]

		if policy == PolicyInvalid {
			return errors.New("Effective policy computed to PolicyInvalid")
		}

		if visibleDynamic != nil {
			if propType == TypeGrowableStack || propType == TypeSizedStack {
				if policy == PolicyVisible {
					stackProp := prop.(Stack)
					if _, ok := visibleDynamic[stackProp.deck().Name()]; ok {
						for _, c := range stackProp.Components() {
							if c == nil {
								continue
							}
							visibleDynamic[c.Deck.Name()][c.DeckIndex] = true
						}
					}

				}
			}
		}

		readSetter.SetProp(propName, applyPolicy(policy, prop, propType))
	}

	return nil

}

//statePlayerIndex is the index of the PlayerState that we're working on (-1
//for Game). preparingForPlayerIndex is the index that we're preparing the
//overall santiized state for, as provied to
//GameManager.SanitizedStateForPlayer()
func deprecatedSanitizeStateObj(readSetter PropertyReadSetter, policy SubStatePolicy, statePlayerIndex PlayerIndex, preparingForPlayerIndex PlayerIndex, defaultPolicy Policy, visibleDynamic map[string]map[int]PlayerIndex) error {

	for propName, propType := range readSetter.Props() {
		prop, err := readSetter.Prop(propName)

		if err != nil {
			return errors.Extend(err, propName+" had an error")
		}

		effectivePolicy, err := deprecatedCalculateEffectivePolicy(prop, propType, policy[propName], statePlayerIndex, preparingForPlayerIndex, defaultPolicy)

		if err != nil {
			return errors.Extend(err, "Couldn't calculate effective policy")
		}

		if effectivePolicy == PolicyInvalid {
			return errors.New("Effective policy computed to PolicyInvalid")
		}

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

	return nil

}

func transativelyMarkDynamicComponentsAsVisible(dynamicComponentValues map[string][]MutableSubState, visibleComponents map[string]map[int]bool) {

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
				visibleComponents[c.Deck.Name()][c.DeckIndex] = true
				//Take note that there's another item to add to the queue to explore.
				workItems = append(workItems, workItem{c.Deck.Name(), c.DeckIndex})
			}
		}

	}
}

func deprecatedTransativelyMarkDynamicComponentsAsVisible(dynamicComponentValues map[string][]MutableSubState, visibleComponents map[string]map[int]PlayerIndex) {

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

func sanitizeDynamicComponentValues(dynamicComponentValues map[string][]MutableSubState, visibleComponents map[string]map[int]bool, transformation map[string]subStateSanitizationTransformation) error {

	for name, slice := range dynamicComponentValues {

		visibleDynamicDeck := visibleComponents[name]

		for i, value := range slice {

			readSetter := value.ReadSetter()

			if _, visible := visibleDynamicDeck[i]; visible {

				if err := sanitizeStateObj(readSetter, transformation[name], nil); err != nil {
					return errors.Extend(err, "Couldn't sanitize random dynamic component")
				}

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
	return nil

}

func deprecatedSanitizeDynamicComponentValues(dynamicComponentValues map[string][]MutableSubState, visibleComponents map[string]map[int]PlayerIndex, dynamicPolicy map[string]SubStatePolicy, preparingForPlayerIndex PlayerIndex, isRandom bool) error {

	for name, slice := range dynamicComponentValues {

		visibleDynamicDeck := visibleComponents[name]

		for i, value := range slice {

			readSetter := value.ReadSetter()

			if player, visible := visibleDynamicDeck[i]; visible {

				//The fact that we do such different things here seems like a bug in how we've structed these methods?
				if isRandom {
					if err := deprecatedSanitizeStateObj(readSetter, dynamicPolicy[name], player, preparingForPlayerIndex, PolicyRandom, nil); err != nil {
						return errors.Extend(err, "Couldn't sanitize random dynamic component")
					}
				} else {
					if err := deprecatedSanitizeStateObj(readSetter, dynamicPolicy[name], player, preparingForPlayerIndex, PolicyVisible, nil); err != nil {
						return errors.Extend(err, "Couldn't sanitize non-random dynamic component")
					}
				}

			} else {
				//Make it a hidden

				for propName, propType := range readSetter.Props() {
					prop, err := readSetter.Prop(propName)

					if err != nil {
						continue
					}

					if isRandom {
						readSetter.SetProp(propName, applyPolicy(PolicyRandom, prop, propType))
					} else {
						readSetter.SetProp(propName, applyPolicy(PolicyHidden, prop, propType))
					}
				}
			}
		}
	}
	return nil

}

func deprecatedCalculateEffectivePolicy(prop interface{}, propType PropertyType, policyGroup GroupPolicy, statePlayerIndex PlayerIndex, preparingForPlayerIndex PlayerIndex, defaultPolicy Policy) (Policy, error) {

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
			return PolicyVisible, errors.New("Unknown policy: " + strconv.Itoa(int(policy)))
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

	return effectivePolicy, nil
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

func randomIntSlice(length int) []int {
	result := make([]int, rand.Intn(length))

	for i := 0; i < len(result); i++ {
		result[i] = randomInt()
	}

	return result
}

func randomBoolSlice(length int) []bool {
	result := make([]bool, rand.Intn(length))

	for i := 0; i < len(result); i++ {
		result[i] = randomBool()
	}

	return result
}

func randomStringSlice(length int) []string {
	result := make([]string, rand.Intn(length))

	for i := 0; i < len(result); i++ {
		result[i] = randomString(16)
	}

	return result
}

func randomPlayerIndexSlice(length int) []PlayerIndex {
	result := make([]PlayerIndex, rand.Intn(length))

	for i := 0; i < len(result); i++ {
		//TODO: ideally we'd actually return a random player index that is
		//valid given the size of hte game.
		result[i] = 0
	}

	return result
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

func randomTimer() *Timer {
	//TODO: actually set some of the fields randomly
	return NewTimer()
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
		case TypePlayerIndex:
			//TODO: ideally we'd return a legitimately random playerIndex. But
			//down here we don't know what the legal range is.
			return 0
		case TypeEnumVar:
			e := input.(enum.Var).CopyVar()
			e.SetValue(e.Enum().RandomValue())
			return e
		case TypeEnumConst:
			e := input.(enum.Const).Copy()
			res, _ := e.Enum().NewConst(e.Enum().RandomValue())
			return res
		case TypeIntSlice:
			return randomIntSlice(5)
		case TypeBoolSlice:
			return randomBoolSlice(5)
		case TypeStringSlice:
			return randomStringSlice(5)
		case TypePlayerIndexSlice:
			return randomPlayerIndexSlice(5)
		case TypeGrowableStack:
			return randomGrowableStack(input.(*GrowableStack))
		case TypeSizedStack:
			return randomSizedStack(input.(*SizedStack))
		case TypeTimer:
			return randomTimer()
		default:
			//Hmm, it's an unknwon type I guess. This shouldn't happen!
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
	case TypePlayerIndex:
		return 0
	case TypeTimer:
		return NewTimer()
	case TypeEnumVar:
		e := input.(enum.Var).CopyVar()
		e.SetValue(e.Enum().DefaultValue())
		return e
	case TypeEnumConst:
		e := input.(enum.Const).Copy()
		res, _ := e.Enum().NewConst(e.Enum().DefaultValue())
		return res
	}

	//Now the ones that are non-stack containers
	switch propType {
	case TypeIntSlice:
		return applySanitizationPolicyIntSlice(policy, input.([]int))
	case TypeBoolSlice:
		return applySanitizationPolicyBoolSlice(policy, input.([]bool))
	case TypeStringSlice:
		return applySanitizationPolicyStringSlice(policy, input.([]string))
	case TypePlayerIndexSlice:
		return applySanitizationPolicyPlayerIndexSlice(policy, input.([]PlayerIndex))
	}

	//Now we're left with len-properties.

	stack := input.(Stack)

	stack.applySanitizationPolicy(policy)

	return input

}

func applySanitizationPolicyIntSlice(policy Policy, input []int) []int {
	if policy == PolicyVisible {
		return input
	}

	if policy == PolicyLen || policy == PolicyOrder {
		return make([]int, len(input))
	}

	if policy == PolicyNonEmpty {
		if len(input) > 0 {
			return make([]int, 1)
		}
		return make([]int, 0)
	}

	//if we get to here it's either PolicyHidden, or an unknown policy. If the
	//latter, it's better to fail by being restrictive.
	return make([]int, 0)
}

func applySanitizationPolicyBoolSlice(policy Policy, input []bool) []bool {
	if policy == PolicyVisible {
		return input
	}

	if policy == PolicyLen || policy == PolicyOrder {
		return make([]bool, len(input))
	}

	if policy == PolicyNonEmpty {
		if len(input) > 0 {
			return make([]bool, 1)
		}
		return make([]bool, 0)
	}

	//if we get to here it's either PolicyHidden, or an unknown policy. If the
	//latter, it's better to fail by being restrictive.
	return make([]bool, 0)
}

func applySanitizationPolicyStringSlice(policy Policy, input []string) []string {
	if policy == PolicyVisible {
		return input
	}

	if policy == PolicyLen || policy == PolicyOrder {
		return make([]string, len(input))
	}

	if policy == PolicyNonEmpty {
		if len(input) > 0 {
			return make([]string, 1)
		}
		return make([]string, 0)
	}

	//if we get to here it's either PolicyHidden, or an unknown policy. If the
	//latter, it's better to fail by being restrictive.
	return make([]string, 0)

}

func applySanitizationPolicyPlayerIndexSlice(policy Policy, input []PlayerIndex) []PlayerIndex {
	if policy == PolicyVisible {
		return input
	}

	if policy == PolicyLen || policy == PolicyOrder {
		return make([]PlayerIndex, len(input))
	}

	if policy == PolicyNonEmpty {
		if len(input) > 0 {
			return make([]PlayerIndex, 1)
		}
		return make([]PlayerIndex, 0)
	}

	//if we get to here it's either PolicyHidden, or an unknown policy. If the
	//latter, it's better to fail by being restrictive.
	return make([]PlayerIndex, 0)
}

func (g *GrowableStack) applySanitizationPolicy(policy Policy) {

	if policy == PolicyVisible {
		return
	}

	if policy == PolicyLen || policy == PolicyOrder {

		//Keep Ids before we blank-out components, but put them in a random
		//order.
		g.overrideIds = make([]string, len(g.indexes))

		for i, c := range g.Components() {
			if c == nil {
				continue
			}
			g.overrideIds[i] = c.Id(g.state())
		}

		if policy == PolicyLen {
			perm := randPermForStack(g)
			shuffledIds := make([]string, len(g.overrideIds))
			for i, j := range perm {
				shuffledIds[i] = g.overrideIds[j]
			}
			g.overrideIds = shuffledIds
		}

		indexes := make([]int, len(g.indexes))

		for i := 0; i < len(indexes); i++ {
			indexes[i] = genericComponentSentinel
		}

		g.indexes = indexes
		return
	}

	//Anything other than PolicyVisible and PolicyLen (at least currently)
	//will move Ids to PossibleIds.
	for _, c := range g.Components() {
		if c == nil {
			continue
		}
		id := c.Id(g.statePtr)
		g.idSeen(id)
	}

	if policy == PolicyNonEmpty {
		if g.NumComponents() == 0 {
			g.indexes = make([]int, 0)
		} else {
			g.indexes = []int{genericComponentSentinel}
		}

		return
	}

	//if we get to here it's either PolicyHidden, or an unknown policy. If the
	//latter, it's better to fail by being restrictive.
	g.indexes = make([]int, 0)
	g.idsLastSeen = make(map[string]int)
	return

}

//returns a random permutation of size stack.Len(). The permutation will be
//predictable given this exact stack and its state, but unpredictable in
//general. This makes it give predictable results for testing but still be
//unguessable if you don't have the stack's game's SecretSalt.
func randPermForStack(stack Stack) []int {

	//TODO: we really only do this in order to have straight-forward testing
	//via golden json blobs. That feels like the wrong trade-off...

	seedStr := stack.state().game.SecretSalt() + strconv.Itoa(stack.state().Version())

	h := fnv.New64()
	h.Write([]byte(seedStr))
	seed := h.Sum64()

	r := rand.New(rand.NewSource(int64(seed)))

	return r.Perm(stack.Len())

}

func (s *SizedStack) applySanitizationPolicy(policy Policy) {

	if policy == PolicyVisible {
		return
	}

	if policy == PolicyLen || policy == PolicyOrder {

		//Keep Ids before we blank-out components, but put them in a random
		//order.
		s.overrideIds = make([]string, len(s.indexes))

		for i, c := range s.Components() {
			if c == nil {
				continue
			}
			s.overrideIds[i] = c.Id(s.state())
		}

		if policy == PolicyLen {
			perm := randPermForStack(s)
			shuffledIds := make([]string, len(s.overrideIds))
			for i, j := range perm {
				shuffledIds[i] = s.overrideIds[j]
			}
			s.overrideIds = shuffledIds
		}

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

	//Anything other than PolicyVisible and PolicyLen (at least currently)
	//will move Ids to PossibleIds.
	for _, c := range s.Components() {
		if c == nil {
			continue
		}
		id := c.Id(s.statePtr)
		s.idSeen(id)
	}

	//if we get to here it's either PolicyHidden, PolicyNonEmpty or an unknown
	//policy. If the latter, it's better to fail by being restrictive.

	hasComponents := s.NumComponents() > 0

	indexes := make([]int, len(s.indexes))
	for i := 0; i < len(indexes); i++ {
		indexes[i] = -1
	}
	s.indexes = indexes

	if policy == PolicyNonEmpty && hasComponents {
		s.indexes[0] = genericComponentSentinel
	}

	if policy == PolicyHidden {
		s.idsLastSeen = make(map[string]int)
	}

	return
}
