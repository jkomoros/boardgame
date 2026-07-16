const MAX_MESSAGES = 200;
const MAX_CHANNELS = 1_000;
const MAX_PLAYERS = 1_000;
const MAX_TEXT_LENGTH = 10_000;

type RecordValue = Readonly<Record<string, unknown>>;

function record(value: unknown, path: string): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${path} must be an object`);
  }
  return value as RecordValue;
}

function string(value: unknown, path: string, allowEmpty = false): string {
  if (typeof value !== 'string' || (!allowEmpty && !value.trim()) || value.length > MAX_TEXT_LENGTH) {
    throw new Error(`${path} must be ${allowEmpty ? 'a bounded string' : 'a bounded non-empty string'}`);
  }
  return value;
}

function boolean(value: unknown, path: string): boolean {
  if (typeof value !== 'boolean') throw new Error(`${path} must be a boolean`);
  return value;
}

function integer(value: unknown, path: string, minimum = 0): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < minimum) {
    throw new Error(`${path} must be a safe integer of at least ${minimum}`);
  }
  return value;
}

function stringArray(value: unknown, path: string, nullAsEmpty = false): string[] {
  if (value === null && nullAsEmpty) return [];
  if (!Array.isArray(value) || value.length > MAX_CHANNELS) {
    throw new Error(`${path} must be an array of at most ${MAX_CHANNELS} strings`);
  }
  const result = value.map((entry, index) => string(entry, `${path}[${index}]`));
  if (new Set(result).size !== result.length) throw new Error(`${path} must not contain duplicates`);
  return result;
}

export interface ChatMessage {
  id: string;
  channel: string;
  sender: number;
  body: string;
  timestamp: number;
}

export interface ChatConfig {
  Enabled: boolean;
  PrebakedOnly: boolean;
  AllowedMessages: string[];
}

export interface ChatReadResponse {
  Messages: ChatMessage[];
  ViewChannels: string[];
  SendChannels: string[];
  UserIDMap: Record<string, number>;
  ChatConfig: ChatConfig;
}

export interface ChatSendResponse {
  MessageID: string;
}

export function decodeChatReadResponse(value: unknown): ChatReadResponse {
  const item = record(value, 'Chat response');
  if (item['Status'] !== 'Success') throw new Error('Chat response.Status must be "Success"');
  const ViewChannels = stringArray(item['ViewChannels'], 'Chat response.ViewChannels', true);
  const SendChannels = stringArray(item['SendChannels'], 'Chat response.SendChannels', true);
  const visible = new Set(ViewChannels);
  for (const channel of SendChannels) {
    if (!visible.has(channel)) throw new Error(`Chat response.SendChannels contains non-viewable channel ${channel}`);
  }

  const rawMessages = item['Messages'] === null ? [] : item['Messages'];
  if (!Array.isArray(rawMessages) || rawMessages.length > MAX_MESSAGES) {
    throw new Error(`Chat response.Messages must be an array of at most ${MAX_MESSAGES} entries`);
  }
  const messageIDs = new Set<string>();
  const Messages = rawMessages.map((value, index): ChatMessage => {
    const message = record(value, `Chat response.Messages[${index}]`);
    const id = string(message['id'], `Chat response.Messages[${index}].id`);
    if (messageIDs.has(id)) throw new Error(`Chat response.Messages contains duplicate id ${id}`);
    messageIDs.add(id);
    const channel = string(message['channel'], `Chat response.Messages[${index}].channel`);
    if (!visible.has(channel)) {
      throw new Error(`Chat response.Messages[${index}].channel is not viewable`);
    }
    const sender = integer(message['sender'], `Chat response.Messages[${index}].sender`, -2);
    if (sender >= MAX_PLAYERS) throw new Error(`Chat response.Messages[${index}].sender is too large`);
    return {
      id,
      channel,
      sender,
      body: string(message['body'], `Chat response.Messages[${index}].body`, true),
      timestamp: integer(message['timestamp'], `Chat response.Messages[${index}].timestamp`),
    };
  });

  const userMap = record(item['UserIDMap'], 'Chat response.UserIDMap');
  if (Object.keys(userMap).length > MAX_PLAYERS) {
    throw new Error(`Chat response.UserIDMap must have at most ${MAX_PLAYERS} entries`);
  }
  const userEntries: Array<[string, number]> = [];
  const mappedPlayers = new Set<number>();
  for (const [userID, rawPlayer] of Object.entries(userMap)) {
    string(userID, 'Chat response.UserIDMap key');
    const player = integer(rawPlayer, `Chat response.UserIDMap.${userID}`);
    if (player >= MAX_PLAYERS) throw new Error(`Chat response.UserIDMap.${userID} is too large`);
    if (mappedPlayers.has(player)) throw new Error(`Chat response.UserIDMap maps player ${player} more than once`);
    mappedPlayers.add(player);
    userEntries.push([userID, player]);
  }
  // Object.fromEntries defines even special keys such as "__proto__" as own
  // data properties; bracket assignment on a normal object would mutate its
  // prototype instead.
  const UserIDMap: Record<string, number> = Object.fromEntries(userEntries);

  const rawConfig = record(item['ChatConfig'], 'Chat response.ChatConfig');
  const PrebakedOnly = boolean(rawConfig['PrebakedOnly'], 'Chat response.ChatConfig.PrebakedOnly');
  const AllowedMessages = stringArray(
    rawConfig['AllowedMessages'],
    'Chat response.ChatConfig.AllowedMessages',
    true,
  );
  if (!PrebakedOnly && AllowedMessages.length > 0) {
    throw new Error('Chat response.ChatConfig.AllowedMessages requires PrebakedOnly');
  }
  return {
    Messages,
    ViewChannels,
    SendChannels,
    UserIDMap,
    ChatConfig: {
      Enabled: boolean(rawConfig['Enabled'], 'Chat response.ChatConfig.Enabled'),
      PrebakedOnly,
      AllowedMessages,
    },
  };
}

export function decodeChatSendResponse(value: unknown): ChatSendResponse {
  const item = record(value, 'Chat send response');
  if (item['Status'] !== 'Success') throw new Error('Chat send response.Status must be "Success"');
  return { MessageID: string(item['MessageID'], 'Chat send response.MessageID') };
}
