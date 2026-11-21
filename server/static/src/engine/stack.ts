// Stacks: where component instances live in a state. Ports stack.go.
//
// A physical stack is a JSON object in exactly the shape Go serializes
// (stackJSONObj). Stack is a handle over that object: it never copies it, so
// the state stays plain data. Growable and sized stacks differ as in Go: a
// growable stack's Indexes is a dense list; a sized stack's is a fixed run of
// slots where -1 (EMPTY_SLOT) means empty.
//
// Errors are thrown as StackError with Go's message text, and are raised
// before anything is mutated, as in Go.

import { EMPTY_SLOT, GENERIC_COMPONENT } from './ids.ts';
import type { EngineState } from './state.ts';

/** stackJSONObj in stack.go. Size and MaxSize are omitted when zero. */
export interface StackJSON {
	Deck: string;
	Indexes: number[];
	/**
	 * One id per slot. Always an array on a physical stack. Go serializes
	 * null here for a concatenated merged stack with no slots at all (see
	 * MergedStackView.toJSON).
	 */
	IDs: string[] | null;
	IDsLastSeen: Record<string, number>;
	ShuffleCount: number;
	Size?: number;
	MaxSize?: number;
}

export class StackError extends Error {
	constructor(message: string) {
		super(message);
		this.name = 'StackError';
	}
}

/** What Go reports for SlotsRemaining on an uncapped stack (math.MaxInt64). */
export const UNLIMITED_SLOTS = Infinity;

export type StackKind = 'growable' | 'sized';

/** The read-only surface shared by physical stacks and merged views. */
export interface StackView {
	readonly deckName: string;
	/** True when every slot is fixed (ImmutableSizedStack() != nil). */
	isSized(): boolean;
	len(): number;
	numComponents(): number;
	/** Deck index per slot: a component, EMPTY_SLOT, or GENERIC_COMPONENT. */
	indexes(): number[];
	ids(): string[];
	idsLastSeen(): Record<string, number>;
	shuffleCount(): number;
	maxSize(): number;
	slotsRemaining(): number;
	/** Deck index at a slot, or null where Go's ComponentAt returns nil. */
	componentAt(index: number): number | null;
	/** The JSON Go's MarshalJSON produces. */
	toJSON(): StackJSON;
}

function setOrDelete(node: StackJSON, key: 'Size' | 'MaxSize', value: number): void {
	if (value === 0) delete node[key];
	else node[key] = value;
}

/** A physical stack: a handle over one StackJSON node inside a state. */
export class Stack implements StackView {
	readonly state: EngineState;
	readonly node: StackJSON;
	readonly kind: StackKind;
	/** The owner path stack_ownership.go gives this stack, e.g. "Game.Draw". */
	readonly path: string;
	/** True for a board space, which cannot be resized. */
	readonly inBoard: boolean;

	constructor(state: EngineState, node: StackJSON, kind: StackKind, path: string, inBoard: boolean) {
		this.state = state;
		this.node = node;
		this.kind = kind;
		this.path = path;
		this.inBoard = inBoard;
	}

	get deckName(): string {
		return this.node.Deck;
	}

	isSized(): boolean {
		return this.kind === 'sized';
	}

	len(): number {
		return this.node.Indexes.length;
	}

	indexes(): number[] {
		return this.node.Indexes.slice();
	}

	numComponents(): number {
		if (this.kind === 'growable') return this.node.Indexes.length;
		let count = 0;
		for (const index of this.node.Indexes) if (index !== EMPTY_SLOT) count++;
		return count;
	}

	shuffleCount(): number {
		return this.node.ShuffleCount;
	}

	ids(): string[] {
		return (this.node.IDs ?? []).slice();
	}

	idsLastSeen(): Record<string, number> {
		return { ...this.node.IDsLastSeen };
	}

	maxSize(): number {
		if (this.kind === 'sized') return this.len();
		return this.node.MaxSize ?? 0;
	}

	slotsRemaining(): number {
		if (this.kind === 'growable') {
			const max = this.node.MaxSize ?? 0;
			if (max <= 0) return UNLIMITED_SLOTS;
			return max - this.len();
		}
		let count = 0;
		for (const index of this.node.Indexes) if (index === EMPTY_SLOT) count++;
		return count;
	}

