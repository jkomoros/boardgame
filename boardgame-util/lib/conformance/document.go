package conformance

import (
	"encoding/json"
	"time"

	"github.com/jkomoros/boardgame"
)

// SchemaVersion is bumped whenever the Document format changes shape.
const SchemaVersion = 1

// Clock is the fixed manual timer clock every export runs under. It is the
// same instant scenario.Run uses, so a scenario and its export agree.
var Clock = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// Document is one exported run. Field order here is the order in the file.
type Document struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Source        Source `json:"source"`
	// Normalization lists every field that is omitted or rewritten.
	Normalization []Rule `json:"normalization"`
	Game          Game   `json:"game"`
	Engine        Engine `json:"engine"`
	// Viewers is the order of Views in every Version: admin (-2), observer
	// (-1), then each player index.
	Viewers []boardgame.PlayerIndex `json:"viewers"`
	// Moves is the static move table, in the engine's registration order.
	Moves []MoveType `json:"moves"`
	// MoveInputSchema and MoveChoiceProjectionSchema are the engine's own
	// schema structures (boardgame.BuildMoveInputSchema and
	// boardgame.BuildMoveChoiceProjectionSchema).
	MoveInputSchema            json.RawMessage `json:"moveInputSchema"`
	MoveChoiceProjectionSchema json.RawMessage `json:"moveChoiceProjectionSchema"`
	// ComputedProperties is GameManager.ComputedPropertyDescriptors.
	ComputedProperties json.RawMessage `json:"computedProperties"`
	// VariantConfig is the game type's full variant configuration.
	VariantConfig json.RawMessage `json:"variantConfig"`
	// Chest is the wire chest: Decks, Enums, Constants and LegalTemplates.
	Chest    json.RawMessage `json:"chest"`
	Versions []Version       `json:"versions"`
}

// Source says where the run came from.
type Source struct {
	// Kind is "scenario" or "golden".
	Kind string `json:"kind"`
	// Path is the golden file, relative to the repository root. Empty for a
	// scenario.
	Path string `json:"path,omitempty"`
	// Description is the golden's Description or the scenario's name.
	Description string `json:"description,omitempty"`
	// Covers is a human note on what the run exercises.
	Covers string `json:"covers,omitempty"`
}

// Rule is one normalization: a field and what the exporter does with it.
type Rule struct {
	// Field is a path in the Document, or the wire/storage name of a field
	// that is absent from it.
	Field string `json:"field"`
	// Action is "omitted", "fixed", "derived", "rewritten" or "kept".
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// Game is everything needed to recreate the game before version 0.
type Game struct {
	Name              string `json:"name"`
	DisplayName       string `json:"displayName"`
	ID                string `json:"id"`
	SecretSalt        string `json:"secretSalt"`
	NumPlayers        int    `json:"numPlayers"`
	MinNumPlayers     int    `json:"minNumPlayers"`
	MaxNumPlayers     int    `json:"maxNumPlayers"`
	DefaultNumPlayers int    `json:"defaultNumPlayers"`
	// Variant is the game's effective variant: the requested values with the
	// configuration's defaults filled in.
	Variant boardgame.Variant `json:"variant"`
	Agents  []string          `json:"agents"`
	// Seating is true when the run seats players through the seating
	// rendezvous (it contains a move with trigger "seat"). The engine tells
	// the game so before version 0, and games behave differently for it:
	// with seating off, seats are never waited for or inactivated as empty.
	Seating bool `json:"seating"`
	// Finished and Winners are the game's final values.
	Finished bool                    `json:"finished"`
	Winners  []boardgame.PlayerIndex `json:"winners"`
	// FinalVersion is the last entry of Versions.
	FinalVersion int `json:"finalVersion"`
}

// Engine pins the engine-level facts a consumer may need to match.
type Engine struct {
	// Clock is the fixed timer clock, RFC 3339. A timer started at any
	// version has Deadline = Clock + its duration.
	Clock                                 string `json:"clock"`
	LegalCatalogVersion                   int    `json:"legalCatalogVersion"`
	MoveInputSchemaFingerprint            string `json:"moveInputSchemaFingerprint"`
	MoveChoiceProjectionSchemaFingerprint string `json:"moveChoiceProjectionSchemaFingerprint"`
	MoveChoiceProjectionSchemaVersion     int    `json:"moveChoiceProjectionSchemaVersion"`
	// PhaseEnum and GroupEnum name the chest enums used for phases and
	// sanitization groups. Empty when the game has none.
	PhaseEnum string `json:"phaseEnum"`
	GroupEnum string `json:"groupEnum"`
}

// MoveType is the static description of one configured move.
type MoveType struct {
	Name     string `json:"name"`
	HelpText string `json:"helpText"`
	// IsFixUp moves are never offered to players; the engine proposes them.
	IsFixUp bool `json:"isFixUp"`
	// NamePublic is false when some viewer may not see the move's name.
	NamePublic bool `json:"namePublic"`
	// AnimationKeys is the canonical name plus any hidden-name alias.
	AnimationKeys    []string    `json:"animationKeys"`
	IsGatheringStart bool        `json:"isGatheringStart"`
	Fields           []MoveField `json:"fields"`
}

// MoveField is one property of a move, sorted by name within a MoveType.
type MoveField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	EnumName string `json:"enumName,omitempty"`
}

