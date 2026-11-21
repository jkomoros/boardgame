/**
 * The rules of Memory, written by hand to mirror `examples/memory` in Go:
 * the same state shape, the same move names, and the same fix-up moves in
 * the same order, so the unmodified Memory renderer animates them.
 *
 * It is plain data and pure functions behind the `GameRules` interface of
 * `../toolkit/table.ts`. The TypeScript engine's rules layer is meant to
 * replace this file; the pattern beside it would not change.
 *
 * What it takes from the engine today is `componentId`, so card ids are the
 * salted hashes the Go core computes. The engine's stacks and `viewFor` are
 * classes, which the pattern sandbox refuses (FINDINGS.md, "The engine in
 * the sandbox"), so the handful of stack operations Memory needs, and its
 * one sanitization policy, are written out below.
 */
import { componentId } from "../../../server/static/src/engine/ids.ts";
import { sha1Hex } from "../../../server/static/src/engine/sha1.ts";

import {
  ADMIN_SEAT,
  type GameIdentity,
  type GameRules,
  type StepOutcome,
  type SteppedVersion,
} from "../toolkit/table.ts";
import type {
  ChestComponent,
  Outcome,
  StackJSON,
  TableInfo,
  ViewerPart,
} from "../toolkit/view.ts";

//
// The deck
//

export const DECK = "cards";

const GENERAL =
  "🚴 ✋ 💘 🎓 🌍 🏖 🏛 ⛺ 🚑 🚕 ⚓ 🕰 🌈 🔥 ⛄ 🎄 🎁 🏆 ⚽ 🎳";
const FOODS =
  "🍒 🍔 🍭 🍇 🍉 🍊 🍌 🍍 🍓 🌽 🥕 🍗 🍕 🍩 🍦 🍺 🌮 🌭 🧀 🥐";
const ANIMALS =
  "🐕 🐄 🐘 🐍 🦀 🏇 🦍 🐈 🐖 🐫 🐁 🐿 🦇 🐓 🦅 🦉 🐋 🦑 🐝 🐡";

export interface CardValues {
  Type: string;
  CardSet: string;
}

/** Every card, by deck index: two of each type, as `newDeck()` builds it. */
function deckCards(): CardValues[] {
  const cards: CardValues[] = [];
  const sets: [string, string][] = [
    ["general", GENERAL],
    ["foods", FOODS],
    ["animals", ANIMALS],
  ];
  for (const [cardSet, types] of sets) {
    for (const type of types.split(" ")) {
      cards.push({ Type: type, CardSet: cardSet });
      cards.push({ Type: type, CardSet: cardSet });
    }
  }
  return cards;
}

/** The public facts the renderer needs once: the deck and message texts. */
export function memoryInfo(): TableInfo<CardValues> {
  const components: ChestComponent<CardValues>[] = deckCards().map((
    values,
    index,
  ) => ({ Index: index, Values: values }));
  return {
    gameName: "memory",
    chest: {
      Decks: { [DECK]: components },
      Enums: {},
      LegalTemplates: {
        "reveal.no_cards_left": "You have no cards left to reveal this turn",
        "hide.cards_still_to_reveal":
          "You still have to reveal more cards before your turn is over",
        "hide.no_cards_to_hide": "no cards left to hide",
      },
    },
    seatColors: [
      "#D32F2F",
      "#1976D2",
      "#388E3C",
      "#F57C00",
      "#7B1FA2",
      "#00796B",
    ],
  };
}

//
// The state
//

export const MOVE_REVEAL = "Reveal Card";
export const MOVE_HIDE = "Hide Cards";
export const MOVE_FINISH_TURN = "Finish Turn";
export const MOVE_CAPTURE = "Capture Cards";
export const MOVE_START_TIMER = "Start Hide Cards Timer";

/** How long two unmatched cards stay face-up before they are hidden. */
export const HIDE_CARDS_MS = 4000;

const EMPTY_SLOT = -1;
const HIDDEN_CARD = -2;

export interface TimerJSON {
  ID: string;
  IsTimer: boolean;
}

