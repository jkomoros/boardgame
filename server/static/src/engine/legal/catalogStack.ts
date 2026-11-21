// Go: legal/catalog_stack.go and legal/catalog_count.go. The stack-shaped
// predicates: occupancy at an index or key, the MayMove/MaySwap family, and
// counts and emptiness.
//
// None of these compute stack legality themselves. Whether a component may
// move is the stack's own answer (LegalComponent.mayMoveTo and friends); a
// predicate only selects the component and carries the stack's error text as
// the `detail` binding.

import type { LegalContext, LegalStack } from './access.ts';
import { legalCompareOp } from './catalogCompare.ts';
import { specArgs, specError, specTemplate, unknownFrom } from './catalogSupport.ts';
import { goAtoi, goQuote } from './goformat.ts';
import { type Result, resolveEnumPath, resolveIntPath, resolveStackPath } from './path.ts';
import { type LegalPredicate, type LegalPredicateConstructor, newLegalPredicate } from './predicate.ts';
import {
  TEMPLATE_COMPONENT_MISSING,
  TEMPLATE_COMPONENT_MISSING_KEY,
  TEMPLATE_COMPONENT_PRESENT_UNEXPECTED,
  TEMPLATE_MAY_NOT_MOVE_ALL_TO,
  TEMPLATE_MAY_NOT_MOVE_COUNT_TO,
  TEMPLATE_MAY_NOT_MOVE_TO,
  TEMPLATE_MAY_NOT_SWAP_COMPONENTS,
  TEMPLATE_NO_COMPONENT_TO_MOVE,
  TEMPLATE_STACK_COUNT,
  TEMPLATE_STACK_EMPTY,
  TEMPLATE_STACK_NOT_EMPTY,
} from './templates.ts';
import {
  LEGAL_COST_CHEAP,
  LEGAL_COST_EXPENSIVE,
  LEGAL_COST_MODERATE,
  type LegalRead,
  type LegalSpec,
  type PropertyType,
  failT,
  passVerdict,
  unknownVerdict,
} from './types.ts';

/** Go: legal.ComponentPresentAt. */
export function componentPresentAt(stackPath: string, idxField: string): LegalSpec {
  return { name: 'componentPresentAt', args: [stackPath, idxField] };
}

/** Go: legal.ComponentAbsentAt. */
export function componentAbsentAt(stackPath: string, idxField: string): LegalSpec {
  return { name: 'componentAbsentAt', args: [stackPath, idxField] };
}

/** Go: legal.ComponentPresentAtKey. */
export function componentPresentAtKey(stackPath: string, keyField: string): LegalSpec {
  return { name: 'componentPresentAtKey', args: [stackPath, keyField] };
}

/** Go: legal.MayMoveTo. */
export function mayMoveTo(srcPath: string, dstPath: string, idxField: string): LegalSpec {
  return { name: 'mayMoveTo', args: [srcPath, dstPath, idxField] };
}

/** Go: legal.MayMoveToSlot. */
export function mayMoveToSlot(srcPath: string, dstPath: string, sourceIndexField: string, destinationSlotField: string): LegalSpec {
  return { name: 'mayMoveToSlot', args: [srcPath, dstPath, sourceIndexField, destinationSlotField] };
}

/** Go: legal.MayMoveToSameSlot. */
export function mayMoveToSameSlot(srcPath: string, dstPath: string, indexField: string): LegalSpec {
  return mayMoveToSlot(srcPath, dstPath, indexField, indexField);
}

/** Go: legal.MayMoveFirstTo. */
export function mayMoveFirstTo(srcPath: string, dstPath: string): LegalSpec {
  return { name: 'mayMoveFirstTo', args: [srcPath, dstPath] };
}

/** Go: legal.MayMoveFirstToSlot. */
export function mayMoveFirstToSlot(srcPath: string, dstPath: string, destinationSlotField: string): LegalSpec {
  return { name: 'mayMoveFirstToSlot', args: [srcPath, dstPath, destinationSlotField] };
}

