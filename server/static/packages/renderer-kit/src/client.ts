/**
 * Experimental game-renderer authoring facade.
 *
 * Game clients should prefer this deliberately small entry point over imports
 * from framework internals. Exports may be added only after a real renderer
 * proves the need; deep implementation modules remain unsupported.
 */
export { html, css } from 'lit';
// The shared layout/state vocabulary. Every generated renderer base already
// carries it, so most renderers never name these; import them when a renderer
// replaces `static styles` outright instead of spreading the base's.
export { layoutStyles, rendererStyles, stateStyles } from './styles/renderer-styles.js';
export { BoardgameBaseGameRenderer } from './components/boardgame-base-game-renderer.js';
export { BoardgameBasePlayerInfoRenderer } from './components/boardgame-base-player-info-renderer.js';
export { BoardgameTableViewBase } from './components/boardgame-table-view-base.js';
export { BoardgameHandViewBase } from './components/boardgame-hand-view-base.js';
export type { SeatPresentation } from './components/boardgame-table-view-base.js';
export { defineEffectTheme, fx } from './effects/effect-spec.js';
export { diffVisibleComponents } from './effects/visible-component-diff.js';
export type { VisibleComponentDiff } from './effects/visible-component-diff.js';
export { motion } from '@boardgame/elements/motion/cohort.js';
export type { MotionStaggerCohortSpec } from '@boardgame/elements/motion/cohort.js';
export type { MotionTransferDeclaration } from '@boardgame/elements/motion/transfer.js';
export type { MotionReleaseDeclaration } from '@boardgame/elements/motion/release.js';
export { motionSilhouette } from '@boardgame/elements/motion/subject.js';
export type {
  MotionSilhouetteShape,
  MotionSilhouetteSnapshot,
  MotionSubjectSnapshot,
} from '@boardgame/elements/motion/subject.js';
export type {
  BurstEffectSpec,
  AfterMotionEffectSpec,
  DecorateMotionEffectSpec,
  EffectAnchor,
  EffectHandle,
  EffectHostAPI,
  EffectIntensity,
  EffectPointAnchor,
  EffectResult,
  EffectSpec,
  EffectTheme,
  EffectTone,
  EffectTransitionContext,
  NamedEffectAnchor,
  MotionEffectAnchor,
  MotionEffectMoment,
  ParallelEffectSpec,
  PointEffectAnchor,
  SubjectEffectAnchor,
  PulseEffectSpec,
  SequenceEffectSpec,
  TrailEffectSpec,
  TravelEffectSpec,
} from './effects/effect-spec.js';
export { glyphForSlug } from '@boardgame/elements/components/companion-avatar-catalog.js';
export type {
  PlayerChipPresentation,
  PlayerChipPresentationChangedDetail,
} from './components/boardgame-base-player-info-renderer.js';
export { isVisibleComponent } from '@boardgame/elements/types/boardgame-types.js';
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
} from '@boardgame/elements/types/boardgame-types.js';
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
} from '@boardgame/elements/moves/action.js';
export { bindMoveAction } from './moves/action-binding.js';
export type { MoveActionBindingOptions } from './moves/action-binding.js';
export type {
  TargetAction,
  TargetActionOptions,
  TargetActionPreview,
  TargetCandidate,
  TargetKey,
} from '@boardgame/elements/moves/target-action.js';
export { targetList } from '@boardgame/elements/moves/target-list.js';
export type { TargetListBinding, TargetListChoice } from '@boardgame/elements/moves/target-list.js';
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
} from '@boardgame/elements/components/spatial-board-geometry.js';
export { piecesFromSizedStacks, rasterBoardArtwork } from '@boardgame/elements/components/spatial-board-geometry.js';
export { cardView, componentView, dieView, tokenView } from '@boardgame/elements/components/component-view.js';
export type { ComponentView, ComponentViewContext, ComponentViewOptions } from '@boardgame/elements/components/component-view.js';
export type { FadingTextAutoMessage, FadingTextSuppress, FadingTextTrigger } from '@boardgame/elements/components/boardgame-fading-text.js';
export type { StatusTextAutoMessage, StatusTextValue } from '@boardgame/elements/components/boardgame-status-text.js';
export type { StatStack } from '@boardgame/elements/components/boardgame-stat.js';
export type { ArtFit } from '@boardgame/elements/components/component-art.js';
export type { ActionBarAlignment, ActionBarOrientation } from '@boardgame/elements/components/boardgame-action-bar.js';
export type { TargetListLayout } from '@boardgame/elements/components/boardgame-target-list.js';
export type { ComponentZoneLayout } from '@boardgame/elements/components/boardgame-component-zone.js';
export type { MarketAttachmentPosition } from '@boardgame/elements/components/boardgame-market.js';
export type { StackSlotGeometry, StackSlotGeometryChangedDetail } from '@boardgame/elements/components/boardgame-component-stack.js';
export type { GameOutcomeViewer } from '@boardgame/elements/components/boardgame-game-outcome.js';
export type { BoardViewportChange } from '@boardgame/elements/components/boardgame-board-viewport.js';
export type { GridPlacementDraft } from '@boardgame/elements/components/boardgame-game-board.js';
export type { SpatialPlacementDraft } from '@boardgame/elements/components/boardgame-spatial-board.js';
export type { DraftControlsBinding } from '@boardgame/elements/components/boardgame-draft-controls.js';
export type {
  InspectorChangeReason,
  InspectorOpenChangedDetail,
} from '@boardgame/elements/components/boardgame-inspector.js';
export type { ReadinessView } from '@boardgame/elements/components/boardgame-readiness.js';
export type { ClientMove, JsonValue } from '@boardgame/elements/types/wire.js';
export type {
  ReadinessKey,
  ReadinessLabels,
  ReadinessParticipant,
  ReadinessPresentation,
  ReadinessState,
} from '@boardgame/elements/status/readiness.js';
export { readinessPresentation } from '@boardgame/elements/status/readiness.js';
export {
  SelectionDraftController,
  type SelectionDraftBinding,
  type SelectionDraftSelectionBinding,
  type SelectionDraftNotice,
  type SelectionOptionBinding,
  type SelectionDraftOptions,
  type SelectionDraftStateOptions,
  type SelectionDraftRebasePolicy,
} from './moves/selection-draft.js';
export {
  AdminPlayerIndex,
  AnyPlayerIndex,
  ObserverPlayerIndex,
  isConcretePlayerIndex,
  isKnownPlayerIndex,
  turnStatusPresentation,
} from '@boardgame/elements/status/turn-status.js';
export type {
  SpecialPlayerIndex,
  TurnStatusContext,
  TurnStatusKind,
  TurnStatusPresentation,
} from '@boardgame/elements/status/turn-status.js';
export type { PlayerPresentation } from '@boardgame/elements/status/player-presentation.js';
export { TimerController } from '@boardgame/elements/timers/timer-service.js';
export type {
  TimerCadence,
  TimerControllerOptions,
  TimerReading,
  TimerReference,
  TimerStatus,
} from '@boardgame/elements/timers/timer-service.js';
export type { TimerDisplayFormat } from '@boardgame/elements/components/boardgame-timer.js';
export { isStackLayout } from '@boardgame/elements/components/boardgame-component-stack.js';
export type { StackLayout } from '@boardgame/elements/components/boardgame-component-stack.js';
export type {
  MoveInputErrorCode,
  MoveInputIssue,
  MoveInputSchema,
  MoveInputSchemaField,
  MoveInputSchemaMove,
  MoveInputValidationResult,
} from '@boardgame/elements/moves/input.js';
export {
  defaultMessageResolver,
  defaultProjectedChoiceMessage,
  MOVE_CHOICE_PROJECTION_SCHEMA_VERSION,
  PROJECTION_FAILED_MESSAGE,
  projectedPlayerChoices,
  projectedStackChoices,
} from '@boardgame/elements/moves/projected-choices.js';
export type {
  MessageDescriptor,
  MessageResolver,
  MoveChoiceProjectionSchemaEntry,
  MoveChoiceProjectionTypes,
  ProjectedMoveChoiceCandidate,
  ProjectedMoveChoiceSet,
  ProjectedMoveChoices,
  ProjectedMoveChoicesStatus,
  ProjectedPlayerChoice,
  ProjectedPlayerChoices,
  ProjectedStackChoices,
} from '@boardgame/elements/moves/projected-choices.js';

