/**
 * Type definitions for boardgame web components.
 *
 * These types define the interfaces for components that participate in the
 * FLIP animation system and the game rendering pipeline.
 */

import type { AnimatingProps } from './animation';
import type { ExpandedStack } from './boardgame-types';
import type { ComponentView } from '../components/component-view';
import type { BoundMoveAction } from '../moves/action';

/**
 * Base interface for animatable components.
 * These components participate in the FLIP animation system.
 */
export interface BoardgameAnimatableItemElement extends HTMLElement {
  /** Unique identifier for the component */
  id: string;

  /** If true, skip animations for this component */
  noAnimate: boolean;

  /** List of property names that should animate (e.g., ['rotated', 'faceUp']) */
  animatingProperties: string[];

  /**
   * Returns current values of all animating properties.
   * Called during FIRST phase of FLIP to capture state before change.
   */
  animatingPropValues(): AnimatingProps;
  animatingPropDefaults(stack: any): AnimatingProps;
}

/**
 * Base interface for game components (cards, tokens, etc).
 * Extends animatable with additional rendering properties.
 */
export interface BoardgameComponentElement extends BoardgameAnimatableItemElement {
  /** Deck name (e.g., 'cards', 'tiles') */
  deck: string;

  /** Index within deck */
  index: number;

  /** Additional values for rendering (game-specific) */
  values: Record<string, any>;

  /** If true, component is disabled/inactive */
  disabled: boolean;

  /** If true, component is interactive (clickable) */
  interactive: boolean;

  /** Scale factor for rendering */
  scale: number;

  /** If true, apply shadow effect */
  shadow: boolean;

  /** If true, use alternate shadow (for rotated cards) */
  altShadow: boolean;

  /** Private historical-presentation capture policy. */
  historicalPresentationPolicy: 'none' | 'clone-default-slot';

  /** Prepare a fresh inert host for temporary departing motion. */
  prepareMotionCarrier(defaults: Readonly<Record<string, unknown>>): void;

  /**
   * Compute properties that should animate between states.
   * Used for complex animations like card flips.
   */
  computeAnimatingProps(): AnimatingProps;

  /**
   * Returns true if component rotates between before/after states.
   * Used to optimize transform calculations during FLIP.
   */
  animationRotates(before: AnimatingProps, after: AnimatingProps): boolean;
}

/**
 * Card component interface.
 * Supports front/back rendering and rotation animations.
 */
export interface BoardgameCardElement extends BoardgameComponentElement {
  /** If true, card shows front face */
  faceUp: boolean;

  /** If true, card is rotated 90 degrees */
  rotated: boolean;

  /** If true, use spacer deck for back rendering */
  spacerDeck: string;
}

/**
 * Token component interface.
 * Simpler than cards - typically circular with no rotation.
 */
export interface BoardgameTokenElement extends BoardgameComponentElement {
  /** If true, token is in active state */
  active: boolean;

  /** If true, token is highlighted */
  highlighted: boolean;
}

/**
 * Component stack interface.
 * Container that lays out multiple components with animations.
 */
export interface BoardgameComponentStackElement extends BoardgameAnimatableItemElement {
  /** Expanded stack snapshot to render, or null while no stack is selected. */
  stack: ExpandedStack | null | undefined;

  /** Layout algorithm used to position the stack's components. */
  layout: StackLayout;

  /** Deck and game names inferred from the current stack. */
  deckName: string;
  gameName: string;

  /** Last-seen counters used by the animation coordinator. */
  idsLastSeen: Readonly<Record<string, number>> | null;

  /** Whether to apply deterministic random rotation outside board layouts. */
  messy: boolean;
  messiness: number;

  /** Whether an empty stack should omit its default spacer component. */
  noDefaultSpacer: boolean;

  /** Number of columns used by the board layout. */
  boardCols: number;

  /** Number of rows used by the board layout. */
  boardRows: number;

  /** Pixel positions used by the spatial layout, indexed by stack slot. */
  spatialPositions: Array<{ top: number; left: number } | null>;

  /** Number of visual placeholder components added to the rendered stack. */
  fauxComponents: number;

  /** Per-component animation delay, expressed as a fraction of animation length. */
  stagger: number;

  /** Whether every component is disabled for a display-only stack. */
  componentsDisabled: boolean;

  /** Explicit untyped escape hatch; prefer componentView.withProperties(...). */
  unsafeComponentAttrs: Record<string, unknown>;

  /** Renderer-scoped typed Lit recipe for component hosts and content. */
  componentView: ComponentView | null;

  /** One exact bound action or explicit null for each logical stack slot. */
  componentActions: readonly (BoundMoveAction<string, object> | null)[];

  /** Currently rendered real and faux component elements. */
  readonly Components: BoardgameComponentElement[];
}

/**
 * Stack layout algorithms.
 */
export type { StackLayout } from '../components/boardgame-component-stack';

/**
 * Component animator interface.
 * Orchestrates FLIP animations for all components in a container.
 */
export interface BoardgameComponentAnimatorElement extends HTMLElement {
  /**
   * Prepare for animation by capturing current state.
   * Called BEFORE state changes (FIRST phase of FLIP).
   */
  prepare(): void;

  /**
   * Execute animation after state change.
   * Returns promise that resolves when all animations complete.
   * Called AFTER state changes (LAST, INVERT, PLAY phases).
   */
  animate(): Promise<void>;

  /**
   * Version counter for faux animating components.
   * Incremented each animation cycle to track component sources.
   */
  idsLastSeen: Map<string, number>;
}

/**
 * Custom events dispatched by animating components.
 */
export interface AnimationEvent extends CustomEvent {
  detail: {
    ele: HTMLElement;
  };
}

declare global {
  interface HTMLElementEventMap {
    'will-animate': AnimationEvent;
    'animation-done': AnimationEvent;
  }
}
