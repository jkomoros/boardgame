import type { ClientMove } from '../client.js';

declare const move: ClientMove;
move.Name satisfies string;
move.Version satisfies number;

// @ts-expect-error animation metadata never exposes serialized move arguments
move.Blob;
// @ts-expect-error animation metadata never exposes the proposer identity
move.Proposer;
// @ts-expect-error installed move metadata is immutable
move.Name = 'Different move';
