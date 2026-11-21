import type { ReactiveControllerHost } from 'lit';

/** Renderer identity shared by local interaction controllers. */
export interface GameSnapshotHost extends ReactiveControllerHost {
  readonly state: object | null;
  readonly gameName: string;
  readonly gameId: string;
  readonly gameVersion: number;
  readonly snapshotEpoch: number;
  readonly viewingAsPlayer: number;
  readonly proposingAsPlayer: number;
  readonly proposingAsAdmin: boolean;
}

export function gameSnapshotKey(host: GameSnapshotHost): string {
  return [
    host.gameName,
    host.gameId,
    host.gameVersion,
    host.snapshotEpoch,
    host.viewingAsPlayer,
    host.proposingAsPlayer,
    host.proposingAsAdmin ? 1 : 0,
  ].join('\u0000');
}
