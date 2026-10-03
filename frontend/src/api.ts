// Typed access to the Go facade.
//
// Wails injects window.go.main.App at load time. Wrapping it here keeps the
// binding shape in one file. The interfaces below restate the Go DTOs in dto.go
// by hand; tests/structural/wire_test.go compares the two field for field.
//
// Every call takes a refusal handler as its last argument and never rejects: a
// refused call hands the reason to the handler and answers null. A call without
// a handler does not compile, so no refusal can go unseen (house robustness
// rule 11).

export interface State {
  name: string
  version: string
  /** Why the record could not be opened; empty when it opened. */
  problem: string
  severities: string[]
}

export interface EventEntry {
  id: number
  definition: number
  symptom: string
  /** The occurrence time in the form a datetime-local field takes. */
  occurredAt: string
  /** The occurrence time as the history shows it. */
  when: string
  severity: string
  note: string
}

export interface Definition {
  id: number
  label: string
}

export interface RecordForm {
  symptom: string
  /** Empty means now. */
  occurredAt: string
  severity: string
  note: string
}

export interface EditForm {
  symptom: string
  /** Empty keeps the time held, so an edit cannot move it by accident. */
  occurredAt: string
  severity: string
  note: string
}

export interface Filter {
  from: string
  to: string
  definitions: number[]
  severities: string[]
}

export interface ReceiptLine {
  kind:
    | 'title'
    | 'range'
    | 'appointment'
    | 'heading'
    | 'when'
    | 'severity'
    | 'note'
    | 'provenance'
    | 'statement'
  text: string
}

export interface ImportResult {
  chosen: boolean
  added: number
  skipped: number
}

export interface Credit {
  work: string
  licence: string
  holder: string
}

export interface About {
  name: string
  version: string
  author: string
  copyright: string
  statement: string
  credits: Credit[]
}

interface Bridge {
  State(): Promise<State>
  Now(): Promise<string>
  About(): Promise<About>
  Record(form: RecordForm): Promise<EventEntry>
  Suggest(typed: string): Promise<Definition[]>
  Symptoms(): Promise<Definition[]>
  History(filter: Filter): Promise<EventEntry[]>
  Edit(id: number, form: EditForm): Promise<EventEntry>
  DeletionPrompt(ids: number[]): Promise<string>
  Delete(ids: number[]): Promise<void>
  Rename(id: number, label: string): Promise<void>
  Receipt(from: string, to: string, appointment: string): Promise<ReceiptLine[] | null>
  Export(): Promise<string>
  Import(): Promise<ImportResult>
  Donate(): Promise<void>
  SavePDF(from: string, to: string, appointment: string): Promise<string>
}

interface WailsWindow {
  go?: { main?: { App?: Bridge } }
}

/** Refused receives the reason a call did not happen, ready to show. */
export type Refused = (reason: string) => void

/** noWindow is the reason given when the page runs outside SymDiary's window. */
export const noWindow = 'SymDiary is not running: this page needs its window.'

const bridge = (): Bridge | null => (window as unknown as WailsWindow).go?.main?.App ?? null

/** sentence turns a refusal from Go into something to show: a capital first letter. */
export function sentence(reason: unknown): string {
  const text = reason instanceof Error ? reason.message : String(reason)
  return text.charAt(0).toUpperCase() + text.slice(1)
}

/** ask makes one call, answering null and telling refused when it did not happen. */
async function ask<T>(invoke: (b: Bridge) => Promise<T>, refused: Refused): Promise<T | null> {
  const b = bridge()
  if (!b) {
    refused(noWindow)
    return null
  }
  try {
    return await invoke(b)
  } catch (reason) {
    refused(sentence(reason))
    return null
  }
}

/** act makes a call that answers nothing: true when it happened, else null. */
async function act(invoke: (b: Bridge) => Promise<void>, refused: Refused): Promise<true | null> {
  const done = await ask(async (b) => {
    await invoke(b)
    return true as const
  }, refused)
  return done
}

export const api = {
  state: (refused: Refused) => ask((b) => b.State(), refused),
  now: (refused: Refused) => ask((b) => b.Now(), refused),
  about: (refused: Refused) => ask((b) => b.About(), refused),
  record: (form: RecordForm, refused: Refused) => ask((b) => b.Record(form), refused),
  suggest: (typed: string, refused: Refused) => ask((b) => b.Suggest(typed), refused),
  symptoms: (refused: Refused) => ask((b) => b.Symptoms(), refused),
  history: (filter: Filter, refused: Refused) => ask((b) => b.History(filter), refused),
  edit: (id: number, form: EditForm, refused: Refused) => ask((b) => b.Edit(id, form), refused),
  deletionPrompt: (ids: number[], refused: Refused) => ask((b) => b.DeletionPrompt(ids), refused),
  remove: (ids: number[], refused: Refused) => act((b) => b.Delete(ids), refused),
  rename: (id: number, label: string, refused: Refused) =>
    act((b) => b.Rename(id, label), refused),
  // appointment is the last appointment for the sheet to name; '' names none
  // (FR-046). The page decides whether to send it; Go refuses one the range
  // does not cover.
  receipt: (from: string, to: string, appointment: string, refused: Refused) =>
    ask(async (b) => (await b.Receipt(from, to, appointment)) ?? [], refused),
  exportRecord: (refused: Refused) => ask((b) => b.Export(), refused),
  importRecord: (refused: Refused) => ask((b) => b.Import(), refused),
  // The page asks for the donation page; it never names an address. The one
  // home for that address is Go's product package, so there is nothing here to
  // keep in step with it.
  donate: (refused: Refused) => act((b) => b.Donate(), refused),
  // The record is drawn by Go, not printed by the browser. Three engines print
  // a page three different ways; one of them draws a PDF the same way every
  // time; that is the document a reader hands to a doctor.
  savePDF: (from: string, to: string, appointment: string, refused: Refused) =>
    ask((b) => b.SavePDF(from, to, appointment), refused),
}
