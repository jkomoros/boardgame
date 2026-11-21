// Go: legal/catalog_compare.go. The general-purpose relations: propAtLeast,
// propCompare, playerBool, propEquals, propNotEquals.

import type { LegalChest, LegalEnumValue } from './access.ts';
import { specArgs, specError, specTemplate, unknownFrom } from './catalogSupport.ts';
import { goAtoi, goQuote } from './goformat.ts';
import { resolveBoolPath, resolveIntPath, resolveLegalPath } from './path.ts';
import { type LegalPredicateConstructor, newLegalPredicate } from './predicate.ts';
import { TEMPLATE_PLAYER_BOOL, TEMPLATE_PROP_AT_LEAST, TEMPLATE_PROP_COMPARE, TEMPLATE_PROP_EQUALS, TEMPLATE_PROP_NOT_EQUALS } from './templates.ts';
import {
  ADMIN_PLAYER_INDEX,
  LEGAL_COST_CHEAP,
  LEGAL_COST_TRIVIAL,
  type LegalSpec,
  OBSERVER_PLAYER_INDEX,
  failT,
  passVerdict,
  unknownVerdict,
} from './types.ts';

/** Go: legalCompareOps. The closed set of comparison operators. */
export const LEGAL_COMPARE_OPS: Readonly<Record<string, (value: number, n: number) => boolean>> = {
  '==': (value, n) => value === n,
  '!=': (value, n) => value !== n,
  '<': (value, n) => value < n,
  '<=': (value, n) => value <= n,
  '>': (value, n) => value > n,
  '>=': (value, n) => value >= n,
};

/** Looks up a comparison operator; undefined for anything outside the closed set. */
export function legalCompareOp(op: string): ((value: number, n: number) => boolean) | undefined {
  return Object.prototype.hasOwnProperty.call(LEGAL_COMPARE_OPS, op) ? LEGAL_COMPARE_OPS[op] : undefined;
}

/** Go: legal.PropAtLeast. */
export function propAtLeast(path: string, n: number): LegalSpec {
  return { name: 'propAtLeast', args: [path, String(n)] };
}

/** Go: legal.PropCompare. `op` is one of ==, !=, <, <=, >, >=. */
export function propCompare(path: string, op: string, n: number): LegalSpec {
  return { name: 'propCompare', args: [path, op, String(n)] };
}

/** Go: legal.PlayerBool. Requires the current player's bool `prop` to be true. */
export function playerBool(prop: string): LegalSpec {
  return { name: 'playerBool', args: [prop] };
}

/** Go: legal.PlayerBoolIs. */
export function playerBoolIs(prop: string, want: boolean): LegalSpec {
  return { name: 'playerBool', args: [prop, String(want)] };
}

/** Go: legal.PropEquals. `value` is matched by the property's own type. */
export function propEquals(path: string, value: string): LegalSpec {
  return { name: 'propEquals', args: [path, value] };
}

/** Go: legal.PropNotEquals. */
export function propNotEquals(path: string, value: string): LegalSpec {
  return { name: 'propNotEquals', args: [path, value] };
}

/** Go: playerBoolWant. Returns the bool, or the error fragment Go formats. */
export function playerBoolWant(arg: string): { ok: true; value: boolean } | { ok: false; error: string } {
  if (arg === 'true') {
    return { ok: true, value: true };
  }
  if (arg === 'false') {
    return { ok: true, value: false };
  }
  return { ok: false, error: `must be "true" or "false", got ${goQuote(arg)}` };
}

/** Go: propEqualsPlayerIndexString. */
function playerIndexString(p: number): string {
  if (p === OBSERVER_PLAYER_INDEX) {
    return 'observer';
  }
  if (p === ADMIN_PLAYER_INDEX) {
    return 'admin';
  }
  return String(p);
}

/** Go: valueMatchesAnyEnumInChest. */
function valueMatchesAnyEnumInChest(value: string, chest: LegalChest): boolean {
  return chest.enums().some(e => e.valueFromString(value) !== null);
}

