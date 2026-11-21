/**
 * The guest renderer host.
 *
 * It does what `boardgame-game-state-manager` and `boardgame-game-view` do
 * for the webapp, with messages in place of Redux, HTTP and the socket:
 *
 *   - an ordered queue of versions, released one at a time as the real
 *     animation gate reports each cycle settled;
 *   - a timer service fed from the timers in each state entry;
 *   - move and legality-preview transports;
 *   - a mount point for the game's player-info renderer.
 *
 * It mounts the real `boardgame-render-game`, which owns the animator, the
 * effect layer and the gate. Every place that needed a workaround is marked
 * SHIM and written up in FINDINGS.md.
 */

import '@client/components/boardgame-render-game.js';
import '@client/components/boardgame-render-player-info.js';
import { BoardgameRenderGame } from '@client/components/boardgame-render-game.js';
import type { BoardgameRenderPlayerInfo } from '@client/components/boardgame-render-player-info.js';
import { expandGameStateSnapshot } from '@client/selectors.js';
import { clientMoveFromWire } from '@client/types/client-move.js';
import { validatePlayerPresentations } from '@client/status/player-presentation.js';
import {
  TimerService,
  TIMER_SERVICE_REQUEST_EVENT,
  type TimerServiceRequestDetail,
} from '@client/timers/timer-service.js';
import {
  MoveSubmissionGate,
  type MovePreviewTransport,
  type MoveSubmissionRequest,
  type MoveTransport,
  type MoveTransportResult,
} from '@client/moves/action.js';
import type { TargetPreviewTransport } from '@client/moves/target-action.js';

import {
  envelope,
  type GuestStateEntry,
  type GuestToHostMessage,
  type HostToGuestMessage,
  type InitMessage,
  type MessageBody,
  type MoveResultMessage,
  type StateMessage,
} from '../protocol.js';
import { VersionQueue } from './version-queue.js';

/** What the guest needs to know about a game whose renderer it bundles. */
export interface BundledGame {
  /** The move-input schema fingerprint compiled into the renderer. */
  readonly moveInputSchemaFingerprint: string;
  /** Whether a `boardgame-render-player-info-<name>` element is bundled. */
  readonly hasPlayerInfo: boolean;
}

/** An entry plus the guest-clock instant it arrived, for its timers. */
interface QueuedEntry extends GuestStateEntry {
  readonly receivedAt: number;
}

interface PendingMove {
  readonly snapshotVersion: number;
  readonly resolve: (result: MoveTransportResult) => void;
  readonly timeout: ReturnType<typeof setTimeout>;
}

interface RunningTimer {
  readonly version: number;
  readonly original: number;
  readonly timeLeftAtReceipt: number;
  readonly receivedAt: number;
  ticked: boolean;
}

/** How long a proposed move may go unanswered before the action reports failure. */
const MOVE_RESPONSE_TIMEOUT_MS = 10_000;

// SHIM (FINDINGS F1): boardgame-render-game loads a game's renderer with a
// path-computed dynamic import(). A guest has nothing to fetch, and every
// renderer it can show is already registered by a static import, so the
// loader resolves immediately. `_loadRendererModule` is TypeScript-private
// but an ordinary prototype method at run time.
(BoardgameRenderGame.prototype as unknown as {
  _loadRendererModule: (modulePath: string) => Promise<unknown>;
})._loadRendererModule = () => Promise.resolve();

export class GuestHost {
  private readonly root: HTMLElement;
  private readonly post: (message: GuestToHostMessage) => void;
  private readonly games: Readonly<Record<string, BundledGame>>;

  private render: BoardgameRenderGame | null = null;
  private seats: HTMLElement | null = null;
  private playerInfos: BoardgameRenderPlayerInfo[] = [];

  private init: InitMessage | null = null;
  private mounted = false;
  private viewer = 0;
  private epoch = 0;

