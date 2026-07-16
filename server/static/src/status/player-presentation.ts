export const MAX_PLAYER_PRESENTATIONS = 128;

export interface PlayerPresentation {
  readonly playerIndex: number;
  readonly label: string;
  readonly color?: string;
}

export interface PlayerPresentationSource {
  readonly DisplayName?: unknown;
}

/** Normalize public server player metadata before it reaches game renderers. */
export function playerPresentations(
  players: readonly PlayerPresentationSource[],
  colors: readonly string[],
): readonly PlayerPresentation[] {
  if (!Array.isArray(players)) throw new Error('playerPresentations: players must be an array');
  if (!Array.isArray(colors)) throw new Error('playerPresentations: colors must be an array');
  if (players.length > MAX_PLAYER_PRESENTATIONS) {
    throw new Error(`playerPresentations: received ${players.length} players; maximum is ${MAX_PLAYER_PRESENTATIONS}`);
  }
  return validatePlayerPresentations(players.map((player, playerIndex) => {
    if (typeof player !== 'object' || player === null) {
      throw new Error(`playerPresentations: player ${playerIndex} must be an object`);
    }
    const rawLabel = player.DisplayName;
    const label = typeof rawLabel === 'string' && rawLabel.trim()
      ? rawLabel.trim()
      : `Player ${playerIndex + 1}`;
    if (label.length > 200) {
      throw new Error(`playerPresentations: player ${playerIndex} label exceeds 200 characters`);
    }
    const color = colors[playerIndex];
    if (color !== undefined && typeof color !== 'string') {
      throw new Error(`playerPresentations: color ${playerIndex} must be a string`);
    }
    return {
      playerIndex,
      label,
      ...(color?.trim() ? { color: color.trim() } : {}),
    };
  }));
}

/** Validate and defensively copy an author- or server-provided presentation list. */
export function validatePlayerPresentations(
  presentations: readonly PlayerPresentation[],
): readonly PlayerPresentation[] {
  if (!Array.isArray(presentations)) {
    throw new Error('playerPresentations: presentations must be an array');
  }
  if (presentations.length > MAX_PLAYER_PRESENTATIONS) {
    throw new Error(
      `playerPresentations: received ${presentations.length} players; maximum is ${MAX_PLAYER_PRESENTATIONS}`,
    );
  }
  return Object.freeze(presentations.map((presentation, position) => {
    if (typeof presentation !== 'object' || presentation === null) {
      throw new Error(`playerPresentations: presentation ${position} must be an object`);
    }
    if (presentation.playerIndex !== position) {
      throw new Error(
        `playerPresentations: presentation ${position} must have playerIndex ${position}, not ${JSON.stringify(presentation.playerIndex)}`,
      );
    }
    if (typeof presentation.label !== 'string' || !presentation.label.trim()) {
      throw new Error(`playerPresentations: player ${position} label must be a non-empty string`);
    }
    const label = presentation.label.trim();
    if (label.length > 200) {
      throw new Error(`playerPresentations: player ${position} label exceeds 200 characters`);
    }
    if (presentation.color !== undefined
      && (typeof presentation.color !== 'string' || !presentation.color.trim())) {
      throw new Error(`playerPresentations: color ${position} must be a non-empty string`);
    }
    return Object.freeze({
      playerIndex: position,
      label,
      ...(presentation.color === undefined ? {} : { color: presentation.color.trim() }),
    });
  }));
}

export function fallbackPlayerPresentation(playerIndex: number): PlayerPresentation {
  if (!Number.isSafeInteger(playerIndex) || playerIndex < 0) {
    throw new Error(`playerPresentation: index must be a non-negative safe integer, not ${JSON.stringify(playerIndex)}`);
  }
  return Object.freeze({ playerIndex, label: `Player ${playerIndex + 1}` });
}