// THE ONE CLASS A GAME MAY EXTEND.
//
// `componentView()` above is the documented escape hatch for a game-owned
// component host: its type parameter is `ElementType extends
// BoardgameComponent`, and its runtime guard refuses anything else with
// "create() must return a registered element extending BoardgameComponent".
// This export is what makes that reachable. Three documents (TUTORIAL.md twice,
// server/README.md once) had told creators to extend it while the single
// supported entry point declined to hand it over; the class's own doc comment
// carries the subclass contract those documents never stated.
export { BoardgameComponent } from '@boardgame/elements/components/boardgame-component.js';
export type { VisualMotionTrackInput } from '@boardgame/elements/motion/component-track.js';
export type { FlipRecord } from '@boardgame/elements/components/boardgame-component.js';

// Importing the facade registers every curated renderer primitive. Game code
// must not depend on transitive or deep component-module side effects. The
// CONCRETE primitives' classes -- card, token, die, stack and the rest -- are
// intentionally not exported: the supported creator contract for those is
// custom-element markup and the facade types, not implementation inheritance.
// `BoardgameComponent` above is the deliberate exception, and the only one.
import '@boardgame/elements/components/boardgame-die.js';
import '@boardgame/elements/components/boardgame-card.js';
import '@boardgame/elements/components/boardgame-token.js';
import '@boardgame/elements/components/boardgame-component-stack.js';
import '@boardgame/elements/components/boardgame-game-board.js';
import '@boardgame/elements/components/boardgame-player-badge.js';
import '@boardgame/elements/components/boardgame-fading-text.js';
import '@boardgame/elements/components/boardgame-action-button.js';
import '@boardgame/elements/components/boardgame-action-bar.js';
import '@boardgame/elements/components/boardgame-draft-controls.js';
import '@boardgame/elements/components/boardgame-inspector.js';
import '@boardgame/elements/components/boardgame-placement-item.js';
import '@boardgame/elements/components/boardgame-readiness.js';
import '@boardgame/elements/components/boardgame-selection-option.js';
import '@boardgame/elements/components/boardgame-target-list.js';
import '@boardgame/elements/components/boardgame-component-zone.js';
import '@boardgame/elements/components/boardgame-deck.js';
import '@boardgame/elements/components/boardgame-market.js';
import '@boardgame/elements/components/boardgame-game-outcome.js';
import '@boardgame/elements/components/boardgame-game-surface.js';
import '@boardgame/elements/components/boardgame-player-grid.js';
import '@boardgame/elements/components/boardgame-player-panel.js';
import '@boardgame/elements/components/boardgame-pips.js';
import '@boardgame/elements/components/boardgame-stat.js';
import '@boardgame/elements/components/boardgame-status-text.js';
import '@boardgame/elements/components/boardgame-timer.js';
import '@boardgame/elements/components/boardgame-turn-status.js';
import '@boardgame/elements/components/boardgame-spatial-board.js';
import '@boardgame/elements/components/boardgame-board-viewport.js';

export type { TrackStep } from '@boardgame/elements/components/boardgame-track.js';
import '@boardgame/elements/components/boardgame-track.js';
import '@boardgame/elements/components/boardgame-mat.js';
export type { DieRollBudget } from '@boardgame/elements/components/boardgame-die.js';

export { facetAvailable, viewGameProp, viewPlayerProp } from "@boardgame/elements/state/visibility.js";
export type { StateFacet, StateVisibility, ViewedValue } from "@boardgame/elements/state/visibility.js";
