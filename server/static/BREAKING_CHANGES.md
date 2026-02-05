# Breaking Changes - Lit 3 Migration

## Summary

The migration from Polymer 3 to Lit 3 introduces breaking changes that affect external game implementations. All games in the `examples/` directory have been migrated and serve as reference implementations.

---

## 🔴 BREAKING: Event System Change (tap → click)

### Impact: HIGH
**Affects:** All games using event handling for move proposals and interactions

### What Changed

**Before (Polymer):**
- Used Polymer's unified `tap` event for touch and click
- `tap` event fired on both touch and click interactions

**After (Lit 3):**
- Uses standard `click` event
- Modern browsers handle touch and click consistently with `click`

### Migration Required

**In boardgame-base-game-renderer:**
```javascript
// OLD - Polymer 3
this.addEventListener('tap', e => this._handleButtonTapped(e));

// NEW - Lit 3
this.addEventListener('click', e => this._handleButtonTapped(e));
```

**In custom components:**
```javascript
// OLD - Polymer 3
this.shadowRoot.addEventListener('tap', (e) => this._handleTap(e));

// NEW - Lit 3
this.renderRoot.addEventListener('click', (e) => this._handleClick(e));
```

**In templates:**
```html
<!-- OLD - Polymer 3 -->
<div on-tap="_handleClick">Click me</div>

<!-- NEW - Lit 3 -->
<div @click="${this._handleClick}">Click me</div>
```

### Who Is Affected

- **External games** extending `BoardgameBaseGameRenderer`
- **Custom components** using tap events
- **Mobile/touch interfaces** (must test on actual devices)

### Action Required

1. Search codebase for `tap` event listeners
2. Replace with `click` event listeners
3. Test on mobile/touch devices
4. Verify move proposals still work

---

## 🟡 BREAKING: Import Paths Changed

### Impact: MEDIUM
**Affects:** All external game renderers

### What Changed

**Before (Polymer):**
```javascript
import { BoardgameBaseGameRenderer } from '../../src/components/boardgame-base-game-renderer.js';
```

**After (Lit 3):**
```typescript
import { BoardgameBaseGameRenderer } from '../../../server/static/src/components/boardgame-base-game-renderer.js';
```

### Why

TypeScript files use `.js` extension in imports (TypeScript convention), and paths may differ based on file location.

### Who Is Affected

All external games importing:
- `boardgame-base-game-renderer`
- `boardgame-die`
- `boardgame-board`
- `boardgame-fading-text`
- Any other migrated components

### Action Required

Update all imports to reference `.ts` files (but keep `.js` extension in import path).

---

## 🟡 BREAKING: @apply CSS Directive Removed

### Impact: MEDIUM
**Affects:** Games using iron-flex-layout or custom @apply mixins

### What Changed

**Before (Polymer with iron-flex-layout):**
```css
.container {
  @apply --layout-horizontal;
  @apply --layout-center;
}
```

**After (Standard CSS):**
```css
.container {
  display: flex;
  flex-direction: row;
  align-items: center;
}
```

### Why

- `@apply` was a non-standard CSS feature
- Removed from web platform standards
- Modern CSS flexbox/grid are well-supported

### Common Replacements

| Old @apply | New CSS |
|------------|---------|
| `@apply --layout-horizontal;` | `display: flex; flex-direction: row;` |
| `@apply --layout-vertical;` | `display: flex; flex-direction: column;` |
| `@apply --layout-center;` | `align-items: center;` |
| `@apply --layout-center-justified;` | `justify-content: center;` |
| `@apply --layout-flex;` | `flex: 1;` |
| `@apply --paper-font-title;` | Use explicit font CSS |

### Who Is Affected

Games with custom CSS using:
- `iron-flex-layout` mixins
- `paper-font` mixins
- Custom `@apply` directives

### Action Required

1. Search for `@apply` in your CSS
2. Replace with standard CSS equivalents
3. Test layouts visually
4. Check responsive behavior

---

## 🟢 BREAKING: Two-Way Data Binding Syntax Changed

