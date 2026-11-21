// Fail-closed evaluation, the `any` compositor, spec resolution rules,
// player-index arithmetic and the Go formatting helpers, on hand-written
// fixtures. These are the corners a recorded Go verdict does not reach: a
// predicate that throws, one that returns nonsense, a constructor that tries
// to smuggle in a nested `any`.
import assert from 'node:assert/strict';
import { test } from 'node:test';
import type { LegalContext } from './access.ts';
import { defaultLegalConstructors, defaultLegalRegistry } from './catalog.ts';
import { propAtLeast } from './catalogCompare.ts';
import { any, playerBoolAt, proposer, proposerIsPlayerFromMove } from './catalogPlayers.ts';
import { stackEmpty } from './catalogStack.ts';
import { goAtoi, goQuote, goStringSlice } from './goformat.ts';
import { legalErrorf, legalTemplatePlaceholders, renderLegalMessage, renderLegalVerdict } from './message.ts';
import { playerIndexEnsureValid, playerIndexEquivalent, playerIndexNext, playerIndexValid, legalPanic } from './playerIndex.ts';
import type { LegalPanic } from './playerIndex.ts';
import {
  type LegalPredicate,
  type LegalPredicateConstructor,
  evalLegalPredicate,
  legalPredicateSerializable,
  legalRegistry,
  newLegalPredicate,
  resolveLegalSpecs,
  isLegalSpecError,
} from './predicate.ts';
import type { LegalSpecError } from './predicate.ts';
import { defaultLegalTemplates, mergeLegalTemplates } from './templates.ts';
import * as fx from './testing/fixtures.ts';
import { ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX, LEGAL_CATALOG_VERSION, type LegalVerdict, OBSERVER_PLAYER_INDEX, failT, passVerdict, unknownVerdict } from './types.ts';

function fixed(name: string, verdict: LegalVerdict | (() => LegalVerdict), fields: Partial<LegalPredicate> = {}): LegalPredicate {
  return newLegalPredicate({
    name,
    reads: [],
    cost: 0,
    evaluate: typeof verdict === 'function' ? verdict : () => verdict,
    ...fields,
  });
}

function fixedConstructor(name: string, verdict: LegalVerdict): LegalPredicateConstructor {
  return { name, construct: () => fixed(name, verdict) };
}

const noMove: LegalContext = { state: fx.state(), move: null, proposer: 0 };
const withMove: LegalContext = { state: fx.state(), move: fx.move({}), proposer: 0 };

test('evalLegalPredicate fails closed on a missing predicate or evaluate function', () => {
  assert.deepEqual(evalLegalPredicate(null, noMove), { outcome: 'unknown', reason: 'predicate was nil' });
  const broken = { ...fixed('broken', passVerdict()), evaluate: undefined } as unknown as LegalPredicate;
  assert.deepEqual(evalLegalPredicate(broken, noMove), { outcome: 'unknown', reason: 'predicate "broken" has no Evaluate function' });
});

test('a verdict that is not pass, fail or unknown is unknown, never legal', () => {
  for (const bad of [{}, { outcome: '' }, { outcome: 'legal' }, null, undefined]) {
    assert.deepEqual(evalLegalPredicate(fixed('p', bad as LegalVerdict), withMove), {
      outcome: 'unknown',
      reason: 'predicate returned invalid verdict',
    });
  }
});

test('a throwing predicate is unknown, with the reason Go gives a panic', () => {
  const thrower = (reads: LegalPredicate['reads']) =>
    fixed(
      'thrower',
      () => {
        throw legalPanic('runtime error: index out of range [-6]');
      },
      { reads },
    );
  assert.deepEqual(evalLegalPredicate(thrower([]), withMove), {
    outcome: 'unknown',
    reason: 'predicate "thrower" panicked: runtime error: index out of range [-6]',
  });
  // With no move and no declared move read, Go assumes an undeclared touch of the move.
  assert.deepEqual(evalLegalPredicate(thrower([{ path: 'game.X', facet: 'values' }]), noMove), {
    outcome: 'unknown',
    reason: 'undeclared move read',
  });
  // A declared move read (either kind) means the panic is reported as itself.
  for (const p of ['move.X', 'players[move.Who].X']) {
    assert.match(evalLegalPredicate(thrower([{ path: p, facet: 'values' }]), noMove).reason ?? '', /^predicate "thrower" panicked/);
  }
});

