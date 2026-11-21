// The default predicate registry: Go's legal.DefaultConstructors() plus
// moves.FrameworkConstructors() ("inProgression"), which is what package
// moves registers as the universal catalog at init.

import { playerBoolConstructor, propAtLeastConstructor, propCompareConstructor, propEqualsConstructor, propNotEqualsConstructor } from './catalogCompare.ts';
import { inPhaseConstructor, inProgressionConstructor, stackConstraintsConstructor } from './catalogFramework.ts';
import {
  allActivePlayersConstructor,
  playerBoolAtConstructor,
  proposerIsCurrentPlayerConstructor,
  proposerIsPlayerFromMoveConstructor,
} from './catalogPlayers.ts';
import {
  componentPropEqualsConstructor,
  componentPropEqualsCurrentPlayerConstructor,
  componentPropNotEqualsConstructor,
  componentPropsEqualConstructor,
  componentPropsNotEqualConstructor,
  revealableCardAtConstructor,
} from './catalogPurpose.ts';
import {
  componentAbsentAtConstructor,
  componentPresentAtConstructor,
  componentPresentAtKeyConstructor,
  mayMoveAllToConstructor,
  mayMoveCountToConstructor,
  mayMoveFirstToConstructor,
  mayMoveFirstToSlotConstructor,
  mayMoveFixedCountToConstructor,
  mayMoveToConstructor,
  mayMoveToSlotConstructor,
  maySwapComponentsByKeyConstructor,
  maySwapComponentsConstructor,
  stackCountConstructor,
  stackEmptyConstructor,
  stackNotEmptyConstructor,
} from './catalogStack.ts';
import { type LegalPredicateConstructor, type LegalRegistry, legalRegistry } from './predicate.ts';

/**
 * Go: legal.DefaultConstructors(), in Go's order, followed by
 * moves.FrameworkConstructors(). `any` is not here: it is resolved by
 * resolveLegalSpecs itself and cannot be shadowed.
 */
export function defaultLegalConstructors(): LegalPredicateConstructor[] {
  return [
    propAtLeastConstructor(),
    propCompareConstructor(),
    playerBoolConstructor(),
    playerBoolAtConstructor(),
    componentPresentAtConstructor(),
    componentAbsentAtConstructor(),
    componentPresentAtKeyConstructor(),
    mayMoveToConstructor(),
    mayMoveToSlotConstructor(),
    mayMoveFirstToConstructor(),
    mayMoveFirstToSlotConstructor(),
    mayMoveAllToConstructor(),
    mayMoveCountToConstructor(),
    mayMoveFixedCountToConstructor(),
    maySwapComponentsConstructor(),
    maySwapComponentsByKeyConstructor(),
    allActivePlayersConstructor(),
    proposerIsCurrentPlayerConstructor(),
    proposerIsPlayerFromMoveConstructor(),
    revealableCardAtConstructor(),
    componentPropEqualsCurrentPlayerConstructor(),
    inPhaseConstructor(),
    stackConstraintsConstructor(),
    stackCountConstructor(),
    stackEmptyConstructor(),
    stackNotEmptyConstructor(),
    propEqualsConstructor(),
    propNotEqualsConstructor(),
    componentPropEqualsConstructor(),
    componentPropNotEqualsConstructor(),
    componentPropsEqualConstructor(),
    componentPropsNotEqualConstructor(),
    inProgressionConstructor(),
  ];
}

/**
 * The registry a game resolves against: the universal catalog overlaid with
 * the game's own constructors (Go: ConfigurePredicateConstructors).
 */
export function defaultLegalRegistry(...gameConstructors: readonly LegalPredicateConstructor[]): LegalRegistry {
  return legalRegistry(defaultLegalConstructors(), gameConstructors);
}
