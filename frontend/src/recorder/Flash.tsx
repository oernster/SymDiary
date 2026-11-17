// A message flashed at the foot of the screen, wherever the page is scrolled
// (FR-107). It is the only place a message is shown, so it is also the page's
// status region: the region stays while nothing shows, because a screen
// reader announces a change inside a region it already knows, not one that
// arrives already holding its words.

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
  return (
    <div role="status">
      {said?.text && <Shown key={said.count} text={said.text} />}
    </div>
  )
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
    <div className="flash" data-testid="flash">
      {text}
    </div>
  )
}
