# Game Renderer Migration to Lit 3 + TypeScript

## Status: ✅ COMPLETE

All game renderers and the base renderer have been successfully migrated from Polymer 3 to Lit 3 with full TypeScript support.

---

## Migration Summary

### What Was Migrated

**Core Components (Phase 0 - Supporting Components):**
- ✅ `boardgame-fading-text` - Animated text overlay component
- ✅ `boardgame-board` - Checkerboard grid component
- ✅ `boardgame-die` - Animated die component

**Base Infrastructure (Phase 1):**
- ✅ `boardgame-base-game-renderer` - Base class for all game renderers

**Example Games (Phase 3):**
1. ✅ **Pig** - Simple dice game (40 lines)
   - `boardgame-render-game-pig.ts`
   - `boardgame-render-player-info-pig.ts`

2. ✅ **Checkers** - Board game with computed properties (54 lines)
   - `boardgame-render-game-checkers.ts`

3. ✅ **Tic Tac Toe** - Grid game with custom component (120 lines total)
   - `boardgame-render-game-tictactoe.ts`
   - `boardgame-render-player-info-tictactoe.ts`
   - `boardgame-tictactoe-cell.ts` (custom component)

4. ✅ **Blackjack** - Card game with @apply replacements (110 lines total)
   - `boardgame-render-game-blackjack.ts`
   - `boardgame-render-player-info-blackjack.ts`

5. ✅ **Memory** - Timed matching game with animation delays (165 lines total)
   - `boardgame-render-game-memory.ts`
   - `boardgame-render-player-info-memory.ts`

6. ✅ **Debug Animations** - Comprehensive test game (405 lines)
   - `boardgame-render-game-debuganimations.ts`

**Total Files Migrated:** 15 TypeScript files

---

## Breaking Changes for External Games

### 🔴 CRITICAL: Event System Change

**Old (Polymer):**
```javascript
this.addEventListener('tap', handler);
```

**New (Lit 3):**
```javascript
this.addEventListener('click', handler);
```

**Impact:** All custom game renderers using `tap` events must change to `click` events.

**Affected Components:**
- `boardgame-base-game-renderer` - Move proposal click handling
- `boardgame-die` - Die interaction
- Any custom components using tap events

**Migration Required:** If your game uses custom tap event handlers, replace with click handlers. Mobile/touch interactions now use standard click events.

---

## Migration Patterns for External Games

### 1. Component Declaration

**Old (Polymer):**
```javascript
import { PolymerElement } from '@polymer/polymer/polymer-element.js';
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class MyGameRenderer extends BoardgameBaseGameRenderer {
  static get template() {
    return html`...`;
  }

  static get is() {
    return "my-game-renderer"
  }
}

customElements.define(MyGameRenderer.is, MyGameRenderer);
```

**New (Lit 3):**
```typescript
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
import { html, css } from 'lit';
import { property } from 'lit/decorators.js';

class MyGameRenderer extends BoardgameBaseGameRenderer {
  static override styles = [
    ...(BoardgameBaseGameRenderer.styles ? [BoardgameBaseGameRenderer.styles] : []),
    css`
      /* Your styles */
    `
  ];

  override render() {
    return html`...`;
  }
}

customElements.define('my-game-renderer', MyGameRenderer);
```

### 2. Properties

**Old (Polymer):**
```javascript
static get properties() {
  return {
    myProp: {
      type: String,
      value: "default"
    },
    computed: {
      type: Number,
      computed: "_computeValue(someProp)"
    }
  };
}
```

**New (Lit 3):**
```typescript
@property({ type: String })
myProp = 'default';

get computed(): number {
  return this._computeValue(this.someProp);
}
```

### 3. CSS: Replacing @apply

**Old (Polymer with iron-flex-layout):**
```css
#container {
  @apply --layout-horizontal;
  @apply --layout-center;
  @apply --layout-flex;
}
```

**New (Standard CSS):**
```css
#container {
  display: flex;
  flex-direction: row;
  align-items: center;
  flex: 1;
}
```

**Common Replacements:**
- `@apply --layout-horizontal;` → `display: flex; flex-direction: row;`
- `@apply --layout-vertical;` → `display: flex; flex-direction: column;`
- `@apply --layout-center;` → `align-items: center;`
- `@apply --layout-center-justified;` → `justify-content: center;`
- `@apply --layout-flex;` → `flex: 1;`

### 4. Template Syntax

**Old (Polymer):**
```html
<boardgame-component-stack
  stack="[[state.Game.DrawStack]]"
  layout="stack"
  messy$="[[messy]]"
  disabled="{{!isCurrentPlayer}}">
</boardgame-component-stack>

<template is="dom-repeat" items="[[players]]">
  <div>Player [[index]]: [[item.name]]</div>
</template>
```

**New (Lit 3):**
```typescript
import { repeat } from 'lit/directives/repeat.js';

html`
  <boardgame-component-stack
    .stack="${this.state?.Game?.DrawStack}"
    layout="stack"
    ?messy="${this.messy}"
    ?disabled="${!this.isCurrentPlayer}">
  </boardgame-component-stack>

  ${repeat(this.players, (player, index) => html`
    <div>Player ${index}: ${player.name}</div>
  `)}
