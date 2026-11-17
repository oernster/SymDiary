// The history: every event, newest first, filtered by dates, symptom and
// severity (FR-020, FR-021), with editing and deletion (FR-023 to FR-026). It
// shows events and a count; nothing else (FR-022).

import { useCallback, useEffect, useState } from 'react'
import { api, type Definition, type EventEntry, type Filter, type Refused } from './api'
import { EventRows } from './EventRows'
import { notGiven } from './EventFields'
import { useDeletion } from './useDeletion'
import { counted } from './words'

interface Props {
  severities: string[]
  refused: Refused
  tell: (text: string) => void
  /** Changes whenever the record changes from outside the pane; the list is read again. */
  revision: number
}

const noFilter: Filter = { from: '', to: '', definitions: [], severities: [] }

/** toggled adds an item to a list or takes it out. */
function toggled<T>(list: T[], item: T): T[] {
  return list.includes(item) ? list.filter((held) => held !== item) : [...list, item]
}

export function HistoryPane({ severities, refused, tell, revision }: Props) {
  const [filter, setFilter] = useState<Filter>(noFilter)
  const [symptoms, setSymptoms] = useState<Definition[]>([])
  const [events, setEvents] = useState<EventEntry[]>([])
  const [chosen, setChosen] = useState<number[]>([])

  const load = useCallback(async () => {
    const [listed, held] = await Promise.all([
      api.history(filter, refused),
      api.symptoms(refused),
    ])
    if (listed) setEvents(listed)
    if (held) setSymptoms(held)
    setChosen([])
  }, [filter, refused])

  // revision is read only to say when to load again: an import beneath the
  // pane changes what the list holds without changing the filter.
  useEffect(() => {
    void load()
  }, [load, revision])

  const deletion = useDeletion(refused, tell, load)

  return (
    <section className="pane history" aria-labelledby="history-title">
      <h1 id="history-title">History</h1>
      <fieldset className="filters">
        <legend>Show</legend>
        <label>
          From <input type="date" value={filter.from}
            onChange={(e) => setFilter({ ...filter, from: e.target.value })} />
        </label>
        <label>
          To <input type="date" value={filter.to}
            onChange={(e) => setFilter({ ...filter, to: e.target.value })} />
        </label>
        <div className="choices" role="group" aria-label="Symptoms">
          <span className="group-label">Symptoms</span>
          {symptoms.length === 0 && <span className="muted">none yet</span>}
          {symptoms.map((symptom) => (
            <label key={symptom.id} className="choice">
              <input type="checkbox" checked={filter.definitions.includes(symptom.id)}
                onChange={() => setFilter({
                  ...filter, definitions: toggled(filter.definitions, symptom.id),
                })} />
              {symptom.label}
            </label>
          ))}
        </div>
        <div className="choices" role="group" aria-label="Severities">
          <span className="group-label">Severity</span>
          {['', ...severities].map((severity) => (
            <label key={severity || 'none'} className="choice">
              <input type="checkbox" checked={filter.severities.includes(severity)}
                onChange={() => setFilter({
                  ...filter, severities: toggled(filter.severities, severity),
                })} />
              {severity || notGiven}
            </label>
          ))}
        </div>
        <button type="button" onClick={() => setFilter(noFilter)}>Clear filters</button>
      </fieldset>

      <div className="list-head">
        <p>{counted(events.length, 'event')}</p>
        <button type="button" className="danger" disabled={chosen.length === 0}
          onClick={() => void deletion.ask(chosen)}>
          Delete chosen ({chosen.length})
        </button>
      </div>

      <EventRows events={events} severities={severities} refused={refused} tell={tell}
        reload={load} askToDelete={(ids) => void deletion.ask(ids)}
        choosing={{ chosen, toggle: (id) => setChosen(toggled(chosen, id)) }} />
      {deletion.dialog}
    </section>
  )
}
