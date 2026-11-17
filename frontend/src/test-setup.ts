// Shared set-up for the front-end suite: each test starts with no bridge, so a
// test that wants the window says so with installBridge.

import '@testing-library/jest-dom/vitest'
import { afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'

// Newer Node carries a localStorage of its own, undefined unless Node is handed
// a storage file, and it sits on the global ahead of the one jsdom provides.
// Measured 2026-11-17 on Node 26.13: sixteen tests read undefined from every
// storage call; with Node's own storage switched off all of them passed. A
// command-line flag would fix it only on a Node that knows the flag, so the
// suite is handed jsdom's storage here, whatever Node it runs on.
declare const jsdom: { window: Window } | undefined
if (typeof jsdom !== 'undefined') {
  for (const name of ['localStorage', 'sessionStorage'] as const) {
    Object.defineProperty(globalThis, name, {
      value: jsdom.window[name],
      configurable: true,
      writable: true,
    })
  }
}

afterEach(() => {
  cleanup()
  delete (window as unknown as { go?: unknown }).go
})
