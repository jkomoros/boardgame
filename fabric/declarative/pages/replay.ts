/**
 * The one thing every page script does: parse a replay into frames.
 * No game logic lives here. A frame is exactly what the replay recorded for
 * one viewer at one version.
 */
import type { Frame } from '../src/index.js';

const checkedIn = new URL('../../../examples/memory/client/scenario-replay.json', import.meta.url);

export interface Replay {
  readonly chest: Record<string, unknown>;
  readonly gameName: string;
  readonly frames: readonly {
    readonly version: number;
    readonly label: string;
    readonly move?: string;
    readonly viewers: readonly {
      readonly viewer: number;
      readonly game: Frame['view'];
      readonly moveLegality?: Frame['legality'];
      readonly projectedMoveChoices?: Frame['choices'];
    }[];
  }[];
}

/** `scenario` is the checked-in Memory replay; any other name is a file in ./replays. */
export async function loadReplay(name: string): Promise<Replay> {
  const url = name === 'scenario' ? checkedIn : new URL(`../replays/${name}.json`, import.meta.url);
  return (await fetch(url)).json() as Promise<Replay>;
}

export function framesFor(replay: Replay, viewer: number): Frame[] {
  return replay.frames.map(frame => {
    const snapshot = frame.viewers.find(candidate => candidate.viewer === viewer)!;
    return {
      view: snapshot.game,
      // Between decision boundaries the engine projects no legality: nothing is legal.
      choices: snapshot.projectedMoveChoices ?? { Sets: [] },
      legality: snapshot.moveLegality ?? {},
      transition: { move: frame.move ?? '' },
    };
  });
}

/** Show `propose-move` events on the page, so a person and a test can both see them. */
export function logProposals(target: EventTarget, log: HTMLElement): void {
  const seen: unknown[] = [];
  (window as unknown as { proposals: unknown[] }).proposals = seen;
  target.addEventListener('propose-move', event => {
    const detail = (event as CustomEvent).detail;
    seen.push(detail);
    log.textContent = `propose-move ${JSON.stringify(detail)}\n` + (log.textContent ?? '');
  });
}
