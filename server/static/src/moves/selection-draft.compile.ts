import {
  SelectionDraftController,
  type BoundMoveAction,
} from '../client.js';
import type { GameSnapshotHost } from './snapshot-controller.js';

type Card = 'clay' | 'ore' | 'wool';
declare const host: GameSnapshotHost;
declare const commit: (selected: readonly Card[]) => BoundMoveAction<'Trade', { Cards: string }>;

const controller = new SelectionDraftController<Card>(host);
const draft = controller.bind({
  candidates: ['clay', 'ore', 'wool'] as const,
  minSelected: 2,
  maxSelected: 3,
  rebase: 'keep-valid',
  action: commit,
});
draft.toggle('clay');
draft.select('ore');
draft.deselect('clay');
draft.isSelected('ore');
draft.action?.activate();
const controls = document.createElement('boardgame-draft-controls');
controls.draft = draft;

// @ts-expect-error candidate keys retain their literal union
draft.toggle('brick');
new SelectionDraftController<Card>(host).bind({
  candidates: ['clay'], action: commit,
  // @ts-expect-error rebase is a closed policy
  rebase: 'merge',
});