  private readonly queue = new VersionQueue<QueuedEntry>();
  private current: QueuedEntry | null = null;
  /** The version shown before `current`, or null if `current` was first. */
  private previousVersion: number | null = null;
  /** True from installing a version until its cycle is released. */
  private busy = false;
  private cycleSequence = 0;
  private activeCycleId = 0;
  /** The cycle that showed a first state, which nothing animates into. */
  private initialCycleId = 0;
  private releasedCycleId = 0;
  private snapshotEpoch = 0;
  /** Bumped by every reset so stale async continuations can tell. */
  private generation = 0;

  private readonly timerService = new TimerService();
  private timers = new Map<string, RunningTimer>();
  private timerFrame: number | null = null;

  private readonly pendingMoves = new Map<string, PendingMove>();
  private readonly submissionGate = new MoveSubmissionGate();

  constructor(
    root: HTMLElement,
    post: (message: GuestToHostMessage) => void,
    games: Readonly<Record<string, BundledGame>>,
  ) {
    this.root = root;
    this.post = post;
    this.games = games;
    // The renderer kit finds its timer service by event, as in the webapp.
    root.addEventListener(TIMER_SERVICE_REQUEST_EVENT, (event: Event) => {
      const request = event as CustomEvent<TimerServiceRequestDetail>;
      if (typeof request.detail?.accept !== 'function') return;
      event.stopPropagation();
      request.detail.accept(this.timerService);
    });
    root.addEventListener('all-animations-done', event => this.cycleReleased(event));
    // A renderer may declare that a buffered successor can cut in early.
    root.addEventListener('motion-cycle-release', event => this.cycleReleased(event));
    // The legacy path: an element that announces a move as a DOM event
    // instead of going through a bound action and the transport.
    root.addEventListener('propose-move', event => this.legacyProposal(event));
  }

  get bundledGames(): readonly string[] {
    return Object.keys(this.games);
  }

  /** The version currently on screen, or null before the first one. */
  get shownVersion(): number | null {
    return this.current?.version ?? null;
  }

  handle(message: HostToGuestMessage): void {
    switch (message.type) {
      case 'init': this.handleInit(message); return;
      case 'state': this.handleState(message); return;
      case 'move-result': this.handleMoveResult(message); return;
    }
  }

  private send(body: MessageBody<GuestToHostMessage>): void {
    this.post(envelope<GuestToHostMessage>(body));
  }

  // -------------------------------------------------------------------------
  // init
  // -------------------------------------------------------------------------

