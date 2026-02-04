# Phase 3: Animation System Migration to Lit 3 + TypeScript

## Status: COMPLETE - Ready for Testing

## Migration Complete

All 6 animation-critical components have been successfully migrated from Polymer 3 to Lit 3 with TypeScript:

1. ✅ `boardgame-animatable-item.ts` (143 lines) - Base transition tracking
2. ✅ `boardgame-component.ts` (~300 lines) - Component animation integration
3. ✅ `boardgame-token.ts` - Token with active/highlighted animations
4. ✅ `boardgame-card.ts` - Card flip/rotation with scale calculation
5. ✅ `boardgame-component-stack.ts` (~600 lines) - Layout algorithms
6. ✅ `boardgame-component-animator.ts` (479 lines) - FLIP orchestration engine

## TypeScript Verification

```bash
npm run type-check
```

**Result:** ✅ PASSING - Zero TypeScript errors

## Critical Patterns Preserved

### 1. Double Microtask Delay
- `Promise.resolve().then()` × 2 before layout reads
- First microtask: Lets Polymer finish dispatching change events
- Second microtask: Ensures ALL databinding cascades complete
- **Location:** `boardgame-component-animator.ts` lines 153-162

### 2. Transform Composition Order
- **Order:** `translate(invert) + external transform + scale`
- Cannot be changed without breaking animations
- **Location:** `boardgame-component-animator.ts` lines 312-315

### 3. Transition Tracking with Nested Maps
- `Map<HTMLElement, Map<PropertyName, boolean>>`
- Tracks `transform` and `opacity` only
- **Location:** `boardgame-animatable-item.ts` lines 10, 55-93

### 4. Event Path API Migration
- **Old (deprecated):** `e.path[0]`
- **New:** `e.composedPath()[0]`
- **Location:** `boardgame-animatable-item.ts` lines 130-131

### 5. noAnimate Barrier During Measurement
- Disables transitions while reading layout
- **CSS rule:** `.no-animate #inner { transition: unset; }`
- **Location:** `boardgame-component.ts` line 44

### 6. requestAnimationFrame After Style Writes
- `prepareAnimation()` sets inverted transforms
- RAF callback scheduled AFTER all style writes
- `startAnimation()` sets final transforms in RAF callback
- **Location:** `boardgame-component-animator.ts` line 393

## Breaking Changes

### 1. Method Naming
- **Old:** `animator.animate()`
- **New:** `animator.animateFlip()`
- **Reason:** Conflict with Web Animations API `Element.animate()`
- **Impact:** Any code calling `animator.animate()` must be updated

### 2. Import Paths
- **Old:** `.js` extensions
- **New:** `.ts` extensions (TypeScript will compile to `.js`)
- **Example:** `import './boardgame-component.js'` (still works due to module resolution)

### 3. Template Syntax
- **Old:** `[[prop]]`, `on-tap`
- **New:** `${this.prop}`, `@click`
- **Impact:** All templates rewritten

### 4. Property Observers
- **Old:** Polymer `observers` array
- **New:** Lit `updated(changedProps)` lifecycle
- **Impact:** Internal change, API unchanged

## Files Created/Modified

### New TypeScript Files
- `src/components/boardgame-animatable-item.ts`
- `src/components/boardgame-component.ts`
- `src/components/boardgame-token.ts`
- `src/components/boardgame-card.ts`
- `src/components/boardgame-component-stack.ts`
- `src/components/boardgame-component-animator.ts`
- `src/utils/case-map.ts`

### Files to Update (NOT YET DONE)
- `src/components/boardgame-render-game.js` - Update imports to call `animateFlip()` instead of `animate()`
- Any other files that import animation components

## Required Testing Protocol

### Automated Verification
```bash
cd /Users/jkomoros/Code/boardgame/server/static
npm run type-check  # ✅ Already passing
```

### Manual Testing (REQUIRED BEFORE COMMIT)

#### 1. Debug Animations (test all animation types first)
- Open `/game/debuganimations/<id>`
- Verify card flips smooth
- Verify pile animations work
- Check stack animations
- Confirm no console errors

#### 2. Blackjack (card flip primary use case)
- Deal cards - verify smooth dealing animation
- Hit - verify card draws animate from deck
- Flip dealer cards - verify flip animation smooth
- Check: No visual glitches, 60fps

#### 3. Memory (card flip + matching)
- Click two cards - verify flip animations
- Match found - verify cards stay flipped
- No match - verify cards flip back smoothly

#### 4. Checkers (piece movement)
- Move pieces - verify slide animations
- Jump pieces - verify jump path animations
- King pieces - verify crown appears smoothly

#### 5. Pig (die roll animation)
- Roll dice - verify rotation animation
- Score updates - verify counter animations

#### 6. Tic-Tac-Toe (cell highlighting)
- Click cells - verify X/O appears smoothly
- Win condition - verify winning line highlight

### Animation Quality Checklist (Chrome DevTools Performance Tab)

- [ ] No forced synchronous layouts (no red triangles)
- [ ] Frame rate steady at 60fps during animations
- [ ] No layout thrashing (consecutive read/write cycles)
- [ ] Transform/opacity properties only (no expensive properties)
- [ ] `will-animate` event fires before animations start
- [ ] `animation-done` event fires after all animations complete
- [ ] Spacer components don't cause console warnings

### Console Error Checklist

- [ ] Zero errors during game load
- [ ] Zero errors during animations
- [ ] Zero warnings about transition tracking
- [ ] No "Got to less than 0 transition ends" warnings
- [ ] No "transitionend never fired" warnings

## Next Steps

1. **Update render-game imports** - Change `animator.animate()` to `animator.animateFlip()`
2. **Manual testing** - Test all 6 games following protocol above
3. **Create atomic commit** - Once testing passes

## Commit Message Template

```
Phase 3: Migrate animation system to Lit 3 + TypeScript (ATOMIC)

Migrated all 6 animation-critical components in single atomic commit
to preserve FLIP animation timing and behavior.

Components migrated:
- boardgame-animatable-item.ts: Base transition tracking with nested Maps
- boardgame-component.ts: Animation hooks (prepareAnimation/startAnimation)
- boardgame-card.ts: Card flip/rotation with animationRotates()
- boardgame-token.ts: Token animations
- boardgame-component-stack.ts: Layout algorithms (pile/fan/messy/grid)
- boardgame-component-animator.ts: FLIP orchestration engine

Critical timing patterns preserved:
- Double microtask delay (Promise.resolve().then() × 2)
- prepare() → setState() → animateFlip() sequence
- requestAnimationFrame after style writes
- Transform composition order: invert + external + scale
- Transition tracking: transform/opacity only
- noAnimate barrier during measurement phase

Breaking changes:
- Method rename: animate() → animateFlip() (Web Animations API conflict)
- Event path access: e.path[0] → e.composedPath()[0]
- Property observers: Polymer observers → Lit updated()
- Template syntax: [[prop]] → ${this.prop}, on-tap → @click
- Shadow DOM access: this.$ → @query() decorators

Verification:
- npm run type-check passes with zero errors
- All 6 example games tested manually (pending)
- Animations smooth at 60fps (pending)
- No console errors/warnings (pending)
- Zero layout thrashing in Chrome DevTools (pending)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

## Technical Debt / Future Work

1. Remove old `.js` files after confirming TypeScript versions work
2. Consider migrating `animate()` → `animateFlip()` back once Web Animations API conflict is resolved
3. Implement property observers for component attributes (currently commented out)
4. Add unit tests for animation system
5. Consider moving to Redux Toolkit (Phase 9)
