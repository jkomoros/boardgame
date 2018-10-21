import { BoardgameAnimatableItem} from './boardgame-animatable-item.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

export class BoardgameComponent extends BoardgameAnimatableItem {
  static get template() {
    return html`
    <style>
      :root {
        /* These are copied and lightly modified from paper-styles/shadow, because we need rotated versions, too. */
        --shadow-elevation-normal: {
          box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                      0 1px 5px 0 rgba(0, 0, 0, 0.12),
                      0 3px 1px -2px rgba(0, 0, 0, 0.2);
        };

        --shadow-elevation-raised: {
          box-shadow: 0 8px 10px 1px rgba(0, 0, 0, 0.14),
                      0 3px 14px 2px rgba(0, 0, 0, 0.12),
                      0 5px 5px -3px rgba(0, 0, 0, 0.4);
        };

        --alt-shadow-elevation-normal: {
          filter: drop-shadow(0 2px 2px rgba(0, 0, 0, 0.14))
            drop-shadow(0 1px 5px rgba(0, 0, 0, 0.12))
            drop-shadow(0 3px 1px rgba(0, 0, 0, 0.2));     
        };

        --alt-shadow-elevation-raised: {
          filter: drop-shadow(0 8px 10px rgba(0, 0, 0, 0.14))
                  drop-shadow(0 3px 14px rgba(0, 0, 0, 0.12))
                  drop-shadow(0 5px 5px rgba(0, 0, 0, 0.4));
        };  


      }

      #outer {
        --default-component-scale: 1.0;
        --component-aspect-ratio: 1.0;
        --default-component-width: 30px;
        --component-effective-scale: var(--component-scale, var(--default-component-scale));
        --component-effective-width: calc(var(--component-effective-scale) * var(--component-width, var(--default-component-width)));
        --component-effective-height: calc(var(--component-effective-width) * var(--component-aspect-ratio));
      }

      .spacer {
        visibility: hidden;
      }

      #outer.interactive {
        cursor: pointer;
      }

      .no-animate #inner {
        transition: unset;
      }

      .disabled {
        filter: saturate(60%);
      }

      #outer {
        cursor: default;
      }

      #outer.shadow #inner {
        @apply --shadow-elevation-normal;
      }

      #outer.alt-shadow #inner {
        @apply --alt-shadow-elevation-normal;
      }

      #outer {
        transition: transform 0.1s ease-in-out;
      }

      #outer.interactive:hover {
        transform:translateY(-0.25em);
      }

      #outer.shadow.interactive:hover #inner {
        @apply --shadow-elevation-raised;
      }

      #outer.alt-shadow.interactive:hover #inner {
        @apply --alt-shadow-elevation-raised;
      }

      #inner {
        /* The second part of this transition is from paper-styles/shadow, because we need to do our own to get rotated */
        transition: transform var(--animation-length, 0.25s) ease-in-out, box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1), filter 0.28s cubic-bezier(0.4,0,0.2,1);
      }

    </style>

    <!-- subclass style should be inserted here -->

    <div id="outer" class\$="{{_classes}}" on-tap="handleTap" style\$="{{_outerStyle}}">
      <div id="inner">
        <!-- subclass content should be inserted here -->
      </div>
    </div>
`;
  }

  static get is() {
    return "boardgame-component"
  }

  static get properties() {
    return {
      //Index is the index of this card in the stack. Included in
      //component-tapped events.
      index: Number,
      //If item is set, it is assumed to be a Component from a stack in a
      //state. faceUp, coContent, spacer, and id will all be set based on
      //its value. A convenient way to stamp out cards with minimal fuss
      //in the default case.
      item: {
        type: Object,
        observer: "_itemChanged"
      },
      //id  should be set to the component's Id from the framework.
      //boardgame-component-animator will look for this id to figure out
      //which cards are logically the same and thus should be animated
      //from one place to another.
      id: String,
      //if true, the card will be rendered differently, no component-
      //tapped events will be triggered.
      disabled: {
        type: Boolean,
        value: false
      },
      //True if the card is neither disabled nor a spacer.
      interactive: {
        type: Boolean,
        readOnly: true,
        computed: "_computeInteractive(spacer, disabled)"
      },
      //If true, the card is "empty" and should take up space but not
      //render a card. Useful for sizedStacks that have an empty slot.
      spacer: {
        type: Boolean,
        reflectToAttribute: true,
        value: false
      },
      //if true, no drop-shadow will be rendered. boardgame-card- stack
      //will set this in some cases, like in a stack, so that all of the
      //shadows for 10's of cards don't multiply together.
      noShadow: Boolean,

      //If true, will use a different shadow rendering using drop-shadow
      //instead of box-shadow. This allows the alpha of the image to be
      //used, but with a lower fidelity rendering.
      altShadow: Boolean,

      //boardgameComponent reflects the attribute boardgame-component to
      //all subclasses, allowing htem to be selected via CSS.
      boardgameComponent: {
        type: Boolean,
        reflectToAttribute: true,
        value: true
      },

      //set _classes to change the classes that are set. call
      //_updateClasses() and it will call _computedClasses. When you
      //override _computedClasses, just add a new item in the observers
      //array that lists all the dependencies. Idelaly this would just be
      //a computed property thta you'd override how it was computed in
      //subclasses, but Polymer doesn't let you re-define a computed
      //property in a subclass.
      _classes: {
        type: String,
      },  

      //Bound into outer's style property. Extension point for subclasses.
      _outerStyle: {
        type: String,
        value: "",
      }
    }
  }

  static get observers() {
    //Your subclass should add an entry like _updateClasses(...allPropNames that could change), where all props includes noAnimate, spacer, noshadow,interactive,and disabled. We don't return one here because if we did it would be called twice--one for yours and one for this one.
    return []
  }

