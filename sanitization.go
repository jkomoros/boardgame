package boardgame

import (
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/errors"
	"hash/fnv"
	"math/rand"
	"sort"
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

/*

A sanitization policy reflects how to tranform a given State property when
presenting to someone outside of the target group. They are returned from your
GameDelegate's SanitizationPolicy method, and the results are used to
configure how properties are modified or elided in state.SanitizedForPlayer.

For most types of properties, there are two effective policies: PolicyVisible,
in which the property is left untouched, and PolicyHidden, in which case the
value is sanitized to its zero value.

However Stacks are much more complicated and have more policies. Even when the
specific components in a stack aren't visible, it can still be important to
know that a given ComponentInstance in one state is the same as another
ComponentInstance in another state, which allows for example animations of the
same logical card from one stack to another, even though the value of the card
is not visible.

In order to do this, every component has a semi-stable Id. This Id is
calculated based on a hash of the component, deck, deckIndex, gameId, and also
a secret salt for the game. This way, the same component in different games
will have different Ids, and if you have never observed the value of the
component in a given game, it is impossible to guess it. However, it is
possible to keep track of the component as it moves between different stacks
within a game.

Every stack has an ordered list of Ids representing the Id for each component.
Components can also be queried for their Id.

Stacks also have an unordered set of IdsLastSeen, which tracks the last time
the Id was affirmitively seen in a stack. The basic time this happens is when
a component is first inserted into a stack. (See below for additional times
when this hapepns)

Different Sanitization Policies will do different things to Ids and
IdsLastSeen, according to the following table:

	| Policy         | Values Behavior                                                  | Ids()                       | IdsLastSeen() | ShuffleCount() |Notes                                                                                                  |
	|----------------|------------------------------------------------------------------|-----------------------------|---------------|----------------|-------------------------------------------------------------------------------------------------------|
	| PolicyVisible  | All values visible                                               | Present                     | Present       | Present        | Visible is effectively no transformation                                                              |
	| PolicyOrder    | All values replaced by generic component                         | Present                     | Present       | Present        | PolicyOrder is similar to PolicyLen, but the order of components is observable                        |
	| PolicyLen      | All values replaced by generic component                         | Sorted Lexicographically    | Present       | Present        | PolicyLen makes it so it's only possible to see the length of a stack, not its order.                 |
	| PolicyNonEmpty | Values will be either 0 components or a single generic component | Absent                      | Present       | Absent         | PolicyNonEmpty makes it so it's only possible to tell if a stack had 0 items in it or more than zero. |
	| PolicyHidden   | Values are completely empty                                      | Absent                      | Absent        | Absent         | PolicyHidden is the most restrictive; stacks look entirely empty.                                     |


However, in some cases it is not possible to keep track of the precise order
of components, even with perfect observation. The canonical example is when a
stack is shuffled. Another example would be when a card is inserted at an
unknown location in a deck.

For this reason, a component's Id is only semi-stable. When one of these
secret moves has occurred, the Ids is randomized. However, in order to be able
to keep track of where the component is, the component is "seen" in
IdsLastSeen immediately before having its Id scrambled, and immediately after.
This procedure is referred to as "scrambling" the Ids.

stack.Shuffle() automatically scrambles the ids of all items in the stack.
SecretMoveComponent, which is similar to the normal MoveComponent, moves the
component to the target stack and then scrambles the Ids of ALL components in
that stack as described above. This is because if only the new item's id
changed, it would be trivial to observe that the new Id is equivalent to the
old Id.

*/
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

func (s *state) SanitizedForPlayer(player PlayerIndex) ImmutableState {

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

func generateSubStateSanitizationTransformation(subState ImmutableSubState, propertyRef StatePropertyRef, delegate GameDelegate, generatingForPlayer PlayerIndex, index PlayerIndex) subStateSanitizationTransformation {

	//Since propertyRef is passed in by value we can modify it locally without a problem

	result := make(subStateSanitizationTransformation)

	for propName := range subState.Reader().Props() {
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

	for deckName := range s.dynamicComponentValues {
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

			var stacks []ImmutableStack

			if propType == TypeStack {
				stacks = []ImmutableStack{prop.(ImmutableStack)}
			} else if propType == TypeBoard {
				stacks = prop.(Board).ImmutableSpaces()
			}

			for _, stack := range stacks {

				if _, ok := visibleDynamic[stack.Deck().Name()]; ok {
					for _, c := range stack.ImmutableComponents() {
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
		for index := range visibleItems {
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
			for _, c := range stack.ImmutableComponents() {
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
		board := input.(Board)
		board.applySanitizationPolicy(policy)
		return input
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
			g.overrideIds = overrideIDsForLen(g)
		}

		indexes := make([]int, len(g.indexes))

		for i := 0; i < len(indexes); i++ {
			indexes[i] = genericComponentSentinel
		}

		g.indexes = indexes
		return
	}

	g.shuffleCount = 0

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

func overrideIDsForLen(stack Stack) []string {
	ids := stack.Ids()
	sort.Strings(ids)
	return ids
}

//returns a random permutation of size stack.Len(). The permutation will be
//predictable given this exact stack and its state, but unpredictable in
//general. This makes it give predictable results for testing but still be
//unguessable if you don't have the stack's game's SecretSalt. This method
//exists even though state has a soruce of randomness because this library
//should only use state.Rand() as a source if it's deterministically called,
//whereas any given state might have multiple sanitizated states created
//implicitly.
func randPermForStack(stack Stack) []int {

	//We want this to be deterministic for two reasons: to have stable goldens
	//to compare against in testing. But also so that the same stack,
	//sanitized with order, will have a somewhat-stable list of IDs. If it
	//changes every time then they might animate on the client, which is what
	//caused #711.

	//We want something unguessable, different per stack, that's semi-stable,
	//but that doesn't change when items in the stack are reordered. Ideally
	//we'd use a stable stack.Id(), but those don't exist. Another option
	//would be to do it based on the secret salt of the game, as well as the
	//StatePropertyRef of this stack in state, which is stable. But we don't
	//actually know that cheaply.

	//As a compromise, iterate through the stack to find the lowest-ID'd
	//component, and use that. That will change if that particular component
	//happens to leave the stack, but that fact is already observable via
	//lastSeenIds, so that's OK>

	lowestComponentId := ""

	for _, c := range stack.Components() {
		if c == nil {
			continue
		}
		if lowestComponentId == "" {
			lowestComponentId = c.ID()
			continue
		}
		if c.ID() < lowestComponentId {
			lowestComponentId = c.ID()
		}
	}

	seedStr := stack.state().game.secretSalt + lowestComponentId

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
			s.overrideIds = overrideIDsForLen(s)
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

	s.shuffleCount = 0

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
