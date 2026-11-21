/**
 * JSON, as a type a Common Fabric cell can hold.
 *
 * A cell's schema comes from its TypeScript type, and a position typed
 * `unknown` is a reference to another piece: whatever is stored there reads
 * back as an empty object (FINDINGS.md of the Memory pattern, F7). The
 * engine's own types say `unknown` wherever Go has an interface value, so a
 * table stores engine data under this type instead and converts at the
 * boundary.
 */
export type Json = null | boolean | number | string | Json[] | {
  [key: string]: Json;
};

export type JsonObject = { [key: string]: Json };

/**
 * A copy of `value` as JSON: every object and array in the result is new,
 * and anything JSON cannot hold (`undefined`, a function) is left out.
 *
 * The copy matters as much as the type. An object passed through from a
 * sealed cell into a released view reaches the store as a reference to the
 * sealed cell, and the whole view is then withheld (FINDINGS.md, F4).
 */
export function toJson(value: unknown): Json {
  if (value === null || value === undefined) return null;
  if (
    typeof value === "boolean" || typeof value === "number" ||
    typeof value === "string"
  ) return value;
  if (Array.isArray(value)) return value.map((item) => toJson(item));
  if (typeof value === "object") return toJsonObject(value);
  return null;
}

/** `toJson` for a value known to be a record. */
export function toJsonObject(value: object): JsonObject {
  const out: JsonObject = {};
  const record = value as Record<string, unknown>;
  for (const key of Object.keys(record)) {
    const item = record[key];
    if (item === undefined || typeof item === "function") continue;
    out[key] = toJson(item);
  }
  return out;
}
