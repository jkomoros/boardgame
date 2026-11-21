/**
 * The renderer-host contract: what whatever mounts a game renderer must
 * supply, written as types so a second host (a Fabric guest, a fixture) can be
 * checked against the same shape as the webapp.
 *
 * This file is renderer-kit code (layer 2, see scripts/layers.json). It must
 * not import the Redux store, the HTTP API, routing or anything else a
 * particular host is built from. The webapp's implementation lives in
 * `src/webapp-renderer-host.ts` and `boardgame-game-view`.
 */
import type { MovePreviewBatchResult } from '../legal/previewLegality.js';

/** The game a request is about. A host may ignore it if it serves one game. */
export interface RendererGameRef {
  readonly gameName: string;
  readonly gameId: string;
}

// ---- Legality preview for board targets ----

/** One batch of candidate arg-sets of a single move, to be checked without applying anything. */
export interface BoardPreviewRequest extends RendererGameRef {
  readonly moveName: string;
  /** One entry per candidate, in the order results must come back. */
  readonly candidates: readonly { readonly args: Readonly<Record<string, string>> }[];
  /** The perspective the host's displayed legality uses. */
  readonly asPlayer: number;
  readonly asAdmin: boolean;
}

export interface BoardPreviewTransport {
  /**
   * Resolves with one result per candidate, in candidate order, or null when
   * the host could not get an answer (the renderer then fails closed and
   * retries). May reject; a rejection is treated like null.
   */
  previewBoard(request: BoardPreviewRequest): Promise<readonly MovePreviewBatchResult[] | null>;
}

// ---- Table actions ----

/**
 * The outcome of a table action. `rate-limited` and `conflict` are the two
 * refusals the Table surface words specially; everything else is `failed`
 * with the host's message, if it has one.
 */
export type TableActionResult =
  | { readonly ok: true; readonly locked?: boolean }
  | {
    readonly ok: false;
    readonly reason: 'rate-limited' | 'conflict' | 'failed';
    readonly message?: string;
  };

/** How a person joins this table: the address to type and a QR image for it. */
export interface TableJoinInfo {
  /** Human-readable address, e.g. `example.com/join`. Empty hides the instructions. */
  readonly address: string;
  /** URL of a QR image for the join address. Empty hides the image. */
  readonly qrUrl: string;
}

/**
 * The actions the table owner's shared display can take. `BoardgameTableViewBase`
 * renders the controls; the host carries them out. Each may reject on a
 * transport or protocol error, which the Table surface reports as a failure.
 */
export interface TableActions {
  setRoomLock(game: RendererGameRef, locked: boolean): Promise<TableActionResult>;
  switchToSolo(game: RendererGameRef): Promise<TableActionResult>;
  skipTurn(game: RendererGameRef): Promise<TableActionResult>;
  joinInfo(game: RendererGameRef): TableJoinInfo;
}

/**
 * What `boardgame-render-game` takes from its host as one object. Members are
 * stable services, not reactive data: reactive inputs (state, move forms, the
 * move transports) stay ordinary properties on the element.
 */
export interface RendererHost {
  /** Board-target legality preview. Absent means the host offers none, and no target is grayed by preview. */
  readonly boardPreview?: BoardPreviewTransport;
  /** Table actions for a Table surface. Absent means the controls report that they are unavailable. */
  readonly tableActions?: TableActions;
}
