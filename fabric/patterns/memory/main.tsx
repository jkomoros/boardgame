/**
 * Memory at one table, for two to six people, drawn by the boardgame client
 * in a sandboxed `cf-iframe`.
 *
 * The pattern holds one sealed table (`table.tsx`) and hands the guest the
 * view computed for whoever is looking, the game's public facts, and the one
 * verb a browser may send. The guest is
 * the generic boardgame guest of `fabric/guest`; nothing in it is specific
 * to Memory but the renderer compiled into it.
 *
 * The view is handed over whole. A value this pattern derived from it would
 * be sealed again.
 */
import {
  computed,
  dataFile,
  type Default,
  NAME,
  pattern,
  type Stream,
  UI,
  type VNode,
  type Writable,
} from "commonfabric";

import type { Proposal, TableInfo, TableView } from "../toolkit/view.ts";
import { memoryInfo } from "./game.ts";
import { move, type SealedTable, viewFor } from "./table.tsx";

/** The guest's document and bundle, written by `npm run build:fabric`. */
const GUEST_HTML = dataFile("../guest/guest.html");
const GUEST_SCRIPT = dataFile("../guest/guest.js");

export interface MemoryInput {
  /** The sealed table, which starts as nothing. */
  table: Writable<Default<SealedTable, null>>;
}

/** A table of Memory. */
export interface MemoryOutput {
  [NAME]: string;
  [UI]: VNode;
  /** What the principal looking may see of the table. */
  view: TableView;
  /** The deck and message texts, which are public. */
  info: TableInfo;
  /** Propose something as whoever sends it. */
  move: Stream<Proposal>;
}

export default pattern<MemoryInput, MemoryOutput>(({ table }) => {
  const view = viewFor({ table });
  const propose = move({ table });
  const info = computed(() => memoryInfo());
  return {
    [NAME]: "Memory",
    [UI]: (
      <div style={{ width: "100%", height: "100vh" }}>
        <cf-iframe
          src={GUEST_HTML}
          script={GUEST_SCRIPT}
          $context={{ view, info, move: propose }}
          resourceKinds={{
            view: "readonly",
            info: "readonly",
            move: "stream",
          }}
        />
      </div>
    ),
    view,
    info,
    move: propose,
  };
});