/** Go: propEqualsFamilyConstructor. */
function propEqualsFamily(name: string, defaultTemplate: string, negate: boolean): LegalPredicateConstructor {
  return {
    name,
    construct(spec, chest) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: ${name} requires 2 args (path, value), got ${args.length}`);
      }
      const [path, value] = args;
      const template = specTemplate(spec, defaultTemplate);

      // The value's possible readings are settled now; which one applies is
      // decided per evaluation by the property's actual type.
      const intWant = goAtoi(value);
      const boolWant = value === 'true' ? true : value === 'false' ? false : null;

      // A constructor-time typo guard, only possible when a chest is at hand.
      if (
        chest &&
        !intWant.ok &&
        boolWant === null &&
        value !== 'observer' &&
        value !== 'admin' &&
        !valueMatchesAnyEnumInChest(value, chest)
      ) {
        specError(
          `legal: ${name}: value ${goQuote(value)} matches no int/bool/playerindex-special and no enum value name in the chest — likely a typo`,
        );
      }

      return newLegalPredicate({
        name,
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'values' }],
        allowedReadTypes: { [path]: ['TypeInt', 'TypeBool', 'TypeEnum', 'TypePlayerIndex'] },
        cost: LEGAL_COST_TRIVIAL,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['value', 'want'] },
        evaluate(ctx) {
          const resolved = resolveLegalPath(path, ctx);
          if (!resolved.ok) {
            return unknownFrom(resolved);
          }
          const { value: val, type } = resolved.value;
          let match: boolean;
          let actual: string;
          switch (type) {
            case 'TypeInt':
              if (!intWant.ok) {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} is an int property but value ${goQuote(value)} does not parse as an int`,
                );
              }
              if (typeof val !== 'number') {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} resolved to an int-typed property but its value was not an int (${typeof val})`,
                );
              }
              actual = String(val);
              match = val === intWant.value;
              break;
            case 'TypeBool':
              if (boolWant === null) {
                return unknownVerdict(`legal: ${name}: value ${goQuote(value)} is not "true" or "false"`);
              }
              if (typeof val !== 'boolean') {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} resolved to a bool-typed property but its value was not a bool (${typeof val})`,
                );
              }
              actual = String(val);
              match = val === boolWant;
              break;
            case 'TypeEnum': {
              const v = (val as LegalEnumValue | null) ?? null;
              if (v === null) {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} resolved to an enum-typed property with no value`,
                );
              }
              const key = v.enum.valueFromString(value);
              if (key === null) {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} is an enum property (enum ${goQuote(v.enum.name)}) but ${goQuote(value)} is not a valid value name`,
                );
              }
              actual = v.string;
              match = v.value === key;
              break;
            }
            case 'TypePlayerIndex': {
              if (typeof val !== 'number') {
                return unknownVerdict(
                  `legal: ${name}: path ${goQuote(path)} resolved to a PlayerIndex-typed property but its value was not a PlayerIndex (${typeof val})`,
                );
              }
              let want: number;
              if (value === 'observer') {
                want = OBSERVER_PLAYER_INDEX;
              } else if (value === 'admin') {
                want = ADMIN_PLAYER_INDEX;
              } else if (intWant.ok) {
                want = intWant.value;
              } else {
                return unknownVerdict(
                  `legal: ${name}: value ${goQuote(value)} is not a valid PlayerIndex (want an integer, "observer", or "admin")`,
                );
              }
              actual = playerIndexString(val);
              match = val === want;
              break;
            }
            default:
              return unknownVerdict(
                `legal: ${name}: path ${goQuote(path)} resolved to property type ${type}, which typed equality does not support (want int, bool, enum, or PlayerIndex)`,
              );
          }
          if (negate) {
            match = !match;
          }
          if (match) {
            return passVerdict();
          }
          return failT(template, { value: actual, want: value });
        },
      });
    },
  };
}

