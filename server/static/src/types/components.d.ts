/**
 * Type definitions for boardgame web components.
 *
 * These types define the interfaces for components that participate in the
 * FLIP animation system and the game rendering pipeline.
 */

import { AnimatingProps } from './animation';

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

  /**
   * Prepare for animation by applying inverse transforms.
   * This makes the component appear in its old position/state.
   * Called during INVERT phase of FLIP.
   */
  prepareAnimation(before: AnimatingProps, transform: string, opacity: string): void;

  /**
   * Start animation to final state by removing inverse transforms.
   * CSS transitions then animate to natural position/state.
   * Called during PLAY phase of FLIP.
   */
  startAnimation(after: AnimatingProps, transform: string, opacity: string): void;
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

  /** Whether to clone content during animation */
  cloneContent: boolean;

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
  /** Array of component data to render */
  stack: ComponentData[];

  /** Layout algorithm: 'stack', 'grid', 'fan', 'pile', 'spread' */
  layout: StackLayout;

  /** Deck name for rendering components */
  deck: string;

  /** Whether stack is interactive */
  interactive: boolean;

  /** Number of columns (for grid layout) */
  numCols: number;

  /** Spacing between components */
  spacing: number;

  /** If true, messy pile layout (random offsets) */
  messy: boolean;

  /** Angle for fan layout (degrees) */
  fanAngle: number;
}

/**
 * Stack layout algorithms.
 */
export type StackLayout = 'stack' | 'grid' | 'fan' | 'pile' | 'spread';

/**
 * Component data structure passed to stacks.
 */
export interface ComponentData {
  /** Component deck name */
  deck: string;

  /** Component index within deck */
  index: number;

  /** Additional rendering values */
  values?: Record<string, any>;

  /** Component ID for animation tracking */
  id?: string;
}

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
