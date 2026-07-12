// renderLegalMessage is the TypeScript mirror of Go's RenderLegalMessage
// (legal_error.go). It renders a declarative-legality message
// ({template, bindings}) against the game's merged LegalTemplates table.
//
// It is the SHARED client renderer for two surfaces (design spec §6,
// Sub-project B / Workstream 6): the pre-submit Preconditions ledger and the
// post-submit ProposeMove rejection envelope. It MUST stay byte-for-byte
// identical to the Go renderer — any divergence is a client bug, pinned by
// renderLegalMessage.test.ts.
//
// Semantics (verbatim from Go RenderLegalMessage):
//   - a nil/undefined message renders "".
//   - the body is table[template], falling back to the RAW template key
//     string when the key is absent (comma-ok, not truthiness: a present but
//     empty body stays "").
//   - each {name} placeholder (name matching /[A-Za-z0-9_]+/) is replaced by
//     its binding value; a MISSING binding renders the bare placeholder NAME
//     (never blank, never throwing). Bindings are absent entirely when the
//     owning entry's `evaluable` is false (#693 guard) — then every
//     placeholder renders bare.
//   - binding values are the closed scalar union string | number | boolean
//     (LegalBindingValue): a string renders verbatim, an integer via its
//     decimal form (Go strconv.Itoa — the wire only ever ships integers), a
//     boolean as "true"/"false" (Go strconv.FormatBool).
import type { PreconditionMessage } from '../types/api';

// Mirrors legalPlaceholderPattern in legal_error.go: `\{([A-Za-z0-9_]+)\}`.
// The global flag makes String.prototype.replace substitute every match, like
// Go's ReplaceAllStringFunc.
const LEGAL_PLACEHOLDER = /\{([A-Za-z0-9_]+)\}/g;

/** The client-side merged template table (chest LegalTemplates: key -> body). */
export type LegalTemplateTable = Record<string, string>;

export function renderLegalMessage(
  message: PreconditionMessage | null | undefined,
  table: LegalTemplateTable | undefined,
): string {
  if (!message) {
    return '';
  }
  // table[template] with comma-ok fallback to the raw key (never coerce a
  // present-but-empty body to the key via `??`/truthiness).
  const body =
    table && Object.prototype.hasOwnProperty.call(table, message.template)
      ? table[message.template]
      : message.template;
  const bindings = message.bindings;
  return body.replace(LEGAL_PLACEHOLDER, (_match: string, name: string): string => {
    if (!bindings || !Object.prototype.hasOwnProperty.call(bindings, name)) {
      return name; // missing binding -> bare placeholder name
    }
    const val = bindings[name];
    switch (typeof val) {
      case 'string':
        return val;
      case 'number':
        return String(val);
      case 'boolean':
        return String(val);
      default:
        return name;
    }
  });
}
