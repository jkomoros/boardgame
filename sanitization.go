package boardgame

import (
	"math"
)

//StatePolicy defines a sanitization policy for a State object. In particular,
//it defines a policy for the Game state, and a single, fixed policy for all
//Player states. Each string returns the policy for the property with that
//name in that sub-state object. Properties with no corresponding policy are
//effectively PolicyNoOp for all groups.
type StatePolicy struct {
	Game   map[string]GroupPolicy
	Player map[string]GroupPolicy
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
	//Non sanitized
	PolicyVisible Policy = iota
	//For groups (e.g. stacks, int slices), return a group that has the same
	//length. For all else, it's effectively PolicyHidden. In practice, stacks
	//will be set so that their NumComponents() is the same, but every
	//component that exists returns the GenericComponent.
	PolicyLen

	//TODO: implement the other policies.
)

//statePlayerIndex is the index of the PlayerState that we're working on (-1
//for Game). preparingForPlayerIndex is the index that we're preparing the
//overall santiized state for, as provied to
//GameManager.SanitizedStateForPlayer()
func sanitizeStateObj(readSetter PropertyReadSetter, policy map[string]GroupPolicy, statePlayerIndex int, preparingForPlayerIndex int) {

	for propName, propType := range readSetter.Props() {
		prop, err := readSetter.Prop(propName)

		if err != nil {
			//TODO: shouldn't we return an error or something?
			continue
		}
		readSetter.SetProp(propName, sanitizeProperty(prop, propType, policy[propName], statePlayerIndex, preparingForPlayerIndex))
	}

}

func sanitizeProperty(prop interface{}, propType PropertyType, policyGroup GroupPolicy, statePlayerIndex int, preparingForPlayerIndex int) interface{} {

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
	}

	return applyPolicy(effectivePolicy, prop, propType)
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
	}

	//Now we're left with len-properties.

	stack := input.(Stack)

	stack.applySanitizationPolicy(policy)

	return input

}
