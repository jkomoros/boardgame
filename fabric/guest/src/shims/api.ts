/**
 * Stands in for server/static/src/api.ts inside the guest bundle.
 *
 * The renderer kit imports the HTTP API directly (see FINDINGS.md, F2 and
 * F4). A guest has no server, so every entry point here fails loudly and
 * reports itself; none of them may be reached while playing a game. The
 * list of names is whatever the import closure of `boardgame-render-game`
 * asks for, so a new server call in the kit shows up as a build error here.
 */

type Reporter = (name: string) => void;

let reporter: Reporter = () => {};
const calls: string[] = [];

/** The guest installs a reporter that forwards to the host as an `error`. */
export function setServerCallReporter(next: Reporter): void {
  reporter = next;
}

/** Names of every server entry point the client tried to use. */
export function serverCallsAttempted(): readonly string[] {
  return calls;
}

function refused(name: string): { status: number; error: string; friendlyError: string } {
  calls.push(name);
  reporter(name);
  return {
    status: 0,
    error: `${name}: the guest has no server`,
    friendlyError: 'This action is not available here',
  };
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  friendlyError?: string;
  code?: string;
  expectedVersion?: number;
  actualVersion?: number;
  failureBody?: Readonly<Record<string, unknown>>;
  aborted?: boolean;
  status: number;
}

export function buildGameUrl(name: string, id: string, path = ''): string {
  refused('buildGameUrl');
  return `guest-has-no-server:${name}/${id}/${path}`;
}

/** Used by the Table surface's host actions (lock room, skip turn, solo). */
export async function apiHttpPost(_url: string, ..._rest: unknown[]): Promise<ApiResponse<unknown>> {
  return refused('apiHttpPost');
}

/** Used by boardgame-render-game's board legality preview. */
export async function movePreviewBatch(..._args: unknown[]): Promise<ApiResponse<never>> {
  return refused('movePreviewBatch');
}
