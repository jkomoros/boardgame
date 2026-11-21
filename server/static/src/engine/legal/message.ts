// Turning a verdict into text: Go's RenderLegalMessage, LegalError.Error()
// and LegalTemplatePlaceholders.
//
// The renderer itself is NOT reimplemented here. It already exists as
// src/legal/renderLegalMessage.ts, the client's byte-for-byte mirror of Go's
// RenderLegalMessage, shared with the ledger UI. The dependency runs one
// way: engine/legal imports src/legal. Nothing in src/legal imports the
// engine.

import { type LegalTemplateTable, renderLegalMessage } from '../../legal/renderLegalMessage.ts';
import type { LegalVerdict } from './types.ts';

export { renderLegalMessage };
export type { LegalTemplateTable };

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