/** Go: legal.MayMoveAllTo. */
export function mayMoveAllTo(srcPath: string, dstPath: string): LegalSpec {
  return { name: 'mayMoveAllTo', args: [srcPath, dstPath] };
}

/** Go: legal.MayMoveCountTo. */
export function mayMoveCountTo(srcPath: string, dstPath: string, countPath: string): LegalSpec {
  return { name: 'mayMoveCountTo', args: [srcPath, dstPath, countPath] };
}

/** Go: legal.MayMoveFixedCountTo. */
export function mayMoveFixedCountTo(srcPath: string, dstPath: string, count: number): LegalSpec {
  return { name: 'mayMoveFixedCountTo', args: [srcPath, dstPath, String(count)] };
}

/** Go: legal.MaySwapComponents. */
export function maySwapComponents(stackPath: string, firstIndexField: string, secondIndexField: string): LegalSpec {
  return { name: 'maySwapComponents', args: [stackPath, firstIndexField, secondIndexField] };
}

/** Go: legal.MaySwapComponentsByKey. */
export function maySwapComponentsByKey(stackPath: string, firstKeyField: string, secondKeyField: string): LegalSpec {
  return { name: 'maySwapComponentsByKey', args: [stackPath, firstKeyField, secondKeyField] };
}

/** Go: legal.StackCount. */
export function stackCount(path: string, op: string, n: number): LegalSpec {
  return { name: 'stackCount', args: [path, op, String(n)] };
}

/** Go: legal.StackEmpty. */
export function stackEmpty(path: string): LegalSpec {
  return { name: 'stackEmpty', args: [path] };
}

/** Go: legal.StackNotEmpty. */
export function stackNotEmpty(path: string): LegalSpec {
  return { name: 'stackNotEmpty', args: [path] };
}

function nilStack(path: string) {
  return unknownVerdict(`legal: stack path ${goQuote(path)} resolved to nil`);
}

/** Go: componentPresentAtConstructor / componentAbsentAtConstructor. */
function componentAtIndex(name: string, defaultTemplate: string, wantPresent: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: ${name} requires 2 args (stackPath, idxField), got ${args.length}`);
      }
      const [stackPath, idxField] = args;
      const template = specTemplate(spec, defaultTemplate);
      return newLegalPredicate({
        name,
        clientEvaluable: true,
        args: [...args],
        reads: [
          { path: stackPath, facet: 'occupancy' },
          { path: idxField, facet: 'values' },
        ],
        requiredReadTypes: { [stackPath]: 'TypeStack', [idxField]: 'TypeInt' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['index'] },
        evaluate(ctx) {
          const idx = resolveIntPath(idxField, ctx);
          if (!idx.ok) {
            return unknownFrom(idx);
          }
          const stack = resolveStackPath(stackPath, ctx);
          if (!stack.ok) {
            return unknownFrom(stack);
          }
          if (stack.value === null) {
            return nilStack(stackPath);
          }
          if ((stack.value.componentAt(idx.value) !== null) === wantPresent) {
            return passVerdict();
          }
          return failT(template, { index: idx.value });
        },
      });
    },
  };
}

/** Go: componentPresentAtConstructor. */
export function componentPresentAtConstructor(): LegalPredicateConstructor {
  return componentAtIndex('componentPresentAt', TEMPLATE_COMPONENT_MISSING, true);
}

/** Go: componentAbsentAtConstructor. */
export function componentAbsentAtConstructor(): LegalPredicateConstructor {
  return componentAtIndex('componentAbsentAt', TEMPLATE_COMPONENT_PRESENT_UNEXPECTED, false);
}

/** Go: componentPresentAtKeyConstructor. */
export function componentPresentAtKeyConstructor(): LegalPredicateConstructor {
  return {
    name: 'componentPresentAtKey',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: componentPresentAtKey requires 2 args (stackPath, keyField), got ${args.length}`);
      }
      const [stackPath, keyField] = args;
      const template = specTemplate(spec, TEMPLATE_COMPONENT_MISSING_KEY);
      return newLegalPredicate({
        name: 'componentPresentAtKey',
        clientEvaluable: true,
        args: [...args],
        reads: [
          { path: stackPath, facet: 'occupancy' },
          { path: keyField, facet: 'values' },
        ],
        requiredReadTypes: { [stackPath]: 'TypeStack', [keyField]: 'TypeEnum' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['key'] },
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
            return nilStack(stackPath);
          }
          if (stack.value.componentAt(key.value.value) !== null) {
            return passVerdict();
          }
          return failT(template, { key: key.value.string });
        },
      });
    },
  };
}

