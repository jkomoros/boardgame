import { LitElement } from 'lit';
import { property } from 'lit/decorators.js';

/** Typed lifecycle shared by every generated game-specific player summary. */
export abstract class BoardgameBasePlayerInfoRenderer<State, PlayerState> extends LitElement {
  @property({ type: Object })
  state: State | null = null;

  @property({ type: Number })
  playerIndex = 0;

  @property({ type: Object })
  playerState: PlayerState | null = null;
}
