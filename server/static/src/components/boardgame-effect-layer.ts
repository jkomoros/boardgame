import { css, html } from 'lit';
import type { AnimationTimingPolicy } from './boardgame-animatable-item.js';
import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { burstParticles } from '../effects/particle-burst.js';

const MAX_ACTIVE_EFFECTS = 8;
const MAX_ACTIVE_PARTICLES = 60;

export type EffectPreset = 'impact' | 'score' | 'success' | 'celebrate';
export type EffectAnchor = string | HTMLElement | { x: number; y: number };

export interface BurstEffectOptions {
  preset?: EffectPreset;
  /** Number of particles. Values are clamped to the framework budget of 24. */
  count?: number;
  /** Maximum travel distance in CSS pixels. Clamped to 8-240. */
  spread?: number;
  /** Total effect duration in milliseconds. Clamped to 120-1200. */
  duration?: number;
  /** Stable seed for repeatable layouts. Defaults to preset + anchor position. */
  seed?: string | number;
  /** Optional CSS colors; semantic Material theme colors are used by default. */
  colors?: readonly string[];
  /** Local delight is immediate by default; choose version for companion sync. */
  timing?: AnimationTimingPolicy;
}

export interface EffectHandle {
  readonly finished: Promise<void>;
  cancel(): void;
}

export interface EffectHostAPI {
  burst(anchor: EffectAnchor, options?: BurstEffectOptions): EffectHandle;
  cancelAll(): void;
}

const PRESET_DEFAULTS: Record<EffectPreset, {
  count: number;
  spread: number;
  duration: number;
  colors: readonly string[];
}> = {
  impact: { count: 8, spread: 42, duration: 320, colors: [
    'var(--boardgame-effect-primary, var(--md-sys-color-primary, #6750a4))',
    'var(--boardgame-effect-glow, var(--md-sys-color-primary-container, #eaddff))',
  ] },
  score: { count: 12, spread: 70, duration: 520, colors: [
    'var(--boardgame-effect-score, #f6bf26)',
    'var(--boardgame-effect-accent, var(--md-sys-color-tertiary, #7d5260))',
  ] },
  success: { count: 14, spread: 78, duration: 560, colors: [
    'var(--boardgame-effect-success, var(--md-sys-color-primary, #2e6b4f))',
    'var(--boardgame-effect-glow, var(--md-sys-color-primary-container, #eaddff))',
  ] },
  celebrate: { count: 20, spread: 110, duration: 720, colors: [
    'var(--boardgame-effect-primary, var(--md-sys-color-primary, #6750a4))',
    'var(--boardgame-effect-accent, var(--md-sys-color-tertiary, #7d5260))',
    'var(--boardgame-effect-score, #f6bf26)',
    'var(--boardgame-effect-success, var(--md-sys-color-primary, #2e6b4f))',
  ] },
};

