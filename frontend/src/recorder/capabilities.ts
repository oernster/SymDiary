// What the phone recorder needs from the browser it opens in (FR-104,
// NFR-PORT-001). No browser is named as a requirement: the page checks for each
// thing it uses on the phone it runs on, naming whatever is missing before
// anything is recorded that could not be kept or sent.

/** The share sheet takes plain text; the contents are an export all the same (D-9). */
export const shareType = 'text/plain'

/** The parts of a browser the recorder relies on, each absent where it is not offered. */
export interface Browser {
  WebAssembly?: unknown
  indexedDB?: unknown
  navigator: {
    serviceWorker?: unknown
    locks?: unknown
    share?: unknown
    canShare?: (data: ShareData) => boolean
  }
}

/** One thing the recorder needs, named as a reader is told when it is missing. */
interface Need {
  said: string
  present: (browser: Browser) => boolean
}

/**
 * canShareFiles answers whether the browser offers file sharing at all. It
 * never asks canShare about a made-up file: on a phone where Send works, that
 * answer came back false and stopped the recorder (Amendment 26). Send asks
 * about the real file when it is pressed and says so if it cannot be shared.
 */
function canShareFiles(browser: Browser): boolean {
  return browser.navigator.share !== undefined && browser.navigator.canShare !== undefined
}

/** Every need, in the order the recorder meets them. */
export const needs: readonly Need[] = [
  { said: 'WebAssembly, to run SymDiary’s rules', present: (b) => b.WebAssembly !== undefined },
  { said: 'IndexedDB, to keep your events on the phone', present: (b) => b.indexedDB !== undefined },
  { said: 'service workers, to open without a signal', present: (b) => b.navigator.serviceWorker !== undefined },
  { said: 'Web Locks, so only one copy writes at a time', present: (b) => b.navigator.locks !== undefined },
  { said: 'file sharing, to send your events to your computer', present: canShareFiles },
]

/** missing answers, in plain words, everything this browser lacks; empty when it lacks nothing. */
export function missing(browser: Browser): string[] {
  return needs.filter((need) => !need.present(browser)).map((need) => need.said)
}

/** cannotRun is what the recorder says, in place of itself, in a browser that lacks something. */
export function cannotRun(lacking: string[]): string {
  return (
    `This browser cannot run SymDiary’s recorder. It lacks ${lacking.join('; ')}. ` +
    'Another browser on this phone may have what it needs: Chrome does.'
  )
}
