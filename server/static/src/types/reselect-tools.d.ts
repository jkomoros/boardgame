/**
 * Type declarations for reselect-tools
 */

declare module 'reselect-tools/src' {
    export function getStateWith(fn: () => any): void;
    export function registerSelectors(selectors: any): void;
}
