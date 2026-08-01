import { html } from '../../src/client.js';
import { PlayerInfoRenderer, registerPlayerInfoRenderer } from './_game_renderer.js';

@registerPlayerInfoRenderer
export class BoardgameRenderPlayerInfoCheckers extends PlayerInfoRenderer {
  override render() {
    //boardgame-stat asks the stack for its own count (and capacity, if it has
    //one). Never count Indexes: a sized stack pads it with a -1 sentinel at
    //every empty slot, so its length is the capacity, not the count.
    return html`
      <boardgame-stat label="Cards" .stack=${this.playerState?.Hand ?? null}></boardgame-stat>
    `;
  }
}
