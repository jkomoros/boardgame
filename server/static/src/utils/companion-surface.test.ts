import test from 'node:test';
import assert from 'node:assert/strict';
import {
  forgetAllCompanionSurfaces,
  rememberSurfaceForGame,
  surfaceForGame,
} from './companion-surface.ts';

class MemoryStorage implements Storage {
  private readonly values = new Map<string, string>();

  get length(): number { return this.values.size; }
  clear(): void { this.values.clear(); }
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  key(index: number): string | null { return [...this.values.keys()][index] ?? null; }
  removeItem(key: string): void { this.values.delete(key); }
  setItem(key: string, value: string): void { this.values.set(key, value); }
}

function installBrowser(sessionStorage: Storage, localStorage = new MemoryStorage()): void {
  Object.defineProperty(globalThis, 'window', {
    configurable: true,
    value: {
      location: { search: '' },
      sessionStorage,
      localStorage,
    },
  });
  Object.defineProperty(globalThis, 'document', {
    configurable: true,
    value: { cookie: 'surface_GAME=table' },
  });
}

test('companion surface intent is tab-scoped and ignores origin-wide residue', () => {
  const sharedLocalStorage = new MemoryStorage();
  sharedLocalStorage.setItem('boardgame-surface:GAME', 'table');

  const handTab = new MemoryStorage();
  installBrowser(handTab, sharedLocalStorage);
  rememberSurfaceForGame('GAME', 'hand');
  assert.equal(surfaceForGame('GAME', true), 'hand');

  // A second tab shares localStorage and cookies, but not sessionStorage. It
  // must not silently become either the Table or Hand selected by another tab.
  installBrowser(new MemoryStorage(), sharedLocalStorage);
  assert.equal(surfaceForGame('GAME', true), null);
});

test('authoritative solo mode suppresses and clearing removes tab intent', () => {
  const tab = new MemoryStorage();
  installBrowser(tab);
  rememberSurfaceForGame('GAME', 'table');
  rememberSurfaceForGame('OTHER', 'hand');
  assert.equal(surfaceForGame('GAME', true), 'table');

  // Even an explicit stale restore URL is presentation intent, not authority.
  window.location.search = '?display=hand';
  assert.equal(surfaceForGame('GAME', false), null);
  assert.equal(surfaceForGame('GAME', true), 'hand');
  window.location.search = '';

  forgetAllCompanionSurfaces();
  assert.equal(surfaceForGame('GAME', true), null);
  assert.equal(surfaceForGame('OTHER', true), null);
});

test.after(() => {
  Reflect.deleteProperty(globalThis, 'window');
  Reflect.deleteProperty(globalThis, 'document');
});
