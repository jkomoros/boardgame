import type { BoundMoveAction } from './action.js';
import type { TargetAction, TargetKey } from './target-action.js';

export interface TargetListChoice<
  Key extends TargetKey,
  MoveName extends string = string,
  Input extends object = object,
> {
  readonly key: Key;
  readonly label: string;
  readonly action: BoundMoveAction<MoveName, Input>;
}

/** One validated, exact-key binding for an accessible list of move targets. */
export interface TargetListBinding<
  Key extends TargetKey,
  MoveName extends string = string,
  Input extends object = object,
> {
  readonly target: TargetAction<Key, MoveName, Input>;
  readonly choices: readonly TargetListChoice<Key, MoveName, Input>[];
}

const bindings = new WeakSet<object>();

/**
 * Adds game-owned labels to a typed target collection without separating the
 * labels from the exact keys or actions they describe.
 */
export function targetList<
  Key extends TargetKey,
  MoveName extends string,
  Input extends object,
>(
  target: TargetAction<Key, MoveName, Input>,
  labelFor: (key: Key, index: number) => string,
): TargetListBinding<Key, MoveName, Input> {
  if (typeof target !== 'object' || target === null
    || !Array.isArray(target.candidates)
    || typeof target.get !== 'function'
    || typeof target.subscribe !== 'function'
    || typeof target.ensurePreview !== 'function') {
    throw new Error('targetList: target must come from move(...).targets(...)');
  }
  if (typeof labelFor !== 'function') throw new Error('targetList: labelFor must be a function');
  const choices = target.candidates.map((candidate, index) => {
    let label: string;
    try {
      label = labelFor(candidate.key, index);
    } catch (error) {
      const detail = error instanceof Error ? `: ${error.message}` : '';
      throw new Error(`targetList: labelFor failed for ${JSON.stringify(candidate.key)} at index ${index}${detail}`);
    }
    if (typeof label !== 'string' || !label.trim()) {
      throw new Error(`targetList: labelFor must return a non-empty string for ${JSON.stringify(candidate.key)} at index ${index}`);
    }
    if (label.length > 200) {
      throw new Error(`targetList: label for ${JSON.stringify(candidate.key)} exceeds 200 characters`);
    }
    return Object.freeze({ key: candidate.key, label: label.trim(), action: candidate.action });
  });
  const binding = Object.freeze({ target, choices: Object.freeze(choices) });
  bindings.add(binding);
  return binding;
}

export function isTargetListBinding(value: unknown): value is TargetListBinding<TargetKey> {
  return typeof value === 'object' && value !== null && bindings.has(value);
}
