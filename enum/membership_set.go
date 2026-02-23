package enum

import "slices"

// ImmutableMembershipSet represents a read-only set of enum values from a
// particular enum. It is used to represent which groups a player belongs to,
// for example in sanitization logic.
type ImmutableMembershipSet interface {
	//Enum returns the Enum this membership set is associated with.
	Enum() Enum
	//Contains returns true if the given EnumKey is in the set.
	Contains(val EnumKey) bool
	//ContainsVal returns true if the given ImmutableVal's value is in the set.
	//If val is nil, returns false.
	ContainsVal(val ImmutableVal) bool
	//Members returns a sorted slice of all EnumKeys in the set.
	Members() []EnumKey
	//Len returns the number of members in the set.
	Len() int
	//ImmutableCopy returns an immutable copy of this membership set.
	ImmutableCopy() ImmutableMembershipSet
	//Copy returns a mutable copy of this membership set.
	Copy() MembershipSet
}

// MembershipSet represents a mutable set of enum values from a particular enum.
type MembershipSet interface {
	ImmutableMembershipSet
	//Add adds the given EnumKey to the set.
	Add(val EnumKey)
	//Remove removes the given EnumKey from the set.
	Remove(val EnumKey)
}

// membershipSet is the unexported implementation of ImmutableMembershipSet and
// MembershipSet.
type membershipSet struct {
	enum Enum
	data map[EnumKey]bool
}

// NewMembershipSet creates a new MembershipSet for this enum containing the
// given members. Any members not valid for this enum are silently ignored.
func (e *enum) NewMembershipSet(members ...EnumKey) MembershipSet {
	data := make(map[EnumKey]bool, len(members))
	for _, m := range members {
		if e.Valid(m) {
			data[m] = true
		}
	}
	return &membershipSet{
		enum: e,
		data: data,
	}
}

func (m *membershipSet) Enum() Enum {
	return m.enum
}

func (m *membershipSet) Contains(val EnumKey) bool {
	return m.data[val]
}

func (m *membershipSet) ContainsVal(val ImmutableVal) bool {
	if val == nil {
		return false
	}
	return m.data[val.Value()]
}

func (m *membershipSet) Members() []EnumKey {
	result := make([]EnumKey, 0, len(m.data))
	for k := range m.data {
		result = append(result, k)
	}
	slices.Sort(result)
	return result
}

func (m *membershipSet) Len() int {
	return len(m.data)
}

func (m *membershipSet) Add(val EnumKey) {
	if m.enum != nil && !m.enum.Valid(val) {
		return
	}
	m.data[val] = true
}

func (m *membershipSet) Remove(val EnumKey) {
	delete(m.data, val)
}

func (m *membershipSet) Copy() MembershipSet {
	data := make(map[EnumKey]bool, len(m.data))
	for k, v := range m.data {
		data[k] = v
	}
	return &membershipSet{enum: m.enum, data: data}
}

func (m *membershipSet) ImmutableCopy() ImmutableMembershipSet {
	return m.Copy()
}
