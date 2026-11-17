import { afterEach, describe, expect, it, vi } from 'vitest'
import { keepUpToDate, type Workers } from './update'

/** workersWith is a service worker container, controlled or not, that can announce a takeover. */
function workersWith(controller: unknown) {
  let changed = () => {}
  const workers: Workers = {
    controller,
    addEventListener: (_type, listener) => {
      changed = listener
    },
  }
  return { workers, takeOver: () => changed() }
}

/** showing sets what the page reports about being on screen, announcing it when told to. */
function showing(state: DocumentVisibilityState, announce = true) {
  Object.defineProperty(document, 'visibilityState', { value: state, configurable: true })
  if (announce) document.dispatchEvent(new Event('visibilitychange'))
}

describe('the recorder keeps itself up to date (FR-105)', () => {
  afterEach(() => showing('visible', false))

  it('reloads at once when a newer release takes over while the page is off screen', () => {
    const { workers, takeOver } = workersWith({})
    const reload = vi.fn()
    keepUpToDate(workers, document, window, () => Promise.resolve(), reload)
    showing('hidden', false)
    takeOver()
    expect(reload).toHaveBeenCalledTimes(1)
  })

  it('waits until the person comes back when a newer release takes over on screen', () => {
    const { workers, takeOver } = workersWith({})
    const reload = vi.fn()
    keepUpToDate(workers, document, window, () => Promise.resolve(), reload)
    showing('visible', false)
    takeOver()
    expect(reload).not.toHaveBeenCalled()
    showing('hidden')
    showing('visible')
    expect(reload).toHaveBeenCalled()
  })

  it('reloads nothing when the first worker takes control on a first visit', () => {
    const { workers, takeOver } = workersWith(null)
    const reload = vi.fn()
    keepUpToDate(workers, document, window, () => Promise.resolve(), reload)
    showing('hidden', false)
    takeOver()
    expect(reload).not.toHaveBeenCalled()
  })

  it('asks for a newer release each time the page comes back on screen', () => {
    const check = vi.fn(() => Promise.resolve())
    keepUpToDate(workersWith({}).workers, document, window, check, vi.fn())
    showing('hidden')
    expect(check).not.toHaveBeenCalled()
    showing('visible')
    showing('hidden')
    showing('visible')
    expect(check).toHaveBeenCalledTimes(2)
  })

  it('handles a check that fails with no signal', () => {
    const failed = Promise.reject(new TypeError('Failed to fetch'))
    const handled = vi.spyOn(failed, 'catch')
    keepUpToDate(workersWith({}).workers, document, window, () => failed, vi.fn())
    showing('visible')
    expect(handled).toHaveBeenCalledTimes(1)
    failed.catch(() => {}) // so a planted regression cannot also leave the run with a stray rejection
  })
})
