// The fields an event is made of, shared by the recording form and the edit
// row so the two cannot drift apart. The time is left to each caller, since
// recording defaults it to now while an edit shows the time held.

import type { Ref } from 'react'
import type { Refused } from './api'
import { SymptomInput } from './SymptomInput'

export interface EventValues {
  symptom: string
  severity: string
  note: string
}

export const emptyValues: EventValues = { symptom: '', severity: '', note: '' }

/** notGiven is how the absence of a severity is offered (FR-004). */
export const notGiven = 'Not given'

interface Props {
  idPrefix: string
  values: EventValues
  onChange: (values: EventValues) => void
  severities: string[]
  refused: Refused
  symptomRef?: Ref<HTMLInputElement>
}

export function EventFields({ idPrefix, values, onChange, severities, refused, symptomRef }: Props) {
  const set = (change: Partial<EventValues>) => onChange({ ...values, ...change })
  return (
    <>
      <div className="field">
        <label htmlFor={`${idPrefix}-symptom`}>Symptom</label>
        <SymptomInput
          id={`${idPrefix}-symptom`}
          value={values.symptom}
          onChange={(symptom) => set({ symptom })}
          refused={refused}
          inputRef={symptomRef}
        />
      </div>
      <fieldset className="field radio-row">
        <legend>Severity</legend>
        {['', ...severities].map((severity) => (
          <label key={severity || 'none'} className="choice">
            <input
              type="radio"
              name={`${idPrefix}-severity`}
              value={severity}
              checked={values.severity === severity}
              onChange={() => set({ severity })}
            />
            {severity || notGiven}
          </label>
        ))}
      </fieldset>
      <div className="field">
        <label htmlFor={`${idPrefix}-note`}>Note</label>
        <textarea
          id={`${idPrefix}-note`}
          rows={3}
          value={values.note}
          onChange={(event) => set({ note: event.target.value })}
        />
      </div>
    </>
  )
}
