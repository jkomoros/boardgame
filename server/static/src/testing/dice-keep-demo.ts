/** Interactive presentation fixture: installs deterministic snapshots, no game rules. */
import { LitElement, html, css } from 'lit';
import { property, query } from 'lit/decorators.js';
import { dieView } from '../components/component-view.js';
import { SelectionDraftController } from '../moves/selection-draft.js';
import type { BoardgameComponentAnimator } from '../components/boardgame-component-animator.js';
import type { ExpandedStack, VisibleComponent } from '../types/boardgame-types.js';
import '../components/boardgame-component-zone.js';
import '../components/boardgame-component-animator.js';

type DemoDie = VisibleComponent<{ Faces: number[] }, { SelectedFace: number; Value: number; RollCount: number }>;
const makeDie = (index: number): DemoDie => ({ ID: `die-${index}`, Index: index, Deck: 'dice', GameName: 'dice-demo',
  Values: { Faces: [1, 2, 3, 4, 5, 6] }, DynamicValues: { SelectedFace: index % 6, Value: index % 6 + 1, RollCount: 0 } });
function stack(items: readonly DemoDie[]): ExpandedStack {
  return { Deck: 'dice', GameName: 'dice-demo', Components: [...items], IDs: items.map(d => d.ID ?? ''),
    Indexes: items.map((_, index) => index), IDsLastSeen: {}, ShuffleCount: 0, Size: items.length };
}

export class DiceKeepDemo extends LitElement {
  static styles = css`
    :host { display:block; max-width:900px; margin:2rem auto; padding:1rem; color:#302d27; font:17px system-ui; }
    h1 { font-size:1.6rem; } p { line-height:1.5; }
    .zones { display:grid; gap:1rem; grid-template-columns:1fr 1fr; margin:1.5rem 0; }
    boardgame-component-zone { --component-width:64px; min-width:0; }
    button { padding:.7rem 1rem; margin:.25rem; font:inherit; border:1px solid #817668; border-radius:.5rem; background:#fffaf1; color:inherit; }
    button:disabled { opacity:.5; }
    @media(max-width:600px) { .zones { grid-template-columns:1fr; } }
  `;
  @property({ attribute: false }) state = { tray: Array.from({ length: 5 }, (_, i) => makeDie(i)), kept: [] as DemoDie[] };
  @property({ type: Number }) gameVersion = 0;
  @property({ type: Boolean }) busy = false;
  readonly gameName = 'dice-demo'; readonly gameId = 'fixture'; readonly snapshotEpoch = 0;
  readonly viewingAsPlayer = 0; readonly proposingAsPlayer = 0; readonly proposingAsAdmin = false;
  private readonly selection = new SelectionDraftController<string>(this);
  private readonly view = dieView({ rollBudget: { durationMs: 900, maxSolidDice: 5 } });
  @query('boardgame-component-animator') private animator!: BoardgameComponentAnimator;

  private binding() { return this.selection.draft({ candidates: this.state.tray.map(d => d.ID ?? ''), maxSelected: 20, rebase: 'keep-valid' }); }

  async install(next: typeof this.state): Promise<void> {
    this.busy = true;
    this.animator.prepare();
    this.state = next; this.gameVersion++;
    await this.updateComplete;
    for (const zone of this.renderRoot.querySelectorAll('boardgame-component-zone')) {
      await zone.updateComplete;
      await zone.shadowRoot!.querySelector('boardgame-component-stack')!.updateComplete;
    }
    const dice = [...this.renderRoot.querySelectorAll('boardgame-component-zone')].flatMap(zone =>
      [...zone.shadowRoot!.querySelectorAll('boardgame-die')]);
    for (let i = 0; i < 4; i++) await Promise.all(dice.map(d => d.updateComplete));
    await Promise.all([this.animator.animateFlip(), ...dice.map(d => d.settled())]);
    this.busy = false;
  }

  private reroll(repeated: boolean): void {
    const tray = this.state.tray.map(die => {
      const face = repeated ? die.DynamicValues!.SelectedFace : (die.DynamicValues!.SelectedFace + 1) % 6;
      return { ...die, DynamicValues: { SelectedFace: face, Value: face + 1, RollCount: (die.DynamicValues!.RollCount ?? 0) + 1 } };
    });
    void this.install({ ...this.state, tray });
  }

  private keep(): void {
    const selected = new Set(this.binding().selected);
    void this.install({ tray: this.state.tray.filter(d => !selected.has(d.ID ?? '')),
      kept: [...this.state.kept, ...this.state.tray.filter(d => selected.has(d.ID ?? ''))] });
  }

  override render() {
    const selection = this.binding();
    return html`
      <h1>Keep and reroll dice</h1>
      <p>Select dice in the tray, then keep them. Rerolls change only tray counters; repeated results still roll.
      This fixture installs sample snapshots to demonstrate presentation and selection.</p>
      <div class="zones">
        <boardgame-component-zone no-default-spacer label="Tray" layout=${this.state.tray.length > 5 ? 'grid' : 'spread'} .stack=${stack(this.state.tray)} .componentView=${this.view} .selection=${selection}></boardgame-component-zone>
        <boardgame-component-zone no-default-spacer label="Kept" layout=${this.state.kept.length > 5 ? 'grid' : 'spread'} .stack=${stack(this.state.kept)} .componentView=${this.view}></boardgame-component-zone>
      </div>
      <button ?disabled=${this.busy || !selection.selected.length} @click=${this.keep}>Keep selected</button>
      <button ?disabled=${this.busy || !this.state.tray.length} @click=${() => this.reroll(false)}>Reroll tray</button>
      <button ?disabled=${this.busy || !this.state.tray.length} @click=${() => this.reroll(true)}>Repeat result</button>
      <button ?disabled=${this.busy || !this.state.kept.length} @click=${() => void this.install({ tray: [...this.state.tray, ...this.state.kept], kept: [] })}>Return kept</button>
      <button ?disabled=${this.busy} @click=${() => void this.install({ tray: Array.from({length: this.state.tray.length + this.state.kept.length === 5 ? 20 : 5}, (_, i) => makeDie(i)), kept: [] })}>Switch 5 / 20 dice</button>
      <p role="status">${this.busy ? 'Playing snapshot…' : `${selection.selected.length} selected`}</p>
      <boardgame-component-animator></boardgame-component-animator>
    `;
  }
}
customElements.define('dice-keep-demo', DiceKeepDemo);
