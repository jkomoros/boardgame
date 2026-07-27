import { fx, html, css, isVisibleComponent } from '../../src/client.js';
import type { EffectSpec } from '../../src/client.js';
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

      .container {
        display: flex;
        flex-direction: row;
      }

      .flex {
        flex: 1;
      }
    `
  ];

  /**
   * Celebrate the number the player can actually read, at the moment they can
   * read it.
   *
   * This used to be an `effectsForTransition` plan, and that hook runs at CYCLE
   * START — the instant the move's snapshot installs, which is the instant the
   * die is thrown. Three things were wrong with that, and all three are
   * consequences of the die becoming a solid that flies:
   *
   *   - the pulse played over a die that was still tumbling, so it celebrated a
   *     number nobody could yet see;
   *   - it was anchored at the die's LAYOUT centre while the solid had travelled
   *     more than half its own width away from it, so the ring was not even
   *     around the die; and
   *   - on a 600ms version slot it was finished before the roll was.
   *
   * `roll-end` is the die saying it has stopped, and its detail carries the
   * value on the landed face — the one the player is now looking at. Firing
   * here also means a roll that is force-settled by the cycle sweep, or that
   * never animated at all (reduced motion), still gets its celebration: the die
   * dispatches `roll-end` on every one of those paths.
   *
   * `timing: 'immediate'`, not `'version'`: the roll it punctuates is itself
   * immediate (a physics bake must not be clamped into a version slot), so
   * there is no shared slot left to align a companion's copy to, and this is
   * local feedback about a die on this screen.
   */
  private _celebrateRoll(event: Event): void {
    const detail = (event as CustomEvent<{ value: number }>).detail;
    const value = detail?.value;
    if (typeof value !== 'number' || !Number.isFinite(value)) return;
    const die = this.state?.Game?.Die?.Components?.[0];
    if (!isVisibleComponent(die)) return;
    const maximum = Math.max(...die.Values.Faces);
    const tone = value === 1 ? 'warning' : value === maximum ? 'reward' : 'attention';
    // The die element itself, not a named anchor: it is the element that just
    // told us it stopped, and by then the solid has travelled back to the
    // middle of its own box, so its layout centre IS where the number is.
    const at = event.currentTarget as HTMLElement;
    const effects: EffectSpec[] = [fx.pulse({
      at,
      tone,
      intensity: value === 1 || value === maximum ? 'medium' : 'small',
    })];
    if (value === maximum) {
      effects.push(fx.burst({ at, tone: 'reward', intensity: 'small' }));
    }
    this.effects?.play(fx.parallel(effects, { key: 'roll-die', timing: 'immediate' }));
  }

  override render() {
    return html`
      <boardgame-game-surface heading="Pig">
        <div class="container">
          <boardgame-die
            .item="${this.state?.Game?.Die?.Components?.[0]}"
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
