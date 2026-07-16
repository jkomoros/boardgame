import {
  createMoveAction,
  type MoveActionService,
  type MoveActionSnapshot,
  type MoveProposalResult,
} from './action.js';
import { bindMoveAction } from './action-binding.js';

type MoveName = 'Roll' | 'Place' | 'Choose' | 'Customize';
type MoveInputs = {
  Roll: Record<string, never>;
  Place: { Slot: number };
  Choose: { Choice: 'red' | 'blue'; Confirm?: boolean };
  Customize: { Label?: string };
};

declare const service: MoveActionService;
declare const snapshot: MoveActionSnapshot;
const roll = createMoveAction<'Roll', MoveName, MoveInputs>('Roll', service, snapshot);
void roll.propose();
const directHandler: () => Promise<MoveProposalResult> = roll.propose;
void directHandler;
void bindMoveAction(roll);

const place = createMoveAction<'Place', MoveName, MoveInputs>('Place', service, snapshot);
void place.with({ Slot: 3 }).propose();
const targets = place.targets([0, 1, 2] as const, Slot => ({ Slot }));
void targets.get(1)?.action.activate();
// @ts-expect-error A target key outside the inferred literal set is rejected.
void targets.get(4);
void bindMoveAction(place.with({ Slot: 3 }));
const choose = createMoveAction<'Choose', MoveName, MoveInputs>('Choose', service, snapshot);
void choose.with({ Choice: 'red' }).propose();
void choose.with({ Choice: 'blue', Confirm: true }).propose();
const customize = createMoveAction<'Customize', MoveName, MoveInputs>('Customize', service, snapshot);
void customize.with({}).propose();
void customize.with({ Label: 'custom' }).propose();
// @ts-expect-error Defaultable creator fields still require explicit binding.
void customize.propose();

// @ts-expect-error Required-input actions cannot be proposed before binding arguments.
void place.propose();
// @ts-expect-error Required-input builders cannot be bound to an interactive element.
void bindMoveAction(place);
// @ts-expect-error Required fields cannot be omitted.
void place.with({});
// @ts-expect-error Extra fields are rejected even on fresh literals.
void place.with({ Slot: 3, Other: 1 });
// @ts-expect-error Native integer inputs cannot be supplied as wire strings.
void place.with({ Slot: '3' });
// @ts-expect-error Target mappers use native generated inputs.
void place.targets([0, 1], Slot => ({ Slot: String(Slot) }));
// @ts-expect-error Target mappers reject extra input fields.
void place.targets([0, 1], Slot => ({ Slot, Other: true }));
// @ts-expect-error Bound actions cannot be rebound as a target collection.
void place.with({ Slot: 1 }).targets([0, 1], Slot => ({ Slot }));
// @ts-expect-error Enum inputs accept only generated values.
void choose.with({ Choice: 'green' });
// @ts-expect-error Arbitrary strings are not generated move names.
void createMoveAction<'Nope', MoveName, MoveInputs>('Nope', service, snapshot);

function exhaust(result: MoveProposalResult): string {
  switch (result.kind) {
    case 'success':
    case 'server-rejection':
    case 'network-failure':
    case 'blocked':
    case 'stale-snapshot':
      return result.requestID;
    default: {
      const impossible: never = result;
      return impossible;
    }
  }
}
void exhaust;