test('admin bypass passes without evaluating, and only for the admin proposer', () => {
  let calls = 0;
  const p = fixed(
    'gate',
    () => {
      calls++;
      return failT('gate.no');
    },
    { adminPolicy: 'bypass' },
  );
  assert.deepEqual(evalLegalPredicate(p, { ...withMove, proposer: ADMIN_PLAYER_INDEX }), { outcome: 'pass' });
  assert.equal(calls, 0);
  assert.equal(evalLegalPredicate(p, { ...withMove, proposer: 0 }).outcome, 'fail');
  assert.equal(evalLegalPredicate(p, { ...withMove, proposer: OBSERVER_PLAYER_INDEX }).outcome, 'fail');
  assert.equal(evalLegalPredicate(p, { ...withMove, proposer: ANY_PLAYER_INDEX }).outcome, 'fail');
});

test('resolveLegalSpecs copies the spec\'s admin policy onto the predicate, but not onto an any', () => {
  const registry = defaultLegalRegistry();
  const [bypassing] = resolveLegalSpecs([proposerIsPlayerFromMove('Who')], registry);
  assert.equal(bypassing.adminPolicy, 'bypass');
  const [plain] = resolveLegalSpecs([{ name: 'proposerIsPlayerFromMove', args: ['Who'] }], registry);
  assert.equal(plain.adminPolicy, '');
  const [compositor] = resolveLegalSpecs([{ ...any(stackEmpty('game.A'), stackEmpty('game.B')), adminPolicy: 'bypass' }], registry);
  assert.equal(compositor.adminPolicy, '');
});

test('any is Kleene: a pass wins, then an unknown, then fail', () => {
  const P = passVerdict();
  const F = failT('child.failed', { n: 1 });
  const U = unknownVerdict('child could not tell');
  const table: [LegalVerdict[], LegalVerdict][] = [
    [[P, P], { outcome: 'pass' }],
    [[F, P], { outcome: 'pass' }],
    [[U, P], { outcome: 'pass' }],
    [[P, U], { outcome: 'pass' }],
    [[F, F], { outcome: 'fail', message: { template: 'legal.any_failed' } }],
    [[F, U], { outcome: 'unknown', message: { template: 'legal.any_failed' }, reason: 'sub-predicate "s1" of "any" was unknown' }],
    [[U, F], { outcome: 'unknown', message: { template: 'legal.any_failed' }, reason: 'sub-predicate "s0" of "any" was unknown' }],
    [[U, U, F], { outcome: 'unknown', message: { template: 'legal.any_failed' }, reason: 'sub-predicate "s0" of "any" was unknown' }],
    [[F, F, F, P], { outcome: 'pass' }],
  ];
  for (const [subs, want] of table) {
    const registry = legalRegistry(subs.map((v, i) => fixedConstructor('s' + i, v)));
    const [compositor] = resolveLegalSpecs([any(...subs.map((_, i) => ({ name: 's' + i })))], registry);
    assert.deepEqual(evalLegalPredicate(compositor, withMove), want, subs.map(s => s.outcome).join(','));
  }
});

test('any stops at the first pass, carries no child bindings, and honours a message override', () => {
  const order: string[] = [];
  const tracking = (name: string, verdict: LegalVerdict): LegalPredicateConstructor => ({
    name,
    construct: () =>
      fixed(name, () => {
        order.push(name);
        return verdict;
      }),
  });
  const registry = legalRegistry([tracking('a', failT('a.no', { secret: 7 })), tracking('b', passVerdict()), tracking('c', passVerdict())]);
  const [compositor] = resolveLegalSpecs([any({ name: 'a' }, { name: 'b' }, { name: 'c' })], registry);
  assert.deepEqual(evalLegalPredicate(compositor, withMove), { outcome: 'pass' });
  assert.deepEqual(order, ['a', 'b']);

  const [overridden] = resolveLegalSpecs([{ ...any({ name: 'a' }, { name: 'a' }), message: 'mine.neither' }], registry);
  assert.deepEqual(evalLegalPredicate(overridden, withMove), { outcome: 'fail', message: { template: 'mine.neither' } });
  assert.deepEqual(overridden.emittedTemplates, ['mine.neither']);
  assert.deepEqual(overridden.emittedBindings, { 'mine.neither': [] });
});

