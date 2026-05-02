/**
 * Shared types, constants, and detection logic for gathering components.
 *
 * This module centralizes the detection contract between the server-side
 * gathering moves and the client-side gathering panel. The contract is:
 * - A selection move has fields named TargetPlayerIndex + Selected{Team,Role,Color}
 * - The Selected* field has EnumName matching "team", "role", or "color"
 * - Available values come from Computed.Global.Available{Teams,Roles,Colors}
 * - Current selections come from Computed.Players[i].{Team,Role,Color}Value
 */
import type { MoveForm, MoveFormField } from '../types/api';

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

export interface PlayerInfo {
  IsEmpty: boolean;
  IsAgent: boolean;
  PhotoUrl?: string;
  DisplayName: string;
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

/** Get ReadyToStartError from computed global properties. */
export function getReadyToStartError(state: any): string {
  return state?.Game?.Computed?.Global?.ReadyToStartError || '';
}

/** Get available enum values from computed global properties. */
export function getAvailableValues(state: any, key: 'AvailableTeams' | 'AvailableRoles' | 'AvailableColors'): EnumValue[] {
  return state?.Game?.Computed?.Global?.[key] || [];
}

/** Get a player's computed selection value. */
export function getPlayerComputedValue(state: any, playerIndex: number, key: 'TeamValue' | 'RoleValue' | 'ColorValue'): string {
  const players = state?.Players;
  if (!players || !players[playerIndex]) return '';
  return players[playerIndex]?.Computed?.[key] || '';
}
