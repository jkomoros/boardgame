/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { PolymerElement } from '@polymer/polymer/polymer-element.js';

import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/polymer/lib/elements/dom-if.js';
import './shared-styles.js';
import './boardgame-ajax.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
import { dom } from '@polymer/polymer/lib/legacy/polymer.dom.js';

class BoardgameMoveForm extends PolymerElement {
  static get template() {
    return html`
    <style include="shared-styles">
      #moves > details {
        margin-left:1em;
      }
      h2 {
        margin-top:0;
        margin-bottom:0;
      }
    </style>
      <h2>Moves</h2>
      <div id="container">
        <template is="dom-repeat" items="{{config}}">
          <details id="moves-{{_normalizeID(item.Name)}}">
            <summary>Move {{item.Name}}</summary>
            <form>
              <p><em>{{item.HelpText}}</em></p>
              <input type="hidden" name="MoveType" value="{{item.Name}}">
              <input type="hidden" name="admin" value="{{boolToInt(admin)}}">
              <input type="hidden" name="player" value="{{moveAsPlayer}}">
              <template is="dom-repeat" items="{{item.Fields}}">
                <strong>{{item.Name}}</strong>
                <template is="dom-if" if="[[_isEnumField(item.Type)]]">
                  <select name="{{item.Name}}">
                    <template is="dom-repeat" items="{{_stringValues(item.Enum.Values)}}">
                      <option value="{{item}}">{{item}}</option>
                    </template>
                  </select>
                </template>
                <template is="dom-if" if="[[!_isEnumField(item.Type)]]">
                  <input name="{{item.Name}}" value="{{_prepareValue(item.DefaultValue)}}">
                </template>
                <br>
              </template>
              <div hidden\$="{{item.Fields}}">
                <em>No modifiable fields</em><br>
              </div>
              <input type="button" on-tap="doSubmitForm" value="Make Move">
            </form>
          </details>
        </template>
        <boardgame-ajax id="ajax" game-path="move" game-route="[[gameRoute]]" method="POST" last-response="{{formResponse}}" content-type="application/x-www-form-urlencoded"></boardgame-ajax>
      </div>
`;
  }

  static get is() {
    return "boardgame-move-form"
  }

  static get properties() {
    return {
      config : Object,
      admin: Boolean,
      gameRoute: Object,
      moveAsPlayer: Number,
      formResponse: {
        type: Object,
        observer: "_formResponseChanged"
      }
    }
  }

  boolToInt(bool) {
    return bool ? "1" : "0"
  }

  _prepareValue(val) {
    if (val === true || val === false) {
      return this.boolToInt(val);
    }
    return val;
  }

  _isEnumField(fieldType) {
    return fieldType == 5;
  }

  _stringValues(obj) {
    let result = [];
    let entries = Object.entries(obj);
    for (let i = 0; i < entries.length; i++) {
      let [key, val] = entries[i];
      result.push(val);
    }
    return result;
  }

  proposeMove(moveName, args) {
    if (!this.config) {
      console.warn("proposeMove called but no forms configed")
      return;
    }
    let moveConfig;
    for (let i = 0; i < this.config.length; i++) {
      let item = this.config[i];
      //TODO: fuzzy matching (remove whitespace and lowercase compare)
      if (item.Name == moveName) {
        moveConfig = item;
        break;
      }
    }

    if (!moveConfig) {
      console.warn("No move of name " + moveName + " found.");
      return;
    }

    let targetEleID = "#moves-" + this._normalizeID(moveConfig.Name);

    let containerEle = this.shadowRoot.querySelector(targetEleID);

    if (!containerEle) {
      console.warn("Couldn't find move dom ele ", targetEleID);
      return;
    }

    let formEle = containerEle.querySelector("form");

    if (!formEle) {
      console.warn("Couldn't find form ele");
      return;
    }

    let inputs = formEle.elements;

    for (var key in args) {
      if (!args.hasOwnProperty(key)) continue;

      let fieldFilled = false;

      for (let i = 0; i < inputs.length; i++) {
        let element = inputs[i];
        if (element.type == "hidden") continue;
        if (element.type == "submit") continue;

        if (element.name == key) {
          element.value = args[key];
          fieldFilled = true;
        }

      }

      if (!fieldFilled) {
        console.warn("Couldn't find argument " + key + " in form.")
        return;
      }

    }
    this.submitForm(formEle);

  }

  doSubmitForm(e) {
    var evt = dom(e);
    console.log(e);
    this.submitForm(evt.localTarget.form);
  }

  submitForm(formEle) {
    var body = {};
    var eles = formEle.elements;
    for (var i = 0; i < eles.length; i++) {
      var ele = eles[i];
      body[ele.name] = ele.value;
    }
    this.$.ajax.body = body;
    this.$.ajax.generateRequest();
  }

  _normalizeID(str) {
    return str.split(" ").join("")
  }

  _formResponseChanged(newValue) {
    if (newValue.Error) {
      this.dispatchEvent(new CustomEvent("show-error", {composed: true, detail: {message: newValue.Error, friendlyMessage: newValue.FriendlyError, title: "Couldn't make move"}}));
    }
  }

}

customElements.define(BoardgameMoveForm.is, BoardgameMoveForm);
