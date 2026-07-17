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
  availableSeats: number;
  requiresSeatPicker: boolean;
  joinTicket: string;
}

export interface SeatOptionsSlot {
  playerIndex: number;
  label: string;
  filled: boolean;
  status: 'open' | 'human' | 'agent' | 'closed';
  available: boolean;
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
  resumed: boolean;
}

export function decodeJoinResponse(value: unknown): JoinResponse {
  const item = record(value, 'Join response');
  const minPlayers = count(item['minPlayers'], 'Join response.minPlayers');
  const maxPlayers = count(item['maxPlayers'], 'Join response.maxPlayers');
  const currentPlayers = count(item['currentPlayers'], 'Join response.currentPlayers');
  const availableSeats = count(item['availableSeats'], 'Join response.availableSeats');
  if (minPlayers > maxPlayers) throw new Error('Join response.minPlayers must not exceed maxPlayers');
  if (currentPlayers > maxPlayers) throw new Error('Join response.currentPlayers must not exceed maxPlayers');
  if (availableSeats > maxPlayers) throw new Error('Join response.availableSeats must not exceed maxPlayers');
  return {
    gameID: string(item['gameID'], 'Join response.gameID'),
    gameName: string(item['gameName'], 'Join response.gameName'),
    gameDisplayName: string(item['gameDisplayName'], 'Join response.gameDisplayName'),
    minPlayers,
    maxPlayers,
    currentPlayers,
    availableSeats,
    requiresSeatPicker: boolean(item['requiresSeatPicker'], 'Join response.requiresSeatPicker'),
    joinTicket: string(item['joinTicket'], 'Join response.joinTicket'),
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
    const filled = boolean(slot['filled'], `Seat options response.slots[${position}].filled`);
    const available = boolean(slot['available'], `Seat options response.slots[${position}].available`);
    const statusValue = string(slot['status'], `Seat options response.slots[${position}].status`);
    if (statusValue !== 'open' && statusValue !== 'human' && statusValue !== 'agent' && statusValue !== 'closed') {
      throw new Error(`Seat options response.slots[${position}].status was not recognized`);
    }
    const status: SeatOptionsSlot['status'] = statusValue;
    const canonical = status === 'open'
      ? !filled && available
      : status === 'closed'
        ? !filled && !available
        : filled && !available;
    if (!canonical) {
      throw new Error(`Seat options response.slots[${position}] contradicted status ${status}`);
    }
    return {
      playerIndex,
      label: string(slot['label'], `Seat options response.slots[${position}].label`),
      filled,
      status,
      available,
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
    resumed: boolean(item['resumed'], 'Join seat response.resumed'),
  };
}

export interface JoinProblem {
  code: string;
  error: string;
  slots?: SeatOptionsSlot[];
}

export function decodeJoinProblem(value: unknown): JoinProblem {
  const item = record(value, 'Join problem');
  const result: JoinProblem = {
    code: string(item['code'], 'Join problem.code'),
    error: string(item['error'], 'Join problem.error'),
  };
  if (item['slots'] !== undefined) {
    // Reuse the strict slot decoder by wrapping the conflict snapshot in the
    // successful seat-options shape. Identity fields are irrelevant here.
    result.slots = decodeSeatOptionsResponse({
      gameID: 'conflict',
      gameName: 'conflict',
      requiresSeatPicker: true,
      slots: item['slots'],
    }).slots;
  }
  return result;
}
