// The public surface of the legal evaluator. See README.md.

export * from './types.ts';
export * from './access.ts';
export {
  type LegalPathKind,
  type ParsedLegalPath,
  type ResolvedLegalProp,
  type Result,
  legalPathReadsMove,
  parseLegalPath,
  resolveBoolPath,
  resolveEnumPath,
  resolveIntPath,
  resolveLegalPath,
  resolvePlayerIndexPath,
  resolveStackPath,
  validateLegalPath,
  validateLegalPathTypes,
} from './path.ts';
export {
  legalPanic,
  isLegalPanic,
  playerIndexEnsureValid,
  playerIndexEquivalent,
  playerIndexNext,
  playerIndexValid,
} from './playerIndex.ts';
export type { LegalPanic } from './playerIndex.ts';
export * from './predicate.ts';
export * from './templates.ts';
export * from './message.ts';
export * from './catalog.ts';
export * from './catalogCompare.ts';
export * from './catalogStack.ts';
export * from './catalogPlayers.ts';
export * from './catalogPurpose.ts';
export * from './catalogFramework.ts';
export * from './plan.ts';
export * from './evaluable.ts';
export { goAtoi, goQuote } from './goformat.ts';
