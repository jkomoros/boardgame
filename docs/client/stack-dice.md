# Stack-managed dice

`dieView()` creates a real `BoardgameComponent` host. Use it with ordinary
`boardgame-component-stack` or `boardgame-component-zone`; stack identity,
selection, actions, layout and structural transfers work as for cards and tokens.
Inline `<boardgame-die .item=${die} .action=${rollAction}>` keeps its own button.

```ts
private readonly dice = dieView<MyDiceStack>({
  properties: () => ({ faceNames: { 6: 'Star' }, symbols: { Star: '★' } }),
  rollBudget: { durationMs: 900, maxSolidDice: 5 },
});
private readonly keep = new SelectionDraftController<string>(this);

render() {
  const selection = this.keep.draft({
    candidates: this.state.Game.Tray.Components.flatMap(die => die?.ID ? [die.ID] : []),
    maxSelected: 5,
    rebase: 'keep-valid',
  });
  return html`<boardgame-component-zone label="Tray" layout="spread"
    .stack=${this.state.Game.Tray} .componentView=${this.dice}
    .selection=${selection}></boardgame-component-zone>`;
}
```

Selection string keys are component IDs, so reordering cannot move a selection
to another die. Removing a candidate prunes the existing controller's selection.
Use `selection.selected` to construct the game's existing keep move.

A changed `DynamicValues.RollCount` triggers a roll, including a repeated result.
Mounting, changing identity, hiding, clearing and correcting a face without a new
counter install a quiet snapshot. The selected face remains an index into
`Values.Faces`; simulation assigns that authoritative value to the landed facet.
The die's inner surface owns roll transforms, while the component host owns
structural movement. Both use the existing animation gate and cancellation
lifecycle. Historical carriers copy only rendered faces, names, glyphs and pose;
they never acquire an item or a roll counter.

The optional budget applies separately to each stack, including when a recipe is
shared between two zones. The first `maxSolidDice` visible dice in slot order use
solids; remaining dice use the existing flat reel. Hidden and empty slots spend
no allowance. Reordering across the boundary changes presentation quietly and
keeps the correct value. `durationMs` caps the playback of the complete trajectory,
including its final pose; it does not truncate a simulated throw. Its supported
range is 0–5000 ms; `maxSolidDice` is an integer from 0–32. All simultaneous dice
share that upper duration bound rather than accumulating serial delays. Reduced
motion bypasses roll simulation and installs the authoritative result directly.

Open `/dice-demo.html` on the development server for a usable five-die keep/reroll
fixture. It installs deterministic sample snapshots, supports pointer and keyboard
selection, moves dice between two ordinary zones, repeats results, and switches
to twenty dice. It adds no scoring or game rules.

Browser measurements on the same Chromium run (50px d6, repeated result):

| Pool | Planning/render update | Roll keyframes | Facets | Longest roll |
|---|---:|---:|---:|---:|
| 5, unbudgeted | 48 ms | 142 | 30 | 778 ms |
| 20, unbudgeted | 162 ms | 755 | 120 | 1983 ms |
| 20, budget 5 solids / 900 ms | 63 ms | 142 | 30 | 778 ms |

These are local observations, not timing assertions. Repeated-result reels need
no visual track; changed reel results use two keyframes. Browser tests assert the
allocation/duration bounds, authoritative values, gate release, hidden and history
cleanup, separate transform ownership, and desktop/phone keep/reroll interaction.
