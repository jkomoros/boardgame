package gamedef

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/enum"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

const (
	movesPkgPath     = "github.com/jkomoros/boardgame/moves"
	behaviorsPkgPath = "github.com/jkomoros/boardgame/behaviors"
	// moveNameSanitizationKey is the core's moveNameSanitizationConfigKey
	// (move_sanitization.go), unexported there.
	moveNameSanitizationKey = "github.com/jkomoros/boardgame.MoveNameSanitization"
)

// Export is the run-time half of the exporter: it reads the live manager
// (moves in registration order, the chest, the constructors, plans,
// progressions, variants) and joins static, the static half from Static,
// into one GameData.
//
// manager is a GameManager built over the game's delegate, as
// testdata/gen/defs and lib/conformance build one:
//
//	manager, err := boardgame.NewGameManager(memory.NewDelegate(), memstore.NewStorageManager())
//
// static should come from Static on the same game's package. When static is
// nil, Export still succeeds and leaves empty what only the static half
// knows: MoveData.Bodies, Traits' FnRefs, Functions, Computed[].Compute,
// Predicates[].Construct, SetUp, End's FnRefs, CurrentPlayerIndex and a
// custom counter's Count. When static is non-nil, Export fails when a
// configured move's Go type has no entry in static.Moves.
//
// Some parts the proposal assigns to the static half are also derived here
// by reflection: MoveData.Stock.Type (the first embedded type declared in
// package moves, or the move's own type when it is declared there) and the
// behavior traits (which behaviors the state structs embed, and
// DrawDiscardPair's struct tags). When static is non-nil its value is used,
// and a disagreement with the run-time value is an error; see joinTraits for
// what counts as a disagreement for traits.
func Export(manager *boardgame.GameManager, static *StaticInfo) (*GameData, error) {
	if manager == nil {
		return nil, errors.New("gamedef: nil manager")
	}
	delegate := manager.Delegate()
	data := &GameData{
		SchemaVersion: SchemaVersion,
		Source:        Source{Exporter: Exporter},
		Name:          delegate.Name(),
		DisplayName:   delegate.DisplayName(),
		Description:   delegate.Description(),
		// Through the interface, as the manager and the conformance header
		// call them: a misspelled override (DefaultNumPlayeres) is not used.
		MinNumPlayers:     delegate.MinNumPlayers(),
		MaxNumPlayers:     delegate.MaxNumPlayers(),
		DefaultNumPlayers: delegate.DefaultNumPlayers(),
		ComponentEnums:    componentEnums(manager),
		LegalTemplates:    map[string]string{},
		Computed:          []ComputedData{},
		Predicates:        []PredicateData{},
		Moves:             []MoveData{},
		Agents:            []AgentData{},
		Functions:         []FunctionEntry{},
	}
	if static != nil {
		data.Source.Package = static.Package
	} else {
		data.Source.Package = reflect.TypeOf(delegate).Elem().PkgPath()
	}

	def, err := DefinitionFor(manager)
	if err != nil {
		return nil, fmt.Errorf("definition: %w", err)
	}
	data.Definition = *def

	chest := manager.Chest()
	if names := chest.ConstantNames(); len(names) > 0 {
		data.Constants = map[string]any{}
		for _, name := range names {
			data.Constants[name] = chest.Constant(name)
		}
	}
	if phases := delegate.PhaseEnum(); phases != nil {
		data.PhaseEnum = phases.Name()
	}
	if groups := delegate.GroupEnum(); groups != nil {
		data.GroupEnum = groups.Name()
	}

	if data.Initial, err = initial(manager); err != nil {
		return nil, fmt.Errorf("initial: %w", err)
	}

	runtimeTraits, err := traitsFor(manager)
	if err != nil {
		return nil, fmt.Errorf("traits: %w", err)
	}
	if data.Traits, err = joinTraits(runtimeTraits, static); err != nil {
		return nil, err
	}

	if data.Variants, err = canonicalJSON(manager.Variants()); err != nil {
		return nil, fmt.Errorf("variants: %w", err)
	}

	if data.Computed, err = computed(manager, static); err != nil {
		return nil, fmt.Errorf("computed: %w", err)
	}

	if configurer, ok := delegate.(interface{ ConfigureLegalTemplates() map[string]string }); ok {
		for key, text := range configurer.ConfigureLegalTemplates() {
			data.LegalTemplates[key] = text
		}
	}

	if configurer, ok := delegate.(interface {
		ConfigurePredicateConstructors() []*boardgame.LegalPredicateConstructor
	}); ok {
		for _, ctor := range configurer.ConfigurePredicateConstructors() {
			if ctor == nil || ctor.Name == "" {
				continue
			}
			entry := PredicateData{Name: ctor.Name}
			if static != nil {
				entry.Construct = static.Predicates[ctor.Name]
			}
			data.Predicates = append(data.Predicates, entry)
		}
	}

	choices, err := boardgame.BuildMoveChoiceProjectionSchema(manager)
	if err != nil {
		return nil, fmt.Errorf("choice projection schema: %w", err)
	}
	choicesByMove := map[string]boardgame.MoveChoiceProjectionSchema{}
	for _, choice := range choices {
		choicesByMove[choice.MoveName] = choice
	}
	for _, move := range manager.ExampleMoves() {
		entry, err := moveData(manager, move, choicesByMove, static)
		if err != nil {
			return nil, fmt.Errorf("move %q: %w", move.Info().Name(), err)
		}
		data.Moves = append(data.Moves, *entry)
	}

	if data.ComponentLegal, err = componentLegal(manager, static); err != nil {
		return nil, fmt.Errorf("component legal: %w", err)
	}

	if static != nil {
		data.SetUp = static.SetUp
		data.End = EndData{ConditionMet: static.End.ConditionMet, PlayerScore: static.End.PlayerScore, Check: static.End.Check}
		data.CurrentPlayerIndex = static.CurrentPlayerIndex
		if static.Functions != nil {
			data.Functions = append([]FunctionEntry{}, static.Functions...)
			sort.SliceStable(data.Functions, func(i, j int) bool { return data.Functions[i].Func < data.Functions[j].Func })
		}
	}
	if scorer, ok := delegate.(interface{ LowScoreWins() bool }); ok {
		data.End.LowScoreWins = scorer.LowScoreWins()
	}

	for _, agent := range manager.Agents() {
		data.Agents = append(data.Agents, AgentData{Name: agent.Name(), DisplayName: agent.DisplayName()})
	}
	return data, nil
}

