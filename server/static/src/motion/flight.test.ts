import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { compileViewportFlight } from './flight.ts';
import type { ViewportGeometry } from './geometry.ts';

const rect = (left: number, top: number, width = 20, height = 20): ViewportGeometry =>
  Object.freeze({ space: 'viewport', left, top, width, height });

describe('compileViewportFlight', () => {
  it('aligns centers and preserves the carrier resting transform', () => {
    const result = compileViewportFlight(
      rect(10, 20, 10, 30),
      rect(100, 200, 30, 10),
      'matrix(1, 0, 0, 1, 4, 5)',
    );
    assert.deepEqual(result.inversion, {
      translateX: -100,
      translateY: -170,
      scale: 1,
      changed: true,
    });
    assert.deepEqual(result.tracks, [{
      target: 'host',
      property: 'transform',
      from: 'translate(-100px, -170px) matrix(1, 0, 0, 1, 4, 5)',
      to: 'matrix(1, 0, 0, 1, 4, 5)',
    }]);
    assert.ok(Object.isFrozen(result));
    assert.ok(Object.isFrozen(result.tracks));
  });

  it('returns no track for a stationary centered carrier', () => {
    const result = compileViewportFlight(rect(10, 20), rect(10.2, 19.8));
    assert.equal(result.inversion.changed, false);
    assert.deepEqual(result.tracks, []);
  });
});
