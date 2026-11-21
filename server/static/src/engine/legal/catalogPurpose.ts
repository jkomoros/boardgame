// Go: legal/catalog_purpose.go and legal/catalog_component_scalar.go. The
// purpose-built predicates (a short hand-written branch over a fixed set of
// paths) and the sealed static component-value comparisons.

import type { LegalComponent, LegalContext, LegalEnumValue } from './access.ts';
import { enumValuesEqual, readEnumProp, specArgs, specError, specTemplate, unknownFrom } from './catalogSupport.ts';
import { goAtoi, goQuote } from './goformat.ts';
import { type Result, resolveEnumPath, resolveIntPath, resolveStackPath } from './path.ts';
import { legalPanic } from './playerIndex.ts';
import type { LegalPanic } from './playerIndex.ts';
import { type LegalComponentFieldRequirement, type LegalPredicateConstructor, newLegalPredicate, unionLegalReads } from './predicate.ts';
import {
  TEMPLATE_ALREADY_REVEALED,
  TEMPLATE_COMPONENT_PROPS_EQUAL,
  TEMPLATE_COMPONENT_PROPS_NOT_EQUAL,
  TEMPLATE_COMPONENT_PROP_EQUALS,
  TEMPLATE_COMPONENT_PROP_NOT_CURRENT_PLAYER,
  TEMPLATE_COMPONENT_PROP_NOT_EQUALS,
  TEMPLATE_NO_CARD_HERE,
} from './templates.ts';
import {
  LEGAL_COST_CHEAP,
  LEGAL_COST_TRIVIAL,
  type LegalRead,
  type LegalSpec,
  type PropertyType,
  failT,
  passVerdict,
  unknownVerdict,
} from './types.ts';

/** Go: legal.RevealableCardAt. */
export function revealableCardAt(hiddenPath: string, visiblePath: string, idxField: string): LegalSpec {
  return { name: 'revealableCardAt', args: [hiddenPath, visiblePath, idxField] };
}

/** Go: legal.ComponentPropEqualsCurrentPlayer. */
export function componentPropEqualsCurrentPlayer(stackPath: string, keyField: string, prop: string): LegalSpec {
  return { name: 'componentPropEqualsCurrentPlayer', args: [stackPath, keyField, prop] };
}

/**
 * Go: revealableCardAtConstructor. Two mirrored stacks: a card is revealable
 * at an index when the hidden stack holds one there. Otherwise the failure
 * says which of two things is true, by occupancy alone: nothing was ever
 * there, or it is already in the visible stack.
 */
