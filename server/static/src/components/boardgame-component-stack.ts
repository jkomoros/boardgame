import { LitElement, html, css, TemplateResult } from 'lit';
import { property, query } from 'lit/decorators.js';
import './boardgame-deck-defaults.js';
import './boardgame-card.js';
import { dashToCamelCase } from '../utils/case-map.js';

// These are the random values we use. We need them to be the same for each key.
const pseudoRandomValues = [
  0.409629, 0.045224, 0.957352, 0.674598, 0.420276, 0.69529, 0.375744, 0.757951, 0.70161, 0.165333, 0.307487,
  0.761283, 0.061829, 0.444936, 0.893498, 0.568637, 0.313256, 0.096547, 0.689491, 0.735802, 0.653278, 0.568639,
  0.168512, 0.868507, 0.484359, 0.17899, 0.531255, 0.326798, 0.62229, 0.38919, 0.699555, 0.906995, 0.525003,
  0.578083, 0.174277, 0.219422, 0.601429, 0.444303, 0.012084, 0.821015, 0.761839, 0.036714, 0.376602, 0.458024,
  0.638299, 0.835771, 0.295093, 0.294265, 0.793988, 0.46952, 0.353375, 0.747226, 0.947925, 0.28802, 0.462328,
  0.61194, 0.814922, 0.38534, 0.245267, 0.439828, 0.213518, 0.347944, 0.753266, 0.462731, 0.804775, 0.284984,
  0.272913, 0.584866, 0.491474, 0.663749, 0.082535, 0.983746, 0.297612, 0.106188, 0.434022, 0.568589, 0.160467,
  0.339668, 0.266839, 0.562368, 0.446072, 0.42395, 0.372456, 0.90581, 0.245894, 0.225044, 0.952531, 0.971619,
  0.344237, 0.169056, 0.318305, 0.61021, 0.37241, 0.604048, 0.499905, 0.189648, 0.62082, 0.493149, 0.754958,
  0.736003, 0.856383, 0.493789, 0.375336, 0.171083, 0.324018, 0.522103, 0.386678, 0.771986, 0.93826, 0.409071,
  0.883499, 0.413002, 0.103175, 0.898144, 0.624233, 0.054771, 0.493724, 0.229437, 0.209876, 0.480461, 0.635064,
  0.336034, 0.373814, 0.189853, 0.789123, 0.586921, 0.393384, 0.347548, 0.153098, 0.99294, 0.923887, 0.115157,
  0.918841, 0.847155, 0.878961, 0.30781, 0.694391, 0.196018, 0.957279, 0.792493, 0.601996, 0.567903, 0.471252,
  0.55565, 0.743476, 0.905305, 0.877299, 0.886972, 0.686297, 0.713567, 0.902545, 0.764396, 0.049177, 0.269324,
  0.848572, 0.269079, 0.610327, 0.119123, 0.974389, 0.759697, 0.932941, 0.747793, 0.883499, 0.740326, 0.841617,
  0.67744, 0.957788, 0.091737, 0.904918, 0.111062, 0.066767, 0.153055, 0.442142, 0.630529, 0.88257, 0.06523,
  0.079217, 0.018493, 0.062141, 0.874116, 0.264976, 0.438222, 0.008716, 0.050499, 0.439269, 0.432986, 0.166319,
  0.555334, 0.994858, 0.513525, 0.560583, 0.326941, 0.167995, 0.980903, 0.12907, 0.944065, 0.535459, 0.260209,
  0.62893, 0.453737
];

const sharedStackList: BoardgameComponentStack[] = [];

