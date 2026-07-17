package boardgame

import (
	"errors"
	"fmt"

	"github.com/jkomoros/boardgame/enum"
)

const moveNameSanitizationConfigKey = "github.com/jkomoros/boardgame.MoveNameSanitization"

type moveNameSanitizationConfig struct {
	policy             string
	hiddenAnimationKey string
	err                string
}

// SetMoveNameSanitization stores move-name sanitization on a MoveConfig's
// custom configuration. Most games use moves.WithMoveNameSanitization.
func SetMoveNameSanitization(config PropertyCollection, policy string, hiddenAnimationKey ...string) {
	configured := moveNameSanitizationConfig{policy: policy}
	if len(hiddenAnimationKey) > 1 {
		configured.err = "move name sanitization accepts at most one hidden animation key"
	} else if len(hiddenAnimationKey) == 1 {
		configured.hiddenAnimationKey = hiddenAnimationKey[0]
	}
	config[moveNameSanitizationConfigKey] = configured
}

func configuredMoveNameSanitization(config PropertyCollection) (map[string]Policy, string, error) {
	configured, ok := config[moveNameSanitizationConfigKey].(moveNameSanitizationConfig)
	if !ok {
		return map[string]Policy{SanitizationDefaultGroup: PolicyVisible}, "", nil
	}
	if configured.err != "" {
		return nil, "", errors.New(configured.err)
	}
	policies := policyFromStructTagWithDefault(configured.policy, SanitizationDefaultGroup, PolicyHidden)
	for group, policy := range policies {
		if policy != PolicyVisible && policy != PolicyHidden {
			return nil, "", fmt.Errorf("move name sanitization group %q must use visible or hidden", group)
		}
	}
	return policies, configured.hiddenAnimationKey, nil
}

// MovePropertySanitizer is the optional override implemented by base.Move.
// Games normally configure move properties with sanitize struct tags instead.
type MovePropertySanitizer interface {
	SanitizationPolicy(propName string, groupMembership map[string]bool) Policy
}

// MoveNamePublic reports whether the move uses the compatibility default where
// its canonical name is visible to every viewer.
func (m *MoveInfo) MoveNamePublic() bool {
	if m == nil || m.moveType == nil {
		return false
	}
	return ResolveSanitizationPolicy(m.moveType.nameSanitization, map[string]bool{SanitizationDefaultGroup: true}, PolicyHidden) == PolicyVisible
}

// MoveNameVisibleToPlayer evaluates canonical-name visibility for a proposed
// or stored move using the same proposer/viewer group semantics as properties.
func (g *Game) MoveNameVisibleToPlayer(move Move, proposer, viewer PlayerIndex, state ImmutableState) (bool, error) {
	if viewer == AdminPlayerIndex {
		return true, nil
	}
	if move == nil || move.Info() == nil || move.Info().moveType == nil {
		return false, errors.New("move was not initialized")
	}
	groups, err := g.moveSanitizationGroupsFor(state, proposer, viewer, move)
	if err != nil {
		return false, err
	}
	return ResolveSanitizationPolicy(move.Info().moveType.nameSanitization, groups, PolicyHidden) == PolicyVisible, nil
}

// MoveJSONForPlayer returns viewer-specific move metadata suitable for the
// client wire. The persisted storage record is never modified or serialized.
func (g *Game) MoveJSONForPlayer(player PlayerIndex, record *MoveStorageRecord) (interface{}, error) {
	if record == nil {
		return nil, nil
	}
	if g == nil {
		return nil, errors.New("game was nil")
	}
	move, err := record.inflate(g)
	if err != nil {
		return nil, err
	}
	if record.Version < 1 {
		return nil, errors.New("move record version must be positive")
	}
	state := g.State(record.Version - 1)
	if state == nil {
		return nil, fmt.Errorf("couldn't load pre-move state at version %d", record.Version-1)
	}
	groups, err := g.moveSanitizationGroupsFor(state, record.Proposer, player, move)
	if err != nil {
		return nil, err
	}

	moveType := move.Info().moveType
	namePolicy := ResolveSanitizationPolicy(moveType.nameSanitization, groups, PolicyHidden)
	animationKey := record.Name
	if player != AdminPlayerIndex && namePolicy != PolicyVisible {
		animationKey = moveType.hiddenAnimationKey
		if animationKey == "" {
			return nil, nil
		}
	}

	properties := make(PropertyCollection)
	for propName, propType := range move.Reader().Props() {
		policy := PolicyVisible
		if player != AdminPlayerIndex {
			policy = move.Info().SanitizationPolicy(propName, groups)
			if sanitizer, ok := move.(MovePropertySanitizer); ok {
				policy = sanitizer.SanitizationPolicy(propName, groups)
			}
		}
		if !movePolicyCarriesValue(policy, propType) {
			continue
		}
		value, err := move.Reader().Prop(propName)
		if err != nil {
			return nil, fmt.Errorf("move property %q: %w", propName, err)
		}
		properties[propName] = applyPolicy(policy, value, propType)
	}

	result := map[string]interface{}{
		"AnimationKey": animationKey,
		"Version":      record.Version,
	}
	if len(properties) > 0 {
		result["Properties"] = properties
	}
	return result, nil
}

func movePolicyCarriesValue(policy Policy, propType PropertyType) bool {
	if policy == PolicyVisible {
		return true
	}
	if policy == PolicyHidden || policy == PolicyInvalid {
		return false
	}
	switch propType {
	case TypeIntSlice, TypeBoolSlice, TypeStringSlice, TypePlayerIndexSlice, TypeEnumSlice:
		return true
	default:
		return false
	}
}

func (g *Game) moveSanitizationGroupsFor(state ImmutableState, proposer, viewer PlayerIndex, move Move) (map[string]bool, error) {
	if state == nil {
		return nil, errors.New("state was nil")
	}
	players := state.ImmutablePlayerStates()

	var proposerState, viewerState ImmutableSubState
	if proposer >= 0 && int(proposer) < len(players) {
		proposerState = players[proposer]
	}
	if viewer >= 0 && int(viewer) < len(players) {
		viewerState = players[viewer]
	}
	proposerMembership, groups := groupMembershipForPlayerState(proposerState)
	viewerMembership, _ := groupMembershipForPlayerState(viewerState)

	moveType := move.Info().moveType
	groupEnum := g.Manager().Delegate().GroupEnum()
	groupNames := moveType.validator.sanitizationPolicyGroupNames(groupEnum)
	for groupName := range moveType.nameSanitization {
		if groupName == SanitizationDefaultGroup {
			continue
		}
		if groupEnum != nil && groupEnum.ValueFromString(groupName) != enum.IllegalValue {
			continue
		}
		groupNames[groupName] = true
	}
	for groupName := range groupNames {
		if proposerState == nil && groupName != sanitizationGroupSelf && groupName != sanitizationGroupOther {
			groups[groupName] = false
			continue
		}
		inGroup, err := g.Manager().computedPlayerGroupMembership(groupName, proposer, viewer, proposerMembership, viewerMembership)
		if err != nil {
			return nil, fmt.Errorf("move sanitization group %q: %w", groupName, err)
		}
		groups[groupName] = inGroup
	}
	return groups, nil
}
