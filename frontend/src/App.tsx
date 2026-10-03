// The shell: the nav band, the pane it chooses, the status line and the
// problem banner shown when the record could not be opened (FR-062).

import { useCallback, useEffect, useRef, useState } from 'react'
import { useRing } from './useRing'
import { useTheme } from './useTheme'
import { themeLabel } from './theme'
import { api, type About, type State } from './api'
import { AboutDialog } from './Dialog'
import { GuideDialog } from './GuideDialog'
import { HistoryPane } from './HistoryPane'
import { ReceiptPane } from './ReceiptPane'
import { RecordPane } from './RecordPane'
import { SymptomsPane } from './SymptomsPane'
import { counted } from './words'
import recordIcon from './assets/icons/record.png'
import historyIcon from './assets/icons/history.png'
import receiptIcon from './assets/icons/receipt.png'
import symptomsIcon from './assets/icons/symptoms.png'
import exportIcon from './assets/icons/export.png'
import importIcon from './assets/icons/import.png'
import guideIcon from './assets/icons/help-info.png'
import lightModeIcon from './assets/icons/light-mode.png'
import darkModeIcon from './assets/icons/dark-mode.png'
import aboutIcon from './assets/icons/application-icon.png'
import donateIcon from './assets/icons/donate.png'

type Pane = 'record' | 'history' | 'receipt' | 'symptoms'

const panes: { pane: Pane; label: string; icon: string }[] = [
  { pane: 'record', label: 'Record', icon: recordIcon },
  { pane: 'history', label: 'History', icon: historyIcon },
  { pane: 'receipt', label: 'Receipt', icon: receiptIcon },
  { pane: 'symptoms', label: 'Symptoms', icon: symptomsIcon },
]

interface Message {
  tone: 'info' | 'error'
  text: string
}

interface BandButtonProps {
  label: string
  icon: string
  /** What the button does, where the picture cannot say it for itself. */
  hint?: string
  /** An extra class, for a button whose picture is not a square glyph. */
  className?: string
  pressed?: boolean
  onClick: () => void
}

function BandButton({ label, icon, hint, className, pressed, onClick }: BandButtonProps) {
  return (
    <button
      type="button"
      className={className ? `band-button ${className}` : 'band-button'}
      title={hint}
      aria-pressed={pressed}
      onClick={onClick}
    >
      <img src={icon} alt="" />
      <span>{label}</span>
    </button>
  )
}

// A beer and a coffee do not say on their own that pressing them leaves the
// application, so the tooltip does.
const donateHint = 'Buy the author a drink (opens your browser)'

// A rule between groups of band buttons: what belongs together is drawn
// together. It is chrome rather than a control, so it takes no focus, paints no
// ring and is hidden from anything reading the page aloud, where the grouping
// it draws carries no meaning.
function Separator() {
  return <span className="band-separator" aria-hidden="true" />
}

export function App() {
  const [state, setState] = useState<State | null>(null)
  const [pane, setPane] = useState<Pane>('record')
  const [message, setMessage] = useState<Message | null>(null)
  const [about, setAbout] = useState<About | null>(null)
  const [guide, setGuide] = useState(false)
  // How many times the record has changed from the band, beneath whichever pane
  // is showing. The band's Import stays usable on every pane, so a pane holding
  // what it read earlier has to be told to read again.
  const [revision, setRevision] = useState(0)

  const refused = useCallback((text: string) => setMessage({ tone: 'error', text }), [])
  const tell = useCallback((text: string) => setMessage({ tone: 'info', text }), [])

  useRing()
  const [theme, toggleTheme] = useTheme()

  // The window opens neutral: nothing focused, nothing ringed, no control
  // lit up unasked. A window is looked at before it is acted in. The sink
  // holds the focus the browser would otherwise leave on the document; it is
  // out of the tab order itself, so the first step enters the ring at its first
  // stop rather than beside the sink.
  const start = useRef<HTMLDivElement>(null)
  useEffect(() => start.current?.focus(), [])

  useEffect(() => {
    void api.state(refused).then((found) => found && setState(found))
  }, [refused])

  const exportRecord = async () => {
    const path = await api.exportRecord(refused)
    if (path) tell(`Your record was exported to ${path}.`)
  }

  const importRecord = async () => {
    const result = await api.importRecord(refused)
    if (!result?.chosen) return
    setRevision((count) => count + 1)
    const skipped =
      result.skipped > 0 ? ` Skipped ${counted(result.skipped, 'event')} already in your record.` : ''
    tell(`Imported ${counted(result.added, 'event')}.${skipped}`)
  }

  const severities = state?.severities ?? []

  return (
    <div className="shell">
      <div ref={start} className="focus-sink" tabIndex={-1} aria-hidden="true" />
      <nav className="band" aria-label={state?.name ?? 'SymDiary'}>
        <div className="band-group">
          {panes.map((entry) => (
            <BandButton key={entry.pane} label={entry.label} icon={entry.icon}
              pressed={pane === entry.pane} onClick={() => setPane(entry.pane)} />
          ))}
        </div>
        <div className="band-group">
          <BandButton label="Import" icon={importIcon} onClick={importRecord} />
          <BandButton label="Export" icon={exportIcon} onClick={exportRecord} />
          <Separator />
          <BandButton label={themeLabel(theme)}
            icon={theme === 'dark' ? lightModeIcon : darkModeIcon}
            hint={`${themeLabel(theme)} (the picture is the one you would move to)`}
            onClick={toggleTheme} />
          <Separator />
          <BandButton label="Guide" icon={guideIcon} onClick={() => setGuide(true)} />
          <BandButton label="About" icon={aboutIcon}
            onClick={() => void api.about(refused).then((found) => found && setAbout(found))} />
          <Separator />
          <BandButton label="Donate" icon={donateIcon} hint={donateHint} className="donate"
            onClick={() => void api.donate(refused)} />
        </div>
      </nav>

      {state?.problem && <p className="problem" role="alert">{state.problem}</p>}

      <div className={`status ${message?.tone ?? ''}`} role="status" aria-live="polite">
        {message && (
          <>
            <span>{message.text}</span>
            <button type="button" className="dismiss" onClick={() => setMessage(null)}>
              Dismiss
            </button>
          </>
        )}
      </div>

      <main>
        {pane === 'record' && <RecordPane severities={severities} refused={refused} tell={tell} />}
        {pane === 'history' && (
          <HistoryPane severities={severities} refused={refused} tell={tell} revision={revision} />
        )}
        {pane === 'receipt' && <ReceiptPane refused={refused} saved={tell} revision={revision} />}
        {pane === 'symptoms' && <SymptomsPane refused={refused} tell={tell} />}
      </main>

      {guide && <GuideDialog onClose={() => setGuide(false)} />}
      {about && <AboutDialog about={about} onClose={() => setAbout(null)} />}
    </div>
  )
}
