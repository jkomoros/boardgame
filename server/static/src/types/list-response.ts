import type {
  GameListItem,
  ManagerAgentInfo,
  ManagerInfo,
  ManagerVariantInfo,
  ManagerVariantValue,
  PlayerInfo,
} from './store.js';

const MAX_MANAGERS = 1_000;
const MAX_GAMES = 1_000;
const MAX_PLAYERS = 1_000;
const MAX_OPTIONS = 10_000;

type RecordValue = Readonly<Record<string, unknown>>;

function record(value: unknown, path: string): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${path} must be an object`);
  }
  return value as RecordValue;
}

function array(value: unknown, path: string, maximum: number): readonly unknown[] {
  if (value === null) return [];
  if (!Array.isArray(value)) throw new Error(`${path} must be an array`);
  if (value.length > maximum) throw new Error(`${path} exceeds the maximum of ${maximum} entries`);
  return value;
}

function string(value: unknown, path: string, allowEmpty = false): string {
  if (typeof value !== 'string' || (!allowEmpty && !value.trim())) {
    throw new Error(`${path} must be ${allowEmpty ? 'a string' : 'a non-empty string'}`);
  }
  return value;
}

function boolean(value: unknown, path: string): boolean {
  if (typeof value !== 'boolean') throw new Error(`${path} must be a boolean`);
  return value;
}

function integer(value: unknown, path: string): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${path} must be a non-negative safe integer`);
  }
  return value;
}

function player(value: unknown, path: string): PlayerInfo {
  const item = record(value, path);
  if (item['PhotoURL'] !== undefined && typeof item['PhotoURL'] !== 'string') {
    throw new Error(`${path}.PhotoURL must be a string when present`);
  }
  return {
    IsEmpty: boolean(item['IsEmpty'], `${path}.IsEmpty`),
    IsAgent: boolean(item['IsAgent'], `${path}.IsAgent`),
    DisplayName: string(item['DisplayName'], `${path}.DisplayName`, true),
    ...(typeof item['PhotoURL'] === 'string' ? { PhotoURL: item['PhotoURL'] } : {}),
  };
}

function agent(value: unknown, path: string): ManagerAgentInfo {
  const item = record(value, path);
  return {
    Name: string(item['Name'], `${path}.Name`),
    DisplayName: string(item['DisplayName'], `${path}.DisplayName`),
  };
}

function variantValue(value: unknown, path: string): ManagerVariantValue {
  const item = record(value, path);
  return {
    Value: string(item['Value'], `${path}.Value`),
    DisplayName: string(item['DisplayName'], `${path}.DisplayName`),
    Description: string(item['Description'], `${path}.Description`, true),
  };
}

function variant(value: unknown, path: string): ManagerVariantInfo {
  const item = record(value, path);
  return {
    Name: string(item['Name'], `${path}.Name`),
    DisplayName: string(item['DisplayName'], `${path}.DisplayName`),
    Description: string(item['Description'], `${path}.Description`, true),
    Values: array(item['Values'], `${path}.Values`, MAX_OPTIONS)
      .map((entry, index) => variantValue(entry, `${path}.Values[${index}]`)),
  };
}

function manager(value: unknown, path: string): ManagerInfo {
  const item = record(value, path);
  const MinNumPlayers = integer(item['MinNumPlayers'], `${path}.MinNumPlayers`);
  const MaxNumPlayers = integer(item['MaxNumPlayers'], `${path}.MaxNumPlayers`);
  const DefaultNumPlayers = integer(item['DefaultNumPlayers'], `${path}.DefaultNumPlayers`);
  if (MinNumPlayers > MaxNumPlayers) throw new Error(`${path} has MinNumPlayers greater than MaxNumPlayers`);
  if (DefaultNumPlayers < MinNumPlayers || DefaultNumPlayers > MaxNumPlayers) {
    throw new Error(`${path}.DefaultNumPlayers must be within its player bounds`);
  }
  return {
    Name: string(item['Name'], `${path}.Name`),
    DisplayName: string(item['DisplayName'], `${path}.DisplayName`),
    Description: string(item['Description'], `${path}.Description`, true),
    DefaultNumPlayers,
    MinNumPlayers,
    MaxNumPlayers,
    Agents: array(item['Agents'], `${path}.Agents`, MAX_OPTIONS)
      .map((entry, index) => agent(entry, `${path}.Agents[${index}]`)),
    Variant: array(item['Variant'], `${path}.Variant`, MAX_OPTIONS)
      .map((entry, index) => variant(entry, `${path}.Variant[${index}]`)),
    SupportsTableHandMode: boolean(item['SupportsTableHandMode'], `${path}.SupportsTableHandMode`),
  };
}

