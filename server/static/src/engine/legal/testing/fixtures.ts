// A small, self-contained implementation of the state-access interface
// (../access.ts) for tests, and a worked example of how to bind it.
//
// It has two entry points:
//   - plain builders (reader, stack, enumOf, ...) for hand-written fixtures;
//   - bindWireFixture, which builds the same objects from a fixture exported
//     by the Go generator (testdata/gen): a state.StorageRecord(), the chest
//     JSON /info ships, and the type tables a state record cannot carry.
//
// The stack here models growable, sized and merged stacks well enough to
// answer Go's MayMoveTo / MayMoveToSlot / MayMoveAllTo / MayMoveCountTo /
// MaySwapComponents for stacks WITHOUT constraints, with Go's exact error
// text. It is test scaffolding, not the engine's stack implementation: the
// real one is bound in its place (see ../README.md). A constrained
// destination throws rather than guessing.
//
// Pure TypeScript: no Node APIs, so a test reads the JSON and hands it in.

import type {
  LegalBoard,
  LegalChest,
  LegalComponent,
  LegalEnum,
  LegalEnumValue,
  LegalFramework,
  LegalMove,
  LegalPhase,
  LegalPropertyReader,
  LegalStack,
  LegalState,
} from '../access.ts';
import type { PropertyType } from '../types.ts';

/** A tagged value for a hand-written reader, where the JS type alone is ambiguous. */
export interface TypedValue {
  type: PropertyType;
  value: unknown;
}

/** A PlayerIndex-typed property value. */
export function playerIndex(value: number): TypedValue {
  return { type: 'TypePlayerIndex', value };
}

/** A property of an explicit type (for types legality never reads, e.g. TypeTimer). */
export function typed(type: PropertyType, value: unknown): TypedValue {
  return { type, value };
}

function isTyped(v: unknown): v is TypedValue {
  return typeof v === 'object' && v !== null && 'type' in v && 'value' in v && Object.keys(v).length === 2;
}

/**
 * A reader over a plain object. number -> TypeInt, boolean -> TypeBool,
 * string -> TypeString, a FixtureStack -> TypeStack, a FixtureBoard ->
 * TypeBoard, an enum value -> TypeEnum; wrap anything else with playerIndex()
 * or typed(). A function value is called on read and may throw, to model a
 * failing read.
 */
export function reader(values: Record<string, unknown>): LegalPropertyReader {
  const props: Record<string, PropertyType> = {};
  const raw: Record<string, unknown> = {};
  for (const [name, v] of Object.entries(values)) {
    if (isTyped(v)) {
      props[name] = v.type;
      raw[name] = v.value;
    } else if (v instanceof FixtureStack) {
      props[name] = 'TypeStack';
      raw[name] = v;
    } else if (v instanceof FixtureBoard) {
      props[name] = 'TypeBoard';
      raw[name] = v;
    } else if (typeof v === 'number') {
      props[name] = 'TypeInt';
      raw[name] = v;
    } else if (typeof v === 'boolean') {
      props[name] = 'TypeBool';
      raw[name] = v;
    } else if (typeof v === 'string') {
      props[name] = 'TypeString';
      raw[name] = v;
    } else if (typeof v === 'object' && v !== null && 'enum' in v && 'string' in v) {
      props[name] = 'TypeEnum';
      raw[name] = v;
    } else {
      throw new Error(`fixture reader: cannot infer a PropertyType for ${name}; wrap it with typed()`);
    }
  }
  return {
    props: () => props,
    prop(name) {
      const v = raw[name];
      return typeof v === 'function' ? (v as () => unknown)() : v;
    },
  };
}

/**
 * An enum from a key -> name table. valueFromString follows Go's
 * enum.ValueFromString: TrimSpace+ToLower match on the names, then the
 * decimal key itself.
 */
export function enumOf(name: string, values: Record<number, string>, parents?: Record<number, number>): LegalEnum {
  const keys = Object.keys(values).map(Number).sort((a, b) => a - b);
  const normalize = (s: string) => s.toLowerCase().trim();
  return {
    name,
    valueFromString(s) {
      const wanted = normalize(s);
      for (const key of keys) {
        if (normalize(values[key]) === wanted) {
          return key;
        }
      }
      if (/^[+-]?[0-9]+$/.test(s) && Object.prototype.hasOwnProperty.call(values, Number(s))) {
        return Number(s);
      }
      return null;
    },
    ancestors(key) {
      if (!parents) {
        return null;
      }
      const chain = [key];
      let current = key;
      while (Object.prototype.hasOwnProperty.call(parents, current) && parents[current] !== current) {
        current = parents[current];
        chain.push(current);
      }
      return chain;
    },
  };
}

