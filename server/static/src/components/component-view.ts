import { nothing, render } from 'lit';
import type {
  Component,
  ExpandedStack,
  OpaqueComponent,
  VisibleComponent,
} from '../types/boardgame-types.js';
import { isVisibleComponent } from '../types/boardgame-types.js';
import './boardgame-card.js';
import './boardgame-token.js';
import type { BoardgameComponent } from './boardgame-component.js';
import type { BoardgameCard } from './boardgame-card.js';
import type { BoardgameToken } from './boardgame-token.js';

type StackValues<S> = S extends ExpandedStack<infer Values, object> ? Values : never;
type StackDynamicValues<S> = S extends ExpandedStack<object, infer DynamicValues> ? DynamicValues : never;

export type ComponentViewContext<S extends ExpandedStack<object, object>> =
  | Readonly<{ kind: 'empty'; component: null; index: number }>
  | Readonly<{ kind: 'hidden'; component: OpaqueComponent; index: number }>
  | Readonly<{
      kind: 'visible';
      component: VisibleComponent<StackValues<S>, StackDynamicValues<S>>;
      index: number;
    }>;

type FrameworkOwnedComponentProperty = 'boardgameComponent' | 'disabled' | 'id' | 'index' | 'item' | 'spacer';

export type SettableComponentProperties<ElementType extends BoardgameComponent> = {
  readonly [Key in keyof ElementType as Key extends FrameworkOwnedComponentProperty
    ? never
    : ElementType[Key] extends (...args: never[]) => unknown
    ? never
    : Key]?: ElementType[Key];
};

export interface ComponentViewOptions<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
> {
  /** Render light-DOM content into a stable component host. */
  readonly render?: (context: ComponentViewContext<S>) => unknown;
  /** Set typed host properties such as card faceUp/rotated or token color. */
  readonly properties?: (context: ComponentViewContext<S>) => SettableComponentProperties<ElementType>;
}

/**
 * An opaque, reusable recipe for one deck's component hosts and content.
 * Create these once on a renderer class; stack updates reuse the same hosts so
 * FLIP animation identity is preserved.
 */
export interface ComponentView<
  S extends ExpandedStack<object, object> = ExpandedStack<object, object>,
  ElementType extends BoardgameComponent = BoardgameComponent,
> {
  readonly __componentViewStack?: S;
  /** Add stack-specific, type-checked host properties without changing the recipe identity. */
  withProperties(properties: SettableComponentProperties<ElementType>): ComponentView<S, ElementType>;
}

interface InternalComponentView<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
> extends ComponentView<S, ElementType> {
  readonly create: () => ElementType;
  readonly options: ComponentViewOptions<S, ElementType>;
  readonly base?: InternalComponentView<S, ElementType>;
  readonly overrides?: SettableComponentProperties<ElementType>;
}

const initialProperties = new WeakMap<BoardgameComponent, Map<PropertyKey, unknown>>();
const appliedProperties = new WeakMap<BoardgameComponent, Set<PropertyKey>>();
const createdComponents = new WeakSet<BoardgameComponent>();
const componentTags = new WeakMap<ComponentView, string>();

export function componentView<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
>(
  create: () => ElementType,
  options: ComponentViewOptions<S, ElementType>,
): ComponentView<S, ElementType> {
  let view!: InternalComponentView<S, ElementType>;
  view = Object.freeze({
    create,
    options,
    withProperties: (properties: SettableComponentProperties<ElementType>) => bindProperties(view, properties),
  }) as InternalComponentView<S, ElementType>;
  return view;
}

/** The common card case, with card properties checked by TypeScript. */
export function cardView<S extends ExpandedStack<object, object>>(
  options: ComponentViewOptions<S, BoardgameCard>,
): ComponentView<S, BoardgameCard> {
  return componentView(
    () => document.createElement('boardgame-card'),
    options,
  );
}

/** The common token case, with token properties checked by TypeScript. */
export function tokenView<S extends ExpandedStack<object, object>>(
  options: ComponentViewOptions<S, BoardgameToken>,
): ComponentView<S, BoardgameToken> {
  return componentView(
    () => document.createElement('boardgame-token'),
    options,
  );
}

