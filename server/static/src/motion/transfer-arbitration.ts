export interface TransferOwnershipFact {
  readonly key: string;
  readonly carrierKind: 'external' | 'stack';
  readonly subjectMatchesCarrier: boolean;
  readonly carrierResolvesExactly: boolean;
  readonly beforeSightings: number;
  readonly afterSightings: number;
}

export type TransferOwnershipDecision = Readonly<{
  key: string;
  disposition: 'automatic' | 'explicit' | 'conflict';
  reason?: 'retained' | 'ambiguous' | 'identity-mismatch';
}>;

/** Decide segment ownership without consulting geometry or mutable DOM. */
export function partitionMotionTransferOwnership(
  facts: readonly TransferOwnershipFact[],
): readonly TransferOwnershipDecision[] {
  return Object.freeze(facts.map(fact => {
    if (fact.carrierKind === 'external') {
      return Object.freeze({ key: fact.key, disposition: 'explicit' as const });
    }
    if (!fact.subjectMatchesCarrier || !fact.carrierResolvesExactly) {
      return Object.freeze({ key: fact.key, disposition: 'conflict' as const, reason: 'identity-mismatch' as const });
    }
    if (fact.beforeSightings > 0) {
      return Object.freeze({ key: fact.key, disposition: 'conflict' as const, reason: 'retained' as const });
    }
    if (fact.afterSightings !== 1) {
      return Object.freeze({ key: fact.key, disposition: 'conflict' as const, reason: 'ambiguous' as const });
    }
    return Object.freeze({ key: fact.key, disposition: 'automatic' as const });
  }));
}
