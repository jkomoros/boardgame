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
