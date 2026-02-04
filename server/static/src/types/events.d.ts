/**
 * Type definitions for custom events used throughout the application.
 *
 * These extend the global HTMLElementEventMap to provide type safety
 * for event listeners and dispatchers.
 */

import { AnimationEvent } from './components';

/**
 * Event dispatched when a move is proposed.
 * Bubbles up to game-state-manager for submission.
 */
export interface ProposeMoveEvent extends CustomEvent {
  detail: {
    /** Move name (e.g., 'DrawCard', 'PlaceToken') */
    name: string;
    /** Move arguments as key-value pairs */
    arguments: Record<string, any>;
  };
}

/**
 * Event dispatched when a component is tapped/clicked.
 */
export interface ComponentTapEvent extends CustomEvent {
  detail: {
    /** Component deck name */
    deck: string;
    /** Component index */
    index: number;
    /** Additional component values */
    values?: Record<string, any>;
  };
}

/**
 * Event dispatched when a form value changes.
 */
export interface FormChangeEvent extends CustomEvent {
  detail: {
    /** Form field name */
    name: string;
    /** New value */
    value: any;
  };
}

/**
 * Extend global HTMLElementEventMap with custom events.
 * This provides type safety when using addEventListener/removeEventListener.
 */
declare global {
  interface HTMLElementEventMap {
    // Animation events (from components.d.ts)
    'will-animate': AnimationEvent;
    'animation-done': AnimationEvent;

    // Game events
    'propose-move': ProposeMoveEvent;
    'component-tap': ComponentTapEvent;

    // Form events
    'form-change': FormChangeEvent;
  }
}

// Export empty object to make this a module
export {};
