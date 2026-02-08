import { LitElement } from 'lit';
import { customElement, property } from 'lit/decorators.js';

// defaultsInstances will be populated, on connected, by any boardgame-
// deck-defaults that has templates.
const defaultsInstances: BoardgameDeckDefaults[] = [];

// template by name is cached template elements to hand out.
const templatesByName: Record<string, HTMLTemplateElement> = {};

// BoardgameDeckDefaults is really just a container for templates with a deck
// property.
@customElement('boardgame-deck-defaults')
export class BoardgameDeckDefaults extends LitElement {
  // gameName can be set explicitly or will be set implicitly when
  // accessed via effectiveGameName.
  @property({ type: String })
  gameName = '';

  get effectiveGameName(): string {
    if (!this.gameName) {
      // If not set, search upwards to find our renderer whose name
      // implicitly contains it.
      let ele: HTMLElement | null = (this.parentNode as ShadowRoot)?.host as HTMLElement;
      while (ele) {
        if (ele.localName.startsWith('boardgame-render-game-')) {
          break;
        }
        // Look up the chain.
        if (ele.parentElement) {
          // Just normal parent
          ele = ele.parentElement;
        } else if ((ele.parentNode as ShadowRoot)?.host) {
          // Cross shadow DOM boundary
          ele = (ele.parentNode as ShadowRoot).host as HTMLElement;
        } else {
          // Unknown situation, just stop walking upward
          ele = null;
        }
      }

      if (ele && ele.localName.startsWith('boardgame-render-game-')) {
        this.gameName = ele.localName.replace('boardgame-render-game-', '');
      }
    }
    return this.gameName;
  }

  override connectedCallback() {
    super.connectedCallback();
    const template = this.querySelector('[deck]');
    if (!template) {
      // We must be just a reader defaults instance. Don't register.
      return;
    }
    defaultsInstances.push(this);
  }

  override disconnectedCallback() {
    super.disconnectedCallback();
    let i = 0;
    while (i < defaultsInstances.length) {
      const item = defaultsInstances[i];
      if (item === this) {
        defaultsInstances.splice(i, 1);
      } else {
        i++;
      }
    }
  }

  templateForDeck(gameName: string, deckName: string): HTMLTemplateElement | null {
    const templateKey = `${gameName}-${deckName}`;

    if (!deckName) {
      // This happens often when a new renderer is loaded and we don't yet
      // have the first state bundle.
      return null;
    }

    if (templatesByName[templateKey]) return templatesByName[templateKey];

    let template: HTMLTemplateElement | null = null;

    // Find the first defaults-instance that has it.
    for (const instance of defaultsInstances) {
      template = instance.querySelector(`[deck="${deckName}"]`) as HTMLTemplateElement;
      if (template) {
        // Verify that it's the right gameName.
        if (instance.effectiveGameName !== gameName) {
          // Sometimes effectiveGameName will be undefined before the deck
          // default is actually embedded in the parentElement (e.g. at
          // renderer boot) and that's OK.
          continue;
        }
        break;
      }
    }

    if (!template) return null;

    // In Lit, we don't need to templatize - we just store the template element
    // and it can be cloned when needed
    templatesByName[templateKey] = template;

    return template;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'boardgame-deck-defaults': BoardgameDeckDefaults;
  }
}
