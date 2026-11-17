import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, render, screen } from '@testing-library/react'
import { Flash, flashFor } from './Flash'

describe('a message flashed wherever the page is scrolled (FR-107)', () => {
  afterEach(() => vi.useRealTimers())

  it('flashes the words, then clears them', () => {
    vi.useFakeTimers()
    render(<Flash said={{ text: 'A sent event is corrected on the desktop.', count: 1 }} />)
    expect(screen.getByTestId('flash')).toHaveTextContent('A sent event is corrected on the desktop.')
    act(() => vi.advanceTimersByTime(flashFor))
    expect(screen.queryByTestId('flash')).toBeNull()
  })

  it('flashes the same words again when they are said again', () => {
    vi.useFakeTimers()
    const { rerender } = render(<Flash said={{ text: 'Refused.', count: 1 }} />)
    act(() => vi.advanceTimersByTime(flashFor))
    expect(screen.queryByTestId('flash')).toBeNull()
    rerender(<Flash said={{ text: 'Refused.', count: 2 }} />)
    expect(screen.getByTestId('flash')).toHaveTextContent('Refused.')
  })

  it('shows nothing when nothing has been said', () => {
    render(<Flash said={null} />)
    expect(screen.queryByTestId('flash')).toBeNull()
  })
})
