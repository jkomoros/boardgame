/**
 * <bg-stack>: a declarative wrapper over `boardgame-component-stack`.
 *
 * In: attributes and JSON. Out: `propose-move`. It builds the real
 * `cardView({ render, properties })` recipe INSIDE trusted code from a
 * `face` description, and turns a `pick` description into the event a bound
 * move action used to deliver by calling a function.
 *
 *   <bg-stack path="Game.Cards" layout="grid" messy
 *             face="Type" pick="Reveal Card:CardIndex"></bg-stack>
 */
import { LitElement, html, css, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import '@boardgame/elements/components/boardgame-component-stack.js';
import type { BoardgameComponentStack } from '@boardgame/elements/components/boardgame-component-stack.js';
import { cardView, tokenView } from '@boardgame/elements/components/component-view.js';
import type { ComponentView } from '@boardgame/elements/components/component-view.js';
import { adoptEarlyProperties, findTable } from './bg-table.js';
import type { BgTable, TableChild } from './bg-table.js';
import { expandStack, isWireStack } from './expand.js';
import type { WireStack } from './expand.js';
import {
  faceProps, faceText, parseFace, parsePick, pickState, proposal, slotKind,
} from './spec.js';
import type { FaceSpec, PickSpec, SlotContext } from './spec.js';

export interface ProposeMoveDetail {
  readonly move: string;
  readonly args: Readonly<Record<string, unknown>>;
  /** Which stack and slot the person activated. */
  readonly stack?: string;
  readonly index?: number;
  /** The state version that was on screen when they did. */
  readonly version: number | null;
}

const REGION_SLOTS = new Set(['footer', 'corner']);
let anonymous = 0;

export class BgStack extends LitElement implements TableChild {
  static override styles = css`
    :host { display: block; }
    .face-text {
      font-family: var(--bg-face-font-family, inherit);
      font-size: var(--bg-face-font-size, 34px);
      line-height: var(--bg-face-line-height, 40px);
      font-weight: var(--bg-face-font-weight, 400);
      /* Face content is clipped to the card: it can never draw outside it. */
      overflow: hidden;
      max-width: 100%;
    }
    [role='button']:focus-visible {
      outline: 3px solid Highlight;
      outline-offset: 2px;
    }
  `;

  /** Dotted path of this stack in the table's installed state, e.g. `Players.0.WonCards`. */
  @property({ type: String }) path = '';

  /** Stable identity within the table. Defaults to `path`. */
  @property({ type: String }) name = '';

  @property({ type: String }) layout = 'stack';
  @property({ type: Boolean }) messy = false;
  @property({ type: Number }) messiness = 1;
  @property({ type: Number, attribute: 'faux-components' }) fauxComponents = 0;
  @property({ type: Number }) stagger = 0;
  @property({ type: Boolean, attribute: 'no-default-spacer' }) noDefaultSpacer = false;
  @property({ type: Number, attribute: 'post-animation-delay' }) postAnimationDelay = 0;

  /** Which component element the stack holds. */
  @property({ type: String }) component: 'card' | 'token' = 'card';

  /** How a component is drawn: a field name, `prop:Field` pairs, or a FaceSpec object. */
  @property({ attribute: 'face' }) face: string | FaceSpec | null = null;

  /** What activating a slot proposes: `Move Name:ArgName` or a PickSpec object. */
  @property({ attribute: 'pick' }) pick: string | PickSpec | null = null;

  /**
   * The state version `stack` belongs to, for callers that feed stacks
   * directly. A path-bound stack takes it from the table's installed frame.
   */
  @property({ type: Number }) version: number | null = null;

  private _stack: WireStack | null = null;

  /** A wire stack given directly, for callers that do not bind by `path`. */
  set stack(value: WireStack | null) {
    const old = this._stack;
    // Tell the table BEFORE anything changes, so it can still measure.
    this._table?.noteChange(false);
    this._stack = value;
    this.requestUpdate('stack', old);
  }

  get stack(): WireStack | null {
    return this._stack;
  }

  private _table: BgTable | null = null;
  private _view: ComponentView | null = null;
  private _viewTag = '';
  private _faceSpec: FaceSpec | null = null;
  private _pickSpec: PickSpec | null = null;
  private _expanded: { Components: readonly unknown[] } | null = null;
  private _key = '';
  private _trustedGesture = false;

  get key(): string {
    return this.name || this.path || this._key;
  }

  private get _inner(): BoardgameComponentStack | null {
    return this.renderRoot.querySelector('boardgame-component-stack');
  }

  override connectedCallback(): void {
    super.connectedCallback();
    adoptEarlyProperties(this, ['stack']);
    this._key ||= `bg-stack-${++anonymous}`;
    this._table = findTable(this.parentNode ?? this);
    this._table?.register(this);
    // A stack arriving is a structural change: nothing old is left to measure.
    this._table?.noteChange(true);
    this.renderRoot.addEventListener('click', this._onClick, true);
    this.renderRoot.addEventListener('component-tapped', this._onTapped);
    this.renderRoot.addEventListener('keydown', this._onKeyDown as EventListener);
  }

  override disconnectedCallback(): void {
    super.disconnectedCallback();
    this.renderRoot.removeEventListener('click', this._onClick, true);
    this.renderRoot.removeEventListener('component-tapped', this._onTapped);
    this.renderRoot.removeEventListener('keydown', this._onKeyDown as EventListener);
    // The table reference was cached at connect: by now this element has no
    // ancestors to walk.
    this._table?.noteChange(true);
    this._table?.unregister(this);
    this._table = null;
  }

  tableChanged(): void {
    this.requestUpdate();
  }

  protected override willUpdate(changed: Map<string, unknown>): void {
    if (changed.has('face')) this._faceSpec = parseFace(this.face);
    if (changed.has('pick')) this._pickSpec = parsePick(this.pick);
    const tag = this.component === 'token' ? 'boardgame-token' : 'boardgame-card';
    if (!this._view || tag !== this._viewTag) {
      this._viewTag = tag;
      // The function-valued recipe exists only here, inside the element.
      const options = {
        render: (context: { index: number; component: unknown }) => this._renderFace(context),
        properties: (context: { index: number; component: unknown }) =>
          this._propertiesFor(context) as never,
      };
      this._view = (tag === 'boardgame-token'
        ? tokenView(options as never) : cardView(options as never)) as ComponentView;
    }
    for (const name of ['layout', 'messy', 'messiness', 'face', 'fauxComponents']) {
      if (changed.has(name) && changed.get(name) !== undefined) {
        this._table?.noteChange(false);
        break;
      }
    }
    const wire = this.path ? this._table?.resolve(this.path) : this._stack;
    this._expanded = (this._table
      ? this._table.expand(wire)
      : isWireStack(wire) ? expandStack(wire, null, '') : null) as typeof this._expanded;
  }

  protected override firstUpdated(): void {
    const inner = this._inner!;
    // FINDING (S9/S12): a stack's id is random per element instance, and the
    // animator keys collection geometry and history by it. Give it the
    // declared, stable key so a recreated stack is the same collection.
    Object.defineProperty(inner, 'id', { configurable: true, get: () => this.key });
  }

  protected override updated(changed: Map<string, unknown>): void {
    const inner = this._inner;
    if (!inner) return;
    if ((changed.has('face') || changed.has('pick')) && inner.stack) {
      // Same stack, different recipe inputs: have the stack re-stamp its hosts.
      inner.stack = inner.stack;
    }
    void inner.updateComplete.then(() => this._applyAccessibility());
  }

  private _slot(index: number, component: unknown): SlotContext {
    return {
      index,
      kind: slotKind(component),
      component: (component ?? null) as SlotContext['component'],
    };
  }

  private _renderFace(context: { index: number; component: unknown }): unknown {
    const face = this._faceSpec;
    const slot = this._slot(context.index, context.component);
    if (!face || slot.kind !== 'visible') return nothing;
    const text = faceText(face.text, slot);
    const regions = Object.entries(face.regions ?? {})
      .filter(([name]) => REGION_SLOTS.has(name));
    // Text bindings only. Nothing a caller supplies is ever parsed as markup.
    return html`${text ? html`<div class="face-text" part="face-text">${text}</div>` : nothing}${
      regions.map(([name, source]) =>
        html`<span slot=${name} part="face-${name}">${faceText(source, slot)}</span>`)}`;
  }

  /** The pick with legality folded in: the frame's candidates unless the spec names its own. */
  private get _effectivePick(): PickSpec | null {
    const pick = this._pickSpec;
    if (!pick || pick.candidates || !pick.arg) return pick;
    const candidates = this._table?.candidates(pick.move, pick.arg);
    return candidates ? { ...pick, candidates } : pick;
  }

  private _propertiesFor(context: { index: number; component: unknown }): Record<string, unknown> {
    const slot = this._slot(context.index, context.component);
    const props = faceProps(this._faceSpec, this._viewTag, slot);
    // Display-only unless a pick makes this slot a target that is enabled now.
    props.disabled = !pickState(this._effectivePick, slot).enabled;
    return props;
  }

  private _components(): Array<HTMLElement & { index: number }> {
    const inner = this._inner;
    if (!inner) return [];
    return [...inner.children].filter(child =>
      child.hasAttribute('boardgame-component')) as Array<HTMLElement & { index: number }>;
  }

  private _applyAccessibility(): void {
    const slots = this._expanded?.Components ?? [];
    for (const element of this._components()) {
      const state = pickState(this._effectivePick, this._slot(element.index, slots[element.index]));
      if (state.target) {
        element.setAttribute('role', 'button');
        element.setAttribute('aria-disabled', String(!state.enabled));
        element.tabIndex = state.enabled ? 0 : -1;
      } else {
        element.removeAttribute('role');
        element.removeAttribute('aria-disabled');
        element.removeAttribute('tabindex');
      }
    }
  }

  // ---- gestures ---------------------------------------------------------

  private readonly _onClick = (event: Event): void => {
    // `component-tapped` is synthesized by the component inside its own click
    // handler, so it cannot say whether a person clicked. Remember that here.
    this._trustedGesture = event.isTrusted;
  };

  private readonly _onTapped = (event: Event): void => {
    event.stopPropagation();
    const trusted = this._trustedGesture;
    this._trustedGesture = false;
    if (!trusted) return;
    const element = event.composedPath().find((target): target is HTMLElement =>
      target instanceof HTMLElement && target.hasAttribute('boardgame-component'));
    if (element) this._activate((element as HTMLElement & { index: number }).index);
  };

  private readonly _onKeyDown = (event: KeyboardEvent): void => {
    if (!event.isTrusted || (event.key !== 'Enter' && event.key !== ' ')) return;
    const element = event.composedPath().find((target): target is HTMLElement =>
      target instanceof HTMLElement && target.hasAttribute('boardgame-component'));
    if (!element) return;
    event.preventDefault();
    this._activate((element as HTMLElement & { index: number }).index);
  };

  private _activate(index: number): void {
    const pick = this._effectivePick;
    if (!pick) return;
    const slot = this._slot(index, this._expanded?.Components[index]);
    if (!pickState(pick, slot).enabled) return;
    const detail: ProposeMoveDetail = Object.freeze({
      ...proposal(pick, slot),
      stack: this.key,
      index,
      version: this._table?.view?.Version ?? this.version ?? null,
    });
    this.dispatchEvent(new CustomEvent('propose-move', { detail, bubbles: true, composed: true }));
  }

  override render() {
    return html`
      <boardgame-component-stack
        layout=${this.layout}
        ?messy=${this.messy}
        .messiness=${this.messiness}
        .fauxComponents=${this.fauxComponents}
        .stagger=${this.stagger}
        ?no-default-spacer=${this.noDefaultSpacer}
        post-animation-delay=${this.postAnimationDelay > 0 ? this.postAnimationDelay : nothing}
        .componentView=${this._view}
        .stack=${this._expanded}><slot></slot></boardgame-component-stack>
    `;
  }
}

customElements.define('bg-stack', BgStack);

declare global {
  interface HTMLElementTagNameMap { 'bg-stack': BgStack; }
}