// Marshal is the checked-in form of a GameData: indented with tabs, HTML
// characters not escaped, one trailing newline. Export followed by Marshal
// is deterministic.
func Marshal(data *GameData) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// ----- initial values -------------------------------------------------------

// initial calls the three constructors and keeps each non-zero property.
func initial(manager *boardgame.GameManager) (Initial, error) {
	delegate := manager.Delegate()
	result := Initial{Players: []map[string]any{}, Dynamic: map[string][]map[string]any{}}
	var err error
	if result.Game, err = nonZeroProps(delegate.GameStateConstructor()); err != nil {
		return result, fmt.Errorf("game: %w", err)
	}
	for i := 0; i < delegate.MaxNumPlayers(); i++ {
		props, err := nonZeroProps(delegate.PlayerStateConstructor(boardgame.PlayerIndex(i)))
		if err != nil {
			return result, fmt.Errorf("player %d: %w", i, err)
		}
		result.Players = append(result.Players, props)
	}
	chest := manager.Chest()
	for _, deckName := range chest.DeckNames() {
		deck := chest.Deck(deckName)
		var entries []map[string]any
		for i := 0; i < deck.Len(); i++ {
			values := delegate.DynamicComponentValuesConstructor(deck)
			if isNil(values) {
				entries = nil
				break
			}
			props, err := nonZeroProps(values)
			if err != nil {
				return result, fmt.Errorf("deck %s component %d: %w", deckName, i, err)
			}
			entries = append(entries, props)
		}
		if entries != nil {
			result.Dynamic[deckName] = entries
		}
	}
	return result, nil
}

