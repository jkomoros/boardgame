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
export { BoardgameTableViewBase } from './components/boardgame-table-view-base.js';
export { BoardgameHandViewBase } from './components/boardgame-hand-view-base.js';
export type { SeatPresentation } from './components/boardgame-table-view-base.js';
export { defineEffectTheme, fx } from './effects/effect-spec.js';
export type {
  BurstEffectSpec,
  EffectAnchor,
  EffectHandle,
  EffectHostAPI,
  EffectIntensity,
  EffectResult,
  EffectSpec,
  EffectTheme,
  EffectTone,
  EffectTransitionContext,
  NamedEffectAnchor,
  ParallelEffectSpec,
  PointEffectAnchor,
  PulseEffectSpec,
  SequenceEffectSpec,
  TravelEffectSpec,
} from './effects/effect-spec.js';
export { glyphForSlug } from './components/companion-avatar-catalog.js';
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
export { targetList } from './moves/target-list.js';
export type { TargetListBinding, TargetListChoice } from './moves/target-list.js';
export { SourceDestinationController } from './moves/source-destination.js';
export type {
  SourceDestinationBinding,
  SourceDestinationHost,
  SourceDestinationOptions,
} from './moves/source-destination.js';
export { PlacementDraftController } from './moves/placement-draft.js';
export type {
  DraftPlacement,
  PlacementItemBinding,
  PlacementDraftBinding,
  PlacementDraftNotice,
  PlacementDraftOptions,
  PlacementDraftRebasePolicy,
  PlacementTargetBinding,
} from './moves/placement-draft.js';
export type {
  BoardPiece,
  BoardPathOverlay,
  BoardPathTone,
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
export type { TargetListLayout } from './components/boardgame-target-list.js';
export type { ComponentZoneLayout } from './components/boardgame-component-zone.js';
export type { GameOutcomeViewer } from './components/boardgame-game-outcome.js';
export type { BoardViewportChange } from './components/boardgame-board-viewport.js';
export type { GridPlacementDraft } from './components/boardgame-game-board.js';
export type { SpatialPlacementDraft } from './components/boardgame-spatial-board.js';
export type { DraftControlsBinding } from './components/boardgame-draft-controls.js';
export type {
  InspectorChangeReason,
  InspectorOpenChangedDetail,
} from './components/boardgame-inspector.js';
export type { ReadinessView } from './components/boardgame-readiness.js';
export type { ClientMove, JsonValue } from './types/api.js';
export type {
  ReadinessKey,
  ReadinessLabels,
  ReadinessParticipant,
  ReadinessPresentation,
  ReadinessState,
} from './status/readiness.js';
export { readinessPresentation } from './status/readiness.js';
export {
  SelectionDraftController,
  type SelectionDraftBinding,
  type SelectionDraftNotice,
  type SelectionOptionBinding,
  type SelectionDraftOptions,
  type SelectionDraftRebasePolicy,
} from './moves/selection-draft.js';
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
export type { PlayerPresentation } from './status/player-presentation.js';
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

// Importing the facade registers every curated renderer primitive. Game code
// must not depend on transitive or deep component-module side effects. Their
// classes are intentionally not exported: the supported creator contract is
// custom-element markup and the facade types, not implementation inheritance.
import './components/boardgame-die.js';
import './components/boardgame-card.js';
import './components/boardgame-token.js';
import './components/boardgame-component-stack.js';
import './components/boardgame-game-board.js';
import './components/boardgame-player-badge.js';
import './components/boardgame-fading-text.js';
import './components/boardgame-action-button.js';
import './components/boardgame-action-bar.js';
import './components/boardgame-draft-controls.js';
import './components/boardgame-inspector.js';
import './components/boardgame-placement-item.js';
import './components/boardgame-readiness.js';
import './components/boardgame-selection-option.js';
import './components/boardgame-target-list.js';
import './components/boardgame-component-zone.js';
import './components/boardgame-game-outcome.js';
import './components/boardgame-game-surface.js';
import './components/boardgame-player-grid.js';
import './components/boardgame-player-panel.js';
import './components/boardgame-status-text.js';
import './components/boardgame-timer.js';
import './components/boardgame-turn-status.js';
import './components/boardgame-spatial-board.js';
import './components/boardgame-board-viewport.js';
