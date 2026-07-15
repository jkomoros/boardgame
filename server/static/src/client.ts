/**
 * Experimental game-renderer authoring facade.
 *
 * Game clients should prefer this deliberately small entry point over imports
 * from framework internals. Exports may be added only after a real renderer
 * proves the need; deep implementation modules remain unsupported.
 */
export { html, css } from 'lit';
export { BoardgameBaseGameRenderer } from './components/boardgame-base-game-renderer.js';
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