// nonZeroProps reads every scalar and slice property of a freshly
// constructed sub-state through its generated reader, and keeps those that
// are not the type's zero value: a non-zero number, true, a non-empty
// string, an enum value the constructor set (by its string, as the state
// JSON stores it), a non-nil slice. Stacks, boards and timers are left to the
// inflater and never kept.
func nonZeroProps(sub boardgame.Reader) (map[string]any, error) {
	result := map[string]any{}
	if isNil(sub) {
		return result, nil
	}
	reader := sub.Reader()
	for _, name := range sortedKeys(reader.Props()) {
		var value any
		keep := false
		switch reader.Props()[name] {
		case boardgame.TypeInt:
			v, err := reader.IntProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v != 0
		case boardgame.TypeBool:
			v, err := reader.BoolProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v
		case boardgame.TypeString:
			v, err := reader.StringProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v != ""
		case boardgame.TypePlayerIndex:
			v, err := reader.PlayerIndexProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = int(v), v != 0
		case boardgame.TypeEnum:
			v, err := reader.ImmutableEnumProp(name)
			if err == nil && !isNil(v) {
				value, keep = v.String(), true
			}
		case boardgame.TypeIntSlice:
			v, err := reader.IntSliceProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v != nil
		case boardgame.TypeBoolSlice:
			v, err := reader.BoolSliceProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v != nil
		case boardgame.TypeStringSlice:
			v, err := reader.StringSliceProp(name)
			if err != nil {
				return nil, err
			}
			value, keep = v, v != nil
		case boardgame.TypePlayerIndexSlice:
			v, err := reader.PlayerIndexSliceProp(name)
			if err != nil {
				return nil, err
			}
			if v != nil {
				ints := make([]int, len(v))
				for i, p := range v {
					ints[i] = int(p)
				}
				value, keep = ints, true
			}
		case boardgame.TypeEnumSlice:
			v, err := reader.ImmutableEnumSliceProp(name)
			if err == nil && !isNil(v) {
				strs := []string{}
				for _, key := range v.Values() {
					strs = append(strs, v.Enum().String(key))
				}
				value, keep = strs, true
			}
		}
		if keep {
			result[name] = value
		}
	}
	return result, nil
}

// isNil reports whether v is nil or a typed nil pointer, map, slice or
// interface.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func, reflect.Chan:
		return rv.IsNil()
	}
	return false
}

// ----- traits ---------------------------------------------------------------

// traitsFor reads the behaviors the game and player state structs embed.
// The FnRef-valued traits (PlayerOrder, a declared GameScore, Turn) are the
// static half's; GameScore is GameScoreFromScore here when the player state
// embeds behaviors.ScoreBehavior.
func traitsFor(manager *boardgame.GameManager) (Traits, error) {
	var result Traits
	example := manager.ExampleState()
	if example == nil {
		return result, errors.New("no example state")
	}
	gameType := structType(reflect.TypeOf(example.ImmutableGameState()))
	if gameType != nil {
		for i := 0; i < gameType.NumField(); i++ {
			field := gameType.Field(i)
			if !field.Anonymous || field.Type.PkgPath() != behaviorsPkgPath {
				continue
			}
			switch field.Type.Name() {
			case "DrawDiscardPair":
				result.Game.DrawDiscardPair = &DrawDiscardPair{Draw: field.Tag.Get("draw"), Discard: field.Tag.Get("discard")}
			case "RoundRobin":
				result.Game.RoundRobin = true
			}
		}
	}
	players := example.ImmutablePlayerStates()
	if len(players) == 0 {
		return result, nil
	}
	playerType := structType(reflect.TypeOf(players[0]))
	if playerType != nil {
		for i := 0; i < playerType.NumField(); i++ {
			field := playerType.Field(i)
			if !field.Anonymous || field.Type.PkgPath() != behaviorsPkgPath {
				continue
			}
			switch field.Type.Name() {
			case "Seat":
				result.Player.Seat = true
			case "InactivePlayer":
				result.Player.Inactive = true
			case "PlayerElimination":
				result.Player.Elimination = true
			case "PlayerRole":
				result.Player.Role = true
			case "PlayerColor":
				result.Player.Color = true
			case "ScoreBehavior":
				result.Player.GameScore = GameScoreFromScore
			}
		}
	}
	return result, nil
}

