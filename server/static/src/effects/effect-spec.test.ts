import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { createEffectTransitionContext, defineEffectTheme, fx } from './effect-spec.ts';

describe('effect descriptors', () => {
  it('separates recipe, semantic tone, and intensity in immutable data', () => {
    const descriptor = fx.burst({
      at: fx.anchor('score'),
      tone: 'reward',
      intensity: 'small',
      key: 'claim-point',
      advanced: { count: 9, palette: ['gold', 'white'] },
    });
    assert.deepEqual(descriptor, {
      kind: 'burst',
      at: { kind: 'named', name: 'score' },
      tone: 'reward',
      intensity: 'small',
      key: 'claim-point',
      advanced: { count: 9, palette: ['gold', 'white'] },
    });
    assert.ok(Object.isFrozen(descriptor));
    assert.ok(Object.isFrozen(descriptor.at));
    assert.ok(Object.isFrozen(descriptor.advanced));
    assert.ok(Object.isFrozen(descriptor.advanced?.palette));
  });

  it('composes the same leaf descriptors without mutating them', () => {
    const pulse = fx.pulse({ at: fx.point(10, 20), tone: 'attention' });
    const travel = fx.travel({ from: fx.anchor('bank'), to: fx.anchor('hand') });
    const composition = fx.sequence([
      fx.parallel([pulse, travel]),
      fx.burst({ at: fx.anchor('hand'), tone: 'reward' }),
    ], { gapMs: 30, intensity: 'medium' });
    assert.equal(composition.kind, 'sequence');
    assert.equal(composition.effects[0].kind, 'parallel');
    assert.ok(Object.isFrozen(composition.effects));
    assert.equal(pulse.tone, 'attention');
  });

  it('rejects ambiguous identity, anchor, point, and timing inputs', () => {
    assert.throws(() => fx.anchor('  '), /anchor name/);
    assert.throws(() => fx.point(Number.NaN, 0), /finite/);
    assert.throws(() => fx.motion('  '), /motion subject ID/);
    assert.throws(() => fx.motion('card', 'middle' as never), /motion moment/);
    assert.throws(() => fx.pulse({ at: fx.point(0, 0), key: '' }), /effect key/);
    assert.throws(() => fx.sequence([], { gapMs: -1 }), /gapMs/);
  });

  it('describes privacy-safe structural departure and arrival points', () => {
    const arrival = fx.motion('card-17');
    const departure = fx.motion('card-17', 'departure');
    assert.deepEqual(arrival, {
      kind: 'motion', subjectId: 'card-17', moment: 'arrival',
    });
    assert.deepEqual(departure, {
      kind: 'motion', subjectId: 'card-17', moment: 'departure',
    });
    assert.equal(Object.isFrozen(arrival), true);
  });

  it('describes a structural trail without independent timing', () => {
    const descriptor = fx.trail({
      subject: 'card-22',
      tone: 'magic',
      intensity: 'subtle',
      advanced: { echoes: 3, lagMs: 18, opacity: 0.4, palette: ['violet'] },
    });
    assert.deepEqual(descriptor, {
      kind: 'trail',
      subject: 'card-22',
      tone: 'magic',
      intensity: 'subtle',
      advanced: { echoes: 3, lagMs: 18, opacity: 0.4, palette: ['violet'] },
    });
    assert.equal(Object.isFrozen(descriptor), true);
    assert.equal(Object.isFrozen(descriptor.advanced), true);
    assert.equal(Object.isFrozen(descriptor.advanced?.palette), true);
    assert.throws(() => fx.trail({ subject: ' ' }), /trail subject ID/);
  });

  it('groups lifecycle-bound motion decoration as immutable data', () => {
    const descriptor = fx.decorateMotion({
      subject: 'card-22',
      trail: { tone: 'magic', intensity: 'small' },
      arrival: fx.burst({ at: fx.motion('card-22'), tone: 'reward' }),
    });
    assert.equal(descriptor.kind, 'decorate-motion');
    assert.equal(descriptor.subject, 'card-22');
    assert.deepEqual(descriptor.effects.map(effect => effect.kind), ['trail', 'burst']);
    assert.equal(Object.isFrozen(descriptor), true);
    assert.equal(Object.isFrozen(descriptor.effects), true);
    assert.throws(
      () => fx.decorateMotion({ subject: 'card-22' }),
      /requires a trail, departure, or arrival/,
    );
  });

  it('copies and freezes game-local themes', () => {
    const reward = ['#ffd700'];
    const theme = defineEffectTheme({ tones: { reward } });
    reward.push('#fff');
    assert.deepEqual(theme.tones?.reward, ['#ffd700']);
    assert.ok(Object.isFrozen(theme));
    assert.ok(Object.isFrozen(theme.tones));
    assert.ok(Object.isFrozen(theme.tones?.reward));
  });

  it('provides a pure initial/transition context with selector-based changes', () => {
    const initial = createEffectTransitionContext({
      before: null,
      after: { score: 1, label: 'A' },
      move: null,
      version: 1,
      snapshotEpoch: 4,
    });
    assert.equal(initial.kind, 'initial');
    assert.equal(initial.changed(state => state.score), true);

    const transition = createEffectTransitionContext({
      before: { score: 1, label: 'A' },
      after: { score: 2, label: 'A' },
      move: { Name: 'Score', Version: 2 },
      version: 2,
      snapshotEpoch: 5,
    });
    assert.equal(transition.kind, 'transition');
    assert.equal(transition.changed(state => state.score), true);
    assert.equal(transition.changed(state => state.label), false);
    assert.ok(Object.isFrozen(transition));
  });
});