test('a throwing child of any is one unknown child, not a thrown any', () => {
  const registry = legalRegistry([
    {
      name: 'boom',
      construct: () =>
        fixed('boom', () => {
          throw new Error('kaboom');
        }),
    },
    fixedConstructor('no', failT('no')),
    fixedConstructor('yes', passVerdict()),
  ]);
  const [failing] = resolveLegalSpecs([any({ name: 'boom' }, { name: 'no' })], registry);
  assert.deepEqual(evalLegalPredicate(failing, withMove), {
    outcome: 'unknown',
    message: { template: 'legal.any_failed' },
    reason: 'sub-predicate "boom" of "any" was unknown',
  });
  const [passing] = resolveLegalSpecs([any({ name: 'boom' }, { name: 'yes' })], registry);
  assert.equal(evalLegalPredicate(passing, withMove).outcome, 'pass');
});

test('any unions its children\'s reads and takes the highest cost', () => {
  const [compositor] = resolveLegalSpecs(
    [any(propAtLeast('player.Score', 1), stackEmpty('game.Deck'), propAtLeast('player.Score', 9), playerBoolAt(proposer(), 'Ready', true))],
    defaultLegalRegistry(),
  );
  assert.deepEqual(compositor.reads, [
    { path: 'player.Score', facet: 'values' },
    { path: 'game.Deck', facet: 'nonempty' },
    { path: 'proposer.Ready', facet: 'values' },
  ]);
  assert.equal(compositor.cost, 1);
  assert.deepEqual(compositor.requiredReadTypes, { 'player.Score': 'TypeInt', 'game.Deck': 'TypeStack', 'proposer.Ready': 'TypeBool' });
  assert.equal(compositor.sub.length, 4);
  assert.equal(compositor.clientEvaluable, false);
  assert.equal(legalPredicateSerializable(compositor), true);
});

test('any is depth one however it is reached', () => {
  const registry = defaultLegalRegistry(
    // A constructor that resolves an `any` of its own.
    { name: 'wrapsAnAny', construct: (_spec, _chest, resolve) => resolve(any(stackEmpty('game.A'), stackEmpty('game.B'))) },
    // A constructor that hand-builds a nested `any` without calling resolve.
    {
      name: 'handBuilt',
      construct: () => fixed('any', passVerdict(), { sub: [fixed('any', passVerdict(), { sub: [fixed('leaf', passVerdict())] })] }),
    },
  );
  const nested = /nested "any" compositor is not allowed/;
  assert.throws(() => resolveLegalSpecs([any(stackEmpty('game.A'), any(stackEmpty('game.B'), stackEmpty('game.C')))], registry), nested);
  assert.throws(() => resolveLegalSpecs([any(stackEmpty('game.A'), { name: 'wrapsAnAny' })], registry), nested);
  assert.throws(() => resolveLegalSpecs([{ name: 'handBuilt' }], registry), nested);
  // Outside an any, a constructor may resolve one.
  assert.equal(resolveLegalSpecs([{ name: 'wrapsAnAny' }], registry)[0].name, 'any');
  assert.throws(() => resolveLegalSpecs([{ name: 'any' }], registry), /requires at least 2 sub-specs, got 0/);
});

test('any cannot be shadowed, and a game constructor overrides a catalog one by name', () => {
  const registry = defaultLegalRegistry(fixedConstructor('any', failT('shadow')), fixedConstructor('stackEmpty', failT('mine')));
  const [compositor, overridden] = resolveLegalSpecs([any(propAtLeast('player.A', 1), propAtLeast('player.B', 1)), stackEmpty('game.S')], registry);
  assert.equal(compositor.sub.length, 2);
  assert.deepEqual(evalLegalPredicate(overridden, withMove), { outcome: 'fail', message: { template: 'mine' } });
  assert.throws(
    () => resolveLegalSpecs([{ name: 'nope' }], registry),
    (err: unknown) => isLegalSpecError(err) && err.message === 'boardgame: legal spec "nope": unknown predicate name "nope"',
  );
});