// joinTraits merges the run-time traits with the static half's. A boolean
// trait is true when either half says so, except that the static half saying
// false while the run-time half says true is not a disagreement only when the
// static half set no boolean at all (it may carry FnRefs alone); otherwise a
// differing boolean, or a differing DrawDiscardPair, is an error. FnRefs come
// from the static half; a declared GameScore method replaces
// GameScoreFromScore, which reflection cannot tell from a promoted one.
func joinTraits(runtime Traits, static *StaticInfo) (Traits, error) {
	if static == nil {
		return runtime, nil
	}
	s := static.Traits
	result := s
	staticSetsBools := s.Game.RoundRobin || s.Player.Seat || s.Player.Inactive || s.Player.Elimination || s.Player.Role || s.Player.Color || s.Game.DrawDiscardPair != nil
	check := func(name string, runtimeValue, staticValue bool) error {
		if staticSetsBools && runtimeValue != staticValue {
			return fmt.Errorf("traits: %s is %v at run time but %v in the static half", name, runtimeValue, staticValue)
		}
		return nil
	}
	for _, c := range []struct {
		name   string
		r, s   bool
		target *bool
	}{
		{"Game.RoundRobin", runtime.Game.RoundRobin, s.Game.RoundRobin, &result.Game.RoundRobin},
		{"Player.Seat", runtime.Player.Seat, s.Player.Seat, &result.Player.Seat},
		{"Player.Inactive", runtime.Player.Inactive, s.Player.Inactive, &result.Player.Inactive},
		{"Player.Elimination", runtime.Player.Elimination, s.Player.Elimination, &result.Player.Elimination},
		{"Player.Role", runtime.Player.Role, s.Player.Role, &result.Player.Role},
		{"Player.Color", runtime.Player.Color, s.Player.Color, &result.Player.Color},
	} {
		if err := check(c.name, c.r, c.s); err != nil {
			return result, err
		}
		*c.target = c.r || c.s
	}
	switch {
	case s.Game.DrawDiscardPair == nil:
		result.Game.DrawDiscardPair = runtime.Game.DrawDiscardPair
	case runtime.Game.DrawDiscardPair != nil && *runtime.Game.DrawDiscardPair != *s.Game.DrawDiscardPair:
		return result, fmt.Errorf("traits: DrawDiscardPair is %+v at run time but %+v in the static half", *runtime.Game.DrawDiscardPair, *s.Game.DrawDiscardPair)
	}
	if result.Player.GameScore == "" {
		result.Player.GameScore = runtime.Player.GameScore
	}
	return result, nil
}

// ----- component legality ---------------------------------------------------

// componentLegal joins the static half's Legal methods, keyed by component
// values type, with the decks: a deck gets the FnRef of its components'
// values type when that type implements interfaces.LegalComponent (what
// moves.DefaultComponent calls). Nil when no deck has one, or static is nil.
//
// A deck whose values implement LegalComponent but whose type the static
// half has no Legal for (a method promoted from another package), or a deck
// mixing values types with different Legal methods, is an error: the loader
// could only guess.
func componentLegal(manager *boardgame.GameManager, static *StaticInfo) (map[string]FnRef, error) {
	if static == nil {
		return nil, nil
	}
	var result map[string]FnRef
	chest := manager.Chest()
	for _, deckName := range chest.DeckNames() {
		var ref FnRef
		for _, c := range chest.Deck(deckName).Components() {
			if c == nil || isNil(c.Values()) {
				continue
			}
			if _, ok := c.Values().(interfaces.LegalComponent); !ok {
				continue
			}
			valuesType := structType(reflect.TypeOf(c.Values()))
			if valuesType == nil {
				return nil, fmt.Errorf("deck %s: component values are not a struct", deckName)
			}
			found, ok := static.ComponentLegal[valuesType.String()]
			if !ok {
				return nil, fmt.Errorf("deck %s: %s implements LegalComponent, but the static half has no Legal declared on it", deckName, valuesType)
			}
			if ref != "" && ref != found {
				return nil, fmt.Errorf("deck %s: components have two Legal methods, %s and %s", deckName, ref, found)
			}
			ref = found
		}
		if ref == "" {
			continue
		}
		if result == nil {
			result = map[string]FnRef{}
		}
		result[deckName] = ref
	}
	return result, nil
}

