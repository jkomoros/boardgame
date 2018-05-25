package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"
)

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

//A sanitization policy reflects how to tranform a given State property when
//presenting to someone outside of the target group.
type Policy int

const (
	//Non sanitized. For non-group properties (e.g. strings, ints, bools), any
	//policy other than PolicyVisible is effectively PolicyHidden.
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
	//components). So for default Stacks, it will either have no components or
	//1 component. And for SizedStack, either all of the slots will be empty,
	//or the first slot will be non-empty. In all cases, the Component
	//present, if there is one, will be the deck's GenericComponent.
	PolicyNonEmpty

	//PolicyHidden returns effectively the zero value for the type. For
	//stacks, the deck it is, and the Size (for SizedStack) is set, but
	//nothing else is.
	PolicyHidden

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

//generateSanitizationTransformation creates a sanitizationTransformation by
//consulting the delegate for each property on each sub-state.
func (s *state) generateSanitizationTransformation(player PlayerIndex) *sanitizationTransformation {

	delegate := s.game.manager.delegate

	result := &sanitizationTransformation{}

	ref := NewStatePropertyRef()
	ref.Group = StateGroupGame

	result.Game = generateSubStateSanitizationTransformation(s.GameState(),
		ref, delegate, player, -1)

	result.Players = make([]subStateSanitizationTransformation, len(s.PlayerStates()))

	for i, playerState := range s.PlayerStates() {
		ref := NewStatePropertyRef()
		ref.Group = StateGroupPlayer
		result.Players[i] = generateSubStateSanitizationTransformation(playerState, ref, delegate, player, PlayerIndex(i))
	}

	result.DynamicComponentValues = make(map[string]subStateSanitizationTransformation)

	for deckName, deckValues := range s.DynamicComponentValues() {
		if len(deckValues) == 0 {
			return nil
		}
		ref := NewStatePropertyRef()
		ref.Group = StateGroupDynamicComponentValues
		ref.DeckName = deckName
		result.DynamicComponentValues[deckName] = generateSubStateSanitizationTransformation(deckValues[0], ref, delegate, player, -1)
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

//applySanitizationTransformation takes a generated sanitizationTransformation
//and applies it to the given tate, returning a new state that has been
//transformed accordingly. The DynamicComponentValues transformations are set
//to Hidden (instead of how they are configured) unless the stacks that
//contain them in Game and Player states resolve to PolicyVisible.
func (s *state) applySanitizationTransformation(transformation *sanitizationTransformation) (State, error) {

	sanitized, err := s.copy(true)

	if err != nil {
		return nil, errors.New("Couldn't copy state: " + err.Error())
	}

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

	err = sanitizeStateObj(sanitized.gameState.ReadSetConfigurer(), transformation.Game, visibleDynamicComponents)

	if err != nil {
		return nil, errors.Extend(err, "Couldn't sanitize game state")
	}

	playerStates := sanitized.playerStates

	for i := 0; i < len(playerStates); i++ {
		err = sanitizeStateObj(playerStates[i].ReadSetConfigurer(), transformation.Players[i], visibleDynamicComponents)
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

//sanitizeStateObj applies the given sanitizationTransformation to the given
//sub-state. It also keeps track of which components within it resolve to
//PolicyVisible, so later that information can be used to only reveal that
//information in DynamicComponentValues if the components they're related to
//were visible.
func sanitizeStateObj(readSetConfigurer PropertyReadSetConfigurer, transformation subStateSanitizationTransformation, visibleDynamic map[string]map[int]bool) error {

	for propName, propType := range readSetConfigurer.Props() {
		prop, err := readSetConfigurer.Prop(propName)

		if err != nil {
			return errors.Extend(err, propName+" had an error")
		}

		policy := transformation[propName]

		if policy == PolicyInvalid {
			return errors.New("Effective policy computed to PolicyInvalid")
		}

		readSetConfigurer.ConfigureProp(propName, applyPolicy(policy, prop, propType))

		if visibleDynamic != nil {

			if policy != PolicyVisible {
				continue
			}

			var stacks []Stack

			if propType == TypeStack {
				stacks = []Stack{prop.(Stack)}
			} else if propType == TypeBoard {
				stacks = prop.(Board).Spaces()
			}

			for _, stack := range stacks {

				if _, ok := visibleDynamic[stack.Deck().Name()]; ok {
					for _, c := range stack.Components() {
						if c == nil {
							continue
						}
						visibleDynamic[c.Deck().Name()][c.DeckIndex()] = true
					}
				}
			}
		}

	}

	return nil

}

//transitivelyMarkDynamicComponentsAsVisible expands which
//dynamiccomponentvalues are visible by extending the visibility throughout
//any items that are in stacks on dynamiccomponentvalues that are visible.
func transativelyMarkDynamicComponentsAsVisible(dynamicComponentValues map[string][]ConfigurableSubState, visibleComponents map[string]map[int]bool) {

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

		for _, stack := range stacksForReader(reader) {
			if _, ok := dynamicComponentValues[stack.Deck().Name()]; !ok {
				//This stack is for a deck that has no dynamic values, can skip.
				continue
			}

			//Ok, if we get to here then we have a stack with items in a deck that does have dynamic values.
			for _, c := range stack.Components() {
				if c == nil {
					continue
				}
				//There can't possibly be a collision because each component may only be in a single stack at a time.
				visibleComponents[c.Deck().Name()][c.DeckIndex()] = true
				//Take note that there's another item to add to the queue to explore.
				workItems = append(workItems, workItem{c.Deck().Name(), c.DeckIndex()})
			}
		}

	}
}

//sanitizeDynamicComponentValues is more complex than just applying a
//straightforward sanitizationTransformation because the components should
//only folow the configured property if the component they're affiliated with
//was PolicyVisible.
func sanitizeDynamicComponentValues(dynamicComponentValues map[string][]ConfigurableSubState, visibleComponents map[string]map[int]bool, transformation map[string]subStateSanitizationTransformation) error {

	for name, slice := range dynamicComponentValues {

		visibleDynamicDeck := visibleComponents[name]

		for i, value := range slice {

			readSetConfigurer := value.ReadSetConfigurer()

			if _, visible := visibleDynamicDeck[i]; visible {

				if err := sanitizeStateObj(readSetConfigurer, transformation[name], nil); err != nil {
					return errors.Extend(err, "Couldn't sanitize random dynamic component")
				}

			} else {
				//Make it a hidden

				for propName, propType := range readSetConfigurer.Props() {
					prop, err := readSetConfigurer.Prop(propName)

					if err != nil {
						continue
					}

					readSetConfigurer.ConfigureProp(propName, applyPolicy(PolicyHidden, prop, propType))

				}
			}
		}
	}
	return nil

}

func applyPolicy(policy Policy, input interface{}, propType PropertyType) interface{} {
	if policy == PolicyVisible {
		return input
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
	case TypeEnum:
		e := input.(enum.Val).ImmutableCopy()
		res, _ := e.Enum().NewImmutableVal(e.Enum().DefaultValue())
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

	if propType == TypeBoard {
		board := input.(MutableBoard)
		board.applySanitizationPolicy(policy)
		return input
	}

	//Now we're left with len-properties.

	stack := input.(MutableStack)

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

func (b *board) applySanitizationPolicy(policy Policy) {
	for _, stack := range b.spaces {
		stack.applySanitizationPolicy(policy)
	}
}

func (g *growableStack) applySanitizationPolicy(policy Policy) {

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
			g.overrideIds[i] = c.ID()
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
		id := c.ID()
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

	seedStr := stack.state().game.secretSalt + strconv.Itoa(stack.state().Version())

	h := fnv.New64()
	h.Write([]byte(seedStr))
	seed := h.Sum64()

	r := rand.New(rand.NewSource(int64(seed)))

	return r.Perm(stack.Len())

}

func (s *sizedStack) applySanitizationPolicy(policy Policy) {

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
			s.overrideIds[i] = c.ID()
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
		id := c.ID()
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
