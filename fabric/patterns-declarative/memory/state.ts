/**
 * The part of Memory's released state the two experiments read by name. The
 * table releases its state as plain JSON (`TableView`'s default), so a
 * pattern that reaches into it says what it expects to find.
 */
import type { Json, JsonObject } from "../../patterns/toolkit/json.ts";

export interface MemoryState extends JsonObject {
  Game: { Cards: Json };
  Players: { WonCards: Json }[];
}
