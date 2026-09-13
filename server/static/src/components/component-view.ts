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
import { validateDieRollBudget } from './boardgame-die.js';
import type { BoardgameDie, DieRollBudget } from './boardgame-die.js';
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
  Properties extends SettableComponentProperties<ElementType> = SettableComponentProperties<ElementType>,
> {
  /** Render light-DOM content into a stable component host. */
  readonly render?: (context: ComponentViewContext<S>) => unknown;
  /** Set typed host properties such as card faceUp/rotated or token color. */
  readonly properties?: (context: ComponentViewContext<S>) => Properties;
}

/**
 * An opaque, reusable recipe for one deck's component hosts and content.
 * Create these once on a renderer class; stack updates reuse the same hosts so
 * FLIP animation identity is preserved.
 */
export interface ComponentView<
  S extends ExpandedStack<object, object> = ExpandedStack<object, object>,
  ElementType extends BoardgameComponent = BoardgameComponent,
  Properties extends SettableComponentProperties<ElementType> = SettableComponentProperties<ElementType>,
> {
  readonly __componentViewStack?: S;
  /** Add stack-specific, type-checked host properties without changing the recipe identity. */
  withProperties(properties: Properties): ComponentView<S, ElementType, Properties>;
}

interface InternalComponentView<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
  Properties extends SettableComponentProperties<ElementType>,
> extends ComponentView<S, ElementType, Properties> {
  readonly create: () => ElementType;
  readonly options: ComponentViewOptions<S, ElementType, Properties>;
  readonly base?: InternalComponentView<S, ElementType, Properties>;
  readonly overrides?: Properties;
  readonly validateProperties?: (properties: Properties) => void;
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
  return createComponentView(create, options);
}

function createComponentView<
  S extends ExpandedStack<object, object>,
  ElementType extends BoardgameComponent,
  Properties extends SettableComponentProperties<ElementType>,
>(
  create: () => ElementType,
  options: ComponentViewOptions<S, ElementType, Properties>,
  validateProperties?: (properties: Properties) => void,
): ComponentView<S, ElementType, Properties> {
  let view!: InternalComponentView<S, ElementType, Properties>;
  view = Object.freeze({
    create,
    options,
    validateProperties,
    withProperties: (properties: Properties) => bindProperties(view, properties),
  }) as InternalComponentView<S, ElementType, Properties>;
  return view;
}

/** The common card case, with card properties checked by TypeScript. */
export function cardView<S extends ExpandedStack<object, object>>(
  options: ComponentViewOptions<S, BoardgameCard>,
): ComponentView<S, BoardgameCard> {
  return componentView(
    () => {
      const card = document.createElement('boardgame-card');
      card.structuredHistoricalPresentation = true;
      return card;
    },
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

/** Dice use ordinary stack identity, layout, actions, and structural motion. */
type DieViewProperties = Omit<
  SettableComponentProperties<BoardgameDie>,
  'action' | 'faces' | 'selectedFaceIndex' | 'stackManaged'
>;

export function dieView<S extends ExpandedStack<object, object>>(
  options: ComponentViewOptions<S, BoardgameDie, DieViewProperties>
    & { readonly rollBudget?: DieRollBudget } = {},
): ComponentView<S, BoardgameDie, DieViewProperties> {
  if (options.rollBudget) validateDieRollBudget(options.rollBudget);
  const budget = options.rollBudget ? Object.freeze({ ...options.rollBudget }) : null;
  return createComponentView(() => {
    const die = document.createElement('boardgame-die');
    die.stackManaged = true;
    die.rollBudget = budget;
    return die;
  }, options, properties => {
    for (const key of ['action', 'faces', 'selectedFaceIndex', 'stackManaged'] as const) {
      if (Object.prototype.hasOwnProperty.call(properties, key)) {
        throw new Error(`dieView(): ${key} is owned by the stack-managed die host`);
      }
    }
  });
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
  const next = Object.assign({}, internal.options.properties?.(context) ?? {}, internal.overrides ?? {});
  internal.validateProperties?.(next);
  render(internal.options.render?.(context) ?? nothing, element);

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
  Properties extends SettableComponentProperties<ElementType>,
>(
  source: InternalComponentView<S, ElementType, Properties>,
  properties: Properties,
): ComponentView<S, ElementType, Properties> {
  source.validateProperties?.(properties);
  const base = source.base ?? source;
  const overrides = Object.freeze({ ...(source.overrides ?? {}), ...properties }) as Properties;
  return Object.freeze({
    create: base.create,
    options: base.options,
    base,
    overrides,
    validateProperties: source.validateProperties,
    withProperties: (next: Properties) => bindProperties(
      { ...source, base, overrides } as InternalComponentView<S, ElementType, Properties>,
      next,
    ),
  }) as InternalComponentView<S, ElementType, Properties>;
}

function asInternalView(view: ComponentView): InternalComponentView<
  ExpandedStack<object, object>, BoardgameComponent, SettableComponentProperties<BoardgameComponent>
> {
  const candidate = view as Partial<InternalComponentView<
    ExpandedStack<object, object>, BoardgameComponent, SettableComponentProperties<BoardgameComponent>
  >>;
  if (typeof candidate.create !== 'function' || !candidate.options || typeof candidate.withProperties !== 'function') {
    throw new Error('boardgame-component-stack: componentView must come from cardView(), tokenView(), dieView(), or componentView()');
  }
  return candidate as InternalComponentView<
    ExpandedStack<object, object>, BoardgameComponent, SettableComponentProperties<BoardgameComponent>
  >;
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
