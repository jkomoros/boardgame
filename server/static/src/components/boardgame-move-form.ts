/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { LitElement, html, css } from 'lit';
import { customElement, property } from 'lit/decorators.js';
import { repeat } from 'lit/directives/repeat.js';
import { when } from 'lit/directives/when.js';
import '@material/web/button/filled-button.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { submitMove } from '../actions/game.js';
import { selectGameError } from '../selectors.js';
import type { RootState } from '../types/store';
import type { JsonValue, MoveForm } from '../types/api';

interface GameRoute {
  name: string;
  id: string;
}

@customElement('boardgame-move-form')
export class BoardgameMoveForm extends connect(store)(LitElement) {
  static styles = css`
    :host {
      display: block;
    }

    #moves > details {
      margin-left: 1em;
    }

    h2 {
      margin-top: 0;
      margin-bottom: 0;
      font-family: var(--md-sys-typescale-title-large-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-title-large-size, 22px);
      font-weight: var(--md-sys-typescale-title-large-weight, 400);
      color: var(--md-sys-color-on-surface, #1C1810);
    }

    details {
      margin-bottom: 8px;
    }

    summary {
      cursor: pointer;
      padding: 8px 12px;
      border-radius: 8px;
      font-family: var(--md-sys-typescale-body-large-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-body-large-size, 16px);
      color: var(--md-sys-color-on-surface, #1C1810);
      background: var(--md-sys-color-surface-container, #F0EBE3);
    }

    summary:hover {
      background: var(--md-sys-color-surface-container-high, #E0D9CE);
    }

    form {
      padding: 8px 12px;
    }

    input[type="text"],
    input[type="number"],
    input:not([type]) {
      padding: 8px 12px;
      border: 1px solid var(--md-sys-color-outline, #857B6E);
      border-radius: 4px;
      font-family: var(--md-sys-typescale-body-medium-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-body-medium-size, 14px);
      color: var(--md-sys-color-on-surface, #1C1810);
      background: var(--md-sys-color-surface, #FAF6F0);
    }

    select {
      padding: 8px 12px;
      border: 1px solid var(--md-sys-color-outline, #857B6E);
      border-radius: 4px;
      font-family: var(--md-sys-typescale-body-medium-font, 'Source Sans 3', sans-serif);
      font-size: var(--md-sys-typescale-body-medium-size, 14px);
      color: var(--md-sys-color-on-surface, #1C1810);
      background: var(--md-sys-color-surface, #FAF6F0);
    }
  `;

  @property({ type: Array })
  config: MoveForm[] = [];

  @property({ type: Boolean })
  admin = false;

  @property({ type: Object, attribute: 'game-route' })
  gameRoute: GameRoute | null = null;

  @property({ type: Number, attribute: 'move-as-player' })
  moveAsPlayer = 0;

  @property({ type: Number, attribute: 'game-version' })
  gameVersion = 0;

  // animating mirrors boardgame-render-game's isAnimating, threaded down
  // through boardgame-admin-controls. While true, submit buttons are
  // disabled unless noAnimationDisable opts out (#721) — proposing a move
  // while the rendered state is still transitioning would be judged against
  // stale-looking state from the user's perspective.
  @property({ type: Boolean })
  animating = false;

  // noAnimationDisable opts a move-form instance out of the animating-driven
  // auto-disable above (e.g. for callers that want moves always available).
  @property({ type: Boolean, attribute: 'no-animation-disable' })
  noAnimationDisable = false;

  private _lastError: string | null = null;

