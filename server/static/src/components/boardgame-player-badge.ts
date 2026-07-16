import { LitElement, css, html, nothing } from 'lit';
import { property } from 'lit/decorators.js';
import { styleMap } from 'lit/directives/style-map.js';
import type { PlayerPresentation } from '../status/player-presentation.js';

/** Inline player identity from an explicit sanitized renderer presentation. */
export class BoardgamePlayerBadge extends LitElement {
  static override styles = css`
    :host {
      display: inline-flex;
      align-items: center;
      gap: var(--boardgame-player-badge-gap, 0.25rem);
      vertical-align: middle;
      min-width: 0;
    }

    .avatar {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 1.5rem;
      height: 1.5rem;
      border-radius: 50%;
      background: var(--boardgame-player-badge-fallback, #6f685f);
      color: white;
      font-weight: 600;
      text-transform: uppercase;
      flex: none;
    }

    :host([compact]) .avatar {
      width: 1rem;
      height: 1rem;
      font-size: 0.5625rem;
    }

    .name {
      max-width: var(--boardgame-player-badge-name-width, 12rem);
      overflow: hidden;
      font-size: 0.8125rem;
      font-weight: 500;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    @media (forced-colors: active) {
      .avatar { border: 1px solid CanvasText; background: Canvas; color: CanvasText; }
    }
  `;

  @property({ attribute: false })
  player: PlayerPresentation | null = null;

  @property({ type: Boolean, reflect: true })
  compact = false;

  override render() {
    const player = this.#validatedPlayer();
    const initial = Array.from(player.label)[0]?.toLocaleUpperCase() ?? '?';
    const avatarLabel = this.compact ? player.label : nothing;
    return html`
      <span
        class="avatar"
        part="avatar"
        style=${styleMap({ backgroundColor: player.color })}
        role=${this.compact ? 'img' : nothing}
        aria-label=${avatarLabel}
        aria-hidden=${this.compact ? nothing : 'true'}>
        ${initial}
      </span>
      ${this.compact ? nothing : html`<span class="name" part="name">${player.label}</span>`}
    `;
  }

  #validatedPlayer(): PlayerPresentation {
    const player = this.player;
    if (typeof player !== 'object' || player === null) {
      throw new Error('boardgame-player-badge: .player must come from renderer.playerPresentation(index)');
    }
    if (!Number.isSafeInteger(player.playerIndex) || player.playerIndex < 0) {
      throw new Error('boardgame-player-badge: playerIndex must be a non-negative safe integer');
    }
    if (typeof player.label !== 'string' || !player.label.trim() || player.label.length > 200) {
      throw new Error('boardgame-player-badge: player label must be a non-empty string of at most 200 characters');
    }
    if (player.color !== undefined
      && (typeof player.color !== 'string' || !player.color.trim()
        || !CSS.supports('color', player.color))) {
      throw new Error(`boardgame-player-badge: player color is not valid CSS: ${JSON.stringify(player.color)}`);
    }
    return player;
  }
}

customElements.define('boardgame-player-badge', BoardgamePlayerBadge);

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-player-badge': BoardgamePlayerBadge;
  }
}
