import { describe, expect, it, vi } from 'vitest'
import { goRules, goSupport, listenForInstall, lockName, onlyCopy, startGo, type RecorderWindow } from './start'

/** The smallest module WebAssembly will compile: the magic number and version 1. */
const emptyModule = btoa(String.fromCharCode(0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00))

/** locksAnswering is a lock manager whose lock is free or held elsewhere. */
function locksAnswering(free: boolean) {
  const request = vi.fn((_name: string, _options: LockOptions, granted: LockGrantedCallback<unknown>) => {
    void granted(free ? ({ name: lockName, mode: 'exclusive' } as Lock) : null)
    return Promise.resolve()
  })
  return { locks: { request } as unknown as LockManager, request }
}

/** goThat stands in for wasm_exec.js's Go, doing what the test says when run. */
function goThat(run: (win: RecorderWindow) => Promise<void>, win: RecorderWindow) {
  return class {
    importObject = {}
    run() {
      return run(win)
    }
  }
}

/** loading answers every script as loaded, after doing what the test says. */
const loading = (act: () => void) => vi.fn(async () => act())

describe('one copy at a time', () => {
  it('answers true when no other copy holds the lock, asking without waiting', async () => {
    const { locks, request } = locksAnswering(true)
    await expect(onlyCopy(locks)).resolves.toBe(true)
    expect(request).toHaveBeenCalledWith(lockName, { ifAvailable: true }, expect.any(Function))
  })

  it('answers false when another copy holds it', async () => {
    await expect(onlyCopy(locksAnswering(false).locks)).resolves.toBe(false)
  })

  it('refuses when the lock cannot be asked for', async () => {
    const locks = { request: () => Promise.reject(new Error('locks unavailable')) } as unknown as LockManager
    await expect(onlyCopy(locks)).rejects.toThrow('locks unavailable')
  })
})

describe('starting the rules', () => {
  it('loads both scripts, names the zone and settles once Go is ready (FR-100)', async () => {
    const win: RecorderWindow = {}
    const load = loading(() => {
      win.Go = goThat(async (w) => {
        w.symdiaryReady?.()
        return new Promise(() => {})
      }, win)
      win.SYMDIARY_WASM_B64 = emptyModule
    })
    await expect(startGo(win, load)).resolves.toBeUndefined()
    expect(load.mock.calls).toEqual([[goSupport], [goRules]])
    expect(win.symdiaryZone).toBe(Intl.DateTimeFormat().resolvedOptions().timeZone)
  })

  it('refuses when Go stops before it is ready', async () => {
    const win: RecorderWindow = {}
    const load = loading(() => {
      win.Go = goThat(() => Promise.resolve(), win)
      win.SYMDIARY_WASM_B64 = emptyModule
    })
    await expect(startGo(win, load)).rejects.toThrow('the rules stopped before they were ready')
  })

  it('refuses, naming the script, when one will not load', async () => {
    const load = vi.fn(() => Promise.reject(new Error(`could not load ${goSupport}`)))
    await expect(startGo({}, load)).rejects.toThrow(goSupport)
  })

  it('refuses when the support script defined no Go', async () => {
    await expect(startGo({}, loading(() => {}))).rejects.toThrow(`${goSupport} did not define Go`)
  })

  it('refuses when the rules script held no rules', async () => {
    const win: RecorderWindow = {}
    const load = loading(() => {
      win.Go = goThat(() => new Promise(() => {}), win)
    })
    await expect(startGo(win, load)).rejects.toThrow(`${goRules} held no rules`)
  })
})

describe('the offer to install', () => {
  /** offer dispatches Chrome's install offer on the window. */
  function offer() {
    const event = new Event('beforeinstallprompt', { cancelable: true })
    window.dispatchEvent(event)
    return event
  }

  it("keeps an offer made before anything watched, holding back Chrome's own banner", () => {
    const offers = listenForInstall(window)
    const made = offer()
    expect(made.defaultPrevented).toBe(true)
    const told = vi.fn()
    offers.watch(told)
    expect(told).toHaveBeenCalledWith(made)
  })

  it('withdraws the offer once installed and stops telling a watcher that left', () => {
    const offers = listenForInstall(window)
    const told = vi.fn()
    const leave = offers.watch(told)
    offer()
    window.dispatchEvent(new Event('appinstalled'))
    expect(told).toHaveBeenLastCalledWith(null)
    leave()
    told.mockClear()
    offer()
    expect(told).not.toHaveBeenCalled()
  })
})