	/**
	 * The deck index at a slot, or null where Go's ComponentAt returns nil:
	 * out of range, an empty slot, or an index past the end of the deck.
	 * A sanitized stack yields GENERIC_COMPONENT.
	 */
	componentAt(index: number): number | null {
		if (index >= this.len() || index < 0) return null;
		const deckIndex = this.node.Indexes[index];
		const deck = this.state.def.Decks[this.node.Deck];
		if (!deck) return null;
		if (deckIndex >= deck.Size) return null;
		if (deckIndex >= 0) return deckIndex;
		if (deckIndex === EMPTY_SLOT) return null;
		return GENERIC_COMPONENT;
	}

	firstComponentIndex(): number {
		if (this.kind === 'growable') return 0;
		for (let i = 0; i < this.node.Indexes.length; i++) if (this.node.Indexes[i] !== EMPTY_SLOT) return i;
		return -1;
	}

	lastComponentIndex(): number {
		if (this.kind === 'growable') return this.len() - 1;
		for (let i = this.node.Indexes.length - 1; i >= 0; i--) if (this.node.Indexes[i] !== EMPTY_SLOT) return i;
		return -1;
	}

	/** The first slot a component could be moved into, or -1. */
	firstSlot(): number {
		if (this.kind === 'growable') return 0;
		for (let i = 0; i < this.node.Indexes.length; i++) if (this.node.Indexes[i] === EMPTY_SLOT) return i;
		return -1;
	}

	lastSlot(): number {
		if (this.kind === 'growable') return this.len();
		for (let i = this.node.Indexes.length - 1; i >= 0; i--) if (this.node.Indexes[i] === EMPTY_SLOT) return i;
		return -1;
	}

	nextSlot(): number {
		return this.kind === 'growable' ? this.lastSlot() : this.firstSlot();
	}

	legalSlot(index: number): boolean {
		if (index < 0) return false;
		if (this.kind === 'growable') return index <= this.len();
		if (index >= this.len()) return false;
		return this.componentAt(index) === null;
	}

	resizable(): boolean {
		return !this.inBoard;
	}

	toJSON(): StackJSON {
		const out: StackJSON = {
			Deck: this.node.Deck,
			Indexes: this.node.Indexes.slice(),
			IDs: (this.node.IDs ?? []).slice(),
			IDsLastSeen: { ...this.node.IDsLastSeen },
			ShuffleCount: this.node.ShuffleCount,
		};
		if (this.node.Size) out.Size = this.node.Size;
		if (this.node.MaxSize) out.MaxSize = this.node.MaxSize;
		return out;
	}

	// ----- internals shared with EngineState ------------------------------

	/** Recomputes IDs from Indexes and the state's secret-move counts. */
	syncIds(): void {
		const ids = new Array<string>(this.node.Indexes.length);
		for (let i = 0; i < ids.length; i++) {
			const deckIndex = this.componentAt(i);
			ids[i] = deckIndex === null ? '' : this.state.componentId(this.node.Deck, deckIndex);
		}
		this.node.IDs = ids;
	}

	private idSeen(id: string): void {
		if (id === '') return;
		this.node.IDsLastSeen[id] = this.state.version;
	}

	/** modificationsAllowed. */
	private modificationsAllowedError(): string | null {
		if (this.state.sanitized) return "Modifications not allowed: stack's state is sanitized";
		const attachment = this.state.stackAttachmentError(this);
		if (attachment !== null) return 'Modifications not allowed: ' + attachment;
		return null;
	}

	private requireModifiable(): void {
		const err = this.modificationsAllowedError();
		if (err !== null) throw new StackError(err);
	}

	/** scrambleIds: see every component, bump its count, see it again. */
	private scrambleIds(): void {
		for (let i = 0; i < this.node.Indexes.length; i++) {
			const deckIndex = this.componentAt(i);
			if (deckIndex === null) continue;
			this.idSeen(this.state.componentId(this.node.Deck, deckIndex));
			this.state.movedSecretly(this.node.Deck, deckIndex);
			this.idSeen(this.state.componentId(this.node.Deck, deckIndex));
		}
	}

