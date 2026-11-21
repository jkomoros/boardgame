import assert from 'node:assert/strict';
import { it } from 'node:test';
import { partitionMotionTransferOwnership } from './transfer-arbitration.ts';

it('partitions transfer ownership from exact identity facts', () => {
  assert.deepEqual(partitionMotionTransferOwnership([
    { key: 'external', carrierKind: 'external', subjectMatchesCarrier: false, carrierResolvesExactly: true, beforeSightings: 0, afterSightings: 0 },
    { key: 'incoming', carrierKind: 'stack', subjectMatchesCarrier: true, carrierResolvesExactly: true, beforeSightings: 0, afterSightings: 1 },
    { key: 'retained', carrierKind: 'stack', subjectMatchesCarrier: true, carrierResolvesExactly: true, beforeSightings: 1, afterSightings: 1 },
    { key: 'wrong', carrierKind: 'stack', subjectMatchesCarrier: false, carrierResolvesExactly: true, beforeSightings: 0, afterSightings: 1 },
  ]), [
    { key: 'external', disposition: 'explicit' },
    { key: 'incoming', disposition: 'automatic' },
    { key: 'retained', disposition: 'conflict', reason: 'retained' },
    { key: 'wrong', disposition: 'conflict', reason: 'identity-mismatch' },
  ]);
});