/** A value of `e`. */
export function enumValue(e: LegalEnum, key: number, names: Record<number, string>): LegalEnumValue {
  return { enum: e, value: key, string: names[key] ?? '' };
}

/** A component of a fixture deck: its static values, and where it currently sits. */
export class FixtureComponent implements LegalComponent {
  readonly deck: string;
  readonly deckIndex: number;
  private readonly reader: LegalPropertyReader | null;
  /** The physical stack holding this component, set when the stack is built. */
  container: FixtureStack | null = null;

  constructor(deck: string, deckIndex: number, values: LegalPropertyReader | null) {
    this.deck = deck;
    this.deckIndex = deckIndex;
    this.reader = values;
  }

  values(): LegalPropertyReader | null {
    return this.reader;
  }

  /** Go: componentInstance.MayMoveTo. */
  mayMoveTo(destination: LegalStack): string | null {
    const dest = destination as FixtureStack;
    const source = this.container;
    if (source === null) {
      return 'component is not in any stack: component was not found in any attached stack; a captured stack may be stale or replaced';
    }
    if (source === dest) {
      return 'source and destination are the same stack';
    }
    if (source.deck !== dest.deck) {
      return 'source and destination use different decks';
    }
    if (dest.slotsRemaining() < 1) {
      return 'destination stack has no slots remaining';
    }
    dest.assertUnconstrained();
    return null;
  }

  /** Go: componentInstance.MayMoveToSlot. */
  mayMoveToSlot(destination: LegalStack, slotIndex: number): string | null {
    const err = this.mayMoveTo(destination);
    if (err !== null) {
      return err;
    }
    const dest = destination as FixtureStack;
    if (slotIndex < 0) {
      return 'slot index must be non-negative';
    }
    if (dest.kind === 'sized') {
      if (slotIndex >= dest.len()) {
        return `slot index ${slotIndex} is out of range (stack length ${dest.len()})`;
      }
      if (dest.componentAt(slotIndex) !== null) {
        return `slot ${slotIndex} is already occupied`;
      }
    } else if (slotIndex > dest.len()) {
      return `slot index ${slotIndex} is out of range (stack length ${dest.len()})`;
    }
    return null;
  }
}

export type FixtureStackKind = 'growable' | 'sized' | 'merged';

/** A growable, sized or merged stack of FixtureComponents. null is an empty slot. */
export class FixtureStack implements LegalStack {
  readonly deck: string;
  readonly kind: FixtureStackKind;
  readonly slots: readonly (FixtureComponent | null)[];
  /** Go: MaxSize() for a growable stack; 0 is unbounded. */
  readonly maxSize: number;
  readonly constrained: boolean;

  constructor(
    deck: string,
    kind: FixtureStackKind,
    slots: readonly (FixtureComponent | null)[],
    options: { maxSize?: number; constrained?: boolean } = {},
  ) {
    this.deck = deck;
    this.kind = kind;
    this.slots = slots;
    this.maxSize = options.maxSize ?? 0;
    this.constrained = options.constrained ?? false;
    if (kind !== 'merged') {
      for (const component of slots) {
        if (component !== null) {
          component.container = this;
        }
      }
    }
  }

  len(): number {
    return this.slots.length;
  }

  numComponents(): number {
    return this.slots.filter(c => c !== null).length;
  }

  componentAt(index: number): LegalComponent | null {
    if (!Number.isInteger(index) || index < 0 || index >= this.slots.length) {
      return null;
    }
    return this.slots[index];
  }

  first(): LegalComponent | null {
    return this.slots.find(c => c !== null) ?? null;
  }

  /** Go: SlotsRemaining(). */
  slotsRemaining(): number {
    if (this.kind === 'merged') {
      throw new Error('fixture stack: SlotsRemaining of a merged stack is not modeled');
    }
    if (this.kind === 'sized') {
      return this.slots.filter(c => c === null).length;
    }
    return this.maxSize <= 0 ? Number.MAX_SAFE_INTEGER : this.maxSize - this.slots.length;
  }

