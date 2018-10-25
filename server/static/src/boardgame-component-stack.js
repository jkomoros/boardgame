import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import '@polymer/iron-flex-layout/iron-flex-layout.js';
import '@polymer/polymer/lib/elements/dom-repeat.js';
import '@polymer/polymer/lib/elements/dom-if.js';
import './boardgame-deck-defaults.js';
import './boardgame-card.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';
import { dashToCamelCase } from '@polymer/polymer/lib/utils/case-map.js';

var BoardgameComponentStack;

var sharedStackList = [];

//These are the random values we use. We need them to be the same for each key.
var pseudoRandomValues = [0.409629, 0.045224, 0.957352, 0.674598, 0.420276, 0.69529, 0.375744, 0.757951, 0.70161, 0.165333, 0.307487, 0.761283, 0.061829, 0.444936, 0.893498, 0.568637, 0.313256, 0.096547, 0.689491, 0.735802, 0.653278, 0.568639, 0.168512, 0.868507, 0.484359, 0.17899, 0.531255, 0.326798, 0.62229, 0.38919, 0.699555, 0.906995, 0.525003, 0.578083, 0.174277, 0.219422, 0.601429, 0.444303, 0.012084, 0.821015, 0.761839, 0.036714, 0.376602, 0.458024, 0.638299, 0.835771, 0.295093, 0.294265, 0.793988, 0.46952, 0.353375, 0.747226, 0.947925, 0.28802, 0.462328, 0.61194, 0.814922, 0.38534, 0.245267, 0.439828, 0.213518, 0.347944, 0.753266, 0.462731, 0.804775, 0.284984, 0.272913, 0.584866, 0.491474, 0.663749, 0.082535, 0.983746, 0.297612, 0.106188, 0.434022, 0.568589, 0.160467, 0.339668, 0.266839, 0.562368, 0.446072, 0.42395, 0.372456, 0.90581, 0.245894, 0.225044, 0.952531, 0.971619, 0.344237, 0.169056, 0.318305, 0.61021, 0.37241, 0.604048, 0.499905, 0.189648, 0.62082, 0.493149, 0.754958, 0.736003, 0.856383, 0.493789, 0.375336, 0.171083, 0.324018, 0.522103, 0.386678, 0.771986, 0.93826, 0.409071, 0.883499, 0.413002, 0.103175, 0.898144, 0.624233, 0.054771, 0.493724, 0.229437, 0.209876, 0.480461, 0.635064, 0.336034, 0.373814, 0.189853, 0.789123, 0.586921, 0.393384, 0.347548, 0.153098, 0.99294, 0.923887, 0.115157, 0.918841, 0.847155, 0.878961, 0.30781, 0.694391, 0.196018, 0.957279, 0.792493, 0.601996, 0.567903, 0.471252, 0.55565, 0.743476, 0.905305, 0.877299, 0.886972, 0.686297, 0.713567, 0.902545, 0.764396, 0.049177, 0.269324, 0.848572, 0.269079, 0.610327, 0.119123, 0.974389, 0.759697, 0.932941, 0.747793, 0.883499, 0.740326, 0.841617, 0.67744, 0.957788, 0.091737, 0.904918, 0.111062, 0.066767, 0.153055, 0.442142, 0.630529, 0.88257, 0.06523, 0.079217, 0.018493, 0.062141, 0.874116, 0.264976, 0.438222, 0.008716, 0.050499, 0.439269, 0.432986, 0.166319, 0.555334, 0.994858, 0.513525, 0.560583, 0.326941, 0.167995, 0.980903, 0.12907, 0.944065, 0.535459, 0.260209, 0.62893, 0.453737];