  private handleInit(message: InitMessage): void {
    const game = this.games[message.gameName];
    if (!game) {
      this.send({
        type: 'error',
        message: `this guest has no renderer for ${JSON.stringify(message.gameName)}; it bundles ${this.bundledGames.join(', ')}`,
      });
      return;
    }
    const presentations = validatePlayerPresentations(message.players);
    this.resetPlayback();
    this.queue.clear();
    this.init = message;
    this.viewer = message.viewer;
    this.mounted = false;
    const generation = this.generation;

    this.root.replaceChildren();
    // The webapp's game view publishes seat colours as custom properties.
    presentations.forEach((presentation, index) => {
      if (presentation.color) this.root.style.setProperty(`--player-${index}-color`, presentation.color);
      else this.root.style.removeProperty(`--player-${index}-color`);
    });

    const seats = document.createElement('div');
    seats.className = 'guest-seats';
    this.seats = seats;
    this.playerInfos = [];
    if (game.hasPlayerInfo) {
      for (const presentation of presentations) {
        const seat = document.createElement('div');
        seat.className = 'guest-seat';
        seat.dataset['player'] = String(presentation.playerIndex);
        const name = document.createElement('span');
        name.className = 'guest-seat-name';
        name.textContent = presentation.label;
        if (presentation.color) name.style.setProperty('--seat-color', presentation.color);
        // SHIM (FINDINGS F7): in the webapp only the Redux-bound roster
        // mounts this element. It needs no store, so the guest mounts it
        // directly and tells it the renderer is already registered.
        const info = document.createElement('boardgame-render-player-info') as BoardgameRenderPlayerInfo;
        info.gameName = message.gameName;
        info.playerIndex = presentation.playerIndex;
        info.rendererLoaded = true;
        info.active = true;
        seat.append(name, info);
        seats.append(seat);
        this.playerInfos.push(info);
      }
    }

    const render = document.createElement('boardgame-render-game') as BoardgameRenderGame;
    render.id = 'render';
    // SHIM (FINDINGS F3): with a game id set, render-game asks
    // companion-surface which surface this browser wants, and that reads
    // location.search, sessionStorage and document.cookie. Reading
    // document.cookie throws in an opaque origin. Declaring the table solo
    // (CompanionMode false) makes that function return before it reads
    // anything.
    render.companionInfo = { CompanionMode: false } as unknown as BoardgameRenderGame['companionInfo'];
    render.gameName = message.gameName;
    // SHIM (FINDINGS F2): with an empty game id the legality-preview refresh
    // returns before it reaches `movePreviewBatch`. The id is only used
    // otherwise as an effect seed and in the transition key.
    render.gameId = '';
    render.chest = message.chest as unknown as BoardgameRenderGame['chest'];
    render.playerPresentations = presentations;
    render.viewingAsPlayer = message.viewer;
    render.proposingAsPlayer = message.viewer;
    render.previewAsPlayer = message.viewer;
    render.moveInputSchemaFingerprint = message.moveInputSchemaFingerprint ?? game.moveInputSchemaFingerprint;
    render.moveTransport = this.moveTransport;
    render.movePreviewTransport = this.movePreviewTransport;
    render.targetPreviewTransport = this.targetPreviewTransport;
    render.moveSubmissionGate = this.submissionGate;
    // SHIM (FINDINGS F5): "socket active" is render-game's word for "the
    // host is live". False withholds the move transport and shows a
    // reconnecting banner, so a guest is always live.
    render.socketActive = true;
    render.active = true;
    this.render = render;
    this.root.append(seats, render);

    void this.waitForRenderer(render).then(() => {
      if (generation !== this.generation || this.render !== render) return;
      this.mounted = true;
      this.send({ type: 'ready', stage: 'mounted', games: this.bundledGames });
      this.pump();
    }, (error: unknown) => {
      this.send({ type: 'error', message: `renderer failed to mount: ${String(error)}` });
    });
  }

  private async waitForRenderer(render: BoardgameRenderGame): Promise<void> {
    for (let attempt = 0; attempt < 200; attempt++) {
      await render.updateComplete;
      if (render.renderer) {
        await render.renderer.updateComplete;
        return;
      }
      if (render.rendererError) throw new Error(render.rendererError);
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
    }
    throw new Error('renderer was not instantiated');
  }

  // -------------------------------------------------------------------------
  // state
  // -------------------------------------------------------------------------

  private handleState(message: StateMessage): void {
    if (!this.init) {
      this.send({ type: 'error', message: 'state arrived before init' });
      return;
    }
    const epoch = message.epoch ?? 0;
    const receivedAt = performance.now();
    const key = `${epoch}:${message.viewer}`;
    let result;
    try {
      result = this.queue.offer(key, message.entries.map(entry => ({ ...entry, receivedAt })));
    } catch (error) {
      this.send({ type: 'error', message: String(error) });
      return;
    }
    if (result.reset || message.viewer !== this.viewer || epoch !== this.epoch) {
      // A different seat or a different history: nothing on screen is a
      // valid "before" for what comes next, so cut instead of animating.
      this.resetPlayback();
      this.viewer = message.viewer;
      this.epoch = epoch;
    }
    if (result.accepted > 0) this.settleMovesThrough(this.queue.lastQueued ?? -1);
    this.pump();
  }

  /** Drop what is on screen and every in-flight cycle, keeping the mount. */
  private resetPlayback(): void {
    this.generation++;
    this.busy = false;
    this.current = null;
    this.previousVersion = null;
    this.activeCycleId = 0;
    this.releasedCycleId = 0;
    this.setTimers(null);
    for (const [requestId, pending] of this.pendingMoves) {
      clearTimeout(pending.timeout);
      pending.resolve({ kind: 'network-failure', error: 'the view was reset' });
      this.pendingMoves.delete(requestId);
    }
    if (this.render) {
      // A null state makes render-game treat the next one as initial.
      this.render.state = null;
      this.render.transitionMove = null;
    }
    for (const info of this.playerInfos) info.state = null;
  }