  assertUnconstrained(): void {
    if (this.constrained) {
      throw new Error('fixture stack: stack constraints are not modeled; bind the real stack implementation');
    }
  }

  /** Go: preparePhysicalComponentTransfer followed by validateComponentTransferPlan. */
  private transfer(destination: LegalStack, count: number, capacityError: string): string | null {
    const dest = destination as FixtureStack;
    if (dest.kind === 'merged') {
      return 'destination is a view, not a physical stack';
    }
    if (this === dest) {
      return 'source and destination are the same stack';
    }
    if (this.deck !== dest.deck) {
      return 'source and destination use different decks';
    }
    if (count < 0) {
      return 'component count must be non-negative';
    }
    const available = this.numComponents();
    if (count > available) {
      return `source stack has ${available} components, cannot move ${count}`;
    }
    const remaining = dest.slotsRemaining();
    if (remaining < count) {
      if (capacityError !== '') {
        return capacityError;
      }
      const unit = (n: number, singular: string) => (n === 1 ? singular : singular + 's');
      return `destination stack has ${remaining} ${unit(remaining, 'slot')} remaining; cannot move ${count} ${unit(count, 'component')}`;
    }
    if (count === 0) {
      return null;
    }
    dest.assertUnconstrained();
    let validated = 0;
    for (const component of this.slots) {
      if (component === null) {
        continue;
      }
      if (validated === count) {
        break;
      }
      const err = component.mayMoveTo(dest);
      if (err !== null) {
        return err;
      }
      validated++;
    }
    return null;
  }

  mayMoveAllTo(destination: LegalStack): string | null {
    if (this.kind === 'merged') {
      return 'MayMoveAllTo is not supported on MergedStacks';
    }
    return this.transfer(destination, this.numComponents(), 'not enough space in the target stack');
  }

  mayMoveCountTo(destination: LegalStack, count: number): string | null {
    if (this.kind === 'merged') {
      return 'MayMoveCountTo is not supported on MergedStacks';
    }
    return this.transfer(destination, count, '');
  }

  /** Go: maySwapComponentsImpl. */
  maySwapComponents(i: number, j: number): string | null {
    if (this.kind === 'merged') {
      return 'MaySwapComponents is not supported on MergedStacks';
    }
    if (i < 0) {
      return 'i must be 0 or greater';
    }
    if (j < 0) {
      return 'j must be 0 or greater';
    }
    if (i >= this.len()) {
      return "i must be less than the stack's length";
    }
    if (j >= this.len()) {
      return "j must be less than the stack's length";
    }
    if (i === j) {
      return 'i and j were the same';
    }
    return null;
  }
}

/** A board: an ordered list of stacks, optionally keyed by an enum. */
export class FixtureBoard implements LegalBoard {
  private readonly spaces: readonly LegalStack[];
  private readonly keyEnum: LegalEnum | null;

  constructor(spaces: readonly LegalStack[], keyEnum: LegalEnum | null = null) {
    this.spaces = spaces;
    this.keyEnum = keyEnum;
  }

  enum(): LegalEnum | null {
    return this.keyEnum;
  }

  spaceAt(index: number): LegalStack | null {
    return Number.isInteger(index) && index >= 0 && index < this.spaces.length ? this.spaces[index] : null;
  }

  spaceAtKey(key: number): LegalStack | null {
    return this.spaceAt(key);
  }
}

/** Components for one deck: `values[i]` are component i's static values (or null for none). */
export function deck(name: string, values: readonly (Record<string, unknown> | null)[]): FixtureComponent[] {
  return values.map((v, i) => new FixtureComponent(name, i, v === null ? null : reader(v)));
}

/** A hand-written state. Everything not given takes the obvious default. */
export interface PlainState {
  game?: Record<string, unknown>;
  players?: readonly Record<string, unknown>[];
  /** Defaults to 0. */
  currentPlayerIndex?: number;
  /** Per seat; defaults to all true. */
  playersMayBeActive?: readonly boolean[];
  /** Per seat; defaults to all false. */
  playersInactive?: readonly boolean[];
  customPlayerOrder?: readonly number[] | null;
}

