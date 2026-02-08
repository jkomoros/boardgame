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
import './shared-styles.js';

import { connect } from 'pwa-helpers/connect-mixin.js';
import { store } from '../store.js';
import { submitMove } from '../actions/game.js';

interface MoveField {
  Name: string;
  Type: number;
  DefaultValue: string | number | boolean;
  Enum?: {
    Values: Record<string, string>;
  };
}

interface MoveConfig {
  Name: string;
  HelpText: string;
  Fields?: MoveField[];
}

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
    }
  `;

  @property({ type: Array })
  config: MoveConfig[] = [];

  @property({ type: Boolean })
  admin = false;

  @property({ type: Object })
  gameRoute: GameRoute | null = null;

  @property({ type: Number })
  moveAsPlayer = 0;

  private boolToInt(bool: boolean): string {
    return bool ? "1" : "0";
  }

  private _prepareValue(val: string | number | boolean): string {
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

    let moveConfig: MoveConfig | undefined;
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

  private async submitForm(formEle: HTMLFormElement): Promise<void> {
    if (!this.gameRoute) return;

    const body: Record<string, string> = {};
    const eles = formEle.elements;
    for (let i = 0; i < eles.length; i++) {
      const ele = eles[i] as HTMLInputElement;
      if (ele.name) {
        body[ele.name] = ele.value;
      }
    }

    const response = await store.dispatch(submitMove(this.gameRoute, body));

    if (response.error) {
      this.dispatchEvent(new CustomEvent("show-error", {
        composed: true,
        detail: {
          message: response.error,
          friendlyMessage: response.friendlyError,
          title: "Couldn't make move"
        }
      }));
    }
  }

  private _normalizeID(str: string): string {
    return str.split(" ").join("");
  }

  render() {
    return html`
      <h2>Moves</h2>
      <div id="container">
        ${repeat(this.config, (item) => item.Name, (item) => html`
          <details id="moves-${this._normalizeID(item.Name)}">
            <summary>Move ${item.Name}</summary>
            <form>
              <p><em>${item.HelpText}</em></p>
              <input type="hidden" name="MoveType" value="${item.Name}">
              <input type="hidden" name="admin" value="${this.boolToInt(this.admin)}">
              <input type="hidden" name="player" value="${this.moveAsPlayer}">
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
              <input type="button" @click="${this.doSubmitForm}" value="Make Move">
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
