// A message flashed at the foot of the screen, wherever the page is scrolled
// (FR-107). The same words also stay at the top of the page; this copy is for
// the reader who is looking somewhere else, so it is hidden from a screen
// reader, which already hears the one at the top.

import { useEffect, useState } from 'react'

/** How long a message stays at the foot of the screen, in milliseconds. */
export const flashFor = 4000

/** One thing said, numbered so the same words said twice flash twice. */
export interface Said {
  text: string
  count: number
}

/** Flash shows the latest thing said; a new count starts it afresh. */
export function Flash({ said }: { said: Said | null }) {
  if (!said || !said.text) return null
  return <Shown key={said.count} text={said.text} />
}

/** Shown is one flash, which clears itself after flashFor. */
function Shown({ text }: { text: string }) {
  const [showing, setShowing] = useState(true)
  useEffect(() => {
    const timer = setTimeout(() => setShowing(false), flashFor)
    return () => clearTimeout(timer)
  }, [])
  if (!showing) return null
  return (
    <div className="flash" aria-hidden="true" data-testid="flash">
      {text}
    </div>
  )
}
