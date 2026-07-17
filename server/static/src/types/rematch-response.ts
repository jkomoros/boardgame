export interface RematchResponse {
  readonly ok: true;
  readonly gameID: string;
  readonly gameName: string;
  readonly roomCode: string;
}

function nonEmptyString(value: unknown, path: string): string {
  if (typeof value !== 'string' || !value.trim()) throw new Error(`${path} must be a non-empty string`);
  return value;
}

export function decodeRematchResponse(value: unknown): RematchResponse {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Rematch response must be an object');
  }
  const item = value as Readonly<Record<string, unknown>>;
  if (item['ok'] !== true) throw new Error('Rematch response.ok must be true');
  const gameID = nonEmptyString(item['gameID'], 'Rematch response.gameID');
  const gameName = nonEmptyString(item['gameName'], 'Rematch response.gameName');
  const roomCode = nonEmptyString(item['roomCode'], 'Rematch response.roomCode');
  if (!/^[A-HJKMNPQRSTUVWXY]{4,5}$/.test(roomCode)) {
    throw new Error('Rematch response.roomCode must be a canonical room code');
  }
  return { ok: true, gameID, gameName, roomCode };
}
