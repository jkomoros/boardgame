import { Element } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-icons/iron-icons.js';
import '@polymer/iron-icons/social-icons.js';
import '@polymer/paper-icon-button/paper-icon-button.js';
import './boardgame-ajax.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameConfigureGameProperties extends Element {
  static get template() {
    return html`
      <paper-icon-button icon="{{_openIcon(gameOpen)}}" alt="{{_openAlt(gameOpen)}}" disabled="{{disabled}}" on-tap="_handleOpenTapped"></paper-icon-button>
      <paper-icon-button icon="{{_visibleIcon(gameVisible)}}" alt="{{_visibleAlt(gameVisible)}}" disabled="{{disabled}}" on-tap="_handleVisibleTapped"></paper-icon-button>
      <boardgame-ajax id="ajax" game-path="configure" game-route="[[gameRoute]]" method="POST" content-type="application/x-www-form-urlencoded" last-response="{{_response}}"></boardgame-ajax>
`;
  }

  static get is() {
    return "boardgame-configure-game-properties"
  }

  static get properties() {
    return {
      gameVisible: {
        type: Boolean,
        value: false,
      },
      gameOpen: {
        type: Boolean,
        value: false,
      },
      admin: Boolean,
      isOwner: Boolean,
      gameRoute: Object,
      configurable: {
        type: Boolean,
        value: false,
      },
      disabled: {
        type: Boolean,
        computed: "_computeDisabled(admin, isOwner, configurable)"
      },
      _response: {
        type: Object,
        observer: "_responseChanged"
      }
    }
  }

  _computeDisabled(admin, isOwner, configurable) {
    return !(admin || isOwner || configurable);
  }

  _visibleIcon(gameVisible) {
    return gameVisible ? "visibility" : "visibility-off"
  }

  _openIcon(gameOpen) {
    return gameOpen ? "social:people" : "social:people-outline"
  }

  _openAlt(gameOpen) {
    return gameOpen ? "Anyone who has the link can join" : "Only specifically invited people may join"
  }

  _visibleAlt(gameVisible) {
    return gameVisible ? "Your game is publicly listed so random people can find it" : "Your game is unlisted so only people you share the link with can find it"
  }

  _handleOpenTapped() {
    this._submit(!this.gameOpen, this.gameVisible);
  }

  _handleVisibleTapped() {
    this._submit(this.gameOpen, !this.gameVisible);
  }

  _submit(open, visible) {
    this.$.ajax.body = {"open": open ? 1 : 0, "visible": visible ? 1 : 0 , "admin" : this.admin ? 1 : 0};
    this.$.ajax.generateRequest();
  }

  _responseChanged(newValue) {
    if (newValue.Status == "Success") {
      //Tell game-view to fetch data now
      this.dispatchEvent(new CustomEvent("refresh-info", {composed: true}));
    } else {
      this.dispatchEvent(new CustomEvent("show-error", {composed: true, detail: {"message" : newValue.Error, "friendlyMessage": newValue.FriendlyError, "title": "Couldn't toggle"}}));
    }
  }
}

customElements.define(BoardgameConfigureGameProperties.is, BoardgameConfigureGameProperties);