export interface MemoryGameState {
  CurrentPlayer: number;
  CardSet: string;
  NumCards: number;
  HiddenCards: StackJSON;
  VisibleCards: StackJSON;
  /** `VisibleCards` laid over `HiddenCards`; recomputed, never stored. */
  Cards: StackJSON;
  HideCardsTimer: TimerJSON;
  UnusedCards: StackJSON;
}

export interface MemoryPlayerState {
  SeatFilled: boolean;
  SeatClosed: boolean;
  PlayerInactive: boolean;
  CardsLeftToReveal: number;
  WonCards: StackJSON;
}

/** The whole game, unsanitized: what the Go server would store. */
export interface MemoryGame {
  Game: MemoryGameState;
  Players: MemoryPlayerState[];
  /** How many times each card's id has been re-rolled, by deck index. */
  SecretMoveCount: number[];
}

export interface PlayerComputed {
  Color: string;
  GameScore: number;
  MayBeActive: boolean;
}

/** What a viewer is shown: `Game.JSONForPlayer(...).CurrentState`. */
export interface MemoryView {
  Game: MemoryGameState;
  Players: MemoryPlayerState[];
  Computed: {
    Global: { CurrentPlayerHasCardsToReveal: boolean };
    Players: PlayerComputed[];
  };
  Components: Record<string, never>;
}

//
// Stacks
//

function emptyStack(size?: number): StackJSON {
  const stack: StackJSON = {
    Deck: DECK,
    Indexes: [],
    IDs: [],
    IDsLastSeen: {},
    ShuffleCount: 0,
  };
  if (size !== undefined) {
    stack.Size = size;
    for (let i = 0; i < size; i++) {
      stack.Indexes.push(EMPTY_SLOT);
      stack.IDs.push("");
    }
  }
  return stack;
}

function copyStack(stack: StackJSON): StackJSON {
  const copy: StackJSON = {
    Deck: stack.Deck,
    Indexes: stack.Indexes.slice(),
    IDs: stack.IDs.slice(),
    IDsLastSeen: { ...stack.IDsLastSeen },
    ShuffleCount: stack.ShuffleCount,
  };
  if (stack.Size !== undefined) copy.Size = stack.Size;
  return copy;
}

function copyGame(game: MemoryGame): MemoryGame {
  return {
    Game: {
      ...game.Game,
      HiddenCards: copyStack(game.Game.HiddenCards),
      VisibleCards: copyStack(game.Game.VisibleCards),
      Cards: copyStack(game.Game.Cards),
      HideCardsTimer: { ...game.Game.HideCardsTimer },
      UnusedCards: copyStack(game.Game.UnusedCards),
    },
    Players: game.Players.map((player) => ({
      ...player,
      WonCards: copyStack(player.WonCards),
    })),
    SecretMoveCount: game.SecretMoveCount.slice(),
  };
}

/** A game being changed into its next version. */
interface Working {
  game: MemoryGame;
  identity: GameIdentity;
  version: number;
}

function idOf(work: Working, deckIndex: number): string {
  return componentId(
    work.identity,
    DECK,
    deckIndex,
    work.game.SecretMoveCount[deckIndex] ?? 0,
  );
}

/** Puts a card into a slot of a sized stack, noting when its id was seen. */
function place(
  work: Working,
  stack: StackJSON,
  slot: number,
  deckIndex: number,
): void {
  const id = idOf(work, deckIndex);
  stack.Indexes[slot] = deckIndex;
  stack.IDs[slot] = id;
  stack.IDsLastSeen[id] = work.version;
}

/** Puts a card at the end of a growable stack. */
function append(work: Working, stack: StackJSON, deckIndex: number): void {
  const id = idOf(work, deckIndex);
  stack.Indexes.push(deckIndex);
  stack.IDs.push(id);
  stack.IDsLastSeen[id] = work.version;
}

function clearSlot(stack: StackJSON, slot: number): void {
  stack.Indexes[slot] = EMPTY_SLOT;
  stack.IDs[slot] = "";
}

/**
 * Reorders a stack and re-rolls the id of every card in it, as a shuffle in
 * the Go core does, so no id links a card to where it was before.
 */
