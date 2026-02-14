import { BoardgameAnimatableItem } from './boardgame-animatable-item.js';
import { html, css, CSSResult, TemplateResult } from 'lit';
import { property, query } from 'lit/decorators.js';
import { classMap } from 'lit/directives/class-map.js';

export class BoardgameComponent extends BoardgameAnimatableItem {
  static override styles: any = css`
    :host {
      --default-component-scale: 1.0;
      --component-aspect-ratio: 1.0;
      --default-component-width: 30px;
      --component-effective-scale: var(--component-scale, var(--default-component-scale));
      --component-effective-width: calc(var(--component-effective-scale) * var(--component-width, var(--default-component-width)));
      --component-effective-height: calc(var(--component-effective-width) * var(--component-aspect-ratio));
    }

    /* Shadow elevation styles - copied from paper-styles */
    :host {
      --shadow-elevation-normal: 0 2px 2px 0 rgba(0, 0, 0, 0.14),
                                  0 1px 5px 0 rgba(0, 0, 0, 0.12),
                                  0 3px 1px -2px rgba(0, 0, 0, 0.2);

      --shadow-elevation-raised: 0 8px 10px 1px rgba(0, 0, 0, 0.14),
                                  0 3px 14px 2px rgba(0, 0, 0, 0.12),
                                  0 5px 5px -3px rgba(0, 0, 0, 0.4);

      --alt-shadow-elevation-normal: drop-shadow(0 2px 2px rgba(0, 0, 0, 0.14))
                                      drop-shadow(0 1px 5px rgba(0, 0, 0, 0.12))
                                      drop-shadow(0 3px 1px rgba(0, 0, 0, 0.2));

      --alt-shadow-elevation-raised: drop-shadow(0 8px 10px rgba(0, 0, 0, 0.14))
                                      drop-shadow(0 3px 14px rgba(0, 0, 0, 0.12))
                                      drop-shadow(0 5px 5px rgba(0, 0, 0, 0.4));
    }

    .spacer {
      visibility: hidden;
    }

    #outer.interactive {
      cursor: pointer;
    }

    /* CRITICAL: noAnimate barrier during measurement */
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
      box-shadow: var(--shadow-elevation-normal);
    }

    #outer.alt-shadow #inner {
      filter: var(--alt-shadow-elevation-normal);
    }

    #outer {
      transition: transform 0.1s ease-in-out;
    }

    #outer.interactive:hover {
      transform: translateY(-0.25em);
    }

    #outer.shadow.interactive:hover #inner {
      box-shadow: var(--shadow-elevation-raised);
    }

    #outer.alt-shadow.interactive:hover #inner {
      filter: var(--alt-shadow-elevation-raised);
    }

    #inner {
      /* The second part of this transition is from paper-styles/shadow */
      transition: transform var(--animation-length, 0.25s) ease-in-out,
                  box-shadow 0.28s cubic-bezier(0.4, 0, 0.2, 1),
                  filter 0.28s cubic-bezier(0.4, 0, 0.2, 1);
    }
  `;

  @property({ type: Number })
  index = 0;

  @property({ type: Object })
  item: any = null;

  @property({ type: String })
  id = '';

  @property({ type: Boolean })
  disabled = false;

  @property({ type: Boolean, attribute: 'boardgame-component', reflect: true })
  boardgameComponent = true;

  @property({ type: Boolean })
  spacer = false;

  @property({ type: Boolean })
  noShadow = false;

  @property({ type: Boolean })
  altShadow = false;

  @query('#inner')
  protected innerElement!: HTMLElement;

  protected _outerStyle = '';

  private _memoizedComposedPropertyDefinition: any = null;

  get interactive(): boolean {
    return !this.spacer && !this.disabled;
  }

  // animatingProperties should return an array of strings of property
  // names that change during animations. animatingPropValues() and
  // animatingPropDefaults() will use this.
  get animatingProperties(): string[] {
    return [];
  }

  // Returns the bundle of properties, as configured by
  // animatingProperties(), at their current value.
  animatingPropValues(): Record<string, any> {
    const result: Record<string, any> = {};
    for (const propName of this.animatingProperties) {
      result[propName] = (this as any)[propName];
    }
    return result;
  }

  // Returns the bundle of animating properties, as defined by
  // animatingProperties(), set to the defaults for the given stack. Used
  // when there isn't an element analog before or after the animation to
  // compare to.
  animatingPropDefaults(stack: any): Record<string, any> {
    const result: Record<string, any> = {};
    for (const propName of this.animatingProperties) {
      result[propName] = stack.stackDefault(propName);
    }
    return result;
  }

  // computeAnimationProps is called by prepareAnimation and startAnimation,
  // passing the raw props and returning the actual properties to set. This is
  // the override point for sub-classes like boardgame-card who actually want
  // to set other properties, not the literal ones we were provided, for
  // performance reasons. The default simply returns props.
  computeAnimationProps(isAfter: boolean, props: Record<string, any>): Record<string, any> {
    return props;
  }

