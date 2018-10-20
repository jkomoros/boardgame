import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import browserDetect from '../node_modules/browser-detect/dist/browser-detect.es5.js';
import './boardgame-component-stack.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class BoardgameComponentAnimator extends PolymerElement {
  static get template() {
    return html`
    <boardgame-component-stack id="stack" no-default-spacer=""></boardgame-component-stack>
`;
  }

  static get is() {
    return "boardgame-component-animator"
  }

  static get properties() {
    return {
      _infoById: Object,
      _lastSeenNodesById: Object,
      _beforeSeenIds: Object,
      _animatingComponents: Array,
      _beforeCollectionOffsets: Object,
      ancestorOffsetParent: Object
    }
  }

  _calculateOffsets(ele) {

    var top = 0;
    var left = 0;
    var width = ele.offsetWidth;
    var height = ele.offsetHeight;

    var offsetEle = ele;
    while (offsetEle) {

      top += offsetEle.offsetTop;
      left += offsetEle.offsetLeft;

      if (offsetEle == this.ancestorOffsetParent) {
        offsetEle = null;
      } else {
        offsetEle = offsetEle.offsetParent;
      }
    }

    return {
      top: top,
      left: left,
      width: width,
      height: height
    };
  }

  ready() {
    super.ready();
    this._lastSeenNodesById = new Map();
    //This calculation takes a bit of time to run, so do it now so that
    //it's ready before the first animation is triggered.
    let browsername = browserDetect().name
  }

  prepare() {

    var collections = this.$.stack._sharedStackList;

    this._beforeCollectionOffsets = new Map();

    var result = {};

    //keep track of all of the ids we've seen this round to make sure we
    //found a home for all of them in the end.
    this._beforeSeenIds = new Set();

    for (var i = 0; i < collections.length; i++) {
      var collection = collections[i];

      var offsetComponent = collection.offsetComponent;
      this._beforeCollectionOffsets.set(collection.id, this._calculateOffsets(offsetComponent));

      var components = collection.Components;
      for (var j = 0; j < components.length; j++) {
        var component = components[j];

        component.resetAnimating();

        //Skip comoonents without ids (e.g. faux-components, spacer
        //components).
        if (component.id == "") continue;

        var record = result[component.id] || {};

        this._beforeSeenIds.add(component.id);

        record.offsets = this._calculateOffsets(component);

        //We use getComputedStyle instead of just card.style.transform,
        //because if the card is in the middle of transforming, we want
        //the exact value at that second, not what the logical final value
        //is.

        var computedStyle = getComputedStyle(component)

        record.beforeTransform = computedStyle.transform;

        if (record.beforeTransform == "none") {
          record.beforeTransform = "";
        }

        record.before = component.animatingPropValues();

        if (component.cloneContent) {
          var newNodes = [];
          //If the card has old-front, that's a signal that content is bad
          //and not worth copying.
          var children = component.children;
          for (var k = 0; k < children.length; k++) {
            var child = children[k];
            if (child.slot) {
              //Skip content that doesn't go in default slot
              continue
            }
            if (child.localName == "dom-bind") {
              continue;
            }
            newNodes.push(child.cloneNode(true));
          }
          if (newNodes.length > 0) {
            this._lastSeenNodesById.set(component.id, newNodes);
          }
        }
        result[component.id] = record;
      }
    }

    this._infoById = result;
  }

  animate() {
    //Wait for the style to be set--but BEFORE a frame is rendered! On
    //Chrome, requestAnimationFrame happens right before this--but
    //microTask timing isn't sufficiently late. On Safari,
    //requestAnimationFrame is already too late, and you'll see a visual
    //glitch if you wait until then.

    //TODO: there's got to be a better way to test which behavior to use,
    //but I don't know how to do it reliably, so falling back to the hack
    //of browser detection.
    let detector = browserDetect();
    if (detector && detector.name == "Safari") {
      console.log("Bam!")
      Promise.resolve().then(() => this._doAnimate());
    } else {
      window.requestAnimationFrame(()=> this._doAnimate());
    }
  }

  _doAnimate() {
    var collections = this.$.stack._sharedStackList;

    //The last seen location of a given card ID
    var idToPossibleCollection = new Map();

    var collectionOffsets = new Map();

    //Turning off animations and setting card flip all require recalcing
    //style so do them once before readback in the second loop.

    for (var i = 0; i < collections.length; i++) {
      var collection = collections[i];
      collection.noAnimate = true;
      var components = collection.Components;
      for (var j = 0; j < components.length; j++) {
        var component = components[j];
        if (component.id == "") continue;
        component.noAnimate = true; 
      }
    }

    //This layout readback is the most important thing to do quickly
    //because if we thrash the DOM there will be a lot of recalc style. So
    //do it in its own pass.
    for (var i = 0; i < collections.length; i++) {
      var collection = collections[i];

      var offsetComponent = collection.offsetComponent;
      collectionOffsets.set(collection.id, this._calculateOffsets(offsetComponent));

      //Note which Ids were last seen here
      this._ingestStack(idToPossibleCollection, collection);

      var components = collection.Components;
      for (var j = 0; j < components.length; j++) {
        var component = components[j];
        if (component.id == "") continue;
        var record = this._infoById[component.id];
        if (!record) {
          record = {};
          this._infoById[component.id] = record;
        }
        record.newOffsets = this._calculateOffsets(component);
      }
    }

    //This is the meat of the method, where we set all layout-affecting
    //properties, append fake dom, etc.
    for (var i = 0; i < collections.length; i++) {

      var collection = collections[i];

      var components = collection.Components;
      for (var j = 0; j < components.length; j++) {
        var component = components[j];

        if (component.id == "") continue;

        var record = this._infoById[component.id];

        if (!record.offsets) {

          //Hmm, a record who didn't have its offsets set in prepare(),
          //presumably because it didn't exist. This MAY be an element who
          //came from a PolicyNonEmpty stack.

          var collectionRecord = idToPossibleCollection.get(component.id);

          if (!collectionRecord) {
            //Nah, we don't know where it came from. Just skip animating it.
            continue;
          }

          var theStack = collectionRecord.stack;
          //We actually want the runner up, if it exists. the winner is
          //the stack it's now in, and teh runner up should be where it
          //just came from.
          if (collectionRecord.runnerUpStack) {
            theStack = collectionRecord.runnerUpStack;
          }

          record.offsets = this._beforeCollectionOffsets.get(theStack.id);

          record.before = component.animatingPropDefaults(theStack);
          
          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;

          theStack.setUnknownAnimationState(component);

          //TODO: can this move up to be next to setting record.before for
          //clarity, or todes it depend on being after
          //theStack.setUnknownAnimationState?
          record.beforeTransform = component.style.transform;

        } else {
          record.afterOpacity = component.style.opacity;
          record.afterTransform = component.style.transform;
        }

        //Mark that we've seen where this one is going.
        this._beforeSeenIds.delete(component.id);

        record.after = component.animatingPropValues();

        var invertTop = record.offsets.top - record.newOffsets.top;
        var invertLeft = record.offsets.left - record.newOffsets.left;

        var scaleFactor= record.offsets.width / record.newOffsets.width;

        //If the before and after are rotated differently then the scale
        //factor will need to compare height vs width to get the right
        //scale factor.
        if (component.animationRotates(record.before, record.after)) {
          scaleFactor = record.offsets.height / record.newOffsets.width;
        }

        //The containing box has physically shrunk (or grown), and the
        //transform will make its apparent edge be that much smaller or
        //bigger, so correct for that.
        invertTop -= (record.newOffsets.height - record.offsets.height) / 2;
        invertLeft -= (record.newOffsets.width - record.offsets.width) / 2;

        //We used to only bother setting transforms for items that had
        //physically moved. However, the browser is smart enough to ignore
        //transforms that are basically no ops. And if we don't set it
        //then cards that don't physically move but do have transform
        //changes won't animate because the transform was set during
        //noAnimate and is never set to anything different. In testing
        //this didn't appear to have any appreciable performance difference.
        var transform = `translateY(${invertTop}px) translateX(${invertLeft}px)`
        var scaleTransform = `scale(${scaleFactor})`
        let beforeInvertedTransform = transform + " " + record.beforeTransform + " " + scaleTransform;

        //TODO: what should opacity be?
        component.prepareAnimation(record.before, beforeInvertedTransform, "1.0");

        var clonedNodes = this._lastSeenNodesById.get(component.id);

        if (clonedNodes && clonedNodes.length > 0) {

          //Clear out old nodes.
          for (var k = 0; k < component.children.length; k++) {
            var child = component.children[k];
            if (child.slot == "fallback") {
              component.removeChild(child);
            }
          }
          for (var k = 0; k < clonedNodes.length; k++) {
            var node = clonedNodes[k];
            node.slot = "fallback";
            component.appendChild(node);
          }
        }
        
      }
    }


    this._animatingComponents = [];

    //Any items still in _beforeSeenIds did not have a specific card to
    //animate to. Let's see if we can figure out which collection they
    //went to.
    for (let id of this._beforeSeenIds) {

      //Which stack do we think this is in now?
      var anonRecord = idToPossibleCollection.get(id);

      if (!anonRecord) {
        //Guess it's a mystery. :-(
        continue;
      }

      var component = anonRecord.stack.newAnimatingComponent();

      var record = this._infoById[id];

      record.after = component.animatingPropDefaults(anonRecord.stack),

      

      this._animatingComponents.push({
        stack: anonRecord.stack,
        component: component,
        after: record.after,
        afterTransform: component.style.transform,
        afterOpacity: component.style.opacity,
      })

      var stackLocation = collectionOffsets.get(anonRecord.stack.id);
      var oldLocation = record.offsets;

      var invertTop = oldLocation.top - stackLocation.top;
      var invertLeft = oldLocation.left - stackLocation.left;

      invertTop -= (stackLocation.height - oldLocation.height) / 2;
      invertLeft -= (stackLocation.width - oldLocation.width) / 2;

      var scaleFactor= oldLocation.width / stackLocation.width;

      if (component.animationRotates(record.before, record.after)) {
        //The before anda after are different rotations which means the
        //invert top and left have to be tweaked.
        scaleFactor = oldLocation.height / stackLocation.width;
      }

      //We used to only bother setting transforms for items that had
      //physically moved. However, the browser is smart enough to ignore
      //transforms that are basically no ops. And if we don't set it
      //then cards that don't physically move but do have transform
      //changes won't animate because the transform was set during
      //noAnimate and is never set to anything different. In testing
      //this didn't appear to have any appreciable performance difference.
      var transform = `translateY(${invertTop}px) translateX(${invertLeft}px)`;
      var scaleTransform = `scale(${scaleFactor})`
      
      let beforeInvertedTransform = component.style.transform = transform + " " + record.beforeTransform + " " + scaleTransform;
      let beforeOpacity = component.style.opacity = "1.0";

      component.prepareAnimation(record.before, beforeInvertedTransform, beforeOpacity);

      var clonedNodes = this._lastSeenNodesById.get(id);
      if (clonedNodes) {
        for (var k = 0; k < clonedNodes.length; k++) {
          var node = clonedNodes[k];
          node.slot = "fallback";
          component.appendChild(node);
        }
      }
    }

    //Wait for styles to be set to do the animations
    window.requestAnimationFrame(() => this._startAnimations());

  }

  _startAnimations() {
    var collections = this.$.stack._sharedStackList;

    for (var i = 0; i < collections.length; i++) {
      var collection = collections[i];
      collection.noAnimate = false;
      var components = collection.Components;
      for (var j = 0; j < components.length; j++) {
        var component = components[j];
        if (component.id == "") continue;
        var record = this._infoById[component.id];
        if (!record) continue;
        component.noAnimate = false;
        component.startAnimation(record.after, record.afterTransform, record.afterOpacity);
      }
    }

    for (var i = 0; i < this._animatingComponents.length; i++) {
      var record = this._animatingComponents[i];
      record.component.noAnimate = false;

      record.component.startAnimation(record.after, record.afterTransform, record.afterOpacity);
    }

  }

  _ingestStack(possibleLocations, stack) {

    var idsLastSeen = stack.idsLastSeen;

    for (var key in idsLastSeen) {
      if (!idsLastSeen.hasOwnProperty(key)) continue;

      if (possibleLocations.has(key)) {

        var record = possibleLocations.get(key);

        if (idsLastSeen[key] > record.version) {
          //new winner
          var newRecord = {
            version: idsLastSeen[key],
            stack: stack,
            runnerUpVersion: record.version,
            runnerUpStack: record.stack
          };
          possibleLocations.set(key, newRecord)
          record = newRecord;
        }

        if (!record.runnerUpStack || idsLastSeen[key] > record.runnerUpVersion) {
          //Found a new second!
          possibleLocations.set(key, {
            version: record.version,
            stack: record.stack,
            runnerUpVersion: idsLastSeen[key],
            runnerUpStack: stack
          })
        }

      } else {
        //We're the first one that's been seen; add it.
        possibleLocations.set(key, {
          version: idsLastSeen[key],
          stack: stack
        })
      }

    }

  }
}

customElements.define(BoardgameComponentAnimator.is, BoardgameComponentAnimator);
