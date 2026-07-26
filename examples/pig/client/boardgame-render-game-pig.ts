import { fx, html, css, isVisibleComponent } from '../../src/client.js';
import type { EffectSpec, EffectTransitionContext } from '../../src/client.js';
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';
import { MoveNames } from './_move_names.js';
import type { MoveName } from './_move_names.js';
import type { State } from './_types.js';

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

      .container {
        display: flex;
        flex-direction: row;
      }

      .flex {
        flex: 1;
      }
    `
  ];

  override effectsForTransition(
    context: EffectTransitionContext<State, MoveName>,
  ): readonly EffectSpec[] {
    if (context.kind === 'initial' || context.move?.AnimationKey !== MoveNames.RollDice) return [];
    const die = context.after.Game.Die.Components[0];
    if (!isVisibleComponent(die)) return [];
    const value = die.DynamicValues?.Value;
    if (value === undefined) return [];
    const maximum = Math.max(...die.Values.Faces);
    const tone = value === 1 ? 'warning' : value === maximum ? 'reward' : 'attention';
    const effects: EffectSpec[] = [fx.pulse({
      at: fx.anchor('pig-die'),
      tone,
      intensity: value === 1 || value === maximum ? 'medium' : 'small',
    })];
    if (value === maximum) {
      effects.push(fx.burst({
        at: fx.anchor('pig-die'),
        tone: 'reward',
        intensity: 'small',
      }));
    }
    return [fx.parallel(effects, { key: 'roll-die', timing: 'version' })];
  }

  override render() {
    return html`
      <boardgame-game-surface heading="Pig">
        <div class="container">
          <boardgame-die
            data-effect-anchor="pig-die"
            .item="${this.state?.Game?.Die?.Components?.[0]}"
            .action="${this.move(MoveNames.RollDice)}">
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
