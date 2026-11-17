import { describe, expect, it, vi } from 'vitest'
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { InstallButton } from './InstallButton'
import type { InstallOffer, InstallOffers, Offered } from './start'

/** offering is a phone whose Chrome offers to install, answering the prompt as told. */
function offering(prompt: () => Promise<void>) {
  let tell: Offered = () => {}
  const offer = Object.assign(new Event('beforeinstallprompt'), { prompt: vi.fn(prompt) }) as InstallOffer
  const offers: InstallOffers = {
    watch(next) {
      tell = next
      return () => {}
    },
  }
  return { offers, offer, make: () => act(() => tell(offer)) }
}

describe('installing the recorder', () => {
  it('shows no button until Chrome offers', () => {
    const { offers, make } = offering(() => Promise.resolve())
    render(<InstallButton offers={offers} />)
    expect(screen.queryByRole('button', { name: 'Install SymDiary' })).toBeNull()
    make()
    expect(screen.getByRole('button', { name: 'Install SymDiary' })).toBeInTheDocument()
  })

  it('asks Chrome once, then hides the button', async () => {
    const { offers, offer, make } = offering(() => Promise.resolve())
    render(<InstallButton offers={offers} />)
    make()
    fireEvent.click(screen.getByRole('button', { name: 'Install SymDiary' }))
    await waitFor(() => expect(offer.prompt).toHaveBeenCalledTimes(1))
    expect(screen.queryByRole('button', { name: 'Install SymDiary' })).toBeNull()
  })

  it('says why when the browser will not install', async () => {
    const { offers, make } = offering(() => Promise.reject(new Error('the prompt was already used')))
    render(<InstallButton offers={offers} />)
    make()
    fireEvent.click(screen.getByRole('button', { name: 'Install SymDiary' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'The browser would not install SymDiary: the prompt was already used.',
    )
  })
})