/**
 * Go: mayMoveConstructor. `useSlot` adds a destination slot field;
 * `firstSource` takes the source stack's first component instead of an
 * indexed one. Two distinct failures: no component at the source
 * ({index}), and the stack rejecting the move ({detail}).
 */
function mayMove(name: string, useSlot: boolean, firstSource: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      let wantArgs = 3;
      let argDescription = 'srcPath, dstPath, sourceIndexField';
      if (firstSource) {
        wantArgs = 2;
        argDescription = 'srcPath, dstPath';
      }
      if (useSlot) {
        wantArgs++;
        argDescription += ', destinationSlotField';
      }
      if (args.length !== wantArgs) {
        specError(`legal: ${name} requires ${wantArgs} args (${argDescription}), got ${args.length}`);
      }
      const [srcPath, dstPath] = args;
      const sourceIndexField = firstSource ? '' : args[2];
      const destinationSlotField = useSlot ? args[wantArgs - 1] : '';

      const noComponentTemplate = specTemplate(spec, TEMPLATE_NO_COMPONENT_TO_MOVE);
      const mayNotMoveTemplate = specTemplate(spec, TEMPLATE_MAY_NOT_MOVE_TO);
      // One override retargets both branches at one key; only bindings on
      // every emission of that key are guaranteed, and here that is none.
      const emittedBindings: Record<string, readonly string[]> =
        noComponentTemplate === mayNotMoveTemplate
          ? { [noComponentTemplate]: [] }
          : { [noComponentTemplate]: ['index'], [mayNotMoveTemplate]: ['detail'] };

      const reads: LegalRead[] = [
        { path: srcPath, facet: 'occupancy' },
        { path: dstPath, facet: 'values' },
      ];
      const requiredReadTypes: Record<string, PropertyType> = { [srcPath]: 'TypeStack', [dstPath]: 'TypeStack' };
      if (!firstSource) {
        reads.push({ path: sourceIndexField, facet: 'values' });
        requiredReadTypes[sourceIndexField] = 'TypeInt';
      }
      if (useSlot) {
        if (destinationSlotField !== sourceIndexField) {
          reads.push({ path: destinationSlotField, facet: 'values' });
        }
        requiredReadTypes[destinationSlotField] = 'TypeInt';
      }

      return newLegalPredicate({
        name,
        clientEvaluable: true,
        args: [...args],
        reads,
        requiredReadTypes,
        cost: LEGAL_COST_MODERATE,
        emittedTemplates: [noComponentTemplate, mayNotMoveTemplate],
        emittedBindings,
        evaluate(ctx) {
          let sourceIndex = 0;
          if (!firstSource) {
            const idx = resolveIntPath(sourceIndexField, ctx);
            if (!idx.ok) {
              return unknownFrom(idx);
            }
            sourceIndex = idx.value;
          }
          let destinationSlot = sourceIndex;
          if (useSlot) {
            const slot = resolveIntPath(destinationSlotField, ctx);
            if (!slot.ok) {
              return unknownFrom(slot);
            }
            destinationSlot = slot.value;
          }
          const src = resolveStackPath(srcPath, ctx);
          if (!src.ok) {
            return unknownFrom(src);
          }
          const dst = resolveStackPath(dstPath, ctx);
          if (!dst.ok) {
            return unknownFrom(dst);
          }
          if (src.value === null || dst.value === null) {
            return unknownVerdict('legal: source or destination stack path resolved to nil');
          }
          const comp = firstSource ? src.value.first() : src.value.componentAt(sourceIndex);
          if (comp === null) {
            return failT(noComponentTemplate, { index: sourceIndex });
          }
          const moveErr = useSlot ? comp.mayMoveToSlot(dst.value, destinationSlot) : comp.mayMoveTo(dst.value);
          if (moveErr !== null) {
            return failT(mayNotMoveTemplate, { detail: moveErr });
          }
          return passVerdict();
        },
      });
    },
  };
}

