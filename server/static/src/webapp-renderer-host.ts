/**
 * The webapp's implementation of the renderer-host services
 * (renderer-host/renderer-host.ts): everything boardgame-render-game and the
 * renderer bases used to do against the Go server directly. Host layer; the
 * kit never imports this file.
 */
import { apiHttpPost, buildGameUrl, movePreviewBatch } from './api.js';
import { decodeHostActionResponse } from './types/host-action-response.js';
import { apiPath } from './util.js';
import type {
  BoardPreviewTransport,
  RendererGameRef,
  TableActionResult,
  TableActions,
} from './renderer-host/renderer-host.js';

type TableActionEndpoint = 'setRoomLock' | 'switchToSolo' | 'hostSkipTurn';

async function postTableAction(
  game: RendererGameRef,
  endpoint: TableActionEndpoint,
  body: Readonly<Record<string, unknown>>,
): Promise<TableActionResult> {
  const response = await apiHttpPost(buildGameUrl(game.gameName, game.gameId, endpoint), body);
  if (!response.data) {
    const message = response.error || response.friendlyError || undefined;
    return {
      ok: false,
      reason: response.status === 429 ? 'rate-limited' : response.status === 409 ? 'conflict' : 'failed',
      ...(message ? { message } : {}),
    };
  }
  // A malformed success body throws, and the Table surface reports it.
  const decoded = decodeHostActionResponse(response.data);
  return { ok: true, ...(decoded.locked === undefined ? {} : { locked: decoded.locked }) };
}

/** Table actions carried out against the Go server's host-action endpoints. */
export const webappTableActions: TableActions = {
  setRoomLock: (game, locked) => postTableAction(game, 'setRoomLock', { locked }),
  switchToSolo: game => postTableAction(game, 'switchToSolo', {}),
  skipTurn: game => postTableAction(game, 'hostSkipTurn', {}),
  joinInfo: game => ({
    address: `${window.location.host}/join`,
    // Self-hosted QR for this page's own origin.
    qrUrl: game.gameName && game.gameId
      ? `${apiPath(`game/${game.gameName}/${game.gameId}/qrcode.png`)}?origin=${encodeURIComponent(window.location.origin)}`
      : '',
  }),
};

/** Board-target legality preview through the server's movePreviewBatch endpoint. */
export const webappBoardPreview: BoardPreviewTransport = {
  previewBoard: async request => {
    const response = await movePreviewBatch(
      request.gameName,
      request.gameId,
      request.moveName,
      request.candidates.map(candidate => ({ Args: { ...candidate.args } })),
      { player: request.asPlayer, admin: request.asAdmin ? 1 : 0 },
    );
    return response.data ? (response.data.Results ?? []) : null;
  },
};