`
```

**Binding Syntax:**
- `[[prop]]` (one-way) → `${this.prop}` (interpolation)
- `{{prop}}` (two-way) → `.prop="${this.prop}"` (property binding)
- `prop$="value"` (attribute) → `prop="value"` (attribute)
- `?attribute` (boolean) → `?attribute="${bool}"` (boolean attribute)

### 5. Two-Way Binding with Paper Elements

**Old (Polymer - automatic two-way binding):**
```html
<paper-toggle-button checked="{{messy}}">Messy</paper-toggle-button>
<paper-slider value="{{scale}}" min="0.5" max="2.0"></paper-slider>
```

**New (Lit 3 - manual event handlers):**
```typescript
html`
  <paper-toggle-button
    ?checked="${this.messy}"
    @checked-changed="${(e: CustomEvent) => { this.messy = e.detail.value; }}">
    Messy
  </paper-toggle-button>

  <paper-slider
    value="${this.scale}"
    @value-changed="${(e: CustomEvent) => { this.scale = e.detail.value; }}"
    min="0.5"
    max="2.0">
  </paper-slider>
`
```

### 6. Dynamic Styles

**Old (Polymer):**
```html
<div style$="--component-scale:[[scale]]">...</div>
```

**New (Lit 3 with styleMap):**
```typescript
import { styleMap } from 'lit/directives/style-map.js';

html`
  <div style="${styleMap({ '--component-scale': this.scale.toString() })}">
    ...
  </div>
`
```

### 7. Property Observers

**Old (Polymer):**
```javascript
static get properties() {
  return {
    item: {
      type: Object,
      observer: "_itemChanged"
    }
  };
}

_itemChanged(newValue, oldValue) {
  // React to change
}
```

**New (Lit 3):**
```typescript
@property({ type: Object })
item: any = null;

override updated(changedProperties: Map<PropertyKey, unknown>) {
  super.updated(changedProperties);

  if (changedProperties.has('item')) {
    this._itemChanged(this.item);
  }
}

private _itemChanged(newValue: any) {
  // React to change
}
```

### 8. Lifecycle Methods

**Old (Polymer):**
```javascript
ready() {
  super.ready();
  this.addEventListener('tap', this._handleTap);
}
```

**New (Lit 3):**
```typescript
private _boundHandleTap?: (e: Event) => void;

override firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
  super.firstUpdated(_changedProperties);

  this._boundHandleTap = (e: Event) => this._handleTap(e);
  this.addEventListener('click', this._boundHandleTap);
}

override disconnectedCallback() {
  super.disconnectedCallback();
  if (this._boundHandleTap) {
    this.removeEventListener('click', this._boundHandleTap);
  }
}
```

### 9. Shadow DOM Queries

**Old (Polymer):**
```javascript
ready() {
  super.ready();
  let element = this.$.myElement;
  // or
  let element = this.shadowRoot.querySelector('my-element');
}
```

**New (Lit 3 - with decorator):**
```typescript
import { query } from 'lit/decorators.js';

@query('#myElement')
private _myElement?: HTMLElement;

// Or with manual query and timing:
override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
  super.firstUpdated(_changedProperties);
  await this.updateComplete; // CRITICAL: Wait for render

  const element = this.renderRoot.querySelector('my-element');
}
```

---

## Animation Override Methods

The base renderer provides two methods that games can override to control animations:

### animationLength(fromMove, toMove): number

**Purpose:** Control animation duration in milliseconds.

**Return Values:**
- `0` (default) - Use default animation length
- `> 0` - Set animation length to this value in ms
- `< 0` - Skip animation entirely

**Example:**
```typescript
override animationLength(fromMove: any, toMove: any): number {
  if (toMove?.Name === 'Quick Move') {
    return 100; // Fast animation
  }
  if (toMove?.Name === 'Slow Move') {
    return 2000; // Slow animation
  }
  return 0; // Default
}
```

### delayAnimation(fromMove, toMove): number

**Purpose:** Delay before installing new state (in milliseconds).

**Use Case:** Show intermediate state before applying changes (e.g., show matched cards before removing them).

**Example (from Memory game):**
```typescript
override delayAnimation(fromMove: any, toMove: any): number {
  if (toMove && toMove.Name === 'Capture Cards') {
    // Show the cards for a second before capturing them
    return 1000;
  }
  return 0;
}
```

---

## Polymer Dependencies Still Required

The following Polymer/Paper components are still used and must remain as dependencies:

**Paper Components:**
- `@polymer/paper-button` - Button component
- `@polymer/paper-toggle-button` - Toggle switches
- `@polymer/paper-slider` - Range sliders
- `@polymer/paper-dropdown-menu` - Dropdown menus
- `@polymer/paper-listbox` - List selection
- `@polymer/paper-item` - List items
- `@polymer/paper-progress` - Progress bars

**Note:** These work via Polymer's compatibility layer. Future phases may replace these with native Lit components or web components.

---

## TypeScript Configuration

