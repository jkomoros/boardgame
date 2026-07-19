// Package choice provides sealed immutable descriptors for finite move-choice
// projection. The descriptors authorize a security-sensitive read model; they
// contain no prompt, title, localized copy, layout, or other UI presentation.
package choice

import (
	"sort"

	"github.com/jkomoros/boardgame"
)

// Projection is an immutable declaration for one creator-owned move field.
// Construct one with PlayerIndexes or EnumValues, then explicitly authorize
// actor-exact disclosure with DiscloseExactAvailabilityToActor.
type Projection struct {
	fieldName      string
	source         boardgame.MoveChoiceSource
	excludedValues []string
	disclosure     boardgame.MoveChoiceDisclosure
	auditRationale string
}

// PlayerIndexes enumerates valid player indexes from the projected state.
func PlayerIndexes(fieldName string) Projection {
	return Projection{fieldName: fieldName, source: boardgame.MoveChoiceSourcePlayers}
}

// EnumValues enumerates the configured enum's canonical string values.
func EnumValues(fieldName string) Projection {
	return Projection{fieldName: fieldName, source: boardgame.MoveChoiceSourceEnumValues}
}

// Excluding removes static implementation sentinels from an enum universe. It
// does not make any remaining candidate legal.
func (p Projection) Excluding(values ...string) Projection {
	p.excludedValues = append([]string(nil), p.excludedValues...)
	p.excludedValues = append(p.excludedValues, values...)
	sort.Strings(p.excludedValues)
	return p
}

// DiscloseExactAvailabilityToActor explicitly authorizes revealing candidate
// membership and exact availability to the player who would propose the move.
// rationale is retained at runtime for security review but never fingerprinted
// or sent to clients.
func (p Projection) DiscloseExactAvailabilityToActor(rationale string) Projection {
	p.disclosure = boardgame.MoveChoiceDisclosureActorExact
	p.auditRationale = rationale
	return p
}

// Declaration returns a defensive copy of the core declaration. It exists for
// moves.WithChoiceProjection; game authors should normally not call it.
func (p Projection) Declaration() boardgame.MoveChoiceProjection {
	return boardgame.MoveChoiceProjection{
		FieldName:      p.fieldName,
		Source:         p.source,
		ExcludedValues: append([]string(nil), p.excludedValues...),
		Disclosure:     p.disclosure,
		AuditRationale: p.auditRationale,
	}
}