  /** Install the next version if nothing is in flight. */
  private pump(): void {
    if (!this.mounted || this.busy || !this.render) return;
    const next = this.queue.peek();
    if (!next) return;

    const renderer = this.render.renderer;
    const fromMove = this.current ? clientMoveFromWire(this.current.move) : null;
    const toMove = clientMoveFromWire(next.move);
    if (renderer && this.current && (fromMove || toMove)) {
      // The renderer's per-transition duration hook, as the state manager
      // consults it. Negative means "do not show this intermediate version".
      const length = renderer.animationLength(fromMove, toMove);
      if (length < 0 && this.queue.length > 1) {
        this.queue.shift();
        this.send({
          type: 'animations-done', version: next.version, viewer: this.viewer,
          epoch: this.epoch, remaining: this.queue.length, how: 'skipped',
        });
        this.pump();
        return;
      }
      this.render.defaultAnimationLength = Math.max(0, length);
    }

    this.queue.shift();
    this.busy = true;
    const initial = this.current === null;
    const cycleId = ++this.cycleSequence;
    this.activeCycleId = cycleId;
    this.initialCycleId = initial ? cycleId : 0;
    this.install(next, cycleId);

    if (initial) {
      // render-game opens no animation cycle for the first state after
      // null, so all-animations-done arrives only if something in the first
      // render happened to animate. Do not depend on it: release once the
      // first render has painted.
      const generation = this.generation;
      void this.afterFirstPaint().then(() => {
        if (generation !== this.generation) return;
        this.release(cycleId);
      });
    }
  }

  private async afterFirstPaint(): Promise<void> {
    const render = this.render;
    if (!render) return;
    await render.updateComplete;
    await render.renderer?.updateComplete;
    await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
  }

  private install(entry: QueuedEntry, cycleId: number): void {
    const render = this.render;
    const init = this.init;
    if (!render || !init) return;
    this.previousVersion = this.current?.version ?? null;
    this.current = entry;
    this.snapshotEpoch += 1;
    this.setTimers(entry);

    // The same expansion the webapp runs: stacks gain their components from
    // the chest. SHIM (FINDINGS F8): it is imported from selectors.ts, the
    // Redux selector module, because that is where the pure function lives.
    const expanded = expandGameStateSnapshot(entry.state as never, init.chest as never, init.gameName);

    // All of these land in one Lit update, in the order render-game's
    // updated() applies them; the state assignment is what starts the cycle.
    render.transitionMove = clientMoveFromWire(entry.move);
    render.motionCycleId = cycleId;
    render.gameVersion = entry.version;
    render.snapshotEpoch = this.snapshotEpoch;
    render.viewingAsPlayer = this.viewer;
    render.proposingAsPlayer = this.viewer;
    render.previewAsPlayer = this.viewer;
    render.currentPlayerIndex = entry.currentPlayerIndex;
    render.gameFinished = entry.outcome.finished;
    render.gameWinners = [...entry.outcome.winners];
    // SHIM (FINDINGS F12): Go omits an empty `Sets`, and the only code that
    // puts it back is the decoder for whole HTTP responses.
    render.projectedMoveChoicesWire = (entry.projectedMoveChoices
      ? { ...entry.projectedMoveChoices, Sets: entry.projectedMoveChoices.Sets ?? [] }
      : null) as BoardgameRenderGame['projectedMoveChoicesWire'];
    // SHIM (FINDINGS F6): render-game only accepts legality wrapped in the
    // HTTP API's move forms, so rebuild the three fields it reads.
    render.moveForms = Object.entries(entry.moveLegality ?? {}).map(([name, legality]) => ({
      Name: name,
      LegalForPlayer: legality.legalForPlayer,
      LegalForAnyone: legality.legalForAnyone,
      ...(legality.error ? { LegalForPlayerError: legality.error } : {}),
    })) as unknown as BoardgameRenderGame['moveForms'];
    render.state = expanded as unknown as BoardgameRenderGame['state'];
    for (const info of this.playerInfos) info.state = expanded as unknown as BoardgameRenderPlayerInfo['state'];
    this.root.dataset['version'] = String(entry.version);
    this.root.dataset['viewer'] = String(this.viewer);
  }

