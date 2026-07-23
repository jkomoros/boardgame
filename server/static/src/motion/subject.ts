export type MotionSilhouetteShape = 'rectangle' | 'rounded-rectangle' | 'circle';

/**
 * Sanitized visual capability for subject-following decoration.
 *
 * A silhouette deliberately contains no DOM, text, image, game property, or
 * computed color. Geometry comes from the separately captured motion plan.
 */
export interface MotionSilhouetteSnapshot {
  readonly kind: 'silhouette';
  readonly shape: MotionSilhouetteShape;
}

export type MotionSubjectSnapshot = MotionSilhouetteSnapshot;

export function motionSilhouette(
  shape: MotionSilhouetteShape = 'rectangle',
): MotionSilhouetteSnapshot {
  if (shape !== 'rectangle' && shape !== 'rounded-rectangle' && shape !== 'circle') {
    throw new Error('motion silhouette shape is invalid');
  }
  return Object.freeze({ kind: 'silhouette', shape });
}

/** Copy only the exact safe protocol; malformed/extended values opt out. */
export function sanitizeMotionSubjectSnapshot(value: unknown): MotionSubjectSnapshot | null {
  if (!value || typeof value !== 'object') return null;
  const candidate = value as Record<string, unknown>;
  if (Object.keys(candidate).some(key => key !== 'kind' && key !== 'shape')) return null;
  if (candidate.kind !== 'silhouette') return null;
  if (candidate.shape !== 'rectangle'
    && candidate.shape !== 'rounded-rectangle'
    && candidate.shape !== 'circle') return null;
  return motionSilhouette(candidate.shape);
}
