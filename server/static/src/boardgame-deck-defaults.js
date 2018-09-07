import { Element } from '@polymer/polymer/polymer-element.js';
import { templatize } from '@polymer/polymer/lib/utils/templatize.js';

//defaultsInstances will be populated, on connected, by any boardgame-
//deck-defaults taht has tempaltes.
let defaultsInstances = [];
//tempalte by name is templatized cached things to hand out.
let templatesByName = {};

//BoardgameDeckDefaults is really just a container for templates with a deck
//property.
class BoardgameDeckDefaults extends Element {
  static get is() {
    return "boardgame-deck-defaults"
  }

  static get properties() {
    return {
      //gameName can be set explicitly or will be set implicitly when
      //accessed via effectiveGameName.
      gameName: {
        type: String
      }
    }
  }

  get effectiveGameName() {
    if (!this.gameName) {
      //If not set, search upwards to find our renderer whose name
      //implicitly contains it.
      let ele = this.parentNode.host;
      while (ele) {
        if (ele.localName.startsWith("boardgame-render-game-")) {
          break;
        }
        //Look up the chain.
        if (ele.parentElement) {
          //Just normal parent
          ele = ele.parentElement;
        } else if(ele.parentNode && ele.parentNode.host) {
          //Cross shadow DOM boundary
          ele = ele.parentNode.host;
        } else {
          //Unknown situation, just stop walking upward
          ele = null;
        }
      }

      if (ele && ele.localName.startsWith("boardgame-render-game-")) {
        this.gameName = ele.localName.replace("boardgame-render-game-", "");
      }

    }
    return this.gameName
  }

  connectedCallback() {
    let template = this.querySelector("[deck]");
    if (!template) {
      //We must be just a reader defaults instance. Don't register.
      return
    }
    defaultsInstances.push(this);
  }

  disconnectedCallback() {
    var i = 0;
    while (i < defaultsInstances.length) {
      var item = defaultsInstances[i];
      if (i == this) {
        defaultsInstances.splice(i, 1);
      } else {
        i++;
      }
    }
  }

  templateForDeck(gameName, deckName) {

    let templateKey = gameName + "-" + deckName

    if (!deckName) {
      //This happens often when a new renderer is loaded and we don't yet
      //have the first state bundle.
      return null;
    }

    if (templatesByName[templateKey]) return templatesByName[templateKey];

    let template;
    let templateParentInstance;

    //Find the first defaults-instance that has it.
    for (let instance of defaultsInstances) {
      template = instance.querySelector("[deck=" + deckName + "]");
      if (template) {

        //Verify that it's the right gameName.
        if (instance.effectiveGameName != gameName) {
          //Sometimes effectiveGameName will be undefined before the deck
          //default is actually embedded in the parentElement (e.g. at
          //renderer boot) and that's OK.
          continue;
        }

        templateParentInstance = instance;
        break;
      }
    }

    if (!template) return null;

    let templateInstance = templatize(template, templateParentInstance);

    templatesByName[templateKey] = templateInstance;

    return templateInstance;
  }
}

customElements.define(BoardgameDeckDefaults.is, BoardgameDeckDefaults);
