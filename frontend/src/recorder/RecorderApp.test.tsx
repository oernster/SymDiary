import { afterEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { sentence } from '../api'
import { aState, anEvent, installBridge } from '../bridge-fake'
import { RecorderApp, notKept, notOffline, notStarted } from './RecorderApp'
import type { InstallOffers } from './start'

/** opened is a load the test settles itself, standing in for Go starting. */
function opened() {
  let settle = () => {}
  const loaded = new Promise<void>((resolve) => {
    settle = resolve
  })
  return { loaded, settle }
}

/** noOffers is a phone where Chrome never offers to install. */
const noOffers: InstallOffers = { watch: () => () => {} }

/** storageAnswers sets what the browser says when asked to keep the phone's events. */
function storageAnswers(persist: () => Promise<boolean>) {
  Object.defineProperty(navigator, 'storage', { configurable: true, value: { persist } })
}

function renderApp(loaded: Promise<void> = Promise.resolve(), worker: Promise<unknown> = Promise.resolve()) {
  render(<RecorderApp loaded={loaded} worker={worker} offers={noOffers} />)
}

afterEach(() => {
  Object.defineProperty(navigator, 'storage', { configurable: true, value: undefined })
})

describe('the phone recorder', () => {
  it('opens About and closes it again (FR-106)', async () => {
    installBridge()
    renderApp()
    const about = screen.getByRole('button', { name: 'About' })
    await waitFor(() => expect(about).toBeEnabled())
    fireEvent.click(about)
    expect(await screen.findByRole('heading', { name: 'SymDiary 1.0.0' })).toBeInTheDocument()
    expect(screen.getByText('© Oliver Ernster 2026')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('takes typing but cannot record until Go has loaded (FR-100)', async () => {
    installBridge()
    const { loaded, settle } = opened()
    renderApp(loaded)
    expect(screen.getByRole('button', { name: 'Record now' })).toBeDisabled()
    expect(screen.getByLabelText('Symptom')).toBeEnabled()
    settle()
    await waitFor(() => expect(screen.getByRole('button', { name: 'Record now' })).toBeEnabled())
  })

  it('says why the rules could not start, leaving Record disabled', async () => {
    installBridge()
    renderApp(Promise.reject(new Error('could not load wasm_exec.js')))
    expect(await screen.findByText(notStarted(new Error('could not load wasm_exec.js')))).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Record now' })).toBeDisabled()
  })

  it('says so when the phone kept no copy to open offline', async () => {
    installBridge()
    const reason = new Error('the worker script was not found')
    renderApp(Promise.resolve(), Promise.reject(reason))
    expect(await screen.findByText(notOffline(reason))).toBeInTheDocument()
  })

  it('shows the problem the facade reports', async () => {
    installBridge({ State: vi.fn(() => Promise.resolve({ ...aState, problem: 'Storage is unavailable.' })) })
    renderApp()
    expect(await screen.findByText('Storage is unavailable.')).toBeInTheDocument()
  })

  it('asks for persistent storage when it opens (FR-080)', async () => {
    installBridge()
    const persist = vi.fn(() => Promise.resolve(true))
    storageAnswers(persist)
    renderApp()
    await waitFor(() => expect(persist).toHaveBeenCalled())
  })

  it('says the events may be cleared when the browser will not keep them (FR-081)', async () => {
    installBridge()
    storageAnswers(() => Promise.resolve(false))
    renderApp()
    expect(await screen.findByText(notKept)).toBeInTheDocument()
  })

  it('says the same when the browser refuses to answer', async () => {
    installBridge()
    storageAnswers(() => Promise.reject(new TypeError('not allowed here')))
    renderApp()
    expect(await screen.findByText(notKept)).toBeInTheDocument()
  })

  it('says nothing about clearing when the browser will keep them', async () => {
    const bridge = installBridge()
    storageAnswers(() => Promise.resolve(true))
    renderApp()
    await waitFor(() => expect(bridge.Waiting).toHaveBeenCalled())
    expect(screen.queryByText(notKept)).toBeNull()
  })

  it('lists the events the phone holds', async () => {
    installBridge({ History: vi.fn(() => Promise.resolve([anEvent])) })
    renderApp()
    expect(await screen.findByText(anEvent.when)).toBeInTheDocument()
    expect(screen.getByText(anEvent.note)).toBeInTheDocument()
  })
})

describe('correcting on the phone', () => {
  /** alreadySent is Go's refusal for an event the desktop holds (application.ErrAlreadySent). */
  const alreadySent = 'this event has been sent to the desktop, so it is corrected there'

  /** opened presses Edit on the only event, answering its editor; the recording form is beside it. */
  async function opened(symptom: string) {
    fireEvent.click(await screen.findByRole('button', { name: 'Edit' }))
    return screen.findByRole('form', { name: `Edit the ${symptom} event` })
  }

  it('edits an unsent event (FR-086)', async () => {
    const corrected = { ...anEvent, note: 'Since lunch.' }
    const bridge = installBridge({
      History: vi.fn(() => Promise.resolve([anEvent])),
      Edit: vi.fn(() => Promise.resolve(corrected)),
    })
    renderApp()
    const editor = await opened(anEvent.symptom)
    fireEvent.change(within(editor).getByLabelText('Note'), { target: { value: 'Since lunch.' } })
    fireEvent.click(within(editor).getByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(bridge.Edit).toHaveBeenCalledWith(anEvent.id, expect.objectContaining({ note: 'Since lunch.' })),
    )
    expect(await screen.findByText(`Saved the ${anEvent.symptom} event of ${anEvent.when}.`)).toBeInTheDocument()
  })

  it('asks before deleting an unsent event, naming it (FR-087)', async () => {
    const prompt = `Delete the Tired event of ${anEvent.when}? This cannot be undone.`
    const bridge = installBridge({
      History: vi.fn(() => Promise.resolve([anEvent])),
      DeletionPrompt: vi.fn(() => Promise.resolve(prompt)),
      Delete: vi.fn(() => Promise.resolve()),
    })
    renderApp()
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(await screen.findByText(prompt)).toBeInTheDocument()
    expect(bridge.Delete).not.toHaveBeenCalled()
    fireEvent.click(screen.getByRole('alertdialog').querySelector('.danger') as HTMLElement)
    await waitFor(() => expect(bridge.Delete).toHaveBeenCalledWith([anEvent.id]))
    expect(await screen.findByText('Deleted 1 event.')).toBeInTheDocument()
  })

  it('refuses to delete a sent event before asking, saying where it is corrected (FR-088)', async () => {
    installBridge({
      History: vi.fn(() => Promise.resolve([anEvent])),
      DeletionPrompt: vi.fn(() => Promise.reject(new Error(alreadySent))),
    })
    renderApp()
    fireEvent.click(await screen.findByRole('button', { name: 'Delete' }))
    expect(await screen.findByText(sentence(alreadySent))).toBeInTheDocument()
    expect(screen.queryByRole('alertdialog')).toBeNull()
  })

  it('refuses to edit a sent event, keeping what was typed (FR-088)', async () => {
    installBridge({
      History: vi.fn(() => Promise.resolve([anEvent])),
      Edit: vi.fn(() => Promise.reject(new Error(alreadySent))),
    })
    renderApp()
    const editor = await opened(anEvent.symptom)
    fireEvent.change(within(editor).getByLabelText('Note'), { target: { value: 'Since lunch.' } })
    fireEvent.click(within(editor).getByRole('button', { name: 'Save' }))
    expect(await screen.findByText(sentence(alreadySent))).toBeInTheDocument()
    expect(within(editor).getByLabelText('Note')).toHaveValue('Since lunch.')
  })
})
