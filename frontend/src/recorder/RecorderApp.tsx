// The phone recorder's window (Amendment 21): the desktop's own recording form,
// the phone's unsent count with Send plus the events the phone holds, which
// can be corrected until they are sent (FR-086 to FR-088).

import { useCallback, useEffect, useState } from 'react'
import { api, said, type About, type EventEntry, type State, type Waiting } from '../api'
import { AboutDialog } from '../Dialog'
import { EventRows } from '../EventRows'
import { RecordPane } from '../RecordPane'
import { useDeletion } from '../useDeletion'
import { Flash, type Said } from './Flash'
import { InstallButton } from './InstallButton'
import { SendPane } from './SendPane'
import type { InstallOffers } from './start'

/** Said when the browser will not promise to keep the phone's events (FR-081). */
export const notKept =
  'This phone may clear events that have not been sent. Sending them is how to keep them.'

/** Said when SymDiary's rules could not start, so nothing can be recorded. */
export const notStarted = (reason: unknown) => `The recorder could not start: ${said(reason)}.`

/** Said when the phone has not kept a copy of the recorder to open offline. */
export const notOffline = (reason: unknown) =>
  `This phone has not kept a copy of the recorder to open without a connection: ${said(reason)}.`

/** The phone's list asks for every event it holds. */
const everything = { from: '', to: '', definitions: [], severities: [] }

interface Props {
  /** Settles once Go has installed the facade (FR-100); refuses with the reason it could not. */
  loaded: Promise<void>
  /** Settles once the service worker that keeps the recorder offline is registered. */
  worker: Promise<unknown>
  offers: InstallOffers
}

export function RecorderApp({ loaded, worker, offers }: Props) {
  const [ready, setReady] = useState(false)
  const [state, setState] = useState<State | null>(null)
  const [waiting, setWaiting] = useState<Waiting | null>(null)
  const [events, setEvents] = useState<EventEntry[]>([])
  const [said, setSaid] = useState<Said | null>(null)
  const [problems, setProblems] = useState<string[]>([])
  const [kept, setKept] = useState(true)
  const [about, setAbout] = useState<About | null>(null)

  /** say flashes words at the foot of the screen, the one place a message shows (FR-107). */
  const say = useCallback(
    (text: string) => setSaid((before) => ({ text, count: (before?.count ?? 0) + 1 })),
    [],
  )
  const refresh = useCallback(async () => {
    setWaiting(await api.waiting(say))
    setEvents((await api.history(everything, say)) ?? [])
  }, [say])
  // A sent event is refused by the rules themselves, in the words of FR-088.
  const deletion = useDeletion(say, say, refresh)

  useEffect(() => {
    const problem = (text: string) =>
      setProblems((known) => (known.includes(text) ? known : [...known, text]))
    loaded.then(
      async () => {
        setState(await api.state(say))
        setReady(true)
        await refresh()
      },
      (reason) => problem(notStarted(reason)),
    )
    worker.catch((reason) => problem(notOffline(reason)))
    // A refusal to answer is no promise either, so it reads as not kept.
    navigator.storage?.persist?.().then(setKept, () => setKept(false))
  }, [loaded, worker, refresh, say])

  const tell = (text: string) => {
    say(text)
    void refresh()
  }

  return (
    <main className="recorder">
      {state?.problem && <p role="alert">{state.problem}</p>}
      {problems.map((text) => (
        <p role="alert" key={text}>
          {text}
        </p>
      ))}
      {!kept && <p role="status">{notKept}</p>}
      <InstallButton offers={offers} />
      <RecordPane severities={state?.severities ?? []} refused={say} tell={tell} ready={ready} />
      <SendPane waiting={waiting} refused={say} tell={tell} changed={() => void refresh()} />
      <section className="pane" aria-labelledby="held-title">
        <h1 id="held-title">On this phone</h1>
        <EventRows events={events} severities={state?.severities ?? []} refused={say}
          tell={say} reload={refresh} askToDelete={(ids) => void deletion.ask(ids)} />
      </section>
      {/* The desktop's own About, naming the works this build ships (FR-106). */}
      <div className="button-row">
        <button type="button" disabled={!ready}
          onClick={() => void api.about(say).then((found) => found && setAbout(found))}>
          About
        </button>
      </div>
      {/* Before the dialogs, so an open dialog sits above it (FR-107). */}
      <Flash said={said} />
      {about && <AboutDialog about={about} onClose={() => setAbout(null)} />}
      {deletion.dialog}
    </main>
  )
}
