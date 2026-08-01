import { BoardgameComponent, componentView } from '../client.js';

// THE CONTRADICTION THIS PINS. `componentView()` is the documented escape hatch
// for a game-owned component host, its type parameter is
// `ElementType extends BoardgameComponent`, and its runtime guard refuses
// anything else. Three documents told creators to write "a custom element
// extending BoardgameComponent" while `client.ts` -- the single supported entry
// point -- declined to export it, so the instruction could not be followed.
// Reaching it through the facade is now a compile-time fact.
class MeeplePiece extends BoardgameComponent {
  tone: 'neutral' | 'raised' = 'neutral';

  // The main override point: which of a subclass's OWN properties, changing,
  // is a visual transition the animator should carry.
  override get animatingProperties(): string[] {
    return ['tone'];
  }
}

const view = componentView<never, MeeplePiece>(
  () => new MeeplePiece(),
  { properties: () => ({ tone: 'raised' }) },
);
void view;

// A subclass's own properties are settable through the recipe...
view.withProperties({ tone: 'neutral' });

// ...and the six the framework owns are not, because the stack writes them as
// it binds state to hosts. `SettableComponentProperties` subtracts exactly
// these, and this is the assertion that it still does.
// @ts-expect-error `item` is bound by the stack, not by the recipe
view.withProperties({ item: null });
// @ts-expect-error `index` is the slot the stack put this host in
view.withProperties({ index: 2 });
// @ts-expect-error `spacer` is what an empty sized-stack slot becomes
view.withProperties({ spacer: true });
// @ts-expect-error `disabled` follows the stack's own interaction state
view.withProperties({ disabled: true });
// @ts-expect-error `id` is component identity, which is what FLIP pairs on
view.withProperties({ id: 'mine' });
// @ts-expect-error `boardgameComponent` is the marker the stack sets itself
view.withProperties({ boardgameComponent: false });
// @ts-expect-error a property the subclass does not have is not settable
view.withProperties({ notAThing: 1 });