/** Builds a LegalState from a PlainState. */
export function state(plain: PlainState = {}): LegalState {
  const game = reader(plain.game ?? {});
  const players = (plain.players ?? []).map(reader);
  return {
    game: () => game,
    players: () => players,
    currentPlayerIndex: () => plain.currentPlayerIndex ?? 0,
    playerMayBeActive: index => plain.playersMayBeActive?.[index] ?? true,
    customPlayerOrder: () => plain.customPlayerOrder ?? null,
    playerIsInactive: index => plain.playersInactive?.[index] ?? false,
  };
}

/** Builds a LegalMove from its field values. */
export function move(fields: Record<string, unknown>): LegalMove {
  const r = reader(fields);
  return { reader: () => r };
}

/** A framework whose phase is fixed and whose progression answer is canned. */
export function framework(phase: LegalPhase | null, progression?: LegalFramework['moveInProgression']): LegalFramework {
  return progression ? { currentPhase: () => phase, moveInProgression: progression } : { currentPhase: () => phase };
}

// ---------------------------------------------------------------------------
// Binding a Go-exported fixture (testdata/fixtures.json).

export interface WireProp {
  type: PropertyType;
  enum?: string;
  value?: unknown;
}

export interface WireSubState {
  props: Record<string, WireProp>;
  stacks?: Record<string, { kind: FixtureStackKind; constrained?: boolean }>;
}

export interface WireStackJSON {
  Deck: string;
  Indexes: number[];
  Size?: number;
  MaxSize?: number;
}

export interface WireFixture {
  game: string;
  currentPlayerIndex: number;
  playersMayBeActive: boolean[] | null;
  playersInactive: boolean[] | null;
  customPlayerOrder: number[] | null;
  phase: { key: number; name: string; ancestors: number[] | null } | null;
  gameState: WireSubState;
  playerStates: WireSubState[] | null;
  move: { name: string; props: Record<string, WireProp> } | null;
  state: { Game: Record<string, unknown>; Players: Record<string, unknown>[] | null };
}

export interface WireChestEnum {
  Values: Record<string, string>;
  Parents?: Record<string, number>;
}

export interface WireGame {
  chest: {
    Decks: Record<string, { Index: number; Values: Record<string, unknown> | null }[]> | null;
    Enums: Record<string, WireChestEnum> | null;
    LegalTemplates?: Record<string, string>;
  };
  componentProps: Record<string, Record<string, WireProp>>;
  /** Enums a property uses that the chest does not ship: name -> key -> display string. */
  otherEnums?: Record<string, Record<string, string>>;
}

export interface BoundFixture {
  state: LegalState;
  move: LegalMove | null;
  framework: LegalFramework;
  chest: LegalChest;
  /** The chest's LegalTemplates table, as shipped. */
  templates: Record<string, string>;
}

const EMPTY_SLOT = -1;

/**
 * Builds the access objects for one exported fixture. Enum-typed state
 * values arrive as their display names and are resolved through the chest's
 * enums; stacks arrive as deck indexes and are resolved through its decks.
 */
