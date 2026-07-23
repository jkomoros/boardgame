/**
 * Type definitions for the FLIP animation system.
 *
 * The boardgame framework uses FLIP (First, Last, Invert, Play) animations
 * to create smooth transitions when components move or change state.
 */

/**
 * Properties that can be animated on a component.
 * These are passed to and from animating components during the FLIP cycle.
 */
export interface AnimatingProps {
  [key: string]: any;
}

/**
 * Rectangle representing an element's offset position and dimensions.
 * Used in the FLIP animation system to calculate transforms.
 */
export interface OffsetRect {
  top: number;
  left: number;
  width: number;
  height: number;
}

/**
 * Record tracking animation state for a component during FLIP.
 */
export interface ComponentAnimationRecord {
  /** Offset rectangle before the state change */
  beforeOffset: OffsetRect;
  /** Offset rectangle after the state change */
  afterOffset: OffsetRect;
  /** Property values before the state change */
  beforeProps: AnimatingProps;
  /** Property values after the state change */
  afterProps: AnimatingProps;
  /** External transform applied by parent (e.g., stack layout) */
  beforeTransform: string;
  /** External transform applied by parent after state change */
  afterTransform: string;
  /** Opacity before the state change */
  beforeOpacity: string;
  /** Opacity after the state change */
  afterOpacity: string;
}

/**
 * Faux component created to animate items from unknown sources.
 * Used when PolicyLen sanitization removes components from the DOM
 * and we need to animate them out from their last known position.
 */
export interface FauxComponent {
  ele: HTMLElement;
  version: number;
  inUse: boolean;
}

/**
 * Interface for components that participate in FLIP animations.
 */
export interface AnimatingComponent extends HTMLElement {
  /** Unique identifier for the component */
  id: string;
  /** If true, this component should not animate */
  noAnimate: boolean;
  /** List of property names that should be animated */
  animatingProperties: string[];
  /** Returns current values of all animating properties */
  animatingPropValues(): AnimatingProps;
  /** Private historical-presentation capture policy. */
  historicalPresentationPolicy?: 'none' | 'clone-default-slot' | 'clone-default-slot-safe';
  motionEndpointOrientation?(state: AnimatingProps): 'natural' | 'quarter-turned';
  legacyAnimationRotationRequested?(before: AnimatingProps, after: AnimatingProps): boolean | null;
}
