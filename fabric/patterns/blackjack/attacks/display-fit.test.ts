/**
 * What each principal is shown of a table of blackjack at the runtime's
 * display boundary, and what the table's owner gets by going around the
 * table: binding its writer from another module, running the module a second
 * time over the live state, and writing the state with other code.
 *
 * This is labs' own display test for its card table
 * (`runtime/html/test/worker-display-fit-card-table.test.ts`) pointed at
 * `../table.tsx`. It needs labs' test harness, so it runs inside a labs
 * checkout that has the viewer-bound release, not from this repository:
 *
 *   cp fabric/patterns/blackjack/attacks/display-fit.test.ts \
 *     "$LABS/runtime/html/test/boardgame-blackjack-display-fit.test.ts"
 *   cd "$LABS/runtime/html" && BOARDGAME_ROOT=/path/to/boardgame \
 *     deno test -A test/boardgame-blackjack-display-fit.test.ts
 *
 * Alice owns the table's space and never sits. Bob and Carol play. Each is a
 * runtime of their own on one memory server. `SERVED=1` adds a serving
 * runtime that runs the table's handlers and computations for them, as a
 * deployment with server execution on does.
 */
import { expect } from "@std/expect";

import { cfcAtom } from "@commonfabric/api/cfc";
import { Identity } from "@commonfabric/identity";
import {
  ACLManager,
  type Cell,
  hostValueOf,
  readProjected,
  Runtime,
} from "@commonfabric/runner";
import {
  createRenderConfidentialityResolver,
  createRuntimeCfcModulePolicySource,
} from "@commonfabric/runner/cfc";
import { defer } from "@commonfabric/async/defer";

import { ExecutorHost } from "../../runner/src/executor/host.ts";
import { EmulatedStorageManager } from "../../runner/src/storage/v2-emulate.ts";
import { newSharedServer } from "../../runner/test/memory-v2-test-utils.ts";
import { readRefusal } from "../src/worker/display-fit.ts";
import type { RenderPolicy } from "../src/worker/types.ts";

const ROOT = Deno.env.get("BOARDGAME_ROOT");
if (!ROOT) throw new Error("set BOARDGAME_ROOT to a boardgame checkout");
const SERVED = Deno.env.get("SERVED") === "1";

const TABLE = "/fabric/patterns/blackjack/table.tsx";

/** `from`'s directory joined with a relative specifier. */
function resolved(from: string, specifier: string): string {
  const parts = from.split("/").slice(0, -1);
  for (const part of specifier.split("/")) {
    if (part === "..") parts.pop();
    else if (part !== ".") parts.push(part);
  }
  return parts.join("/");
}

/** The table's module and every file of this repository it imports. */
async function tableFiles(): Promise<{ name: string; contents: string }[]> {
  const files = new Map<string, string>();
  const queue = [TABLE];
  while (queue.length > 0) {
    const name = queue.pop()!;
    if (files.has(name)) continue;
    const source = await Deno.readTextFile(ROOT + name);
    files.set(
      name,
      name.endsWith(".tsx") ? `/// <cts-enable />\n${source}` : source,
    );
    for (const match of source.matchAll(/from\s+["'](\.[^"']+)["']/g)) {
      queue.push(resolved(name, match[1]));
    }
  }
  return [...files].map(([name, contents]) => ({ name, contents }));
}

const TABLE_FILES = await tableFiles();

interface Program {
  main: string;
  files: { name: string; contents: string }[];
}

const TABLE_PROGRAM: Program = { main: TABLE, files: TABLE_FILES };

/**
 * A pattern of another module that binds the table module's own `move`
 * handler to a table it is handed.
 */
const BORROWED_MOVE_PROGRAM: Program = {
  main: "/fabric/patterns/blackjack/borrow.tsx",
  files: [
    {
      name: "/fabric/patterns/blackjack/borrow.tsx",
      contents: `/// <cts-enable />
import { pattern } from "commonfabric";
import { move } from "./table.tsx";

// deno-lint-ignore no-explicit-any
export default pattern<{ theirs: any }>(({ theirs }) => ({
  move: move({ table: theirs }),
}));
`,
    },
    ...TABLE_FILES,
  ],
};