export class BoardgameComponentStack extends LitElement {
  static styles = css`
    :host {
      width: 100%;
    }

    #container {
      position: relative;
      box-sizing: border-box;
      display: flex;
      flex-direction: row;
      align-items: center;
    }

    #container #slot-holder {
      display: flex;
      flex-direction: row;
      align-items: center;
    }

    #animating-components [boardgame-component] {
      /* don't mess with the general layout because the card will be gone soon */
      position: absolute;
    }

    #container ::slotted([boardgame-component]),
    #container [boardgame-component] {
      transition: transform var(--animation-length, 0.25s) ease-in-out, opacity var(--animation-length, 0.25s) ease-in-out;
    }

    #container.no-animate ::slotted([boardgame-component]),
    #container.no-animate [boardgame-component] {
      transition: unset;
    }

    #container.grid #slot-holder,
    #container.stack #slot-holder {
      flex-wrap: wrap;
    }

    #container ::slotted([boardgame-component]),
    #container [boardgame-component] {
      margin: 1em;
    }

    #container.pile {
      padding: calc(calc(var(--pile-scale, 1.0) * 3em) + 3em);
      justify-content: center;
    }

    #container.pile ::slotted([boardgame-component]),
    #container.pile [boardgame-component] {
      position: absolute;
      margin: 0;
    }

    .pile ::slotted([boardgame-component].bcc-first),
    .pile [boardgame-component].bcc-first {
      position: relative;
    }

    .stack ::slotted([boardgame-component]),
    .stack [boardgame-component] {
      position: absolute;
      top: 6px;
      left: 0px;
    }

    .stack ::slotted([boardgame-component].bcc-first),
    .stack [boardgame-component].bcc-first,
    .stack [boardgame-component]#spacer {
      z-index: 10;
      position: relative;
      top: 0px;
    }

    .stack ::slotted([boardgame-component]:nth-child(2)),
    .stack [boardgame-component]:nth-child(2) {
      top: 1px;
      z-index: 9;
    }

    .stack ::slotted([boardgame-component]:nth-child(3)),
    .stack [boardgame-component]:nth-child(3) {
      z-index: 8;
      top: 2px;
    }

    .stack ::slotted([boardgame-component]:nth-child(4)),
    .stack [boardgame-component]:nth-child(4) {
      z-index: 7;
      top: 3px;
    }

    .stack ::slotted([boardgame-component]:nth-child(5)),
    .stack [boardgame-component]:nth-child(5) {
      z-index: 6;
      top: 4px;
    }

    .stack ::slotted([boardgame-component]:nth-child(6)),
    .stack [boardgame-component]:nth-child(6) {
      z-index: 5;
      top: 5px;
    }

    #container.spread #slot-holder {
      display: flex;
      flex-direction: row;
      justify-content: space-around;
      width: 100%;
    }

    #container.spread #slot-holder ::slotted(dom-repeat) {
      display: none;
    }

    .spread ::slotted([boardgame-component]:hover),
    .spread [boardgame-component]:hover,
    .fan ::slotted([boardgame-component]:hover),
    .fan [boardgame-component]:hover {
      z-index: 1;
    }

    #container.spread ::slotted([boardgame-component]),
    #container.spread [boardgame-component] {
      margin-right: calc(100px * -0.75);
    }

    #container.spread ::slotted([boardgame-component].bcc-last),
    #container.spread [boardgame-component].bcc-last {
      margin-right: 0;
    }

    #container.fan ::slotted([boardgame-component]),
    #container.fan [boardgame-component] {
      margin-right: calc(100px * -0.5);
    }

    #container.fan ::slotted([boardgame-component][rotated]),
    #container.fan [boardgame-component][rotated] {
      margin-right: calc(100px * -0.25);
    }

    #container.fan ::slotted([boardgame-component].bcc-last),
    #container.fan [boardgame-component].bcc-last {
      margin-right: 0;
    }

    #container.fan {
      box-sizing: border-box;
      padding: 1em 0;
    }
  `;

  @property({ type: String })
  layout = 'stack';

  @property({ type: Object })
  stack: any = null;

  @property({ type: String })
  deckName = '';

  @property({ type: String })
  gameName = '';

  @property({ type: Boolean })
  messy = false;

  @property({ type: Number })
  messiness = 1.0;

