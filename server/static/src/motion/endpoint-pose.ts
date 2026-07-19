export type MotionEndpointOrientation = 'natural' | 'quarter-turned';

export function motionAxesDiffer(
  before: MotionEndpointOrientation,
  after: MotionEndpointOrientation,
): boolean {
  if ((before !== 'natural' && before !== 'quarter-turned')
    || (after !== 'natural' && after !== 'quarter-turned')) {
    throw new Error('motion endpoint orientation must be natural or quarter-turned');
  }
  return before !== after;
}