test('the catalog is the Go catalog: 32 predicates plus inProgression, each named once', () => {
  const names = defaultLegalConstructors().map(c => c.name);
  assert.equal(names.length, 33);
  assert.equal(new Set(names).size, 33);
  assert.ok(!names.includes('any'));
  assert.equal(LEGAL_CATALOG_VERSION, 6);
});

test('every catalog predicate\'s default template keys have default text, with only bound placeholders', () => {
  // Go: legal/templates_test.go and templates_bindings_test.go. Each
  // constructor is given a minimal valid spec; what it says it can emit must
  // be renderable from the default table.
  const minimal: Record<string, { args?: string[]; sub?: { name: string; args: string[] }[] }> = {
    propAtLeast: { args: ['player.A', '1'] },
    propCompare: { args: ['player.A', '==', '1'] },
    playerBool: { args: ['A'] },
    playerBoolAt: { args: ['player.A', 'true'] },
    componentPresentAt: { args: ['game.S', 'move.I'] },
    componentAbsentAt: { args: ['game.S', 'move.I'] },
    componentPresentAtKey: { args: ['game.S', 'move.K'] },
    mayMoveTo: { args: ['game.S', 'game.T', 'move.I'] },
    mayMoveToSlot: { args: ['game.S', 'game.T', 'move.I', 'move.J'] },
    mayMoveFirstTo: { args: ['game.S', 'game.T'] },
    mayMoveFirstToSlot: { args: ['game.S', 'game.T', 'move.J'] },
    mayMoveAllTo: { args: ['game.S', 'game.T'] },
    mayMoveCountTo: { args: ['game.S', 'game.T', 'move.N'] },
    mayMoveFixedCountTo: { args: ['game.S', 'game.T', '2'] },
    maySwapComponents: { args: ['game.S', 'move.I', 'move.J'] },
    maySwapComponentsByKey: { args: ['game.S', 'move.I', 'move.J'] },
    allActivePlayers: { sub: [{ name: 'playerBool', args: ['A'] }] },
    proposerIsCurrentPlayer: {},
    proposerIsPlayerFromMove: { args: ['Who'] },
    revealableCardAt: { args: ['game.S', 'game.T', 'move.I'] },
    componentPropEqualsCurrentPlayer: { args: ['game.S', 'move.K', 'Color'] },
    inPhase: { args: ['1'] },
    inProgression: { args: ['Move'] },
    stackConstraints: { args: ['S', 'T'] },
    stackCount: { args: ['game.S', '==', '1'] },
    stackEmpty: { args: ['game.S'] },
    stackNotEmpty: { args: ['game.S'] },
    propEquals: { args: ['player.A', '1'] },
    propNotEquals: { args: ['player.A', '1'] },
    componentPropEquals: { args: ['game.S', 'first-occupied', 'F', 'v'] },
    componentPropNotEquals: { args: ['game.S', 'first-occupied', 'F', 'v'] },
    componentPropsEqual: { args: ['game.S', 'first-occupied', 'game.T', 'first-occupied', 'F'] },
    componentPropsNotEqual: { args: ['game.S', 'first-occupied', 'game.T', 'first-occupied', 'F'] },
  };
  const table = defaultLegalTemplates();
  const emitted = new Set<string>(['legal.any_failed']);
  for (const ctor of defaultLegalConstructors()) {
    assert.ok(minimal[ctor.name], `no minimal spec for ${ctor.name}`);
    const [p] = resolveLegalSpecs([{ name: ctor.name, ...minimal[ctor.name] }], defaultLegalRegistry());
    assert.ok(p.emittedTemplates.length > 0, ctor.name);
    assert.ok(p.emittedBindings, ctor.name);
    assert.deepEqual(Object.keys(p.emittedBindings).sort(), [...p.emittedTemplates].sort(), ctor.name);
    for (const key of p.emittedTemplates) {
      emitted.add(key);
      assert.ok(Object.prototype.hasOwnProperty.call(table, key), `${ctor.name} emits ${key}, which has no default text`);
      for (const placeholder of legalTemplatePlaceholders(table[key])) {
        assert.ok(p.emittedBindings[key].includes(placeholder), `${key} references {${placeholder}}, which ${ctor.name} does not bind`);
      }
    }
  }
  // The only default keys no bare predicate emits are the behavior helpers' overrides.
  const helperOnly = Object.keys(table).filter(key => !emitted.has(key)).sort();
  assert.deepEqual(helperOnly, [
    'legal.player_active',
    'legal.player_already_submitted',
    'legal.player_inactive',
    'legal.player_not_admin',
    'legal.player_not_submitted',
    'legal.seat_not_closed',
    'legal.seat_not_filled',
  ]);
});

