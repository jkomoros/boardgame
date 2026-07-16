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
  SpatialBoardKey,
} from './components/spatial-board-geometry.js';
export { piecesFromSizedStacks } from './components/spatial-board-geometry.js';
export { cardView, componentView, tokenView } from './components/component-view.js';
export type { ComponentView, ComponentViewContext, ComponentViewOptions } from './components/component-view.js';
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
import './components/boardgame-spatial-board.js';