function finiteOr(value: number | undefined, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function findInOpenRoots(root: Document | ShadowRoot, id: string): HTMLElement | null {
  const escaped = CSS.escape(id);
  const direct = root.getElementById?.(id)
    ?? root.querySelector<HTMLElement>(`[data-effect-anchor="${escaped}"]`);
  if (direct) return direct as HTMLElement;
  for (const element of root.querySelectorAll<HTMLElement>('*')) {
    if (element.shadowRoot) {
      const nested = findInOpenRoots(element.shadowRoot, id);
      if (nested) return nested;
    }
  }
  return null;
}

export class BoardgameEffectLayer extends BoardgameAnimatableItem implements EffectHostAPI {
  static override styles = css`
    :host {
      position: fixed;
      inset: 0;
      z-index: 1000;
      overflow: hidden;
      pointer-events: none;
      contain: strict;
    }

    #particles {
      position: absolute;
      inset: 0;
      overflow: hidden;
      pointer-events: none;
    }

    .particle {
      position: absolute;
      width: var(--particle-size);
      height: var(--particle-size);
      border-radius: 999px;
      background: var(--particle-color);
      box-shadow: 0 0 calc(var(--particle-size) * 1.4) var(--particle-color);
      will-change: transform, opacity;
    }

    .particle:nth-child(3n) {
      border-radius: 2px;
    }
  `;

  private readonly _activeCancels = new Set<() => void>();
  private _activeParticleCount = 0;

  override render() {
    return html`<div id="particles" aria-hidden="true"></div>`;
  }

  override disconnectedCallback(): void {
    this.cancelAll();
    super.disconnectedCallback();
  }

  cancelAll(): void {
    for (const cancel of [...this._activeCancels]) cancel();
  }

  burst(anchor: EffectAnchor, options: BurstEffectOptions = {}): EffectHandle {
    const point = this._anchorPoint(anchor);
    if (!point || window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return { finished: Promise.resolve(), cancel: () => {} };
    }

    const requestedPreset = options.preset;
    const preset: EffectPreset = requestedPreset
      && Object.prototype.hasOwnProperty.call(PRESET_DEFAULTS, requestedPreset)
      ? requestedPreset
      : 'impact';
    const defaults = PRESET_DEFAULTS[preset];
    const duration = Math.max(120, Math.min(1200, finiteOr(options.duration, defaults.duration)));
    const requestedColors = Array.isArray(options.colors)
      ? options.colors.filter(color => typeof color === 'string' && color.trim())
      : [];
    const colors = requestedColors.length ? requestedColors : defaults.colors;

    // Keep repeated clicks and effect-heavy games within one predictable DOM
    // budget. Oldest decorative flourishes yield first; game state never does.
    while (
      this._activeCancels.size >= MAX_ACTIVE_EFFECTS
      || this._activeParticleCount >= MAX_ACTIVE_PARTICLES
    ) {
      const oldest = this._activeCancels.values().next().value as (() => void) | undefined;
      if (!oldest) break;
      oldest();
    }
    const requestedCount = Math.min(
      finiteOr(options.count, defaults.count),
      MAX_ACTIVE_PARTICLES - this._activeParticleCount,
    );
    const layout = burstParticles(
      requestedCount,
      finiteOr(options.spread, defaults.spread),
      options.seed ?? `${preset}:${Math.round(point.x)}:${Math.round(point.y)}`,
    );
    const container = this.shadowRoot?.querySelector<HTMLElement>('#particles');
    if (!container) return { finished: Promise.resolve(), cancel: () => {} };

    const particles: HTMLElement[] = [];
    const animations: Animation[] = [];
    let resolveFinished!: () => void;
    let settled = false;
    const finished = new Promise<void>(resolve => { resolveFinished = resolve; });
    const cleanup = () => {
      if (settled) return;
      settled = true;
      this._activeCancels.delete(cancel);
      this._activeParticleCount = Math.max(0, this._activeParticleCount - layout.length);
      for (const particle of particles) particle.remove();
      resolveFinished();
    };
    const cancel = () => {
      for (const animation of animations) animation.cancel();
      cleanup();
    };
    this._activeCancels.add(cancel);
    this._activeParticleCount += layout.length;

    layout.forEach((particle, index) => {
      const element = document.createElement('i');
      element.className = 'particle';
      element.style.left = `${point.x}px`;
      element.style.top = `${point.y}px`;
      element.style.setProperty('--particle-size', `${particle.sizePx}px`);
      element.style.setProperty('--particle-color', colors[index % colors.length]);
      container.appendChild(element);
      particles.push(element);

      const radians = particle.angleDeg * Math.PI / 180;
      const dx = Math.cos(radians) * particle.distancePx;
      const dy = Math.sin(radians) * particle.distancePx;
      const animation = this.play(element, [
        { transform: 'translate(-50%, -50%) scale(0.35)', opacity: 0 },
        { transform: 'translate(-50%, -50%) scale(1)', opacity: 1, offset: 0.16 },
        { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) rotate(${particle.rotationDeg}deg) scale(0)`, opacity: 0 },
      ], {
        duration,
        delay: particle.delayMs,
        easing: 'cubic-bezier(0.2, 0.8, 0.2, 1)',
        fill: 'both',
      }, {
        gated: false,
        timing: options.timing ?? 'immediate',
      });
      if (animation) animations.push(animation);
      else element.remove();
    });

    if (animations.length === 0) cleanup();
    else Promise.all(animations.map(animation => animation.finished.catch(() => {}))).then(cleanup);
    return { finished, cancel };
  }

  private _anchorPoint(anchor: EffectAnchor): { x: number; y: number } | null {
    if (typeof anchor === 'object' && !(anchor instanceof HTMLElement)) {
      return Number.isFinite(anchor.x) && Number.isFinite(anchor.y) ? anchor : null;
    }
    const root = this.getRootNode();
    const element = typeof anchor === 'string'
      ? (root instanceof Document || root instanceof ShadowRoot
        ? findInOpenRoots(root, anchor)
        : null)
      : anchor;
    if (!element?.isConnected) {
      console.warn('[effects] burst: could not resolve anchor', anchor);
      return null;
    }
    const rect = element.getBoundingClientRect();
    return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
  }
}

customElements.define('boardgame-effect-layer', BoardgameEffectLayer);