/** Go: mayMoveToConstructor. */
export function mayMoveToConstructor(): LegalPredicateConstructor {
  return mayMove('mayMoveTo', false, false);
}

/** Go: mayMoveToSlotConstructor. */
export function mayMoveToSlotConstructor(): LegalPredicateConstructor {
  return mayMove('mayMoveToSlot', true, false);
}

/** Go: mayMoveFirstToConstructor. */
export function mayMoveFirstToConstructor(): LegalPredicateConstructor {
  return mayMove('mayMoveFirstTo', false, true);
}

/** Go: mayMoveFirstToSlotConstructor. */
export function mayMoveFirstToSlotConstructor(): LegalPredicateConstructor {
  return mayMove('mayMoveFirstToSlot', true, true);
}

/** Go: mayMoveAllToConstructor. */
export function mayMoveAllToConstructor(): LegalPredicateConstructor {
  return {
    name: 'mayMoveAllTo',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: mayMoveAllTo requires 2 args (srcPath, dstPath), got ${args.length}`);
      }
      const [srcPath, dstPath] = args;
      const template = specTemplate(spec, TEMPLATE_MAY_NOT_MOVE_ALL_TO);
      return newLegalPredicate({
        name: 'mayMoveAllTo',
        clientEvaluable: false,
        args: [...args],
        reads: [
          { path: srcPath, facet: 'values' },
          { path: dstPath, facet: 'values' },
        ],
        requiredReadTypes: { [srcPath]: 'TypeStack', [dstPath]: 'TypeStack' },
        cost: LEGAL_COST_EXPENSIVE,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['detail'] },
        evaluate(ctx) {
          const src = resolveStackPath(srcPath, ctx);
          if (!src.ok) {
            return unknownFrom(src);
          }
          const dst = resolveStackPath(dstPath, ctx);
          if (!dst.ok) {
            return unknownFrom(dst);
          }
          if (src.value === null || dst.value === null) {
            return unknownVerdict('legal: source or destination stack path resolved to nil');
          }
          const err = src.value.mayMoveAllTo(dst.value);
          if (err !== null) {
            return failT(template, { detail: err });
          }
          return passVerdict();
        },
      });
    },
  };
}

/** Go: newMayMoveCountPredicate. */
function mayMoveCountPredicate(
  name: string,
  args: readonly string[],
  srcPath: string,
  dstPath: string,
  template: string,
  reads: LegalRead[],
  requiredReadTypes: Record<string, PropertyType>,
  resolveCount: (ctx: LegalContext) => Result<number>,
): LegalPredicate {
  return newLegalPredicate({
    name,
    clientEvaluable: false,
    args: [...args],
    reads,
    requiredReadTypes,
    cost: LEGAL_COST_EXPENSIVE,
    emittedTemplates: [template],
    emittedBindings: { [template]: ['detail'] },
    evaluate(ctx) {
      const src = resolveStackPath(srcPath, ctx);
      if (!src.ok) {
        return unknownFrom(src);
      }
      const dst = resolveStackPath(dstPath, ctx);
      if (!dst.ok) {
        return unknownFrom(dst);
      }
      const count = resolveCount(ctx);
      if (!count.ok) {
        return unknownFrom(count);
      }
      if (src.value === null || dst.value === null) {
        return unknownVerdict('legal: source or destination stack path resolved to nil');
      }
      const err = src.value.mayMoveCountTo(dst.value, count.value);
      if (err !== null) {
        return failT(template, { detail: err });
      }
      return passVerdict();
    },
  });
}

/** Go: mayMoveCountToConstructor. */
export function mayMoveCountToConstructor(): LegalPredicateConstructor {
  return {
    name: 'mayMoveCountTo',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: mayMoveCountTo requires 3 args (srcPath, dstPath, countPath), got ${args.length}`);
      }
      const [srcPath, dstPath, countPath] = args;
      const template = specTemplate(spec, TEMPLATE_MAY_NOT_MOVE_COUNT_TO);
      return mayMoveCountPredicate(
        'mayMoveCountTo',
        args,
        srcPath,
        dstPath,
        template,
        [
          { path: srcPath, facet: 'values' },
          { path: dstPath, facet: 'values' },
          { path: countPath, facet: 'values' },
        ],
        { [srcPath]: 'TypeStack', [dstPath]: 'TypeStack', [countPath]: 'TypeInt' },
        ctx => resolveIntPath(countPath, ctx),
      );
    },
  };
}

