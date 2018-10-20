import { PolymerElement } from '@polymer/polymer/polymer-element.js';

export class BoardgameAnimatableItem extends PolymerElement {

  static get is() {
    return "boardgame-animatable-item"
  }

  static get properties() {
    return {
      //If true, no animations will be played.
      noAnimate: {
        type: Boolean,
        value: false
      },
      //An inner counter keeping track of how many transitionend 's we expect
      //to receive before we fire `animation-done`.
      _animatingCount: {
        type: Number,
        value: 0,
      },
      //The names that have been called so far to _statedAnimating
      _animatingCalledNames: Object
    }
  }

    //resetAnimating should be called when we expect animating count to be zero, 
  resetAnimating() {
    //if (this._animatingCount != 0) console.warn(this, this._animatingCount, "Was not zero when expected");
    this._animatingCount = 0;
    this._animatingCalledNames = {};
  }

  //_startingAnimation is called whenever we have just changed a property that
  //_will later fire a transitionend (unless noAnimate is true). Will skip
  //_calls that have already been called with that name passed since the last
  //_resetAnimating called. This is because for example cards call
  //__updateInnerTransform multiple times, but only one transitionend fires.
  _startingAnimation(name) {
    //If called during no animation, ignore it.
    if (this.noAnimate) return;

    if (!this._animatingCalledNames) this._animatingCalledNames = {};
    if (this._animatingCalledNames[name]) return;
    this._animatingCalledNames[name] = true;

    //Keep track that we expect one more transitionend to be received.
    this._animatingCount++;
    if (this._animatingCount == 1) {
      //This was the first one, fire a will-animate.
      this.dispatchEvent(new CustomEvent('will-animate', {composed: true, detail:{ele: this}}));
    }
  }

  //_endingAnimation is the handler for transitionend.
  _endingAnimation(e) {
    if (this.noAnimate) {
      console.warn("_endingAnimation called when noAnimate was true", e, this);
      return;
    }
    this._animatingCount--;
    if (this._animatingCount < 0) this._animatingCount = 0;
    if (this._animatingCount == 0) {
      //all of the animations we were expecting to finish are finished.
      this.dispatchEvent(new CustomEvent('animation-done', {composed: true, detail:{ele:this}}));
    }
  }

}

customElements.define(BoardgameAnimatableItem.is, BoardgameAnimatableItem);
