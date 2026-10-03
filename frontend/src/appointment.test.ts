import { describe, expect, it } from 'vitest'
import { heldAppointment, sheetAppointment, sinceAppointment } from './appointment'

const today = '2026-09-22'

describe('the last appointment', () => {
  it('is held when it is today or earlier', () => {
    expect(heldAppointment('2026-09-02', today)).toBe('2026-09-02')
    expect(heldAppointment(today, today)).toBe(today)
  })

  it('is none when nothing was kept', () => {
    expect(heldAppointment(null, today)).toBe('')
    expect(heldAppointment('', today)).toBe('')
  })

  it('is none when it is still to come', () => {
    // A range from a day not yet reached to today runs backwards.
    expect(heldAppointment('2026-09-23', today)).toBe('')
  })

  it('is none when the value is not a date', () => {
    expect(heldAppointment('soon', today)).toBe('')
    expect(heldAppointment('2026-9-2', today)).toBe('')
  })

  it('starts the range on its own day, which the range includes', () => {
    expect(sinceAppointment('2026-09-02', today)).toEqual(['2026-09-02', today])
  })
})

describe('the appointment on the sheet', () => {
  const held = '2026-09-02'

  it('is named wherever the range covers it, ends included', () => {
    expect(sheetAppointment(held, held, today, true)).toBe(held)
    expect(sheetAppointment(held, held, '2026-09-15', true)).toBe(held)
    expect(sheetAppointment(held, '2026-08-01', today, true)).toBe(held)
    expect(sheetAppointment(held, '2026-08-23', held, true)).toBe(held)
  })

  it('is not named where the range does not cover it', () => {
    expect(sheetAppointment(held, '2026-09-10', today, true)).toBe('')
    expect(sheetAppointment(held, '2026-08-01', '2026-08-31', true)).toBe('')
  })

  it('is not named when the reader says not to; nor when there is none', () => {
    expect(sheetAppointment(held, held, today, false)).toBe('')
    expect(sheetAppointment('', '2026-08-01', today, true)).toBe('')
  })
})
