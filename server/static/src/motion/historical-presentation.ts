export type HistoricalPresentationPolicy =
  | 'none'
  /** Master-compatible clone: preserves authored IDs and reference attributes. */
  | 'clone-default-slot'
  /** Safer clone for new components that do not depend on document identity. */
  | 'clone-default-slot-safe';

export interface HistoricalPresentationSource extends HTMLElement {
  readonly historicalPresentationPolicy?: HistoricalPresentationPolicy;
}

export interface HistoricalPresentation {
  readonly kind: 'cloned-default-slot';
  readonly identity: 'preserve' | 'strip';
}

const presentations = new WeakMap<HistoricalPresentation, Readonly<{
  sourceTagName: string;
  nodes: readonly Node[];
}>>();

function stripDocumentIdentity(node: Node): void {
  if (!(node instanceof Element)) return;
  node.removeAttribute('id');
  node.removeAttribute('autofocus');
  node.removeAttribute('tabindex');
  for (const descendant of node.querySelectorAll('[id], [autofocus], [tabindex]')) {
    descendant.removeAttribute('id');
    descendant.removeAttribute('autofocus');
    descendant.removeAttribute('tabindex');
  }
}

/** Capture only already-rendered, unslotted light DOM. Never component state. */
export function captureHistoricalPresentation(
  source: HistoricalPresentationSource,
): HistoricalPresentation | null {
  const policy = source.historicalPresentationPolicy ?? 'none';
  if (policy === 'none') return null;
  const stripIdentity = policy === 'clone-default-slot-safe';
  const nodes: Node[] = [];
  for (const child of source.childNodes) {
    if (child instanceof HTMLElement && child.slot) continue;
    if (child instanceof Element && child.localName === 'dom-bind') continue;
    const clone = child.cloneNode(true);
    if (stripIdentity) stripDocumentIdentity(clone);
    nodes.push(clone);
  }
  if (nodes.length === 0) return null;
  const presentation = Object.freeze({
    kind: 'cloned-default-slot' as const,
    identity: stripIdentity ? 'strip' as const : 'preserve' as const,
  });
  presentations.set(presentation, Object.freeze({
    sourceTagName: source.localName,
    nodes: Object.freeze(nodes),
  }));
  return presentation;
}

/** Install detached clones into the framework-reserved historical slot. */
export function installHistoricalPresentation(
  target: HTMLElement,
  presentation: HistoricalPresentation,
): boolean {
  const capturedPresentation = presentations.get(presentation);
  if (!capturedPresentation) return false;
  if (presentation.identity === 'strip'
    && target.localName !== capturedPresentation.sourceTagName) return false;
  try {
    for (const existing of [...target.children]) {
      if ((existing as HTMLElement).slot === 'motion-history') existing.remove();
    }
    for (const captured of capturedPresentation.nodes) {
      const clone = captured.cloneNode(true);
      if (presentation.identity === 'strip') stripDocumentIdentity(clone);
      if (clone instanceof HTMLElement) clone.slot = 'motion-history';
      target.append(clone);
    }
    return true;
  } catch {
    return false;
  }
}
