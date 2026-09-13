import { fx, html, css, isVisibleComponent } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { MoveNames } from './_move_names.js';

@registerGameRenderer
export class BoardgameRenderGamePig extends GameRenderer {
  static override styles = [
    ...(GameRenderer.styles ? [GameRenderer.styles] : []),
    css`
      /*
       * The die's size is its own custom property: it draws a solid inside a
       * box of this size, so height/width on the element (which this rule
       * used to set, on a .die class nothing ever carried) cannot resize it.
       */
      boardgame-die {
        --die-size: 100px;
      }
    `
  ];

  /** Add one restrained cue only for Pig's best possible roll. */
  private _celebrateRoll(event: Event): void {
    const detail = (event as CustomEvent<{ value: number }>).detail;
    const value = detail?.value;
    if (typeof value !== 'number' || !Number.isFinite(value)) return;
    const die = this.state?.Game?.Die?.Components?.[0];
    if (!isVisibleComponent(die)) return;
    const maximum = Math.max(...die.Values.Faces);
    if (value !== maximum) return;
    this.effects?.play(fx.pulse({
      at: event.currentTarget as HTMLElement,
      tone: 'reward',
      intensity: 'subtle',
      key: 'maximum-roll',
      timing: 'immediate',
    }));
  }

  override render() {
    return html`
      <boardgame-game-surface heading="Pig">
        <div class="horizontal">
          <boardgame-die
            .item="${this.state?.Game.Die.Components[0] ?? null}"
            .action="${this.move(MoveNames.RollDice)}"
            @roll-end="${this._celebrateRoll}">
          </boardgame-die>
          <div class="flex"></div>
          <boardgame-action-button .action="${this.move(MoveNames.DoneTurn)}">
            Done
          </boardgame-action-button>
        </div>
        <boardgame-turn-status
          slot="status"
          .turn=${this.turnStatus}>
        </boardgame-turn-status>
      </boardgame-game-surface>
    `;
  }
}
