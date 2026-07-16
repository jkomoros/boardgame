import {
  PlacementDraftController,
  type BoundMoveAction,
  type DraftPlacement,
} from '../client.js';
import type { GameSnapshotHost } from './snapshot-controller.js';

type Tile = 'a' | 'b';
type Square = 0 | 1 | 2;
declare const host: GameSnapshotHost;
declare const commit: (
  placements: readonly DraftPlacement<Tile, Square>[],
) => BoundMoveAction<'Play', { Word: string }>;

const controller = new PlacementDraftController<Tile, Square>(host);
const draft = controller.bind({
  items: ['a', 'b'] as const,
  targets: [0, 1, 2] as const,
  minPlacements: 2,
  rebase: 'keep-valid',
  action: commit,
});
draft.selectItem('a');
draft.place(0);
draft.assign('b', 1);
draft.item('a').select();
draft.target(2).place();
draft.targetFor('a');
draft.itemAt(1);
draft.action?.activate();
const controls = document.createElement('boardgame-draft-controls');
controls.draft = draft;

// @ts-expect-error item keys retain their literal union
draft.selectItem('c');
// @ts-expect-error target keys retain their literal union
draft.place(4);
// @ts-expect-error item bindings retain their literal union
draft.item('c');
// @ts-expect-error target bindings retain their literal union
draft.target(4);
new PlacementDraftController<Tile, Square>(host).bind({
  items: ['a'], targets: [0], action: commit,
  // @ts-expect-error rebase is a closed policy
  rebase: 'merge',
});