function shuffle(work: Working, stack: StackJSON, purpose: string): void {
  const cards = stack.Indexes.filter((index) => index >= 0);
  const draw = drawsFor(work.identity, purpose);
  for (let i = cards.length - 1; i > 0; i--) {
    const j = Math.floor(draw() * (i + 1));
    [cards[i], cards[j]] = [cards[j], cards[i]];
  }
  let next = 0;
  for (let slot = 0; slot < stack.Indexes.length; slot++) {
    if (stack.Indexes[slot] < 0) continue;
    const deckIndex = cards[next++];
    work.game.SecretMoveCount[deckIndex] =
      (work.game.SecretMoveCount[deckIndex] ?? 0) + 1;
    place(work, stack, slot, deckIndex);
  }
  stack.ShuffleCount += 1;
}

/**
 * A stream of numbers in [0, 1) that depends only on the game's identity and
 * `purpose`. The salt is the only fresh randomness a game ever draws.
 *
 * TODO(engine): the Go core draws from a port of `math/rand` seeded from the
 * same identity (`engine/rand.ts`), which is a class and so cannot be
 * imported here. These draws are deterministic but are not Go's.
 */
function drawsFor(identity: GameIdentity, purpose: string): () => number {
  let counter = 0;
  return () => {
    const digest = sha1Hex(
      `${identity.gameId}:${identity.secretSalt}:${purpose}:${counter++}`,
    );
    return parseInt(digest.slice(0, 13), 16) / 0x10000000000000;
  };
}

/** `VisibleCards` laid over `HiddenCards`, as the Go merged stack is. */
function mergedCards(game: MemoryGameState): StackJSON {
  const hidden = game.HiddenCards;
  const visible = game.VisibleCards;
  const merged = emptyStack();
  if (hidden.Size !== undefined) merged.Size = hidden.Size;
  merged.ShuffleCount = hidden.ShuffleCount + visible.ShuffleCount;
  for (let slot = 0; slot < hidden.Indexes.length; slot++) {
    const shown = (visible.Indexes[slot] ?? EMPTY_SLOT) !== EMPTY_SLOT;
    merged.Indexes.push(shown ? visible.Indexes[slot] : hidden.Indexes[slot]);
    merged.IDs.push(shown ? visible.IDs[slot] : hidden.IDs[slot]);
  }
  // As in Go, an id last seen at version 0 does not enter a merged stack's
  // record.
  for (const part of [visible, hidden]) {
    for (const id of Object.keys(part.IDsLastSeen)) {
      const seen = part.IDsLastSeen[id];
      if (seen > (merged.IDsLastSeen[id] ?? 0)) merged.IDsLastSeen[id] = seen;
    }
  }
  return merged;
}

function occupied(stack: StackJSON): number[] {
  const slots: number[] = [];
  stack.Indexes.forEach((index, slot) => {
    if (index !== EMPTY_SLOT) slots.push(slot);
  });
  return slots;
}

//
// Setup
//

const NUM_CARDS: Record<string, number> = { small: 10, medium: 20, large: 40 };
const CARD_SETS = ["all", "general", "foods", "animals"];