export function bindWireFixture(fixture: WireFixture, game: WireGame): BoundFixture {
  const enums = new Map<string, { e: LegalEnum; names: Record<number, string> }>();
  for (const [name, def] of Object.entries(game.chest.Enums ?? {})) {
    const names: Record<number, string> = {};
    for (const [key, value] of Object.entries(def.Values)) {
      names[Number(key)] = value;
    }
    const parents = def.Parents
      ? Object.fromEntries(Object.entries(def.Parents).map(([k, v]) => [Number(k), v]))
      : undefined;
    enums.set(name, { e: enumOf(name, names, parents), names });
  }
  const chestEnums = [...enums.values()].map(entry => entry.e);
  for (const [name, values] of Object.entries(game.otherEnums ?? {})) {
    const names: Record<number, string> = {};
    for (const [key, value] of Object.entries(values)) {
      names[Number(key)] = value;
    }
    enums.set(name, { e: enumOf(name, names), names });
  }

  const enumByName = (enumName: string | undefined, propName: string) => {
    const found = enumName === undefined ? undefined : enums.get(enumName);
    if (found === undefined) {
      throw new Error(`fixture: enum ${enumName} for ${propName} is not in the chest`);
    }
    return found;
  };

  const enumFromName = (prop: WireProp, propName: string, display: unknown): LegalEnumValue => {
    const found = enumByName(prop.enum, propName);
    const key = Object.keys(found.names).map(Number).find(k => found.names[k] === display);
    if (key === undefined) {
      throw new Error(`fixture: ${JSON.stringify(display)} is not a value of enum ${prop.enum}`);
    }
    return enumValue(found.e, key, found.names);
  };

  // One component object per deck index, shared by every stack, so that a
  // component knows which physical stack holds it.
  const components = new Map<string, FixtureComponent[]>();
  for (const [deckName, entries] of Object.entries(game.chest.Decks ?? {})) {
    const propTable = game.componentProps[deckName] ?? {};
    components.set(
      deckName,
      entries.map(entry => {
        const values: Record<string, unknown> = {};
        for (const [propName, prop] of Object.entries(propTable)) {
          const rawValue = entry.Values?.[propName];
          values[propName] = prop.type === 'TypeEnum' ? enumFromName(prop, propName, rawValue) : typed(prop.type, rawValue);
        }
        return new FixtureComponent(deckName, entry.Index, entry.Values === null ? null : reader(values));
      }),
    );
  }

  const bindStack = (json: WireStackJSON, meta: { kind: FixtureStackKind; constrained?: boolean } | undefined) => {
    const deckComponents = components.get(json.Deck) ?? [];
    const slots = json.Indexes.map(index => {
      if (index === EMPTY_SLOT) {
        return null;
      }
      const component = deckComponents[index];
      if (component === undefined) {
        throw new Error(`fixture: deck ${json.Deck} has no component ${index}`);
      }
      return component;
    });
    return new FixtureStack(json.Deck, meta?.kind ?? (json.Size !== undefined ? 'sized' : 'growable'), slots, {
      maxSize: json.MaxSize ?? 0,
      constrained: meta?.constrained ?? false,
    });
  };

  const bindSubState = (values: Record<string, unknown>, meta: WireSubState): LegalPropertyReader => {
    const out: Record<string, unknown> = {};
    // Physical stacks first, so a merged view never claims a component.
    const names = Object.keys(meta.props).sort((a, b) => {
      const merged = (n: string) => (meta.stacks?.[n]?.kind === 'merged' ? 1 : 0);
      return merged(a) - merged(b);
    });
    for (const name of names) {
      const prop = meta.props[name];
      const rawValue = values[name];
      if (prop.type === 'TypeStack') {
        out[name] = rawValue ? bindStack(rawValue as WireStackJSON, meta.stacks?.[name]) : typed('TypeStack', null);
      } else if (prop.type === 'TypeEnum') {
        out[name] = enumFromName(prop, name, rawValue);
      } else {
        out[name] = typed(prop.type, rawValue);
      }
    }
    return reader(out);
  };

  const gameReader = bindSubState(fixture.state.Game, fixture.gameState);
  const playerReaders = (fixture.state.Players ?? []).map((values, i) =>
    bindSubState(values, (fixture.playerStates ?? [])[i]),
  );

  let boundMove: LegalMove | null = null;
  if (fixture.move !== null) {
    const fields: Record<string, unknown> = {};
    for (const [name, prop] of Object.entries(fixture.move.props)) {
      if (prop.type === 'TypeEnum') {
        const found = enumByName(prop.enum, name);
        fields[name] = enumValue(found.e, prop.value as number, found.names);
      } else {
        fields[name] = typed(prop.type, prop.value);
      }
    }
    boundMove = move(fields);
  }

  const phase: LegalPhase | null = fixture.phase;
  return {
    state: {
      game: () => gameReader,
      players: () => playerReaders,
      currentPlayerIndex: () => fixture.currentPlayerIndex,
      playerMayBeActive: index => (fixture.playersMayBeActive ?? [])[index] ?? true,
      customPlayerOrder: () => fixture.customPlayerOrder,
      playerIsInactive: index => (fixture.playersInactive ?? [])[index] ?? false,
    },
    move: boundMove,
    framework: { currentPhase: () => phase },
    chest: { enums: () => chestEnums },
    templates: game.chest.LegalTemplates ?? {},
  };
}