export function revealableCardAtConstructor(): LegalPredicateConstructor {
  return {
    name: 'revealableCardAt',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: revealableCardAt requires 3 args (hiddenPath, visiblePath, idxField), got ${args.length}`);
      }
      const [hiddenPath, visiblePath, idxField] = args;
      const noCardTemplate = specTemplate(spec, TEMPLATE_NO_CARD_HERE);
      const alreadyRevealedTemplate = specTemplate(spec, TEMPLATE_ALREADY_REVEALED);
      return newLegalPredicate({
        name: 'revealableCardAt',
        clientEvaluable: true,
        args: [...args],
        reads: [
          { path: hiddenPath, facet: 'occupancy' },
          { path: visiblePath, facet: 'occupancy' },
          { path: idxField, facet: 'values' },
        ],
        requiredReadTypes: { [hiddenPath]: 'TypeStack', [visiblePath]: 'TypeStack', [idxField]: 'TypeInt' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [noCardTemplate, alreadyRevealedTemplate],
        emittedBindings: { [noCardTemplate]: [], [alreadyRevealedTemplate]: [] },
        evaluate(ctx) {
          const idx = resolveIntPath(idxField, ctx);
          if (!idx.ok) {
            return unknownFrom(idx);
          }
          const hidden = resolveStackPath(hiddenPath, ctx);
          if (!hidden.ok) {
            return unknownFrom(hidden);
          }
          const visible = resolveStackPath(visiblePath, ctx);
          if (!visible.ok) {
            return unknownFrom(visible);
          }
          if (hidden.value === null) {
            return unknownVerdict(`legal: stack path ${goQuote(hiddenPath)} resolved to nil`);
          }
          if (visible.value === null) {
            return unknownVerdict(`legal: stack path ${goQuote(visiblePath)} resolved to nil`);
          }
          if (hidden.value.componentAt(idx.value) !== null) {
            return passVerdict();
          }
          if (visible.value.componentAt(idx.value) === null) {
            return failT(noCardTemplate);
          }
          return failT(alreadyRevealedTemplate);
        },
      });
    },
  };
}

/**
 * Go: componentPropEqualsCurrentPlayerConstructor. The component at an
 * enum-keyed slot must have the same enum value for `prop` as the current
 * player does (a checkers token's Color against the player's Color).
 */
export function componentPropEqualsCurrentPlayerConstructor(): LegalPredicateConstructor {
  return {
    name: 'componentPropEqualsCurrentPlayer',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: componentPropEqualsCurrentPlayer requires 3 args (stackPath, keyField, prop), got ${args.length}`);
      }
      const [stackPath, keyField, prop] = args;
      const playerPath = 'player.' + prop;
      const template = specTemplate(spec, TEMPLATE_COMPONENT_PROP_NOT_CURRENT_PLAYER);
      return newLegalPredicate({
        name: 'componentPropEqualsCurrentPlayer',
        clientEvaluable: true,
        args: [...args],
        reads: [
          { path: stackPath, facet: 'values' },
          { path: keyField, facet: 'values' },
          { path: playerPath, facet: 'values' },
        ],
        requiredReadTypes: { [stackPath]: 'TypeStack', [keyField]: 'TypeEnum', [playerPath]: 'TypeEnum' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['prop'] },
        evaluate(ctx) {
          const key = resolveEnumPath(keyField, ctx);
          if (!key.ok) {
            return unknownFrom(key);
          }
          if (key.value === null) {
            return unknownVerdict(`legal: enum path ${goQuote(keyField)} resolved to nil`);
          }
          const stack = resolveStackPath(stackPath, ctx);
          if (!stack.ok) {
            return unknownFrom(stack);
          }
          if (stack.value === null) {
            return unknownVerdict(`legal: stack path ${goQuote(stackPath)} resolved to nil`);
          }
          const comp = stack.value.componentAt(key.value.value);
          if (comp === null) {
            return unknownVerdict(`legal: no component at key ${key.value.string} in stack ${goQuote(stackPath)}`);
          }
          const values = comp.values();
          if (values === null) {
            // Go dereferences comp.Values().Reader() unguarded here.
            throw legalPanic('runtime error: invalid memory address or nil pointer dereference');
          }
          const compVal = readEnumProp(values, prop);
          if (!compVal.ok) {
            return unknownFrom(compVal);
          }
          if (compVal.value === null) {
            return unknownVerdict(`legal: component's enum property ${goQuote(prop)} resolved to nil`);
          }
          const playerVal = resolveEnumPath(playerPath, ctx);
          if (!playerVal.ok) {
            return unknownFrom(playerVal);
          }
          if (playerVal.value === null) {
            return unknownVerdict(`legal: enum path ${goQuote(playerPath)} resolved to nil`);
          }
          if (enumValuesEqual(compVal.value, playerVal.value)) {
            return passVerdict();
          }
          return failT(template, { prop });
        },
      });
    },
  };
}

/**
 * Go: legal.ComponentSelector. Names one component of a stack, from a closed
 * vocabulary: the first occupied slot, the n-th occupied slot, or the slot an
 * int move field indexes.
 */
export interface ComponentSelector {
  readonly stackPath: string;
  /** The selector's wire spelling, e.g. "occupied-ordinal:1". */
  readonly wire: string;
}

/** Go: legal.FirstOccupied. */
export function firstOccupied(stackPath: string): ComponentSelector {
  return { stackPath, wire: 'first-occupied' };
}

/** Go: legal.OccupiedOrdinal. Zero-based among occupied slots. */
export function occupiedOrdinal(stackPath: string, ordinal: number): ComponentSelector {
  return { stackPath, wire: 'occupied-ordinal:' + ordinal };
}

/** Go: legal.MoveIndex. `moveIndexPath` is an int-valued move.* path. */
export function moveIndex(stackPath: string, moveIndexPath: string): ComponentSelector {
  return { stackPath, wire: 'move-index:' + moveIndexPath };
}