/** A pattern of another module that runs the table over a table it is handed. */
const LENT_STATE_PROGRAM: Program = {
  main: "/fabric/patterns/blackjack/lend.tsx",
  files: [
    {
      name: "/fabric/patterns/blackjack/lend.tsx",
      contents: `/// <cts-enable />
import { pattern } from "commonfabric";
import Table from "./table.tsx";

// deno-lint-ignore no-explicit-any
export default pattern<{ table: any }>(({ table }) => {
  const second = Table({ table } as Parameters<typeof Table>[0]);
  return { view: second.view, peek: second.peek, move: second.move };
});
`,
    },
    ...TABLE_FILES,
  ],
};

/**
 * A pattern of another module with a handler of its own that writes a table
 * it is handed: the sender into every seat.
 */
const FORGING_PROGRAM: Program = {
  main: "/fabric/patterns/blackjack/forge.tsx",
  files: [
    {
      name: "/fabric/patterns/blackjack/forge.tsx",
      contents: `/// <cts-enable />
import { handler, pattern, type Writable } from "commonfabric";

// deno-lint-ignore no-explicit-any
const forge = handler<{ occupant: string }, { table: Writable<any> }>(
  (event, { table }) => {
    const found = table.get();
    table.set({
      ...found,
      // deno-lint-ignore no-explicit-any
      seats: found.seats.map((seat: any) => ({ ...seat, principal: event.occupant })),
      revision: found.revision + 1,
    });
  },
);

// deno-lint-ignore no-explicit-any
export default pattern<{ theirs: any }>(({ theirs }) => ({
  forge: forge({ table: theirs }),
}));
`,
    },
    ...TABLE_FILES,
  ],
};

type Table = Cell<Record<string, unknown>>;

interface Seat {
  readonly name: string;
  readonly did: string;
  readonly runtime: Runtime;
  readonly storageManager: EmulatedStorageManager;
}

/** The part of a view this test reads. */
interface View {
  phase: string;
  revision: number;
  version: number;
  viewer: number;
  seats: { label: string }[];
  entries: {
    version: number;
    currentPlayerIndex: number;
    state: {
      Players: { HiddenHand: { Indexes: number[] }; Hand: { Indexes: number[] } }[];
    };
  }[];
}

const at = (table: Table, ...path: string[]): Cell<unknown> =>
  path.reduce(
    (cell, key) => cell.key(key as never) as Cell<unknown>,
    table as Cell<unknown>,
  );

const viewOf = (table: Table): Cell<View> => at(table, "view") as Cell<View>;

const until = async <T>(
  cell: Cell<T>,
  done: (value: T) => boolean,
): Promise<void> => {
  const reached = defer<void>();
  const cancel = cell.sink((value) => {
    if (done(value as T)) reached.resolve();
  });
  await reached.promise;
  cancel();
};

/** Resolves once `table`'s view is of the table after `revision` proposals. */
const seen = async (table: Table, revision: number): Promise<void> => {
  await until(viewOf(table), (view) => view?.revision === revision);
  await viewOf(table).pull();
};

/**
 * What keeps `viewer`'s principal from being shown `cell` as `holder`'s
 * runtime holds it, or undefined.
 */
const refusal = (
  viewer: Seat,
  cell: Cell<unknown>,
  space: string,
  holder: Seat = viewer,
) => {
  const policy: RenderPolicy = {
    declassifyConfidentiality: [],
    maxConfidentiality: [cfcAtom.user(viewer.did)],
  };
  const resolveConfidentiality = createRenderConfidentialityResolver({
    actingPrincipal: viewer.did,
    memberSpaces: [space],
    modulePolicyResolver:
      createRuntimeCfcModulePolicySource(holder.runtime).resolve,
  });
  const read = readProjected(cell, hostValueOf);
  return readRefusal(cell, [read.consumed], policy, { resolveConfidentiality });
};

const send = async (
  seat: Seat,
  table: Table,
  stream: string,
  event: unknown,
): Promise<unknown> => {
  const tx = seat.runtime.edit();
  // deno-lint-ignore no-explicit-any
  (table.withTx(tx) as any).key(stream).send(event);
  const error = (await tx.commit().settled).error;
  await seat.runtime.idle();
  await seat.storageManager.synced();
  await seat.runtime.idle();
  return error;
};

