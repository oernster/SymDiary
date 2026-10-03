import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { HistoryPane } from './HistoryPane'
import { anEvent, installBridge, severities, tired } from './bridge-fake'

const second = { ...anEvent, id: 2, symptom: 'Headache', when: '21 Sep 2026 09:30', note: '' }

/** shown renders the history over a facade holding two events. */
function shown(extra: Parameters<typeof installBridge>[0] = {}) {
  const refused = vi.fn()
  const tell = vi.fn()
  const bridge = installBridge({
    History: vi.fn(() => Promise.resolve([anEvent, second])),
    Symptoms: vi.fn(() => Promise.resolve([tired])),
    ...extra,
  })
  render(<HistoryPane severities={severities} refused={refused} tell={tell} revision={0} />)
  return { bridge, refused, tell }
}

describe('the history', () => {
  it('shows what was recorded, with a count', async () => {
    shown()
    expect(await screen.findByText('2 events')).toBeInTheDocument()
    expect(screen.getByText(anEvent.when)).toBeInTheDocument()
    expect(screen.getByText(anEvent.note)).toBeInTheDocument()
  })

  it('narrows by dates, symptom and severity, then clears again', async () => {
    const { bridge } = shown()
    await screen.findByText('2 events')

    fireEvent.change(screen.getByLabelText(/From/), { target: { value: '2026-09-01' } })
    fireEvent.change(screen.getByLabelText(/To/), { target: { value: '2026-09-30' } })
    fireEvent.click(await screen.findByLabelText('Tired'))
    fireEvent.click(screen.getByLabelText('Moderate'))

    await waitFor(() =>
      expect(bridge.History).toHaveBeenLastCalledWith({
        from: '2026-09-01',
        to: '2026-09-30',
        definitions: [tired.id],
        severities: ['Moderate'],
      }),
    )

    fireEvent.click(screen.getByRole('button', { name: 'Clear filters' }))
    await waitFor(() =>
      expect(bridge.History).toHaveBeenLastCalledWith({
        from: '',
        to: '',
        definitions: [],
        severities: [],
      }),
    )
  })

  it('asks before deleting, naming the event; deletes only when confirmed', async () => {
    const prompt = `Delete the Tired event of ${anEvent.when}? This cannot be undone.`
    const { bridge, tell } = shown({
      DeletionPrompt: vi.fn(() => Promise.resolve(prompt)),
      Delete: vi.fn(() => Promise.resolve()),
    })
    await screen.findByText('2 events')

    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' })[0])
    expect(await screen.findByText(prompt)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(bridge.Delete).not.toHaveBeenCalled()

    fireEvent.click(screen.getAllByRole('button', { name: 'Delete' })[0])
    await screen.findByText(prompt)
    fireEvent.click(screen.getByRole('alertdialog').querySelector('.danger') as HTMLElement)
    await waitFor(() => expect(bridge.Delete).toHaveBeenCalledWith([anEvent.id]))
    expect(tell).toHaveBeenCalledWith('Deleted 1 event.')
  })

  it('counts the events chosen for a bulk deletion', async () => {
    const { bridge } = shown({ DeletionPrompt: vi.fn(() => Promise.resolve('Delete 2 events?')) })
    await screen.findByText('2 events')

    const bulk = screen.getByRole('button', { name: /Delete chosen/ })
    expect(bulk).toBeDisabled()
    fireEvent.click(screen.getByLabelText(`Choose the ${anEvent.symptom} event of ${anEvent.when}`))
    fireEvent.click(screen.getByLabelText(`Choose the ${second.symptom} event of ${second.when}`))
    fireEvent.click(screen.getByRole('button', { name: /Delete chosen \(2\)/ }))

    await waitFor(() => expect(bridge.DeletionPrompt).toHaveBeenCalledWith([anEvent.id, second.id]))
    expect(await screen.findByText('Delete 2 events?')).toBeInTheDocument()
  })

  it('opens an editor on a row and closes it again', async () => {
    const { tell } = shown({ Edit: vi.fn(() => Promise.resolve(anEvent)) })
    await screen.findByText('2 events')

    fireEvent.click(screen.getAllByRole('button', { name: 'Edit' })[0])
    expect(await screen.findByRole('form', { name: /Edit the Tired event/ })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('form')).toBeNull()

    fireEvent.click(screen.getAllByRole('button', { name: 'Edit' })[0])
    fireEvent.click(await screen.findByRole('button', { name: 'Save' }))
    await waitFor(() =>
      expect(tell).toHaveBeenCalledWith(`Saved the ${anEvent.symptom} event of ${anEvent.when}.`),
    )
  })
})
