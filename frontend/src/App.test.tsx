import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { App } from './App'
import { aState, anImport, installBridge, noWindowShown } from './bridge-fake'

describe('the shell', () => {
  it('opens on the recording form and moves between the screens', async () => {
    installBridge()
    render(<App />)
    expect(await screen.findByRole('heading', { name: 'Record a symptom' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /History/ }))
    expect(await screen.findByRole('heading', { name: 'History' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Receipt/ }))
    expect(await screen.findByRole('heading', { name: 'Symptom record' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Symptoms/ }))
    expect(await screen.findByRole('heading', { name: 'Symptoms' })).toBeInTheDocument()
  })

  it('shows the problem when the record could not be opened', async () => {
    installBridge({
      State: vi.fn(() => Promise.resolve({ ...aState, problem: 'Your record could not be opened.' })),
    })
    render(<App />)
    expect(await screen.findByRole('alert')).toHaveTextContent('Your record could not be opened.')
  })

  it('says where the record went and what an import did', async () => {
    installBridge({
      Export: vi.fn(() => Promise.resolve('C:\\Users\\Oliver\\Downloads\\record.json')),
      Import: vi.fn(() => Promise.resolve(anImport)),
    })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Export/ }))
    expect(await screen.findByText(/Your record was exported to/)).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Import/ }))
    expect(
      await screen.findByText('Imported 3 events. Skipped 1 event already in your record.'),
    ).toBeInTheDocument()
  })

  it('shows an imported event in the record already on screen', async () => {
    // Save PDF writes the record the store holds when it is pressed. An import
    // made from the band while the record was showing once left the screen on
    // the old record, so the reader saved events they had never seen.
    let imported = false
    const before = [{ kind: 'title', text: 'SYMPTOM RECORD' }]
    const after = [...before, { kind: 'heading', text: 'Headache - 1 recorded event' }]
    const bridge = installBridge({
      Receipt: vi.fn(() => Promise.resolve(imported ? after : before)),
      Import: vi.fn(() => {
        imported = true
        return Promise.resolve({ chosen: true, added: 1, skipped: 0 })
      }),
    })
    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: /Receipt/ }))
    await waitFor(() => expect(screen.getByLabelText(/^To/)).toHaveValue('2026-09-22'))
    fireEvent.click(screen.getByRole('button', { name: 'Show the record' }))
    await screen.findByLabelText('The symptom record')

    fireEvent.click(screen.getByRole('button', { name: /Import/ }))

    expect(await screen.findByText('Headache - 1 recorded event')).toBeInTheDocument()
    expect(bridge.Receipt).toHaveBeenCalledTimes(2)
  })

  it('shows an imported event in the history already on screen', async () => {
    let imported = false
    const event = { id: 9, definition: 2, symptom: 'Headache', occurredAt: '2026-09-20T08:00',
      when: '20 Sep 2026 08:00', severity: '', note: '' }
    installBridge({
      History: vi.fn(() => Promise.resolve(imported ? [event] : [])),
      Import: vi.fn(() => {
        imported = true
        return Promise.resolve({ chosen: true, added: 1, skipped: 0 })
      }),
    })
    render(<App />)
    fireEvent.click(await screen.findByRole('button', { name: /History/ }))
    expect(await screen.findByText('0 events')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: /Import/ }))

    expect(await screen.findByText('1 event')).toBeInTheDocument()
  })

  it('says nothing when a file dialog is cancelled', async () => {
    installBridge({
      Export: vi.fn(() => Promise.resolve('')),
      Import: vi.fn(() => Promise.resolve({ chosen: false, added: 0, skipped: 0 })),
    })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Export/ }))
    fireEvent.click(screen.getByRole('button', { name: /Import/ }))
    await waitFor(() => expect(screen.queryByText(/Imported/)).toBeNull())
    expect(screen.queryByText(/exported to/)).toBeNull()
  })

  it('opens the Guide and About, then closes them again', async () => {
    installBridge()
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Guide/ }))
    expect(await screen.findByRole('heading', { name: 'How SymDiary works' })).toBeInTheDocument()
    expect(screen.getByText(/It does not diagnose\./)).toBeInTheDocument()
    fireEvent.keyDown(window, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())

    fireEvent.click(screen.getByRole('button', { name: /About/ }))
    expect(await screen.findByRole('heading', { name: 'SymDiary 1.0.0' })).toBeInTheDocument()
    // The notice carries the symbol and the year, both written out.
    expect(screen.getByText('© Oliver Ernster 2026')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Close' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  })

  it('draws the band in reading order, with the separators taking no turn', async () => {
    installBridge()
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    // The ring follows the page, so this is the order a Tab press walks as well
    // as the order the eye reads. The rules between the groups are chrome: they
    // are not buttons, so they cannot appear here.
    const band = screen.getByRole('navigation')
    const labels = Array.from(band.querySelectorAll('button')).map((b) => b.textContent)
    expect(labels).toEqual([
      'Record',
      'History',
      'Receipt',
      'Symptoms',
      'Import',
      'Export',
      // The window opens dark, so the theme button offers the light it would
      // move to (FR-073). It stands alone between two rules.
      'Light mode',
      'Guide',
      'About',
      'Donate',
    ])
    expect(band.querySelectorAll('.band-separator')).toHaveLength(3)
  })

  it('dresses the window from the bar and keeps the choice', async () => {
    installBridge()
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    expect(document.documentElement.dataset.theme).toBe('dark')
    fireEvent.click(screen.getByRole('button', { name: /Light mode/ }))

    await waitFor(() => expect(document.documentElement.dataset.theme).toBe('light'))
    // The button now offers the way back, by its words and by its picture.
    expect(screen.getByRole('button', { name: /Dark mode/ })).toBeInTheDocument()
    expect(window.localStorage.getItem('symdiary.theme')).toBe('light')
  })

  it('opens every dialog on its first stop, never on the page of words', async () => {
    installBridge()
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    // The Guide's body is reachable by Tab and is not a control, so the dialog
    // opens past it, on Close.
    fireEvent.click(screen.getByRole('button', { name: /Guide/ }))
    await screen.findByRole('heading', { name: 'How SymDiary works' })
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Close' })),
    )
  })

  it('offers the donation page last in the band, saying that the browser opens', async () => {
    const bridge = installBridge({ Donate: vi.fn(() => Promise.resolve()) })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    const donate = screen.getByRole('button', { name: /Donate/ })
    // The picture says nothing on its own about leaving the application.
    expect(donate).toHaveAttribute('title', 'Buy the author a drink (opens your browser)')
    // Last in the band: it belongs to nothing on screen, so nothing else is
    // reached by accident on the way to it.
    const band = screen.getByRole('navigation')
    const buttons = Array.from(band.querySelectorAll('button'))
    expect(buttons[buttons.length - 1]).toBe(donate)

    fireEvent.click(donate)
    // The page asks for the donation page and names no address: the one home
    // for that address is Go's product package.
    await waitFor(() => expect(bridge.Donate).toHaveBeenCalledTimes(1))
    expect(bridge.Donate).toHaveBeenCalledWith()
  })

  it('says so when the donation page could not be opened', async () => {
    installBridge({ Donate: vi.fn(() => Promise.reject(new Error('the desktop refused'))) })
    render(<App />)
    await screen.findByRole('heading', { name: 'Record a symptom' })

    fireEvent.click(screen.getByRole('button', { name: /Donate/ }))
    expect(await screen.findByText('The desktop refused')).toBeInTheDocument()
  })

  it('shows a refusal and lets it be dismissed', async () => {
    render(<App />)
    expect(await screen.findByText(noWindowShown)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Dismiss' }))
    expect(screen.queryByText(noWindowShown)).toBeNull()
  })
})
