// The phone recorder's entry (Amendment 21). The form shows at once while Go
// loads behind it (FR-100); a newer copy takes over from an older one (FR-103).

import React from 'react'
import ReactDOM from 'react-dom/client'
import { said } from '../api'
import { RecorderApp } from './RecorderApp'
import { cannotRun, missing, type Browser } from './capabilities'
import { listenForInstall, reloadOnReturn, startGo, takeOver, type RecorderWindow } from './start'
import { TakenOver } from './TakenOver'
import { keepUpToDate } from './update'
import '../theme.css'
import '../styles.css'
import '../lists.css'
import '../dialogs.css'
import './recorder.css'

/** The service worker build-recorder.ps1 writes beside the page. */
const workerScript = 'sw.js'

// Before anything awaits, so an offer Chrome makes early is not missed.
const offers = listenForInstall(window)

/** settled marks a refusal as handled now; whoever is handed the promise later still hears it. */
function settled<T>(promise: Promise<T>): Promise<T> {
  promise.catch(() => {})
  return promise
}

/** stopped shows one sentence in place of the recorder. */
function stopped(view: ReactDOM.Root, text: string) {
  view.render(
    <main className="recorder">
      <p role="alert">{text}</p>
    </main>,
  )
}

/** reload starts the page afresh, which takes the recorder back (FR-103). */
const reload = () => window.location.reload()

/** pause stops this copy and starts it again once the person comes back to it (FR-103). */
function pause(view: ReactDOM.Root) {
  view.render(<TakenOver reload={reload} />)
  reloadOnReturn(document, window, reload)
}

async function open(root: HTMLElement) {
  const view = ReactDOM.createRoot(root)
  // Before anything starts, so nothing is recorded that this browser could not keep or send (FR-104).
  const lacking = missing(window as unknown as Browser)
  if (lacking.length > 0) {
    stopped(view, cannotRun(lacking))
    return
  }
  try {
    // The copy just opened always shows the recorder; an older one pauses (FR-103).
    await takeOver(navigator.locks, () => pause(view))
  } catch (reason) {
    // Two copies writing one store could lose events, so not knowing is a stop.
    stopped(view, `The recorder could not check whether it is already open: ${said(reason)}.`)
    return
  }
  const loaded = settled(startGo(window as unknown as RecorderWindow))
  const worker = settled(
    navigator.serviceWorker
      ? navigator.serviceWorker.register(workerScript)
      : Promise.reject(new Error('this browser keeps no service worker')),
  )
  // A refused registration is already said on screen by RecorderApp (notOffline).
  worker.then(
    (registration) =>
      keepUpToDate(navigator.serviceWorker, document, window, () => registration.update(), reload),
    () => {},
  )
  view.render(
    <React.StrictMode>
      <RecorderApp loaded={loaded} worker={worker} offers={offers} />
    </React.StrictMode>,
  )
}

const root = document.getElementById('root')
if (root) void open(root)