function setUp(
  identity: GameIdentity,
  numPlayers: number,
  options: Record<string, string>,
): MemoryGame {
  const cards = deckCards();
  const cardSet = CARD_SETS.includes(options["cardset"])
    ? options["cardset"]
    : "all";
  const numCards = NUM_CARDS[options["numcards"]] ?? NUM_CARDS["small"];
  const game: MemoryGame = {
    Game: {
      CurrentPlayer: 0,
      CardSet: cardSet,
      NumCards: numCards,
      HiddenCards: emptyStack(numCards),
      VisibleCards: emptyStack(numCards),
      Cards: emptyStack(numCards),
      HideCardsTimer: { ID: "", IsTimer: true },
      UnusedCards: emptyStack(),
    },
    Players: [],
    SecretMoveCount: cards.map(() => 0),
  };
  for (let seat = 0; seat < numPlayers; seat++) {
    game.Players.push({
      SeatFilled: true,
      SeatClosed: false,
      PlayerInactive: false,
      CardsLeftToReveal: seat === 0 ? 2 : 0,
      WonCards: emptyStack(),
    });
  }
  const work: Working = { game, identity, version: 0 };

  // Choose which pairs are in play: shuffle the types the table asked for
  // and take as many as the board holds. Deck indexes 2n and 2n+1 are the
  // two cards of one type.
  //
  // TODO(engine): the Go game keeps the cards not in play in `UnusedCards`,
  // so that every component is in exactly one stack. Nothing draws them, and
  // listing 110 ids in every version made the table six times larger, so
  // here the stack stays empty.
  const types: number[] = [];
  for (let type = 0; type < cards.length / 2; type++) {
    if (cardSet === "all" || cards[2 * type].CardSet === cardSet) {
      types.push(type);
    }
  }
  const draw = drawsFor(identity, "deal");
  for (let i = types.length - 1; i > 0; i--) {
    const j = Math.floor(draw() * (i + 1));
    [types[i], types[j]] = [types[j], types[i]];
  }
  const taken: number[] = [];
  for (const type of types.slice(0, numCards / 2)) {
    taken.push(2 * type, 2 * type + 1);
  }
  taken.forEach((deckIndex, slot) =>
    place(work, game.Game.HiddenCards, slot, deckIndex)
  );
  shuffle(work, game.Game.HiddenCards, "hidden");
  game.Game.Cards = mergedCards(game.Game);
  return game;
}

//
// Legality
//

function typeAt(stack: StackJSON, slot: number, cards: CardValues[]): string {
  return cards[stack.Indexes[slot]]?.Type ?? "";
}

/** Why `seat` may not reveal the card at `slot`, or `""`. */
function revealError(game: MemoryGame, seat: number, slot: number): string {
  if (seat !== ADMIN_SEAT && seat !== game.Game.CurrentPlayer) {
    return "It is not your turn";
  }
  const player = game.Players[game.Game.CurrentPlayer];
  if (player === undefined || player.CardsLeftToReveal < 1) {
    return "You have no cards left to reveal this turn";
  }
  if (!Number.isInteger(slot) || slot < 0 || slot >= game.Game.NumCards) {
    return "there is no card at that index";
  }
  if (game.Game.HiddenCards.Indexes[slot] === EMPTY_SLOT) {
    return game.Game.VisibleCards.Indexes[slot] === EMPTY_SLOT
      ? "there is no card at that index"
      : "that card has already been revealed";
  }
  return "";
}

/** Why `seat` may not hide the face-up cards, or `""`. */
function hideError(game: MemoryGame, seat: number): string {
  if (seat !== ADMIN_SEAT && seat !== game.Game.CurrentPlayer) {
    return "It is not your turn";
  }
  const player = game.Players[game.Game.CurrentPlayer];
  if (player === undefined || player.CardsLeftToReveal > 0) {
    return "You still have to reveal more cards before your turn is over";
  }
  if (occupied(game.Game.VisibleCards).length < 1) {
    return "no cards left to hide";
  }
  return "";
}

function outcome(game: MemoryGame): Outcome {
  const left = occupied(game.Game.HiddenCards).length +
    occupied(game.Game.VisibleCards).length;
  if (left > 0) return { finished: false, winners: [] };
  const scores = game.Players.map((player) => player.WonCards.Indexes.length);
  const best = Math.max(...scores, 0);
  const winners: number[] = [];
  scores.forEach((score, seat) => {
    if (score === best) winners.push(seat);
  });
  return { finished: true, winners };
}

//
// Moves
//

/** The first fix-up move legal at `game`, in the order Go configures them. */
function nextFixUp(game: MemoryGame, cards: CardValues[]): string {
  const visible = occupied(game.Game.VisibleCards);
  const player = game.Players[game.Game.CurrentPlayer];
  if (player.CardsLeftToReveal <= 0 && visible.length === 0) {
    return MOVE_FINISH_TURN;
  }
  if (visible.length === 2) {
    const matched =
      typeAt(game.Game.VisibleCards, visible[0], cards) ===
        typeAt(game.Game.VisibleCards, visible[1], cards);
    if (matched) return MOVE_CAPTURE;
    if (game.Game.HideCardsTimer.ID === "") return MOVE_START_TIMER;
  }
  return "";
}