test('player index validity, next and ensureValid follow the delegate\'s answers', () => {
  const three = fx.state({ players: [{}, {}, {}], playersMayBeActive: [true, false, true] });
  assert.equal(playerIndexValid(0, three), true);
  assert.equal(playerIndexValid(1, three), false);
  assert.equal(playerIndexValid(3, three), false);
  assert.equal(playerIndexValid(-4, three), false);
  for (const special of [OBSERVER_PLAYER_INDEX, ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX]) {
    assert.equal(playerIndexValid(special, three), true);
    assert.equal(playerIndexNext(special, three), special);
  }
  assert.equal(playerIndexNext(0, three), 2, 'skips the inactive seat');
  assert.equal(playerIndexNext(2, three), 0, 'wraps');
  assert.equal(playerIndexEnsureValid(1, three), 2);
  assert.equal(playerIndexEnsureValid(2, three), 2);
  assert.equal(playerIndexEnsureValid(7, three), 0, 'an index past the end wraps to the first active seat');

  const nobody = fx.state({ players: [{}, {}], playersMayBeActive: [false, false] });
  assert.equal(playerIndexNext(0, nobody), 0, 'no active seat: stays put');
});

test('a custom player order replaces seat order', () => {
  const s = fx.state({ players: [{}, {}, {}, {}], playersMayBeActive: [true, true, false, true], customPlayerOrder: [3, 1, 2, 0] });
  assert.equal(playerIndexNext(3, s), 1);
  assert.equal(playerIndexNext(1, s), 0, 'skips inactive seat 2');
  assert.equal(playerIndexNext(0, s), 3, 'wraps');
  assert.equal(playerIndexNext(2, s), 0);
  assert.equal(playerIndexNext(9, s), 9, 'not in the order: unchanged');
  const broken = fx.state({ players: [{}, {}], customPlayerOrder: [0, 5] });
  assert.equal(playerIndexNext(0, broken), 0, 'an out-of-range entry stops the walk');
});

test('Next panics exactly where Go indexes outside the player slice', () => {
  const two = fx.state({ players: [{}, {}] });
  assert.throws(() => playerIndexNext(-7, two), { name: 'LegalPanic', message: 'runtime error: index out of range [-6]' });
  assert.throws(() => playerIndexNext(0, fx.state()), { name: 'LegalPanic', message: 'runtime error: index out of range [0] with length 0' });
});

test('player index equivalence: observer matches nothing, admin and any match everything else', () => {
  const O = OBSERVER_PLAYER_INDEX;
  const A = ADMIN_PLAYER_INDEX;
  const Y = ANY_PLAYER_INDEX;
  const cases: [number, number, boolean][] = [
    [0, 0, true],
    [0, 1, false],
    [O, O, false],
    [O, 0, false],
    [0, O, false],
    [O, A, false],
    [A, O, false],
    [O, Y, false],
    [A, 0, true],
    [0, A, true],
    [A, A, true],
    [A, Y, true],
    [Y, 3, true],
    [3, Y, true],
    [Y, Y, true],
    [-4, 0, false],
    [0, -4, false],
    [-4, A, false],
    [-4, -4, false],
  ];
  for (const [p, other, want] of cases) {
    assert.equal(playerIndexEquivalent(p, other), want, `${p} ~ ${other}`);
  }
});