	private removeComponentAt(componentIndex: number): number {
		const deckIndex = this.node.Indexes[componentIndex];
		if (this.kind === 'growable') {
			this.node.Indexes.splice(componentIndex, 1);
			this.state.updateIndexForAllComponents(this);
		} else {
			// As in Go, a sized stack leaves the removed component's entry
			// in the state's component index; the insert that follows
			// replaces it.
			this.node.Indexes[componentIndex] = EMPTY_SLOT;
		}
		return deckIndex;
	}

	private insertComponentAt(slotIndex: number, deckIndex: number): void {
		if (this.kind === 'growable') this.node.Indexes.splice(slotIndex, 0, deckIndex);
		else this.node.Indexes[slotIndex] = deckIndex;
		this.idSeen(this.state.componentId(this.node.Deck, deckIndex));
		if (this.kind === 'growable') this.state.updateIndexForAllComponents(this);
		else this.state.componentAdded(this.node.Deck, deckIndex, this, slotIndex);
	}

	/**
	 * componentInstance.MayMoveTo for the component deckIndex of this
	 * stack's deck. Like Go, it finds the component's source through the
	 * state's component index rather than assuming it is in this stack.
	 */
	private mayMoveToError(deckIndex: number, dest: Stack): string | null {
		if (dest.state !== this.state) return 'source and destination belong to different states';
		const destAttachment = this.state.stackAttachmentError(dest);
		if (destAttachment !== null) return 'destination stack is not attached: ' + destAttachment;
		const found = this.state.locateComponent(this.node.Deck, deckIndex);
		if (typeof found === 'string') return 'component is not in any stack: ' + found;
		const source = found.stack;
		if (source === dest) return 'source and destination are the same stack';
		const sourceAttachment = this.state.stackAttachmentError(source);
		if (sourceAttachment !== null) return 'source stack is not attached: ' + sourceAttachment;
		if (dest.node.Deck !== source.node.Deck) return 'source and destination use different decks';
		if (dest.slotsRemaining() < 1) return 'destination stack has no slots remaining';
		return null;
	}

	/**
	 * MayMoveTo for the component at componentIndex: the slot-independent
	 * checks MoveTo performs. Returns null when the move would be allowed,
	 * otherwise Go's error text.
	 */
	mayMoveComponentTo(componentIndex: number, dest: Stack): string | null {
		const deckIndex = this.componentAt(componentIndex);
		if (deckIndex === null) return 'no component at index ' + componentIndex;
		return this.mayMoveToError(deckIndex, dest);
	}

	/** moveComonentImpl's checks, without the mutation. */
	private moveComponentError(componentIndex: number, dest: Stack, slotIndex: number): string | null {
		if (this.state !== dest.state) return 'Source and destination belong to different states';
		const sourceErr = this.modificationsAllowedError();
		if (sourceErr !== null) return "Source doesn't allow modifications: " + sourceErr;
		const destErr = dest.modificationsAllowedError();
		if (destErr !== null) return "Destination doesn't allow modifications: " + destErr;
		if (this === dest) return 'source and destination stack are the same. Use SwapComponents instead';
		const deckIndex = this.componentAt(componentIndex);
		if (deckIndex === null) {
			return 'The effective index, ' + componentIndex + ' does not point to an existing component in Source';
		}
		const may = this.mayMoveToError(deckIndex, dest);
		if (may !== null) return may;
		if (!dest.legalSlot(slotIndex)) {
			return 'The effective slot index, ' + slotIndex + ' does not point to a legal slot.';
		}
		return null;
	}

	// ----- moves ----------------------------------------------------------

	/**
	 * ComponentInstance.MoveTo: moves the component at componentIndex to
	 * slotIndex of another stack. In a growable destination a slot is any
	 * index from 0 through len; in a sized one it must be an empty slot.
	 */
	moveComponent(componentIndex: number, dest: Stack, slotIndex: number): void {
		if (slotIndex < 0) throw new StackError('Invalid slotIndex');
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentImpl(source.slotIndex, dest, slotIndex);
	}

	/**
	 * Go's moves are methods on the component (`stack.ComponentAt(i).MoveTo`),
	 * and the component finds its own source stack through the state's
	 * component index. This does the same lookup, so the rare case where
	 * Go's index is stale fails here exactly as it does there.
	 */
	private sourceOf(componentIndex: number): { stack: Stack; slotIndex: number } {
		const deckIndex = this.componentAt(componentIndex);
		if (deckIndex === null) throw new StackError('no component at index ' + componentIndex);
		const found = this.state.locateComponent(this.node.Deck, deckIndex);
		if (typeof found === 'string') {
			throw new StackError('The source component was not in a mutable stack: ' + found);
		}
		return found;
	}

