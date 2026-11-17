import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
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
    expect(await screen.findByText(`${anEvent.when} ${anEvent.symptom}`)).toBeInTheDocument()
  })
})
