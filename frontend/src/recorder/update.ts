// Keeping the phone recorder up to date with nobody doing anything (FR-105).
// The service worker installs a new release in the background and takes over;
// this makes the open page follow it at a moment that cannot cost an entry
// being typed. It also asks for a newer release whenever the person comes back.

import { reloadOnReturn } from './start'

/** The parts of the browser's service worker container this uses. */
export interface Workers {
  controller: unknown
  addEventListener(type: 'controllerchange', listener: () => void): void
}

/**
 * keepUpToDate reloads the page once a newer release has taken over: at once
 * while the page is off screen; otherwise the next time the person comes back
 * to it, so a half-typed entry is never swapped away under them. The first
 * worker taking control on a first visit is not a new release, so it reloads
 * nothing. Each return to the screen asks for a newer release through check;
 * a check that fails, as one does with no signal, is ignored.
 */
export function keepUpToDate(
  workers: Workers,
  doc: Document,
  win: Window,
  check: () => Promise<unknown>,
  reload: () => void,
): void {
  let controlled = workers.controller !== null && workers.controller !== undefined
  workers.addEventListener('controllerchange', () => {
    if (!controlled) {
      controlled = true
      return
    }
    if (doc.visibilityState === 'hidden') reload()
    else reloadOnReturn(doc, win, reload)
  })
  doc.addEventListener('visibilitychange', () => {
    if (doc.visibilityState === 'visible') check().catch(() => {})
  })
}
