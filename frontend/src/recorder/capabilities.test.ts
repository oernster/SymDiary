import { describe, expect, it } from 'vitest'
import { cannotRun, missing, needs, type Browser } from './capabilities'

/** everything is a browser offering all the recorder needs. */
function everything(): Browser {
  return {
    WebAssembly: {},
    indexedDB: {},
    navigator: { serviceWorker: {}, locks: {}, share: () => {}, canShare: () => true },
  }
}

/** Each need, with what takes it away from a browser that had everything. */
const removals: [string, (b: Browser) => void][] = [
  ['WebAssembly', (b) => delete b.WebAssembly],
  ['IndexedDB', (b) => delete b.indexedDB],
  ['service workers', (b) => delete b.navigator.serviceWorker],
  ['Web Locks', (b) => delete b.navigator.locks],
  ['file sharing', (b) => delete b.navigator.share],
]

describe('what the recorder needs from the browser (FR-104)', () => {
  it('finds nothing missing in a browser that has everything', () => {
    expect(missing(everything())).toEqual([])
  })

  it('checks for every need it names', () => {
    expect(removals).toHaveLength(needs.length)
  })

  it.each(removals)('names only %s when it is missing', (name, remove) => {
    const browser = everything()
    remove(browser)
    const lacking = missing(browser)
    expect(lacking).toHaveLength(1)
    expect(lacking[0]).toMatch(new RegExp(`^${name},`))
  })

  it('counts sharing as missing when the browser says it cannot share a file', () => {
    const browser = everything()
    browser.navigator.canShare = () => false
    expect(missing(browser)).toEqual([expect.stringMatching(/^file sharing,/)])
  })

  it('counts sharing as missing when asking throws', () => {
    const browser = everything()
    browser.navigator.canShare = () => {
      throw new TypeError('files are not supported')
    }
    expect(missing(browser)).toEqual([expect.stringMatching(/^file sharing,/)])
  })

  it('says the browser cannot run the recorder and names everything it lacks', () => {
    const lacking = missing({ navigator: {} })
    const said = cannotRun(lacking)
    expect(said).toMatch(/^This browser cannot run SymDiary’s recorder\./)
    for (const need of needs) expect(said).toContain(need.said)
  })
})