  //Returns a template for the subclass. The subtemplate's style will be
  //inserted right after the super template's style, and then the node
  //iwth id of import (if it exists) will have all of its children
  //imported in order and parented under #inner.
  static combinedTemplate(subTemplate) {
    let content = BoardgameComponent.template

    let result = document.importNode(content, true);

    let styleEle = subTemplate.content.querySelector("style");

    let injectionSite = result.content.children[1];

    result.content.insertBefore(styleEle,injectionSite);

    let innerEle = result.content.querySelector("#inner")


    let eleToImport = subTemplate.content.querySelector("#import")

    if (eleToImport) {
      while (eleToImport.children.length > 0) {
        innerEle.appendChild(eleToImport.children[0]);
      }
    }
    return result;
  }

  get willNotAnimate() {
    if (super.willNotAnimate) return true;
    //Spacer causes us to be visibility:hidden, which won't generate a
    //transitionend in chrome. See https://github.com/digitaledgeit/js-
    //transition-auto/issues/1
    if (this.spacer) return true;
    return false;
  }

  //obj.properties, smooshed down all the way to the upper.
  get _composedPropertyDefinition() {
    //TODO: can we get rid of this? Doesn't seem to be used, and I believe
    //Polymer does this for us now.
    if (!this._memoizedComposedPropertyDefinition) {
      let result = {};
      let obj = this;
      while (obj) {
        let props = obj.constructor.properties;
        if (!props) break;
        for (let key of Object.keys(props)) {
          result[key] = props[key];
        }
        obj = obj.__proto__;
      }
      this._memoizedComposedPropertyDefinition = result;
    }
    return this._memoizedComposedPropertyDefinition;
  }

  //animatingProperties should return an array of strings of property
  //names that change during animations. animatingPropValues() and
  //animatingPropDefaults() will use this.
  get animatingProperties() {
    return [];
  }

  //Returns the bundle of properties, as configured by
  //animatingProperties(), at their current value.
  animatingPropValues() {
    let result = {};
    for (let propName of this.animatingProperties) {
      result[propName] = this[propName];
    }
    return result;
  }

  //Returns the bundle of animating properties, as defined by
  //animatingProperties(), set to the defaults for the given stack. Used
  //when there isn't an element analog before or after the animation to
  //compare to.
  animatingPropDefaults(stack) {
    let result = {};
    for (let propName of this.animatingProperties) {
      result[propName] = stack.stackDefault(propName);
    }
    return result;
  }

  //computeAnimationProps is called by prepareAnimation and startAnimation,
  //passing the raw props and returning the actual properties to set. This is
  //the override point for sub-classes like boardgame-card who actually want
  //to set other properties, not the literal ones we were provided, for
  //performance reasons. The default simply returns props.
  computeAnimationProps(isAfter, props) {
    return props
  }

  //prepareAnimation is called after the new state is databound but just
  //before animation starts. Will call computeAnimationProps to get the final
  //props to set, which is an override point for subClasses. beforeProps is
  //what this element--or one like it--returned from animatingPropValues()
  //before the databinding happened. Transform is the transform to set on the
  //top-level element. This often isn't the literal transform from before, but
  //one that has been modified to be the previous transform, combined with the
  //inversion transform to move the component visually back to where it was.
  prepareAnimation(beforeProps, transform, opacity) {
    let props = this.computeAnimationProps(false, beforeProps);
    this.setProperties(props);
    this.style.transform = transform;
    this.style.opacty = opacity;
  }

  //startAnimation is called after the new state is databound and after
  //prepareAnimatino. Will call computeAnimationProps to get the final props
  //to set, which is an override point for subClasses.  afterProps is what
  //this element--or one like it--returned from animatingPropValues() after
  //the databinding happened. transform and opacity are the final values for
  //those two properties in their final location.
  startAnimation(afterProps, transform, opacity) {
    let props = this.computeAnimationProps(true, afterProps);
    this.setProperties(props);
    this.style.transform = transform;
    this.style.opacity = opacity;
    this._startingAnimation("default");
  }

  //prepareForBeingAnimatingComponent is called if the component is going
  //to be an animating component; that is it was created within
  //stack.newAnimatingComponent().
  prepareForBeingAnimatingComponent(stack) {
    //Do nothing; subclasses might do something.
  }

  //cloneContent returns whether we should clone the content of this
  //element during animating. Defaults to false; subclasses might
  //override.
  get cloneContent() {
    return false;
  }

  //animationRotates should return true if the before and after have a
  //different rotated property.
  animationRotates(beforeProps, afterProps) {
    return false;
  }

  _updateClasses() {
    //We can ignore the arguments, all that's necessary is that we're
    //called whenever the value could have changed.
    this._classes = this._computeClasses();
  }

  handleTap(e) {
    if (!this.interactive) {
      return;
    }
    this.dispatchEvent(new CustomEvent('component-tapped', {composed: true, detail: {index: this.index}}));
  }

  _itemChanged(newValue) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.spacer = true;
      return;
    }
    this.id = newValue.Id || "";
  }

  _computeInteractive(spacer, disabled) {
    return !spacer && !disabled;
  }

  _computeClasses() {
    let result = [];
    if (this.spacer) {
      result.push("spacer");
    }
    if (!this.noShadow) {
      result.push(this.altShadow ? "alt-shadow" : "shadow");
    }
    if (this.interactive) {
      result.push("interactive");
    }
    if (this.disabled) {
      result.push("disabled");
    }
    if (this.noAnimate) {
      result.push("no-animate")
    }
    return result.join(" ");
  }
}

customElements.define(BoardgameComponent.is, BoardgameComponent);