/** The hole cards `view` shows face up, by seat. */
const holeCards = (view: View | undefined): number[][] => {
  const head = view?.entries[view.entries.length - 1];
  return (head?.state.Players ?? []).map((player) =>
    player.HiddenHand.Indexes.filter((index) => index >= 0)
  );
};

const atTable = async (
  body: (world: {
    space: string;
    alice: Seat;
    bob: Seat;
    carol: Seat;
    table: Table;
    join: (seat: Seat, table: Table) => Promise<Table>;
    secondRun: (
      argument: Record<string, unknown>,
      program?: Program,
    ) => Promise<Table>;
  }) => Promise<void>,
): Promise<void> => {
  const names = ["alice", "bob", "carol"];
  const identities = await Promise.all(
    names.map((name) => Identity.fromPassphrase(`boardgame blackjack ${name}`)),
  );
  const experimental = { viewerBoundRelease: true, serverExecution: SERVED };
  const space = identities[0].did();
  const server = newSharedServer({ subscriptionRefreshDelayMs: 0 });
  const [alice, bob, carol] = identities.map((identity, index): Seat => {
    const storageManager = EmulatedStorageManager.connectTo(server, {
      as: identity,
    });
    const runtime = new Runtime({
      storageManager,
      apiUrl: new URL(import.meta.url),
      cfcEnforcementMode: "enforce-strict",
      cfcFlowLabels: "persist",
      experimental,
    });
    return { name: names[index], did: identity.did(), runtime, storageManager };
  });
  const service = await Identity.fromPassphrase("boardgame blackjack service");
  const host = SERVED
    ? new ExecutorHost({
      server,
      serviceIdentity: service.did(),
      // deno-lint-ignore require-await
      createRuntime: async () => {
        const storageManager = EmulatedStorageManager.connectTo(server, {
          as: service,
        });
        const runtime = new Runtime({
          storageManager,
          apiUrl: new URL(import.meta.url),
          servingPosture: true,
          cfcEnforcementMode: "enforce-strict",
          cfcFlowLabels: "persist",
          experimental,
        });
        return {
          runtime,
          dispose: async () => {
            await runtime.dispose();
            await storageManager.close();
          },
        };
      },
      policy: { flushDeadlineMs: 5_000, idleParkMs: 600_000 },
    })
    : undefined;
  try {
    const acl = new ACLManager(alice.runtime, space);
    await acl.set(alice.did as never, "OWNER");
    await acl.set(bob.did as never, "WRITE");
    await acl.set(carol.did as never, "WRITE");

    let runs = 0;
    const run = async (
      argument: Record<string, unknown>,
      program: Program = TABLE_PROGRAM,
    ): Promise<Table> => {
      const tx = alice.runtime.edit();
      const pattern = await alice.runtime.patternManager.compilePattern(
        program,
        { space: space as never, tx },
      );
      const result = alice.runtime.getCell<Record<string, unknown>>(
        space as never,
        `blackjack table ${runs++}`,
        undefined,
        tx,
      );
      // deno-lint-ignore no-explicit-any
      alice.runtime.run(tx, pattern as any, argument, result);
      alice.runtime.prepareTxForCommit(tx);
      expect((await tx.commit().settled).error).toBeUndefined();
      await result.pull();
      await alice.runtime.idle();
      await alice.runtime.patternManager.flushCompileCacheWrites();
      await alice.storageManager.synced();
      await alice.runtime.idle();
      return result;
    };
    const join = async (seat: Seat, table: Table): Promise<Table> => {
      const mine = seat.runtime.getCellFromLink(
        table.getAsNormalizedFullLink(),
      ) as Table;
      await mine.sync();
      expect(await seat.runtime.start(mine)).toBeTruthy();
      await seat.runtime.idle();
      await seat.storageManager.synced();
      return mine;
    };
    await body({
      space,
      alice,
      bob,
      carol,
      table: await run({}),
      join,
      secondRun: run,
    });
  } finally {
    await host?.close();
    for (const seat of [alice, bob, carol]) {
      await seat.runtime.idle();
      await seat.runtime.dispose();
      await seat.storageManager.close();
    }
    await server.close();
  }
};

