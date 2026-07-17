import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
  captureOffsetGeometry,
  captureViewportGeometry,
  centeredInversionDelta,
  geometryCenter,
  solveFlipGeometry,
} from './geometry.ts';

describe('motion geometry', () => {
  it('captures viewport geometry without retaining a mutable DOMRect', () => {
    const source = { top: 10, left: 20, width: 30, height: 40 };
    const element = { getBoundingClientRect: () => source };
    const captured = captureViewportGeometry(element);
    source.left = 999;
    assert.deepEqual(captured, {
      space: 'viewport', top: 10, left: 20, width: 30, height: 40,
    });
  });

  it('preserves the animator offset-parent coordinate space', () => {
    const ancestor = {
      offsetTop: 7,
      offsetLeft: 11,
      offsetWidth: 100,
      offsetHeight: 80,
      offsetParent: null,
    } as unknown as HTMLElement;
    const parent = {
      offsetTop: 5,
      offsetLeft: 3,
      offsetWidth: 50,
      offsetHeight: 40,
      offsetParent: ancestor,
    } as unknown as HTMLElement;
    const element = {
      offsetTop: 13,
      offsetLeft: 17,
      offsetWidth: 30,
      offsetHeight: 20,
      offsetParent: parent,
    } as unknown as HTMLElement;
    assert.deepEqual(captureOffsetGeometry(element, ancestor), {
      space: 'offset',
      top: 25,
      left: 31,
      width: 30,
      height: 20,
    });
  });

  it('computes centers and the inversion from a resting target to a source', () => {
    const resting = { space: 'viewport' as const, top: 40, left: 80, width: 20, height: 10 };
    const source = { space: 'viewport' as const, top: 10, left: 20, width: 40, height: 30 };
    assert.deepEqual(geometryCenter(resting), { space: 'viewport', x: 90, y: 45 });
    assert.deepEqual(centeredInversionDelta(resting, source), {
      space: 'viewport', x: -50, y: -20,
    });
  });

  it('solves translation, scaling, rotation-aware scaling, and visible change', () => {
    const before = { space: 'offset' as const, top: 10, left: 20, width: 40, height: 80 };
    const after = { space: 'offset' as const, top: 30, left: 50, width: 20, height: 40 };
    assert.deepEqual(solveFlipGeometry(before, after, { beforeTransform: 'rotate(3deg)' }), {
      translateX: -20,
      translateY: 0,
      scale: 2,
      changed: true,
      invertedTransform: 'translateY(0px) translateX(-20px) rotate(3deg) scale(2)',
    });
    assert.equal(solveFlipGeometry(before, after, { rotates: true }).scale, 4);
    assert.equal(solveFlipGeometry(after, after).changed, false);
  });

  it('falls back to a finite identity scale for zero-sized geometry', () => {
    const before = { space: 'offset' as const, top: 0, left: 0, width: 0, height: 0 };
    const after = { space: 'offset' as const, top: 0, left: 0, width: 0, height: 0 };
    const solution = solveFlipGeometry(before, after);
    assert.equal(solution.scale, 1);
    assert.equal(solution.changed, false);
    assert.equal(solution.invertedTransform.includes('NaN'), false);
    assert.equal(solution.invertedTransform.includes('Infinity'), false);
  });

  it('never emits non-finite translations from malformed snapshots', () => {
    const before = {
      space: 'offset' as const,
      top: Number.NaN,
      left: Number.POSITIVE_INFINITY,
      width: 10,
      height: 10,
    };
    const after = {
      space: 'offset' as const,
      top: 0,
      left: 0,
      width: 10,
      height: 10,
    };
    const solution = solveFlipGeometry(before, after);
    assert.equal(solution.translateX, 0);
    assert.equal(solution.translateY, 0);
    assert.equal(solution.invertedTransform.includes('NaN'), false);
    assert.equal(solution.invertedTransform.includes('Infinity'), false);
  });
});