// ----- computed properties --------------------------------------------------

// computed lists the computed properties in the order the delegate declares
// them, each with the descriptor the manager installed.
func computed(manager *boardgame.GameManager, static *StaticInfo) ([]ComputedData, error) {
	descriptors := map[string]boardgame.ComputedPropertyDescriptor{}
	for _, d := range manager.ComputedPropertyDescriptors() {
		descriptors[string(d.Scope)+"/"+d.Name] = d
	}
	// ComputedProperty keeps its descriptor unexported; its Name and Scope
	// are read by reflection to recover declaration order.
	var order []string
	for _, property := range manager.Delegate().ConfigureComputedProperties() {
		descriptor := reflect.ValueOf(property).FieldByName("descriptor")
		if !descriptor.IsValid() {
			order = nil
			break
		}
		order = append(order, descriptor.FieldByName("Scope").String()+"/"+descriptor.FieldByName("Name").String())
	}
	if len(order) != len(descriptors) {
		// Fall back to the manager's own (sorted) order.
		order = order[:0]
		for _, d := range manager.ComputedPropertyDescriptors() {
			order = append(order, string(d.Scope)+"/"+d.Name)
		}
	}
	result := []ComputedData{}
	for _, key := range order {
		d, ok := descriptors[key]
		if !ok {
			return nil, fmt.Errorf("computed property %s is declared but not installed", key)
		}
		typeName, ok := typeNames[d.Type]
		if !ok {
			return nil, fmt.Errorf("computed property %s has unsupported type %v", d.Name, d.Type)
		}
		entry := ComputedData{Name: d.Name, Scope: string(d.Scope), Type: typeName}
		if static != nil {
			entry.Compute = static.Computed[d.Name]
		}
		result = append(result, entry)
	}
	return result, nil
}

// ----- moves ----------------------------------------------------------------

