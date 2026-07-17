const MAX_DOCUMENT_EFFECTS = 8;
const MAX_DOCUMENT_PARTICLES = 60;

export interface EffectReservation {
  readonly particles: number;
  release(): void;
}

interface BudgetState {
  effects: number;
  particles: number;
}

const documentBudgets = new WeakMap<Document, BudgetState>();

function stateFor(document: Document): BudgetState {
  let state = documentBudgets.get(document);
  if (!state) {
    state = { effects: 0, particles: 0 };
    documentBudgets.set(document, state);
  }
  return state;
}

/**
 * Atomically admits one finite visual effect. Particle-heavy requests degrade
 * to the remaining document budget; existing effects are never killed to make
 * room for newer decoration.
 */
export function reserveEffectBudget(document: Document, requestedParticles: number): EffectReservation | null {
  const state = stateFor(document);
  if (state.effects >= MAX_DOCUMENT_EFFECTS) return null;
  const safeRequested = Number.isFinite(requestedParticles)
    ? Math.max(1, Math.round(requestedParticles))
    : 1;
  const particles = Math.min(safeRequested, MAX_DOCUMENT_PARTICLES - state.particles);
  if (particles <= 0) return null;

  state.effects += 1;
  state.particles += particles;
  let released = false;
  return Object.freeze({
    particles,
    release(): void {
      if (released) return;
      released = true;
      state.effects = Math.max(0, state.effects - 1);
      state.particles = Math.max(0, state.particles - particles);
    },
  });
}

export function effectBudgetSnapshot(document: Document): Readonly<BudgetState> {
  const state = stateFor(document);
  return Object.freeze({ ...state });
}
