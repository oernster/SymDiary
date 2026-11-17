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

/** canShareAFile asks whether a file of the kind Send shares could be shared. */
function canShareAFile(browser: Browser): boolean {
  const { share, canShare } = browser.navigator
  if (!share || !canShare) return false
  try {
    return canShare({ files: [new File([''], 'SymDiary.txt', { type: shareType })] })
  } catch {
    return false
  }
}

/** Every need, in the order the recorder meets them. */
export const needs: readonly Need[] = [
  { said: 'WebAssembly, to run SymDiary’s rules', present: (b) => b.WebAssembly !== undefined },
  { said: 'IndexedDB, to keep your events on the phone', present: (b) => b.indexedDB !== undefined },
  { said: 'service workers, to open without a signal', present: (b) => b.navigator.serviceWorker !== undefined },
  { said: 'Web Locks, so only one copy writes at a time', present: (b) => b.navigator.locks !== undefined },
  { said: 'file sharing, to send your events to your computer', present: canShareAFile },
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
