// The recording form, where SymDiary opens (FR-001 to FR-009).

import { useRef, useState, type FormEvent } from 'react'
import { api, type Refused } from './api'
import { EventFields, emptyValues, type EventValues } from './EventFields'

interface Props {
  severities: string[]
  refused: Refused
  tell: (text: string) => void
  /**
   * False while the rules are still loading: the form takes typing but cannot
   * record yet (FR-100). The desktop's rules are there from the start.
   */
  ready?: boolean
}

export function RecordPane({ severities, refused, tell, ready = true }: Props) {
  const [values, setValues] = useState<EventValues>(emptyValues)
  // An empty time means now, read by the backend at the moment of recording.
  const [time, setTime] = useState('')
  const [latest, setLatest] = useState('')
  const symptomRef = useRef<HTMLInputElement>(null)

  const chooseTime = async () => {
    const now = await api.now(refused)
    if (now) {
      setTime(now)
      setLatest(now)
    }
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const saved = await api.record({ ...values, occurredAt: time }, refused)
    if (!saved) return // everything entered stays on the form (FR-009)
    tell(`Recorded ${saved.symptom} at ${saved.when}.`)
    setValues(emptyValues)
    setTime('')
    symptomRef.current?.focus()
  }

  return (
    <form className="pane record" onSubmit={submit} aria-labelledby="record-title">
      <h1 id="record-title">Record a symptom</h1>
      <EventFields
        idPrefix="record"
        values={values}
        onChange={setValues}
        severities={severities}
        refused={refused}
        symptomRef={symptomRef}
      />
      <div className="field">
        <span className="label" id="record-time-label">
          Time
        </span>
        {time === '' ? (
          <div className="time-row">
            <span>Now</span>
            <button type="button" onClick={chooseTime}>
              Set a different time
            </button>
          </div>
        ) : (
          <div className="time-row">
            <input
              type="datetime-local"
              aria-labelledby="record-time-label"
              value={time}
              max={latest}
              onChange={(event) => setTime(event.target.value)}
            />
            <button type="button" onClick={() => setTime('')}>
              Use the current time
            </button>
          </div>
        )}
      </div>
      <button type="submit" className="primary" disabled={!ready}>
        Record now
      </button>
    </form>
  )
}