  private cycleReleased(event: Event): void {
    const cycleId = (event as CustomEvent<{ cycleId?: unknown }>).detail?.cycleId;
    if (typeof cycleId !== 'number') return;
    this.release(cycleId);
  }

  private release(cycleId: number): void {
    // Exactly one release per installed cycle; render-game also closes an
    // interrupted cycle under its old id, which must not advance the queue.
    if (cycleId !== this.activeCycleId || cycleId === this.releasedCycleId) return;
    this.releasedCycleId = cycleId;
    this.busy = false;
    const shown = this.current;
    const how = cycleId === this.initialCycleId ? 'initial' : 'played';
    if (shown) {
      const previous = this.previousVersion;
      this.send({
        type: 'animations-done', version: shown.version, viewer: this.viewer,
        epoch: this.epoch, remaining: this.queue.length, how,
        ...(previous !== null && shown.version !== previous + 1 ? { jumpedFrom: previous } : {}),
      });
    }
    this.pump();
  }

  // -------------------------------------------------------------------------
  // timers
  // -------------------------------------------------------------------------

  private setTimers(entry: QueuedEntry | null): void {
    const next = new Map<string, RunningTimer>();
    for (const [timerId, info] of Object.entries(entry?.timers ?? {})) {
      const previous = this.timers.get(timerId);
      next.set(timerId, {
        version: entry!.version,
        // The duration is whatever was left the first time the timer was seen.
        original: Math.max(previous?.original ?? info.TimeLeft, info.TimeLeft),
        timeLeftAtReceipt: info.TimeLeft,
        receivedAt: entry!.receivedAt,
        ticked: previous?.ticked ?? false,
      });
    }
    this.timers = next;
    this.updateTimers();
  }

  private updateTimers(): void {
    this.timerFrame = null;
    const now = performance.now();
    const infos: Record<string, { TimeLeft: number; originalTimeLeft: number }> = {};
    let running = false;
    for (const [timerId, timer] of this.timers) {
      const left = Math.max(0, timer.timeLeftAtReceipt - (now - timer.receivedAt));
      infos[timerId] = { TimeLeft: left, originalTimeLeft: timer.original };
      if (left > 0) {
        running = true;
      } else if (!timer.ticked) {
        timer.ticked = true;
        // The host has no clock of its own to wake it; ask it to look.
        this.send({ type: 'tick', timerId, version: timer.version });
      }
    }
    this.timerService.update(infos);
    if (running) this.timerFrame = requestAnimationFrame(() => this.updateTimers());
  }

  // -------------------------------------------------------------------------
  // moves
  // -------------------------------------------------------------------------

  private readonly moveTransport: MoveTransport = {
    submit: request => this.submit(request),
  };

  private submit(request: MoveSubmissionRequest): Promise<MoveTransportResult> {
    const current = this.current;
    if (!current || request.snapshotVersion !== current.version || this.queue.length > 0) {
      // What is on screen is already behind what the host has sent.
      return Promise.resolve({
        kind: 'server-rejection',
        error: 'the board has moved on',
        code: 'STALE_SNAPSHOT',
        expectedVersion: request.snapshotVersion,
        actualVersion: this.queue.lastQueued ?? current?.version ?? 0,
      });
    }
    return new Promise<MoveTransportResult>(resolve => {
      const timeout = setTimeout(() => {
        this.pendingMoves.delete(request.requestID);
        resolve({ kind: 'network-failure', error: 'the host did not answer the move' });
      }, MOVE_RESPONSE_TIMEOUT_MS);
      this.pendingMoves.set(request.requestID, {
        snapshotVersion: request.snapshotVersion, resolve, timeout,
      });
      this.send({
        type: 'propose-move',
        requestId: request.requestID,
        name: request.name,
        arguments: { ...request.arguments },
        snapshotVersion: request.snapshotVersion,
        viewer: this.viewer,
      });
    });
  }

