import type { AnimationTimingPolicy } from './timing.ts';

export interface MotionTransferDeclaration {
  readonly key: string;
  readonly subjectId: string;
  readonly source: string;
  readonly carrier: string;
  readonly durationMs?: number;
  readonly timing?: AnimationTimingPolicy;
}

export interface CompiledMotionTransferDeclaration {
  readonly key: string;
  readonly subjectId: string;
  readonly source: string;
  readonly carrier: string;
  readonly durationMs: number;
  readonly timing: AnimationTimingPolicy;
}

function nonempty(value: unknown, label: string): string {
  if (typeof value !== 'string' || !value.trim()) {
    throw new Error(`motion transfer ${label} must be a nonempty string`);
  }
  return value.trim();
}

/** Copy and atomically validate one transition's ordered presentation intent. */
export function compileMotionTransferDeclarations(
  declarations: readonly MotionTransferDeclaration[],
): readonly CompiledMotionTransferDeclaration[] {
  if (!Array.isArray(declarations)) {
    throw new Error('motion transfer declarations must be an array');
  }
  const keys = new Set<string>();
  const subjects = new Set<string>();
  const carriers = new Set<string>();
  const compiled = declarations.map(declaration => {
    const key = nonempty(declaration?.key, 'key');
    const subjectId = nonempty(declaration?.subjectId, 'subjectId');
    const source = nonempty(declaration?.source, 'source');
    const carrier = nonempty(declaration?.carrier, 'carrier');
    const durationMs = declaration.durationMs ?? 500;
    if (!Number.isFinite(durationMs) || durationMs < 0) {
      throw new Error('motion transfer durationMs must be finite and nonnegative');
    }
    if (keys.has(key)) throw new Error(`duplicate motion transfer key ${key}`);
    if (subjects.has(subjectId)) throw new Error(`duplicate motion transfer subject ${subjectId}`);
    if (carriers.has(carrier)) throw new Error(`duplicate motion transfer carrier ${carrier}`);
    keys.add(key);
    subjects.add(subjectId);
    carriers.add(carrier);
    return Object.freeze({
      key,
      subjectId,
      source,
      carrier,
      durationMs,
      timing: declaration.timing ?? 'version',
    });
  });
  return Object.freeze(compiled);
}

export function motionTransfer(
  declaration: MotionTransferDeclaration,
): CompiledMotionTransferDeclaration {
  return compileMotionTransferDeclarations([declaration])[0];
}