	private moveComponentImpl(componentIndex: number, dest: Stack, slotIndex: number): void {
		const err = this.moveComponentError(componentIndex, dest, slotIndex);
		if (err !== null) throw new StackError(err);
		const moved = this.removeComponentAt(componentIndex);
		dest.insertComponentAt(slotIndex, moved);
		this.syncIds();
		dest.syncIds();
	}

	/**
	 * ComponentInstance.SecretMoveTo: MoveTo, then scramble the id of every
	 * component in the destination so the insertion point cannot be
	 * observed.
	 */
	secretMoveComponent(componentIndex: number, dest: Stack, slotIndex: number): void {
		if (slotIndex < 0) throw new StackError('Invalid slotIndex');
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentImpl(source.slotIndex, dest, slotIndex);
		dest.scrambleIds();
		dest.syncIds();
	}

	moveComponentToFirstSlot(componentIndex: number, dest: Stack): void {
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentImpl(source.slotIndex, dest, dest.firstSlot());
	}

	moveComponentToLastSlot(componentIndex: number, dest: Stack): void {
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentImpl(source.slotIndex, dest, dest.lastSlot());
	}

	moveComponentToNextSlot(componentIndex: number, dest: Stack): void {
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentImpl(source.slotIndex, dest, dest.nextSlot());
	}

	/** ComponentInstance.SlideToFirstSlot. */
	slideToFirstSlot(componentIndex: number): void {
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentToExtreme(source.slotIndex, true);
	}

	/** ComponentInstance.SlideToLastSlot. */
	slideToLastSlot(componentIndex: number): void {
		const source = this.sourceOf(componentIndex);
		source.stack.moveComponentToExtreme(source.slotIndex, false);
	}

	private moveComponentToExtreme(componentIndex: number, isStart: boolean): void {
		this.requireModifiable();
		if (this.componentAt(componentIndex) === null) {
			throw new StackError('component index does not point to a component');
		}
		const moved = this.removeComponentAt(componentIndex);
		const targetSlot = isStart ? this.firstSlot() : this.lastSlot();
		this.insertComponentAt(targetSlot, moved);
		this.syncIds();
	}

	private transferError(dest: Stack, count: number, historicalCapacityError: string): string | null {
		// validateComponentTransferEndpoints
		if (this === dest) return 'source and destination are the same stack';
		if (this.state !== dest.state) return 'source and destination belong to different states';
		const sourceErr = this.modificationsAllowedError();
		if (sourceErr !== null) return "source doesn't allow modifications: " + sourceErr;
		const destErr = dest.modificationsAllowedError();
		if (destErr !== null) return "destination doesn't allow modifications: " + destErr;
		if (this.node.Deck !== dest.node.Deck) return 'source and destination use different decks';
		// preparePhysicalComponentTransfer
		if (count < 0) return 'component count must be non-negative';
		if (count > this.numComponents()) {
			return 'source stack has ' + this.numComponents() + ' components, cannot move ' + count;
		}
		const available = dest.slotsRemaining();
		if (available < count) {
			if (historicalCapacityError !== '') return historicalCapacityError;
			return (
				'destination stack has ' + available + ' ' + (available === 1 ? 'slot' : 'slots') +
				' remaining; cannot move ' + count + ' ' + (count === 1 ? 'component' : 'components')
			);
		}
		// validateLiveComponentTransferPlan: each component that will move
		// must pass MayMoveTo against the state as it is now.
		let validated = 0;
		for (let i = 0; i < this.node.Indexes.length && validated < count; i++) {
			const deckIndex = this.componentAt(i);
			if (deckIndex === null) continue;
			const may = this.mayMoveToError(deckIndex, dest);
			if (may !== null) return may;
			validated++;
		}
		return null;
	}

	private commitTransfer(dest: Stack, count: number): void {
		for (let remaining = count; remaining > 0; remaining--) {
			const moved = this.removeComponentAt(this.firstComponentIndex());
			dest.insertComponentAt(dest.nextSlot(), moved);
		}
		this.syncIds();
		dest.syncIds();
	}