export function createComponentForView(view: ComponentView): BoardgameComponent {
  const internal = asInternalView(view);
  const base = internal.base ?? internal;
  const component = internal.create();
  assertComponentHost(component);
  if (createdComponents.has(component)) {
    throw new Error('componentView(): create() returned a component host it returned before; return a fresh element each time');
  }
  createdComponents.add(component);
  const expectedTag = componentTags.get(base);
  if (expectedTag && expectedTag !== component.localName) {
    throw new Error(`componentView(): create() changed host type from <${expectedTag}> to <${component.localName}>`);
  }
  componentTags.set(base, component.localName);
  component.setAttribute('boardgame-component', '');
  return component;
}

export function updateComponentFromView(
  view: ComponentView,
  element: BoardgameComponent,
  component: Component | null | undefined,
  index: number,
): void {
  const internal = asInternalView(view);
  const context = contextFor(component, index);
  render(internal.options.render?.(context) ?? nothing, element);

  const next = Object.assign({}, internal.options.properties?.(context) ?? {}, internal.overrides ?? {});
  const initial = initialProperties.get(element) ?? new Map<PropertyKey, unknown>();
  const previous = appliedProperties.get(element) ?? new Set<PropertyKey>();
  const nextKeys = new Set<PropertyKey>(Reflect.ownKeys(next));

  for (const key of nextKeys) {
    if (!initial.has(key)) initial.set(key, Reflect.get(element, key));
    Reflect.set(element, key, Reflect.get(next, key));
  }
  for (const key of previous) {
    if (!nextKeys.has(key)) Reflect.set(element, key, initial.get(key));
  }
  initialProperties.set(element, initial);
  appliedProperties.set(element, nextKeys);
}

/** True when two values use the same host/content recipe, even if overrides differ. */
export function sameComponentViewRecipe(first: ComponentView | null | undefined, second: ComponentView | null | undefined): boolean {
  if (!first || !second) return first === second;
  const firstInternal = asInternalView(first);
  const secondInternal = asInternalView(second);
  return (firstInternal.base ?? firstInternal) === (secondInternal.base ?? secondInternal);
}

function bindProperties<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
>(
  source: InternalComponentView<S, ElementType>,
  properties: SettableComponentProperties<ElementType>,
): ComponentView<S, ElementType> {
  const base = source.base ?? source;
  const overrides = Object.freeze({ ...(source.overrides ?? {}), ...properties });
  return Object.freeze({
    create: base.create,
    options: base.options,
    base,
    overrides,
    withProperties: (next: SettableComponentProperties<ElementType>) => bindProperties(
      { ...source, base, overrides } as InternalComponentView<S, ElementType>,
      next,
    ),
  }) as InternalComponentView<S, ElementType>;
}

function asInternalView(view: ComponentView): InternalComponentView<ExpandedStack<object, object>, BoardgameComponent> {
  const candidate = view as Partial<InternalComponentView<ExpandedStack<object, object>, BoardgameComponent>>;
  if (typeof candidate.create !== 'function' || !candidate.options || typeof candidate.withProperties !== 'function') {
    throw new Error('boardgame-component-stack: componentView must come from cardView(), tokenView(), or componentView()');
  }
  return candidate as InternalComponentView<ExpandedStack<object, object>, BoardgameComponent>;
}

function contextFor(component: Component | null | undefined, index: number): ComponentViewContext<ExpandedStack<object, object>> {
  if (component === null || component === undefined) return { kind: 'empty', component: null, index };
  if (isVisibleComponent(component)) return { kind: 'visible', component, index };
  return { kind: 'hidden', component, index };
}

function assertComponentHost(component: unknown): asserts component is BoardgameComponent {
  if (!(component instanceof HTMLElement)
    || typeof (component as Partial<BoardgameComponent>).animatingPropValues !== 'function'
    || typeof (component as Partial<BoardgameComponent>).animatingPropDefaults !== 'function'
    || typeof (component as Partial<BoardgameComponent>).playAnimation !== 'function') {
    throw new Error('componentView(): create() must return a registered element extending BoardgameComponent');
  }
  const registered = customElements.get(component.localName);
  if (!registered || !(component instanceof registered)) {
    throw new Error(`componentView(): create() returned unregistered <${component.localName || 'unknown'}>`);
  }
}
