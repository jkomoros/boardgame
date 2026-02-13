/**
 * TypeScript declarations for test globals
 */

import { Store } from 'redux';

declare global {
  interface Window {
    /**
     * Redux store exposed for testing purposes
     * Set by the exposeStore() helper in fixtures.ts
     */
    __TEST_STORE__?: Store;
  }
}

export {};
