/*

	auto is a package designed to make it trivial to create MoveTypeConfigs
	for your moves.

	Use Case

	Creating MoveConfig's is a necessary part of installing moves on your
	GameManager, but it's verbose and error-prone. You need to create a lot of
	extra structs, and then remember to provide the right properties in your
	config. And to use many of the powerful moves in the moves package, you
	need to write a lot of boilerplate methods to integrate correctly.
	Finally, you end up repeating yourself often--which makes it a pain if you
	change the name of a move.

	Take this example:

		//+autoreader
		type MoveDealInitialCards struct {
			moves.DealComponentsUntilPlayerCountReached
		}

		var moveDealInitialCardsConfig = boardgame.MoveConfig {
			Name: "Deal Initial Cards",
			HelpText: "Deal initial cards to players",
			Constructor: func() boardgame.Move {
				return new(MoveDealInitialCards)
			},
		}

		func (m *MoveDealInitialCards) GameStack(gState boardgame.MutableSubState) boardgame.MutableStack {
			return gState.(*gameState).DrawStack
		}

		func (m *MoveDealInitialCards) PlayerStack(pState boardgame.MutablePlayerState) boardgame.MutableStack {
			return pState.(*playerState).Hand
		}

		func (m *MoveDealInitialCards) TargetCount() int {
			return 2
		}

		func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
			return moves.Add(
				&moveDealInitialCardsConfig,
			)
		}

	auto.Config (and its panic-y sibling auto.MustConfig) help reduce this
	signficantly:

		func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {
			return moves.Add(
				auto.MustConfig(
					new(moves.DealComponentsUntilPlayerCountReached),
					with.GameStack("DrawStack"),
					with.PlayerStack("Hand"),
					with.TargetCount(2),
				)
			)
		}

	Basic Usage

	auto.Config takes an example struct representing your move, and then a
	list of 0 to n interfaces.CustomConfigurationOption. These options are
	given a boardgame.PropertyCollection and then add specific properties to
	it, and then stash that on the CustomConfiguration property of the
	returned MoveTypeConfig. Different move methods will then reach into that
	configuration to alter the behavior of moves of that type.

	The moves/with package defines a large collection of
	CustomConfigurationOption for use with the moves in the moves package.

	Moves that are used with auto.Config must satisfy the AutoConfigurableMove
	interface, which adds one method: DeriveName() string. auto.Config
	primarily consistsn of some set up and then using those return values as
	fields on the returned MoveConfig. These methods are implemented in
	moves.Base, which means that any move structs that embed moves.Base
	(directly or indirectly) can be used with auto.Config.

	moves.Base does a fair bit of magic in these methods to implement much of
	the logic of auto.Config. In general, if you pass a configuration option
	(via with.MoveName, for example) then that option will be used for that
	method. moves.Base.DeriveName() also will use reflection to automatically
	set a struct name like "MoveDealInitialCards" to "Deal Initial Cards". All
	of the moves in the moves package will also automatically return
	reasonable names for DeriveName(), so in many cases you can use those
	structs directly without having to pass with.MoveName().

	Other moves in the moves package, like DealCountComponents, will use
	configuration, like with.GameStack(), to power their default GameStack()
	method.

	All moves in the moves package are designed to return an error from
	ValidConfiguration(), which means that if you forgot to pass a required
	configuration property (e.g. you don't override GameStack and also don't
	provide with.GameStack), when you try to create NewGameManager() and all
	moves' ValidConfiguration() is checked, you'll get an error. This helps
	catch mis-configurations during boot time.

	Refer to the documentation of the various methods in that package for
	their precise behavior and how to configure them.

	Idiomatic Move Definition and Installation

	auto.Config is at the core of idiomatic definition and installation of
	moves, and typically is used for every move you install in your game. The
	following paragraphs describe the high-level idioms to follow.

	Never create your own MoveConfig objects--it's just another global
	variable that clutters up your code and makes it harder to change.
	Instead, use auto.Config. There are some rare cases where you do want to
	refer to the move by name (and not rely on finicky string-based lookup),
	such as when you want an Agent to propose a speciifc type of move. In
	those cases use auto.Config to create the move type config, then save the
	resulting config's Name to a global variable that you use elsewhere, and
	then pass the created config to moves.Add() (and its cousins)

	In general, you should only create a bespoke Move struct in your game if
	it is not possible to use one of the off-the-shelf moves from the moves
	package, combined with configuarion options, to do what you want. In
	practice this means that only if you need to override a method on one of
	the base moves do you need to create a bespoke struct. This typically
	allows you to drastically reduce the number of bespoke move structs your
	game defines, saving thousands of lines of code (each bespoke struct also
	has hundreds of lines of auto-generated PropertyReader code).

	If you do create a bespoke struct, name it like this: "MoveNameOfMyMove",
	so that moves.Base's default DeriveName() will give it a reasonable name
	automatically (in this example, "Name Of My Move").

	In many cases if you subclass powerful moves like DealCountComponents the
	default HelpText() value is sufficient (especially if it's a FixUp
	move that won't ever be seen by players). In other cases, with.HelpText()
	is often the only config option you will pass to auto.Config.

	If your move will be a FixUp move that doesn't sublcass one of the more
	advanced fix up moves (like RoundRobin or DealCountComponents), embed
	moves.FixUp into your struct. That will cause IsFixUp to return the right
	value even without using with.IsFixUp--because with.IsFixUp is easy to
	forget given that it's often in a different file. In almost all cases if
	you use with.IsFixUp you should simply embed moves.FixUp instead.

	auto.MustConfig is like auto.Config, but instead of returning a MoveConfig
	and an error, it simply returns a MoveConfig--and panics if it would have
	returned an error. Since your GameDelegate's ConfigureMoves() is typically
	called during the boot-up sequence of your game, it is safe to use
	auto.MustConfig exclusively, which saves many lines of boilerplate error
	checking.

*/
package auto
