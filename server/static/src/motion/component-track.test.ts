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
    ]), /host/);
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
    assert.throws(() => componentMotionTracks([
      { ...curved, samples: [{ offset: 0, value: 'none' }] },
    ]), /at least two samples/);
  });

  it('pins linear easing for sampled tracks only', () => {
    const [sampled] = componentMotionTracks([{ target: 'visual', property: 'opacity', curve: (p) => String(p) }]);
    const [eased] = componentMotionTracks([{ target: 'host', property: 'opacity', from: '0', to: '1' }]);
    assert.equal(componentMotionTrackEasing(sampled), 'linear');
    assert.equal(componentMotionTrackEasing(eased), undefined);
  });
});
