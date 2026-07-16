import { html } from 'lit';
import {
  BoardgameBasePlayerInfoRenderer,
  type PlayerChipPresentation,
} from './boardgame-base-player-info-renderer.js';

interface PlayerState {
  readonly token: string;
}

interface State {
  readonly Players: readonly PlayerState[];
}

class TypedPlayerInfo extends BoardgameBasePlayerInfoRenderer<State, PlayerState> {
  override get chip(): PlayerChipPresentation {
    return { text: this.playerState?.token ?? '', color: 'rebeccapurple' };
  }

  override render() {
    return html`${this.playerState?.token}`;
  }
}

const info = new TypedPlayerInfo();
info.state = { Players: [{ token: 'X' }] };
info.playerIndex = 0;
info.playerState?.token;

// @ts-expect-error playerState is derived and cannot drift from state/playerIndex
info.playerState = { token: 'O' };
// @ts-expect-error chip text must be a string
const invalidText: PlayerChipPresentation = { text: 1 };
// @ts-expect-error chip color must be a string
const invalidColor: PlayerChipPresentation = { color: false };
