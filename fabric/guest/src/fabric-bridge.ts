/**
 * The guest's side of a Common Fabric `cf-iframe`.
 *
 * A pattern hands the guest these named resources (see
 * `fabric/patterns/memory/main.tsx` and `fabric/patterns/blackjack/main.tsx`):
 *
 *   view      cell, read-only   what whoever only watches may see (`TableView`)
 *   seat<N>   cell, read-only   seat N's own view, for N below `info.seatViews`
 *   info      cell, read-only   the game's public facts (`TableInfo`)
 *   me        cell, read-only   the viewer's own principal
 *   move      stream            a `Proposal` to the table's one writer
 *   identify  stream            asks the pattern to fill in `me`
 *
 * A game whose seats all see the same (Memory) hands over `view` alone and
 * says `seatViews: 0`. A game with hidden hands (blackjack) also hands over
 * one view a seat. The bridge reads `view` for who is seated, finds the
 * viewer's seat from `me`, and then reads that one seat's cell and no other:
 * it never asks for a seat that is not the viewer's.
 *
 * This module turns those into the message protocol of `PROTOCOL.md`, which
 * the renderer host (`GuestHost`) already speaks: the latest value of the
 * view becomes an `init` (when the table, the viewer or the deal changes) and
 * a `state` window; a `propose-move` or a `tick` from the host becomes an
 * event on `move`. Before the deal it draws the seats itself, with the
 * game's name and setup options from `info`.
 *
 * The proposer is never sent. The pattern takes it from the runtime.
 */
import { connectFabric, type CellReadRefusal } from '@commonfabric/iframe-sandbox/guest';

import {
  envelope,
  type GuestStateEntry,
  type GuestToHostMessage,
  type HostToGuestMessage,
  type MessageBody,
} from '../protocol.js';
import type { Proposal, TableInfo, TableView, ViewEntry } from '../../patterns/toolkit/view.js';
import type { GuestHost } from './guest-host.js';

type AnyState = GuestStateEntry['state'];
type View = TableView<AnyState>;
type Info = TableInfo<unknown>;

const OBSERVER = -1;

/** How often, and how many times, a refused tick is sent again. */
const TICK_RETRY_MS = 700;
const TICK_RETRIES = 8;

export const FABRIC_CSS = `
  #fabric-status { padding: 8px 12px; font-size: 14px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap;
    border-bottom: 1px solid var(--md-sys-color-outline-variant, #ccc); }
  #fabric-status .who { font-weight: 600; }
  #fabric-status .note { opacity: 0.75; }
  #fabric-status .error { color: var(--md-sys-color-error, #b3261e); }
  #fabric-lobby { padding: 24px 16px; max-width: 420px; margin: 0 auto; display: flex; flex-direction: column; gap: 12px; }
  #fabric-lobby h1 { font-family: var(--md-sys-typescale-display-small-font); font-weight: 400; margin: 0; }
  #fabric-lobby ol { margin: 0; padding-left: 20px; }
  #fabric-lobby .row { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  #fabric-lobby input, #fabric-lobby select, #fabric-lobby button, #fabric-status button {
    font: inherit; padding: 6px 12px; border-radius: 8px; border: 1px solid var(--md-sys-color-outline, #777);
    background: var(--md-sys-color-surface, #fff); color: inherit; }
  #fabric-lobby button, #fabric-status button { cursor: pointer; background: var(--md-sys-color-primary, #2b5); color: var(--md-sys-color-on-primary, #fff); border-color: transparent; }
  #fabric-lobby button:disabled { opacity: 0.4; cursor: default; }
  [hidden] { display: none !important; }
`;

function element<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  properties: Partial<HTMLElementTagNameMap[K]> = {},
  ...children: (Node | string)[]
): HTMLElementTagNameMap[K] {
  const node = Object.assign(document.createElement(tag), properties);
  node.append(...children);
  return node;
}

/** A private, mutable copy, as `postMessage` would have handed the host. */
function copy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

