import type { ProjectedMoveChoicesWire } from '../types/api.js';
import { playerPresentations } from '../status/player-presentation.js';
import { expandGameStateSnapshot } from '../selectors.js';
import type { GameFromServer } from '../types/game-state.js';
import type { GameChest } from '../types/store.js';
import type {
  RendererFixtureGameContract,
  RendererFixtureLegality,
  RendererFixtureSnapshot,
  RendererFixtureSurface,
} from './renderer-fixture.js';

/** Portable output of boardgame-util/lib/scenario.Run. */
export interface ScenarioReplay {
  readonly schemaVersion: 1;
  readonly name: string;
  readonly gameName: string;
  readonly chest: GameChest;
  readonly moveInputSchemaFingerprint: string;
  readonly frames: readonly {
    readonly label: string;
    readonly version: number;
    readonly viewers: readonly {
      readonly viewer: number;
      readonly game: GameFromServer;
      readonly zeroInputMoves?: readonly string[];
      readonly projectedMoveChoices?: ProjectedMoveChoicesWire;
      readonly moveLegality: Readonly<Record<string, RendererFixtureLegality>>;
    }[];
  }[];
}

/**
 * Adapt one real engine decision boundary to the existing fixture API. This
 * displays recorded states; fixture clicks still record proposals rather than
 * inventing a next state. Advance explicitly to review each scripted step.
 */
export function scenarioFixtureSnapshot<Contract extends RendererFixtureGameContract>(
  replay: ScenarioReplay,
  frameIndex: number,
  viewer: number,
  options: { readonly gameName: string; readonly surface?: RendererFixtureSurface },
): RendererFixtureSnapshot<Contract> {
  if (replay.schemaVersion !== 1 || replay.gameName !== options.gameName) {
    throw new Error('Scenario schema/game does not match this renderer');
  }
  if (!Number.isInteger(frameIndex) || frameIndex < 0 || frameIndex >= replay.frames.length) {
    throw new Error('Scenario frame is out of range');
  }
  const frame = replay.frames[frameIndex];
  if (!frame) throw new Error('Scenario frame is missing');
  const snapshot = frame.viewers.find(candidate => candidate.viewer === viewer);
  if (!snapshot || snapshot.game.Version !== frame.version || snapshot.game.Name !== replay.gameName) {
    throw new Error('Scenario has no matching viewer/version snapshot');
  }
  const game = snapshot.game;
  return {
    schemaVersion: 1,
    // The caller selects its generated contract and we verify the game name.
    // Expansion is shared with live rendering, never a parallel simulator.
    state: expandGameStateSnapshot(game.CurrentState, replay.chest, replay.gameName, game.ActiveTimers) as Contract['State'],
    ...(snapshot.projectedMoveChoices ? { projectedMoveChoices: snapshot.projectedMoveChoices } : {}),
    requireRecordedPreviews: true,
    recordedZeroInputMoves: snapshot.zeroInputMoves ?? [],
    timers: game.ActiveTimers ?? {},
    playerPresentations: playerPresentations(game.CurrentState.Players.map(() => ({})), []),
    viewingAsPlayer: viewer,
    currentPlayerIndex: game.CurrentPlayerIndex,
    moveLegality: snapshot.moveLegality as Readonly<Record<Contract['MoveName'], RendererFixtureLegality>>,
    version: frame.version,
    outcome: { finished: game.Finished, winners: game.Winners ?? [] },
    surface: options.surface ?? 'game',
    serverMoveInputSchemaFingerprint: replay.moveInputSchemaFingerprint,
  };
}