	/** MayMoveAllTo: null when MoveAllTo would succeed, else Go's error. */
	mayMoveAllTo(dest: Stack): string | null {
		return this.transferError(dest, this.numComponents(), 'not enough space in the target stack');
	}

	/** MayMoveCountTo: null when MoveCountTo would succeed, else Go's error. */
	mayMoveCountTo(dest: Stack, count: number): string | null {
		return this.transferError(dest, count, '');
	}

	/** Stack.MoveAllTo: each component in order to the destination's next slot. */
	moveAllTo(dest: Stack): void {
		const count = this.numComponents();
		const err = this.transferError(dest, count, 'Not enough space in the target stack');
		if (err !== null) throw new StackError(err);
		this.commitTransfer(dest, count);
	}

	/** Stack.MoveCountTo: the first count components, as MoveAllTo. */
	moveCountTo(dest: Stack, count: number): void {
		const err = this.transferError(dest, count, '');
		if (err !== null) throw new StackError(err);
		this.commitTransfer(dest, count);
	}

	// ----- shuffles -------------------------------------------------------

	/**
	 * Stack.PublicShuffle: permutes the slots (empty ones included, in a
	 * sized stack) with the state's generator. Ids are left alone.
	 */
	publicShuffle(): void {
		this.requireModifiable();
		this.permute();
		this.syncIds();
	}

	/** Stack.Shuffle: PublicShuffle, then scramble every component's id. */
	shuffle(): void {
		this.requireModifiable();
		this.permute();
		this.scrambleIds();
		this.syncIds();
	}

	private permute(): void {
		const current = this.node.Indexes;
		const perm = this.state.rand().perm(current.length);
		const next = new Array<number>(current.length);
		for (let i = 0; i < perm.length; i++) next[i] = current[perm[i]];
		this.node.Indexes = next;
		this.state.updateIndexForAllComponents(this);
		this.node.ShuffleCount++;
	}

	/** MaySwapComponents: null when SwapComponents would succeed. */
	maySwapComponents(i: number, j: number): string | null {
		const err = this.modificationsAllowedError();
		if (err !== null) return err;
		if (i < 0) return 'i must be 0 or greater';
		if (j < 0) return 'j must be 0 or greater';
		if (i >= this.len()) return "i must be less than the stack's length";
		if (j >= this.len()) return "j must be less than the stack's length";
		if (i === j) return 'i and j were the same';
		return null;
	}

	/** Stack.SwapComponents: swaps two slots; either may be empty. */
	swapComponents(i: number, j: number): void {
		const err = this.maySwapComponents(i, j);
		if (err !== null) throw new StackError(err);
		const indexes = this.node.Indexes;
		const t = indexes[i];
		indexes[i] = indexes[j];
		indexes[j] = t;
		this.state.componentAdded(this.node.Deck, this.componentAt(i), this, i);
		this.state.componentAdded(this.node.Deck, this.componentAt(j), this, j);
		this.syncIds();
	}

	/**
	 * Stack.SortComponents. `less` receives deck indexes (EMPTY_SLOT for an
	 * empty slot). Go uses sort.Sort, which is not stable and whose exact
	 * permutation for equal elements is an implementation detail of Go's
	 * pdqsort; this port uses a stable sort, so `less` must be a strict total
	 * order over the stack's contents for results to match Go.
	 */
	sortComponents(less: (a: number, b: number) => boolean): void {
		this.requireModifiable();
		const sorted = this.node.Indexes.slice();
		sorted.sort((a, b) => (less(a, b) ? -1 : less(b, a) ? 1 : 0));
		this.node.Indexes = sorted;
		this.state.updateIndexForAllComponents(this);
		this.syncIds();
	}

	// ----- resizing -------------------------------------------------------

	expandSize(newSlots: number): void {
		this.requireModifiable();
		if (!this.resizable()) throw new StackError('stack is not resizable');
		if (this.kind === 'growable') {
			if (this.maxSize() === 0) {
				throw new StackError("Can't expand maxSize; maxSize is currently not set. Use SetSize first");
			}
			if (newSlots <= 0) {
				throw new StackError('Expand size must have a positive non-zero number for newSlots');
			}
			setOrDelete(this.node, 'MaxSize', this.maxSize() + newSlots);
			return;
		}
		if (newSlots < 1) throw new StackError("Can't add 0 or negative slots to a sized stack");
		for (let i = 0; i < newSlots; i++) this.node.Indexes.push(EMPTY_SLOT);
		setOrDelete(this.node, 'Size', (this.node.Size ?? 0) + newSlots);
		this.syncIds();
	}

