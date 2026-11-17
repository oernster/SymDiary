// The events of a list, each with Edit and Delete (FR-023 to FR-026; on the
// phone FR-086 to FR-088). Shared by the history and the phone recorder.

import { useState } from 'react'
import type { EventEntry, Refused } from './api'
import { EditRow } from './EditRow'

/** Where several events can be chosen at once, as in the history. */
export interface Choosing {
  chosen: number[]
  toggle: (id: number) => void
}

interface Props {
  events: EventEntry[]
  severities: string[]
  refused: Refused
  tell: (text: string) => void
  /** Reads the list again after an edit. */
  reload: () => Promise<void>
  askToDelete: (ids: number[]) => void
  choosing?: Choosing
}

export function EventRows({ events, severities, refused, tell, reload, askToDelete, choosing }: Props) {
  const [editing, setEditing] = useState<number | null>(null)

  return (
    <ul className="events">
      {events.map((event) => (
        <li key={event.id}>
          {editing === event.id ? (
            <EditRow event={event} severities={severities} refused={refused}
              onCancel={() => setEditing(null)}
              onSaved={(saved) => {
                setEditing(null)
                tell(`Saved the ${saved.symptom} event of ${saved.when}.`)
                void reload()
              }} />
          ) : (
            <div className="event">
              {choosing && (
                <input type="checkbox" checked={choosing.chosen.includes(event.id)}
                  aria-label={`Choose the ${event.symptom} event of ${event.when}`}
                  onChange={() => choosing.toggle(event.id)} />
              )}
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
  )
}