  @property({ type: Object })
  idsLastSeen: any = null;

  @property({ type: Boolean })
  noAnimate = false;

  @property({ type: Boolean })
  noDefaultSpacer = false;

  @property({ type: Number })
  fauxComponents = 0;

  @query('#container')
  private container!: HTMLElement;

  @query('#components')
  private componentsSlot!: HTMLSlotElement;

  @query('#faux-components')
  private fauxComponentsContainer!: HTMLElement;

  @query('#animating-components')
  private animatingComponentsContainer!: HTMLElement;

  private _componentPool: any[] = [];
  private _pileScaleFactor = 1.0;
  private _randomRotationOffset = 0;
  private _id = '';
  private _style = '';

  get _sharedStackList(): BoardgameComponentStack[] {
    return sharedStackList;
  }

  get deckDefaults(): any {
    let ele = this.shadowRoot!.querySelector('boardgame-deck-defaults');
    if (ele) return ele;

    ele = document.createElement('boardgame-deck-defaults');
    this.shadowRoot!.appendChild(ele);
    return ele;
  }

  get offsetComponent(): any {
    const components = this._realComponents;
    if (components.length > 0) return components[0];
    const component = this.shadowRoot!.querySelector('[boardgame-component]');
    if (component) return component;
    return this;
  }

  get id(): string {
    return this._id;
  }

  get Components(): any[] {
    return this._realComponents.concat(this._fauxComponents);
  }

  private get _realComponents(): any[] {
    if (!this.componentsSlot) return [];

    const components = this.componentsSlot.assignedNodes({ flatten: true });
    const result: any[] = [];

    for (let i = 0; i < components.length; i++) {
      const component = components[i];
      if (!(component as any).localName || (component as Element).getAttribute('boardgame-component') !== '') {
        continue;
      }
      result.push(component);
    }

    return result;
  }

  private get _fauxComponents(): any[] {
    if (!this.fauxComponentsContainer) return [];
    const fauxComponents = this.fauxComponentsContainer.querySelectorAll('[boardgame-component]');
    return [...fauxComponents];
  }

  get templateClass(): any {
    return this.deckDefaults.templateForDeck(this.gameName, this.deckName);
  }

  connectedCallback() {
    super.connectedCallback();
    sharedStackList.push(this);
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    let i = 0;
    while (i < sharedStackList.length) {
      const item = sharedStackList[i];
      if (item === this) {
        sharedStackList.splice(i, 1);
      } else {
        i++;
      }
    }
  }

  override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
    super.firstUpdated(_changedProperties);
    this._componentPool = [];
    if (this.componentsSlot) {
      this.componentsSlot.addEventListener('slotchange', () => this._slotChanged(false));
    }
    this._slotChanged(true);
    this._randomRotationOffset = Math.floor(Math.random() * 21);
    this._id = this._randomId(8);

