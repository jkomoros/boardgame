import '@material/web/button/filled-button.js';
import '@material/web/select/filled-select.js';
import '@material/web/select/select-option.js';
import '@material/web/switch/switch.js';
import '@material/web/slider/slider.js';
import '@material/web/radio/radio.js';
import '@material/web/icon/icon.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';

import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { when } from 'lit/directives/when.js';

import {
  createGame,
  updateSelectedMangerIndex,
  updateNumPlayers,
  updateAgentName,
  updateVariantOption,
  updateOpen,
  updateVisible
} from '../actions/list.js';

import {
  selectManagers,
  selectSelectedManagerIndex,
  selectCreateGameNumPlayers,
  selectCreateGameAgents,
  selectCreateGameOpen,
  selectCreateGameVisible,
  selectCreateGameVariantOptions
} from '../selectors.js';

import type { MdFilledSelect } from '@material/web/select/filled-select.js';
import type { MdSlider } from '@material/web/slider/slider.js';
import type { MdSwitch } from '@material/web/switch/switch.js';
import type { MdRadio } from '@material/web/radio/radio.js';

interface AgentInfo {
  Name: string;
  DisplayName: string;
}

interface VariantValue {
  Value: string;
  DisplayName: string;
  Description: string;
}

interface VariantInfo {
  Name: string;
  DisplayName: string;
  Description: string;
  Values: VariantValue[];
}

interface ManagerInfo {
  Name: string;
  DisplayName: string;
  Description: string;
  DefaultNumPlayers: number;
  MinNumPlayers: number;
  MaxNumPlayers: number;
  Variant: VariantInfo[] | null;
  Agents: AgentInfo[];
}

// Empty manager for templates
const EMPTY_MANAGER: ManagerInfo = {
  Name: '',
  DisplayName: '',
  Description: '',
  DefaultNumPlayers: 0,
  MinNumPlayers: 0,
  MaxNumPlayers: 0,
  Variant: [],
  Agents: [],
};

