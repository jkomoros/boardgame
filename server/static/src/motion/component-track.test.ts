import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  compileComponentMotionTracks,
  componentMotionTracks,
} from './component-track.ts';

describe('component motion tracks', () => {
  it('compiles structural host and component-owned visual channels immutably', () => {
    const visual = [{
      target: 'visual' as const,
      property: 'transform' as const,
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
        target: 'host', property: 'transform',
        from: 'translate(20px, 10px)', to: 'none',
      },
      { target: 'host', property: 'opacity', from: '0', to: '1' },
      {
        target: 'visual', property: 'transform',
        from: 'rotateY(0deg)', to: 'rotateY(180deg)',
      },
    ]);
    assert.equal(Object.isFrozen(tracks), true);
    assert.equal(Object.isFrozen(tracks[0]), true);
    visual[0].to = 'tampered';
    assert.equal(tracks[2].to, 'rotateY(180deg)');
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
      { target: 'visual', property: 'transform', from: 'none', to: 'rotate(1deg)' },
      { target: 'visual', property: 'opacity', from: '0', to: '1' },
    ]);
    assert.throws(() => componentMotionTracks([
      { target: 'visual', property: 'opacity', from: 'hidden', to: '1' },
    ]), /must be finite/);
  });
});