    const attrs = this._attributesForComponents();
    for (const key of attrs.keys()) {
      const titleKey = key[0].toUpperCase() + key.slice(1, key.length);
      // TODO: Implement property observer equivalent for Lit
      // this._createPropertyObserver('component' + titleKey, '_componentPropChanged');
    }
  }

  protected updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('layout') || changedProperties.has('messy')) {
      this._updateComponentClasses();
    }

    if (changedProperties.has('stack')) {
      this._stackChanged(this.stack);
    }

    if (changedProperties.has('gameName')) {
      this._gameNameChanged();
    }

    if (changedProperties.has('_pileScaleFactor')) {
      this._style = this._computeStyle(this._pileScaleFactor);
    }
  }

  private _computeStyle(pileScaleFactor: number): string {
    return `--pile-scale:${pileScaleFactor}`;
  }

  private _gameNameChanged() {
    this._componentPool = [];
  }

  returnComponent(ele: any) {
    this._componentPool.push(ele);
  }

  newComponent(): any {
    if (this._componentPool.length > 0) {
      return this._componentPool.pop();
    }

    const templateClass = this.templateClass;

    if (templateClass) {
      const instance = new templateClass({});
      for (const child of instance.children) {
        if (child.nodeType !== 1) continue;
        (child as any).instance = instance;
        return child;
      }
      console.warn('None of the nodes printed by the template are an actual node.');
      return null;
    }

    console.warn('No template class to auto stamp');
    return null;
  }

  stackDefault(propName: string): any {
    const propCount: any = {};
    let maxCount = 0;
    let maxVal: any = false;
    for (const c of this.Components) {
      const val = c[propName];
      if (val === undefined) {
        continue;
      }
      propCount[val] = (propCount[val] || 0) + 1;
      if (propCount[val] > maxCount) {
        maxCount = propCount[val];
        maxVal = val;
      }
    }
    return maxVal;
  }

  setUnknownAnimationState(card: any) {
    card.style.transform = 'scale(0.6)';
    card.style.opacity = '0.0';
  }

  newAnimatingComponent(): any {
    const component = this.newComponent();
    component.noAnimate = true;
    component.prepareForBeingAnimatingComponent(this);
    this.setUnknownAnimationState(component);
    this.animatingComponentsContainer.appendChild(component);
    component.addEventListener('transitionend', (e: Event) => this._clearAnimatingComponents(e));
    return component;
  }

  private _clearAnimatingComponents(e: Event) {
    const container = this.animatingComponentsContainer;
    while (container.children.length > 0) {
      const child = container.children[0];
      if ((child as any).beforeOrphaned) (child as any).beforeOrphaned();
      container.removeChild(child);
    }
  }

  private _stackChanged(newValue: any) {
    if (newValue) {
      if (newValue.Deck) {
        this.deckName = newValue.Deck;
      }
      if (newValue.GameName) {
        this.gameName = newValue.GameName;
      }
      this.idsLastSeen = newValue.IDsLastSeen || {};
    } else {
      this.deckName = '';
      this.gameName = '';
      this.idsLastSeen = null;
    }

    const repeater = this.querySelector('dom-repeat');
    if (repeater) {
      (repeater as any).items = newValue ? newValue.Components : [];
      return;
    }

    this._generateChildren();
  }

  private _attributesForComponents(): Map<string, any> {
    const result = new Map();

    for (const name of Object.getOwnPropertyNames(this)) {
      if (!name.startsWith('component')) {
        continue;
      }
      let finalName = name.replace('component', '');
      finalName = finalName[0].toLowerCase() + finalName.slice(1, finalName.length);
      result.set(finalName, (this as any)[name]);
    }

    for (const attr of this.attributes) {
      if (!attr.name.startsWith('component-')) {
        continue;
      }
      const name = attr.name.replace('component-', '');
      const finalName = dashToCamelCase(name);
      result.set(finalName, attr.value);
    }

    return result;
  }

  private _insertNodes(componentsInfo: any[], hostEle: HTMLElement) {
    const componentCount = hostEle.querySelectorAll('[boardgame-component]').length;
    const childrenToAdd = componentsInfo.length - componentCount;

    if (childrenToAdd > 0) {
      let firstNonComponentEle: Element | null = null;
      for (let i = 0; i < hostEle.children.length; i++) {
        const ele = hostEle.children[i];
        if (!ele.hasAttribute('boardgame-component')) {
          firstNonComponentEle = ele;
          break;
        }
      }

      for (let i = 0; i < childrenToAdd; i++) {
        const ele = this.newComponent();
        if (!ele) break;
        hostEle.insertBefore(ele, firstNonComponentEle);
      }
    } else if (childrenToAdd < 0) {
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
        } while (!eleToRemove.hasAttribute('boardgame-component'));

        this._componentPool.push(hostEle.removeChild(eleToRemove));
      }
    }

    const attrs = this._attributesForComponents();
    const indexAttrs = 'indexAttributes';
    const attributesToIndex = attrs.get(indexAttrs) ? attrs.get(indexAttrs).split(',') : [];

    let componentIndex = 0;

    for (let i = 0; i < hostEle.children.length; i++) {
      const ele = hostEle.children[i] as any;

      if (!ele.hasAttribute('boardgame-component')) continue;

      ele.item = componentsInfo[componentIndex];
      ele.index = componentIndex;

      if (ele.instance) {
        ele.instance.item = componentsInfo[componentIndex];
        ele.instance.index = componentIndex;
      }

      for (const key of attrs.keys()) {
        if (key === indexAttrs) continue;
        let valToSet = attrs.get(key);
        const propConfigs = ele._composedPropertyDefinition;
        if (propConfigs[key] && propConfigs[key].type === Boolean && valToSet === '') {
          valToSet = true;
        }
        ele[key] = valToSet;
      }

      for (const name of attributesToIndex) {
        const finalName = dashToCamelCase(name);
        ele[finalName] = componentIndex;
        ele.setAttribute(name, componentIndex);
      }

      componentIndex++;
    }
  }

  private _generateChildren() {
    this._insertNodes(this.stack ? this.stack.Components : [], this);
  }

  private _slotChanged(firstRender: boolean) {
    const realComponents = this._realComponents;
    const fauxComponentsContainer = this.fauxComponentsContainer;

    const wantSpacer = realComponents.length < 1 && !this.noDefaultSpacer;
    const haveSpacer = !!this.shadowRoot!.querySelector('#container>[boardgame-component][spacer]');

    if (wantSpacer && !haveSpacer) {
      const spacer = this.newComponent();
      if (spacer) {
        spacer.spacer = true;
        spacer.id = 'spacer';
        this.container.insertBefore(spacer, fauxComponentsContainer);
      }
    }

    const targetNumSpacers = wantSpacer ? 1 : 0;
    let spacers = this.shadowRoot!.querySelectorAll('#container>[boardgame-component][spacer]');

    while (spacers.length > targetNumSpacers) {
      this.container.removeChild(spacers[0]);
      spacers = this.shadowRoot!.querySelectorAll('#container>[boardgame-component][spacer]');
    }

    if (firstRender && realComponents.length < 1) return;

    if (realComponents.length < this.fauxComponents) {
      const targetNumFauxComponents = this.fauxComponents - realComponents.length;
      const info: any[] = [];

      for (let i = 0; i < targetNumFauxComponents; i++) {
        info.push(undefined);
      }

      this._insertNodes(info, fauxComponentsContainer);
    }

    this._updateComponentClasses();
  }

  private _updateComponentClasses() {
    const components = this.Components;
    let lastPileScaleFactor = 0.0;

    for (let i = 0; i < components.length; i++) {
      const component = components[i];

      const classes = ['bcc-first', 'bcc-last'];

      for (let j = 0; j < classes.length; j++) {
        component.classList.remove(classes[j]);
      }

      if (i === 0) {
        component.classList.add('bcc-first');
      }

      const transformPieces: string[] = [];
      const id = component.id || i.toString();

      if (this.messy && this.layout !== 'pile') {
        transformPieces.push(`rotate(${this._messyRotationForId(id)}deg)`);
      }

      if (this.layout === 'pile') {
        const offsets = this._pileOffsetsForId(id, components.length);
        transformPieces.push(`translate(${offsets.x}px, ${offsets.y}px)`);
        lastPileScaleFactor = offsets.scaleFactor;
        transformPieces.push(`rotate(${this._messyRotationForId(id)}deg)`);
      }

      component.style.transform = transformPieces.join(' ');

      if (i === components.length - 1) {
        component.classList.add('bcc-last');
      }

      if (this.layout !== 'stack') {
        component.noShadow = false;
        continue;
      }

      if (i < 4) {
        component.noShadow = false;
      } else {
        component.noShadow = true;
      }
    }

    if (this.layout === 'pile') {
      this._pileScaleFactor = lastPileScaleFactor;
    }

    if (this.layout === 'fan') {
      this._fanComponents();
    }
  }

  private _fanComponents() {
    const components = this.Components;

    let maxRotation = 20;
    const minRotation = maxRotation * -1;

    let maxTranslate = -1.0;
    let minTranslate = 1.5;

    const rotated = this.stackDefault('rotated');

    if ((components.length < 8 && rotated) || (!rotated && components.length < 3)) {
      const percent = 0.5;
      maxRotation *= percent;
      maxTranslate *= percent;
      minTranslate *= percent;
    }

    const rotationSpread = maxRotation - minRotation;
    const translateSpread = maxTranslate - minTranslate;

    for (let i = 0; i < components.length; i++) {
      const component = components[i];

      const percent = i / (components.length - 1);
      const rotation = percent * rotationSpread + minRotation;
      const rotationTransformation = `rotate(${rotation}deg)`;

      let translateRadians = 3.0 * percent - 1.5;
      if (percent < 0) {
        translateRadians = translateRadians * -1;
      }

      const translate = Math.cos(translateRadians) * translateSpread + minTranslate;
      const translateTransformation = `translateY(${translate}em)`;

      component.style.transform += rotationTransformation + ' ' + translateTransformation;
    }
  }

  private _messyRotationForId(id: string): number {
    let index = Math.abs(this._hashCode(id));
    index %= pseudoRandomValues.length;
    return (pseudoRandomValues[index]! * 8 - 4) * this.messiness;
  }

  private _pileOffsetsForId(
    id: string,
    numComponents: number
  ): { x: number; y: number; scaleFactor: number } {
    let x = this._randomOffsetForId(id, true);
    let y = this._randomOffsetForId(id, false);

    const triangleWidth = 0.5;

    if (y > 0) {
      let negative = x < 0;
      x = Math.abs(x);

      if (x > 1 - triangleWidth) {
        const rectX = x - (1.0 - triangleWidth);

        if (rectX < y) {
          y *= -1;
          x = 1 + rectX;
        }
      }

      if (negative) x *= -1;
    }

    const smallestSize = 20;
    const largestSize = 50;

    const lowestExpectedComponents = 5;
    const highestExpectedComponents = 25;

    const expectedComponentRange = highestExpectedComponents - lowestExpectedComponents;
    const clampedComponents = Math.min(
      Math.max(numComponents, lowestExpectedComponents),
      highestExpectedComponents
    );

    const multiplier = (clampedComponents - lowestExpectedComponents) / expectedComponentRange;
    const finalSize = smallestSize + (largestSize - smallestSize) * multiplier;

    x *= finalSize;
    y *= finalSize;

    x *= -1;
    y *= -1;

    return { x: x, y: y, scaleFactor: multiplier };
  }

  private _randomOffsetForId(id: string, x: boolean): number {
    if (x) id = id + 'right';
    id = id + this._id;

    let index = Math.abs(this._hashCode(id));
    index %= pseudoRandomValues.length;

    return pseudoRandomValues[index]! * 2 - 1;
  }

  private _hashCode(str: string): number {
    let hash = 0;
    if (str.length === 0) return hash;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash;
    }
    return hash;
  }

  private _classes(layout: string, noAnimate: boolean): string {
    const result: string[] = [];
    if (layout) {
      result.push(layout);
    }
    if (noAnimate) {
      result.push('no-animate');
    }
    return result.join(' ');
  }

  private _randomId(length: number): string {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';

    for (let i = 0; i < length; i++) {
      text += possible.charAt(Math.floor(Math.random() * possible.length));
    }

    return text;
  }

  render(): TemplateResult {
    return html`
      <div id="container" class="${this._classes(this.layout, this.noAnimate)}" style="${this._style}">
        <div id="slot-holder">
          <slot id="components"></slot>
        </div>
        <div id="faux-components"></div>
        <div id="animating-components"></div>
      </div>
    `;
  }
}

customElements.define('boardgame-component-stack', BoardgameComponentStack);