/** Go: legal.ComponentPropEquals. */
export function componentPropEquals(selector: ComponentSelector, field: string, value: string): LegalSpec {
  return { name: 'componentPropEquals', args: [selector.stackPath, selector.wire, field, value] };
}

/** Go: legal.ComponentPropNotEquals. */
export function componentPropNotEquals(selector: ComponentSelector, field: string, value: string): LegalSpec {
  return { name: 'componentPropNotEquals', args: [selector.stackPath, selector.wire, field, value] };
}

/** Go: legal.ComponentPropsEqual. */
export function componentPropsEqual(left: ComponentSelector, right: ComponentSelector, field: string): LegalSpec {
  return { name: 'componentPropsEqual', args: [left.stackPath, left.wire, right.stackPath, right.wire, field] };
}

/** Go: legal.ComponentPropsNotEqual. */
export function componentPropsNotEqual(left: ComponentSelector, right: ComponentSelector, field: string): LegalSpec {
  return { name: 'componentPropsNotEqual', args: [left.stackPath, left.wire, right.stackPath, right.wire, field] };
}

/** Go: parsedComponentSelector. */
interface ParsedSelector {
  stackPath: string;
  kind: 'first-occupied' | 'occupied-ordinal' | 'move-index';
  ordinal: number;
  indexPath: string;
}

/** Go: parseComponentSelector. Throws the bare error fragment; callers add their prefix. */
function parseSelector(stackPath: string, wire: string): ParsedSelector {
  if (stackPath === '') {
    specError('stack path is empty');
  }
  if (wire === 'first-occupied') {
    return { stackPath, kind: 'first-occupied', ordinal: 0, indexPath: '' };
  }
  if (wire.startsWith('occupied-ordinal:')) {
    const raw = wire.slice('occupied-ordinal:'.length);
    const ordinal = goAtoi(raw);
    if (!ordinal.ok || ordinal.value < 0) {
      specError(`occupied ordinal must be a non-negative integer, got ${goQuote(raw)}`);
    }
    return { stackPath, kind: 'occupied-ordinal', ordinal: ordinal.value, indexPath: '' };
  }
  if (wire.startsWith('move-index:')) {
    const path = wire.slice('move-index:'.length);
    if (!path.startsWith('move.') || path.length === 'move.'.length) {
      specError(`move index selector requires a move.* path, got ${goQuote(path)}`);
    }
    return { stackPath, kind: 'move-index', ordinal: 0, indexPath: path };
  }
  return specError(`unknown component selector ${goQuote(wire)}`);
}

function parseSelectorWithPrefix(prefix: string, stackPath: string, wire: string): ParsedSelector {
  try {
    return parseSelector(stackPath, wire);
  } catch (err) {
    return specError(prefix + (err instanceof Error ? err.message : String(err)));
  }
}

function selectorReads(s: ParsedSelector): LegalRead[] {
  const reads: LegalRead[] = [{ path: s.stackPath, facet: 'values' }];
  if (s.kind === 'move-index') {
    reads.push({ path: s.indexPath, facet: 'values' });
  }
  return reads;
}

function selectorTypes(selectors: readonly ParsedSelector[]): Record<string, PropertyType> {
  const out: Record<string, PropertyType> = {};
  for (const s of selectors) {
    out[s.stackPath] = 'TypeStack';
    if (s.kind === 'move-index') {
      out[s.indexPath] = 'TypeInt';
    }
  }
  return out;
}

/**
 * Go: parsedComponentSelector.resolve. The selected component, null when the
 * selector names an empty or out-of-range position (Go: (nil, nil)), or Go's
 * error text.
 */
function resolveSelector(s: ParsedSelector, ctx: LegalContext): Result<LegalComponent | null> {
  const stack = resolveStackPath(s.stackPath, ctx);
  if (!stack.ok) {
    return stack;
  }
  if (stack.value === null) {
    return { ok: false, error: `stack ${goQuote(s.stackPath)} was nil` };
  }
  if (s.kind === 'first-occupied') {
    return { ok: true, value: stack.value.first() };
  }
  if (s.kind === 'occupied-ordinal') {
    let seen = 0;
    const len = stack.value.len();
    for (let i = 0; i < len; i++) {
      const component = stack.value.componentAt(i);
      if (component === null) {
        continue;
      }
      if (seen === s.ordinal) {
        return { ok: true, value: component };
      }
      seen++;
    }
    return { ok: true, value: null };
  }
  const index = resolveIntPath(s.indexPath, ctx);
  if (!index.ok) {
    return index;
  }
  if (index.value < 0 || index.value >= stack.value.len()) {
    return { ok: true, value: null };
  }
  return { ok: true, value: stack.value.componentAt(index.value) };
}