// Version is everything about one state version.
type Version struct {
	Version int `json:"version"`
	// Settled is true when the engine was waiting for outside input at this
	// version: setup finished, or a proposal and all its fix-ups applied.
	// The server reports legality and projected choices only for these.
	Settled bool `json:"settled"`
	// Finished and Winners are the game's values as of this version.
	Finished bool                    `json:"finished"`
	Winners  []boardgame.PlayerIndex `json:"winners"`
	// Move produced this version. Null for version 0.
	Move *MoveRecord `json:"move"`
	// State is the unsanitized state as storage holds it.
	State json.RawMessage `json:"state"`
	// Views has one entry per Document.Viewers, in that order.
	Views []View `json:"views"`
}

// MoveRecord is the stored move record, minus its timestamp.
type MoveRecord struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	// Initiator is the version of the move that started this causal chain.
	Initiator int `json:"initiator"`
	// Phase is the phase enum value before the move applied; PhaseName is
	// its string value, empty when the game has no phase enum.
	Phase     int                   `json:"phase"`
	PhaseName string                `json:"phaseName"`
	Proposer  boardgame.PlayerIndex `json:"proposer"`
	// Arguments is the move's serialized properties.
	Arguments json.RawMessage `json:"arguments"`
	// IsFixUp is the move type's IsFixUp.
	IsFixUp bool `json:"isFixUp"`
	// Trigger says what made the engine apply the move: "setup" (the chain
	// the engine starts when the game is created), "propose" (a player
	// proposed it), "timer" (a timer fired), "seat" (a player was seated and
	// a fix-up check forced), or "fixup" (it followed another move in the
	// same chain). Only "propose", "timer" and "seat" are outside input.
	Trigger string `json:"trigger"`
}

// View is what one viewer is sent for one version.
type View struct {
	Viewer boardgame.PlayerIndex `json:"viewer"`
	// Game is Game.JSONForPlayer for this viewer and state, normalized as
	// the Document's normalization section says. CurrentState carries the
	// sanitized Game, Players and Components plus Computed and Visibility.
	Game json.RawMessage `json:"game"`
	// Move is Game.MoveJSONForPlayer: null for version 0 and for a move this
	// viewer may see nothing of; otherwise AnimationKey (the name, or the
	// hidden-name alias), Version and the Properties the viewer may see.
	Move json.RawMessage `json:"move"`
	// Legality has one entry per move offered to this viewer, in move-table
	// order. Fix-up moves and private moves unavailable to the viewer are
	// absent, as on the wire.
	Legality []MoveLegality `json:"legality"`
	// ProjectedMoveChoices is the actor-only choice snapshot. Present only
	// at a settled version, for a seated player, in a game that declares
	// choice projections.
	ProjectedMoveChoices json.RawMessage `json:"projectedMoveChoices,omitempty"`
}

