// Turning a verdict into text: Go's RenderLegalMessage, LegalError.Error()
// and LegalTemplatePlaceholders.
//
// This is the only TypeScript renderer. The client had a copy of its own
// (src/legal/renderLegalMessage.ts); nothing in the client imported it, and
// it could not load in a Common Fabric pattern (it kept a global regular
// expression at module scope, which the pattern sandbox refuses as stateful
// data), so it was removed and its tests moved here as
// render-message.test.ts.

import type { LegalMessage, LegalVerdict } from './types.ts';

/** The merged template table (chest LegalTemplates: key -> body). */
export type LegalTemplateTable = Record<string, string>;

/**
 * Go: RenderLegalMessage. A nil message renders ''. The body is
 * `table[template]`, or the raw key when the table has none (a present but
 * empty body stays ''). Each `{name}` is replaced by its binding; a missing
 * binding renders the bare name.
 */
export function renderLegalMessage(
  message: LegalMessage | null | undefined,
  table: LegalTemplateTable | undefined,
): string {
  if (!message) {
    return '';
  }
  const body =
    table && Object.prototype.hasOwnProperty.call(table, message.template)
      ? table[message.template]
      : message.template;
  const bindings = message.bindings;
  // Mirrors legalPlaceholderPattern in legal_error.go. Built per call: a
  // global regular expression carries state, so it cannot live at module
  // scope in a pattern.
  return body.replace(/\{([A-Za-z0-9_]+)\}/g, (_match: string, name: string): string => {
    if (!bindings || !Object.prototype.hasOwnProperty.call(bindings, name)) {
      return name;
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

/**
 * Go: LegalTemplatePlaceholders. The distinct {name} placeholders a template
 * body references, in first-appearance order.
 */
export function legalTemplatePlaceholders(body: string): string[] {
  const out: string[] = [];
  for (const match of body.matchAll(/\{([A-Za-z0-9_]+)\}/g)) {
    if (!out.includes(match[1])) {
      out.push(match[1]);
    }
  }
  return out;
}

/**
 * Go: GameManager.LegalRenderVerdict, i.e. LegalVerdict.Error() rendered
 * against the game's template table. The text move.Legal() returns for this
 * verdict: '' for a pass; the rendered message when there is one (a template
 * key missing from the table renders as the key itself); otherwise the
 * unknown reason; otherwise Go's generic fallback.
 */
export function renderLegalVerdict(verdict: LegalVerdict, table: LegalTemplateTable | undefined): string {
  if (verdict.outcome === 'pass') {
    return '';
  }
  if (verdict.message) {
    return renderLegalMessage(verdict.message, table);
  }
  if (verdict.reason) {
    return verdict.reason;
  }
  return 'boardgame: move is not legal';
}

/**
 * Go: legal.Errorf. The structured failure a LegalCustom function returns to
 * keep its message a template key plus bindings.
 */
export function legalErrorf(templateKey: string, bindings?: Record<string, string | number | boolean>): LegalVerdict {
  return bindings === undefined
    ? { outcome: 'fail', message: { template: templateKey } }
    : { outcome: 'fail', message: { template: templateKey, bindings } };
}
