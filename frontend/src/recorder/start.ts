// Starting the phone recorder (Amendment 21): one copy at a time, Chrome's
// offer to install it, then SymDiary's own Go rules (D-10).

/** The Web Locks name every copy of the recorder asks for. */
export const lockName = 'symdiary-recorder'

/** Said by an older copy once a newer one has taken over, so only one copy writes the phone's events (FR-103). */
export const takenOver =
  'SymDiary has been opened again on this phone, so this copy has stopped. The newer one has your events. You can close this one.'

/** The scripts build-recorder.ps1 places beside the page. */
export const goSupport = 'wasm_exec.js'
export const goRules = 'symdiary-wasm.js'

/** Go's own runtime for the browser, as wasm_exec.js defines it. */
interface GoRuntime {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

/** What the page and Go leave for each other on the window (cmd/recorder/main.go). */
export interface RecorderWindow {
  Go?: new () => GoRuntime
  SYMDIARY_WASM_B64?: string
  symdiaryZone?: string
  symdiaryReady?: () => void
}

/**
 * takeOver makes this copy the one that writes the phone's events, taking the
 * recorder's lock from any copy already holding it (FR-103). It settles once
 * the lock is held; lost is called if a newer copy later takes it from this one.
 */
export function takeOver(locks: LockManager, lost: () => void): Promise<void> {
  return new Promise((held, refuse) => {
    locks
      .request(lockName, { steal: true }, () => {
        held()
        // A promise that never settles keeps the lock until the page goes.
        return new Promise<void>(() => {})
      })
      .catch((reason: unknown) => {
        // The browser answers a lock taken by another copy with AbortError.
        if (reason instanceof DOMException && reason.name === 'AbortError') lost()
        else refuse(reason)
      })
  })
}

/** loadScript adds one of the page's own scripts, settling once it has run. */
export function loadScript(src: string): Promise<void> {
  return new Promise((loaded, refuse) => {
    const script = document.createElement('script')
    script.src = src
    script.onload = () => loaded()
    script.onerror = () => refuse(new Error(`could not load ${src}`))
    document.body.append(script)
  })
}

/**
 * startGo runs SymDiary's rules, settling once Go has installed the facade
 * and refusing with the reason when it cannot (FR-100).
 */
export async function startGo(win: RecorderWindow, load = loadScript): Promise<void> {
  await load(goSupport)
  await load(goRules)
  if (!win.Go) throw new Error(`${goSupport} did not define Go`)
  if (!win.SYMDIARY_WASM_B64) throw new Error(`${goRules} held no rules`)
  const binary = Uint8Array.from(atob(win.SYMDIARY_WASM_B64), (c) => c.charCodeAt(0))
  const go = new win.Go()
  const { instance } = await WebAssembly.instantiate(binary, go.importObject)
  // Go's clock in the browser is a fixed offset from start-up, so the page names the zone.
  win.symdiaryZone = Intl.DateTimeFormat().resolvedOptions().timeZone
  return new Promise((ready, refuse) => {
    win.symdiaryReady = () => ready()
    // Go calls symdiaryReady and then waits forever; finishing first means it never got there.
    go.run(instance).then(() => refuse(new Error('the rules stopped before they were ready')), refuse)
  })
}

/** Chrome's offer to install the page, as beforeinstallprompt carries it. */
export interface InstallOffer extends Event {
  prompt(): Promise<void>
}

/** Told the current offer whenever it changes; null once there is none. */
export type Offered = (offer: InstallOffer | null) => void

/** The install offers a component can watch. */
export interface InstallOffers {
  watch(tell: Offered): () => void
}

/**
 * listenForInstall catches Chrome's offer to install the moment it is made,
 * ahead of React, so an early offer is not missed. Chrome 157's menu offers
 * no reliable way to install the page, so this is the way in (measured).
 */
export function listenForInstall(win: Window): InstallOffers {
  let latest: InstallOffer | null = null
  let tell: Offered = () => {}
  win.addEventListener('beforeinstallprompt', (event) => {
    event.preventDefault()
    latest = event as InstallOffer
    tell(latest)
  })
  win.addEventListener('appinstalled', () => {
    latest = null
    tell(null)
  })
  return {
    watch(next) {
      tell = next
      next(latest)
      return () => {
        tell = () => {}
      }
    },
  }
}
