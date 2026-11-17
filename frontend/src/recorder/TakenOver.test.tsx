import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import { TakenOver, takenOver } from './TakenOver'

describe('a copy that a newer one took over (FR-103)', () => {
  it('says it has paused and that it starts again by itself', () => {
    render(<TakenOver reload={vi.fn()} />)
    expect(screen.getByRole('alert')).toHaveTextContent(takenOver)
  })

  it('reloads the page when Reload is tapped', () => {
    const reload = vi.fn()
    render(<TakenOver reload={reload} />)
    fireEvent.click(screen.getByRole('button', { name: 'Reload' }))
    expect(reload).toHaveBeenCalledTimes(1)
  })
})
