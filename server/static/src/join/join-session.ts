export interface JoinOperation {
  readonly generation: number;
  readonly controller: AbortController;
}

/**
 * Owns the lifetime of one visible /join route. It deliberately contains no
 * UI state: its only job is to make it impossible for work started by an old
 * route or superseded request to commit into the current flow.
 */
export class JoinSessionScope {
  private routeKey: string | null = null;
  private generation = 0;
  private operation: AbortController | null = null;

  activate(routeKey: string): boolean {
    if (this.routeKey === routeKey) return false;
    this.invalidate();
    this.routeKey = routeKey;
    return true;
  }

  deactivate(): void {
    this.invalidate();
    this.routeKey = null;
  }

  begin(): JoinOperation {
    if (this.routeKey === null) throw new Error('Cannot begin a join operation for an inactive route');
    this.operation?.abort();
    const controller = new AbortController();
    this.operation = controller;
    return { generation: ++this.generation, controller };
  }

  isCurrent(operation: JoinOperation): boolean {
    return this.routeKey !== null
      && !operation.controller.signal.aborted
      && operation.generation === this.generation;
  }

  private invalidate(): void {
    this.generation++;
    this.operation?.abort();
    this.operation = null;
  }
}

export function codeFromJoinRoute(route: string): string | null {
  const query = route.includes('?') ? route.slice(route.indexOf('?') + 1) : route.replace(/^\?/, '');
  const code = new URLSearchParams(query).get('code');
  return code ? code.toUpperCase() : null;
}
