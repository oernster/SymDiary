// Deleting events: ask first, naming them, then delete only on confirmation
// (FR-025, FR-026; on the phone FR-087). Shared by the history and the phone.

import { useState } from 'react'
import { api, type Refused } from './api'
import { ConfirmDialog } from './Dialog'
import { counted } from './words'

interface Pending {
  prompt: string
  ids: number[]
}

/**
 * useDeletion answers ask, which puts the question for some events, plus the
 * dialog to place in the page while the question is open.
 */
export function useDeletion(refused: Refused, tell: (text: string) => void, reload: () => Promise<void>) {
  const [pending, setPending] = useState<Pending | null>(null)

  const ask = async (ids: number[]) => {
    const prompt = await api.deletionPrompt(ids, refused)
    if (prompt) setPending({ prompt, ids })
  }

  const confirm = async () => {
    if (!pending) return
    const ids = pending.ids
    setPending(null)
    if (await api.remove(ids, refused)) {
      tell(`Deleted ${counted(ids.length, 'event')}.`)
      await reload()
    }
  }

  const dialog = pending && (
    <ConfirmDialog text={pending.prompt} confirmLabel="Delete"
      onConfirm={confirm} onCancel={() => setPending(null)} />
  )
  return { ask, dialog }
}
