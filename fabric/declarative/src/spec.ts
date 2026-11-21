/**
 * The declarative vocabulary: how a component face and a click are described
 * as data. Pure; no DOM. These replace the two function-valued inputs a
 * renderer class passes to a stack today:
 *
 *   cardView({ render, properties })        ->  FaceSpec
 *   move(name).targets(slots, slot => args) ->  PickSpec
 */
import { resolvePath } from './expand.js';

/** Where one output value comes from. A bare string is a field of `Values`. */
export type Source =
  | string
  | Readonly<{
    /** Field of the component's static `Values` (`Suit`), or a longer path (`DynamicValues.Health`). */
    field?: string;
    /** Constant output. Wins over `field`. */
    value?: string | number | boolean;
    /** Mapping table from the field's value to the output. */
    map?: Readonly<Record<string, string | number | boolean>>;
    /** Output when the field is absent or has no `map` entry. */
    default?: string | number | boolean;
  }>;

export interface FaceSpec {
  /** Text placed in the component's default slot (the centre of a card face). */
  readonly text?: Source;
  /** Text placed in a named region of the element: a card's `footer` or `corner` slot. */
  readonly regions?: Readonly<Record<string, Source>>;
  /** Host properties of the component element, from an allow-list. */
  readonly props?: Readonly<Record<string, Source | number | boolean>>;
}

export interface PickSpec {
  /** Move proposed when a slot is activated. */
  readonly move: string;
  /** Name of the move argument that receives the slot value. */
  readonly arg?: string;
  /** What the slot value is: its index in the stack, or the component id. */
  readonly value?: 'index' | 'id';
  /** Constant arguments merged into every proposal. */
  readonly args?: Readonly<Record<string, string | number | boolean | null>>;
  /** Which slots are targets at all. Default `occupied`. */
  readonly slots?: 'all' | 'occupied' | 'hidden' | 'visible' | 'empty';
  /** Legal slot values, when legality is known as data. Absent means every target slot. */
  readonly candidates?: readonly (number | string)[];
  /** False disables every slot (not this viewer's turn, say). Default true. */
  readonly enabled?: boolean;
}

export type SlotKind = 'empty' | 'hidden' | 'visible';

export interface SlotContext {
  readonly kind: SlotKind;
  readonly index: number;
  /** The expanded component: null, `{ ID }`, or a visible component. */
  readonly component: Readonly<Record<string, unknown>> | null;
}

/**
 * Host properties a face spec may set, per component element. Anything else is
 * ignored: `item`, `id`, `index`, `disabled` and `spacer` belong to the stack,
 * and no entry here can carry markup or a function.
 */
const allowedProps: Readonly<Record<string, ReadonlySet<string>>> = {
  'boardgame-card': new Set([
    'suit', 'rank', 'rotated', 'tall', 'aspectRatio', 'art', 'artFit', 'backArt',
    'frontColor', 'inkColor', 'noShadow', 'altShadow',
  ]),
  'boardgame-token': new Set(['color', 'type', 'active', 'highlighted', 'noShadow', 'altShadow']),
};

/** Art is fetched by the trusted element, so an arbitrary URL is an exfiltration channel. */
const urlProps = new Set(['art', 'backArt']);
function safeAssetUrl(value: unknown): string {
  return typeof value === 'string' && /^(data:image\/|\.{0,2}\/[^/])/.test(value) ? value : '';
}

/**
 * Parse the `face` attribute. Three spellings, all data:
 *   face="Type"                      text of field Type
 *   face="suit:Suit rank:Rank"       property mapping, `prop:Field` pairs
 *   face='{"props":{...}}'           the full JSON form
 */
export function parseFace(input: unknown): FaceSpec | null {
  if (input === null || input === undefined || input === '') return null;
  if (typeof input === 'object') return input as FaceSpec;
  if (typeof input !== 'string') return null;
  const trimmed = input.trim();
  if (trimmed.startsWith('{')) return JSON.parse(trimmed) as FaceSpec;
  if (!trimmed.includes(':')) return { text: trimmed };
  const props: Record<string, Source | boolean> = {};
  for (const pair of trimmed.split(/\s+/)) {
    const [prop, field] = pair.split(':');
    if (!prop) continue;
    props[prop] = field === undefined || field === '' ? true : field;
  }
  return { props };
}