/** Applies one move, already known to be legal, to a copy of `game`. */
function apply(
  game: MemoryGame,
  version: number,
  identity: GameIdentity,
  moveName: string,
  slot: number,
): SteppedVersion<MemoryGame> {
  const next = copyGame(game);
  const work: Working = { game: next, identity, version };
  const state = next.Game;
  const stepped: SteppedVersion<MemoryGame> = {
    game: next,
    moveName,
    cancelTimers: [],
  };
  switch (moveName) {
    case MOVE_REVEAL: {
      const deckIndex = state.HiddenCards.Indexes[slot];
      clearSlot(state.HiddenCards, slot);
      place(work, state.VisibleCards, slot, deckIndex);
      next.Players[state.CurrentPlayer].CardsLeftToReveal -= 1;
      break;
    }
    case MOVE_HIDE: {
      if (state.HideCardsTimer.ID !== "") {
        stepped.cancelTimers.push(state.HideCardsTimer.ID);
        state.HideCardsTimer.ID = "";
      }
      for (const shown of occupied(state.VisibleCards)) {
        const deckIndex = state.VisibleCards.Indexes[shown];
        clearSlot(state.VisibleCards, shown);
        place(work, state.HiddenCards, shown, deckIndex);
      }
      break;
    }
    case MOVE_CAPTURE: {
      const won = next.Players[state.CurrentPlayer].WonCards;
      for (const shown of occupied(state.VisibleCards)) {
        const deckIndex = state.VisibleCards.Indexes[shown];
        clearSlot(state.VisibleCards, shown);
        append(work, won, deckIndex);
      }
      break;
    }
    case MOVE_START_TIMER: {
      // The id is public once the timer runs, and says nothing of the salt.
      const id = sha1Hex(
        `${identity.gameId}:${identity.secretSalt}:timer:${version}`,
      ).slice(0, 16).toUpperCase();
      state.HideCardsTimer.ID = id;
      stepped.startTimer = {
        id,
        durationMs: HIDE_CARDS_MS,
        moveName: MOVE_HIDE,
      };
      break;
    }
    case MOVE_FINISH_TURN: {
      state.CurrentPlayer = (state.CurrentPlayer + 1) % next.Players.length;
      next.Players[state.CurrentPlayer].CardsLeftToReveal = 2;
      break;
    }
  }
  state.Cards = mergedCards(state);
  return stepped;
}

/** The most fix-up moves one proposal may cause, as a guard. */
const FIX_UP_CAP = 8;

function step(
  game: MemoryGame,
  version: number,
  identity: GameIdentity,
  moveName: string,
  args: Record<string, string>,
  seat: number,
): StepOutcome<MemoryGame> {
  if (outcome(game).finished) {
    return { error: "the game is over", versions: [] };
  }
  let slot = -1;
  let error: string;
  if (moveName === MOVE_REVEAL) {
    const raw = args["CardIndex"];
    slot = typeof raw === "string" && /^\d+$/.test(raw) ? Number(raw) : -1;
    error = revealError(game, seat, slot);
  } else if (moveName === MOVE_HIDE) {
    error = hideError(game, seat);
  } else {
    // Fix-up moves are the table's own; no seat may propose one.
    error = "that move cannot be proposed";
  }
  if (error !== "") return { error, versions: [] };

  const cards = deckCards();
  const versions: SteppedVersion<MemoryGame>[] = [];
  let current = apply(game, version + 1, identity, moveName, slot);
  versions.push(current);
  while (versions.length <= FIX_UP_CAP && !outcome(current.game).finished) {
    const fixUp = nextFixUp(current.game, cards);
    if (fixUp === "") break;
    current = apply(
      current.game,
      version + 1 + versions.length,
      identity,
      fixUp,
      -1,
    );
    versions.push(current);
  }
  return { error: "", versions };
}

//
// What a viewer sees
//