  stateChanged(state: RootState): void {
    const error = selectGameError(state);
    // Show error if it changed and is new
    if (error && error !== this._lastError) {
      this._lastError = error;
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
        bubbles: true,
        detail: {
          message: error,
          friendlyMessage: error,
          title: "Couldn't make move"
        }
      }));
    } else if (!error) {
      this._lastError = null;
    }
  }

  private boolToInt(bool: boolean): string {
    return bool ? "1" : "0";
  }

  private _prepareValue(val: JsonValue): string {
    if (val === true || val === false) {
      return this.boolToInt(val);
    }
    return String(val);
  }

  private _isEnumField(fieldType: number): boolean {
    return fieldType === 5;
  }

  private _stringValues(obj: Record<string, string>): string[] {
    const result: string[] = [];
    const entries = Object.entries(obj);
    for (let i = 0; i < entries.length; i++) {
      const [, val] = entries[i];
      result.push(val);
    }
    return result;
  }

  proposeMove(moveName: string, args: Record<string, string | number>): void {
    if (!this.config) {
      console.warn("proposeMove called but no forms configed");
      return;
    }

    let moveConfig: MoveForm | undefined;
    for (let i = 0; i < this.config.length; i++) {
      const item = this.config[i];
      // TODO: fuzzy matching (remove whitespace and lowercase compare)
      if (item.Name === moveName) {
        moveConfig = item;
        break;
      }
    }

    if (!moveConfig) {
      console.warn("No move of name " + moveName + " found.");
      return;
    }

    const targetEleID = "#moves-" + this._normalizeID(moveConfig.Name);
    const containerEle = this.shadowRoot!.querySelector(targetEleID);

    if (!containerEle) {
      console.warn("Couldn't find move dom ele ", targetEleID);
      return;
    }

    const formEle = containerEle.querySelector("form") as HTMLFormElement;

    if (!formEle) {
      console.warn("Couldn't find form ele");
      return;
    }

    const inputs = formEle.elements;

    for (const key in args) {
      if (!Object.prototype.hasOwnProperty.call(args, key)) continue;

      let fieldFilled = false;

      for (let i = 0; i < inputs.length; i++) {
        const element = inputs[i] as HTMLInputElement | HTMLSelectElement;
        if (element.type === "hidden") continue;
        if (element.type === "submit") continue;

        if (element.getAttribute('name') === key) {
          // Set enum values differently
          if (element.type === "select-one") {
            (element as HTMLSelectElement).selectedIndex = args[key] as number;
          } else {
            element.value = String(args[key]);
          }
          fieldFilled = true;
        }
      }

      if (!fieldFilled) {
        console.warn("Couldn't find argument " + key + " in form.");
        return;
      }
    }

    this.submitForm(formEle);
  }

  private doSubmitForm(e: Event): void {
    const target = e.target as HTMLElement;
    const form = (target as HTMLInputElement).form;
    if (form) {
      this.submitForm(form);
    }
  }

  private submitForm(formEle: HTMLFormElement): void {
    if (!this.gameRoute) return;

    const body: Record<string, string> = {};
    const eles = formEle.elements;
    for (let i = 0; i < eles.length; i++) {
      const ele = eles[i] as HTMLInputElement;
      if (ele.name) {
        body[ele.name] = ele.value;
      }
    }

    // Dispatch action - errors will be handled via Redux state in stateChanged()
    store.dispatch(submitMove(this.gameRoute, body));
  }

  private _normalizeID(str: string): string {
    return str.split(" ").join("");
  }

  render() {
    return html`
      <h2>Moves</h2>
      <div id="container">
        ${repeat(this.config || [], (item) => item.Name, (item) => html`
          <details id="moves-${this._normalizeID(item.Name)}">
            <summary>Move ${item.Name}</summary>
            <form>
              <p><em>${item.HelpText}</em></p>
              <input type="hidden" name="MoveType" value="${item.Name}">
              <input type="hidden" name="admin" value="${this.boolToInt(this.admin)}">
              <input type="hidden" name="player" value="${this.moveAsPlayer}">
              <input type="hidden" name="ExpectedVersion" value="${this.gameVersion}">
              ${repeat(item.Fields || [], (field) => field.Name, (field) => html`
                <strong>${field.Name}</strong>
                ${when(
                  this._isEnumField(field.Type),
                  () => html`
                    <select name="${field.Name}">
                      ${repeat(
                        this._stringValues(field.Enum?.Values || {}),
                        (val) => val,
                        (val) => html`<option value="${val}">${val}</option>`
                      )}
                    </select>
                  `,
                  () => html`
                    <input name="${field.Name}" value="${this._prepareValue(field.DefaultValue)}">
                  `
                )}
                <br>
              `)}
              <div ?hidden="${item.Fields && item.Fields.length > 0}">
                <em>No modifiable fields</em><br>
              </div>
              <md-filled-button
                ?disabled="${this.animating && !this.noAnimationDisable}"
                @click="${this.doSubmitForm}">Make Move</md-filled-button>
            </form>
          </details>
        `)}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-move-form': BoardgameMoveForm;
  }
}
