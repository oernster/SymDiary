// The last appointment's own rules (FR-046). Pure, with no DOM and no React, so
// every rule is reachable from a test by calling it; the receipt pane drives it.
//
// An appointment is a date boundary and nothing more. It is not an event and it
// is not part of the record: it lives in the window's own storage beside the
// theme, so the record and its export stay the user's observations alone. The
// sheet never mentions it; a range chosen this way prints exactly as the same
// dates chosen by hand.

/** Where the last appointment is kept, in the window's own storage. */
export const appointmentKey = 'symdiary.lastAppointment'

/** RangeChoice is how the receipt's range was chosen. */
export type RangeChoice = 'since' | 'custom'

/** The words each choice is offered under. */
export const rangeLabels: Record<RangeChoice, string> = {
  since: 'Since last appointment',
  custom: 'Custom dates',
}

/** The shape of an ISO date as a date field gives it. */
const isoDate = /^\d{4}-\d{2}-\d{2}$/

/**
 * The appointment a stored value means, given today; '' when there is none.
 *
 * An appointment is today or earlier. Anything else is none at all: a value
 * from a future version, a truncated write, a key edited by hand or a date
 * still to come must not leave the pane offering a range that runs backwards.
 * ISO dates of one shape order as strings do, so no date is parsed to compare.
 */
export function heldAppointment(raw: string | null, today: string): string {
  if (raw === null || !isoDate.test(raw) || raw > today) {
    return ''
  }
  return raw
}

/** The range "Since last appointment" means: that day to today, both included. */
export function sinceAppointment(appointment: string, today: string): [string, string] {
  return [appointment, today]
}
