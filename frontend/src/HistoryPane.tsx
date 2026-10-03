// The history: every event, newest first, filtered by dates, symptom and
// severity (FR-020, FR-021), with editing and deletion (FR-023 to FR-026). It
// shows events and a count; nothing else (FR-022).

import { useCallback, useEffect, useState } from 'react'
import { api, type Definition, type EventEntry, type Filter, type Refused } from './api'
import { ConfirmDialog } from './Dialog'
import { EditRow } from './EditRow'
import { notGiven } from './EventFields'
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

interface Pending {
  prompt: string
  ids: number[]
}

export function HistoryPane({ severities, refused, tell, revision }: Props) {
  const [filter, setFilter] = useState<Filter>(noFilter)
  const [symptoms, setSymptoms] = useState<Definition[]>([])
  const [events, setEvents] = useState<EventEntry[]>([])
  const [chosen, setChosen] = useState<number[]>([])
  const [editing, setEditing] = useState<number | null>(null)
  const [pending, setPending] = useState<Pending | null>(null)

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

  const askToDelete = async (ids: number[]) => {
    const prompt = await api.deletionPrompt(ids, refused)
    if (prompt) setPending({ prompt, ids })
  }

  const confirmDelete = async () => {
    if (!pending) return
    const ids = pending.ids
    setPending(null)
    if (await api.remove(ids, refused)) {
      tell(`Deleted ${counted(ids.length, 'event')}.`)
      await load()
    }
  }

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
          onClick={() => askToDelete(chosen)}>
          Delete chosen ({chosen.length})
        </button>
      </div>

      <ul className="events">
        {events.map((event) => (
          <li key={event.id}>
            {editing === event.id ? (
              <EditRow event={event} severities={severities} refused={refused}
                onCancel={() => setEditing(null)}
                onSaved={(saved) => {
                  setEditing(null)
                  tell(`Saved the ${saved.symptom} event of ${saved.when}.`)
                  void load()
                }} />
            ) : (
              <div className="event">
                <input type="checkbox" checked={chosen.includes(event.id)}
                  aria-label={`Choose the ${event.symptom} event of ${event.when}`}
                  onChange={() => setChosen(toggled(chosen, event.id))} />
                <span className="when">{event.when}</span>
                <span className="symptom">{event.symptom}</span>
                <span className="severity">{event.severity}</span>
                <span className="note">{event.note}</span>
                <span className="row-actions">
                  <button type="button" onClick={() => setEditing(event.id)}>Edit</button>
                  <button type="button" onClick={() => askToDelete([event.id])}>Delete</button>
                </span>
              </div>
            )}
          </li>
        ))}
      </ul>

      {pending && (
        <ConfirmDialog text={pending.prompt} confirmLabel="Delete"
          onConfirm={confirmDelete} onCancel={() => setPending(null)} />
      )}
    </section>
  )
}
