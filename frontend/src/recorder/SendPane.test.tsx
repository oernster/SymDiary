import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { installBridge } from '../bridge-fake'
import { SendPane, againWarning } from './SendPane'

const parcel = { name: 'SymDiary-phone-2026-11-16-1712.txt', text: '{}', events: [4, 5] }
const twoWaiting = { count: 2, oldest: '16 Nov 2026 08:05', hasBatch: true }

/** shareAnswers sets what the phone's share sheet does when asked. */
function shareAnswers(outcome: Promise<void>) {
  const share = vi.fn(() => outcome)
  Object.assign(navigator, { share, canShare: () => true })
  return share
}

function renderPane(waiting = twoWaiting) {
  const refused = vi.fn()
  const tell = vi.fn()
  const changed = vi.fn()
  render(<SendPane waiting={waiting} refused={refused} tell={tell} changed={changed} />)
  return { refused, tell, changed }
}

describe('sending to the desktop', () => {
  it('names the unsent count and the oldest (FR-082)', () => {
    installBridge()
    renderPane()
    expect(screen.getByText('2 not yet sent, oldest 16 Nov 2026 08:05.')).toBeInTheDocument()
  })

  it('offers no send with nothing waiting, nor a repeat before a first send (FR-092, FR-097)', () => {
    installBridge()
    renderPane({ count: 0, oldest: '', hasBatch: false })
    expect(screen.getByRole('button', { name: 'Send' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Send again' })).toBeDisabled()
  })

  it('shares a .txt file and marks it sent once the share completes (FR-089, FR-090)', async () => {
    const bridge = installBridge({
      Prepare: vi.fn(() => Promise.resolve(parcel)),
      Confirm: vi.fn(() => Promise.resolve()),
    })
    const share = shareAnswers(Promise.resolve())
    const { tell, changed } = renderPane()
    fireEvent.click(screen.getByRole('button', { name: 'Send' }))
    await waitFor(() => expect(bridge.Confirm).toHaveBeenCalledWith([4, 5]))
    const shared = (share.mock.calls[0] as unknown as [ShareData])[0].files?.[0]
    expect(shared?.name).toBe(parcel.name)
    expect(shared?.type).toBe('text/plain')
    expect(tell).toHaveBeenCalledWith(`Sent 2 to the desktop as ${parcel.name}.`)
    expect(changed).toHaveBeenCalled()
  })

  it('marks nothing when the share sheet is dismissed (FR-091)', async () => {
    const bridge = installBridge({ Prepare: vi.fn(() => Promise.resolve(parcel)) })
    shareAnswers(Promise.reject(new DOMException('Share canceled', 'AbortError')))
    const { refused } = renderPane()
    fireEvent.click(screen.getByRole('button', { name: 'Send' }))
    await waitFor(() => expect(bridge.Prepare).toHaveBeenCalled())
    expect(bridge.Confirm).not.toHaveBeenCalled()
    expect(refused).not.toHaveBeenCalled()
  })

  it('says so when a share fails for another reason', async () => {
    installBridge({ Prepare: vi.fn(() => Promise.resolve(parcel)) })
    shareAnswers(Promise.reject(new DOMException('Permission denied', 'NotAllowedError')))
    const { refused } = renderPane()
    fireEvent.click(screen.getByRole('button', { name: 'Send' }))
    await waitFor(() => expect(refused).toHaveBeenCalled())
  })

  it('asks before sending the last batch again, naming the risk (FR-095, FR-096)', async () => {
    const bridge = installBridge({
      Again: vi.fn(() => Promise.resolve(parcel)),
      Confirm: vi.fn(() => Promise.resolve()),
    })
    shareAnswers(Promise.resolve())
    renderPane()
    fireEvent.click(screen.getByRole('button', { name: 'Send again' }))
    expect(screen.getByText(againWarning)).toBeInTheDocument()
    expect(bridge.Again).not.toHaveBeenCalled()
    const buttons = screen.getAllByRole('button', { name: 'Send again' })
    fireEvent.click(buttons[buttons.length - 1])
    await waitFor(() => expect(bridge.Confirm).toHaveBeenCalledWith([4, 5]))
  })
})
