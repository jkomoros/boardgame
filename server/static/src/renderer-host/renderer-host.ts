/**
 * The renderer-host contract: what whatever mounts a game renderer must
 * supply, written as types so a second host (a Fabric guest, a fixture) can be
 * checked against the same shape as the webapp.
 *
 * This file is renderer-kit code (layer 2, see scripts/layers.json). It must
 * not import the Redux store, the HTTP API, routing or anything else a
 * particular host is built from. The webapp's implementation lives in
 * `src/webapp/renderer-host.ts` and `boardgame-game-view`.
 */
import type { MovePreviewBatchResult } from '../legal/previewLegality.js';
import type { AnimatableRegistry } from '../motion/animatable-registry.js';
import type { VersionAnimationContext } from '../motion/timing.js';
import type { MovePreviewTransport, MoveSubmissionGate, MoveTransport } from '../moves/action.js';
import type { TargetPreviewTransport } from '../moves/target-action.js';
import type { TimerService } from '../timers/timer-service.js';

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

// ---- Which renderer to mount, and loading it ----

/** The companion surface a renderer is mounted for; null is the ordinary solo renderer. */
export type RendererSurface = 'table' | 'hand' | null;

export interface RendererSurfaceRequest {
  readonly gameId: string;
  /** The game's companion mode as last reported; undefined before it is known. */
  readonly companionMode: boolean | undefined;
}

export interface RendererModuleRequest {
  readonly gameName: string;
  readonly surface: RendererSurface;
  /** 0 for the first load of this renderer identity; a fresh positive number for each retry. */
  readonly attempt: number;
}

/** A request for a game's player-info renderer module. */
export interface PlayerInfoModuleRequest {
  readonly gameName: string;
  /** 0 for the first load; a fresh positive number for each retry. */
  readonly attempt: number;
}

// ---- The state queue ----

/**
 * The ordered queue of state versions. The host owns it and installs one
 * version at a time by setting the mount's state inputs; the mount reports
 * that a version's animations settled, and the host then releases the next.
 */
export interface RendererStateQueue {
  /**
   * The install cycle `cycleId` has settled (or may be cut over); release the
   * next queued version if there is one. Stale or repeated cycle ids are
   * ignored by the queue.
   */
  readyForNextState(cycleId?: number): void;
}

/** `detail` of the mount's `all-animations-done` and `motion-cycle-release` events. */
export interface RendererCycleEventDetail {
  readonly cycleId: number;
}

/**
 * What `boardgame-render-game` takes from its host as one object. Members are
 * stable services, not reactive data: reactive inputs (state, move forms, the
 * move transports) are the properties in `RendererMoveInputs` and the
 * element's other properties.
 *
 * Every member is optional so a partial host (a test, a static preview) still
 * mounts: without `loadRendererModule` no renderer loads and the element shows
 * a renderer error; without `resolveSurface` the solo renderer is used.
 */
export interface RendererHost {
  /**
   * The queue the mount's settlement is reported to. Today the mount announces
   * settlement with `all-animations-done` and `motion-cycle-release` events
   * (`RendererCycleEventDetail`) and the host forwards them here, because the
   * webapp decides per cycle whether local progress may advance a
   * server-scheduled companion slot.
   */
  readonly stateQueue?: RendererStateQueue;
  /**
   * The clock behind every `TimerController` in the renderer. The mount
   * answers timer-service requests from its subtree with this service.
   */
  readonly timerService?: TimerService;
  /** Which surface this browser presents for a game. Presentation intent only; it grants nothing. */
  resolveSurface?(request: RendererSurfaceRequest): RendererSurface;
  /**
   * Makes the game's renderer element class available: resolves once the
   * custom element `boardgame-render-game-<gameName>[-table|-hand]` is
   * defined, however the host gets the code (a dynamic import, a bundle).
   */
  loadRendererModule?(request: RendererModuleRequest): Promise<unknown>;
  /**
   * Where the framework's shipped token art (`token_<type>.svg`) is served
   * from, by token type. The mount installs it for layer-1 tokens; without it
   * a token with no art of its own draws its generated solid, or nothing.
   */
  resolveTokenArt?(type: string): string;
  /**
   * Makes the game's player-info renderer available: resolves once
   * `boardgame-render-player-info-<gameName>` is defined. A host mounts one
   * per seat with `mountPlayerInfo()` (components/boardgame-render-player-info.ts),
   * wherever its own layout puts player summaries; the webapp's roster is one
   * such place, not the only one.
   */
  loadPlayerInfoModule?(request: PlayerInfoModuleRequest): Promise<unknown>;
  /** Board-target legality preview. Absent means the host offers none, and no target is grayed by preview. */
  readonly boardPreview?: BoardPreviewTransport;
  /** Table actions for a Table surface. Absent means the controls report that they are unavailable. */
  readonly tableActions?: TableActions;
}

/**
 * The move-related reactive inputs a host binds on `boardgame-render-game`.
 * They are properties, not `RendererHost` members, because they change with
 * the connection and the viewer and must re-render the renderer when they do.
 */
export interface RendererMoveInputs {
  /** Carries a proposed move to the authority. Null while moves cannot be sent. */
  moveTransport: MoveTransport | null;
  /** Legality of one fully-specified move, without applying it. */
  movePreviewTransport: MovePreviewTransport | null;
  /** Legality of many targets of one move, without applying anything. */
  targetPreviewTransport: TargetPreviewTransport | null;
  /** Serializes submissions across every control in the renderer. */
  moveSubmissionGate: MoveSubmissionGate | null;
}

/**
 * What the mount (`boardgame-render-game`) itself provides: the animation
 * gate, and the registry and timing context that animatable items find by
 * walking up their ancestors. A host does not reimplement these; it uses them
 * to pipe in participants that live outside the mount's subtree.
 */
export interface RendererMount {
  /** Ambient registry animatable items self-register with. */
  readonly animatableRegistry: AnimatableRegistry;
  /** Scoped timing policy for the installed version; null outside companion scheduling. */
  animationContext: VersionAnimationContext | null;
  /** True while the animation gate is open for the current install cycle. */
  readonly isAnimating: boolean;
  /** Admit a participant outside the mount's subtree to the open gate. */
  gateWillAnimate(event: CustomEvent): void;
  /** Settle a participant admitted with `gateWillAnimate`. */
  gateAnimationDone(event: CustomEvent): void;
}