// MoveLegality is one move form's legality for one viewer.
type MoveLegality struct {
	Name string `json:"name"`
	// LegalForPlayer is Move.Legal(state, viewer) == nil for the move bound
	// to its defaults. Always false for the observer.
	LegalForPlayer bool `json:"legalForPlayer"`
	// LegalForPlayerError is that call's error text when it failed.
	LegalForPlayerError string `json:"legalForPlayerError,omitempty"`
	// LegalForAnyone is Move.Legal(state, admin) == nil.
	LegalForAnyone bool `json:"legalForAnyone"`
	// Defaults is the default-bound value of every move field.
	Defaults map[string]json.RawMessage `json:"defaults"`
	// Preconditions is the declarative-legality ledger, for moves that
	// opted in: name, args, verdict ("pass", "fail", "unknown"), message
	// (template key, plus bindings when evaluable), evaluable, provisional.
	Preconditions json.RawMessage `json:"preconditions,omitempty"`
}

// Normalization is the list written into every Document.
func Normalization() []Rule {
	return []Rule{
		{"versions[].move.timestamp (MoveStorageRecord.Timestamp)", "omitted", "Wall-clock time the move was applied."},
		{"game.created, game.modified (GameStorageRecord.Created, Modified)", "omitted", "Wall-clock times."},
		{"versions[].state.Timers[].Deadline", "fixed", "The run uses a manual clock fixed at engine.clock that never advances, so a deadline is engine.clock plus the timer's duration. A golden file's own wall-clock deadlines are not compared."},
		{"versions[].state.Timers[].ID and CurrentState timer ID", "kept", "Timer ids are drawn from State.Rand, so they are a deterministic function of the game id, salt and version."},
		{"versions[].state.Timers[].Move.Timestamp", "kept", "The timer's intended move has never been applied, so this is always Go's zero time."},
		{"versions[].views[].game.Diagram", "omitted", "A debug rendering produced by the Go delegate."},
		{"versions[].views[].game.ActiveTimers", "derived", "On the wire this comes from the manager's process-local scheduler, not from a state version. It is rebuilt from the active timer records of versions[].state as {id: {TimeLeft: milliseconds until Deadline at engine.clock}}, and verified against the scheduler whenever the version is the head."},
		{"versions[].views[].game.Version", "rewritten", "On the wire this is the head version of the game at request time. It is set to the version being viewed."},
		{"versions[].views[].game.Finished, Winners", "rewritten", "On the wire these are the game's current values. They are set to the values as of the version being viewed."},
		{"versions[].views[].legality[].defaults", "rewritten", "The wire's Fields array is in Go map-iteration order and repeats static type information. It is reduced to an object of default values keyed by field name; types are in moves[].fields."},
		{"versions[].views[].legality[] HelpText, IsGatheringStart", "omitted", "Static per move type; see moves[]."},
		{"versions[].views[].legality on versions where settled is false", "derived", "The server reports legality only for the head of a settled game. For other versions it is evaluated against that version's state with moves bound to that state."},
		{"versions[].views[].projectedMoveChoices", "kept", "Present only where the server would send it: a settled version, a seated player viewing as themself, a game with declared projections."},
		{"all embedded JSON", "rewritten", "Re-encoded with object keys sorted and insignificant whitespace removed. Compare structurally."},
		{"game.id, game.secretSalt", "kept", "From the golden file, or derived by SHA-256 from the scenario seed. Never random."},
		{"Players, CompanionInfo, HasEmptySlots and other /info fields", "omitted", "Server session and seating data, not engine output."},
	}
}
