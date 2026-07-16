import type {
  CompanionInfo,
  CompanionSeatPresentation,
  GameChest,
  PlayerInfo,
} from './store.js';
import type {
  GameInfoResponse,
  GameVersionResponse,
  JsonValue,
  MoveForm,
  MoveFormField,
  PreconditionEntry,
  ServerStateBundle,
} from './api.js';
import type { GameFromServer, RawGameState, TimerInfo } from './game-state.js';

const MAX_PLAYERS = 1_000;
const MAX_FORMS = 10_000;
const MAX_BUNDLES = 10_000;

type RecordValue = Readonly<Record<string, unknown>>;
type JsonObject = { [key: string]: JsonValue };

function record(value: unknown, path: string): RecordValue {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${path} must be an object`);
  }
  return value as RecordValue;
}

function array(value: unknown, path: string, maximum: number): readonly unknown[] {
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

function integer(value: unknown, path: string, nonNegative = false): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || (nonNegative && value < 0)) {
    throw new Error(`${path} must be ${nonNegative ? 'a non-negative ' : 'a '}safe integer`);
  }
  return value;
}

function optionalString(value: unknown, path: string): void {
  if (value !== undefined && typeof value !== 'string') throw new Error(`${path} must be a string when present`);
}

function optionalBoolean(value: unknown, path: string): void {
  if (value !== undefined && typeof value !== 'boolean') throw new Error(`${path} must be a boolean when present`);
}

function jsonValue(value: unknown, path: string, depth = 0): JsonValue {
  if (depth > 20) throw new Error(`${path} exceeds the maximum JSON nesting depth`);
  if (value === null || typeof value === 'string' || typeof value === 'boolean') return value;
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (Array.isArray(value)) {
    if (value.length > MAX_FORMS) throw new Error(`${path} exceeds the maximum of ${MAX_FORMS} entries`);
    return value.map((entry, index) => jsonValue(entry, `${path}[${index}]`, depth + 1));
  }
  const item = record(value, path);
  if (Object.keys(item).length > MAX_FORMS) {
    throw new Error(`${path} exceeds the maximum of ${MAX_FORMS} properties`);
  }
  return Object.fromEntries(Object.entries(item).map(([key, entry]) => [
    key,
    jsonValue(entry, `${path}.${key}`, depth + 1),
  ]));
}

function jsonObject(value: unknown, path: string): JsonObject {
  record(value, path);
  return jsonValue(value, path) as JsonObject;
}

function decodeStringMap(value: unknown, path: string): Record<string, string> {
  const item = record(value, path);
  if (Object.keys(item).length > MAX_FORMS) {
    throw new Error(`${path} exceeds the maximum of ${MAX_FORMS} properties`);
  }
  return Object.fromEntries(Object.entries(item).map(([key, entry]) => [
    key,
    string(entry, `${path}.${key}`, true),
  ]));
}

function decodeDecks(value: unknown, path: string): Record<string, readonly unknown[]> {
  const item = record(value, path);
  if (Object.keys(item).length > MAX_FORMS) {
    throw new Error(`${path} exceeds the maximum of ${MAX_FORMS} decks`);
  }
  return Object.fromEntries(Object.entries(item).map(([deckName, rawDeck]) => [
    deckName,
    array(rawDeck, `${path}.${deckName}`, MAX_FORMS).map((rawComponent, componentIndex) => {
      const component = record(rawComponent, `${path}.${deckName}[${componentIndex}]`);
      const Index = integer(component['Index'], `${path}.${deckName}[${componentIndex}].Index`, true);
      if (Index !== componentIndex) {
        throw new Error(
          `${path}.${deckName}[${componentIndex}].Index must equal its canonical index ${componentIndex}`,
        );
      }
      return {
        Index,
        Values: jsonObject(component['Values'], `${path}.${deckName}[${componentIndex}].Values`),
      };
    }),
  ]));
}

function decodeEnums(value: unknown, path: string): Record<string, { Values?: Record<string, string> }> {
  const item = record(value, path);
  if (Object.keys(item).length > MAX_FORMS) {
    throw new Error(`${path} exceeds the maximum of ${MAX_FORMS} enums`);
  }
  return Object.fromEntries(Object.entries(item).map(([enumName, rawEnum]) => {
    const enumItem = record(rawEnum, `${path}.${enumName}`);
    const Values = enumItem['Values'] === undefined
      ? undefined
      : decodeStringMap(enumItem['Values'], `${path}.${enumName}.Values`);
    return [enumName, Values === undefined ? {} : { Values }];
  }));
}

function decodeChest(value: unknown, path: string): GameChest {
  const item = record(value, path);
  const Decks = item['Decks'] === undefined
    ? undefined
    : item['Decks'] === null ? {} : decodeDecks(item['Decks'], `${path}.Decks`);
  const Enums = item['Enums'] === undefined
    ? undefined
    : item['Enums'] === null ? {} : decodeEnums(item['Enums'], `${path}.Enums`);
  const Constants = item['Constants'] === undefined
    ? undefined
    : item['Constants'] === null ? {} : jsonObject(item['Constants'], `${path}.Constants`);
  const LegalTemplates = item['LegalTemplates'] === undefined
    ? undefined
    : decodeStringMap(item['LegalTemplates'], `${path}.LegalTemplates`);
  return {
    ...(Decks === undefined ? {} : { Decks }),
    ...(Enums === undefined ? {} : { Enums }),
    ...(Constants === undefined ? {} : { Constants }),
    ...(LegalTemplates === undefined ? {} : { LegalTemplates }),
  } as unknown as GameChest;
}

function decodePlayer(value: unknown, path: string): PlayerInfo {
  const item = record(value, path);
  const IsEmpty = boolean(item['IsEmpty'], `${path}.IsEmpty`);
  const IsAgent = boolean(item['IsAgent'], `${path}.IsAgent`);
  const DisplayName = string(item['DisplayName'], `${path}.DisplayName`, true);
  optionalString(item['PhotoURL'], `${path}.PhotoURL`);
  return {
    IsEmpty,
    IsAgent,
    DisplayName,
    ...(typeof item['PhotoURL'] === 'string' ? { PhotoURL: item['PhotoURL'] } : {}),
  };
}

function decodeSeat(value: unknown, path: string): CompanionSeatPresentation {
  const item = record(value, path);
  return {
    playerIndex: integer(item['playerIndex'], `${path}.playerIndex`, true),
    displayName: string(item['displayName'], `${path}.displayName`),
    avatarSlug: string(item['avatarSlug'], `${path}.avatarSlug`),
  };
}

function decodeCompanion(value: unknown, path: string): CompanionInfo | null {
  if (value === null || value === undefined) return null;
  const item = record(value, path);
  const CompanionMode = boolean(item['CompanionMode'], `${path}.CompanionMode`);
  const RoomCode = string(item['RoomCode'], `${path}.RoomCode`, true);
  const RoomLocked = boolean(item['RoomLocked'], `${path}.RoomLocked`);
  const SeatPresentations = item['SeatPresentations'] === null || item['SeatPresentations'] === undefined
    ? []
    : array(item['SeatPresentations'], `${path}.SeatPresentations`, MAX_PLAYERS)
      .map((seat, index) => decodeSeat(seat, `${path}.SeatPresentations[${index}]`));
  const Absent = item['Absent'] === null || item['Absent'] === undefined
    ? []
    : array(item['Absent'], `${path}.Absent`, MAX_PLAYERS)
      .map((playerIndex, index) => integer(playerIndex, `${path}.Absent[${index}]`, true));
  optionalBoolean(item['IsHost'], `${path}.IsHost`);
  return {
    CompanionMode,
    RoomCode,
    RoomLocked,
    SeatPresentations,
    Absent,
    ...(typeof item['IsHost'] === 'boolean' ? { IsHost: item['IsHost'] } : {}),
  } as CompanionInfo;
}

function decodeField(value: unknown, path: string): MoveFormField {
  const item = record(value, path);
  const Name = string(item['Name'], `${path}.Name`);
  const Type = integer(item['Type'], `${path}.Type`, true);
  const DefaultValue = jsonValue(item['DefaultValue'], `${path}.DefaultValue`);
  optionalString(item['EnumName'], `${path}.EnumName`);
  if (item['Enum'] !== undefined) record(item['Enum'], `${path}.Enum`);
  return {
    Name,
    Type,
    DefaultValue,
    ...(typeof item['EnumName'] === 'string' ? { EnumName: item['EnumName'] } : {}),
  };
}

function decodePrecondition(value: unknown, path: string): PreconditionEntry {
  const item = record(value, path);
  const name = string(item['name'], `${path}.name`);
  let args: string[] | undefined;
  if (item['args'] !== undefined) {
    args = array(item['args'], `${path}.args`, 1_000)
      .map((arg, index) => string(arg, `${path}.args[${index}]`, true));
  }
  if (item['verdict'] !== 'pass' && item['verdict'] !== 'fail' && item['verdict'] !== 'unknown') {
    throw new Error(`${path}.verdict must be "pass", "fail", or "unknown"`);
  }
  const verdict = item['verdict'];
  const evaluable = boolean(item['evaluable'], `${path}.evaluable`);
  optionalBoolean(item['provisional'], `${path}.provisional`);
  let message: PreconditionEntry['message'];
  if (item['message'] !== undefined) {
    const rawMessage = record(item['message'], `${path}.message`);
    const template = string(rawMessage['template'], `${path}.message.template`);
    let bindings: Record<string, string | number | boolean> | undefined;
    if (rawMessage['bindings'] !== undefined) {
      const rawBindings = record(rawMessage['bindings'], `${path}.message.bindings`);
      bindings = {};
      for (const [bindingName, binding] of Object.entries(rawBindings)) {
        if (typeof binding !== 'string' && typeof binding !== 'number' && typeof binding !== 'boolean') {
          throw new Error(`${path}.message.bindings.${bindingName} must be a string, number, or boolean`);
        }
        bindings[bindingName] = binding;
      }
    }
    message = { template, ...(bindings ? { bindings } : {}) };
  }
  return {
    name,
    verdict,
    evaluable,
    ...(args ? { args } : {}),
    ...(message ? { message } : {}),
    ...(typeof item['provisional'] === 'boolean' ? { provisional: item['provisional'] } : {}),
  };
}

function decodeMoveForm(value: unknown, path: string): MoveForm {
  const item = record(value, path);
  const Name = string(item['Name'], `${path}.Name`);
  const HelpText = string(item['HelpText'], `${path}.HelpText`, true);
  let Fields: MoveFormField[] | undefined;
  // Go encodes nil slices as null and populated slices as arrays. Both null
  // and omission mean "this move has no creator-authored fields".
  if (item['Fields'] !== undefined && item['Fields'] !== null) {
    Fields = array(item['Fields'], `${path}.Fields`, MAX_FORMS)
      .map((field, index) => decodeField(field, `${path}.Fields[${index}]`));
  }
  optionalBoolean(item['LegalForPlayer'], `${path}.LegalForPlayer`);
  optionalString(item['LegalForPlayerError'], `${path}.LegalForPlayerError`);
  optionalBoolean(item['LegalForAnyone'], `${path}.LegalForAnyone`);
  optionalBoolean(item['IsGatheringStart'], `${path}.IsGatheringStart`);
  let Preconditions: PreconditionEntry[] | undefined;
  if (item['Preconditions'] !== undefined && item['Preconditions'] !== null) {
    Preconditions = array(item['Preconditions'], `${path}.Preconditions`, MAX_FORMS)
      .map((entry, index) => decodePrecondition(entry, `${path}.Preconditions[${index}]`));
  }
  return {
    Name,
    HelpText,
    ...(Fields ? { Fields } : {}),
    ...(typeof item['LegalForPlayer'] === 'boolean' ? { LegalForPlayer: item['LegalForPlayer'] } : {}),
    ...(typeof item['LegalForPlayerError'] === 'string' ? { LegalForPlayerError: item['LegalForPlayerError'] } : {}),
    ...(typeof item['LegalForAnyone'] === 'boolean' ? { LegalForAnyone: item['LegalForAnyone'] } : {}),
    ...(typeof item['IsGatheringStart'] === 'boolean' ? { IsGatheringStart: item['IsGatheringStart'] } : {}),
    ...(Preconditions ? { Preconditions } : {}),
  };
}

function decodeForms(value: unknown, path: string): MoveForm[] | null {
  if (value === null || value === undefined) return null;
  return array(value, path, MAX_FORMS).map((form, index) => decodeMoveForm(form, `${path}[${index}]`));
}

function decodeRawState(value: unknown, path: string): RawGameState {
  const item = record(value, path);
  const Version = item['Version'] === undefined
    ? undefined
    : integer(item['Version'], `${path}.Version`, true);
  const Game = jsonObject(item['Game'], `${path}.Game`);
  const Players = array(item['Players'], `${path}.Players`, MAX_PLAYERS)
    .map((player, index) => jsonObject(player, `${path}.Players[${index}]`));
  let Computed: RawGameState['Computed'];
  if (item['Computed'] !== undefined) {
    const rawComputed = record(item['Computed'], `${path}.Computed`);
    const Global = rawComputed['Global'] === undefined
      ? undefined
      : jsonObject(rawComputed['Global'], `${path}.Computed.Global`);
    const computedPlayers = rawComputed['Players'] === undefined
      ? undefined
      : array(rawComputed['Players'], `${path}.Computed.Players`, MAX_PLAYERS)
        .map((player, index) => jsonObject(player, `${path}.Computed.Players[${index}]`));
    Computed = {
      ...(Global === undefined ? {} : { Global }),
      ...(computedPlayers === undefined ? {} : { Players: computedPlayers }),
    };
  }
  let Components: RawGameState['Components'];
  if (item['Components'] !== undefined) {
    const rawComponents = record(item['Components'], `${path}.Components`);
    if (Object.keys(rawComponents).length > MAX_FORMS) {
      throw new Error(`${path}.Components exceeds the maximum of ${MAX_FORMS} properties`);
    }
    Components = Object.fromEntries(Object.entries(rawComponents).map(([deckName, entries]) => [
      deckName,
      array(entries, `${path}.Components.${deckName}`, MAX_FORMS)
        .map((entry, index) => jsonValue(entry, `${path}.Components.${deckName}[${index}]`)),
    ]));
  }
  return {
    ...(Version === undefined ? {} : { Version }),
    Game,
    Players,
    ...(Computed === undefined ? {} : { Computed }),
    ...(Components === undefined ? {} : { Components }),
  } as RawGameState;
}

function decodeTimer(value: unknown, path: string): TimerInfo {
  const item = record(value, path);
  if (typeof item['TimeLeft'] !== 'number' || !Number.isFinite(item['TimeLeft']) || item['TimeLeft'] < 0) {
    throw new Error(`${path}.TimeLeft must be a non-negative finite number`);
  }
  if (item['originalTimeLeft'] !== undefined
    && (typeof item['originalTimeLeft'] !== 'number' || !Number.isFinite(item['originalTimeLeft']))) {
    throw new Error(`${path}.originalTimeLeft must be finite when present`);
  }
  optionalString(item['ID'], `${path}.ID`);
  return {
    TimeLeft: item['TimeLeft'],
    ...(typeof item['originalTimeLeft'] === 'number'
      ? { originalTimeLeft: item['originalTimeLeft'] }
      : {}),
    ...(typeof item['ID'] === 'string' ? { ID: item['ID'] } : {}),
  } as TimerInfo;
}

function decodeStringArray(value: unknown, path: string, maximum: number): string[] {
  return array(value, path, maximum)
    .map((entry, index) => string(entry, `${path}[${index}]`, true));
}

function decodeGame(value: unknown, path: string): GameFromServer {
  const item = record(value, path);
  const CurrentState = decodeRawState(item['CurrentState'], `${path}.CurrentState`);
  const timers = record(item['ActiveTimers'], `${path}.ActiveTimers`);
  const ActiveTimers = Object.fromEntries(Object.entries(timers).map(([id, timer]) => [
    id,
    decodeTimer(timer, `${path}.ActiveTimers.${id}`),
  ]));
  const Version = integer(item['Version'], `${path}.Version`, true);
  const CurrentPlayerIndex = integer(item['CurrentPlayerIndex'], `${path}.CurrentPlayerIndex`);
  const Finished = boolean(item['Finished'], `${path}.Finished`);
  const Winners = item['Winners'] === null || item['Winners'] === undefined
    ? []
    : array(item['Winners'], `${path}.Winners`, MAX_PLAYERS)
      .map((winner, index) => integer(winner, `${path}.Winners[${index}]`, true));
  optionalString(item['Diagram'], `${path}.Diagram`);
  const Name = string(item['Name'], `${path}.Name`);
  const ID = string(item['ID'], `${path}.ID`);
  const NumPlayers = integer(item['NumPlayers'], `${path}.NumPlayers`, true);
  if (NumPlayers > MAX_PLAYERS) throw new Error(`${path}.NumPlayers exceeds the maximum of ${MAX_PLAYERS}`);
  const Agents = decodeStringArray(item['Agents'], `${path}.Agents`, MAX_PLAYERS);
  const Variant = item['Variant'] === null || item['Variant'] === undefined
    ? {}
    : decodeStringMap(item['Variant'], `${path}.Variant`);
  return {
    Name,
    ID,
    NumPlayers,
    Agents,
    Variant,
    CurrentState,
    ActiveTimers,
    Version,
    CurrentPlayerIndex,
    Finished,
    Winners,
    ...(typeof item['Diagram'] === 'string' ? { Diagram: item['Diagram'] } : {}),
  } as GameFromServer;
}

function decodeBundle(value: unknown, path: string): ServerStateBundle {
  const item = record(value, path);
  const Game = decodeGame(item['Game'], `${path}.Game`);
  const Forms = decodeForms(item['Forms'], `${path}.Forms`);
  const ViewingAsPlayer = integer(item['ViewingAsPlayer'], `${path}.ViewingAsPlayer`);
  return { Game, Forms, ViewingAsPlayer, Move: item['Move'] };
}

export function decodeGameInfoResponse(value: unknown): GameInfoResponse {
  const item = record(value, 'Game info response');
  if (item['Status'] !== 'Success') throw new Error('Game info response.Status must be "Success"');
  const Chest = decodeChest(item['Chest'], 'Game info response.Chest');
  const Players = array(item['Players'], 'Game info response.Players', MAX_PLAYERS)
    .map((player, index) => decodePlayer(player, `Game info response.Players[${index}]`));
  const HasEmptySlots = boolean(item['HasEmptySlots'], 'Game info response.HasEmptySlots');
  const GameOpen = boolean(item['GameOpen'], 'Game info response.GameOpen');
  const GameVisible = boolean(item['GameVisible'], 'Game info response.GameVisible');
  const IsOwner = boolean(item['IsOwner'], 'Game info response.IsOwner');
  const Game = decodeGame(item['Game'], 'Game info response.Game');
  const Forms = decodeForms(item['Forms'], 'Game info response.Forms');
  const ViewingAsPlayer = integer(item['ViewingAsPlayer'], 'Game info response.ViewingAsPlayer');
  const StateVersion = integer(item['StateVersion'], 'Game info response.StateVersion', true);
  const LegalCatalogVersion = integer(item['LegalCatalogVersion'], 'Game info response.LegalCatalogVersion', true);
  const MoveInputSchemaFingerprint = string(
    item['MoveInputSchemaFingerprint'],
    'Game info response.MoveInputSchemaFingerprint',
  );
  const CompanionInfo = decodeCompanion(item['CompanionInfo'], 'Game info response.CompanionInfo');
  return {
    Status: 'Success', Chest, Players, HasEmptySlots, GameOpen, GameVisible,
    IsOwner, Game, Forms, ViewingAsPlayer, StateVersion, LegalCatalogVersion,
    MoveInputSchemaFingerprint, CompanionInfo,
  };
}

export function decodeGameVersionResponse(value: unknown): GameVersionResponse {
  const item = record(value, 'Game version response');
  if (item['Status'] !== 'Success') throw new Error('Game version response.Status must be "Success"');
  const Bundles = array(item['Bundles'], 'Game version response.Bundles', MAX_BUNDLES)
    .map((bundle, index) => decodeBundle(bundle, `Game version response.Bundles[${index}]`));
  optionalString(item['Error'], 'Game version response.Error');
  return {
    Status: 'Success',
    Bundles,
    ...(typeof item['Error'] === 'string' ? { Error: item['Error'] } : {}),
  };
}
