import '../client.js';
import type { DieComponent } from './boardgame-die.js';

// THE TYPED PIN ON THE INDEX. `selectedFaceIndex` is an index into `faces`, and
// the one thing a compiler can hold is that it is a number of the die's own
// naming -- so a renderer that still writes the old `selectedFace` fails here
// rather than silently setting an expando that nothing reads.
const die = document.createElement('boardgame-die');
die.faces = [10, 20, 30, 40, 50, 60];
die.selectedFaceIndex = 2;
die.faceNames = { 30: 'Star' };
die.symbols = { Star: '★' };
die.stateVersion = 7;
die.disabled = false;

// The value accessor a die at rest exposes. Readonly and derived: the index
// says which face, this says what is on it.
const shown: number | null = die.value;
void shown;

declare const component: DieComponent;
die.item = component;
die.item = null;

// `document.createElement` resolves the element type through
// HTMLElementTagNameMap, so a die queried out of the DOM is a die and not a
// bare Element -- which is what makes every line above a real check.
const found = document.querySelector('boardgame-die');
void (found ? found.value : null);

// @ts-expect-error the index is a number, not a face name
die.selectedFaceIndex = 'Star';
// @ts-expect-error the old spelling named the face, not the index into it
die.selectedFace = 2;
// @ts-expect-error `value` is what the die shows, not something a caller sets
die.value = 30;
// @ts-expect-error face values are numbers; a name belongs in faceNames
die.faces = ['Star'];
// @ts-expect-error a card is not a die: it has no Values.Faces
die.item = { ID: 'x', Values: { Rank: 3 } };