  /** A newer version arriving is the ordinary sign that a move was applied. */
  private settleMovesThrough(version: number): void {
    for (const [requestId, pending] of this.pendingMoves) {
      if (version <= pending.snapshotVersion) continue;
      clearTimeout(pending.timeout);
      this.pendingMoves.delete(requestId);
      pending.resolve({ kind: 'success' });
    }
  }

  private handleMoveResult(message: MoveResultMessage): void {
    const pending = this.pendingMoves.get(message.requestId);
    if (!pending) return;
    clearTimeout(pending.timeout);
    this.pendingMoves.delete(message.requestId);
    if (message.ok) {
      pending.resolve({ kind: 'success' });
      return;
    }
    pending.resolve({
      kind: 'server-rejection',
      error: message.error ?? 'the move was refused',
      ...(message.actualVersion !== undefined ? {
        code: 'STALE_SNAPSHOT',
        expectedVersion: pending.snapshotVersion,
        actualVersion: message.actualVersion,
      } : {}),
    });
  }

  private legacyProposal(event: Event): void {
    const detail = (event as CustomEvent<{ name?: unknown; arguments?: unknown }>).detail;
    const current = this.current;
    if (!current || typeof detail?.name !== 'string') return;
    // As the webapp's game view does, drop a move made mid-animation.
    if (this.busy || this.queue.length > 0) return;
    const args: Record<string, string> = {};
    for (const [key, value] of Object.entries((detail.arguments ?? {}) as Record<string, unknown>)) {
      args[key] = String(value);
    }
    this.send({
      type: 'propose-move',
      requestId: `guest-legacy-${current.version}-${Date.now()}`,
      name: detail.name,
      arguments: args,
      snapshotVersion: current.version,
      viewer: this.viewer,
    });
  }

  // -------------------------------------------------------------------------
  // legality preview, answered from the entry on screen
  // -------------------------------------------------------------------------

  private legal(name: string, input: Readonly<Record<string, string>>): boolean {
    const entry = this.current;
    if (!entry || entry.outcome.finished) return false;
    const base = entry.moveLegality?.[name]?.legalForPlayer ?? false;
    const fields = Object.keys(input);
    if (fields.length === 0) return base;
    const choices = entry.projectedMoveChoices;
    const set = choices?.Status === 'ready' && choices.StateVersion === entry.version
      ? choices.Sets?.find(candidate => candidate.MoveName === name)
      : undefined;
    const field = fields[0];
    if (set && fields.length === 1 && field === set.FieldName) {
      return set.Candidates.find(candidate => String(candidate.Value) === input[set.FieldName])?.Available ?? false;
    }
    // No per-candidate fact was sent for this input. The default-bound
    // verdict is the best the guest has; the host remains the judge.
    return base;
  }

  private stale(snapshotVersion: number): { kind: 'stale-snapshot'; expectedVersion: number; actualVersion: number } | null {
    const actual = this.current?.version ?? -1;
    return snapshotVersion === actual ? null : { kind: 'stale-snapshot', expectedVersion: snapshotVersion, actualVersion: actual };
  }

  private readonly movePreviewTransport: MovePreviewTransport = {
    preview: async request => this.stale(request.snapshotVersion) ?? {
      kind: 'success',
      legal: this.legal(request.name, request.arguments),
    },
  };

  private readonly targetPreviewTransport: TargetPreviewTransport = {
    previewTargets: async request => this.stale(request.snapshotVersion) ?? {
      kind: 'success',
      results: request.candidates.map(candidate => ({
        id: candidate.id,
        legal: this.legal(request.name, candidate.arguments),
      })),
    },
  };
}
