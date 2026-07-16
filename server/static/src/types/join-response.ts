type RecordValue = Readonly<Record<string, unknown>>;

const MAX_SEATS = 10_000;

function record(value: unknown, path: string): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${path} must be an object`);
  }
  return value as RecordValue;
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

function index(value: unknown, path: string): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${path} must be a non-negative safe integer`);
  }
  return value;
}

function count(value: unknown, path: string): number {
  const result = index(value, path);
  if (result > MAX_SEATS) throw new Error(`${path} must not exceed ${MAX_SEATS}`);
  return result;
}

export interface JoinResponse {
  gameID: string;
  gameName: string;
  gameDisplayName: string;
  minPlayers: number;
  maxPlayers: number;
  currentPlayers: number;
  requiresSeatPicker: boolean;
}

export interface SeatOptionsSlot {
  playerIndex: number;
  label: string;
  filled: boolean;
  avatarSlug?: string;
  displayName?: string;
}

export interface SeatOptionsResponse {
  gameID: string;
  gameName: string;
  slots: SeatOptionsSlot[];
  requiresSeatPicker: boolean;
}

export interface JoinSeatResponse {
  gameID: string;
  gameName: string;
  playerIndex: number;
}

export function decodeJoinResponse(value: unknown): JoinResponse {
  const item = record(value, 'Join response');
  const minPlayers = count(item['minPlayers'], 'Join response.minPlayers');
  const maxPlayers = count(item['maxPlayers'], 'Join response.maxPlayers');
  const currentPlayers = count(item['currentPlayers'], 'Join response.currentPlayers');
  if (minPlayers > maxPlayers) throw new Error('Join response.minPlayers must not exceed maxPlayers');
  if (currentPlayers > maxPlayers) throw new Error('Join response.currentPlayers must not exceed maxPlayers');
  return {
    gameID: string(item['gameID'], 'Join response.gameID'),
    gameName: string(item['gameName'], 'Join response.gameName'),
    gameDisplayName: string(item['gameDisplayName'], 'Join response.gameDisplayName'),
    minPlayers,
    maxPlayers,
    currentPlayers,
    requiresSeatPicker: boolean(item['requiresSeatPicker'], 'Join response.requiresSeatPicker'),
  };
}

export function decodeSeatOptionsResponse(value: unknown): SeatOptionsResponse {
  const item = record(value, 'Seat options response');
  if (!Array.isArray(item['slots']) || item['slots'].length > MAX_SEATS) {
    throw new Error(`Seat options response.slots must be an array of at most ${MAX_SEATS} entries`);
  }
  const slots = item['slots'].map((value, position): SeatOptionsSlot => {
    const slot = record(value, `Seat options response.slots[${position}]`);
    const playerIndex = index(slot['playerIndex'], `Seat options response.slots[${position}].playerIndex`);
    if (playerIndex !== position) {
      throw new Error(`Seat options response.slots[${position}].playerIndex must equal ${position}`);
    }
    const avatarSlug = slot['avatarSlug'] === undefined
      ? undefined
      : string(slot['avatarSlug'], `Seat options response.slots[${position}].avatarSlug`, true);
    const displayName = slot['displayName'] === undefined
      ? undefined
      : string(slot['displayName'], `Seat options response.slots[${position}].displayName`, true);
    return {
      playerIndex,
      label: string(slot['label'], `Seat options response.slots[${position}].label`),
      filled: boolean(slot['filled'], `Seat options response.slots[${position}].filled`),
      ...(avatarSlug !== undefined ? { avatarSlug } : {}),
      ...(displayName !== undefined ? { displayName } : {}),
    };
  });
  return {
    gameID: string(item['gameID'], 'Seat options response.gameID'),
    gameName: string(item['gameName'], 'Seat options response.gameName'),
    slots,
    requiresSeatPicker: boolean(item['requiresSeatPicker'], 'Seat options response.requiresSeatPicker'),
  };
}

export function decodeJoinSeatResponse(value: unknown): JoinSeatResponse {
  const item = record(value, 'Join seat response');
  return {
    gameID: string(item['gameID'], 'Join seat response.gameID'),
    gameName: string(item['gameName'], 'Join seat response.gameName'),
    playerIndex: index(item['playerIndex'], 'Join seat response.playerIndex'),
  };
}