func moveData(manager *boardgame.GameManager, move boardgame.Move, choices map[string]boardgame.MoveChoiceProjectionSchema, static *StaticInfo) (*MoveData, error) {
	info := move.Info()
	name := info.Name()
	moveType := structType(reflect.TypeOf(move))
	if moveType == nil {
		return nil, errors.New("the move is not a pointer to a struct")
	}
	entry := &MoveData{
		Name:     name,
		HelpText: move.HelpText(),
		IsFixUp:  base.IsFixUp(move),
		GoType:   moveType.String(),
		Base:     BaseDefault,
		Fields:   map[string]FieldData{},
	}
	embedded := embeddedTypes(moveType)
	switch {
	case embedded[movesType("CurrentPlayer")] || moveType == movesType("CurrentPlayer"):
		entry.Base = BaseCurrentPlayer
	case embedded[movesType("AnyPlayer")] || moveType == movesType("AnyPlayer"):
		entry.Base = BaseAnyPlayer
	}

	config := info.CustomConfiguration()
	stockType := nearestStock(moveType)
	if static != nil {
		s, ok := static.Moves[entry.GoType]
		if !ok {
			return nil, fmt.Errorf("the static half has no entry for Go type %s", entry.GoType)
		}
		if s.StockType != stockType {
			return nil, fmt.Errorf("stock ancestor is %q at run time but %q in the static half", stockType, s.StockType)
		}
		entry.Bodies = s.Bodies
	}
	if stockType != "" {
		serialized, err := serializeConfig(config)
		if err != nil {
			return nil, err
		}
		if err := addResolvedConfig(move, stockType, serialized); err != nil {
			return nil, err
		}
		entry.Stock = &StockData{Type: stockType, Config: serialized}
	}

	// Fields: what the base does not supply.
	resolved, err := boardgame.ResolveMoveInputFields(move)
	if err != nil {
		return nil, err
	}
	inputs := map[string]boardgame.MoveInputField{}
	for _, field := range resolved {
		inputs[field.Name] = field
	}
	props := move.ReadSetter().Props()
	for _, fieldName := range sortedKeys(props) {
		if fieldName == "TargetPlayerIndex" && entry.Base != BaseDefault {
			continue
		}
		propType := props[fieldName]
		typeName, ok := typeNames[propType]
		if !ok {
			return nil, fmt.Errorf("field %s has unsupported type %v", fieldName, propType)
		}
		field := FieldData{Type: typeName, Input: string(inputs[fieldName].Disposition), Codec: string(inputs[fieldName].Codec)}
		if propType == boardgame.TypeEnum || propType == boardgame.TypeEnumSlice {
			if value, err := move.ReadSetter().ImmutableEnumProp(fieldName); err == nil && value != nil && value.Enum() != nil {
				field.Enum = value.Enum().Name()
			}
		}
		if structField, ok := moveType.FieldByName(fieldName); ok {
			field.Sanitize = structField.Tag.Get("sanitize")
		}
		entry.Fields[fieldName] = field
	}

	if sanitization, ok := config[moveNameSanitizationKey]; ok {
		policy := reflect.ValueOf(sanitization).FieldByName("policy")
		if !policy.IsValid() || policy.Kind() != reflect.String {
			return nil, errors.New("unrecognized move name sanitization configuration")
		}
		entry.NameSanitization = policy.String()
	}
	if keys := info.AnimationKeys(); len(keys) > 1 {
		entry.HiddenAnimationKey = keys[1]
	}

	if phaser, ok := move.(interface{ InstalledLegalPhases() []enum.EnumKey }); ok {
		if phases := phaser.InstalledLegalPhases(); phases != nil {
			entry.LegalPhases = make([]int, len(phases))
			for i, phase := range phases {
				entry.LegalPhases[i] = int(phase)
			}
		}
	}

	if group := moves.InstalledProgression(config); group != nil {
		progression, err := progressionData(moves.DescribeProgression(group), name, static)
		if err != nil {
			return nil, fmt.Errorf("progression: %w", err)
		}
		entry.Progression = &progression
	}
	if multi, ok := move.(interfaces.AllowMultipleInProgression); ok {
		entry.AllowMultipleInProgression = multi.AllowMultipleInProgression()
	}

	specs, declared, hasCustom, ok := manager.LegalPlanSpecs(name)
	if !ok {
		return nil, errors.New("LegalPlanSpecs does not know the move")
	}
	if specs == nil {
		specs = []boardgame.LegalSpec{}
	}
	entry.Plan = PlanData{Declared: declared, Specs: specs, HasCustom: hasCustom}

	if choice, ok := choices[name]; ok {
		if entry.Choices, err = canonicalJSON(choice); err != nil {
			return nil, err
		}
	}
	if seater, ok := move.(interfaces.SeatPlayerMover); ok && seater.IsSeatPlayerMove() {
		entry.SeatsPlayer = true
	}
	if starter, ok := move.(interfaces.GatheringStartMover); ok && starter.IsGatheringStartMove() {
		entry.GatheringStart = true
	}
	// moves/installers.go: an ordered phase must end with a phaseToStarter
	// (PhaseToStart) or a NoOp.
	if _, ok := move.(interface {
		PhaseToStart(enum.EnumKey) (enum.EnumKey, error)
	}); ok || moveType == movesType("NoOp") || embedded[movesType("NoOp")] {
		entry.EndsPhaseProgression = true
	}
	return entry, nil
}

// ResolvedSlotFieldKey is the Stock.Config key under which Export writes the
// move property a MoveComponentToSlot reads its slot from. Go stores nothing
// for it: when WithSlotField is absent the move discovers its only int
// property every time it needs it (moves.ResolvedSlotField). The exporter
// writes what Go resolved, as DrawToPlayer's ValidConfiguration itself
// records ResolvedDrawSource and ResolvedDrawDestination, so the engine's
// registry never re-implements the inference.
const ResolvedSlotFieldKey = movesPkgPath + ".ResolvedSlotField"

