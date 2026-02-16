/**
 * Type definitions for Polymer elements during migration.
 * These bridge types allow TypeScript to compile while Polymer components
 * are being incrementally migrated to Lit 3.
 */

declare module '@polymer/polymer/polymer-element.js' {
  export class PolymerElement extends HTMLElement {
    static get properties(): any;
    static get template(): any;
    static get observers(): string[];
    ready(): void;
    connectedCallback(): void;
    disconnectedCallback(): void;
    $: any;
    root: ShadowRoot | null;
    shadowRoot: ShadowRoot | null;
    updateStyles(properties?: { [key: string]: string }): void;
    fire(type: string, detail?: any, options?: any): CustomEvent;
    dispatchEvent(event: Event): boolean;
    addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void;
    removeEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void;
  }
}

declare module '@polymer/polymer/lib/utils/html-tag.js' {
  export function html(strings: TemplateStringsArray, ...values: any[]): any;
}

declare module '@polymer/lit-element' {
  import { LitElement as LitElementBase } from 'lit';
  export { LitElementBase as LitElement };
  export { html, css } from 'lit';
}