/** A view entry as the renderer host takes one, for the seat `viewer`. */
function entryFor(entry: Readonly<ViewEntry<AnyState>>, viewer: number): GuestStateEntry {
  const part = entry.viewers[String(viewer)] ?? entry.viewers[String(OBSERVER)];
  return copy({
    version: entry.version,
    state: entry.state,
    move: entry.move ?? null,
    currentPlayerIndex: entry.currentPlayerIndex,
    timers: entry.timers,
    outcome: entry.outcome,
    // A version a fix-up replaced at once offers nothing to do.
    ...(entry.settled && part
      ? {
        moveLegality: part.moveLegality,
        zeroInputMoves: part.zeroInputMoves,
        projectedMoveChoices: part.projectedMoveChoices,
      }
      : {}),
  }) as GuestStateEntry;
}

/**
 * What the bridge saw, for tests and for measuring: each distinct view as it
 * arrived, and each proposal as it was sent and settled. Bounded, and read
 * only from outside (`window.__boardgameFabricLog`).
 */
interface BridgeLogEntry {
  /** `performance.now()`, and the wall clock for comparing two guests. */
  readonly at: number;
  readonly wall: number;
  readonly what: 'view' | 'sent' | 'settled' | 'refused';
  readonly revision?: number;
  readonly epoch?: number;
  readonly version?: number;
  readonly kind?: string;
  readonly detail?: string;
}
const bridgeLog: BridgeLogEntry[] = [];
(globalThis as { __boardgameFabricLog?: BridgeLogEntry[] }).__boardgameFabricLog = bridgeLog;
function logged(entry: Omit<BridgeLogEntry, 'at' | 'wall'>): void {
  if (bridgeLog.length >= 2000) bridgeLog.shift();
  bridgeLog.push({ at: performance.now(), wall: Date.now(), ...entry });
}

function describe(cause: unknown): string {
  return cause instanceof Error ? cause.message : String(cause);
}

/**
 * Connects to the embedding pattern and drives `host`, which is mounted in
 * `board`. `makeHost` receives the function the host posts its messages to.
 */
