import { LitElement, type PropertyValues } from 'lit';
import { property } from 'lit/decorators.js';

export interface PlayerChipPresentation {
  readonly text?: string;
  readonly color?: string;
}

export interface PlayerChipPresentationChangedDetail {
  readonly text: string;
  readonly color: string;
}

interface PlayerInfoState<PlayerState> {
  readonly Players: readonly PlayerState[];
}

const CHIP_PRESENTATION_CHANGED = 'player-chip-presentation-changed';

/** Typed lifecycle shared by every generated game-specific player summary. */
export abstract class BoardgameBasePlayerInfoRenderer<
  State extends PlayerInfoState<PlayerState>,
  PlayerState,
> extends LitElement {
  @property({ type: Object })
  state: State | null = null;

  @property({ type: Number, attribute: 'player-index' })
  playerIndex = 0;

  private _lastChipPresentation: PlayerChipPresentationChangedDetail | null = null;

  /** The exact player substate selected by `playerIndex`; never separately synchronized. */
  get playerState(): PlayerState | null {
    this._validatePlayerIndex();
    if (!this.state) return null;
    return this.state.Players[this.playerIndex] ?? null;
  }

  /** Override to customize the roster chip without dispatching events manually. */
  get chip(): PlayerChipPresentation {
    return {};
  }

  protected override updated(changedProperties: PropertyValues<this>): void {
    super.updated(changedProperties);
    this._validatePlayerIndex();
    const next = this._readChipPresentation();
    if (next.text === this._lastChipPresentation?.text && next.color === this._lastChipPresentation?.color) return;
    this._lastChipPresentation = next;
    this.dispatchEvent(new CustomEvent<PlayerChipPresentationChangedDetail>(CHIP_PRESENTATION_CHANGED, {
      bubbles: true,
      composed: true,
      detail: next,
    }));
  }

  private _validatePlayerIndex(): void {
    if (!Number.isSafeInteger(this.playerIndex) || this.playerIndex < 0) {
      throw new Error('player-info renderer: playerIndex must be a non-negative safe integer');
    }
    if (this.state && !Array.isArray(this.state.Players)) {
      throw new Error('player-info renderer: state.Players must be an array');
    }
    if (this.state && this.playerIndex >= this.state.Players.length) {
      throw new Error(
        `player-info renderer: playerIndex ${this.playerIndex} is outside the ${this.state.Players.length}-player state`,
      );
    }
  }

  private _readChipPresentation(): PlayerChipPresentationChangedDetail {
    const chip: unknown = this.chip;
    if (!chip || typeof chip !== 'object' || Array.isArray(chip)) {
      throw new Error('player-info renderer: chip must return an object with optional text and color strings');
    }
    const record = chip as Readonly<Record<string, unknown>>;
    const extra = Object.keys(record).filter(key => key !== 'text' && key !== 'color');
    if (extra.length > 0) {
      throw new Error(`player-info renderer: chip contains unknown field${extra.length === 1 ? '' : 's'} ${extra.join(', ')}`);
    }
    if (record['text'] !== undefined && typeof record['text'] !== 'string') {
      throw new Error('player-info renderer: chip.text must be a string when provided');
    }
    if (record['color'] !== undefined && typeof record['color'] !== 'string') {
      throw new Error('player-info renderer: chip.color must be a CSS color string when provided');
    }
    const text = typeof record['text'] === 'string' ? record['text'] : '';
    const color = typeof record['color'] === 'string' ? record['color'].trim() : '';
    if (color && !CSS.supports('color', color)) {
      throw new Error(`player-info renderer: chip.color ${JSON.stringify(color)} is not a valid CSS color`);
    }
    return Object.freeze({ text, color });
  }
}