BoardgameComponentStack = class extends PolymerElement {
  static get template() {
    return html`
    <style>

      :host {
        width:100%;
      }

      #container {
        position: relative;
        box-sizing: border-box;
        @apply --layout-horizontal;
        @apply --layout-center;
      }

      #container #slot-holder {
        @apply --layout-horizontal;
        @apply --layout-center;
      }

      #animating-components [boardgame-component] {
        /* don't mess with the general layout because the card will be gone soon */
        position: absolute;
      }

      #container ::slotted([boardgame-component]), #container [boardgame-component] {
        transition: transform var(--animation-length, 0.25s) ease-in-out, opacity var(--animation-length, 0.25s) ease-in-out;
      }

      #container.no-animate ::slotted([boardgame-component]), #container.no-animate [boardgame-component] {
        transition: unset;
      }

      #container.grid #slot-holder, #container.stack #slot-holder {
         @apply --layout-wrap;
      }

      #container ::slotted([boardgame-component]), #container [boardgame-component] {
        margin: 1em;
      }

      #container.pile {
        padding: calc(calc(var(--pile-scale, 1.0) * 3em) + 3em);
        /* TODO: it might be that this should always be true for container */
        @apply --layout-center-justified;
      }

      #container.pile ::slotted([boardgame-component]), #container.pile [boardgame-component] {
        position:absolute;
        margin:0;
      }

     .pile ::slotted([boardgame-component].bcc-first), .pile [boardgame-component].bcc-first {
        position:relative;
      }

      /* most cards in stacks should be at the front, and a few pixels down so
      /* you can see them behind the first few cards */

      .stack ::slotted([boardgame-component]), .stack [boardgame-component] {
        position: absolute;
        top: 6px;
        left: 0px;
      }

      /* the first child should actually take up layout space for the entire
      /* stack, and not be offset any. */

      .stack ::slotted([boardgame-component].bcc-first), .stack [boardgame-component].bcc-first, .stack [boardgame-component]#spacer {
        z-index: 10;
        position: relative;
        top:0px;
      }

      /* The second card should stick out justa bit */

      .stack ::slotted([boardgame-component]:nth-child(2)), .stack [boardgame-component]:nth-child(2) {
        top:1px;
        z-index: 9;
      }

      /* The third card should stick out just a bit more */

      .stack ::slotted([boardgame-component]:nth-child(3)), .stack [boardgame-component]:nth-child(3) {
        z-index: 8;
        top:2px;
      }

      .stack ::slotted([boardgame-component]:nth-child(4)), .stack [boardgame-component]:nth-child(4) {
        z-index: 7;
        top:3px;
      }

      .stack ::slotted([boardgame-component]:nth-child(5)), .stack [boardgame-component]:nth-child(5) {
        z-index: 6;
        top:4px;
      }

      .stack ::slotted([boardgame-component]:nth-child(6)), .stack [boardgame-component]:nth-child(6) {
        z-index: 5;
        top:5px;
      }

      #container.spread #slot-holder {
        
        /* If we just applied layout-around-justified to #container.spread, it
        /* wouldn't work because the children are actually the four containers */
        @apply --layout-horizontal;
        @apply --layout-around-justified;
        width:100%;
      }

      #container.spread #slot-holder ::slotted(dom-repeat) {
        /* If we don't do this then oddly enough the dom-repeats will take up space in the layout! */
        display: none;
      }

      .spread ::slotted([boardgame-component]:hover), .spread [boardgame-component]:hover, .fan ::slotted([boardgame-component]:hover), .fan [boardgame-component]:hover {
        z-index:1;
      }

      #container.spread ::slotted([boardgame-component]), #container.spread [boardgame-component] {
        /* default card width */
        margin-right: calc(100px * -0.75);
      }

      #container.spread ::slotted([boardgame-component].bcc-last), #container.spread [boardgame-component].bcc-last {
        margin-right: 0;
      }

      #container.fan ::slotted([boardgame-component]), #container.fan [boardgame-component] {
        /* default card width */
        margin-right: calc(100px * -0.5);
      }

      #container.fan ::slotted([boardgame-component][rotated]), #container.fan [boardgame-component][rotated] {
        /* default card width */
        margin-right: calc(100px * -0.25);
      }

      #container.fan ::slotted([boardgame-component].bcc-last), #container.fan [boardgame-component].bcc-last {
        margin-right: 0;
      }

      #container.fan {
        /* give a bit more vertical breathing room for the arc */
        box-sizing:border-box;
        padding: 1em 0;
      }

    </style>

    <div id="container" class\$="{{_classes(layout, noAnimate)}}" style\$="{{_style}}">
      <div id="slot-holder">
        <slot id="components"></slot>
      </div>
      <!-- spacer will be created here if necessary --> 
      <div id="faux-components">
        <!--Faux components will go here -->
      </div>
      <div id="animating-components"></div>
    </div>
`;
  }

  static get is() {
    return "boardgame-component-stack"
  }

  static get properties() {
    return {
      //layout sets the layout style to use. One of `stack` (default),
      //`grid`, `spread`, `fan` and `pile`
      layout: {
        type: String,
        observer: "_updateComponentClasses",
        value: "stack"
      },
      //stack should be set to the state value corresponding to a stack.
      //It will set idsLastSeen, and also set the first dom-repeat it
      //finds as a child's items to stack.Components.
      stack : {
        type: Object,
        value: null,
        observer: "_stackChanged"
      },
      //deckName, if set, is the deck that this stack is associated
      //with. Will be auto-set based on the value of stack.
      deckName: {
        type: String,
        value: "",
      },
      //gameName is the name of the gametype we're associated with, for
      //looking up which template to instantiate for our children. Will
      //be auto-set based on stack.
      gameName: {
        type: String,
        value: "",
        observer: "_gameNameChanged"
      },
      //if true, all cards will have a slight rotation transformation
      //applied to make the layout not look artificially tidy
      messy: {
        type: Boolean,
        value: false,
        observer: "_updateComponentClasses"
      },
      //If the cards are messy, how messy are they? 1.0 is default
      //messiness, 2.0 is twice as messy.
      messiness: {
        type: Number,
        value: 1.0,
        observer: "_updateComponentClasses"
      },
      //idsLastSeen should be the IdsLastSeen for the stack that this
      //stack represents. boardgame-component-animator will use this
      //value to figure out where to animate from and to.
      idsLastSeen: Object,
      //if true, animations will be suppressed
      noAnimate: Boolean,
      //If there are no cards, by default we'll inject a single spacer
      //(invisibled) card, just to make sure the overall layout doesn't
      //jump when a card is added to this stack. If this property is
      //true, we won't do that.
      noDefaultSpacer: Boolean,
      //If there aren't more than fauxCards "real" cards in the stack,
      //we'll make fake cards until there are this many cards. This is
      //useful if you have, for example, a very big draw deck and only
      //want to actually render the first few cards.
      fauxComponents: {
        type: Number,
        value: 0
      },

      _templateClass: {
        type: Object,
      },

      _pileScaleFactor: {
        type: Number,
        value: 1.0,
      },

      _style: {
        type: String,
        computed: "_computeStyle(_pileScaleFactor)",
      },

      _randomRotationOffset: {
        type: Number,
        value: 0
      },
    }
  }

  get _sharedStackList() {
    return sharedStackList;
  }

  _computeStyle(pileScaleFactor) {
    return "--pile-scale:" + pileScaleFactor;
  }

  get deckDefaults() {
    let ele = this.shadowRoot.querySelector("boardgame-deck-defaults");
    if (ele) return ele;

    ele = document.createElement("boardgame-deck-defaults");
    this.shadowRoot.appendChild(ele);
    return ele;

  }

  _gameNameChanged() {
    //If gameName was changed and we're still around, the components in
    //our pool are no longer valid.
    this._componentPool = [];
  }

  //Components that are returned can be passed out via newComponent
  //later.
  returnComponent(ele) {
    this._componentPool.push(ele);
  }

  //Returns a new component for this stack set reasonably, based on
  //componentType and autoSetType.
  newComponent() {

    if (this._componentPool.length > 0) {
      return this._componentPool.pop();
    }

    let templateClass = this.templateClass;

    if (templateClass) {
        let instance = new templateClass({});
        for (let child of instance.children) {
          if (child.nodeType != 1) continue;
          child.instance = instance;
          return child;
        }
        console.warn("None of the nodes printed by the template are an actual node.");
        return null;
    }

    console.warn("No template class to auto stamp")

    return null;
  }

  //stackDefault goes through the entire stack, and returns the most
  //popular value for propName, skipping elements that don't have it.
  stackDefault(propName) {
    //Just default to setting whatever the most common type is in the
    //stack.
    let propCount = {};
    let maxCount = 0;
    let maxVal = false;
    for (let c of this.Components) {
      let val = c[propName]
      if (val === undefined) {
        continue;
      }
      propCount[val] = (propCount[val] || 0) + 1
      if (propCount[val] > maxCount) {
        maxCount = propCount[val];
        maxVal = val;
      }
    }
    return maxVal;
  }

  connectedCallback() {
    sharedStackList.push(this);
  }

  disconnectedCallback() {
    var i = 0;
    while (i < sharedStackList.length) {
      var item = sharedStackList[i];
      if (i == this) {
        sharedStackList.splice(i, 1);
      } else {
        i++;
      }
    }
  }

  //offsetCard returns a card in the stack to use for positioning.
  get offsetComponent() {
    var components = this._realComponents;
    if (components.length > 0) return components[0];
    var component = this.shadowRoot.querySelector("[boardgame-component]");
    if (component) return component;
    return this;
  }

  //id will be set to a long, stable Id for this stack during its
  //lifetime. Primarily useful to identify itself to the animation-
  //coordinator.
  get id() {
    return this._id;
  }

  ready() {
    this._componentPool = [];
    super.ready();
    this.$.components.addEventListener("slotchange", () => this._slotChanged(false));
    //If the children are already stamped, update now.
    this._slotChanged(true);
    //Generate a random rotation offset from 0 to 20 that will make this
    //stack have rotatons that are stable as long as the stack exists
    //(which it will in general because Polymer databinding will reuse
    //elements) but different from most other stacks.
    this._randomRotationOffset = Math.floor(Math.random() * 21);
    this._id = this._randomId(8);
    let attrs = this._attributesForComponents();
    for (let key of attrs.keys()) {
      let titleKey = key[0].toUpperCase() + key.slice(1,key.length);
      this._createPropertyObserver("component" + titleKey, "_componentPropChanged")
    }
  }

  _componentPropChanged() {
    //If a prop that we forward to our children changed, just regenerate
    //children.
    //TODO: only change the attribute that was changed.
    this._generateChildren();
    this._slotChanged(false);
  }

  //setUnknownAnimationState sets the "final" state of the styling on a
  //card that is flying to the unknown state. Used by boardgame-
  //component-animator.
  setUnknownAnimationState(card) {
    card.style.transform = "scale(0.6)";
    card.style.opacity = "0.0";
  }

  //Returns a new card to animate its position. When its transition ends it will be removed.
  newAnimatingComponent() {
    var component = this.newComponent();
    component.noAnimate = true;
    component.prepareForBeingAnimatingComponent(this);
    this.setUnknownAnimationState(component);
    this.$["animating-components"].appendChild(component);
    component.addEventListener("transitionend", e => this._clearAnimatingComponents(e));
    return component;
  }

  _clearAnimatingComponents(e) {
    var container = this.$['animating-components'];
    while(container.children.length > 0) {
      var child = container.children[0];
      if (child.beforeOrphaned) child.beforeOrphaned();
      //TODO: we should have an _animatingComponentPool too.
      container.removeChild(child);
    }
  }

  _stackChanged(newValue) {
    if (newValue) {
      if (newValue.Deck) {
        this.deckName = newValue.Deck;
      }
      if (newValue.GameName) {
        this.gameName = newValue.GameName;
      }
      this.idsLastSeen = newValue.IdsLastSeen || {};
    } else {
      this.deckName = "";
      this.gameName = "";
      this.idsLastSeen = null;
    }
    var repeater = this.querySelector("dom-repeat");
    if (repeater) {
      repeater.items = newValue ? newValue.Components : [];
      return;
    }

    //We didn't have a repeater to stamp into. We'll have to generate
    //the components ourselves!
    this._generateChildren();
  }

  get templateClass() {
    //We don't memoize this because if the renderer changes then the
    //templateClass we use should change.
    return this.deckDefaults.templateForDeck(this.gameName, this.deckName);
  }

  _attributesForComponents() {

    //Want to forward any attribute set on us that starts
    //with "component-". One complication is that attributes that are
    //databound to us without using "$=" will not show up as attributes,
    //only properties.


    let result = new Map();

    for (let name of Object.getOwnPropertyNames(this)) {
      if (!name.startsWith("component")) {
        continue;
      }
      let finalName = name.replace("component", "");
      finalName = finalName[0].toLowerCase() + finalName.slice(1, finalName.length);
      result.set(finalName, this[name]);
    }

    for (let attr of this.attributes) {
      if (!attr.name.startsWith("component-")) {
        continue;
      }
      let name = attr.name.replace("component-", "");
      let finalName = dashToCamelCase(name);
      result.set(finalName, attr.value);
    }

    return result;

  }

  get _realComponents() {
    //flatten:true will return the fallback nodes, which is where inject
    //the _generateChildren nodes.
    var components = this.$.components.assignedNodes({flatten: true});

    var result = [];

    for (var i = 0; i < components.length; i++) {
      var component = components[i];

      if (!component.localName || component.getAttribute("boardgame-component") !== "") {
        //Skip text nodes and nodes that aren't cards
        continue;
      }

      result.push(component);
    }

    return result;
  }

  get _fauxComponents() {
    var fauxComponents = this.$['faux-components'].querySelectorAll("[boardgame-component]");

    return [...fauxComponents];
  }

  get Components() {
    return this._realComponents.concat(this._fauxComponents);
  }

  _insertNodes(componentsInfo, hostEle) {

    //We can't just do a hostEle.children.length, because some of them
    //might be e.g. fading-text.
    let componentCount = hostEle.querySelectorAll("[boardgame-component]").length;

    let childrenToAdd =  componentsInfo.length - componentCount;
  
    if (childrenToAdd > 0) {
      //We want to put the cards right in front of the first non
      //boardgame-component, if it exists.
      let firstNonComponentEle = null;
      for (let i = 0; i < hostEle.children.length; i++) {
        let ele = hostEle.children[i];
        if (!ele.hasAttribute("boardgame-component")) {
          firstNonComponentEle = ele;
          break;
        }
      }

      for (let i = 0; i < childrenToAdd; i++) {
        //if firstNonComponentEle is null, insertBefore is just
        //equivalent to appendChild.
        let ele = this.newComponent();
        //During boot the ele might not be defined yet.
        if (!ele) break;
        hostEle.insertBefore(ele, firstNonComponentEle);
      }

    } else if (childrenToAdd < 0) {
      //We have childrent to remove Offset will be how many items from
      //the back that we go to remove. Every time that an element we're
      //looking at is NOT a boardgame-component, we increase by one. We
      //shouldn't have to worry about falling off the front because the
      //componentCount calculation above should be aligned.
      let offset = 0;
      for (let i = 0; i > childrenToAdd; i--) {

        let eleToRemove;
        let firstLoop = true;

        do {
          if (firstLoop) {
            firstLoop = false;
          } else {
            offset++;
          }
          eleToRemove = hostEle.children[hostEle.children.length - offset - 1];
        } while(!eleToRemove.hasAttribute("boardgame-component"))

        //We used to remove this from hostEle.children[0], but that
        //triggered a weird bug described in #476 comment where Chrome
        //would get confused about the elements still being there and
        //getting in weird assigneNode limbo. Removing it from the back
        //doesn't appear to trigger the issue.
        this._componentPool.push(hostEle.removeChild(eleToRemove));
      }
    }

    let attrs = this._attributesForComponents();

    const indexAttrs = "indexAttributes";

    let attributesToIndex = attrs.get(indexAttrs) ? attrs.get(indexAttrs).split(",") : [];

    //i will not be the component index, because some children might not
    //correspond to components.
    let componentIndex = 0;

    for (let i = 0; i < hostEle.children.length; i++) {
      let ele = hostEle.children[i];

      if (!ele.hasAttribute("boardgame-component")) continue;

      ele.item = componentsInfo[componentIndex];
      ele.index = componentIndex;

      //In case they used template databinding on the instance we need
      //to update its binding too.
      if (ele.instance) {
        ele.instance.item = componentsInfo[componentIndex];
        ele.instance.index = componentIndex;
      }

      //TODO: we really only need to set these once, when the element is
      //initially crated. However, the ele is a document-fragment, and
      //even after appendingChild we don't get a clean ref to it. So
      //shrug, just do extra work here.
      for (let key of attrs.keys()) {
        if (key == indexAttrs) continue;
        let valToSet = attrs.get(key);
        let propConfigs = ele._composedPropertyDefinition;
        if (propConfigs[key] && propConfigs[key].type == Boolean && valToSet === "") {
          valToSet = true;
        }
        ele[key] = valToSet;
      }
      for (let name of attributesToIndex) {
        let finalName = dashToCamelCase(name);
        ele[finalName] = componentIndex;
        ele.setAttribute(name, componentIndex);
      }

      componentIndex++
    }
  }

  _generateChildren() {

    //When we stamped the components into the shadow root's slot, we
    //used to have to clear out anything that was here, because
    //otherwise they'd prevent the "default" contents of the slot from
    //showing. But now that we support other elements in the stack, we
    //no longer have to remove them.

    //We used to stamp 'defaults' into the this.$.components. However,
    //if we do that then it's not possible to target styles in the game
    //renderer at the template because the objects are in the shadow dom
    //and thus can't be reached by styles. So just stamp into the actual
    //children.
    this._insertNodes(this.stack ? this.stack.Components : [], this);

  }

  _slotChanged(firstRender) {

    var realComponents = this._realComponents;

    let fauxComponentsContainer = this.$["faux-components"];

    let wantSpacer = realComponents.length < 1 && !this.noDefaultSpacer;

    let haveSpacer = !!this.shadowRoot.querySelector("#container>[boardgame-component][spacer]")

    //Add a spacer if we need one.
    if (wantSpacer && !haveSpacer) {
      //Need to add one.
      let spacer = this.newComponent();
      //During boot (or switch) sometimes we don't have spacers
      if (spacer) {
        spacer.spacer = true;
        spacer.id = "spacer";
        this.$.container.insertBefore(spacer, fauxComponentsContainer);
      }
    }

    //Trim down any extra spacers. It's possible for too many to be
    //added in weird race conditions.
    let targetNumSpacers = wantSpacer ? 1 : 0;

    let spacers = this.shadowRoot.querySelectorAll("#container>[boardgame-component][spacer]")

    while (spacers.length > targetNumSpacers) {
      this.$.container.removeChild(spacers[0]);
      spacers = this.shadowRoot.querySelectorAll("#container>[boardgame-component][spacer]")
    }

    if (firstRender && realComponents.length < 1 ) return;

    if (realComponents.length < this.fauxComponents) {

      let targetNumFauxComponents = this.fauxComponents - realComponents.length;

      let info = [];

      for (let i = 0; i < targetNumFauxComponents; i++) {
        //boardgame-component, if its item is set to null, will treat it
        //as a spacer. But undefined it will just ignore.
        info.push(undefined);
      }

      var that = this;

      this._insertNodes(info, fauxComponentsContainer);

    }

    this._updateComponentClasses();

  }

  _updateComponentClasses() {

    //Called when layout changes, or when items are added or removed from
    //the slot.

    //TODO: the fact that we do this imperatively (and trample on whatever
    //shadow setting was explicilty set by the element to start) seems
    //like a smell. However, doing this just with CSS was quite hard.

    var components = this.Components;

    let lastPileScaleFactor = 0.0;

    for (var i = 0; i < components.length; i++) {
      var component = components[i];

      var classes = ["bcc-first", "bcc-last"];

      for (var j = 0; j < classes.length; j++) {
        component.classList.remove(classes[j]);
      }

      if (i == 0) {
        component.classList.add("bcc-first");
      }

      let transformPieces = [];

      let id = component.id || i.toString();

      if (this.messy && this.layout != "pile") {
        transformPieces.push("rotate(" + this._messyRotationForId(id) + "deg)");
      }

      if (this.layout == "pile") {
        let id = component.id || i.toString();
        let offsets = this._pileOffsetsForId(id, components.length);
        transformPieces.push("translate(" + offsets.x + "px, " + offsets.y + "px)");
        lastPileScaleFactor = offsets.scaleFactor;
        transformPieces.push("rotate(" + this._messyRotationForId(id) + "deg)");
      }

      //We always set transform even if it's "" because _fanCards
      //assumes that transform has been reset in this layout pass.
      component.style.transform = transformPieces.join(" ");

      if (i == components.length - 1) {
        component.classList.add("bcc-last");
      }

      if (this.layout != "stack") {
        component.noShadow = false;
        continue
      }

      if (i < 4) {
        component.noShadow = false;
      } else {
        component.noShadow = true;
      }
    }

    if (this.layout == "pile") {
      this._pileScaleFactor = lastPileScaleFactor;
    }

    if (this.layout == "fan") {
      this._fanComponents();
    }

  }

  _fanComponents() {
    var components = this.Components;

    //TODO: set the amount of max rotation based on length of stack;
    var maxRotation = 20;
    var minRotation = maxRotation * -1;
    
    var maxTranslate = -1.0;
    var minTranslate = 1.5;

    //stackDefault will just return false if the items in the stack
    //don't have a rotated property.
    let rotated = this.stackDefault("rotated");
    
    if (components.length < 8 && rotated || !rotated && components.length < 3) {
      var percent = 0.5
      maxRotation *= percent;
      minRotation *= percent;
      maxTranslate *= percent;
      minTranslate *= percent;
    }

    var rotationSpread = maxRotation - minRotation;
    var translateSpread = maxTranslate - minTranslate;


    for (var i = 0; i < components.length; i++) {
        var component = components[i];

        var percent = (i / (components.length - 1));

        var rotation = percent * rotationSpread + minRotation;

        //TODO: preserve the messiness transformations
        var rotationTransformation = "rotate(" + rotation + "deg)";

        var translateRadians = 3.0 * percent - 1.5;

        if (percent < 0) {
          translateRadians = translateRadians * -1;
        }

        var translate = Math.cos(translateRadians) * translateSpread + minTranslate;

        var translateTransformation = "translateY(" + translate + "em)";

        component.style.transform += rotationTransformation + " " + translateTransformation;

    }

  }

  _messyRotationForId(id) {
    let index = Math.abs(this._hashCode(id));
    index %= pseudoRandomValues.length;
    return (pseudoRandomValues[index] * 8 - 4) * this.messiness;
  }

  _pileOffsetsForId(id, numComponents) {
    let x = this._randomOffsetForId(id, true);
    let y = this._randomOffsetForId(id, false);


    //x and y are spread evenly within a rectangle, but we want a stack
    //that gets narrower at the top.

    //The solution is to cut off two triangles from the top half, and
    //slide them down and out to the edges below.

    //As a diagram:

    /*
    -------------------------
    |  /|               |\  |
    | / |               | \ |
    |/  |               |  \|
    -------------------------
    |                       |
    |                       |
    |                       |
    -------------------------

    to:

           -------------------
          /|                 |\ 
         / |                 | \
        /  |                 |  \
       ---------------------------
      /|                         |\
     / |                         | \
    /  |                         |  \
    ----------------------------------
    */

    //The range for x is -1.0 to 1.0. This is how much from the left and
    //right to cut in. 1.0 would be a complete triangle, 0.0 would be an
    //untouched rectangle.
    const triangleWidth = 0.5;

    //Only process the top half
    if (y > 0) {
      //Flip to all positive to make the logic easier for now
      let negative = x < 0;
      x = Math.abs(x);

      if (x > (1 - triangleWidth)) {
        //It's in the triangle. Is it in the outer half?
        let rectX = x - (1.0 - triangleWidth);

        if (rectX < y) {
          //Yup!

          //Flip upside down
          y *= -1;

          //Move outward 
          x = 1 + rectX;

        }

      }

      if (negative) x *= -1;
    }

    //Scale the size of the pile based on how many components there are.

    let smallestSize = 20;
    let largestSize = 50;

    let lowestExpectedComponents = 5;
    let highestExpectedComponents = 25;

    let expectedComponentRange = highestExpectedComponents - lowestExpectedComponents;

    let clampedComponents = Math.min(Math.max(numComponents, lowestExpectedComponents), highestExpectedComponents);

    let multiplier = (clampedComponents - lowestExpectedComponents)  / expectedComponentRange;

    let finalSize = smallestSize + (largestSize - smallestSize) * multiplier;

    x *= finalSize;
    y *= finalSize;

    //Higher x's are actually downward because we'll be adding to our
    //position, so flip the calculation.
    x *= -1;
    y *= -1;

    return {x: x, y: y, scaleFactor: multiplier};
  }

  //Returns between -1 and 1
  _randomOffsetForId(id, x) {

    if (x) id = id + "right";

    //The same component in different stacks should get a different
    //location.
    id = id + this._id;

    let index = Math.abs(this._hashCode(id));

    index %= pseudoRandomValues.length;

    //TODO: make this a constant
    return (pseudoRandomValues[index] * 2 - 1);
  }

  _hashCode(str) {
    //From http://erlycoder.com/49/javascript-hash-functions-to-convert-string-into-integer-hash-
    var hash = 0;
    if (str.length == 0) return hash;
    for (let i = 0; i < str.length; i++) {
        let char = str.charCodeAt(i);
        hash = ((hash<<5)-hash)+char;
        hash = hash & hash;
    }
    return hash;
  }

  _classes(layout, noAnimate) {
    var result = [];
    if (layout) {
      result.push(layout);
    }
    if (noAnimate) {
      result.push("no-animate");
    }
    return result.join(" ");
  }

  _copyObj(obj) {
    let copy = {}
    for (let attr in obj) {
      if (obj.hasOwnProperty(attr)) copy[attr] = obj[attr]
    }
    return copy
  }

  _randomId(length) {
      var text = "";
      var possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";

      for( var i=0; i < length; i++ ) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
      }

      return text;
  }
}

customElements.define(BoardgameComponentStack.is, BoardgameComponentStack);