Games can be written in either TypeScript (`.ts`) or JavaScript (`.js`). TypeScript is recommended for:
- Better IDE support
- Type safety
- Early error detection

**Basic TypeScript Types:**
```typescript
// Use 'any' for game state (complex Go-generated objects)
@property({ type: Object })
state: any = null;

@property({ type: Object })
chest: any = null;

// Specific types for your properties
@property({ type: Boolean })
myFlag = false;

@property({ type: Number })
myNumber = 0;

@property({ type: String })
myString = '';

@property({ type: Array })
myArray: string[] = [];
```

---

## Testing Your Migration

### 1. Type Check
```bash
cd server/static
npm run type-check
```

Should complete with no errors.

### 2. Visual Testing
```bash
cd /path/to/boardgame
./dev.sh
```

Then test each game:
- Create a game
- Make moves
- Verify animations work
- Check for console errors
- Test on mobile/touch devices

### 3. Checklist per Game

- [ ] Game renders correctly
- [ ] Move proposals work (clicks register)
- [ ] `isCurrentPlayer` computed property updates
- [ ] Animation hooks fire (if overridden)
- [ ] State binding updates UI
- [ ] Paper components respond to interactions
- [ ] No console errors or warnings
- [ ] CSS layouts not broken (@apply replacements correct)
- [ ] Mobile/touch events work (click, not tap)

---

## Common Migration Issues

### Issue: "Cannot find module" errors

**Solution:** Check import paths. TypeScript files should use:
```typescript
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
```

Note: Use `.js` extension even for `.ts` files (TypeScript convention).

### Issue: Property not updating

**Solution:**
1. Ensure property uses `@property()` decorator
2. Check if property is private (should be public/protected for reactive properties)
3. Verify parent is passing with `.prop="${value}"` not `prop="${value}"`

### Issue: Events not firing

**Solution:**
1. Change `tap` to `click`
2. Ensure event listeners are bound in `firstUpdated()`
3. Add cleanup in `disconnectedCallback()`

### Issue: CSS not applying

**Solution:**
1. Use `static override styles` array, not `<style>` tag
2. Extend parent styles if needed
3. Replace all `@apply` with standard CSS

### Issue: Two-way binding not working

**Solution:** Lit doesn't have automatic two-way binding. Add manual event handlers:
```typescript
@checked-changed="${(e: CustomEvent) => { this.prop = e.detail.value; }}"
```

---

## Migration Checklist for External Games

Use this checklist when migrating your own game:

- [ ] Create `.ts` file alongside `.js` file
- [ ] Update imports to use TypeScript base renderer
- [ ] Replace `PolymerElement` with `LitElement` or `BoardgameBaseGameRenderer`
- [ ] Convert `static get template()` to `override render()`
- [ ] Convert `static get properties()` to `@property()` decorators
- [ ] Convert computed properties to getters
- [ ] Replace `@apply` with standard CSS flexbox
- [ ] Replace `tap` events with `click` events
- [ ] Convert `dom-repeat` to `repeat()` directive
- [ ] Fix binding syntax (`[[]]` → `${}`, `{{}}` → `.prop`)
- [ ] Add two-way binding event handlers for paper-* components
- [ ] Convert property observers to `updated()` method
- [ ] Convert `ready()` to `firstUpdated()` with cleanup
- [ ] Add TypeScript types to properties and methods
- [ ] Run type check: `npm run type-check`
- [ ] Test functionality in browser
- [ ] Test on mobile/touch devices
- [ ] Check console for errors

---

## Example: Complete Migration

See `examples/pig/client/` for the simplest example of a complete migration.

See `examples/debuganimations/client/` for the most comprehensive example including:
- 11 reactive properties
- Two-way binding with paper components
- Dynamic styling with styleMap
- Shadow DOM queries
- 14 @apply CSS replacements

---

## Getting Help

If you encounter issues migrating your game:

1. Check this document for patterns
2. Look at the example games for reference
3. Run `npm run type-check` to catch type errors
4. Check browser console for runtime errors
5. File an issue at: https://github.com/jkomoros/boardgame/issues

---

## Commits

The migration was completed in the following commits:

**Phase 0 - Supporting Components:**
- `64a337aa` - Migrate boardgame-fading-text to Lit 3
- `4f81d224` - Migrate boardgame-board to Lit 3
- `aa3a03fe` - Migrate boardgame-die to Lit 3

**Phase 1 - Base Renderer:**
- `aa30d88e` - Migrate boardgame-base-game-renderer to Lit 3 + TypeScript

**Phase 3 - Games:**
- `3e41bc5b` - Migrate Pig game renderer to Lit 3
- `1614ebe5` - Migrate Checkers game renderer to Lit 3
- `037a8347` - Migrate Tic Tac Toe game renderer + cell component to Lit 3
- `60a81705` - Migrate Blackjack game renderer to Lit 3
- `19f7962d` - Migrate Memory game renderer to Lit 3
- `c5540ee9` - Migrate Debug Animations game renderer to Lit 3

**Merge:**
- `b779481b` - Merge migrate-example-games into architecture-documentation
- `f97be6ee` - Fix: Resolve all merge conflicts in boardgame-fading-text.ts