	contractSize(newSize: number): void {
		this.requireModifiable();
		if (!this.resizable()) throw new StackError('stack is not resizable');
		if (this.kind === 'growable') {
			if (this.maxSize() === 0) {
				throw new StackError("Can't expand maxSize; maxSize is currently not set. Use SetSize first");
			}
			if (newSize <= 0) {
				throw new StackError('Contract size must be given a positive, non-zero number for newSize');
			}
			if (newSize > this.maxSize()) throw new StackError('Contract size must be passed a smaller size');
			if (newSize < this.numComponents()) {
				throw new StackError("Can't set the max size to a size smaller than the current number of components");
			}
			setOrDelete(this.node, 'MaxSize', newSize);
			return;
		}
		if (newSize > this.len()) throw new StackError('Contract size cannot be used to grow a sized stack');
		if (newSize < this.numComponents()) {
			throw new StackError(
				'The proposed newSize for stack would not be sufficient to contain the components currently in the stack',
			);
		}
		if (newSize < 0) newSize = 0;
		while (this.len() > newSize) {
			const slotIndex = this.lastSlot();
			if (slotIndex === -1) {
				throw new StackError('there was an unexpected error contracting size of stack: no more slots to remove');
			}
			// Go does not refresh the state's component index here
			// (stack.go, sizedStack.ContractSize), so components to the
			// right of a removed slot keep a stale entry. Reproduced
			// deliberately: see EngineState.locateComponent.
			this.node.Indexes.splice(slotIndex, 1);
		}
		setOrDelete(this.node, 'Size', newSize);
		this.syncIds();
	}

	setSize(newSize: number): void {
		this.requireModifiable();
		if (this.kind === 'growable') {
			if (this.maxSize() === newSize) return;
			if (newSize < 0) newSize = 0;
			if (this.maxSize() === 0) {
				setOrDelete(this.node, 'MaxSize', newSize);
				return;
			}
			if (this.maxSize() > newSize) {
				this.contractSize(newSize);
				return;
			}
			this.expandSize(newSize - this.maxSize());
			return;
		}
		if (this.len() === newSize) return;
		if (this.len() > newSize) {
			this.contractSize(newSize);
			return;
		}
		this.expandSize(newSize - this.len());
	}

	sizeToFit(): void {
		if (this.kind === 'growable') this.setSize(this.len());
		else this.contractSize(this.numComponents());
	}
}

/**
 * A merged stack: a read-only view over other stacks (mergedStack in
 * stack.go). A concatenation lays its parts end to end. An overlap shows, for
 * each slot, the first part that has a component there.
 */
export class MergedStackView implements StackView {
	readonly parts: readonly StackView[];
	readonly overlap: boolean;

	constructor(parts: readonly StackView[], overlap: boolean) {
		this.parts = parts;
		this.overlap = overlap;
	}

	get deckName(): string {
		return this.parts.length === 0 ? '' : this.parts[0].deckName;
	}

	isSized(): boolean {
		for (const part of this.parts) if (!part.isSized()) return false;
		return true;
	}

	len(): number {
		if (this.parts.length === 0) return 0;
		if (this.overlap) return this.parts[0].len();
		let total = 0;
		for (const part of this.parts) total += part.len();
		return total;
	}

	componentAt(index: number): number | null {
		if (this.parts.length === 0) return null;
		if (this.overlap) {
			for (const part of this.parts) {
				const c = part.componentAt(index);
				if (c !== null) return c;
			}
			return null;
		}
		for (const part of this.parts) {
			if (index < part.len()) return part.componentAt(index);
			index -= part.len();
		}
		return null;
	}

	indexes(): number[] {
		const length = this.len();
		const out = new Array<number>(length);
		for (let i = 0; i < length; i++) {
			const c = this.componentAt(i);
			out[i] = c === null ? EMPTY_SLOT : c;
		}
		return out;
	}

