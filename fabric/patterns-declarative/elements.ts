/**
 * JSX types for the boardgame declarative elements (`fabric/declarative`),
 * for a Common Fabric pattern.
 *
 * A pattern imports this module for its side effect on the type checker:
 *
 *   import "../elements.ts";
 *
 * Common Fabric's `JSX.IntrinsicElements` lists the `cf-*` elements and has
 * no index signature, so a tag it does not list is a type error. The
 * interface is open to declaration merging from a pattern's own modules,
 * which is all this file does. It admits nothing: whether `<bg-table>` does
 * anything in a browser is decided by whether the page defined it.
 *
 * Every prop is a PROPERTY of the element (Common Fabric's renderer assigns
 * properties; only `data-*` and `aria-*` are attributes), so names are the
 * camelCase ones, not the attribute spellings the HTML pages use.
 */
import type { CellLike, Stream } from "commonfabric";

type In<T> = T | CellLike<T>;

/** What a handler bound to one of these elements receives. */
export interface ElementEvent<D> {
  type: string;
  detail: D;
}

type On<D> = Stream<ElementEvent<D>>;

/** `propose-move`: a person activated a slot or an action. */
export interface ProposeMoveDetail {
  move: string;
  args: Record<string, string | number | boolean | null>;
  stack?: string;
  index?: number;
  version: number | null;
}

/** `timer-elapsed`: a timer of the version on screen ran out by the browser's clock. */
export interface TimerElapsedDetail {
  timerId: string;
  version: number | null;
}

type SourceSpec = string | {
  field?: string;
  value?: string | number | boolean;
  map?: Record<string, string | number | boolean>;
  default?: string | number | boolean;
};

export interface FaceSpec {
  text?: SourceSpec;
  regions?: Record<string, SourceSpec>;
  props?: Record<string, SourceSpec | number | boolean>;
}

export interface PickSpec {
  move: string;
  arg?: string;
  value?: "index" | "id";
  args?: Record<string, string | number | boolean | null>;
  slots?: "all" | "occupied" | "hidden" | "visible" | "empty";
  candidates?: (number | string)[];
  enabled?: boolean;
}

interface Common {
  id?: string;
  className?: string;
  style?: string | Record<string, string | number>;
  children?: unknown;
}

interface BgTableProps extends Common {
  /** Stable identity; animation survives recreation of the element under it. */
  name?: string;
  gameName?: string;
  /** The game's component catalogue (`TableInfo.chest`). */
  chest?: In<unknown>;
  /** The released `TableView`, untouched. */
  tableView?: In<unknown>;
  /** The viewer's own principal; the table finds their seat. */
  principal?: In<string | undefined>;
  animationLength?: number;
  animationLengths?: Record<string, number>;
  holdBefore?: Record<string, number>;
  "onpropose-move"?: On<ProposeMoveDetail>;
  "ontimer-elapsed"?: On<TimerElapsedDetail>;
}

interface BgStackProps extends Common {
  path?: string;
  name?: In<string>;
  /** A wire stack given directly, for a stack that does not bind by `path`. */
  stack?: In<unknown>;
  /** The state version `stack` belongs to. */
  version?: In<number>;
  layout?: "stack" | "grid" | "fan" | "spread" | "pile" | string;
  messy?: boolean;
  messiness?: number;
  fauxComponents?: number;
  stagger?: number;
  noDefaultSpacer?: boolean;
  component?: "card" | "token";
  face?: string | FaceSpec;
  pick?: string | PickSpec;
  "onpropose-move"?: On<ProposeMoveDetail>;
}

declare global {
  namespace JSX {
    interface IntrinsicElements {
      "bg-table": BgTableProps;
      "bg-stack": BgStackProps;
      "bg-callout": Common & { message?: string; path?: string; name?: string };
      "bg-turn-status": Common & { playerLabels?: In<string[]> };
      "bg-seat": Common & { index?: number; flags?: string };
      "bg-value": Common & { path?: string };
      "bg-action": Common & {
        move?: string;
        args?: Record<string, string | number | boolean | null>;
        disabled?: boolean;
        "onpropose-move"?: On<ProposeMoveDetail>;
      };
    }
  }
}