export function startFabricBridge(
  board: HTMLElement,
  makeHost: (post: (message: GuestToHostMessage) => void) => GuestHost,
): void {
  const fabric = connectFabric();
  const viewCell = fabric.cell<View>('view');
  const infoCell = fabric.cell<Info>('info');
  const meCell = fabric.cell<string>('me');
  const move = fabric.stream<Proposal>('move');
  const identify = fabric.stream<void>('identify');

  // A diagnostic, read only from outside: whether the host hands over the
  // part of the view at `path`, and what it holds. FINDINGS.md of the Memory
  // pattern used it to find which parts of a view the display gate refuses.
  (globalThis as { __boardgameFabricProbe?: unknown }).__boardgameFabricProbe = async (path: (string | number)[]) => {
    try {
      const cell = path.length === 0 ? viewCell : (viewCell as unknown as {
        key: (...keys: (string | number)[]) => { pull(): Promise<unknown> };
      }).key(...path);
      const value = await cell.pull();
      return {
        ok: true,
        kind: Array.isArray(value) ? 'array' : value === null ? 'null' : typeof value,
        keys: value !== null && typeof value === 'object' ? Object.keys(value as object) : [],
      };
    } catch (cause) {
      return { ok: false, error: describe(cause) };
    }
  };

  // ----- document ----------------------------------------------------------

  const who = element('span', { className: 'who' });
  const note = element('span', { className: 'note' });
  const problem = element('span', { className: 'error' });
  const again = element('button', { type: 'button', textContent: 'New game', hidden: true });
  const status = element('div', { id: 'fabric-status' }, who, note, again, problem);

  const seatList = element('ol');
  const nameField = element('input', { placeholder: 'Your name', maxLength: 24, autocomplete: 'off' });
  const sit = element('button', { type: 'button', textContent: 'Sit down' });
  const leave = element('button', { type: 'button', textContent: 'Stand up' });
  // One select for each setup option the game declares in `info`.
  const options = element('div', { className: 'row' });
  const start = element('button', { type: 'button', textContent: 'Deal' });
  const lobbyNote = element('div', { className: 'note' });
  const title = element('h1');
  const lobby = element('div', { id: 'fabric-lobby', hidden: true },
    title,
    seatList,
    element('div', { className: 'row' }, nameField, sit, leave),
    options,
    element('div', { className: 'row' }, start),
    lobbyNote);
  document.body.insertBefore(status, board);
  document.body.insertBefore(lobby, board);

  // ----- state -------------------------------------------------------------

  let hydrated = false;
  let refusal: CellReadRefusal | undefined;
  /** What the current `init` was computed from; a change needs a new one. */
  let initKey = '';
  let sending = false;
  let lastStamp = '';
  /** The setup options the lobby's selects were last built from. */
  let optionsKey = '';
  /** The viewer's own seat's cell, when the game has one a seat. */
  let ownSeat = OBSERVER;
  let ownCell: ReturnType<typeof fabric.cell<View>> | undefined;
  let stopOwn: (() => void) | undefined;
  let ownRefusal: CellReadRefusal | undefined;

  const host = makeHost(message => onHostMessage(message));

  function toHost(body: MessageBody<HostToGuestMessage>): void {
    host.handle(envelope<HostToGuestMessage>(body));
  }

  function seatOf(view: Readonly<View> | undefined): number {
    const me = meCell.get();
    if (!view || typeof me !== 'string' || me === '') return OBSERVER;
    return view.seats.findIndex(seat => seat.principal === me);
  }

  /**
   * Reads the one seat's view that is the viewer's, and stops reading any
   * other. `seat` is `OBSERVER` for a watcher and for a game with one view.
   */
  function readOwnSeat(seat: number): void {
    if (seat === ownSeat) return;
    stopOwn?.();
    ownSeat = seat;
    ownCell = undefined;
    stopOwn = undefined;
    ownRefusal = undefined;
    if (seat === OBSERVER) return;
    const cell = fabric.cell<View>(`seat${seat}`);
    ownCell = cell;
    stopOwn = cell.sink(() => { ownRefusal = undefined; sync(); }, {
      onRefused: cause => { ownRefusal = cause; sync(); },
    });
    void cell.pull().then(() => sync(), cause => {
      problem.textContent = describe(cause);
    });
  }

  /** The cell the board is being drawn from right now. */
  function sourceCell(): ReturnType<typeof fabric.cell<View>> {
    return ownCell !== undefined && ownCell.get() !== undefined ? ownCell : viewCell;
  }

  // ----- pattern to host ---------------------------------------------------

  function sync(): void {
    problem.textContent = '';
    if (!hydrated) {
      who.textContent = 'Loading…';
      return;
    }
    // `watch` is what everyone may see: it says who is seated, and so which
    // seat is the viewer's.
    const watch = viewCell.get();
    const info = infoCell.get();
    const seat = seatOf(watch);
    const seatViews = info?.seatViews ?? 0;
    readOwnSeat(seat !== OBSERVER && seat < seatViews ? seat : OBSERVER);
    // The seat's own view, once it has arrived for this deal. Until then a
    // seated player is shown what a watcher is, as a watcher: nothing is ever
    // drawn for a viewer it was not computed for.
    const own = ownCell?.get();
    const ownReady = own !== undefined && watch !== undefined && own.epoch === watch.epoch &&
      own.phase !== 'seating';
    const view = ownReady ? own : watch;
    const viewer = seatViews > 0 && !ownReady ? OBSERVER : seat;
    again.hidden = true;
    const stamp = `${viewer}:${view?.revision}:${view?.epoch}:${view?.version}`;
    if (view && stamp !== lastStamp) {
      lastStamp = stamp;
      logged({ what: 'view', revision: view.revision, epoch: view.epoch, version: view.version, detail: `viewer ${viewer}` });
    }

    if (!view || view.phase === 'seating' || !info) {
      initKey = '';
      board.hidden = true;
      lobby.hidden = false;
      drawLobby(view, seat, info);
      if (refusal) {
        // The runtime would not hand over the view. Before the table's
        // first proposal there may be nothing it can release, so sitting
        // down stays possible; after that this says why nothing shows.
        note.textContent = `The table's view was withheld by ${refusal.refusedBy ?? 'the runtime'}` +
          (refusal.labelSource ? ` (${refusal.labelSource}).` : '.');
      }
      return;
    }

    lobby.hidden = true;
    board.hidden = false;
    const key = `${view.gameName}:${view.epoch}:${viewer}:${view.seats.length}`;
    if (key !== initKey) {
      initKey = key;
      toHost({
        type: 'init',
        gameName: view.gameName,
        chest: copy(info.chest) as never,
        viewer,
        players: view.seats.map((held, index) => ({
          playerIndex: index,
          label: held.label,
          color: info.seatColors[index % info.seatColors.length],
        })),
        // The engine's own fingerprint of its move inputs, so the renderer
        // checks itself against the table it is actually playing at.
        moveInputSchemaFingerprint: info.moveInputSchemaFingerprint,
      });
    }
    if (view.entries.length > 0) {
      toHost({
        type: 'state',
        viewer,
        epoch: view.epoch,
        entries: view.entries.map(entry => entryFor(entry, viewer)),
      });
    }

    const head = view.entries[view.entries.length - 1];
    const turn = head ? view.seats[head.currentPlayerIndex]?.label ?? '' : '';
    who.textContent = seat === OBSERVER ? 'Watching' : `You are ${view.seats[seat]?.label ?? ''}`;
    if (view.phase === 'finished' && head) {
      const winners = head.outcome.winners.map(index => view.seats[index]?.label ?? `Player ${index + 1}`);
      note.textContent = winners.length === 1 ? `${winners[0]} won.` : `A tie: ${winners.join(' and ')}.`;
      again.hidden = seat === OBSERVER;
    } else if (head) {
      note.textContent = head.currentPlayerIndex === seat ? 'Your turn.' : `${turn}'s turn.`;
    }
    if (seat !== OBSERVER && viewer === OBSERVER) {
      note.textContent = ownRefusal
        ? `Your own view was withheld by ${ownRefusal.refusedBy ?? 'the runtime'}; you are shown what a watcher is.`
        : 'Waiting for your own view…';
    }
  }

  function drawLobby(view: Readonly<View> | undefined, seat: number, info: Readonly<Info> | undefined): void {
    title.textContent = info?.displayName ?? '';
    const key = JSON.stringify(info?.setupOptions ?? []);
    if (key !== optionsKey) {
      optionsKey = key;
      options.replaceChildren(...(info?.setupOptions ?? []).map(option => {
        const select = element('select', { name: option.name, title: option.displayName },
          ...option.values.map(value => element('option', { value: value.name, textContent: value.displayName })));
        select.value = option.default;
        return element('label', {}, `${option.displayName} `, select);
      }));
    }
    const seats = view?.seats ?? [];
    seatList.replaceChildren(...seats.map((held, index) =>
      element('li', { textContent: held.label + (index === seat ? ' (you)' : '') })));
    const min = view?.minPlayers ?? 2;
    const max = view?.maxPlayers ?? 6;
    const known = typeof meCell.get() === 'string' && meCell.get() !== '';
    who.textContent = seat === OBSERVER ? 'Not seated' : `You are ${seats[seat]?.label ?? ''}`;
    note.textContent = '';
    nameField.hidden = seat !== OBSERVER;
    sit.hidden = seat !== OBSERVER;
    // Without a view there is no telling who is seated; the table decides.
    sit.disabled = sending || (view !== undefined && !known) || seats.length >= max;
    leave.hidden = seat === OBSERVER;
    leave.disabled = sending;
    start.disabled = sending || seat === OBSERVER || seats.length < min;
    for (const select of options.querySelectorAll('select')) select.disabled = start.disabled;
    lobbyNote.textContent = seats.length === 0
      ? 'Nobody is seated yet.'
      : seats.length < min
      ? `Waiting for ${min - seats.length} more to sit.`
      : seat === OBSERVER
      ? 'Sit down to play, or watch.'
      : 'Deal when everyone is seated.';
  }

  // ----- host to pattern ---------------------------------------------------

  /** Sends a proposal; resolves false, with the reason shown, if it failed. */
  async function propose(proposal: Proposal): Promise<boolean> {
    logged({ what: 'sent', kind: proposal.kind, detail: proposal.name ?? '' });
    try {
      await move.send(proposal);
      logged({ what: 'settled', kind: proposal.kind });
      return true;
    } catch (cause) {
      logged({ what: 'refused', kind: proposal.kind, detail: describe(cause) });
      problem.textContent = describe(cause);
      return false;
    }
  }

  /** A lobby button: one proposal at a time, the buttons held meanwhile. */
  async function act(proposal: Proposal): Promise<void> {
    if (sending) return;
    sending = true;
    sync();
    const before = viewCell.get()?.revision ?? 0;
    const sent = await propose(proposal);
    sending = false;
    if (sent) {
      const after = await viewCell.pull().catch(() => undefined);
      if ((after?.revision ?? 0) <= before) problem.textContent = 'The table refused that.';
    }
    sync();
  }

  sit.addEventListener('click', () => void act({ kind: 'sit', label: nameField.value }));
  leave.addEventListener('click', () => void act({ kind: 'leave' }));
  start.addEventListener('click', () => void act({
    kind: 'start',
    options: Object.fromEntries([...options.querySelectorAll('select')].map(select => [select.name, select.value])),
  }));
  again.addEventListener('click', () => void act({ kind: 'restart' }));

  async function onProposal(message: Extract<GuestToHostMessage, { type: 'propose-move' }>): Promise<void> {
    const epoch = sourceCell().get()?.epoch;
    const sent = await propose({
      kind: 'play',
      name: message.name,
      arguments: { ...message.arguments },
    });
    // A refused proposal writes nothing and says nothing, so the only sign
    // of one is that the table did not move on. A newer state settles the
    // proposal by itself; this is for the case where none comes.
    const view = sent ? await sourceCell().pull().catch(() => undefined) : undefined;
    const advanced = view !== undefined &&
      (view.epoch !== epoch || view.version > message.snapshotVersion);
    if (!advanced) {
      toHost({
        type: 'move-result',
        requestId: message.requestId,
        ok: false,
        error: sent ? 'The table refused that move.' : problem.textContent || 'The move could not be sent.',
        ...(view ? { actualVersion: view.version } : {}),
      });
    }
  }

  /**
   * A timer ran out on this guest's clock. The table checks it against its
   * own, which may be a little behind, so a refused tick is sent again while
   * the version the timer belongs to is still the head.
   */
  async function onTick(timerId: string, version: number, attempt = 0): Promise<void> {
    const view = viewCell.get();
    if (!view || view.phase !== 'playing' || view.version !== version) return;
    await propose({ kind: 'tick', timerId });
    const after = await viewCell.pull().catch(() => undefined);
    if (!after || after.version !== version || attempt >= TICK_RETRIES) return;
    setTimeout(() => void onTick(timerId, version, attempt + 1), TICK_RETRY_MS);
  }

  function onHostMessage(message: GuestToHostMessage): void {
    switch (message.type) {
      case 'propose-move':
        void onProposal(message);
        return;
      case 'tick':
        void onTick(message.timerId, message.version);
        return;
      case 'error':
        console.warn('[boardgame guest]', message.message);
        problem.textContent = message.message;
        return;
      default:
        return;
    }
  }

  // ----- start -------------------------------------------------------------

  const onRefused = (cause: CellReadRefusal): void => {
    refusal = cause;
    sync();
  };
  const stops = [
    viewCell.sink(() => { refusal = undefined; sync(); }, { onRefused }),
    infoCell.sink(sync),
    meCell.sink(sync),
  ];

  void (async () => {
    try {
      await Promise.all([viewCell.pull(), infoCell.pull(), meCell.pull()]);
    } catch (cause) {
      problem.textContent = describe(cause);
    }
    hydrated = true;
    sync();
    // The one event not sent from a gesture: it asks the pattern to write
    // this viewer's own principal where only they read it, so the guest can
    // tell which seat is theirs. It changes nothing once that is written.
    const me = meCell.get();
    if (typeof me !== 'string' || me === '') {
      try {
        await identify.send(undefined as void);
        await meCell.pull();
      } catch (cause) {
        // A viewer who may not send events is an observer.
        console.warn('[boardgame guest] identify refused:', describe(cause));
      }
      sync();
    }
  })();

  globalThis.addEventListener('pagehide', () => {
    stopOwn?.();
    stops.forEach(stop => stop());
    fabric.disconnect();
  }, { once: true });
}