// addResolvedConfig adds to a serialized stock configuration what Go infers
// at boot but does not store. Every other stock type either stores its
// resolution in the configuration (DrawToPlayer) or infers nothing the
// engine needs from the configuration.
func addResolvedConfig(move boardgame.Move, stockType string, config map[string]any) error {
	if stockType != "MoveComponentToSlot" {
		return nil
	}
	field, err := moves.ResolvedSlotField(move)
	if err != nil {
		return fmt.Errorf("slot field: %w", err)
	}
	config[ResolvedSlotFieldKey] = field
	return nil
}

var movesTypes = map[string]reflect.Type{
	"CurrentPlayer": reflect.TypeOf(moves.CurrentPlayer{}),
	"AnyPlayer":     reflect.TypeOf(moves.AnyPlayer{}),
	"NoOp":          reflect.TypeOf(moves.NoOp{}),
	"Default":       reflect.TypeOf(moves.Default{}),
}

func movesType(name string) reflect.Type { return movesTypes[name] }

// structType dereferences pointers to a struct type, or returns nil.
func structType(t reflect.Type) reflect.Type {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}
	return t
}

// embeddedTypes is every struct type t embeds, at any depth.
func embeddedTypes(t reflect.Type) map[reflect.Type]bool {
	result := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.Anonymous {
				continue
			}
			inner := structType(field.Type)
			if inner == nil || result[inner] {
				continue
			}
			result[inner] = true
			walk(inner)
		}
	}
	walk(t)
	return result
}

// nearestStock is the move's nearest stock ancestor's name in package moves:
// the move's own type when package moves declares it, else the first
// embedded type (depth first, in field order) that package moves declares.
// "" when that is moves.Default or there is none.
func nearestStock(t reflect.Type) string {
	var find func(reflect.Type) reflect.Type
	find = func(t reflect.Type) reflect.Type {
		if t.PkgPath() == movesPkgPath {
			return t
		}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.Anonymous {
				continue
			}
			inner := structType(field.Type)
			if inner == nil {
				continue
			}
			if found := find(inner); found != nil {
				return found
			}
		}
		return nil
	}
	found := find(t)
	if found == nil || found == movesType("Default") {
		return ""
	}
	return found.Name()
}

// progressionData converts package moves' description to the schema's. A
// custom counter's Count is the static half's FnRef for the position
// "<move name>" followed by "/<child index>" for each step from the root.
func progressionData(desc moves.ProgressionDescription, position string, static *StaticInfo) (ProgressionData, error) {
	result := ProgressionData{Kind: desc.Kind, Name: desc.Name}
	switch desc.Kind {
	case moves.ProgressionKindMove:
		return result, nil
	case moves.ProgressionKindSerial, moves.ProgressionKindParallel:
		result.Children = []ProgressionData{}
		for i, child := range desc.Children {
			converted, err := progressionData(child, position+"/"+strconv.Itoa(i), static)
			if err != nil {
				return result, err
			}
			result.Children = append(result.Children, converted)
		}
	case moves.ProgressionKindRepeat:
		if desc.Child == nil {
			return result, errors.New("repeat group without a child")
		}
		child, err := progressionData(*desc.Child, position+"/0", static)
		if err != nil {
			return result, err
		}
		result.Child = &child
	default:
		return result, fmt.Errorf("group at %s cannot be described (%s): %s", position, desc.Kind, desc.Detail)
	}
	if desc.Kind != moves.ProgressionKindSerial {
		if desc.Count == nil {
			return result, fmt.Errorf("%s group at %s has no counter", desc.Kind, position)
		}
		counter := CounterData{Kind: desc.Count.Kind}
		if desc.Count.N != nil {
			counter.N = *desc.Count.N
		}
		if desc.Count.Min != nil {
			counter.Min = *desc.Count.Min
		}
		if desc.Count.Max != nil {
			counter.Max = *desc.Count.Max
		}
		if counter.Kind == CounterCustom && static != nil {
			counter.Count = static.Counters[position]
		}
		result.Count = &counter
	}
	return result, nil
}

// ----- move configuration ---------------------------------------------------