/** Bob and Carol sit and Bob deals: three proposals. Alice only watches. */
const deal = async (
  { bob, carol }: { bob: Seat; carol: Seat },
  table: Table,
  bobs: Table,
  carols: Table,
): Promise<number> => {
  expect(await send(bob, bobs, "move", { kind: "sit", label: "Bob" }))
    .toBeUndefined();
  await seen(carols, 1);
  expect(await send(carol, carols, "move", { kind: "sit", label: "Carol" }))
    .toBeUndefined();
  await seen(bobs, 2);
  expect(
    await send(bob, bobs, "move", {
      kind: "start",
      options: { maxrounds: "1" },
    }),
  ).toBeUndefined();
  for (const one of [table, bobs, carols]) await seen(one, 3);
  return 3;
};

const label = SERVED ? " (served)" : "";

Deno.test(`blackjack at the display boundary${label}`, async (t) => {
  await atTable(async ({ space, alice, bob, carol, table, join }) => {
    const bobs = await join(bob, table);
    const carols = await join(carol, table);
    await deal({ bob, carol }, table, bobs, carols);

    await t.step("shows each player their own hole card and no other", () => {
      // Bob has the first turn, so his hole card is already turned over;
      // Carol's is still hidden, and only her view shows it.
      const bobSees = holeCards(viewOf(bobs).get());
      const carolSees = holeCards(viewOf(carols).get());
      expect(bobSees.map((cards) => cards.length)).toEqual([0, 0]);
      expect(carolSees.map((cards) => cards.length)).toEqual([0, 1]);
      expect(refusal(bob, viewOf(bobs), space)).toBeUndefined();
      expect(refusal(carol, viewOf(carols), space)).toBeUndefined();
    });

    await t.step("shows the owner, who holds no seat, no hole card", () => {
      const ownerSees = holeCards(viewOf(table).get());
      expect(ownerSees.map((cards) => cards.length)).toEqual([0, 0]);
      expect(refusal(alice, viewOf(table), space)).toBeUndefined();
    });

    await t.step("shows nobody another principal's view", () => {
      expect(refusal(alice, viewOf(bobs), space, bob)).toBeDefined();
      expect(refusal(alice, viewOf(carols), space, carol)).toBeDefined();
      expect(refusal(carol, viewOf(bobs), space, bob)).toBeDefined();
      expect(refusal(bob, viewOf(carols), space, carol)).toBeDefined();
      expect(refusal(bob, viewOf(table), space, alice)).toBeDefined();
    });

    await t.step("shows nobody the table or its salt", async () => {
      const state = table.getArgumentCell<Record<string, unknown>>()!.key(
        "table",
      );
      await state.pull();
      await at(table, "peek").pull();
      for (const seat of [alice, bob, carol]) {
        expect(refusal(seat, state, space, alice)).toBeDefined();
        expect(refusal(seat, at(table, "peek"), space, alice)).toBeDefined();
      }
    });
  });
});

Deno.test(
  `blackjack against a second binding of its writer${label}`,
  async (t) => {
    await atTable(async ({ space, alice, bob, carol, table, join, secondRun }) => {
      const bobs = await join(bob, table);
      const carols = await join(carol, table);
      expect(await send(bob, bobs, "move", { kind: "sit", label: "Bob" }))
        .toBeUndefined();
      await seen(table, 1);
      const state = table.getArgumentCell<Record<string, unknown>>()!.key(
        "table",
      );
      const borrowed = await secondRun({ theirs: state }, BORROWED_MOVE_PROGRAM);

      await t.step("makes the owner's own proposals and no others", async () => {
        // A seat in Bob's name: the proposer is the owner, whatever the
        // request says, so this seats the owner, as sitting at the table
        // would.
        await send(alice, borrowed, "move", {
          kind: "sit",
          label: "Eve",
          proposer: bob.did,
          principal: bob.did,
        });
        await seen(table, 2);
        const mine = viewOf(table).get();
        expect(mine.seats.map((seat) => seat.label)).toEqual(["Bob", "Eve"]);
        expect(mine.viewer).toBe(1);
        expect(viewOf(bobs).get().viewer).toBe(0);
        // She leaves again, so that she only watches the deal.
        await send(alice, borrowed, "move", { kind: "leave" });
        await seen(table, 3);
        await seen(carols, 3);
        expect(await send(carol, carols, "move", { kind: "sit", label: "Carol" }))
          .toBeUndefined();
        await seen(bobs, 4);
        // The deal is a seated player's: the owner's is refused.
        await send(alice, borrowed, "move", { kind: "start" });
        expect(await send(bob, bobs, "move", { kind: "start" })).toBeUndefined();
        for (const one of [table, bobs, carols]) await seen(one, 5);
        // A move in the game, through the borrowed handler, by nobody seated.
        await send(alice, borrowed, "move", {
          kind: "play",
          name: "Current Player Hit",
          arguments: {},
        });
        await viewOf(table).pull();
        expect(viewOf(table).get().revision).toBe(5);
      });

      await t.step("shows the owner no hole card", () => {
        expect(holeCards(viewOf(table).get()).map((cards) => cards.length))
          .toEqual([0, 0]);
        expect(refusal(alice, viewOf(bobs), space, bob)).toBeDefined();
        expect(refusal(alice, viewOf(carols), space, carol)).toBeDefined();
      });

      await t.step("shows the players the hands it dealt", () => {
        expect(holeCards(viewOf(carols).get())[1].length).toBe(1);
        expect(holeCards(viewOf(bobs).get())[1].length).toBe(0);
        expect(refusal(bob, viewOf(bobs), space)).toBeUndefined();
        expect(refusal(carol, viewOf(carols), space)).toBeUndefined();
      });
    });
  },
);