/** Go: mayMoveFixedCountToConstructor. */
export function mayMoveFixedCountToConstructor(): LegalPredicateConstructor {
  return {
    name: 'mayMoveFixedCountTo',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: mayMoveFixedCountTo requires 3 args (srcPath, dstPath, count), got ${args.length}`);
      }
      const [srcPath, dstPath] = args;
      const parsed = goAtoi(args[2]);
      if (!parsed.ok) {
        specError(`legal: mayMoveFixedCountTo: arg 3 (count) must be an integer: ${parsed.error}`);
      }
      const count = parsed.value;
      if (count < 0) {
        specError(`legal: mayMoveFixedCountTo: count must be non-negative, got ${count}`);
      }
      const template = specTemplate(spec, TEMPLATE_MAY_NOT_MOVE_COUNT_TO);
      return mayMoveCountPredicate(
        'mayMoveFixedCountTo',
        args,
        srcPath,
        dstPath,
        template,
        [
          { path: srcPath, facet: 'values' },
          { path: dstPath, facet: 'values' },
        ],
        { [srcPath]: 'TypeStack', [dstPath]: 'TypeStack' },
        () => ({ ok: true, value: count }),
      );
    },
  };
}

/** Go: maySwapComponentsConstructor. `keyed` reads enum-valued fields instead of int ones. */
function maySwap(name: string, keyed: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: ${name} requires 3 args (stackPath, firstIndexField, secondIndexField), got ${args.length}`);
      }
      const [stackPath, firstField, secondField] = args;
      const template = specTemplate(spec, TEMPLATE_MAY_NOT_SWAP_COMPONENTS);
      const fieldType: PropertyType = keyed ? 'TypeEnum' : 'TypeInt';
      const reads: LegalRead[] = [
        { path: stackPath, facet: 'count' },
        { path: firstField, facet: 'values' },
      ];
      if (secondField !== firstField) {
        reads.push({ path: secondField, facet: 'values' });
      }
      return newLegalPredicate({
        name,
        clientEvaluable: true,
        args: [...args],
        reads,
        requiredReadTypes: { [stackPath]: 'TypeStack', [firstField]: fieldType, [secondField]: fieldType },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['detail'] },
        evaluate(ctx) {
          const stack = resolveStackPath(stackPath, ctx);
          if (!stack.ok) {
            return unknownFrom(stack);
          }
          if (stack.value === null) {
            return nilStack(stackPath);
          }
          let first: number;
          let second: number;
          if (keyed) {
            // Go resolves both before checking either error.
            const firstVal = resolveEnumPath(firstField, ctx);
            const secondVal = resolveEnumPath(secondField, ctx);
            if (!firstVal.ok) {
              return unknownFrom(firstVal);
            }
            if (!secondVal.ok) {
              return unknownFrom(secondVal);
            }
            if (firstVal.value === null || secondVal.value === null) {
              return unknownVerdict('legal: swap enum path resolved to nil');
            }
            first = firstVal.value.value;
            second = secondVal.value.value;
          } else {
            const firstVal = resolveIntPath(firstField, ctx);
            if (!firstVal.ok) {
              return unknownFrom(firstVal);
            }
            const secondVal = resolveIntPath(secondField, ctx);
            if (!secondVal.ok) {
              return unknownFrom(secondVal);
            }
            first = firstVal.value;
            second = secondVal.value;
          }
          const err = (stack.value as LegalStack).maySwapComponents(first, second);
          if (err !== null) {
            return failT(template, { detail: err });
          }
          return passVerdict();
        },
      });
    },
  };
}