### Impact: LOW (Pattern change, not a bug)
**Affects:** Games using Polymer's automatic two-way binding

### What Changed

**Before (Polymer - automatic):**
```html
<paper-toggle-button checked="{{myProperty}}"></paper-toggle-button>
<paper-slider value="{{scale}}"></paper-slider>
```

**After (Lit 3 - manual):**
```typescript
html`
  <paper-toggle-button
    ?checked="${this.myProperty}"
    @checked-changed="${(e: CustomEvent) => { this.myProperty = e.detail.value; }}">
  </paper-toggle-button>

  <paper-slider
    value="${this.scale}"
    @value-changed="${(e: CustomEvent) => { this.scale = e.detail.value; }}">
  </paper-slider>
`
```

### Why

Lit doesn't provide automatic two-way binding - it's explicit by design for clarity and predictability.

### Who Is Affected

Games using:
- `paper-toggle-button`
- `paper-slider`
- `paper-input`
- `paper-dropdown-menu` / `paper-listbox`
- Any component with two-way binding

### Action Required

1. Identify `{{property}}` bindings
2. Add manual event handlers
3. Test interactivity works

---

## 🟢 BREAKING: Template Syntax Changed

### Impact: LOW (Mechanical change)
**Affects:** All game templates

### What Changed

**Data Binding:**
```html
<!-- OLD - Polymer -->
<div>[[property]]</div>
<div>{{twoWayProperty}}</div>
<div class$="[[computedClass]]"></div>

<!-- NEW - Lit 3 -->
<div>${this.property}</div>
<div .property="${this.twoWayProperty}"></div>
<div class="${this.computedClass}"></div>
```

**Boolean Attributes:**
```html
<!-- OLD - Polymer -->
<button disabled="[[!isEnabled]]"></button>

<!-- NEW - Lit 3 -->
<button ?disabled="${!this.isEnabled}"></button>
```

**Property Binding:**
```html
<!-- OLD - Polymer -->
<boardgame-component-stack stack="[[state.Game.DrawStack]]">

<!-- NEW - Lit 3 -->
<boardgame-component-stack .stack="${this.state?.Game?.DrawStack}">
```

**Event Binding:**
```html
<!-- OLD - Polymer -->
<button on-tap="_handleClick"></button>

<!-- NEW - Lit 3 -->
<button @click="${this._handleClick}"></button>
```

**Loops:**
```javascript
// OLD - Polymer
html`
  <template is="dom-repeat" items="[[players]]">
    <div>[[item.name]]</div>
  </template>
`

// NEW - Lit 3
import { repeat } from 'lit/directives/repeat.js';

html`
  ${repeat(this.players, (player) => html`
    <div>${player.name}</div>
  `)}
`
```

### Who Is Affected

All games with templates.

### Action Required

Mechanical find-replace in templates (see GAME_RENDERER_MIGRATION.md for complete guide).

---

## 🟢 BREAKING: Property Declaration Syntax Changed

### Impact: LOW (Mechanical change)
**Affects:** All game components with properties

### What Changed

**Before (Polymer):**
```javascript
static get properties() {
  return {
    myProp: {
      type: String,
      value: "default"
    },
    computed: {
      type: Number,
      computed: "_computeValue(dep1, dep2)"
    }
  };
}
```

**After (Lit 3):**
```typescript
import { property } from 'lit/decorators.js';

@property({ type: String })
myProp = 'default';

get computed(): number {
  return this._computeValue(this.dep1, this.dep2);
}
```

### Who Is Affected

All game components declaring properties.

### Action Required

1. Import `property` decorator
2. Convert to decorator syntax
3. Convert computed properties to getters

---

## 🟢 BREAKING: Lifecycle Methods Changed

### Impact: LOW (Pattern change)
**Affects:** Games with lifecycle hooks

### What Changed

**Before (Polymer):**
```javascript
ready() {
  super.ready();
  this.addEventListener('tap', this._handleTap);
}
```

