/**
 * Shared types, constants, and detection logic for gathering components.
 *
 * This module centralizes the detection contract between the server-side
 * gathering moves and the client-side gathering panel. The contract is:
 * - A selection move has fields named TargetPlayerIndex + Selected{Team,Role,Color}
 * - The Selected* field has EnumName matching "team", "role", or "color"
 * - Raw Computed.Global is expanded onto state.Game.Computed
 * - Raw Computed.Players[i] is expanded onto state.Players[i].Computed
 */
import type { MoveForm, MoveFormField } from '../types/api';
export type { PlayerInfo } from '../types/store';

// ---- Player index constants ----

export const OBSERVER_PLAYER_INDEX = -1;
export const ADMIN_PLAYER_INDEX = -2;
export const ANY_PLAYER_INDEX = -3;

// ---- Shared types ----

export interface EnumValue {
  Key: number;
  Name: string;
  /** CSS color string, present only for color enum values. */
  CSSColor?: string;
}

// ---- Start move detection ----

/**
 * @deprecated Use IsGatheringStart on MoveForm instead.
 * Kept as a fallback for servers that haven't been updated yet.
 */
export const START_MOVE_NAMES = new Set([
  'Confirm Players',
  'Close All Seats',
  'Start Game',
  'Begin Game',
  'Finalize Set Up',
]);

// ---- Selection move detection ----

/**
 * Checks if a move form matches the gathering selection signature:
 * has a TargetPlayerIndex field AND a field with the given name and EnumName.
 */
function hasSelectionSignature(form: MoveForm, fieldName: string, enumName: string): boolean {
  if (!form.Fields) return false;
  const hasTarget = form.Fields.some((f: MoveFormField) => f.Name === 'TargetPlayerIndex');
  const hasSelection = form.Fields.some((f: MoveFormField) => f.Name === fieldName && f.EnumName === enumName);
  return hasTarget && hasSelection;
}

/** Find the team selection move form by field signature. */
export function findTeamMoveForm(moveForms: MoveForm[] | null): MoveForm | null {
  if (!moveForms) return null;
  return moveForms.find(f => f.LegalForAnyone && hasSelectionSignature(f, 'SelectedTeam', 'team')) ?? null;
}

/** Find the role selection move form by field signature. */
export function findRoleMoveForm(moveForms: MoveForm[] | null): MoveForm | null {
  if (!moveForms) return null;
  return moveForms.find(f => f.LegalForAnyone && hasSelectionSignature(f, 'SelectedRole', 'role')) ?? null;
}

/** Find the color selection move form by field signature. */
export function findColorMoveForm(moveForms: MoveForm[] | null): MoveForm | null {
  if (!moveForms) return null;
  return moveForms.find(f => f.LegalForAnyone && hasSelectionSignature(f, 'SelectedColor', 'color')) ?? null;
}

/** Find the start/confirm move form, preferring the server-side IsGatheringStart
 *  marker with a fallback to name matching for backward compatibility. */
export function findStartMoveForm(moveForms: MoveForm[] | null): MoveForm | null {
  if (!moveForms) return null;
  // Prefer the server-side marker (no string matching needed)
  const byMarker = moveForms.find(f => f.IsGatheringStart && f.LegalForAnyone);
  if (byMarker) return byMarker;
  // Fallback to name matching for servers that don't yet send IsGatheringStart
  return moveForms.find(f => START_MOVE_NAMES.has(f.Name) && f.LegalForAnyone) ?? null;
}

/** Check if any gathering picker moves are present and legal. */
export function hasPickerMoves(moveForms: MoveForm[] | null): boolean {
  if (!moveForms) return false;
  return moveForms.some(f =>
    f.LegalForAnyone && f.Fields && (
      hasSelectionSignature(f, 'SelectedTeam', 'team') ||
      hasSelectionSignature(f, 'SelectedRole', 'role') ||
      hasSelectionSignature(f, 'SelectedColor', 'color')
    )
  );
}

// ---- State accessors ----

function optionalRecord(value: unknown, path: string): Readonly<Record<string, unknown>> | null {
  if (value === null || value === undefined) return null;
  if (typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`Gathering state ${path} must be an object when present`);
  }
  return value as Readonly<Record<string, unknown>>;
}

function gameComputed(state: unknown): Readonly<Record<string, unknown>> | null {
  const root = optionalRecord(state, 'root');
  const game = optionalRecord(root?.['Game'], 'Game');
  return optionalRecord(game?.['Computed'], 'Game.Computed');
}

/** Get ReadyToStartError from expanded global computed properties. */
export function getReadyToStartError(state: unknown): string {
  const value = gameComputed(state)?.['ReadyToStartError'];
  if (value === undefined || value === null) return '';
  if (typeof value !== 'string') {
    throw new Error('Gathering state Game.Computed.ReadyToStartError must be a string');
  }
  return value;
}

/** Get available enum values from expanded global computed properties. */
export function getAvailableValues(
  state: unknown,
  key: 'AvailableTeams' | 'AvailableRoles' | 'AvailableColors',
): EnumValue[] {
  const value = gameComputed(state)?.[key];
  if (value === undefined || value === null) return [];
  if (!Array.isArray(value)) {
    throw new Error(`Gathering state Game.Computed.${key} must be an array`);
  }
  return value.map((entry, index) => {
    const item = optionalRecord(entry, `Game.Computed.${key}[${index}]`);
    if (!item || !Number.isSafeInteger(item['Key']) || typeof item['Name'] !== 'string'
      || !item['Name'].trim()
      || (item['CSSColor'] !== undefined && typeof item['CSSColor'] !== 'string')) {
      throw new Error(`Gathering state Game.Computed.${key}[${index}] is not a valid enum value`);
    }
    return {
      Key: item['Key'] as number,
      Name: item['Name'],
      ...(item['CSSColor'] === undefined ? {} : { CSSColor: item['CSSColor'] }),
    };
  });
}

/** Get a player's expanded computed selection value. */
export function getPlayerComputedValue(
  state: unknown,
  playerIndex: number,
  key: 'TeamValue' | 'RoleValue' | 'ColorValue',
): string {
  if (!Number.isSafeInteger(playerIndex) || playerIndex < 0) {
    throw new Error(`Gathering player index must be a non-negative safe integer, got ${playerIndex}`);
  }
  const root = optionalRecord(state, 'root');
  const players = root?.['Players'];
  if (players === undefined || players === null) return '';
  if (!Array.isArray(players)) throw new Error('Gathering state Players must be an array');
  const player = optionalRecord(players[playerIndex], `Players[${playerIndex}]`);
  if (!player) return '';
  const computed = optionalRecord(player['Computed'], `Players[${playerIndex}].Computed`);
  const value = computed?.[key];
  if (value === undefined || value === null) return '';
  if (typeof value !== 'string') {
    throw new Error(`Gathering state Players[${playerIndex}].Computed.${key} must be a string`);
  }
  return value;
}
