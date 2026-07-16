import type { UserInfo } from './store.js';

type RecordValue = Readonly<Record<string, unknown>>;

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

function optionalString(value: unknown, path: string): string | undefined {
  if (value === undefined) return undefined;
  return string(value, path, true);
}

function timestamp(value: unknown, path: string): number | undefined {
  if (value === undefined) return undefined;
  // Go emits Unix nanoseconds, which exceed JavaScript's safe-integer range.
  // The UI treats these as opaque ordering/display metadata, so finite and
  // non-negative is the honest wire contract.
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
    throw new Error(`${path} must be a non-negative finite number when present`);
  }
  return value;
}

function user(value: unknown, path: string): UserInfo | null {
  if (value === null || value === undefined) return null;
  const item = record(value, path);
  const DisplayName = optionalString(item['DisplayName'], `${path}.DisplayName`);
  const Email = optionalString(item['Email'], `${path}.Email`);
  const PhotoURL = optionalString(item['PhotoURL'], `${path}.PhotoURL`);
  const Created = timestamp(item['Created'], `${path}.Created`);
  const LastSeen = timestamp(item['LastSeen'], `${path}.LastSeen`);
  return {
    ID: string(item['ID'], `${path}.ID`),
    ...(DisplayName !== undefined ? { DisplayName } : {}),
    ...(Email !== undefined ? { Email } : {}),
    ...(PhotoURL !== undefined ? { PhotoURL } : {}),
    ...(Created !== undefined ? { Created } : {}),
    ...(LastSeen !== undefined ? { LastSeen } : {}),
  };
}

export interface AuthResponse {
  User: UserInfo | null;
  AdminAllowed: boolean;
  Message?: string;
}

export function decodeAuthResponse(value: unknown): AuthResponse {
  const item = record(value, 'Auth response');
  if (item['Status'] !== 'Success') throw new Error('Auth response.Status must be "Success"');
  if (item['AdminAllowed'] !== undefined && typeof item['AdminAllowed'] !== 'boolean') {
    throw new Error('Auth response.AdminAllowed must be a boolean when present');
  }
  const Message = optionalString(item['Message'], 'Auth response.Message');
  return {
    User: user(item['User'], 'Auth response.User'),
    AdminAllowed: item['AdminAllowed'] === true,
    ...(Message !== undefined ? { Message } : {}),
  };
}