	numComponents(): number {
		if (this.parts.length === 0) return 0;
		if (!this.overlap) {
			let total = 0;
			for (const part of this.parts) total += part.numComponents();
			return total;
		}
		let count = 0;
		const length = this.parts[0].len();
		for (let i = 0; i < length; i++) if (this.componentAt(i) !== null) count++;
		return count;
	}

	ids(): string[] {
		if (this.parts.length === 0) return [];
		const cached = this.parts.map(part => part.ids());
		if (this.overlap) {
			const result = new Array<string>(cached[0].length);
			for (let i = 0; i < result.length; i++) {
				result[i] = '';
				if (cached[0][i] !== '') {
					result[i] = cached[0][i];
					continue;
				}
				for (let depth = 1; depth < cached.length; depth++) {
					const id = cached[depth][i];
					if (id !== undefined && id !== '') {
						result[i] = id;
						break;
					}
				}
			}
			return result;
		}
		const result: string[] = [];
		for (const ids of cached) for (const id of ids) result.push(id);
		return result;
	}

	/**
	 * The latest version each id was seen in any part. As in Go, an id last
	 * seen at version 0 never enters the result (`val > result[key]` with a
	 * zero default).
	 */
	idsLastSeen(): Record<string, number> {
		const result: Record<string, number> = {};
		for (const part of this.parts) {
			const seen = part.idsLastSeen();
			for (const key of Object.keys(seen)) {
				const previous = Object.prototype.hasOwnProperty.call(result, key) ? result[key] : 0;
				if (seen[key] > previous) result[key] = seen[key];
			}
		}
		return result;
	}

	shuffleCount(): number {
		let count = 0;
		for (const part of this.parts) count += part.shuffleCount();
		return count;
	}

	maxSize(): number {
		if (this.parts.length === 0) return 0;
		if (this.overlap) return this.parts[0].maxSize();
		let count = 0;
		for (const part of this.parts) count += part.maxSize();
		return count;
	}

	slotsRemaining(): number {
		if (this.parts.length === 0) return 0;
		if (this.overlap) {
			let count = 0;
			const length = this.len();
			for (let i = 0; i < length; i++) if (this.componentAt(i) === null) count++;
			return count;
		}
		let count = 0;
		for (const part of this.parts) {
			const remaining = part.slotsRemaining();
			if (remaining === UNLIMITED_SLOTS) return UNLIMITED_SLOTS;
			count += remaining;
		}
		return count;
	}

	/** mergedStack.Valid's structural rules; null when valid. */
	validityError(): string | null {
		if (this.parts.length === 0) return 'no sub-stacks provided';
		for (let i = 0; i < this.parts.length; i++) {
			const part = this.parts[i];
			if (part instanceof MergedStackView) {
				const nested = part.validityError();
				if (nested !== null) return 'stack ' + i + ' is an invalid merged stack: ' + nested;
			}
		}
		const deck = this.parts[0].deckName;
		for (let i = 0; i < this.parts.length; i++) {
			if (this.parts[i].deckName !== deck) return 'stack ' + i + ' had a different deck than other sub-stacks';
		}
		if (!this.overlap) return null;
		for (let i = 0; i < this.parts.length; i++) {
			if (!this.parts[i].isSized()) {
				return 'stack ' + i + ' was not fixed size, but overlap stacks require them all to be fixed size';
			}
		}
		const length = this.parts[0].len();
		for (let i = 0; i < this.parts.length; i++) {
			if (this.parts[i].len() !== length) return 'stack ' + i + ' was not the same length as the others';
		}
		return null;
	}

	toJSON(): StackJSON {
		const ids = this.ids();
		const out: StackJSON = {
			Deck: this.deckName,
			Indexes: this.indexes(),
			// mergedStack.IDs builds a concatenation by appending to a nil
			// slice (stack.go:1282), so with no slots at all Go marshals
			// null, not []. An overlap always allocates and marshals [].
			IDs: !this.overlap && ids.length === 0 ? null : ids,
			IDsLastSeen: this.idsLastSeen(),
			ShuffleCount: this.shuffleCount(),
		};
		if (this.isSized()) {
			const size = this.len();
			if (size !== 0) out.Size = size;
		} else {
			const max = this.maxSize();
			if (max !== 0) out.MaxSize = max;
		}
		return out;
	}
}
