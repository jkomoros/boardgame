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
      //The names that have been called so far to _statedAnimating. Map of objects, to map of property names.
      _expectedTransitionEnds: Object,
      _outstandingTransitonEnds: Number,
    }
  }

  ready() {
    super.ready();
    this.resetAnimating();
    this.addEventListener("transitionend", e => this._transitionEnded(e));
    this.shadowRoot.addEventListener("transitionend", e => this._transitionEnded(e));
  }

  //willNotAnimate says whehter based on our current settings we expect this ele
  //and propName to fire a transitionend. Subclasses can override, but should
  //call this. Default answer is false, but this will return true if noAnimate
  //is true.
  willNotAnimate(ele, propName) {
    if (this.noAnimate) {
      return true
    }
    return false
  }

    //resetAnimating should be called when we expect animating count to be zero, 
  resetAnimating() {
    //if (this._animatingCount != 0) console.warn(this, this._animatingCount, "Was not zero when expected");
    this._expectedTransitionEnds = new Map();
    this._outstandingTransitonEnds = 0;
  }

  //_expectTransitionEnd is called whenever we have just changed a property
  //_that will later fire a transitionend, with the specific ele (this,
  //_#inner, #outer), and propertyName we expect to fire. We only care about
  //_transform and opacity changes; ignore everything else. We also will
  //_ignore things that this.willNotAnimate() tell us won't animate.
  _expectTransitionEnd(ele, propName) {
    if (!this._expectedTransitionEnds) {
      //This happens the first time state is installed. No biggie, just skip;
      //it.
      return;
    }

    if (propName != "transform" && propName != "opacity") return;
    if (this.willNotAnimate(ele, propName)) {
      //Sometimes we will have already told us to expect one, but later we
      //realize that we actually won't. This can happen for example the first
      //time a non-spacer card is set to a spacer--we update the inner
      //transform, then set spacer, then later update again. In those cases,
      //we should forget the one we previously told ourselves to expect.
      this._removeExpectedTransition(ele, propName);
      return;
    }

    let expectedPropsMap = this._expectedTransitionEnds.get(ele);
    if (!expectedPropsMap) {
      expectedPropsMap = new Map();
      this._expectedTransitionEnds.set(ele,expectedPropsMap);
    }
  
    //Already set!
    if (expectedPropsMap.get(propName)) return;

    expectedPropsMap.set(propName, true);
    this._outstandingTransitonEnds++;

    if (this._outstandingTransitonEnds == 1) {
      //This was the first one, fire a will-animate.
      this.dispatchEvent(new CustomEvent('will-animate', {composed: true, detail:{ele: this}}));
    }
  }

  //removes the ele and propName from the map, and returns whether it was in there.
  _removeExpectedTransition(ele, propName) {
    if (!this._expectedTransitionEnds) return false;
    let expectedPropsMap = this._expectedTransitionEnds.get(ele);
    if (!expectedPropsMap) return false;
    if (!expectedPropsMap.get(propName)) return false;
    expectedPropsMap.delete(propName);
    this._outstandingTransitonEnds--;
    if (this._outstandingTransitonEnds < 0) {
      console.warn("Got to less than 0 transition ends somehow", e);
      this._outstandingTransitonEnds = 0;
    }
    if (expectedPropsMap.size == 0) {
      this._expectedTransitionEnds.delete(ele);
    }
    return true
  }

  //_transitionEnded is the handler for transitionend. It will fire for _any_
  //_transition that ended on ourselves or our shadow root. We only care about
  //_transform and opacity changes; ignore everything else, because we'll
  //_heard about every property that changes, including box-shadow and others
  //_that are non-semantic.
  _transitionEnded(e) {

    if (e.propertyName != "transform" && e.propertyName != "opacity") return;
    if (!e.path || e.path.length < 1) return;

    let ele = e.path[0];

    let changeMade = this._removeExpectedTransition(ele, e.propertyName);

    if (changeMade && this._outstandingTransitonEnds == 0) {
      //all of the animations we were expecting to finish are finished.
      this.dispatchEvent(new CustomEvent('animation-done', {composed: true, detail:{ele:this}}));
    }
  }

}

customElements.define(BoardgameAnimatableItem.is, BoardgameAnimatableItem);