/** Go: maySwapComponentsConstructor("maySwapComponents", false). */
export function maySwapComponentsConstructor(): LegalPredicateConstructor {
  return maySwap('maySwapComponents', false);
}

/** Go: maySwapComponentsConstructor("maySwapComponentsByKey", true). */
export function maySwapComponentsByKeyConstructor(): LegalPredicateConstructor {
  return maySwap('maySwapComponentsByKey', true);
}

/** Go: stackCountConstructor. Reads only the count facet. */
export function stackCountConstructor(): LegalPredicateConstructor {
  return {
    name: 'stackCount',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: stackCount requires 3 args (path, op, n), got ${args.length}`);
      }
      const [path, op] = args;
      const cmp = legalCompareOp(op);
      if (cmp === undefined) {
        specError(`legal: stackCount: unknown op ${goQuote(op)} (expected one of ==, !=, <, <=, >, >=)`);
      }
      const parsed = goAtoi(args[2]);
      if (!parsed.ok) {
        specError(`legal: stackCount: arg 3 (n) must be an integer: ${parsed.error}`);
      }
      const n = parsed.value;
      const template = specTemplate(spec, TEMPLATE_STACK_COUNT);
      return newLegalPredicate({
        name: 'stackCount',
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'count' }],
        requiredReadTypes: { [path]: 'TypeStack' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['count', 'op', 'n'] },
        evaluate(ctx) {
          const stack = resolveStackPath(path, ctx);
          if (!stack.ok) {
            return unknownFrom(stack);
          }
          if (stack.value === null) {
            return nilStack(path);
          }
          const count = stack.value.numComponents();
          if (cmp(count, n)) {
            return passVerdict();
          }
          return failT(template, { count, op, n });
        },
      });
    },
  };
}

/** Go: stackEmptinessConstructor. Reads only the nonempty facet. */
function stackEmptiness(name: string, defaultTemplate: string, wantEmpty: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 1) {
        specError(`legal: ${name} requires 1 arg (path), got ${args.length}`);
      }
      const path = args[0];
      const template = specTemplate(spec, defaultTemplate);
      return newLegalPredicate({
        name,
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'nonempty' }],
        requiredReadTypes: { [path]: 'TypeStack' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: [] },
        evaluate(ctx) {
          const stack = resolveStackPath(path, ctx);
          if (!stack.ok) {
            return unknownFrom(stack);
          }
          if (stack.value === null) {
            return nilStack(path);
          }
          if ((stack.value.numComponents() === 0) === wantEmpty) {
            return passVerdict();
          }
          return failT(template);
        },
      });
    },
  };
}

/** Go: stackEmptyConstructor. */
export function stackEmptyConstructor(): LegalPredicateConstructor {
  return stackEmptiness('stackEmpty', TEMPLATE_STACK_EMPTY, true);
}

/** Go: stackNotEmptyConstructor. */
export function stackNotEmptyConstructor(): LegalPredicateConstructor {
  return stackEmptiness('stackNotEmpty', TEMPLATE_STACK_NOT_EMPTY, false);
}
