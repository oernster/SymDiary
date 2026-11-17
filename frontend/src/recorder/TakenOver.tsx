// What an older copy of the recorder shows once a newer one has taken over
// (FR-103). It reloads itself when the person comes back to it
// (reloadOnReturn); the button covers the one case no event reaches, both
// copies on screen at once.

/** Why this copy stopped and how it starts again. */
export const takenOver =
  'SymDiary was opened in another window, so this one has paused to keep your events safe. It starts again by itself when you come back to it; Reload starts it now.'

interface Props {
  /** reload starts this page afresh, which takes the recorder back (FR-103). */
  reload: () => void
}

export function TakenOver({ reload }: Props) {
  return (
    <main className="recorder">
      <p role="alert">{takenOver}</p>
      <button type="button" className="primary" onClick={reload}>
        Reload
      </button>
    </main>
  )
}
