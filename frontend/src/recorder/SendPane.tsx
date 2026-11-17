// Sending the phone's events to the desktop (FR-082, FR-089 to FR-097).
//
// The file goes through the phone's share sheet as .txt: Chrome on Android
// refuses to share a .json file whatever it is declared as (M-4, D-9). A batch
// is marked sent only when the share completes; a dismissed share marks
// nothing (FR-090, FR-091).

import { useState } from 'react'
import { api, type Parcel, type Refused, type Waiting } from '../api'
import { ConfirmDialog } from '../Dialog'
import { shareType } from './capabilities'

/** What a share sheet dismissed without choosing an app rejects with. */
const cancelled = 'AbortError'

/** The words of the Send again confirmation (FR-096). */
export const againWarning =
  'Send the last batch again? If any of its events has since been corrected or ' +
  'deleted on the desktop, importing this file there brings the phone’s copy back.'

interface Props {
  waiting: Waiting | null
  refused: Refused
  tell: (text: string) => void
  /** Asked after a send, so the count reads what is now true. */
  changed: () => void
}

/** share hands a parcel to the share sheet: true when it completed. */
async function share(parcel: Parcel, refused: Refused): Promise<boolean> {
  const file = new File([parcel.text], parcel.name, { type: shareType })
  if (!navigator.share || !navigator.canShare?.({ files: [file] })) {
    refused('This phone offers no way to share a file from here.')
    return false
  }
  try {
    await navigator.share({ files: [file], title: parcel.name })
    return true
  } catch (reason) {
    if (!(reason instanceof DOMException && reason.name === cancelled)) {
      refused(`The file was not shared: ${String(reason)}`)
    }
    return false
  }
}

export function SendPane({ waiting, refused, tell, changed }: Props) {
  const [asking, setAsking] = useState(false)

  const send = async (parcel: Parcel | null) => {
    if (!parcel || !(await share(parcel, refused))) return
    if (await api.confirm(parcel.events, refused)) {
      tell(`Sent ${parcel.events.length} to the desktop as ${parcel.name}.`)
    }
    changed()
  }

  const count = waiting?.count ?? 0
  return (
    <section className="pane send" aria-labelledby="send-title">
      <h1 id="send-title">Send to the desktop</h1>
      <p>
        {count === 0
          ? 'Nothing is waiting to be sent.'
          : `${count} not yet sent, oldest ${waiting?.oldest}.`}
      </p>
      <div className="button-row">
        <button type="button" className="primary" disabled={count === 0}
          onClick={async () => send(await api.prepare(refused))}>
          Send
        </button>
        <button type="button" disabled={!waiting?.hasBatch} onClick={() => setAsking(true)}>
          Send again
        </button>
      </div>
      {asking && (
        <ConfirmDialog text={againWarning} confirmLabel="Send again"
          onCancel={() => setAsking(false)}
          onConfirm={async () => {
            setAsking(false)
            await send(await api.again(refused))
          }} />
      )}
    </section>
  )
}
