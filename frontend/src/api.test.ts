import { describe, expect, it, vi } from 'vitest'
import { api, noWindow, sentence } from './api'
import { aState, installBridge } from './bridge-fake'

describe('the bridge', () => {
  it('answers null and says so when the window is not there', async () => {
    const refused = vi.fn()
    expect(await api.state(refused)).toBeNull()
    expect(refused).toHaveBeenCalledWith(noWindow)
  })

  it('hands back what the facade answered', async () => {
    installBridge()
    const refused = vi.fn()
    expect(await api.state(refused)).toEqual(aState)
    expect(refused).not.toHaveBeenCalled()
  })

  it('turns a refusal into a sentence and answers null', async () => {
    installBridge({
      Record: vi.fn(() => Promise.reject(new Error('the event was not saved: disk full'))),
    })
    const refused = vi.fn()
    const saved = await api.record(
      { symptom: 'Tired', occurredAt: '', severity: '', note: '' },
      refused,
    )
    expect(saved).toBeNull()
    expect(refused).toHaveBeenCalledWith('The event was not saved: disk full')
  })

  it('answers true for a call that answers nothing, null when refused', async () => {
    const bridge = installBridge({ Delete: vi.fn(() => Promise.resolve()) })
    const refused = vi.fn()
    expect(await api.remove([1], refused)).toBe(true)
    expect(bridge.Delete).toHaveBeenCalledWith([1])

    installBridge({ Rename: vi.fn(() => Promise.reject('no such symptom')) })
    expect(await api.rename(1, 'X', refused)).toBeNull()
    expect(refused).toHaveBeenCalledWith('No such symptom')
  })

  it('reads a receipt of nothing as an empty list', async () => {
    installBridge({ Receipt: vi.fn(() => Promise.resolve(null)) })
    expect(await api.receipt('2026-09-01', '2026-09-30', '', vi.fn())).toEqual([])
  })

  it('leaves an empty reason alone', () => {
    expect(sentence('')).toBe('')
  })
})
