// The button that installs the recorder to the home screen, shown only while
// Chrome is offering to.

import { useEffect, useState } from 'react'
import { said } from '../api'
import type { InstallOffer, InstallOffers } from './start'

interface Props {
  offers: InstallOffers
}

export function InstallButton({ offers }: Props) {
  const [offer, setOffer] = useState<InstallOffer | null>(null)
  const [failed, setFailed] = useState('')

  useEffect(() => offers.watch(setOffer), [offers])

  const install = async () => {
    if (!offer) return
    setOffer(null) // Chrome answers each offer once
    try {
      await offer.prompt()
    } catch (reason) {
      setFailed(`The browser would not install SymDiary: ${said(reason)}.`)
    }
  }

  return (
    <>
      {failed && <p role="alert">{failed}</p>}
      {offer && (
        <button type="button" className="primary" onClick={() => void install()}>
          Install SymDiary
        </button>
      )}
    </>
  )
}
