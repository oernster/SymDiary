// The receipt: the factual record for a date range, shown here and saved as a
// PDF the reader keeps (FR-040 to FR-044). Every word on it comes from the
// backend's receipt; the page only lays it out on screen, while the document
// itself is drawn by Go so that it reads the same on every desktop.
//
// The range can start at the last appointment the user marked (FR-046). That
// only fills in the two dates: the record is asked for exactly as it is for
// dates chosen by hand, so the sheet cannot tell the two apart.

import { useEffect, useState } from 'react'
import { api, type ReceiptLine, type Refused } from './api'
import {
  appointmentKey, heldAppointment, rangeLabels, sinceAppointment, type RangeChoice,
} from './appointment'
import crest from './assets/icons/application-icon.png'
import { keepStored, readStored } from './storage'

interface Props {
  refused: Refused
  /** saved says where the document went, once it has gone there. */
  saved: (message: string) => void
}

/** defaultDays is how far back the range starts when the pane opens. */
const defaultDays = 30

/** daysBefore answers the ISO date a number of days before an ISO date. */
export function daysBefore(iso: string, days: number): string {
  const [year, month, day] = iso.split('-').map(Number)
  const date = new Date(Date.UTC(year, month - 1, day - days))
  return date.toISOString().slice(0, 10)
}

/**
 * Whether a line is the sheet's letterhead: the first line, a provenance one.
 * The position is part of the test and not only the kind, so a mark can never
 * arrive anywhere other than the head of the sheet.
 */
function isLetterhead(line: ReceiptLine, index: number): boolean {
  return index === 0 && line.kind === 'provenance'
}

export function ReceiptPane({ refused, saved }: Props) {
  const [today, setToday] = useState('')
  const [appointment, setAppointment] = useState('')
  const [choice, setChoice] = useState<RangeChoice>('custom')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [lines, setLines] = useState<ReceiptLine[]>([])

  /** chooseSince sets the range to the appointment's day through today. */
  const chooseSince = (held: string, now: string) => {
    const [start, end] = sinceAppointment(held, now)
    setChoice('since')
    setFrom(start)
    setTo(end)
  }

  useEffect(() => {
    void api.now(refused).then((now) => {
      if (!now) return
      const date = now.slice(0, 10)
      const held = heldAppointment(readStored(appointmentKey), date)
      setToday(date)
      setAppointment(held)
      if (held) {
        chooseSince(held, date)
      } else {
        setTo(date)
        setFrom(daysBefore(date, defaultDays))
      }
    })
  }, [refused])

  /**
   * markAppointment keeps the date marked, else forgets it when the field is
   * cleared. A date still to come is not an appointment that has happened, so
   * it is held as none.
   */
  const markAppointment = (value: string) => {
    const held = heldAppointment(value, today)
    keepStored(appointmentKey, held || null)
    setAppointment(held)
    if (held) {
      chooseSince(held, today)
    } else {
      setChoice('custom')
    }
  }

  /** A date changed by hand is a custom range, whatever it was before. */
  const setByHand = (set: (value: string) => void) => (value: string) => {
    setChoice('custom')
    set(value)
  }

  const show = async () => {
    setLines([])
    const found = await api.receipt(from, to, refused)
    if (found) setLines(found)
  }

  const savePDF = async () => {
    const path = await api.savePDF(from, to, refused)
    // An empty path is a cancelled dialog, which is not worth announcing.
    if (path) saved(`Your symptom record was saved to ${path}.`)
  }

  return (
    <section className="pane receipt-pane" aria-labelledby="receipt-title">
      <h1 id="receipt-title">Symptom record</h1>
      <div className="receipt-controls">
        <label>
          Last appointment <input type="date" value={appointment} max={today}
            onChange={(e) => markAppointment(e.target.value)} />
        </label>
        {appointment && (
          <fieldset className="field choices">
            <legend>Range</legend>
            {(['since', 'custom'] as const).map((kind) => (
              <label key={kind} className="choice">
                <input
                  type="radio"
                  name="receipt-range"
                  value={kind}
                  checked={choice === kind}
                  onChange={() => (kind === 'since' ? chooseSince(appointment, today) : setChoice(kind))}
                />
                {rangeLabels[kind]}
              </label>
            ))}
          </fieldset>
        )}
      </div>
      <div className="receipt-controls">
        <label>
          From <input type="date" value={from} onChange={(e) => setByHand(setFrom)(e.target.value)} />
        </label>
        <label>
          To <input type="date" value={to} onChange={(e) => setByHand(setTo)(e.target.value)} />
        </label>
        <button type="button" onClick={show}>Show the record</button>
        <button type="button" className="primary" disabled={lines.length === 0}
          onClick={() => void savePDF()}>
          Save PDF
        </button>
      </div>
      {lines.length > 0 && (
        <article className="receipt" aria-label="The symptom record">
          {lines.map((line, index) => (
            <p key={index} className={`line ${line.kind}`}>
              {/*
                The mark sits with the words rather than above them, because
                the pair is the letterhead: a picture and the address it
                belongs to. It is decorative, so it carries no alt text; the
                line beside it already names the program.
              */}
              {isLetterhead(line, index) && <img className="mark" src={crest} alt="" />}
              {line.text}
            </p>
          ))}
        </article>
      )}
    </section>
  )
}