/**
 * The `order` policy on a stack: every card becomes the hidden card, ids and
 * occupancy stay. This is the only place a face is removed, so it is the
 * whole of what keeps a hidden card hidden.
 */
function orderOnly(stack: StackJSON): StackJSON {
  const copy = copyStack(stack);
  copy.Indexes = copy.Indexes.map((index) =>
    index === EMPTY_SLOT ? EMPTY_SLOT : HIDDEN_CARD
  );
  return copy;
}

function sanitize(game: MemoryGame, _viewer: number): MemoryView {
  const info = memoryInfo();
  const state: MemoryGameState = {
    CurrentPlayer: game.Game.CurrentPlayer,
    CardSet: game.Game.CardSet,
    NumCards: game.Game.NumCards,
    HiddenCards: orderOnly(game.Game.HiddenCards),
    VisibleCards: copyStack(game.Game.VisibleCards),
    Cards: emptyStack(),
    HideCardsTimer: { ...game.Game.HideCardsTimer },
    UnusedCards: copyStack(game.Game.UnusedCards),
  };
  state.Cards = mergedCards(state);
  const current = game.Players[game.Game.CurrentPlayer];
  return {
    Game: state,
    Players: game.Players.map((player) => ({
      SeatFilled: player.SeatFilled,
      SeatClosed: player.SeatClosed,
      PlayerInactive: player.PlayerInactive,
      CardsLeftToReveal: player.CardsLeftToReveal,
      WonCards: copyStack(player.WonCards),
    })),
    Computed: {
      Global: {
        CurrentPlayerHasCardsToReveal: current !== undefined &&
          current.CardsLeftToReveal > 0,
      },
      Players: game.Players.map((player, seat) => ({
        Color: info.seatColors[seat % info.seatColors.length],
        GameScore: player.WonCards.Indexes.length,
        MayBeActive: true,
      })),
    },
    Components: {},
  };
}

/** Copied from the Go server's output for Memory; the client checks neither. */
const CHOICE_FINGERPRINT =
  "sha256:5c046960abdccaff37af1d4964959fc340eeb9649ce94e66aa842c25006dcfab";

function viewerPart(
  game: MemoryGame,
  version: number,
  viewer: number,
): ViewerPart {
  const over = outcome(game).finished;
  const current = game.Game.CurrentPlayer;
  const firstHidden = occupied(game.Game.HiddenCards)[0] ?? -1;
  const revealForAnyone = !over &&
    revealError(game, ADMIN_SEAT, firstHidden) === "";
  const hideForAnyone = !over && hideError(game, ADMIN_SEAT) === "";
  const mine = viewer === current;
  const part: ViewerPart = {
    moveLegality: {
      [MOVE_HIDE]: {
        legalForPlayer: mine && hideForAnyone,
        legalForAnyone: hideForAnyone,
      },
      [MOVE_REVEAL]: {
        legalForPlayer: mine && revealForAnyone,
        legalForAnyone: revealForAnyone,
      },
    },
    zeroInputMoves: [MOVE_HIDE],
    projectedMoveChoices: {
      StateVersion: version,
      Status: "ready",
      MoveChoiceProjectionSchemaFingerprint: CHOICE_FINGERPRINT,
      ProjectionSchemaVersion: 1,
      Sets: [],
    },
  };
  if (mine && revealForAnyone) {
    part.projectedMoveChoices.Sets.push({
      MoveName: MOVE_REVEAL,
      FieldName: "CardIndex",
      Source: "stack-slots",
      Candidates: occupied(game.Game.HiddenCards).map((slot) => ({
        Value: slot,
        Available: true,
      })),
    });
  }
  return part;
}

/**
 * Memory, as a table plays it.
 *
 * A function, not a constant: the pattern sandbox admits only plain data at
 * module scope, and a record of functions is not plain data (FINDINGS.md,
 * "The engine in the sandbox").
 */
export function memoryRules(): GameRules<MemoryGame, MemoryView> {
  return {
    name: "memory",
    minPlayers: 2,
    maxPlayers: 6,
    setUp,
    step,
    currentPlayer: (game) => game.Game.CurrentPlayer,
    outcome,
    sanitize,
    viewerPart,
  };
}