Deno.test(`blackjack against a second run on its table${label}`, async (t) => {
  await atTable(async ({ space, alice, bob, carol, table, join, secondRun }) => {
    const bobs = await join(bob, table);
    const carols = await join(carol, table);
    const proposals = await deal({ bob, carol }, table, bobs, carols);
    const state = table.getArgumentCell<Record<string, unknown>>()!.key(
      "table",
    );
    const second = await secondRun({ table: state }, LENT_STATE_PROGRAM);

    await t.step("computes the owner's own view again", async () => {
      await seen(second, proposals);
      expect(holeCards(viewOf(second).get()).map((cards) => cards.length))
        .toEqual([0, 0]);
    });

    await t.step("shows the owner none of it", async () => {
      expect(refusal(alice, viewOf(second), space)).toBeDefined();
      await at(second, "peek").pull();
      expect(refusal(alice, at(second, "peek"), space)).toBeDefined();
    });

    await t.step("leaves the live table as it was", () => {
      expect(refusal(alice, viewOf(table), space)).toBeUndefined();
      expect(refusal(bob, viewOf(bobs), space)).toBeUndefined();
      expect(refusal(alice, viewOf(bobs), space, bob)).toBeDefined();
    });
  });
});

Deno.test(`blackjack against a write by other code${label}`, async (t) => {
  await atTable(async ({ space, alice, bob, carol, table, join, secondRun }) => {
    const bobs = await join(bob, table);
    const carols = await join(carol, table);
    const proposals = await deal({ bob, carol }, table, bobs, carols);
    const state = table.getArgumentCell<Record<string, unknown>>()!.key(
      "table",
    );
    const forging = await secondRun({ theirs: state }, FORGING_PROGRAM);

    await t.step("leaves the table as it was", async () => {
      await send(alice, forging, "forge", { occupant: alice.did });
      await viewOf(table).pull();
      const view = viewOf(table).get();
      expect(view.revision).toBe(proposals);
      expect(view.seats.map((seat) => seat.label)).toEqual(["Bob", "Carol"]);
      // Nobody's seat changed hands: the owner still holds none.
      expect(view.viewer).toBe(-1);
      expect(viewOf(bobs).get().viewer).toBe(0);
      expect(viewOf(carols).get().viewer).toBe(1);
    });

    await t.step("shows the owner no hole card", () => {
      expect(holeCards(viewOf(table).get()).map((cards) => cards.length))
        .toEqual([0, 0]);
      expect(refusal(alice, viewOf(bobs), space, bob)).toBeDefined();
    });

    await t.step("shows the players their hands still", () => {
      expect(holeCards(viewOf(carols).get())[1].length).toBe(1);
        expect(holeCards(viewOf(bobs).get())[1].length).toBe(0);
      expect(refusal(bob, viewOf(bobs), space)).toBeUndefined();
      expect(refusal(carol, viewOf(carols), space)).toBeUndefined();
    });
  });
});
