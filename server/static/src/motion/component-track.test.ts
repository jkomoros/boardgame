import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  compileComponentMotionTracks,
  componentMotionKeyframes,
  componentMotionTrackEasing,
  componentMotionTracks,
} from './component-track.ts';

describe('component motion tracks', () => {
  it('compiles structural host and component-owned visual channels immutably', () => {
    const visual: {
      target: 'visual', property: 'transform', from: string, to: string,
    }[] = [{
      target: 'visual',
      property: 'transform',
      from: 'rotateY(0deg)',
      to: 'rotateY(180deg)',
    }];
    const tracks = compileComponentMotionTracks({
      needsHostTransition: true,
      invertedTransform: 'translate(20px, 10px)',
      finalTransform: '',
      beforeOpacity: '0',
      finalOpacity: '1',
      visualTracks: visual,
    });

    assert.deepEqual(tracks, [
      {
        target: 'host', property: 'transform', timeline: 'eased',
        samples: [
          { offset: 0, value: 'translate(20px, 10px)' },
          { offset: 1, value: 'none' },
        ],
      },
      {
        target: 'host', property: 'opacity', timeline: 'eased',
        samples: [{ offset: 0, value: '0' }, { offset: 1, value: '1' }],
      },
      {
        target: 'visual', property: 'transform', timeline: 'eased',
        samples: [
          { offset: 0, value: 'rotateY(0deg)' },
          { offset: 1, value: 'rotateY(180deg)' },
        ],
      },
    ]);
    assert.equal(Object.isFrozen(tracks), true);
    assert.equal(Object.isFrozen(tracks[0]), true);
    visual[0].to = 'tampered';
    assert.equal(tracks[2].samples[1].value, 'rotateY(180deg)');
  });

  it('drops no-ops and unusable opacity without inventing work', () => {
    assert.deepEqual(compileComponentMotionTracks({
      needsHostTransition: false,
      invertedTransform: '',
      finalTransform: '',
      beforeOpacity: 'not-a-number',
      finalOpacity: '1',
      visualTracks: [{
        target: 'visual', property: 'transform', from: 'none', to: 'none',
      }],
    }), []);
  });

  it('rejects competing writers for one target/property channel', () => {
    assert.throws(() => componentMotionTracks([
      { target: 'visual', property: 'transform', from: 'a', to: 'b' },
      { target: 'visual', property: 'transform', from: 'c', to: 'd' },
    ]), /multiple owners/);
  });

  it('normalizes transform emptiness and bounds finite opacity', () => {
    assert.deepEqual(componentMotionTracks([
      { target: 'visual', property: 'transform', from: '', to: ' rotate(1deg) ' },
      { target: 'visual', property: 'opacity', from: '-2', to: '3' },
    ]), [
      {
        target: 'visual', property: 'transform', timeline: 'eased',
        samples: [{ offset: 0, value: 'none' }, { offset: 1, value: 'rotate(1deg)' }],
      },
      {
        target: 'visual', property: 'opacity', timeline: 'eased',
        samples: [{ offset: 0, value: '0' }, { offset: 1, value: '1' }],
      },
    ]);
    assert.throws(() => componentMotionTracks([
      { target: 'visual', property: 'opacity', from: 'hidden', to: '1' },
    ]), /must be finite/);
  });

  it('compiles a frozen two-keyframe WAAPI boundary', () => {
    const [track] = componentMotionTracks([
      { target: 'visual', property: 'opacity', from: '0.2', to: '0.8' },
    ]);
    const frames = componentMotionKeyframes(track);
    assert.deepEqual(frames, [
      { offset: 0, opacity: '0.2' },
      { offset: 1, opacity: '0.8' },
    ]);
    assert.equal(Object.isFrozen(frames), true);
    assert.equal(Object.isFrozen(frames[0]), true);
  });

  it('emits one offset-carrying keyframe per sample of a sampled track', () => {
    // An implementation that emitted only the endpoints would flatten every
    // curve into a straight line while every other assertion still passed.
    const [track] = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (p) => `translateX(${p * p * 4}px)`, resolution: 5,
    }]);
    assert.deepEqual(componentMotionKeyframes(track), [
      { offset: 0, transform: 'translateX(0px)' },
      { offset: 0.25, transform: 'translateX(0.25px)' },
      { offset: 0.5, transform: 'translateX(1px)' },
      { offset: 0.75, transform: 'translateX(2.25px)' },
      { offset: 1, transform: 'translateX(4px)' },
    ]);
  });
});

