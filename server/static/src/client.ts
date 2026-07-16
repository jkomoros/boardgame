/**
 * Experimental game-renderer authoring facade.
 *
 * Game clients should prefer this deliberately small entry point over imports
 * from framework internals. Exports may be added only after a real renderer
 * proves the need; deep implementation modules remain unsupported.
 */
export { html, css } from 'lit';
export { BoardgameBaseGameRenderer } from './components/boardgame-base-game-renderer.js';
export { BoardgameBasePlayerInfoRenderer } from './components/boardgame-base-player-info-renderer.js';
export type {
  PlayerChipPresentation,
  PlayerChipPresentationChangedDetail,
} from './components/boardgame-base-player-info-renderer.js';
export { isVisibleComponent } from './types/boardgame-types.js';
export type {
  Component,
  CatalogComponent,
  DeepReadonly,
  ExpandedBoard,
  ExpandedStack,
  FullGameState,
  GameChest,
  OpaqueComponent,
  VisibleComponent,
} from './types/boardgame-types.js';
export {
  assertMoveInputSchemaFingerprint,
  serializeCreatorMoveInput,
  serializeCreatorMoveInputForServer,
  validateCreatorMoveInput,
  MoveInputValidationError,
  StaleMoveInputSchemaError,
} from './moves/input.js';
export type {
  BoundMoveAction,
  MoveActionAvailability,
  MoveActionBuilder,
  MoveActionFor,
  MoveActionPrecondition,
  MoveActionPreview,
  MoveActionReason,
  MoveActionReasonCode,
  MoveActionSubmission,
  MoveActionTelemetryEvent,
  MoveProposalResult,
} from './moves/action.js';
export { bindMoveAction } from './moves/action-binding.js';
export type { MoveActionBindingOptions } from './moves/action-binding.js';
export type {
  TargetAction,
  TargetActionOptions,
  TargetActionPreview,
  TargetCandidate,
  TargetKey,
} from './moves/target-action.js';
export { SourceDestinationController } from './moves/source-destination.js';
export type {
  SourceDestinationBinding,
  SourceDestinationHost,
  SourceDestinationOptions,
} from './moves/source-destination.js';
export type {
  BoardPiece,
  BoardGeometry,
  BoardGeometryFactory,
  BoardGeometrySpace,
  NormalizedBoardPoint,
  NormalizedBoardRegion,
  NormalizedBoardSpace,
  RasterBoardArtwork,
  RasterBoardArtworkOptions,
  RasterArtworkFit,
  SpatialBoardKey,
} from './components/spatial-board-geometry.js';
export { piecesFromSizedStacks, rasterBoardArtwork } from './components/spatial-board-geometry.js';
export { cardView, componentView, tokenView } from './components/component-view.js';
export type { ComponentView, ComponentViewContext, ComponentViewOptions } from './components/component-view.js';
export type { FadingTextAutoMessage, FadingTextSuppress, FadingTextTrigger } from './components/boardgame-fading-text.js';
export type { StatusTextAutoMessage, StatusTextValue } from './components/boardgame-status-text.js';
export type { ActionBarAlignment, ActionBarOrientation } from './components/boardgame-action-bar.js';
export type { ComponentZoneLayout } from './components/boardgame-component-zone.js';
export type { GameOutcomeViewer } from './components/boardgame-game-outcome.js';
export {
  AdminPlayerIndex,
  AnyPlayerIndex,
  ObserverPlayerIndex,
  isConcretePlayerIndex,
  isKnownPlayerIndex,
  turnStatusPresentation,
} from './status/turn-status.js';
export type {
  SpecialPlayerIndex,
  TurnStatusContext,
  TurnStatusKind,
  TurnStatusPresentation,
} from './status/turn-status.js';
export { TimerController } from './timers/timer-service.js';
export type {
  TimerCadence,
  TimerControllerOptions,
  TimerReading,
  TimerReference,
  TimerStatus,
} from './timers/timer-service.js';
export type { TimerDisplayFormat } from './components/boardgame-timer.js';
export { isStackLayout } from './components/boardgame-component-stack.js';
export type { StackLayout } from './components/boardgame-component-stack.js';
export type {
  MoveInputErrorCode,
  MoveInputIssue,
  MoveInputSchema,
  MoveInputSchemaField,
  MoveInputSchemaMove,
  MoveInputValidationResult,
} from './moves/input.js';

// Importing the facade registers the zero-configuration primitives used by
// the Pig proving renderer. Their classes are intentionally not exported yet:
// the supported contract is the custom-element markup, not implementation
// methods inherited from today's legacy elements.
import './components/boardgame-die.js';
import './components/boardgame-fading-text.js';
import './components/boardgame-action-button.js';
import './components/boardgame-action-bar.js';
import './components/boardgame-component-zone.js';
import './components/boardgame-game-outcome.js';
import './components/boardgame-game-surface.js';
import './components/boardgame-player-grid.js';
import './components/boardgame-player-panel.js';
import './components/boardgame-status-text.js';
import './components/boardgame-timer.js';
import './components/boardgame-turn-status.js';
import './components/boardgame-spatial-board.js';
