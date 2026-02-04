# Lit 3 + TypeScript Migration Plan

## Current Status: Phase 0 - Foundation & Tooling Setup (COMPLETE - Ready to Commit)

This document tracks the incremental migration from Polymer 3 + old lit-element to modern Lit 3 with strict TypeScript.

---

## Phase Checklist

- [x] Phase 0.1: Create TypeScript configuration files (tsconfig.json, tsconfig.strict.json)
- [x] Phase 0.2: Create Vite configuration
- [x] Phase 0.3: Update package.json with new scripts and dependencies
- [x] Phase 0.4: Create Polymer bridge type definitions
- [x] Phase 0.5: Create animation type definitions
- [x] Phase 0.6: Update .gitignore for TypeScript artifacts
- [x] Phase 0.7: Install dependencies (Workaround: Firebase excluded from install, listed in package.json)
- [x] Phase 0.8: Verify TypeScript compiles with allowJs (✅ npm run type-check passes)
- [x] Phase 0.9: Commit Phase 0 (COMMITTED: 54ae58bc)

- [x] Phase 1: Type Infrastructure (COMMITTED: e93a0b44)
  - [x] Create Redux store type definitions (store.d.ts)
  - [x] Create component type definitions (components.d.ts)
  - [x] Create event type definitions (events.d.ts)
  - [x] Verify TypeScript compiles (✅ npm run type-check passes)

- [ ] Phase 2: Migrate Leaf Components (PARTIAL - Simple components only)
  - [x] boardgame-404-view.ts - Static error page
  - [x] boardgame-player-chip.ts - Player avatar with color hash
  - [ ] boardgame-player-roster-item - DEFERRED (depends on paper-*, needs Phase 5)
  - [ ] boardgame-player-roster - DEFERRED (depends on paper-dialog, dom-repeat)
- [ ] Phase 3: Animation System (CRITICAL - ATOMIC)
- [ ] Phase 4: Orchestration Layer
- [ ] Phase 5: UI & Form Components
- [ ] Phase 6: Upgrade Old Lit-Element Components
- [ ] Phase 7: External Game Renderer Compatibility
- [ ] Phase 8: Strict TypeScript Enforcement
- [ ] Phase 9: Redux Toolkit Migration (Optional)
- [ ] Phase 10: Cleanup & Optimization

---

## Current Issues

### Firebase 5.11.1 / gRPC Compilation Failure (RESOLVED)
**Problem:** Firebase 5.11.1 depends on gRPC 1.20.0 which won't compile on Node 20 + ARM64 macOS.

**Error:** String concatenation warning in `stats_data.cc` treated as error due to `-Werror`

**Solution Applied:**
- Firebase is listed in package.json for documentation purposes
- Firebase is NOT installed in node_modules (frontend-only TypeScript migration)
- This allows TypeScript compilation and Lit 3 migration to proceed
- Runtime Firebase usage depends on existing backend server
- Consider upgrading Firebase to v9+ in a separate migration phase

---

## Files Created/Modified

### Created:
- `/Users/jkomoros/Code/boardgame/server/static/tsconfig.json` - Base TypeScript config (lenient)
- `/Users/jkomoros/Code/boardgame/server/static/tsconfig.strict.json` - Strict TypeScript target
- `/Users/jkomoros/Code/boardgame/server/static/vite.config.ts` - Vite build configuration
- `/Users/jkomoros/Code/boardgame/server/static/src/types/polymer-bridge.d.ts` - Polymer type bridges
- `/Users/jkomoros/Code/boardgame/server/static/src/types/animation.d.ts` - Animation system types

### Modified:
- `/Users/jkomoros/Code/boardgame/server/static/package.json` - Added TypeScript, Vite, Lit 3
- `/Users/jkomoros/Code/boardgame/.gitignore` - Added TypeScript build artifacts

---

## Next Steps

1. Resolve Firebase/gRPC installation issue
2. Complete Phase 0 verification (TypeScript compiles)
3. Commit Phase 0
4. Begin Phase 1: Type Infrastructure

---

## Migration Principles

1. **Incremental:** Each phase must leave the system in a working state
2. **Atomic where critical:** Animation system (Phase 3) migrated all at once
3. **Type-safe:** Gradually enable strict TypeScript flags
4. **Preserve behavior:** All 6 example games must work after each phase
5. **FLIP animations:** Must remain smooth and identical to original

---

## Component Inventory (33 total)

### Animation-Critical (Phase 3 - Atomic):
- boardgame-animatable-item.js
- boardgame-component.js
- boardgame-card.js
- boardgame-token.js
- boardgame-component-stack.js
- boardgame-component-animator.js

### Old Lit-Element (Phase 6 - Upgrade 2.3.1 → 3.x):
- boardgame-app.js
- boardgame-game-view.js
- boardgame-list-games-view.js
- boardgame-create-game.js
- boardgame-user.js

### Other Polymer (26 components - Phases 2, 4, 5, 7):
- See full component list in detailed plan

---

## Testing Checklist (Run After Each Phase)

### Automated:
```bash
npm run type-check
npm run lint
```

### Manual (6 Example Games):
- [ ] Blackjack: Cards animate, flips work
- [ ] Checkers: Pieces move smoothly
- [ ] Debug Animations: All animation types functional
- [ ] Memory: Card flips, matches animate
- [ ] Pig: Die rolls, score updates
- [ ] Tic-Tac-Toe: Cell clicks, win detection

### Animation Validation:
- [ ] FLIP animations smooth (60fps)
- [ ] No console errors
- [ ] No layout thrashing (check Performance tab)
- [ ] Transform composition preserved
- [ ] Faux animating components work

---

## Rollback Strategy

Each phase is a git commit. To rollback:
```bash
git revert HEAD  # Revert last phase
git reset --hard <commit>  # Reset to specific phase
```

---

## Reference Files

- Full migration plan: (this document is the plan - update as needed)
- Architecture docs: `/Users/jkomoros/Code/boardgame/server/static/src/ARCHITECTURE.md`
- Blueprint codebase: `/Users/jkomoros/Code/card-web/` (for Lit 3 + TypeScript patterns)
- Components: `/Users/jkomoros/Code/boardgame/server/static/src/components/`