describe('curve tracks', () => {
  it('samples a curve at uniform offsets spanning [0,1]', () => {
    const [track] = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (p) => `translateX(${p * 10}px)`, resolution: 5,
    }]);
    assert.equal(track.timeline, 'sampled');
    assert.deepEqual(track.samples.map(s => s.offset), [0, 0.25, 0.5, 0.75, 1]);
    assert.equal(track.samples[2].value, 'translateX(5px)');
    assert.ok(Object.isFrozen(track));
  });

  it('defaults resting to curve(1)', () => {
    const [track] = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (p) => `translateX(${p}px)`, resolution: 3,
    }]);
    assert.equal(track.resting, 'translateX(1px)');
  });

  it('clamps resolution rather than rejecting it', () => {
    const [lo] = componentMotionTracks([{ target: 'visual', property: 'opacity', curve: (p) => String(p), resolution: 0 }]);
    assert.equal(lo.samples.length, 2);
    const [hi] = componentMotionTracks([{ target: 'visual', property: 'opacity', curve: (p) => String(p), resolution: 9999 }]);
    assert.equal(hi.samples.length, 256);
  });

  it('rewrites endpoint tracks into two samples and leaves them eased', () => {
    const [track] = componentMotionTracks([{ target: 'host', property: 'opacity', from: '0', to: '1' }]);
    assert.equal(track.timeline, 'eased');
    assert.deepEqual(track.samples, [{ offset: 0, value: '0' }, { offset: 1, value: '1' }]);
    assert.equal(track.resting, undefined);
  });

  it('refuses a curve on the host channel', () => {
    assert.throws(() => componentMotionTracks([
      { target: 'host', property: 'transform', curve: () => 'none' } as never,
      // Not just /host/: the generic "target must be host or visual" message
      // would satisfy that too, and this test must fail if the wrong throw wins.
    ]), /curves are not allowed on the host channel/);
  });

  it('refuses a constant curve instead of silently vacating the channel', () => {
    assert.throws(() => componentMotionTracks([
      { target: 'visual', property: 'transform', curve: () => 'none' },
    ]), /constant/);
  });

  it('revalidates an already-compiled track without losing its timeline', () => {
    // The animator re-runs a component's planned track list through the
    // compiler for ownership checks, so compiled tracks are legal input.
    const [curved] = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (p) => `translateX(${p}px)`, resolution: 4,
    }]);
    const [round] = componentMotionTracks([curved]);
    assert.deepEqual(round, curved);
    assert.notEqual(round, curved);
    assert.equal(componentMotionTrackEasing(round), 'linear');
    // Re-freezing is the whole point of the branch, so pin it.
    assert.equal(Object.isFrozen(round), true);
    assert.equal(Object.isFrozen(round.samples), true);
    assert.equal(Object.isFrozen(round.samples[0]), true);
    assert.throws(() => componentMotionTracks([
      { ...curved, samples: [{ offset: 0, value: 'none' }] },
    ]), /at least two samples/);
  });

  it('revalidates a compiled track as strictly as it compiles one', () => {
    const [curved] = componentMotionTracks([{
      target: 'visual', property: 'transform',
      curve: (p) => `translateX(${p}px)`, resolution: 3,
    }]);
    const sample = (offset: number, value: string) => ({ offset, value });

    // Decreasing offsets compile happily today and then throw inside
    // element.animate at playback time, far from the producer that caused it.
    assert.throws(() => componentMotionTracks([{
      ...curved,
      samples: [sample(0.9, 'translateX(0px)'), sample(0.1, 'translateX(1px)')],
    }]), /strictly increase/);

    // Offsets that never reach the endpoints leave the channel undefined
    // outside the sampled window.
    assert.throws(() => componentMotionTracks([{
      ...curved,
      samples: [sample(0.3, 'translateX(0px)'), sample(0.4, 'translateX(1px)')],
    }]), /span \[0,1\]/);

    // Non-uniform spacing silently retimes the trajectory the curve encoded.
    assert.throws(() => componentMotionTracks([{
      ...curved,
      samples: [
        sample(0, 'translateX(0px)'),
        sample(0.01, 'translateX(0.5px)'),
        sample(1, 'translateX(1px)'),
      ],
    }]), /uniformly spaced/);

    // A constant sampled track is exactly what the constant-curve throw exists
    // to reject; the from === to elision only covers eased tracks.
    assert.throws(() => componentMotionTracks([{
      ...curved,
      samples: [
        sample(0, 'translateX(0px)'),
        sample(0.5, 'translateX(0px)'),
        sample(1, 'translateX(0px)'),
      ],
    }]), /constant/);

    // The host channel stays structural whichever door the track comes in.
    assert.throws(() => componentMotionTracks([
      { ...curved, target: 'host' as const },
    ]), /curves are not allowed on the host channel/);

    // An eased track is a two-endpoint transition by definition; extra samples
    // would be silently reinterpreted by the kernel's effect-level easing.
    assert.throws(() => componentMotionTracks([{
      ...curved,
      timeline: 'eased' as const,
    }]), /exactly two samples/);

    // Carrying both forms would silently drop the curve.
    assert.throws(() => componentMotionTracks([
      { ...curved, curve: (p: number) => `translateX(${p}px)` } as never,
    ]), /both samples and a curve/);
  });

  it('clamps a NaN resolution to the default instead of rejecting it', () => {
    // The contract is "clamped, never rejected"; NaN has no magnitude to clamp
    // toward, so it falls back the same way an absent resolution does.
    const [track] = componentMotionTracks([{
      target: 'visual', property: 'opacity',
      curve: (p) => String(p), resolution: Number.NaN,
    }]);
    assert.equal(track.samples.length, 64);
  });

  it('pins linear easing for sampled tracks only', () => {
    const [sampled] = componentMotionTracks([{ target: 'visual', property: 'opacity', curve: (p) => String(p) }]);
    const [eased] = componentMotionTracks([{ target: 'host', property: 'opacity', from: '0', to: '1' }]);
    assert.equal(componentMotionTrackEasing(sampled), 'linear');
    assert.equal(componentMotionTrackEasing(eased), undefined);
  });
});
