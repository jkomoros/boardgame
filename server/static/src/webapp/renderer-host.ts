/**
 * The webapp's implementation of the renderer-host services
 * (renderer-host/renderer-host.ts): everything boardgame-render-game and the
 * renderer bases used to do against the Go server directly. Host layer; the
 * kit never imports this file.
 */
import { apiHttpPost, buildGameUrl, movePreviewBatch } from '../api.js';
import { decodeHostActionResponse } from '../types/host-action-response.js';
import { apiPath } from '../util.js';
import { surfaceForGame } from '../utils/companion-surface.js';
import type { TimerService } from '../timers/timer-service.js';
import type {
  BoardPreviewTransport,
  PlayerInfoModuleRequest,
  RendererGameRef,
  RendererHost,
  RendererModuleRequest,
  RendererStateQueue,
  TableActionResult,
  TableActions,
} from '../renderer-host/renderer-host.js';

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

/**
 * Loads a game's renderer module from the webapp's `game-src` tree, where the
 * build links each game's `client/` directory. The path is relative to this
 * file, which sits two levels below `server/static` like the components do.
 */
export function rendererModulePath(request: RendererModuleRequest): string {
  const suffix = request.surface === null ? '' : `-${request.surface}`;
  const base = `../../game-src/${request.gameName}/boardgame-render-game-${request.gameName}${suffix}.ts`;
  // A retry needs a new URL: a failed dynamic import is cached by specifier.
  return request.attempt > 0 ? `${base}?retry=${request.attempt}` : base;
}

/** The player-info renderer module for a game, in the same `game-src` tree. */
export function playerInfoModulePath(request: PlayerInfoModuleRequest): string {
  const base = `../../game-src/${request.gameName}/boardgame-render-player-info-${request.gameName}.ts`;
  return request.attempt > 0 ? `${base}?retry=${request.attempt}` : base;
}

export interface WebappRendererHostOptions {
  /** The state queue: the page's boardgame-game-state-manager. */
  readonly stateQueue?: RendererStateQueue;
  readonly timerService?: TimerService;
  /** Replaces the dynamic import, for tests that observe or stall module loads. */
  readonly importModule?: (modulePath: string) => Promise<unknown>;
}

/**
 * The webapp's RendererHost: the Go server for previews and table actions,
 * the page URL and tab storage for the companion surface, and dynamic import
 * from `game-src` for renderer modules.
 */
export function createWebappRendererHost(options: WebappRendererHostOptions = {}): RendererHost {
  // Use /* @vite-ignore */ to allow fully dynamic imports in dev mode.
  const importModule = options.importModule
    ?? ((modulePath: string) => import(/* @vite-ignore */ modulePath) as Promise<unknown>);
  return {
    ...(options.stateQueue ? { stateQueue: options.stateQueue } : {}),
    ...(options.timerService ? { timerService: options.timerService } : {}),
    resolveSurface: request => surfaceForGame(request.gameId, request.companionMode),
    loadRendererModule: request => importModule(rendererModulePath(request)),
    loadPlayerInfoModule: request => importModule(playerInfoModulePath(request)),
    boardPreview: webappBoardPreview,
    tableActions: webappTableActions,
  };
}