@customElement('boardgame-create-game')
export class BoardgameCreateGame extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
    }

    .card {
      background: var(--md-sys-color-surface-container-low);
      padding: 20px;
      margin: 12px 0;
      border-radius: 12px;
      box-shadow: var(--md-sys-elevation-1);
    }

    md-switch {
      margin-right: 12px;
    }

    .secondary {
      font-family: var(--md-sys-typescale-body-small-font);
      font-size: var(--md-sys-typescale-body-small-size);
      line-height: var(--md-sys-typescale-body-small-line-height);
      color: var(--md-sys-color-on-surface-variant);
      margin-top: 4px;
    }

    .game .secondary {
      margin-bottom: -8px;
    }

    .variant > div {
      margin-right: 12px;
    }

    [hidden] {
      display: none !important;
    }

    .layout {
      display: flex;
      gap: 16px;
    }

    .vertical {
      flex-direction: column;
    }

    .horizontal {
      flex-direction: row;
    }

    .center {
      align-items: center;
    }

    .justified {
      justify-content: space-between;
    }

    .flex {
      flex-grow: 1;
    }

    md-filled-select {
      min-width: 250px;
    }

    md-slider {
      width: 250px;
      --md-slider-active-track-color: var(--md-sys-color-primary);
      --md-slider-handle-color: var(--md-sys-color-primary);
    }

    [role="radiogroup"] {
      display: flex;
      flex-direction: column;
      gap: 12px;
      margin: 12px 0;
    }

    .radio-label {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
    }

    .radio-label span {
      font-family: var(--md-sys-typescale-body-medium-font);
      font-size: var(--md-sys-typescale-body-medium-size);
      line-height: var(--md-sys-typescale-body-medium-line-height);
      color: var(--md-sys-color-on-surface);
    }

    .switch-label {
      display: flex;
      align-items: center;
      gap: 12px;
      cursor: pointer;
    }

    .switch-label span {
      font-family: var(--md-sys-typescale-body-medium-font);
      font-size: var(--md-sys-typescale-body-medium-size);
      line-height: var(--md-sys-typescale-body-medium-line-height);
      color: var(--md-sys-color-on-surface);
    }

    md-icon {
      margin-right: 8px;
      --md-icon-size: 20px;
    }

    md-filled-button {
      --md-filled-button-container-height: 40px;
    }

    /* Typography for labels and headings */
    label, .label-text {
      font-family: var(--md-sys-typescale-label-large-font);
      font-size: var(--md-sys-typescale-label-large-size);
      line-height: var(--md-sys-typescale-label-large-line-height);
      font-weight: var(--md-sys-typescale-label-large-weight);
      color: var(--md-sys-color-on-surface);
    }

    .player-count {
      font-family: var(--md-sys-typescale-body-large-font);
      font-size: var(--md-sys-typescale-body-large-size);
      line-height: var(--md-sys-typescale-body-large-line-height);
      color: var(--md-sys-color-on-surface);
      font-weight: 600;
    }
  `;

  @property({ type: Number })
  private _selectedManagerIndex = 0;

  @property({ type: Array })
  private _managers: ManagerInfo[] = [];

  @property({ type: Number })
  private _numPlayers = 0;

  @property({ type: Array })
  private _agents: string[] = [];

  @property({ type: Array })
  private _variantOptions: number[] = [];

  @property({ type: Boolean })
  private _open = false;

  @property({ type: Boolean })
  private _visible = false;

  constructor() {
    super();
    this._managers = [];
    this._numPlayers = 0;
  }

  stateChanged(state: any): void {
    this._managers = selectManagers(state);
    this._selectedManagerIndex = selectSelectedManagerIndex(state);
    this._numPlayers = selectCreateGameNumPlayers(state);
    this._agents = selectCreateGameAgents(state);
    this._variantOptions = selectCreateGameVariantOptions(state);
    this._open = selectCreateGameOpen(state);
    this._visible = selectCreateGameVisible(state);
  }

  private _handleSelectedManagerIndexChanged(e: Event): void {
    const select = e.target as MdFilledSelect;
    const selectedIndex = this._managers.findIndex(m => m.Name === select.value);
    store.dispatch(updateSelectedMangerIndex(selectedIndex));
  }

  private _handleOpenChanged(e: Event): void {
    const mdSwitch = e.target as MdSwitch;
    store.dispatch(updateOpen(mdSwitch.selected));
  }

  private _handleVisibleChanged(e: Event): void {
    const mdSwitch = e.target as MdSwitch;
    store.dispatch(updateVisible(mdSwitch.selected));
  }

  private get _selectedManager(): ManagerInfo {
    if (!this._managers || this._selectedManagerIndex < 0 || this._selectedManagerIndex >= this._managers.length) {
      return EMPTY_MANAGER;
    }
    return this._managers[this._selectedManagerIndex];
  }

  private get _variants(): VariantInfo[] {
    return this._selectedManager.Variant || [];
  }

  private _handleAgentSelectedChanged(e: Event): void {
    const radio = e.target as MdRadio;
    if (!radio.checked) return;
    const playerIndex = parseInt(radio.getAttribute('data-player-index') || '0', 10);
    store.dispatch(updateAgentName(playerIndex, radio.value));
  }

  private _handleVariantOptionChanged(e: Event): void {
    const select = e.target as MdFilledSelect;
    const variantIndex = parseInt(select.getAttribute('data-variant-index') || '0', 10);
    const variant = this._variants[variantIndex];
    const optionIndex = variant.Values.findIndex(v => v.Value === select.value);
    store.dispatch(updateVariantOption(variantIndex, optionIndex));
  }

  private _handleSliderValueChanged(e: Event): void {
    const slider = e.target as MdSlider;
    const value = slider.value;
    if (value !== undefined) {
      store.dispatch(updateNumPlayers(value));
    }
  }

  private get _managerHasAgents(): boolean {
    return this._selectedManager.Agents.length === 0;
  }

  private get _managerHasFixedPlayerCount(): boolean {
    return this._selectedManager.MinNumPlayers === this._selectedManager.MaxNumPlayers;
  }

  private serialize(): Record<string, string> {
    const body: Record<string, string> = {};

    // md-filled-select: direct value access (no selectedItem)
    const selects = [...this.shadowRoot!.querySelectorAll<MdFilledSelect>("md-filled-select")];
    for (const select of selects) {
      if (!select.name) continue;
      body[select.name] = select.value;
    }

    // md-slider
    const sliders = [...this.shadowRoot!.querySelectorAll<MdSlider>("md-slider")];
    for (const slider of sliders) {
      if (!slider.name) continue;
      body[slider.name] = String(slider.value);
    }

    // md-switch: .selected instead of .checked
    const switches = [...this.shadowRoot!.querySelectorAll<MdSwitch>("md-switch")];
    for (const mdSwitch of switches) {
      if (!mdSwitch.name) continue;
      body[mdSwitch.name] = mdSwitch.selected ? "1" : "0";
    }

    // md-radio: manual grouping (no md-radio-group)
    Object.assign(body, this._getRadioGroupValues());

    return body;
  }

  private _getRadioGroupValues(): Record<string, string> {
    const values: Record<string, string> = {};
    const radios = [...this.shadowRoot!.querySelectorAll<MdRadio>("md-radio")];
    for (const radio of radios) {
      if (!radio.name || !radio.checked) continue;
      values[radio.name] = radio.value;
    }
    return values;
  }

  private createGame(): void {
    store.dispatch(createGame(this.serialize()));
  }

  render() {
    return html`
      <div class="vertical layout">
        <div class="horizontal layout center game">
          <md-filled-select
            name="manager"
            label="Game Type"
            .value="${this._selectedManager.Name}"
            @change="${this._handleSelectedManagerIndexChanged}">
            ${this._managers.map((item) => html`
              <md-select-option value="${item.Name}">
                <div slot="headline">${item.DisplayName}</div>
                <div slot="supporting-text">${item.Description}</div>
              </md-select-option>
            `)}
          </md-filled-select>
          <div class="vertical layout">
            ${when(
              this._managerHasFixedPlayerCount,
              () => html`
                <div class="secondary">
                  <strong>${this._selectedManager.MinNumPlayers}</strong> players
                </div>
              `,
              () => html`
                <div class="secondary">Number of Players</div>
                <md-slider
                  name="numplayers"
                  labeled
                  min="${this._selectedManager.MinNumPlayers}"
                  max="${this._selectedManager.MaxNumPlayers}"
                  .value="${this._numPlayers}"
                  @change="${this._handleSliderValueChanged}">
                </md-slider>
              `
            )}
          </div>
          <div class="flex"></div>
          <md-filled-button @click="${this.createGame}" raised>
            Create Game
          </md-filled-button>
        </div>

        <div class="horizontal layout justified" ?hidden="${this._managerHasAgents}">
          ${repeat(this._agents, (_, index) => index, (item, index) => html`
            <div class="flex">
              <div class="vertical layout">
                Player ${index}
                <div role="radiogroup" aria-label="Agent for player ${index}" @change="${this._handleAgentSelectedChanged}">
                  <label class="radio-label">
                    <md-radio
                      name="agent-player-${index}"
                      value=""
                      data-player-index="${index}"
                      ?checked="${item === ''}">
                    </md-radio>
                    <span>Real Live Human</span>
                  </label>
                  ${repeat(this._selectedManager.Agents, (agent) => agent.Name, (agent) => html`
                    <label class="radio-label">
                      <md-radio
                        name="agent-player-${index}"
                        value="${agent.Name}"
                        data-player-index="${index}"
                        ?checked="${item === agent.Name}">
                      </md-radio>
                      <span>${agent.DisplayName}</span>
                    </label>
                  `)}
                </div>
              </div>
            </div>
          `)}
        </div>

        <div class="horizontal layout variant">
          ${repeat(this._variants, (variant) => variant.Name, (variant, index) => html`
            <div class="vertical layout">
              <md-filled-select
                label="${variant.DisplayName}"
                name="variant_${variant.Name}"
                data-variant-index="${index}"
                .value="${variant.Values[this._variantOptions[index]]?.Value || ''}"
                @change="${this._handleVariantOptionChanged}">
                ${variant.Values.map(value => html`
                  <md-select-option value="${value.Value}">
                    <div slot="headline">${value.DisplayName}</div>
                    <div slot="supporting-text">${value.Description}</div>
                  </md-select-option>
                `)}
              </md-filled-select>
              <div class="secondary">${variant.Description}</div>
            </div>
          `)}
        </div>

        <div class="horizontal layout">
          <label class="switch-label">
            <md-switch
              name="visible"
              ?selected="${this._visible}"
              @change="${this._handleVisibleChanged}">
              <md-icon slot="icon">visibility</md-icon>
            </md-switch>
            <span>Allow strangers to find the game</span>
          </label>
          <label class="switch-label">
            <md-switch
              name="open"
              ?selected="${this._open}"
              @change="${this._handleOpenChanged}">
              <md-icon slot="icon">people</md-icon>
            </md-switch>
            <span>Allow anyone who can view the game to join</span>
          </label>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-create-game': BoardgameCreateGame;
  }
}
