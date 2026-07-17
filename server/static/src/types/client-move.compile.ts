import type { ClientMove } from '../client.js';

declare const move: ClientMove;
move.AnimationKey satisfies string;
move.Version satisfies number;
move.Properties?.Example satisfies import('../client.js').JsonValue | undefined;

// @ts-expect-error animation metadata never exposes serialized move arguments
move.Blob;
// @ts-expect-error animation metadata never exposes the proposer identity
move.Proposer;
// @ts-expect-error installed move metadata is immutable
move.AnimationKey = 'Different move';