/** Go: propAtLeastConstructor. */
export function propAtLeastConstructor(): LegalPredicateConstructor {
  return {
    name: 'propAtLeast',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 2) {
        specError(`legal: propAtLeast requires 2 args (path, n), got ${args.length}`);
      }
      const path = args[0];
      const parsed = goAtoi(args[1]);
      if (!parsed.ok) {
        specError(`legal: propAtLeast: arg 2 (n) must be an integer: ${parsed.error}`);
      }
      const n = parsed.value;
      const template = specTemplate(spec, TEMPLATE_PROP_AT_LEAST);
      return newLegalPredicate({
        name: 'propAtLeast',
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'values' }],
        requiredReadTypes: { [path]: 'TypeInt' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['value', 'min'] },
        evaluate(ctx) {
          const value = resolveIntPath(path, ctx);
          if (!value.ok) {
            return unknownFrom(value);
          }
          if (value.value >= n) {
            return passVerdict();
          }
          return failT(template, { value: value.value, min: n });
        },
      });
    },
  };
}

/** Go: propCompareConstructor. */
export function propCompareConstructor(): LegalPredicateConstructor {
  return {
    name: 'propCompare',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 3) {
        specError(`legal: propCompare requires 3 args (path, op, n), got ${args.length}`);
      }
      const [path, op] = args;
      const cmp = legalCompareOp(op);
      if (cmp === undefined) {
        specError(`legal: propCompare: unknown op ${goQuote(op)} (expected one of ==, !=, <, <=, >, >=)`);
      }
      const parsed = goAtoi(args[2]);
      if (!parsed.ok) {
        specError(`legal: propCompare: arg 3 (n) must be an integer: ${parsed.error}`);
      }
      const n = parsed.value;
      const template = specTemplate(spec, TEMPLATE_PROP_COMPARE);
      return newLegalPredicate({
        name: 'propCompare',
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'values' }],
        requiredReadTypes: { [path]: 'TypeInt' },
        cost: LEGAL_COST_CHEAP,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['value', 'op', 'n'] },
        evaluate(ctx) {
          const value = resolveIntPath(path, ctx);
          if (!value.ok) {
            return unknownFrom(value);
          }
          if (cmp(value.value, n)) {
            return passVerdict();
          }
          return failT(template, { value: value.value, op, n });
        },
      });
    },
  };
}

/** Go: playerBoolConstructor. */
export function playerBoolConstructor(): LegalPredicateConstructor {
  return {
    name: 'playerBool',
    construct(spec) {
      const args = specArgs(spec);
      if (args.length !== 1 && args.length !== 2) {
        specError(`legal: playerBool requires 1 or 2 args (prop, optional want), got ${args.length}`);
      }
      const prop = args[0];
      let want = true;
      if (args.length === 2) {
        const parsed = playerBoolWant(args[1]);
        if (!parsed.ok) {
          specError(`legal: playerBool: arg 2 (want) ${parsed.error}`);
        }
        want = parsed.value;
      }
      const path = 'player.' + prop;
      const template = specTemplate(spec, TEMPLATE_PLAYER_BOOL);
      return newLegalPredicate({
        name: 'playerBool',
        clientEvaluable: true,
        args: [...args],
        reads: [{ path, facet: 'values' }],
        requiredReadTypes: { [path]: 'TypeBool' },
        cost: LEGAL_COST_TRIVIAL,
        emittedTemplates: [template],
        emittedBindings: { [template]: ['prop', 'want'] },
        evaluate(ctx) {
          const value = resolveBoolPath(path, ctx);
          if (!value.ok) {
            return unknownFrom(value);
          }
          if (value.value === want) {
            return passVerdict();
          }
          return failT(template, { prop, want: String(want) });
        },
      });
    },
  };
}

/** Go: propEqualsConstructor. */
export function propEqualsConstructor(): LegalPredicateConstructor {
  return propEqualsFamily('propEquals', TEMPLATE_PROP_EQUALS, false);
}

/** Go: propNotEqualsConstructor. */
export function propNotEqualsConstructor(): LegalPredicateConstructor {
  return propEqualsFamily('propNotEquals', TEMPLATE_PROP_NOT_EQUALS, true);
}