/** Go: componentScalar. A component's string or enum field, read for comparison. */
type ComponentScalar = { type: 'TypeString'; text: string } | { type: 'TypeEnum'; enum: LegalEnumValue | null };

/** Go: readComponentScalar. */
function readScalar(component: LegalComponent, field: string): Result<ComponentScalar> {
  const reader = component.values();
  if (reader === null) {
    return { ok: false, error: 'selected component has no values reader' };
  }
  const props = reader.props();
  if (!Object.prototype.hasOwnProperty.call(props, field)) {
    return { ok: false, error: `component field ${goQuote(field)} does not exist` };
  }
  const type = props[field];
  if (type !== 'TypeString' && type !== 'TypeEnum') {
    return { ok: false, error: `component field ${goQuote(field)} has unsupported PropertyType ${type}` };
  }
  let value: unknown;
  try {
    value = reader.prop(field);
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : String(err) };
  }
  if (type === 'TypeString') {
    return { ok: true, value: { type, text: typeof value === 'string' ? value : '' } };
  }
  return { ok: true, value: { type, enum: (value as LegalEnumValue | null) ?? null } };
}

/** Go: componentScalar.display. */
function display(s: ComponentScalar): string {
  if (s.type === 'TypeEnum') {
    return s.enum !== null ? s.enum.string : '';
  }
  return s.text;
}

/** Go: componentScalar.equalsLiteral. */
function equalsLiteral(s: ComponentScalar, value: string): Result<boolean> {
  if (s.type === 'TypeString') {
    return { ok: true, value: s.text === value };
  }
  if (s.enum === null) {
    return { ok: false, error: 'enum component field has no enum' };
  }
  const key = s.enum.enum.valueFromString(value);
  if (key === null) {
    return { ok: false, error: `${goQuote(value)} is not a value in enum ${goQuote(s.enum.enum.name)}` };
  }
  return { ok: true, value: s.enum.value === key };
}

/** Go: componentScalar.equals. */
function scalarsEqual(a: ComponentScalar, b: ComponentScalar): Result<boolean> {
  if (a.type !== b.type) {
    return { ok: false, error: `component scalar types differ (${a.type} and ${b.type})` };
  }
  if (a.type === 'TypeString') {
    return { ok: true, value: a.text === (b as { text: string }).text };
  }
  const other = (b as { enum: LegalEnumValue | null }).enum;
  if (a.enum === null || other === null) {
    return { ok: false, error: 'enum component field was nil' };
  }
  return { ok: true, value: enumValuesEqual(a.enum, other) };
}

/** Go's `%v` of the (component, err) pair when no component was selected. */
function selectionError(result: Result<LegalComponent | null>): string {
  return result.ok ? '<nil>' : result.error;
}

/** Go: componentPropLiteralConstructor. One component's field against a literal. */
function componentPropLiteral(name: string, defaultTemplate: string, negate: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 4) {
        specError(`legal: ${name} requires 4 args (stackPath, selector, field, value), got ${args.length}`);
      }
      const selector = parseSelectorWithPrefix(`legal: ${name}: `, args[0], args[1]);
      const field = args[2];
      const want = args[3];
      if (field === '') {
        specError(`legal: ${name} requires a non-empty component field`);
      }
      const template = specTemplate(spec, defaultTemplate);
      const requirement: LegalComponentFieldRequirement = {
        stackPath: selector.stackPath,
        field,
        allowedTypes: ['TypeString', 'TypeEnum'],
        requiredEnumValue: want,
      };
      return newLegalPredicate({
        name,
        args: [...args],
        reads: selectorReads(selector),
        cost: LEGAL_COST_TRIVIAL,
        requiredReadTypes: selectorTypes([selector]),
        requiredComponentFields: [requirement],
        clientEvaluable: false,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['prop', 'value', 'want'] },
        evaluate(ctx) {
          const component = resolveSelector(selector, ctx);
          if (!component.ok || component.value === null) {
            return unknownVerdict(`legal: ${name} could not select a component: ${selectionError(component)}`);
          }
          const actual = readScalar(component.value, field);
          if (!actual.ok) {
            return unknownVerdict(`legal: ${name}: ${actual.error}`);
          }
          const matches = equalsLiteral(actual.value, want);
          if (!matches.ok) {
            return unknownVerdict(`legal: ${name}: ${matches.error}`);
          }
          if (matches.value !== negate) {
            return passVerdict();
          }
          return failT(template, { prop: field, value: display(actual.value), want });
        },
      });
    },
  };
}

