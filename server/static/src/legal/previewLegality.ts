// Pure helpers + types for board legality preview (the client half of the
// server's movePreviewBatch endpoint). Kept free of lit/DOM imports so the
// decision logic is unit-testable under `node --test`; the DOM glue lives in
// boardgame-render-game (dispatch + debounce) and each game renderer's
// previewSpec() (candidate enumeration).
import type { MovePreviewBatchResult } from '../api.js';

/**
 * PreviewCandidate ties one board space to the move args that would target it.
 * A game renderer produces these in previewSpec(); boardgame-render-game batches
 * their args to the server and feeds the illegal spaces back for graying.
 */
export interface PreviewCandidate {
  /** The board space (cell index) this candidate corresponds to. */
  space: number;
  /** The move's field args for targeting this space (fieldName -> string). */
  args: Record<string, string>;
}

/**
 * MovePreviewSpec is what a game renderer returns from previewSpec() to opt into
 * per-target legality preview: the move type to check, plus one candidate per
 * board space to test.
 */
export interface MovePreviewSpec {
  moveName: string;
  candidates: PreviewCandidate[];
}

/**
 * disabledSpacesFromResults maps a batch preview's results (in candidate order)
 * to the board spaces that should be grayed. A candidate is grayed when its
 * result is illegal OR missing — a short or absent response fails SAFE: never
 * leave a space clickable that the server might reject. The returned list is the
 * illegal candidates' spaces (candidate.space, not the array index, so
 * non-contiguous boards map correctly).
 */
export function disabledSpacesFromResults(
  candidates: PreviewCandidate[],
  results: readonly MovePreviewBatchResult[] | undefined
): number[] {
  const disabled: number[] = [];
  for (let i = 0; i < candidates.length; i++) {
    const r = results?.[i];
    if (!r || !r.Legal) disabled.push(candidates[i].space);
  }
  return disabled;
}

/**
 * samePreviewSpaces reports whether two grayed-space lists are identical, so the
 * caller can skip re-assigning previewDisabledSpaces (and the board re-render it
 * triggers) when a refresh produced the same set. disabledSpacesFromResults
 * yields spaces in stable candidate order, so an order-sensitive compare is both
 * correct and cheap.
 */
export function samePreviewSpaces(a: number[], b: number[]): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return false;
  }
  return true;
}

/**
 * previewOutcome is the pure decision a completed batch-preview refresh makes,
 * extracted so the (otherwise DOM/async) staleness + error-retention logic is
 * unit-testable:
 *   - 'drop-stale'   — a newer refresh superseded this one (seq moved) OR the
 *                      renderer this response was for is no longer mounted; do
 *                      nothing (never gray a board a later refresh owns).
 *   - 'keep-on-error'— still current, but the server returned no data (network/
 *                      server error); leave the prior graying rather than
 *                      flashing the whole board back to enabled.
 *   - 'apply'        — still current, mounted, and data present; gray from it.
 * Staleness is checked before the error branch: a superseded errored response
 * must not touch the board.
 */
export function previewOutcome(o: {
  startedSeq: number;
  currentSeq: number;
  rendererStillMounted: boolean;
  hasData: boolean;
}): 'apply' | 'drop-stale' | 'keep-on-error' {
  if (o.startedSeq !== o.currentSeq || !o.rendererStillMounted) return 'drop-stale';
  if (!o.hasData) return 'keep-on-error';
  return 'apply';
}