  // prepareAnimation is called after the new state is databound but just
  // before animation starts. Will call computeAnimationProps to get the final
  // props to set, which is an override point for subClasses. beforeProps is
  // what this element--or one like it--returned from animatingPropValues()
  // before the databinding happened. Transform is the transform to set on the
  // top-level element. This often isn't the literal transform from before, but
  // one that has been modified to be the previous transform, combined with the
  // inversion transform to move the component visually back to where it was.
  prepareAnimation(beforeProps: Record<string, any>, transform: string, opacity: string) {
    const props = this.computeAnimationProps(false, beforeProps);
    this.setProperties(props);
    this.style.transform = transform;
    this.style.opacity = opacity;
  }

  // startAnimation is called after the new state is databound and after
  // prepareAnimation. Will call computeAnimationProps to get the final props
  // to set, which is an override point for subClasses. afterProps is what
  // this element--or one like it--returned from animatingPropValues() after
  // the databinding happened. transform and opacity are the final values for
  // those two properties in their final location.
  startAnimation(afterProps: Record<string, any>, transform: string, opacity: string) {
    const props = this.computeAnimationProps(true, afterProps);
    this.setProperties(props);
    this.style.transform = transform;
    this._expectTransitionEnd(this, 'transform');
    if (this.style.opacity !== opacity) {
      this.style.opacity = opacity;
      this._expectTransitionEnd(this, 'opacity');
    }
  }

  // prepareForBeingAnimatingComponent is called if the component is going
  // to be an animating component; that is it was created within
  // stack.newAnimatingComponent().
  prepareForBeingAnimatingComponent(stack: any) {
    // Do nothing; subclasses might do something.
  }

  // cloneContent returns whether we should clone the content of this
  // element during animating. Defaults to false; subclasses might
  // override.
  get cloneContent(): boolean {
    return false;
  }

  // animationRotates should return true if the before and after have a
  // different rotated property.
  animationRotates(beforeProps: Record<string, any>, afterProps: Record<string, any>): boolean {
    return false;
  }

  override willNotAnimate(ele: HTMLElement, propName: string): boolean {
    if (super.willNotAnimate(ele, propName)) return true;

    // Spacer causes us to be visibility:hidden, which won't generate a
    // transitionend in chrome. See https://github.com/digitaledgeit/js-
    // transition-auto/issues/1
    if (this.spacer) {
      // Spacer only makes inner and outer visibility:hidden
      if (ele.id === 'outer') return true;
      if (ele.id === 'inner') return true;
    }
    return false;
  }

  handleTap(e: Event) {
    if (!this.interactive) {
      return;
    }
    this.dispatchEvent(new CustomEvent('component-tapped', { composed: true, detail: { index: this.index } }));
  }

  protected override updated(changedProperties: Map<string, any>) {
    super.updated(changedProperties);

    if (changedProperties.has('item')) {
      this._itemChanged(this.item);
    }
  }

  protected _itemChanged(newValue: any) {
    if (newValue === undefined) return;
    if (newValue === null) {
      this.spacer = true;
      return;
    }
    this.spacer = false;
    this.id = newValue.ID || '';
  }

  protected _computeClasses(): Record<string, boolean> {
    return {
      spacer: this.spacer,
      shadow: !this.noShadow && !this.altShadow,
      'alt-shadow': !this.noShadow && this.altShadow,
      interactive: this.interactive,
      disabled: this.disabled,
      'no-animate': this.noAnimate
    };
  }

  private setProperties(props: Record<string, any>) {
    for (const key in props) {
      if (props.hasOwnProperty(key)) {
        (this as any)[key] = props[key];
      }
    }
  }

  // obj.properties, smooshed down all the way to the upper.
  get _composedPropertyDefinition(): any {
    // TODO: can we get rid of this? Doesn't seem to be used, and I believe
    // Lit does this for us now.
    if (!this._memoizedComposedPropertyDefinition) {
      const result: any = {};
      let obj: any = this;
      while (obj) {
        const props = obj.constructor.properties;
        if (!props) break;
        for (const key of Object.keys(props)) {
          result[key] = props[key];
        }
        obj = Object.getPrototypeOf(obj);
      }
      this._memoizedComposedPropertyDefinition = result;
    }
    return this._memoizedComposedPropertyDefinition;
  }

  override render(): TemplateResult {
    return html`
      <div id="outer" class="${classMap(this._computeClasses())}" @click="${this.handleTap}" style="${this._outerStyle}">
        <div id="inner">
          <slot></slot>
        </div>
      </div>
    `;
  }
}

customElements.define('boardgame-component', BoardgameComponent);