function game(value: unknown, path: string): GameListItem {
  const item = record(value, path);
  // Players is omitted entirely (not even `null`) on entries coming from
  // AllGames: server/api/main.go's doListGames populates AllGames straight
  // from storage.ListGames's CombinedStorageRecord, which carries no
  // per-player roster -- only the Participating/Visible lists are enriched
  // with Players (see doListGames). Treat "absent" as "no roster to show"
  // rather than a decode error; a present-but-malformed Players array is
  // still rejected below.
  const players = item['Players'] === undefined
    ? []
    : array(item['Players'], `${path}.Players`, MAX_PLAYERS)
      .map((entry, index) => player(entry, `${path}.Players[${index}]`));
  return {
    ID: string(item['ID'], `${path}.ID`),
    Name: string(item['Name'], `${path}.Name`),
    Players: players,
    ReadableLastActivity: string(item['ReadableLastActivity'], `${path}.ReadableLastActivity`, true),
    Open: boolean(item['Open'], `${path}.Open`),
    Visible: boolean(item['Visible'], `${path}.Visible`),
  };
}

export function decodeManagersResponse(value: unknown): ManagerInfo[] {
  const item = record(value, 'Manager list response');
  if (item['Status'] !== 'Success') throw new Error('Manager list response.Status must be "Success"');
  return array(item['Managers'], 'Manager list response.Managers', MAX_MANAGERS)
    .map((entry, index) => manager(entry, `Manager list response.Managers[${index}]`));
}

export interface GamesListResponse {
  ParticipatingActiveGames: GameListItem[];
  ParticipatingFinishedGames: GameListItem[];
  VisibleActiveGames: GameListItem[];
  VisibleJoinableActiveGames: GameListItem[];
  AllGames: GameListItem[];
}

export function decodeGamesListResponse(value: unknown): GamesListResponse {
  const item = record(value, 'Games list response');
  if (item['Status'] !== 'Success') throw new Error('Games list response.Status must be "Success"');
  const games = (key: keyof GamesListResponse, optional = false): GameListItem[] => array(
    optional && item[key] === undefined ? null : item[key],
    `Games list response.${key}`,
    MAX_GAMES,
  ).map((entry, index) => game(entry, `Games list response.${key}[${index}]`));
  return {
    ParticipatingActiveGames: games('ParticipatingActiveGames'),
    ParticipatingFinishedGames: games('ParticipatingFinishedGames'),
    VisibleActiveGames: games('VisibleActiveGames'),
    VisibleJoinableActiveGames: games('VisibleJoinableActiveGames'),
    AllGames: games('AllGames', true),
  };
}

export interface CreateGameResponse {
  GameName: string;
  GameID: string;
  CompanionRoomCode?: string;
}

export function decodeCreateGameResponse(value: unknown): CreateGameResponse {
  const item = record(value, 'Create game response');
  if (item['Status'] !== 'Success') throw new Error('Create game response.Status must be "Success"');
  if (item['CompanionRoomCode'] !== undefined && typeof item['CompanionRoomCode'] !== 'string') {
    throw new Error('Create game response.CompanionRoomCode must be a string when present');
  }
  return {
    GameName: string(item['GameName'], 'Create game response.GameName'),
    GameID: string(item['GameID'], 'Create game response.GameID'),
    ...(typeof item['CompanionRoomCode'] === 'string'
      ? { CompanionRoomCode: item['CompanionRoomCode'] }
      : {}),
  };
}