// serializeConfig is a move config's CustomConfiguration(), keys as stored,
// values serialized by kind (see serializeValue).
func serializeConfig(config boardgame.PropertyCollection) (map[string]any, error) {
	result := map[string]any{}
	for key, value := range config {
		serialized, err := serializeValue(reflect.ValueOf(value))
		if err != nil {
			return nil, fmt.Errorf("config %s: %w", key, err)
		}
		result[key] = serialized
	}
	return result, nil
}

var (
	progressionGroupType = reflect.TypeOf((*moves.MoveProgressionGroup)(nil)).Elem()
	enumType             = reflect.TypeOf((*enum.Enum)(nil)).Elem()
	immutableValType     = reflect.TypeOf((*enum.ImmutableVal)(nil)).Elem()
)

// serializeValue turns one configuration value into JSON-ready data:
//
//   - a progression group: its ProgressionData;
//   - an enum: its name; an enum value: its string;
//   - a string, bool or integer kind (enum keys included): itself;
//   - a value of a type without unexported fields: what encoding/json makes
//     of it (so a LegalSpec keeps its lowercase keys);
//   - otherwise a struct is an object of its fields by Go name (unexported
//     ones included, read by reflection), a slice or array a list, a
//     pointer or interface what it points to (null when nil), and a map
//     with string keys an object.
//
// Functions, channels and other kinds are an error.
func serializeValue(v reflect.Value) (any, error) {
	if !v.IsValid() {
		return nil, nil
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil, nil
		}
		return serializeValue(v.Elem())
	}
	if v.CanInterface() {
		if v.Type().Implements(progressionGroupType) && !(v.Kind() == reflect.Pointer && v.IsNil()) {
			progression, err := progressionData(moves.DescribeProgression(v.Interface().(moves.MoveProgressionGroup)), "config", nil)
			if err != nil {
				return nil, err
			}
			return progression, nil
		}
		if v.Type().Implements(enumType) && !(v.Kind() == reflect.Pointer && v.IsNil()) {
			return v.Interface().(enum.Enum).Name(), nil
		}
		if v.Type().Implements(immutableValType) && !(v.Kind() == reflect.Pointer && v.IsNil()) {
			return v.Interface().(enum.ImmutableVal).String(), nil
		}
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return nil, nil
		}
		return serializeValue(v.Elem())
	}
	if v.CanInterface() && plainType(v.Type(), map[reflect.Type]bool{}) {
		return canonicalJSON(v.Interface())
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil, nil
		}
		list := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			item, err := serializeValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			list[i] = item
		}
		return list, nil
	case reflect.Struct:
		object := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			item, err := serializeValue(v.Field(i))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", v.Type().Field(i).Name, err)
			}
			object[v.Type().Field(i).Name] = item
		}
		return object, nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map with %s keys", v.Type().Key())
		}
		if v.IsNil() {
			return nil, nil
		}
		object := map[string]any{}
		for _, key := range v.MapKeys() {
			item, err := serializeValue(v.MapIndex(key))
			if err != nil {
				return nil, err
			}
			object[key.String()] = item
		}
		return object, nil
	}
	return nil, fmt.Errorf("cannot serialize a value of kind %s (%s)", v.Kind(), v.Type())
}

// plainType reports whether encoding/json serializes t completely: no
// unexported struct field, function, channel, interface or progression
// group anywhere in it.
func plainType(t reflect.Type, seen map[reflect.Type]bool) bool {
	if seen[t] {
		return true
	}
	seen[t] = true
	if t.Implements(progressionGroupType) || t.Implements(enumType) || t.Implements(immutableValType) {
		return false
	}
	switch t.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return plainType(t.Elem(), seen)
	case reflect.Map:
		return t.Key().Kind() == reflect.String && plainType(t.Elem(), seen)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() || strings.HasPrefix(field.Tag.Get("json"), "-") || !plainType(field.Type, seen) {
				return false
			}
		}
		return true
	}
	return false
}

// canonicalJSON is v's JSON with sorted object keys and no extra whitespace,
// as lib/conformance writes header values.
func canonicalJSON(v any) (json.RawMessage, error) {
	blob, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}