/** Go: componentPropsPairConstructor. The same field on two selected components. */
function componentPropsPair(name: string, defaultTemplate: string, negate: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 5) {
        specError(`legal: ${name} requires 5 args (leftStack, leftSelector, rightStack, rightSelector, field), got ${args.length}`);
      }
      const left = parseSelectorWithPrefix(`legal: ${name} left selector: `, args[0], args[1]);
      const right = parseSelectorWithPrefix(`legal: ${name} right selector: `, args[2], args[3]);
      const field = args[4];
      if (field === '') {
        specError(`legal: ${name} requires a non-empty component field`);
      }
      const template = specTemplate(spec, defaultTemplate);
      const comparableTypeGroup = name + ':' + field;
      const requirements: LegalComponentFieldRequirement[] = [
        { stackPath: left.stackPath, field, allowedTypes: ['TypeString', 'TypeEnum'], comparableTypeGroup },
      ];
      if (right.stackPath !== left.stackPath) {
        requirements.push({ stackPath: right.stackPath, field, allowedTypes: ['TypeString', 'TypeEnum'], comparableTypeGroup });
      }
      return newLegalPredicate({
        name,
        args: [...args],
        reads: unionLegalReads([selectorReads(left), selectorReads(right)]),
        cost: LEGAL_COST_TRIVIAL,
        requiredReadTypes: selectorTypes([left, right]),
        requiredComponentFields: requirements,
        clientEvaluable: false,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['prop', 'left', 'right'] },
        evaluate(ctx) {
          const leftComponent = resolveSelector(left, ctx);
          if (!leftComponent.ok || leftComponent.value === null) {
            return unknownVerdict(`legal: ${name} could not select left component: ${selectionError(leftComponent)}`);
          }
          const rightComponent = resolveSelector(right, ctx);
          if (!rightComponent.ok || rightComponent.value === null) {
            return unknownVerdict(`legal: ${name} could not select right component: ${selectionError(rightComponent)}`);
          }
          const leftValue = readScalar(leftComponent.value, field);
          if (!leftValue.ok) {
            return unknownVerdict(`legal: ${name}: ${leftValue.error}`);
          }
          const rightValue = readScalar(rightComponent.value, field);
          if (!rightValue.ok) {
            return unknownVerdict(`legal: ${name}: ${rightValue.error}`);
          }
          const matches = scalarsEqual(leftValue.value, rightValue.value);
          if (!matches.ok) {
            return unknownVerdict(`legal: ${name}: ${matches.error}`);
          }
          if (matches.value !== negate) {
            return passVerdict();
          }
          return failT(template, { prop: field, left: display(leftValue.value), right: display(rightValue.value) });
        },
      });
    },
  };
}

/** Go: componentPropEqualsConstructor. */
export function componentPropEqualsConstructor(): LegalPredicateConstructor {
  return componentPropLiteral('componentPropEquals', TEMPLATE_COMPONENT_PROP_EQUALS, false);
}

/** Go: componentPropNotEqualsConstructor. */
export function componentPropNotEqualsConstructor(): LegalPredicateConstructor {
  return componentPropLiteral('componentPropNotEquals', TEMPLATE_COMPONENT_PROP_NOT_EQUALS, true);
}

/** Go: componentPropsEqualConstructor. */
export function componentPropsEqualConstructor(): LegalPredicateConstructor {
  return componentPropsPair('componentPropsEqual', TEMPLATE_COMPONENT_PROPS_EQUAL, false);
}

/** Go: componentPropsNotEqualConstructor. */
export function componentPropsNotEqualConstructor(): LegalPredicateConstructor {
  return componentPropsPair('componentPropsNotEqual', TEMPLATE_COMPONENT_PROPS_NOT_EQUAL, true);
}
