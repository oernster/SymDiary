import { render, screen, waitFor } from '@testing-library/react'
import { fireEvent } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { RecordPane } from './RecordPane'
import { anEvent, installBridge, severities, tired } from './bridge-fake'

/** paneWith renders the recording form over a fake facade. */
function paneWith(record: ReturnType<typeof vi.fn>) {
  const refused = vi.fn()
  const tell = vi.fn()
  const bridge = installBridge({ Record: record, Suggest: vi.fn(() => Promise.resolve([tired])) })
  render(<RecordPane severities={severities} refused={refused} tell={tell} />)
  return { bridge, refused, tell }
}

describe('the recording form', () => {
  it('records what was entered and clears afterwards', async () => {
    const record = vi.fn(() => Promise.resolve(anEvent))
    const { tell } = paneWith(record)

    fireEvent.change(screen.getByLabelText('Symptom'), { target: { value: 'Tired' } })
    fireEvent.click(screen.getByLabelText('Moderate'))
    fireEvent.change(screen.getByLabelText('Note'), { target: { value: 'Since lunch.' } })
    fireEvent.click(screen.getByRole('button', { name: 'Record now' }))

    await waitFor(() => expect(record).toHaveBeenCalled())
    expect(record).toHaveBeenCalledWith({
      symptom: 'Tired',
      severity: 'Moderate',
      note: 'Since lunch.',
      occurredAt: '',
    })
    await waitFor(() => expect(screen.getByLabelText('Symptom')).toHaveValue(''))
    expect(screen.getByLabelText('Note')).toHaveValue('')
    expect(screen.getByLabelText('Symptom')).toHaveFocus()
    expect(tell).toHaveBeenCalledWith(`Recorded Tired at ${anEvent.when}.`)
  })

  it('keeps everything entered when the event was not saved', async () => {
    const record = vi.fn(() => Promise.reject(new Error('the event was not saved: disk full')))
    const { refused, tell } = paneWith(record)

    fireEvent.change(screen.getByLabelText('Symptom'), { target: { value: 'Headache' } })
    fireEvent.change(screen.getByLabelText('Note'), { target: { value: 'Since lunch.' } })
    fireEvent.click(screen.getByRole('button', { name: 'Record now' }))

    await waitFor(() => expect(refused).toHaveBeenCalled())
    expect(screen.getByLabelText('Symptom')).toHaveValue('Headache')
    expect(screen.getByLabelText('Note')).toHaveValue('Since lunch.')
    expect(tell).not.toHaveBeenCalled()
  })

  it('offers the current time and lets another be set', async () => {
    const record = vi.fn(() => Promise.resolve(anEvent))
    paneWith(record)
    expect(screen.getByText('Now')).toBeDefined()

    fireEvent.click(screen.getByRole('button', { name: 'Set a different time' }))
    const field = await screen.findByLabelText('Time')
    expect(field).toHaveValue('2026-09-22T17:12')
    fireEvent.change(field, { target: { value: '2026-09-22T12:30' } })
    fireEvent.change(screen.getByLabelText('Symptom'), { target: { value: 'Headache' } })
    fireEvent.click(screen.getByRole('button', { name: 'Record now' }))

    await waitFor(() => expect(record).toHaveBeenCalled())
    expect(record).toHaveBeenCalledWith(expect.objectContaining({ occurredAt: '2026-09-22T12:30' }))
    // Recording clears the form, so the next one starts at the current time again.
    expect(await screen.findByText('Now')).toBeInTheDocument()
  })

  it('cannot record while the rules are still loading (FR-100)', () => {
    installBridge()
    render(<RecordPane severities={severities} refused={vi.fn()} tell={vi.fn()} ready={false} />)
    expect(screen.getByRole('button', { name: 'Record now' })).toBeDisabled()
    expect(screen.getByLabelText('Symptom')).toBeEnabled()
  })

  it('goes back to the current time when asked', async () => {
    paneWith(vi.fn())
    fireEvent.click(screen.getByRole('button', { name: 'Set a different time' }))
    await screen.findByLabelText('Time')

    fireEvent.click(screen.getByRole('button', { name: 'Use the current time' }))
    expect(screen.getByText('Now')).toBeInTheDocument()
    expect(screen.queryByLabelText('Time')).toBeNull()
  })
})