**After (Lit 3):**
```typescript
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

### Why

- Lit uses standard web component lifecycle
- Requires explicit event listener cleanup
- Better memory management

### Who Is Affected

Games with:
- Event listeners added in lifecycle
- DOM queries in lifecycle
- Initialization logic

### Action Required

1. Convert `ready()` to `firstUpdated()`
2. Add `disconnectedCallback()` for cleanup
3. Store bound event handlers for removal

---

## 🟢 BREAKING: Shadow DOM Query Changed

### Impact: LOW (API change)
**Affects:** Games querying shadow DOM

### What Changed

**Before (Polymer):**
```javascript
ready() {
  super.ready();
  let element = this.$.myElement; // Automatic ID map
  // or
  let element = this.shadowRoot.querySelector('my-element');
}
```

**After (Lit 3 - Option 1: Decorator):**
```typescript
import { query } from 'lit/decorators.js';

@query('#myElement')
private _myElement?: HTMLElement;

// Access: this._myElement
```

**After (Lit 3 - Option 2: Manual):**
```typescript
override async firstUpdated(_changedProperties: Map<PropertyKey, unknown>) {
  super.firstUpdated(_changedProperties);
  await this.updateComplete; // CRITICAL

  const element = this.renderRoot.querySelector('my-element');
}
```

### Why

- Lit doesn't automatically map IDs
- More explicit and type-safe
- `renderRoot` instead of `shadowRoot` for consistency

### Who Is Affected

Games querying shadow DOM elements.

### Action Required

1. Replace `this.$` with `@query()` decorator
2. Replace `shadowRoot` with `renderRoot`
3. Use `await this.updateComplete` for timing

---

## Dependencies Still Required

These Polymer/Paper dependencies are still required and should remain in `package.json`:

```json
{
  "@polymer/paper-button": "^3.0.1",
  "@polymer/paper-toggle-button": "^3.0.1",
  "@polymer/paper-slider": "^3.0.1",
  "@polymer/paper-dropdown-menu": "^3.0.1",
  "@polymer/paper-listbox": "^3.0.1",
  "@polymer/paper-item": "^3.0.1",
  "@polymer/paper-progress": "^3.0.1"
}
```

These work via Polymer's compatibility layer. Future migrations may replace them.

---

## Backward Compatibility

**There is NO backward compatibility.** The old Polymer 3 renderers will not work with the new Lit 3 base renderer.

**Options:**
1. **Migrate your game** to Lit 3 (recommended - see GAME_RENDERER_MIGRATION.md)
2. **Pin to old version** of boardgame library (not recommended)
3. **Use old base renderer** (`.js` files still exist, but unmaintained)

---

## Testing Checklist

After migrating, verify:

- [ ] Game loads without errors
- [ ] Move proposals work (clicks register)
- [ ] Animations play correctly
- [ ] Computed properties update
- [ ] Paper components (buttons, sliders) work
- [ ] Mobile/touch events work
- [ ] Layout looks correct (CSS not broken)
- [ ] No console errors or warnings
- [ ] TypeScript compiles (if using TS)

---

## Migration Timeline

All breaking changes were introduced in the following commits:

- `64a337aa` - `boardgame-fading-text` migration (tap→click)
- `aa3a03fe` - `boardgame-die` migration (tap→click)
- `aa30d88e` - `boardgame-base-game-renderer` migration (tap→click, CRITICAL)

**Date:** February 5, 2026
**Branch:** `architecture-documentation`

---

## Need Help?

1. Check GAME_RENDERER_MIGRATION.md for complete migration patterns
2. Look at `examples/` directory for reference implementations
3. File issues at: https://github.com/jkomoros/boardgame/issues

---

## Summary Table

| Change | Impact | Affects | Fix Difficulty |
|--------|--------|---------|----------------|
| tap → click events | HIGH | All games | Easy |
| @apply CSS removed | MEDIUM | Games with custom CSS | Medium |
| Import paths changed | MEDIUM | All games | Easy |
| Two-way binding | LOW | Games with paper-* | Easy |
| Template syntax | LOW | All games | Easy |
| Property syntax | LOW | All games | Easy |
| Lifecycle methods | LOW | Games with init logic | Medium |
| Shadow DOM queries | LOW | Few games | Easy |
