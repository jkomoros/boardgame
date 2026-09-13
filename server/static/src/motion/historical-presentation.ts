export type HistoricalPresentationPolicy =
  | 'none'
  /** Master-compatible clone: preserves authored IDs and reference attributes. */
  | 'clone-default-slot'
  /** Safer clone for new components that do not depend on document identity. */
  | 'clone-default-slot-safe';

/** Explicit, component-owned display facts; never an item or arbitrary host state. */
export type HistoricalAppearance = Readonly<Record<string, string | number | boolean>>;

export interface HistoricalPresentationSource extends HTMLElement {
  readonly historicalPresentationPolicy?: HistoricalPresentationPolicy;
  readonly historicalPresentationSlots?: Readonly<Record<string, string>> | null;
  captureHistoricalAppearance?(): HistoricalAppearance | null;
  installHistoricalAppearance?(appearance: HistoricalAppearance): () => void;
}

export interface HistoricalPresentation {
  readonly kind: 'cloned-default-slot';
  readonly identity: 'preserve' | 'strip';
}

const presentations = new WeakMap<HistoricalPresentation, Readonly<{
  sourceTagName: string;
  nodes: readonly Readonly<{ node: Node; slot: string }>[];
  appearance: HistoricalAppearance | null;
}>>();
const installations = new WeakMap<HTMLElement, () => void>();

/** Leaf primitives may copy their resolved display values, never their data source. */
function clonePresentationTree(source: Node): Node {
  const clone = source.cloneNode(true);
  const visit = (from: Node, to: Node): void => {
    if (from instanceof Element && to instanceof Element) {
      (from as Element & { copyHistoricalPresentationTo?(target: Element): void })
        .copyHistoricalPresentationTo?.(to);
    }
    [...from.childNodes].forEach((child, index) => {
      const target = to.childNodes[index];
      if (target) visit(child, target);
    });
  };
  visit(source, clone);
  return clone;
}

function stripDocumentIdentity(node: Node): void {
  if (!(node instanceof Element)) return;
  for (const element of [node, ...node.querySelectorAll('[id], [autofocus], [tabindex]')]) {
    element.removeAttribute('id');
    element.removeAttribute('autofocus');
    element.removeAttribute('tabindex');
  }
}

/** Capture only component-approved, already-rendered public presentation. */
export function captureHistoricalPresentation(
  source: HistoricalPresentationSource,
): HistoricalPresentation | null {
  const policy = source.historicalPresentationPolicy ?? 'none';
  if (policy === 'none') return null;
  const safe = policy === 'clone-default-slot-safe';
  const slots = safe ? source.historicalPresentationSlots : null;
  const nodes: Array<Readonly<{ node: Node; slot: string }>> = [];
  // Preserve the legacy element-only contract; bare text/comments are not art.
  for (const child of source.children) {
    const sourceSlot = child.getAttribute('slot') ?? '';
    const slot = slots
      ? (Object.prototype.hasOwnProperty.call(slots, sourceSlot) ? slots[sourceSlot] : undefined)
      : !sourceSlot ? safe ? 'motion-history' : 'fallback' : undefined;
    if (!slot || child.localName === 'dom-bind') continue;
    const node = safe ? clonePresentationTree(child) : child.cloneNode(true);
    if (safe) stripDocumentIdentity(node);
    nodes.push(Object.freeze({ node, slot }));
  }
  const appearance = safe ? source.captureHistoricalAppearance?.() ?? null : null;
  if (nodes.length === 0 && !appearance) return null;
  const presentation = Object.freeze({
    kind: 'cloned-default-slot' as const,
    identity: safe ? 'strip' as const : 'preserve' as const,
  });
  presentations.set(presentation, Object.freeze({
    sourceTagName: source.localName,
    nodes: Object.freeze(nodes),
    appearance: appearance ? Object.freeze({ ...appearance }) : null,
  }));
  return presentation;
}

/** Clear only nodes and overrides installed by this module, never authored slots. */
export function clearHistoricalPresentation(target: HTMLElement): void {
  installations.get(target)?.();
}

/** A generation-safe disposer cannot clear a newer installation on the same host. */
export function historicalPresentationDisposer(target: HTMLElement): () => void {
  return installations.get(target) ?? (() => {});
}

/** Install detached clones without replacing the live component's authored content. */
export function installHistoricalPresentation(
  target: HistoricalPresentationSource,
  presentation: HistoricalPresentation,
): boolean {
  const captured = presentations.get(presentation);
  if (!captured) return false;
  const safe = presentation.identity === 'strip';
  if (safe && target.localName !== captured.sourceTagName) return false;
  const nodes: Node[] = [];
  let restore: (() => void) | undefined;
  try {
    // Stage first: a failing custom leaf must not leave a half-installed face.
    for (const entry of captured.nodes) {
      const node = safe ? clonePresentationTree(entry.node) : entry.node.cloneNode(true);
      if (safe) {
        stripDocumentIdentity(node);
        if (node instanceof Element) {
          node.setAttribute('inert', '');
          node.setAttribute('aria-hidden', 'true');
        }
      }
      if (node instanceof Element) node.setAttribute('slot', entry.slot);
      nodes.push(node);
    }
    clearHistoricalPresentation(target);
    if (!safe) {
      // This public slot replacement is specifically the legacy contract.
      for (const child of [...target.children]) {
        if (['fallback', 'motion-history'].includes(child.getAttribute('slot') ?? '')) child.remove();
      }
    }
    if (captured.appearance) restore = target.installHistoricalAppearance?.(captured.appearance);
    target.append(...nodes);
    let disposed = false;
    const dispose = () => {
      if (disposed) return;
      disposed = true;
      nodes.forEach(node => node.parentNode?.removeChild(node));
      restore?.();
      if (installations.get(target) === dispose) installations.delete(target);
    };
    installations.set(target, dispose);
    return true;
  } catch {
    nodes.forEach(node => node.parentNode?.removeChild(node));
    restore?.();
    return false;
  }
}