test('goAtoi is strconv.Atoi', () => {
  assert.deepEqual(goAtoi('42'), { ok: true, value: 42 });
  assert.deepEqual(goAtoi('-7'), { ok: true, value: -7 });
  assert.deepEqual(goAtoi('+7'), { ok: true, value: 7 });
  assert.deepEqual(goAtoi('007'), { ok: true, value: 7 });
  for (const bad of ['', ' 1', '1 ', '1.0', '1e3', '0x10', '1_000', '--1', 'one', '١']) {
    assert.deepEqual(goAtoi(bad), { ok: false, error: `strconv.Atoi: parsing ${goQuote(bad)}: invalid syntax` }, bad);
  }
  assert.deepEqual(goAtoi('99999999999999999999'), {
    ok: false,
    error: 'strconv.Atoi: parsing "99999999999999999999": value out of range',
  });
  // The documented divergence: fits int64, but not exactly representable here.
  assert.equal(goAtoi('9007199254740993').ok, false);
  assert.deepEqual(goAtoi('9007199254740991'), { ok: true, value: 9007199254740991 });
});

test('goQuote and goStringSlice are %q and %v', () => {
  assert.equal(goQuote('game.Cards'), '"game.Cards"');
  assert.equal(goQuote('say "hi"\\now'), '"say \\"hi\\"\\\\now"');
  assert.equal(goQuote('a\tb\nc\u0001'), '"a\\tb\\nc\\x01"');
  assert.equal(goQuote('héllo ♠'), '"héllo ♠"');
  assert.equal(goStringSlice(['A', 'B']), '[A B]');
  assert.equal(goStringSlice([]), '[]');
  assert.equal(goStringSlice(undefined), '[]');
});

test('renderLegalVerdict is the text move.Legal() returns', () => {
  const table = mergeLegalTemplates({ 'reveal.no_cards_left': 'You have no cards left to reveal this turn' });
  assert.equal(renderLegalVerdict(passVerdict(), table), '');
  assert.equal(renderLegalVerdict(failT('legal.prop_at_least', { value: 0, min: 1 }), table), 'requires at least 1, but the current value is 0');
  assert.equal(renderLegalVerdict(failT('reveal.no_cards_left'), table), 'You have no cards left to reveal this turn');
  // An unregistered key renders as itself: how a plain LegalCustom error reads.
  assert.equal(renderLegalVerdict(failT('the timer is already active'), table), 'the timer is already active');
  // A missing binding renders its bare name; nothing throws.
  assert.equal(renderLegalVerdict(failT('legal.stack_count', { count: 2 }), table), "requires the stack's count (2) to be op n");
  assert.equal(renderLegalVerdict(unknownVerdict('reads hidden property'), table), 'reads hidden property');
  // An unknown that carries a message renders the message, not the reason.
  assert.equal(
    renderLegalVerdict({ outcome: 'unknown', message: { template: 'legal.any_failed' }, reason: 'sub-predicate "x" of "any" was unknown' }, table),
    'none of the required conditions were satisfied',
  );
  assert.equal(renderLegalVerdict({ outcome: 'unknown' }, table), 'boardgame: move is not legal');
  assert.equal(renderLegalVerdict(failT('legal.stack_empty'), undefined), 'legal.stack_empty');
});

test('the shared client renderer is the one this package renders with', () => {
  const message = legalErrorf('k', { n: 3, flag: false, who: 'ana' }).message;
  assert.equal(renderLegalMessage(message, { k: '{who} has {n} ({flag}) {missing}' }), 'ana has 3 (false) missing');
  assert.deepEqual(legalErrorf('k'), { outcome: 'fail', message: { template: 'k' } });
  assert.deepEqual(legalTemplatePlaceholders('{a} then {b_2} then {a} but not {no-dash} or { spaced }'), ['a', 'b_2']);
});

test('mergeLegalTemplates overlays a game\'s table on the defaults without mutating them', () => {
  const merged = mergeLegalTemplates({ 'legal.stack_empty': 'empty it first', 'game.own': 'mine' }, null, { 'game.own': 'later wins' });
  assert.equal(merged['legal.stack_empty'], 'empty it first');
  assert.equal(merged['game.own'], 'later wins');
  assert.equal(merged['legal.any_failed'], 'none of the required conditions were satisfied');
  assert.equal(defaultLegalTemplates()['legal.stack_empty'], 'requires the stack to be empty');
  assert.equal(Object.keys(defaultLegalTemplates()).length, 38);
});