/**
 * Parse the `pick` attribute:
 *   pick="Reveal Card:CardIndex"     move name, then the argument that gets the slot index
 *   pick='{"move":"...","arg":"..."}'
 */
export function parsePick(input: unknown): PickSpec | null {
  if (input === null || input === undefined || input === '') return null;
  if (typeof input === 'object') return input as PickSpec;
  if (typeof input !== 'string') return null;
  const trimmed = input.trim();
  if (trimmed.startsWith('{')) return JSON.parse(trimmed) as PickSpec;
  const colon = trimmed.lastIndexOf(':');
  return colon < 0
    ? { move: trimmed }
    : { move: trimmed.slice(0, colon), arg: trimmed.slice(colon + 1) };
}

export function resolveSource(source: Source | number | boolean, slot: SlotContext): unknown {
  if (typeof source === 'number' || typeof source === 'boolean') return source;
  const spec = typeof source === 'string' ? { field: source } : source;
  if (spec.value !== undefined) return spec.value;
  let raw: unknown;
  if (spec.field === '$index') raw = slot.index;
  else if (spec.field === '$kind') raw = slot.kind;
  else if (spec.field && slot.kind === 'visible') {
    // A hidden or empty slot has no fields; that is the whole point of hiding.
    raw = resolvePath(slot.component, spec.field.includes('.') ? spec.field : `Values.${spec.field}`);
  }
  if (spec.map && raw !== undefined && raw !== null
    && Object.prototype.hasOwnProperty.call(spec.map, String(raw))) {
    return spec.map[String(raw)];
  }
  if (raw === undefined || raw === null) return spec.default;
  return spec.map ? (spec.default ?? raw) : raw;
}

/** Text for one region. Always a string; it is only ever bound as text. */
export function faceText(source: Source | undefined, slot: SlotContext): string {
  if (source === undefined) return '';
  const value = resolveSource(source, slot);
  return value === undefined || value === null ? '' : String(value);
}

/** Allow-listed host properties for one slot. */
export function faceProps(
  face: FaceSpec | null,
  tag: string,
  slot: SlotContext,
): Record<string, unknown> {
  const allowed = allowedProps[tag];
  const result: Record<string, unknown> = {};
  if (!face?.props || !allowed) return result;
  for (const [prop, source] of Object.entries(face.props)) {
    if (!allowed.has(prop)) continue;
    let value = resolveSource(source, slot);
    if (urlProps.has(prop)) value = safeAssetUrl(value);
    // A property with no value for this slot keeps the element's own default.
    if (value === undefined) continue;
    result[prop] = value;
  }
  return result;
}

export function slotKind(component: unknown): SlotKind {
  if (component === null || component === undefined) return 'empty';
  return typeof (component as { Values?: unknown }).Values === 'object' ? 'visible' : 'hidden';
}

function slotValue(pick: PickSpec, slot: SlotContext): number | string {
  return pick.value === 'id' ? String(slot.component?.ID ?? '') : slot.index;
}

/** Whether a slot is a target of the pick, and whether it may be activated now. */
export function pickState(
  pick: PickSpec | null,
  slot: SlotContext,
): Readonly<{ target: boolean; enabled: boolean }> {
  if (!pick || !pick.move) return { target: false, enabled: false };
  const which = pick.slots ?? 'occupied';
  const target = which === 'all'
    || (which === 'occupied' && slot.kind !== 'empty')
    || which === slot.kind;
  if (!target) return { target: false, enabled: false };
  if (pick.enabled === false) return { target: true, enabled: false };
  if (pick.candidates && !pick.candidates.includes(slotValue(pick, slot))) {
    return { target: true, enabled: false };
  }
  return { target: true, enabled: true };
}

/** The JSON detail of a `propose-move` event for one activated slot. */
export function proposal(
  pick: PickSpec,
  slot: SlotContext,
): Readonly<{ move: string; args: Record<string, unknown> }> {
  const args: Record<string, unknown> = { ...(pick.args ?? {}) };
  if (pick.arg) args[pick.arg] = slotValue(pick, slot);
  return { move: pick.move, args };
}
