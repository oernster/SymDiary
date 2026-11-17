# SymDiary: Software Requirements Specification

Status: baselined 2026-09-22. Derived the same day from
`health-symptom-tracker-requirements.md` (the source document, section 1.5).
Changes from here arrive as numbered amendments with a reason (section 6).
No questions remain open (Appendix B).

## 1. Introduction

### 1.1 Purpose

SymDiary records symptoms at the moment they are noticed. Later, it saves a
short factual record as a PDF to take to a doctor. It is a recorder, not a
diagnostician: it keeps what the user observed and when; interpreting that is
for the user and their doctor.

> When you notice a symptom, record it. When you see your doctor, take the
> record with you.

### 1.2 Intended audience

The owner (Oliver Ernster), who is the developer and the first user, plus any
AI assistant working on the code. The repository is public under GPL-3.0.

### 1.3 Scope

In scope for version 1:

- Recording a symptom event: symptom, time, optional severity, optional note.
- Reusable symptom definitions, offered by autocomplete.
- A history of events, filtered by date range, symptom and severity.
- Editing and deleting events.
- A symptom record over a chosen date range, saved as a PDF.
- A machine-readable export of the whole record.
- A Windows desktop application in Go with Wails, plus the house setup program.
- The same application on Linux and macOS, packaged as a Flatpak and a signed
  DMG (Amendment 7).
- A phone recorder for Android: records offline, then hands its events to the
  desktop as an export file through the phone's share sheet (Amendment 21,
  section 3.10).

Out of scope (decided by the source document, sections 10 and 15):

- Any diagnosis, suggested diagnosis or statement of what a symptom may mean.
- Any recommendation about medication or treatment.
- Any AI, generated text or interpretation of the user's observations.
- Health scores, risk scores, trend statements and colour-coded health ratings.
- Any claim that two events are related.
- Rewriting, normalising or translating the user's words into medical terms.
- Reminders, notifications or any prompt to record.
- Accounts, sign-in, cloud storage and synchronisation between machines.
- Advertising, analytics and telemetry.
- Event types other than symptoms (source section 14; see 3.8).
- A recorder on iPhone, until a device can measure it (Amendment 21, D-1).
- Any live link between phone and desktop; the handover is a file (Amendment 21,
  D-5).
- A single-file setup program for Linux or macOS. Those two are packaged the way
  each platform expects, a Flatpak and a signed DMG; the bespoke setup program
  stays Windows-only (Amendment 7).

### 1.4 Definitions

| Term | Meaning |
|---|---|
| Symptom | A label the user supplies for what they noticed, such as `Tired`. Stored exactly as typed. |
| Symptom definition | A symptom kept for reuse, so it is offered the next time. |
| Event | One observation: a symptom at a point in time, with optional severity and note. The fundamental unit of data. |
| Occurred at | When the user says the symptom happened. Defaults to the recording instant; editable. |
| Recorded at | When SymDiary created the event. Set once, never edited. |
| Severity | An optional user choice from a fixed list (FR-004). Never computed. |
| Note | Optional free text attached to an event, kept byte for byte as entered. |
| History | The list of events in the application, newest first. |
| Receipt | The symptom record for a date range: shown in the window and saved as a PDF (section 3.4). |
| Last appointment | One date, today or earlier, the user marks as their most recent appointment. A range boundary kept by the window, not part of the record (FR-046). |
| Export | The machine-readable file holding the whole record (section 3.5). |
| Record | Everything SymDiary stores: definitions plus events. |
| Local time | The time in the zone the desktop is set to. |
| Reference machine | Oliver's desktop, Windows 11 build 26200. |
| Recorder | The installable web application on the phone. |
| Desktop application | SymDiary as it ships on Windows, macOS and Linux. |
| Send | The recorder writing an export file of events then offering it to the phone's share sheet. |
| Sent | An event that has been in a send the share sheet completed (FR-090). |
| Batch | The events one send carried. |

### 1.5 References

- Source document: `C:\Users\Oliver\Downloads\health-symptom-tracker-requirements.md`
  (read 2026-09-22). Section numbers cited as "source N".
- `C:\Users\Oliver\Development\WhatDay\REQUIREMENTS.md`: the shape of this document.
- `C:\Users\Oliver\Development\ed-voyage-companion`: the Go + Wails delivery
  reference (build, gate, setup program).
- WCAG 2.2 level AA.

## 2. Overall description

### 2.1 Product perspective

A new, standalone desktop application for Windows, macOS and Linux
(Amendment 7). It has no server and no data other than the record and its own
settings. Its own code has no network access; the one fetch is the window
toolkit's, on a Windows machine missing WebView2 (Amendment 17, NFR-PRIV-001).

The phone recorder (Amendment 21) is a second client of the same domain and
application, compiled to WebAssembly and served as static files from
`symdiary.com/app/`. It has no server either; its only connections fetch its own
files from that site (NFR-PRIV-004). It reaches the desktop only as a file the
user moves.

### 2.2 User classes

One: the person whose symptoms are recorded, using their own user account.
A doctor reads the saved record and never uses the application. No
administrator rights are needed at any point.

### 2.3 Operating environment

Windows 10 or 11 on x64 (Amendment 17). Windows 11 ships the WebView2 runtime
the window renders through; where it is missing, as it can be on Windows 10,
the application and the setup program offer to install it before opening.
Built and tested on the reference machine, which runs Windows 11.

Linux and macOS carry the same application (Amendment 7). Linux is a Flatpak on
the GNOME runtime, which supplies the webkit2gtk-4.1 that Wails renders through;
macOS is a signed and notarized DMG for Apple Silicon. Both have now been built
on their own platform and run: the Flatpak's window opens and the DMG is
notarized. The gate adds to that on every run by compiling and vetting the whole
module for all three, which settles that nothing Windows-only can be written
without the next run saying so. What each platform run did and did not settle is
in TESTING.md.

Where each platform keeps the two files it owns:

| | The record | The run log |
|---|---|---|
| Windows | `%APPDATA%\SymDiary\symdiary.db` | `%LOCALAPPDATA%\SymDiary\SymDiary.log` |
| Linux | `$XDG_CONFIG_HOME/SymDiary/symdiary.db` | `$XDG_STATE_HOME/SymDiary/SymDiary.log` |
| macOS | `~/Library/Application Support/SymDiary/symdiary.db` | `~/Library/Logs/SymDiary/SymDiary.log` |

Inside the Flatpak both variables already point at the sandbox, so the same
rules land under `~/.var/app/uk.codecrafter.SymDiary`.

The phone recorder runs installed to the home screen on Android, in any
browser that has what it needs (NFR-PORT-001, FR-104), keeping its events in
that browser's storage for `symdiary.com`. Seen working end to end in Chrome
157 on a Samsung Galaxy Z Flip 5 (Appendix A).

### 2.4 Constraints

- C-1 Language: Go 1.26, with SymDiary's own code and its libraries pure Go and
  no cgo in the Windows build; on Linux and macOS Wails' own window code is
  built with cgo. Frontend React with TypeScript on Vite.
- C-2 Toolkit: Wails v2 (v2.12.0 measured on the reference machine).
- C-3 Storage: SQLite through `modernc.org/sqlite` (pure Go).
- C-4 Layering: `internal/{domain,application,infrastructure}` with the UI in
  the root package and `frontend/`, plus an explicit composition root; enforced
  by a structural test.
- C-5 The setup program is ported from ED Voyage Companion's `installer/`.
- C-6 Licence: GPL-3.0 (already in the repository).
- C-7 `VERSION` at the repository root is the single source of the version.
- C-8 The README header (the `# SymDiary` title and the owner's opening line)
  stays verbatim; additions go beneath it.
- C-9 The phone recorder runs `internal/domain` and `internal/application`
  compiled to WebAssembly; no rule is written a second time for the phone
  (Amendment 21, D-10).

### 2.5 Assumptions and dependencies

| ID | Assumption | Owner | Status |
|---|---|---|---|
| A-1 | One person's record is small: under 20,000 events over ten years (about five a day). | Oliver | Accepted 2026-09-23, untested. It sets the scale the design assumes rather than a threshold anything enforces; nothing in the product fails at a larger record and nothing measures one (NFR-PERF-002). |
| A-2 | The desktop's clock is right; SymDiary trusts it for the default occurrence time. | Oliver | Accepted 2026-09-23. The application reads the clock and offers the instant as the default, which the user may replace (FR-003), so a wrong clock is corrected by the person rather than detected by the program. |
| A-3 | Oliver supplies the application artwork as a master PNG with a transparent background. | Oliver | Satisfied. The masters are in `assets/` plus `donate.png` at the root; `tools/genicons.py` derives every shipped size from them. |
| A-4 | A PDF the application draws itself opens on a machine that has never met SymDiary; it says the same thing on every desktop. | Claude | Measured and true (Appendix A, M-1) |
| A-5 | An installed recorder's events survive weeks of the recorder going unopened on Android. | Oliver | Accepted 2026-11-16, untested beyond a single evening. Chrome granted persistent storage (M-3), which is the measured half. The probe stays installed on the owner's phone as a running check; FR-081 and FR-082 bound the loss to unsent events if it proves false. |

## 3. Requirements

Priorities use MoSCoW. Every requirement names the test or check that verifies
it. Each name was read off the tree rather than written from intent: a reference
that does not resolve is a requirement nobody can check.

### 3.1 Functional requirements: recording

**FR-001 Record now**
- Priority: Must
- Requirement: When the user presses `Record now` with a symptom entered, the
  recording service shall store one event whose occurrence time is the
  recording instant, unless the user has set another time (FR-003).
- Acceptance: Given the clock reads 2026-09-22T17:12:00+01:00, symptom `Tired`
  and note `Only been awake for about 10 minutes.`, when `Record now` is
  pressed, then the history's first entry reads 22 Sep 2026 17:12, `Tired`,
  with that note and no severity.
- Verified by: `internal/application/record_test.go::TestRecordNowUsesTheClock`

**FR-002 Symptom required**
- Priority: Must
- Requirement: If `Record now` is pressed with an empty symptom (nothing but
  white space), then the recording form shall store nothing and shall say that
  a symptom is needed.
- Verified by: `internal/domain/event_test.go::TestBlankSymptomRefused`

**FR-003 Set the occurrence time**
- Priority: Must
- Requirement: The recording form shall let the user replace the default
  occurrence time with another local date and time before recording.
- Rationale: Source 6: at 15:00 the user remembers a headache at about 12:30.
- Acceptance: Given the clock reads 15:00 and the user sets 12:30 the same
  day, when `Tired` is recorded, then the event's occurred-at is 12:30 and its
  recorded-at is 15:00.
- Verified by: `record_test.go::TestRetrospectiveTimeKept`

**FR-004 Optional severity**
- Priority: Must
- Requirement: The recording form shall offer severity as an optional choice
  of `Mild`, `Moderate` or `Severe`, with no choice selected by default.
- Verified by: `internal/domain/severity_test.go::TestSeverityIsOptional`

**FR-005 Optional note**
- Priority: Must
- Requirement: The recording service shall store the note exactly as entered,
  with no trimming, spelling correction or other change.
- Acceptance: A note of `  Much worse than yesterday.  ` (leading and trailing
  spaces) reads back identical, byte for byte.
- Verified by: `internal/infrastructure/store/store_test.go::TestTheStoreContract`, subtest "a note and symptom round trip byte for byte"

**FR-006 Occurrence time never in the future**
- Priority: Must
- Requirement: If the user sets an occurrence time later than the current
  instant, then the recording form shall store nothing and shall say that an
  occurrence cannot be in the future.
- Verified by: `event_test.go::TestFutureOccurrenceRefused`

**FR-007 Recorded-at set once**
- Priority: Must
- Requirement: The recording service shall set recorded-at when the event is
  created; no later operation shall change it.
- Verified by: `edit_test.go::TestEditLeavesRecordedAtAlone`

**FR-008 Form after recording**
- Priority: Must
- Requirement: When an event has been recorded, the recording form shall clear
  its symptom, note, severity and time, then place the keyboard focus on the
  symptom field.
- Rationale: The next observation starts from a clean form (source 3).
- Verified by: `frontend/src/RecordPane.test.tsx`, "records what was entered and clears afterwards"

**FR-009 Recording failure**
- Priority: Must
- Requirement: If the store cannot write an event, then the recording form
  shall keep everything the user entered and shall say that the event was not
  saved, naming the reason.
- Verified by: `record_test.go::TestWriteFailureLosesNothing`

### 3.2 Functional requirements: symptom definitions

**FR-010 New symptom becomes a definition**
- Priority: Must
- Requirement: When an event is recorded with a symptom that matches no
  definition, the recording service shall create a definition holding the
  symptom exactly as typed.
- Verified by: `record_test.go::TestNewSymptomBecomesDefinition`

**FR-011 Autocomplete**
- Priority: Must
- Requirement: While the user types in the symptom field, the recording form
  shall list every definition containing the typed text, matched without
  regard to letter case, most recently used first.
- Acceptance: Given definitions `Tired`, `Headache` and `Back pain`, typing
  `he` lists `Headache` only; typing `a` lists `Headache` and `Back pain`
  (Amendment 1).
- Verified by: `internal/domain/definition_test.go::TestSuggestionsMatchIgnoringCase`

**FR-012 Free entry**
- Priority: Must
- Requirement: The recording form shall accept a symptom that is not in the
  list, without asking the user to confirm it.
- Verified by: `frontend/src/RecordPane.test.tsx`, "records what was entered and clears afterwards", which types a symptom no definition holds

**FR-013 Labels kept as typed**
- Priority: Must
- Requirement: The application shall display every symptom exactly as the
  user typed it; no code path shall change its case, spelling or wording.
- Verified by: `store_test.go::TestTheStoreContract`, subtest "a note and symptom round trip byte for byte"

**FR-014 Near duplicates**
- Priority: Must
- Requirement: When the typed symptom differs from an existing definition only
  in letter case or surrounding spaces, the recording service shall use the
  existing definition, keeping the label as first typed (Q-2).
- Verified by: `definition_test.go::TestCaseVariantReusesDefinition`

### 3.3 Functional requirements: history, editing and deletion

**FR-020 History order**
- Priority: Must
- Requirement: The history shall list events by occurrence time, newest first,
  showing for each the local date and time, the symptom, the severity where
  given and the note where given.
- Verified by: `internal/application/history_test.go::TestNewestFirst`

**FR-021 Filters**
- Priority: Must
- Requirement: The history shall filter events by an inclusive local date
  range, by one or more symptoms and by one or more severities; filters
  combine, each narrowing the others.
- Acceptance: Given 12 events of `Tired` and 3 of `Headache` in September,
  filtering to `Headache` from 1 to 30 September lists 3.
- Verified by: `internal/domain/filter_test.go::TestFiltersCombine`

**FR-022 No analysis**
- Priority: Must
- Requirement: The history shall show events and counts only; it shall show no
  trend, average, score, chart or statement about what the events mean.
- Verified by: inspection against source 10.

**FR-023 Edit an event**
- Priority: Must
- Requirement: The application shall let the user change an event's symptom,
  occurrence time, severity and note.
- Verified by: `internal/application/edit_test.go::TestEditKeepsOccurredAt`, plus `frontend/src/EditRow.test.tsx`, "opens with what the event holds"

**FR-024 Edit keeps the occurrence time**
- Priority: Must
- Requirement: When an event is edited without its occurrence time being
  changed, the edit service shall keep the occurrence time exactly as it was.
- Acceptance: Given an event occurring at 09:30, when its note is edited at
  14:00, then it still occurs at 09:30.
- Verified by: `edit_test.go::TestEditKeepsOccurredAt`

**FR-025 Delete an event**
- Priority: Must
- Requirement: When the user asks to delete an event, the application shall
  show a confirmation naming the event's symptom and local date and time; the
  event is removed only when the user confirms.
- Verified by: `internal/application/edit_test.go::TestDeleteNeedsConfirmation`,
  `frontend/src/HistoryPane.test.tsx`, "asks before deleting, naming the event; deletes only when confirmed"

**FR-026 Delete many**
- Priority: Could
- Requirement: Where several events are selected, the delete confirmation shall
  state how many events will be removed.
- Verified by: `frontend/src/HistoryPane.test.tsx`, "counts the events chosen for a bulk deletion"

**FR-027 Edit and delete failure**
- Priority: Must
- Requirement: If the store cannot write an edit or a deletion, then the
  application shall leave the event as it was and shall say so, naming the
  reason.
- Verified by: `edit_test.go::TestFailedEditChangesNothing`

**FR-028 Rename a definition**
- Priority: Should
- Requirement: The application shall let the user rename a symptom definition;
  every event using it then shows the new name (Q-3).
- Verified by: `internal/application/history_test.go::TestRenameReachesEveryEvent`

### 3.4 Functional requirements: the receipt

**FR-040 Save a receipt as a PDF**
- Priority: Must
- Requirement: When the user asks for a receipt over a local date range, the
  application shall show it and shall, on request, write it to a PDF file at a
  location the user chooses (Amendment 16).
- The document shall be drawn by the application itself rather than by the
  window's browser engine, so that a record reads the same whichever desktop
  produced it.
- Acceptance: the saved document is a PDF; it holds the receipt of FR-041 with
  the framing of FR-045; every page carries `Page N of M`; no page begins with
  what was observed at an event whose time was on the page before; cancelling
  the dialog writes nothing and says nothing.
- Once the record is showing, changing the range shall show the record for the
  new range without being asked again (Amendment 19). Save PDF writes the
  range in the fields, so the record on screen must always be that one; an
  answer for a range since left is discarded.
- Each save dialog shall filter to the kind of file it is about to write and
  shall suggest a name carrying that extension. macOS applies a dialog's filter
  to the name it is given, so a document offered under another kind's filter is
  written under a name that says one thing while its contents say another.
- Verified by: `internal/application/history_test.go::TestReceiptFromTheStore`,
  `pdf_test.go`, `internal/infrastructure/pdf/layout_test.go` and
  `internal/infrastructure/pdf/sheet_test.go`, plus opening a saved document on
  the reference machine (A-4). The dialog's own kind is held by
  `pdf_test.go::TestTheSaveDialogAsksForTheKindOfFileItIsActuallyWriting`,
  proved to bite by planting the defect it was written for. The record following
  the range is held by `frontend/src/ReceiptPane.test.tsx`, "follows a change of
  dates once the record is showing" and "never lets a slow answer for an earlier
  range overwrite the current one", the second proved to bite by planting the
  discard out.

**FR-041 Receipt contents**
- Priority: Must
- Requirement: The receipt shall hold, in order: the title `SYMPTOM RECORD`;
  the date range; the last appointment where FR-046 names one; then, for each
  symptom recorded in the range, a heading of
  the symptom and its number of events, followed by each event's local date
  and time, severity where given and note where given. The fixed framing of
  FR-045 opens this order and sits within it, which it does not otherwise
  disturb.
- Acceptance: Given `Headache` on 25 Aug 12:30 and 2 Sep 08:10 plus `Tired`
  on 28 Aug 07:45 and 22 Sep 17:12, the receipt for 23 Aug to 22 Sep reads
  `Headache - 2 recorded events`, `25 Aug 12:30`, `02 Sep 08:10`, then
  `Tired - 2 recorded events`, `28 Aug 07:45`, `22 Sep 17:12`, each event
  followed by its severity and note where given.
- Event times leave out the year while the whole range lies in one year, as the
  source example does; a range crossing a year names the year on every event
  (Amendment 2).
- Verified by: `internal/domain/receipt_test.go::TestReceiptMatchesTheAcceptanceExample`,
  `TestReceiptAcrossAYearNamesTheYear`

**FR-042 Arithmetic only**
- Priority: Must
- Requirement: The receipt shall contain only what the user recorded, counts of
  events, the last appointment the user marked where FR-046 names it and the
  fixed framing of FR-045; it shall contain no interpretation, trend,
  comparison or suggestion.
- The framing is fixed text that names no event and changes with no event, so
  it states nothing about what was recorded. Anything that varies with the
  record belongs to FR-041 and is held by this requirement. The appointment is
  a date the user typed, printed as typed; it is set beside the range, never
  among the events, so the sheet compares nothing with it.
- Verified by: `receipt_test.go::TestReceiptHoldsNoOtherText`, which asserts
  every line of a generated receipt is a title, a range, the appointment the
  user marked, a heading, a count, a field of a recorded event or one of the
  two framing lines.

**FR-043 Empty range**
- Priority: Must
- Requirement: If the chosen range holds no events, then the receipt service
  shall say so and shall produce no receipt.
- Verified by: `receipt_test.go::TestEmptyRangeGivesNoReceipt`

**FR-044 Receipt order**
- Priority: Must
- Requirement: The receipt shall order symptom groups by each symptom's
  earliest event in the range, oldest first; within a group, events run oldest
  to newest. Names play no part in the order (Q-4).
- Rationale: The source list says "chronological" while its example runs
  newest first; the owner chose oldest first, never alphabetical.
- Verified by: `receipt_test.go::TestReceiptOrder`

**FR-045 Printed framing**
- Priority: Must
- Requirement: The receipt shall carry, above the title and once only, a line
  naming the program that produced it and its home address (Amendment 14).
  Immediately below the date range it shall carry a statement that the record
  is not a diagnosis but the person's own notes, printed for a healthcare
  professional.
- Acceptance: Given the receipt of FR-041's example, its first line reads
  `Generated by SymDiary (https://symdiary.com)`, its fourth line reads `This
  record is not a diagnosis. It is one person's own notes of what they observed
  and when, printed for a healthcare professional to read.` and neither line
  appears anywhere else on the sheet.
- The wording avoids saying the record is guidance for a professional to use.
- The application's own icon shall be drawn beside that line and beside no
  other, so a sheet lying on a desk says what produced it at a glance
  (Amendment 13). It identifies the producer; it adds nothing to what the notes
  say and nothing that reads as a finding.
- Every page of the document shall say which page it is and how many there are
  (Amendment 15). A record read in the wrong order is worse than one that is
  hard to read; a sheaf of paper handed across a desk can be dropped.
- The document shall carry nothing the record does not hold: no date of
  printing, no window title, no address of its own (Amendment 16).
- Verified by `ReceiptPane.test.tsx`, which asserts exactly one mark on screen,
  in the first line, carrying no alt text; by
  `internal/infrastructure/pdf/sheet_test.go`, which asserts the mark is drawn
  on the document and that a record with none is still written; and by
  `tests/structural/linekind_test.go::TestEveryLineKindIsDrawnInTheDocument`,
  which holds the document's styles to the domain's own line kinds so a kind
  cannot reach the paper with no style at all. Proved to bite by planting a
  missing style.
  MHRA guidance v1.10f treats software that provides information to help a
  healthcare professional reach a clinical decision as a separate category
  (page 12); a printed sheet that calls itself an input to that decision is
  describing a purpose SymDiary does not have. Notes read by a professional are
  what section 4.1 confirmed SymDiary to be, so that is what the sheet says.
- The words and the address live in `internal/product`, which is the one home
  for the facts every layer may name. The domain is handed them rather than
  reading them, so no layer boundary is crossed to print a line.
- Verified by: `internal/domain/receipt_test.go::TestTheFramingOpensTheRecord`,
  `TestTheFramingIsTheSameWhateverTheRecord`

**FR-046 Since last appointment** (Amendments 18 and 19)
- Priority: Should
- Requirement: The receipt pane shall let the user mark one date, today or
  earlier, as their last appointment. While one is held, the pane shall offer
  two ranges, `Since last appointment` (from that date to today, both
  included) and `Custom dates`; it shall open on the first. While none is held,
  it shall offer custom dates alone, opening on the last thirty days as before.
  Changing either date by hand makes the range custom.
- The choice of range fills the two dates and nothing else: the receipt is
  the one the same dates give when chosen by hand.
- Where the range covers the last appointment (its first and last days
  included), the receipt shall carry the line `Last appointment: DD Mon YYYY`
  directly under the date range, above the statement of FR-045 (Amendment 19).
  The rule is the range, not how it was chosen: a custom range spanning the
  appointment names it too. A range that does not cover it names nothing. The
  receipt service refuses an appointment outside the range rather than
  printing it.
- Beside the date the pane shall offer `Show on the sheet`, ticked when the
  pane opens and enabled only while the range covers the appointment; unticked,
  the sheet names no appointment.
- The line names the most recent appointment only, since only one is held;
  `Last` says so rather than implying it is the only one. It is never placed
  among the events: arranging the record around the appointment would be the
  software presenting the data, which section 4.1 keeps out.
- An appointment is a date boundary, not an event and not part of the record.
  It lives in the window's own storage beside the theme (FR-073), so it is
  neither exported nor imported. Storage that refuses to be read holds no
  appointment. A value that is not a date is none as well; so is a date still
  to come.
- Nothing prompts the user to mark one (section 1.3: no prompts).
- Acceptance: Given a last appointment of 2 Sep 2026 on 22 Sep 2026, the pane
  opens on `Since last appointment` with the range 2 Sep to 22 Sep and asks for
  the receipt of exactly those dates, its fourth line reading `Last
  appointment: 02 Sep 2026`. The same holds for 1 Aug to 22 Sep and for 23 Aug
  to 2 Sep; 10 Sep to 22 Sep and 1 Aug to 31 Aug name no appointment. Given
  none, it opens on 23 Aug to 22 Sep and offers no choice of range.
- Verified by: `frontend/src/appointment.test.ts`;
  `frontend/src/ReceiptPane.test.tsx`, "since the last appointment";
  `internal/domain/receipt_test.go::TestTheLastAppointmentIsNamedOnlyWithinTheRange`,
  `TestTheLastAppointmentSitsUnderTheRange`, `TestNoAppointmentAddsNoLine`;
  `receipt_test.go::TestTheSheetNamesTheLastAppointmentTheWindowSends`. Proved
  to bite by planting a future appointment let through, a range starting the
  day after and a receipt service accepting an appointment outside its range.

### 3.5 Functional requirements: export

**FR-050 Export**
- Priority: Must
- Requirement: When the user asks to export, the export service shall write the
  whole record to a JSON file at a path the user chooses in a save dialog. The
  dialog shall open in the user's Downloads folder, which the user may leave
  (Amendment 4).
- Verified by: `internal/infrastructure/export/export_test.go::TestExportWritesEveryEventAndReadsBack`

**FR-051 Export format**
- Priority: Must
- Requirement: The export shall carry a format version plus every definition
  and every event with its symptom, occurred-at, recorded-at, severity and
  note; times are RFC 3339 with their UTC offset.
- Verified by: `export_test.go::TestFormatIsStable` against a committed sample
  file.

**FR-052 Export failure**
- Priority: Must
- Requirement: If the export file cannot be written, then the export service
  shall leave no partial file behind and shall say so, naming the path and the
  reason.
- Verified by: `export_test.go::TestFailedExportLeavesNothing`

**FR-053 Import**
- Priority: Should
- Requirement: The application shall read an export file back, adding its
  events to the record and skipping any event already present (Q-5).
- Verified by: `internal/application/transfer_test.go::TestExportThenImportRoundTrip`, `TestImportSkipsRepeatsWithinTheFile`

### 3.6 Functional requirements: storage and lifecycle

**FR-060 Local store**
- Priority: Must
- Requirement: The application shall keep the record in one SQLite file, in the
  folder the platform keeps a program's own configuration in: on Windows
  `%APPDATA%\SymDiary\symdiary.db` (Amendment 7 names the other two in 2.3).
- Verified by: `internal/infrastructure/store/unavailable_test.go::TestTheRecordSitsInTheFolderThePlatformNames`,
  which reads the folder from the platform through `os.UserConfigDir`

**FR-061 First run**
- Priority: Must
- Requirement: If no record file exists, then the application shall create an
  empty one without reporting a fault; the recording form is then ready.
- Verified by: `store_test.go::TestMissingFileGivesEmptyRecord`

**FR-062 Unreadable record**
- Priority: Must
- Requirement: If the record file exists but cannot be opened, then the
  application shall open its window, say that the record could not be read
  and name the file with the reason. It shall not overwrite the file.
- Rationale: House robustness rule 1: nothing before the window may end the run.
- Verified by: `store_test.go::TestCorruptFileIsNeverReplaced`

**FR-063 Schema versions**
- Priority: Must
- Requirement: The store shall record its schema version; when it opens a file
  of an older version, the store shall upgrade it in one transaction.
- Verified by: `store_test.go::TestNewerSchemaRefused` and `TestReopenKeepsEverything`. The migration list holds one entry, so no file older than the current schema exists to upgrade from; the loop that would do it runs on every open and is exercised whenever a fresh file is made

**FR-064 Single instance**
- Priority: Must
- Requirement: When SymDiary starts while another instance is running for the
  same user, the new instance shall bring the running window forward and exit.
- Verified by: a manual check with two copies started; the lock is Wails' own and needs two real processes, so no test covers it; it is checked by hand in a real build

**FR-065 Log**
- Priority: Must
- Requirement: The application shall point standard error at a log in the folder
  the platform keeps a program's own state in, as its first act: on Windows
  `%LOCALAPPDATA%\SymDiary\SymDiary.log` (Amendment 7 names the other two in 2.3).
  The log shall never contain a symptom, a note or any other part of the record.
- Note: pointing standard error at a file is Windows-only work, because the
  failure it answers is Windows-only: a windowed run started from a shortcut is
  handed a handle of 0 and everything written there is lost. Elsewhere the run
  keeps its own standard error and the crash report is copied to the log.
- Verified by: `runlog_test.go`, including `TestEveryPlatformsLogFolder`, plus
  `tests/structural/logscan_test.go::TestOnlyTheKnownPlacesWriteToTheLog`,
  which holds the log's writers to a declared list. Nothing about a write to
  standard error says whether its arguments came from the record, so the
  promise is kept by there being few enough writers to read, all of them known
  and each a constant sentence plus an error or a stack. A new one fails the
  test and has to be declared, which is the moment to ask what it puts in the
  file.

**FR-066 About**
- Priority: Must
- Requirement: The application shall offer an About dialog stating the
  product name, the version, the author, the copyright notice and the
  open-source works it is built with, each with its licence, plus the statement
  in FR-067. The notice shall read `© Oliver Ernster 2026` (Amendment 8): the
  symbol says it, so the word beside it says it twice.
- Note: the year is the year of the first release, not the year the program is
  run in. A notice that follows the clock claims a date nothing was published
  on.
- Verified by: `internal/application/about_test.go`;
  `frontend/src/App.test.tsx::opens the Guide and About, then closes them again`

**FR-068 Guide** (Amendment 5)
- Priority: Must
- Requirement: The application shall offer a Guide naming every control with
  the picture that control draws, stating how to record quickly, what SymDiary
  will not do and how the record is kept.
- Rationale: The house guide, as PigeonPost and ClearBudget carry one.
- Verified by: `frontend/src/App.test.tsx::opens the Guide and About, then closes them again`

**FR-069 Donation** (Amendment 6)
- Priority: Should
- Requirement: The application shall offer a button in the bar that hands a
  donation address to the desktop for the user's browser to open. The
  application shall not itself fetch that page; the address shall be stated in
  one place in the application.
- Rationale: The house donate button. SymDiary is free and stays free: nothing
  is withheld behind a donation, so the ask is a postscript rather than a
  prompt.
- Acceptance: Pressing Donate asks the desktop for exactly one address, that
  address is `https://www.paypal.com/ncp/payment/4XP3AYNMPQGUC`; no connection
  is opened by SymDiary itself.
- Verified by: `donate_test.go`;
  `frontend/src/App.test.tsx::offers the donation page last in the band, saying that the browser opens`;
  `tests/structural/boundary_test.go::TestNoNetworkImports`

**FR-073 Light and dark** (Amendment 12)
- Priority: Should
- Requirement: The application shall open dark and shall offer a button in the
  bar that moves it between dark and light. The choice shall be remembered for
  the next opening. The button shall show the mode it would move to, in its
  picture and in its words. The setup program shall behave the same way.
- Rationale: The owner's decision, superseding Q-9. SymDiary no longer reads the
  Windows app mode: the window changing under the reader because the desktop
  reached dusk is a surprise, while the button is one press away.
- The button stands between two rules, to the left of Guide: it belongs to
  neither the transfer group before it nor the help group after it.
- Acceptance: A window opened for the first time is dark and its button reads
  `Light mode` over the sun. Pressing it makes the window light and the button
  read `Dark mode` over the moon. Closing and reopening keeps light.
- The choice lives in the window's own storage rather than in the record or in
  a file of its own: it is a preference about this machine's window, not part
  of what the user recorded. Storage that refuses to be read or written leaves
  the default showing rather than ending the page: a theme is not worth a dead
  window, while the cost of a refusal is one press next time.
- Verified by: `frontend/src/theme.test.ts`, `useTheme.test.tsx`,
  `App.test.tsx`, "dresses the window from the bar and keeps the choice", plus
  a manual pass over the setup program.

**FR-067 Not medical advice**
- Priority: Must
- Requirement: The About dialog shall state that SymDiary records observations
  only and gives no medical advice.
- Verified by: `about_test.go::TestAboutCarriesTheStatement`

### 3.7 Functional requirements: setup program

**FR-070 Per-user install**
- Priority: Must
- Requirement: The setup program shall install SymDiary to
  `%LOCALAPPDATA%\Programs\SymDiary` without requesting administrator rights.
- Verified by: manual install on the reference machine.

**FR-071 Routes**
- Priority: Must
- Requirement: The setup program shall offer install, update, going back to an
  older version, repair and uninstall, each on its own screen, registered in
  the Windows Apps list.
- Verified by: manual; each route exercised once.

**FR-072 Uninstall keeps the record**
- Priority: Must
- Requirement: When uninstalling, the setup program shall remove the program
  files, shortcuts and registry entries; it shall remove the record only where
  the user ticks a box naming the record file.
- Rationale: The record is the user's (source 12); losing it by removing the
  program would be accidental loss (source 13).
- Verified by `setup.TestTheRecordIsNeverClearedWithTheLeftovers`, which fails
  if the record's folder is among the folders an uninstall clears unasked or
  inside one of them; also by
  `setup.TestTheRecordGoesOnlyWhenItIsAskedFor`. The tick reaching the flag was
  settled by hand on 2026-09-23 against the built setup program, as a pair of
  runs: with the box left alone the record survived, with it ticked the record
  went. No test can cover that hop, because it runs through the setup page,
  which has no build step to type check it.

**FR-074 The licence is explained, not just named** (Amendment 13)
- Priority: Should
- Requirement: The setup program shall carry a licence screen, reachable from
  its header on every screen but progress, stating in ordinary words what the
  licence permits and requires, followed by the published licence text in full.
  Every word of it shall come from the application rather than being written
  into the page.
- Rationale: Naming a licence explains nothing. "GNU General Public Licence,
  version 3" tells most people nothing at all; somebody about to install a
  program deserves to know what they are being given in words they already
  use. The full text follows for anyone who wants the thing itself.
- Acceptance: The screen names the product and the licence in one sentence,
  lists what the reader may do and what they must do in turn, then shows the
  licence text in a pane that scrolls.
- The text shown shall be the text the repository publishes, which
  `tests/structural/licence_test.go` holds byte for byte.
- The pane shall read itself down at the house pace and step aside the moment
  the reader touches it (FR-075).
- A copy of the licence shall also be delivered with the program on every
  platform: in the install folder on Windows, under `share/licenses` inside
  the Flatpak, in the bundle's resources on macOS. Section 4 of the licence
  asks that every recipient be given a copy along with the program;
  a screen in a setup program is not a copy the recipient keeps. Verified by
  looking in the three places after a build.
- The works the setup program is itself built from are named at the foot of
  the licence text. They sat in a box above it until the box and the pane
  together no longer fitted the fixed window, at which point the box would
  have pushed the heading off the screen.

**FR-075 The licence reads itself** (Amendment 13)
- Priority: Should
- Requirement: The licence pane shall hold still when the screen opens, read
  itself down slowly, hold at the end, rewind and repeat. Any manual reading
  input shall suspend the cycle, which shall then resume from where the reader
  left it rather than restarting.
- Rationale: A licence is long and most people will not scroll it. A pane that
  reads itself lets somebody watch it go by; one that stops the moment they
  touch it never fights them for the scrollbar.
- Acceptance: The pace is the application's own, stated once: still for five
  seconds, then one pixel every two ticks of forty milliseconds, five seconds
  at the foot, fifteen pixels a tick back up, two seconds at the top. A manual
  input holds it for two and a half seconds of stillness.
- The pane is a text view, so it draws no focus ring in any state, Tab
  included (NFR-USE-002).

### 3.8 Non-functional requirements

**NFR-USE-001 Recording speed**: With SymDiary's window open, recording an event
of an existing symptom with no note shall take no more than 4 keystrokes after
the first letters of the symptom: choose the suggestion, then press Enter to
record. The keystroke count itself is not instrumented. What the suite holds is
the path those keys drive: `frontend/src/RecordPane.test.tsx` covers entering a
symptom, recording it and the form clearing afterwards, while
`useRing.test.tsx` and `ring.test.ts` hold the keyboard model that gets the
focus there. Counting the keys would mean a test asserting a number that the
layout could change without the behaviour changing, which is worth doing only
if recording ever stops feeling quick.

**NFR-USE-002 Keyboard**: Every action shall be reachable by keyboard alone
(Amendment 8). Tab and Right shall step the focus ring forward and Shift+Tab and
Left shall step it back, both wrapping at the ends; Enter and Space shall fire
the focused control; Escape shall close an open dialog. A field holding text
keeps the horizontal arrows for its caret and is left with Tab. The window shall
open with nothing focused; a dialog shall open on its first control, passing
over its scrolling body. The window shall hand the page the keyboard as it
opens, so the first key pressed reaches the ring without the page being clicked
first (Amendment 11). Verified by `frontend/src/ring.test.ts`,
`useRing.test.tsx` and `window_test.go::TestReadyAsksTheWindowForTheKeyboard`,
plus a manual pass.

**NFR-USE-003 Contrast**: Text shall meet WCAG 2.2 AA contrast (4.5:1 for body
text) in the theme in use; each ring colour shall meet the 3:1 WCAG 2.2 asks
of a non-text indicator against the surfaces it is drawn on. Verified by a test
over the colour tokens.

**NFR-USE-004 Focus ring** (Amendment 8): A control shall show no ring at rest,
the ring colour while it is hovered or keyboard-focused, then the danger colour
permanently while it is disabled; a disabled control's fill shall be muted so
the ring reads against it. No container shall take focus or paint a ring. The
accent colour shall never be used as a ring. Verified by
`tests/structural/focus_test.go`, each assertion proved by planting.

**NFR-PERF-001 Startup**: The window shall be usable promptly after it is
started. Not instrumented, by the owner's decision on 2026-09-22: startup is
quick enough in use, so nothing times it. The earlier form of this requirement
named a 2 second budget measured between two lines in the run log; only one of
those lines was ever written, so the budget was never a reading. Timing it would
mean the page reporting itself ready and the log carrying that second line,
which is worth doing only if startup ever stops feeling quick.

**NFR-PERF-002 History**: The history shall answer a filter change promptly at
the sizes one person's record reaches (A-1). Not instrumented, by the owner's
decision on 2026-09-23, on the same ground as NFR-PERF-001: filtering is quick
in use and nothing times it. The earlier form named 300 ms at the 95th
percentile over 100 changes against a generated store of 20,000 events; no such
benchmark was ever written, so the number was a target rather than a reading and
a requirement nobody can check is worse than one that says what it is. Measuring
it would mean a generated store and a Go benchmark over the filter path, which
is worth doing if the history ever stops feeling immediate.

**NFR-PERF-003 Receipt**: Saving a record as a PDF shall be prompt at the sizes
a record reaches. Not instrumented, for the same reason as NFR-PERF-002: the
earlier form named 2 s for over 1,000 events and nothing ever timed it. What has
been read instead is the document itself, on each platform: see TESTING.md,
where documents of one to three pages were written and looked right; the suite
checks that a 120 event record runs to the pages its layout predicts
(`sheet_test.go::TestALongRecordRunsToSeveralPages`) rather than timing it.

**NFR-PRIV-001 No network**: The application and the setup program shall open
no network connection of their own. Verified by `tests/structural/boundary_test.go::TestNoNetworkImports`
(forbids `net` and every `net/` package in the repository's own Go code) plus a
Content-Security-Policy carrying `default-src 'self'` and `connect-src 'none'`
on the frontend, checked by `TestThePageOpensNoConnection` reading `index.html`. The one exception is
Wails' own: on Windows, where WebView2 is missing or older than 94.0.992.31, the
toolkit asks first; if the user agrees, it downloads Microsoft's bootstrapper
before the window opens (Amendment 17). Nothing from the record is involved; the
boundary test cannot see that code, since it lives in the Wails module.

**NFR-PRIV-002 No telemetry**: No component shall collect or send usage data.
Verified by the same boundary test plus a dependency review in `go.mod` and
`package.json`.

**NFR-PRIV-003 Encryption at rest**: The record is not encrypted at rest; it is
protected by the user's own account only. This non-claim shall be stated in the
README (Q-6).

**NFR-REL-001 Crash safety**: An event reported as recorded shall survive a
power loss one second later. The store runs SQLite in WAL mode with
`synchronous=FULL`. Verified by `store_test.go::TestDurabilityPragmas`.

**NFR-MAINT-001 Coverage**: `internal/domain` and `internal/application` shall
be held at 100% statement coverage by `test.ps1`, which `build.ps1` runs before
it builds anything. `build.ps1 -Fast` skips the gate for a working loop and
prints that it has; a release is never cut with it.

**NFR-MAINT-002 Structure**: A structural test shall enforce the layering
(C-4), domain purity (no `os`, `time.Now` or `database/sql` in the domain), the
400-line module cap and its 381 to 400 danger band.

**NFR-MAINT-003 Checks**: `gofmt`, `go vet`, `staticcheck`, `tsc --noEmit`,
`eslint` and Vitest shall report nothing.

**NFR-MAINT-004 Docs**: The repository shall carry `README.md` (with who it is
for and not for), this specification, `ARCHITECTURE.md`, `TESTING.md`,
`DEVELOPMENT.md` and `VERSION`.

**NFR-MAINT-005 Wire**: A structural test shall compare the Go DTOs with the
hand-written TypeScript interfaces, field for field; a second shall compare
the receipt's line kinds with the union the page admits. The first sees that a
line carries a kind; only the second sees which kinds exist.

**NFR-PRIV-004 The recorder opens no connection of its own**: After its files
have loaded, the recorder shall open no connection. The page carries a
Content-Security-Policy of `default-src 'self'`, `script-src 'self'
'wasm-unsafe-eval'` and `connect-src 'none'`; the second is what lets the Go
code compile (M-6) and grants no network access. The
only fetches are the recorder's service worker collecting the recorder's own
files from its site, when a version installs and for any file missing from its
cache. Nothing from the record is involved. Verified by
`tests/structural/boundary_test.go::TestThePageOpensNoConnection`, which reads
`frontend/recorder.html` for `default-src 'self'` and `connect-src 'none'`.

**NFR-PRIV-005 The host**: The recorder shall be served as static files from an
origin that sets no cookie and on which no page loads a script from another
origin (Q-13). Verified by reading the response headers of every file the
recorder loads, plus
`tests/structural/site_test.go::TestNoSitePageLoadsAScriptFromAnotherSite`
over every page under `docs/`.

**NFR-PRIV-006 Encryption at rest on the phone**: The recorder's events are not
encrypted at rest; they are protected by the phone's own lock. This non-claim
shall be stated in the README beside NFR-PRIV-003's.

**NFR-REL-002 Durability on the phone**: An event the recorder reports as
recorded shall still be held after the phone closes the recorder one second
later. Verified by probe M-3, then a manual check on each platform.

**NFR-USE-005 Recording speed on the phone**: Recording an event of an existing
symptom with no note shall take no more than 2 taps after the first letters of
the symptom: the suggestion, then `Record`. Held as NFR-USE-001 is: by a test of
the path rather than a count of taps.

**NFR-USE-006 Contrast**: The recorder shall meet NFR-USE-003's contrast rule,
from the same colour tokens.

**NFR-USE-007 Touch targets**: Every control on the recorder shall present a
target of at least 44 by 44 CSS pixels (WCAG 2.2, 2.5.5). **Not yet met or
verified:** no test checks it and the recorder's buttons have not been
measured; the shared button rules set no minimum size. To be verified by a
structural test over the recorder's styles.

**NFR-PORT-001 Platforms**: The recorder shall run installed to the home
screen on Android in any browser offering WebAssembly, IndexedDB, service
workers, Web Locks and file sharing (D-1, Q-14, FR-104). No browser is named
as a requirement. Seen working end to end only in Chrome 157 on a Samsung
Galaxy Z Flip 5.

**NFR-MAINT-006 Coverage**: The recorder's rules and use cases shall be held at
100% statement coverage by the same gate as NFR-MAINT-001.

**NFR-MAINT-007 The file on both sides**: The recorder's file shall be held to
FR-051. The recorder writes it through the desktop's own export code
(`export.File.Encode`), so the two sides share one implementation; the
desktop reads it back (FR-093).

### 3.9 Won't this time

| Item | Reason |
|---|---|
| Other event types (medication, meal, sleep) | Source 14: version 1 stays on symptoms. The event table carries a kind column so they can follow without a migration of existing rows. |
| The recorder on iPhone | D-1: no device to measure it on. Every behaviour this document relies on (offline opening, storage that survives, sharing a file, telling a completed share from a cancelled one) was measured on Android alone; claiming it for iPhone would be a guess. |
| A PDF on the phone | D-4. The layout rules (FR-040 onwards) keep one home. |
| Events travelling from the desktop to the phone | D-3: one direction only. |
| Editing or deleting a sent event on the phone | D-3. |
| A live link between phone and desktop, over Wi-Fi or anything else | D-5; NFR-PRIV-001. |
| An app store listing | 4.1: the store category is promotional material that fixes intended purpose. A listing reopens Q-8. |
| Filters on the phone's list | The phone's list is a check that recording worked; the history with filters stays on the desktop. |
| Encryption at rest on the phone | NFR-PRIV-003's non-claim extends to the phone (NFR-PRIV-006). |
| Synchronisation between machines | Source 11: no cloud. The export is the portability path. |
| Reminders | Source 10: no nagging. |
| Charts and trends | Source 10: no interpretation. |
| Edit history (audit trail) | Source 13 allows one; left out of version 1 (Q-7). |
| Update check | No network (NFR-PRIV-001). |

### 3.10 Functional requirements: phone recorder (Amendment 21)

A recorder for Android, served from `symdiary.com/app/`, that records symptoms offline and hands them to the desktop as an export file. Every behaviour below that the phone's browser owns was measured on a real phone before it was written (Appendix A, M-2 to M-6). Each `Verified by` names tests that exist; where none does yet, the line says so.

Decisions taken by the owner on 2026-11-16:

| ID | Decision |
|---|---|
| D-1 | The recorder runs on Android first. iPhone waits until it can be measured: the owner has no iPhone, so none of M-2 to M-5 can be read there (revised 2026-11-16 from "iPhone and Android from the first release"). |
| D-2 | The phone keeps every event it records; nothing is removed after a send. |
| D-3 | Once an event has been sent, the desktop owns its corrections; the phone edits only what it has not sent. The export format stays at version 1. |
| D-4 | The PDF stays on the desktop; the phone does not draw one. |
| D-5 | No Wi-Fi link: the handover is a file moved by the phone's own sharing, so NFR-PRIV-001 stands unchanged on the desktop. |
| D-6 | A send carries only the events not yet sent; `Send again` repeats the last batch for a file that never arrived. Sending the whole phone record is not offered (Q-11). |
| D-7 | The recorder is served from `symdiary.com/app/`, by the site's own GitHub Pages from this repository. The site loads no script from another origin, so nothing but SymDiary's own code can reach the recorder's storage (Q-13, NFR-PRIV-005) (Oliver's ruling). |
| D-8 | The recorder lives in the SymDiary repository (Q-16). |
| D-9 | The recorder shares a `.txt` file holding export format version 1 unchanged; the desktop's Import dialog lists `*.txt` beside `*.json` (Q-17, M-4). |
| D-10 | The recorder runs SymDiary's own Go domain and application, compiled to WebAssembly; the rules have one home (Q-12, M-6). |

Priorities: 26 Must, 1 Should (FR-094), 2 Could (FR-084, FR-085). The Musts are the smallest recorder that records, keeps and hands over; the expensive parts (a PDF on the phone, two-way transfer, a live link) sit in 3.9.

#### Recording on the phone

**FR-076 Record on the phone**
- Priority: Must
- Requirement: When the user presses `Record` with a symptom entered, the
  recorder shall store one event whose occurrence time is the recording
  instant, unless the user has set another time.
- Acceptance: Given the phone's clock reads 2026-11-16T08:05:00+00:00, symptom
  `Tired` and no note, when `Record` is pressed, then the recorder's list shows
  16 Nov 2026 08:05, `Tired`, with no note and no severity.
- Verified by: `internal/application/record_test.go::TestRecordNowUsesTheClock`,
  the rule the phone runs; recording on the phone was checked by hand

**FR-077 The desktop's recording rules hold on the phone**
- Priority: Must
- Requirement: The recorder shall refuse every event the desktop refuses under
  FR-002 (blank symptom) and FR-006 (future occurrence), with the same wording.
- Rationale: one product, one set of rules; a refusal on one side that is
  accepted on the other teaches the user that the rule is negotiable.
- Verified by: `internal/facade/facade_test.go::TestRecordRefusals`, run over
  the phone's own store; there is one implementation of the rules (D-10)

**FR-078 Recording offline**
- Priority: Must
- Requirement: While the phone has no network connection, the recorder shall
  open from the home screen and record.
- Verified by: probe M-2, then a manual check on each platform in D-1.

**FR-079 Kept on the phone**
- Priority: Must
- Requirement: The recorder shall keep its events in the browser's storage for
  its own address, so they remain after the recorder is closed and after the
  phone restarts.
- Verified by: probe M-3;
  `internal/infrastructure/memstore/memstore_test.go::TestTheRecordSurvivesReopening`

**FR-080 Persistent storage asked for**
- Priority: Must
- Requirement: When the recorder first opens, the recorder shall ask the
  browser to keep its storage persistent.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "asks for persistent storage when it opens (FR-080)"

**FR-081 Persistent storage refused**
- Priority: Must
- Requirement: If the browser declines persistent storage, then the recorder
  shall say on its main screen that the phone may clear unsent events and that
  sending is how to keep them.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "says the events may be cleared when the browser will
  not keep them (FR-081)"

**FR-082 Unsent count**
- Priority: Must
- Requirement: The recorder shall show the number of events not yet sent,
  together with the date of the oldest of them.
- Rationale: browser storage can be cleared by the phone. The desktop holds the
  master record, so the loss is bounded by what has not been sent; the user
  needs to see that bound.
- Acceptance: Given three unsent events, the oldest recorded 12 Nov, the main
  screen reads `3 not yet sent, oldest 12 Nov 2026` followed by that event's
  time.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "names the unsent count and the oldest (FR-082)";
  `internal/application/outbox_test.go::TestWaitingCountsTheUnsentAndNamesTheOldest`

**FR-083 Symptom suggestions on the phone**
- Priority: Must
- Requirement: The recorder shall offer the symptoms it has already recorded,
  matched as FR-011 matches them and folded as FR-014 folds them.
- Verified by: the FR-011 and FR-014 tests (`TestSuggestionsMatchIgnoringCase`,
  `TestCaseVariantReusesDefinition`), since the phone runs the same rules

**FR-084 Symptoms seeded from the desktop**
- Priority: Could
- Requirement: When the user opens a desktop export file in the recorder, the
  recorder shall add the file's symptom definitions to its suggestions.
- Rationale: the same words on both sides keep the desktop's list from
  growing near-duplicates.
- **Not built in this release.** To be verified by a recorder test "an opened
  export adds its definitions"

**FR-085 Seeding takes no events**
- Priority: Could
- Requirement: When the recorder reads a desktop export file, the recorder
  shall store none of the file's events.
- Rationale: D-3, one direction only.
- **Not built in this release.** To be verified by a recorder test "an opened
  export adds no events"

**FR-100 The form before the rules**
- Priority: Must
- Requirement: While the Go rules are still loading, the recorder shall show
  the recording form, accept typing into it and show `Record` as disabled.
- Rationale: D-10. The rules took 496 ms to start on the measured phone (M-6);
  typing a symptom takes longer than that, so nobody waits if the form does not.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "takes typing but cannot record until Go has loaded
  (FR-100)"; `frontend/src/recorder/start.test.ts` "loads both scripts, names the zone and settles once Go is
  ready (FR-100)"

**FR-101 A failed write on the phone**
- Priority: Must
- Requirement: If the phone's storage cannot write an event, then the recorder
  shall keep everything entered in the form and shall say that the event was
  not saved, naming the reason.
- Rationale: FR-009's rule, on the phone.
- Verified by: `memstore_test.go::TestAFailedSaveChangesNothing`;
  `frontend/src/RecordPane.test.tsx` "keeps everything entered when the event
  was not saved", the form the phone shares

**FR-102 Stored data from a newer recorder**
- Priority: Must
- Requirement: If the phone's storage holds data written by a newer recorder
  than the one running, then the recorder shall refuse to open it and shall
  say so, rather than read it.
- Rationale: 4.4's rule for the desktop's own schema, on the phone.
- Verified by: `memstore_test.go::TestASnapshotThatCannotBeTrustedIsRefused`,
  its case for a newer schema

**FR-103 One copy at a time**
- Priority: Must
- Requirement: When the recorder opens while another copy is open on the same
  phone, the copy just opened shall show the recorder and the older copy shall
  pause and say so. When the user comes back to a paused copy, it shall reload
  itself and show the recorder, pausing the other in turn; while paused it
  shall also offer a Reload button. Where a copy cannot tell whether another is
  open, it shall start nothing and say so.
- Rationale: two copies writing the phone's storage could write over each
  other's events, so only one may write. The one the user is looking at is the
  one they mean to use (Oliver's ruling): a copy left open in a Chrome tab must
  never block the home screen icon; reopening the app must never leave the
  user on a stopped page with nothing to do.
- Verified by: `frontend/src/recorder/start.test.ts`, "takes the lock from any
  copy already holding it", "stops this copy once a newer one takes the lock
  away", "refuses when the lock cannot be asked for", "reloads once the page is
  brought back to the screen", "reloads when Chrome restores the page from its
  cache" and "never reloads while the page stays hidden or on a fresh load";
  `frontend/src/recorder/TakenOver.test.tsx`, "reloads the page when Reload is
  tapped"; checked by hand on the phone with a Chrome tab and the icon

**FR-104 The recorder checks its browser**
- Priority: Must
- Requirement: When the recorder opens in a browser lacking WebAssembly,
  IndexedDB, service workers, Web Locks or the ability to share a `.txt` file,
  it shall start nothing and shall say that this browser cannot run it, naming
  each thing missing and what it is for.
- Rationale: no browser is required by name (Oliver's ruling), so the check
  happens on the phone itself. Recording events a browser could not keep or
  send would lose them.
- Verified by: `frontend/src/recorder/capabilities.test.ts`, "finds nothing
  missing in a browser that has everything", "names only %s when it is missing"
  for each need, "counts sharing as missing when the browser says it
  cannot share a file" and "says the browser cannot run the recorder and names
  everything it lacks"

#### Correcting on the phone

**FR-086 Edit before sending**
- Priority: Must
- Requirement: While an event has not been sent, the recorder shall let the
  user change its symptom, occurrence time, severity and note.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "edits an unsent event (FR-086)"

**FR-087 Delete before sending**
- Priority: Must
- Requirement: While an event has not been sent, when the user asks to delete
  it, the recorder shall ask first, naming the event, then delete it only on
  confirmation.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "asks before deleting an unsent event, naming it
  (FR-087)"

**FR-088 Sent events are final on the phone**
- Priority: Must
- Requirement: If the user tries to edit or delete a sent event, then the
  recorder shall refuse, saying that a sent event is corrected on the desktop.
- Verified by: `frontend/src/recorder/RecorderApp.test.tsx` "refuses to delete a sent event before asking, saying
  where it is corrected (FR-088)" and "refuses to edit a sent event, keeping
  what was typed (FR-088)"; `internal/application/edit_sent_test.go::TestASentEventIsNeitherEditedNorDeleted`

#### Sending

**FR-089 Send**
- Priority: Must
- Requirement: When the user presses `Send`, the recorder shall write a file in
  export format version 1 (FR-051) holding every event not yet sent, then offer
  it to the phone's share sheet (D-6).
- Acceptance: Given five events of which two are sent, when `Send` is pressed,
  then the file holds the three unsent events and no others.
- Verified by: probe M-4; `frontend/src/recorder/SendPane.test.tsx` "shares a .txt file and marks it sent once the
  share completes (FR-089, FR-090)";
  `outbox_test.go::TestPrepareCarriesTheUnsentAndTheSymptomsTheyUse`

**FR-090 Sent only when the share completes**
- Priority: Must
- Requirement: When the share sheet reports the file shared, the recorder shall
  mark the batch's events as sent.
- Verified by: probe M-5; `frontend/src/recorder/SendPane.test.tsx` "shares a .txt file and marks it sent once the
  share completes (FR-089, FR-090)";
  `outbox_test.go::TestConfirmMarksTheBatchAndAgainRepeatsIt`

**FR-091 Share cancelled or failed**
- Priority: Must
- Requirement: If the share sheet is dismissed or reports a failure, then the
  recorder shall mark nothing as sent.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "marks nothing when the share sheet is dismissed (FR-091)"

**FR-092 Nothing to send**
- Priority: Must
- Requirement: While no event is unsent, the recorder shall show `Send` as
  disabled.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "offers no send with nothing waiting, nor a repeat before
  a first send (FR-092, FR-097)"

**FR-093 The desktop reads what the phone wrote**
- Priority: Must
- Requirement: The desktop application shall import a file the recorder wrote
  under FR-053, unchanged.
- Acceptance: Given a desktop record of 40 events and a recorder file of 3 new
  events, when the file is imported, then the desktop reports 3 added, 0
  skipped; importing it again reports 0 added, 3 skipped.
- Verified by: `export_test.go::TestAPhoneSendReadsAsText`, which writes
  through `export.File.Encode`, the code the recorder sends with, then reads
  the `.txt` back; the import and a repeat import skipping everything were
  checked by hand with a file sent from the phone.

**FR-094 File name**
- Priority: Should
- Requirement: The recorder shall name each file
  `SymDiary-phone-<date>-<time>.txt` in local time, so a reader can tell
  batches apart in a folder.
- Rationale: Chrome on Android refuses to share a `.json` file and shares a
  `.txt` one (M-4, D-9). The contents stay format version 1.
- Acceptance: a send at 16 Nov 2026 17:12 writes
  `SymDiary-phone-2026-11-16-1712.txt`.
- Verified by: `outbox_test.go::TestPrepareCarriesTheUnsentAndTheSymptomsTheyUse`;
  `internal/facade/outbox_test.go::TestThePhoneSendsWhatItRecorded`

**FR-098 The desktop's Import lists text files**
- Priority: Must
- Requirement: When the user opens the desktop application's Import dialog,
  the dialog shall list files ending `.txt` as well as files ending `.json`.
- Rationale: D-9. The reader decides by contents (the `format` and `version`
  envelope), so a `.txt` that is not a SymDiary record is refused as any foreign
  file is (FR-053).
- Acceptance: Given `SymDiary-phone-2026-11-16-1712.txt` in Downloads holding
  one event, when Import is opened, then the file is listed; when it is chosen,
  then the desktop reports one event imported.
- Verified by: `window_test.go::TestTheImportDialogListsTextAndJSON` and
  `export_test.go::TestAPhoneSendReadsAsText`; that the dialog shows both is
  checked by hand in a real build.

**FR-099 A foreign text file is refused**
- Priority: Must
- Requirement: If the user imports a `.txt` file that is not a SymDiary
  record, then the desktop application shall import nothing and shall say the
  file is not a SymDiary record, naming it.
- Verified by: `export_test.go::TestATextFileThatIsNotARecordIsRefusedByName`

**FR-095 Send again**
- Priority: Must
- Requirement: When the user confirms `Send again`, the recorder shall write a
  file holding the last batch, then offer it to the phone's share sheet.
- Rationale: FR-090 marks a batch sent once the share sheet completes, which is
  not proof the file reached the desktop (M-5). Without this, a lost file
  leaves its events marked sent with no way to send them.
- Acceptance: Given a last batch of three events, followed by two events
  recorded since, when `Send again` is confirmed, then the file holds the same
  three events and not the two.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "asks before sending the last batch again, naming the
  risk (FR-095, FR-096)"; `outbox_test.go::TestConfirmMarksTheBatchAndAgainRepeatsIt`

**FR-096 Send again warns first**
- Priority: Must
- Requirement: When the user presses `Send again`, the recorder shall ask
  first, saying that an event of that batch since corrected or deleted on the
  desktop comes back if the file is imported again.
- Rationale: the desktop matches events by all five fields
  (`internal/domain/event.go:47`), so the phone's original copy is a new event
  to it once the desktop's copy has changed.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "asks before sending the last batch again, naming the
  risk (FR-095, FR-096)"

**FR-097 No batch to repeat**
- Priority: Must
- Requirement: While no send has completed, the recorder shall show
  `Send again` as disabled.
- Verified by: `frontend/src/recorder/SendPane.test.tsx` "offers no send with nothing waiting, nor a repeat before
  a first send (FR-092, FR-097)"

## 4. Other requirements

### 4.1 Legal and regulatory

GPL-3.0. SymDiary makes no medical claim and performs no clinical function; the
design keeps it outside what UK MHRA guidance treats as software as a medical
device, which turns on intended medical purpose such as diagnosis, prevention,
prediction, prognosis, treatment or alleviation.

Owner's ruling, 2026-09-22 (Q-8): of those purposes only monitoring could be
argued at all, while SymDiary is a note-taking application. It stores what the user
typed and gives it back. None of the other purposes apply.

The ruling is what the product is built to stay inside, so it is a constraint on
every later change rather than a note about this version. The line is
interpretation: the moment SymDiary scores, trends, alerts, predicts or says two
events are related, the argument that it only takes notes is gone. Section 1.3
already puts every one of those out of scope; FR-042 holds the record to the
user's own entries alone.

Confirmed 2026-09-22 against the published text, which is what a regulator reads
rather than this document. Two sources, both read in full at the passages that
bear on SymDiary:

- The Medical Devices Regulations 2002 (SI 2002/618), regulation 2, as
  published on legislation.gov.uk and stated there to be current to 21
  September 2026. "Software" is named in the definition of a medical device;
  the qualifying purposes are diagnosis, prevention, monitoring, treatment or
  alleviation of disease, the injury and handicap limbs, investigation or
  modification of the anatomy and control of conception. "Intended purpose" is
  what the manufacturer supplies "on the labelling, the instructions for use
  and/or the promotional materials", so the README, the Guide, About and the
  store listings are the text that settles it, not this document.
- MHRA, "Medical device stand-alone software including apps (including
  IVDMDs)" v1.10f, last updated 1 July 2023. Page 21 (Monitoring) lists, among
  the examples unlikely to be devices, software that simply replaces a written
  diary or log of symptoms that can be used when consulting with the patient's
  doctor. That is SymDiary's intended purpose. Page 10 puts software that stores
  medical data without change on the non-medical side of the medical-purpose
  chart; it states as well that an electronic patient record simply replacing a
  paper file does not meet the definition. Appendix 1 (symptom checkers)
  governs software that matches entered symptoms to conditions, which SymDiary
  does not do, so it does not apply.

The ruling therefore stands on the nearest published example there is. The
constraint on every later change is the guidance's own caveat, in its words:
"the addition of features that enhance the data presented may bring it into the
remit of the UK MDR 2002". That is broader than the list above, because it
reaches presentation and not only inference.

Two further readings taken from the same text and held as standing rules:

- A disclaimer carries no weight on its own. Page 11: a general disclaimer such
  as "this product is not a medical device" is not acceptable where medical
  claims are made elsewhere in the labelling or promotional literature. The
  README's closing line is accurate because nothing else makes a claim; it is
  not what keeps SymDiary outside the definition.
- The store category is promotional material. Page 11 names the app store
  description and category among the materials that fix intended purpose, so
  SymDiary is never listed under a medical or health category. The Flatpak
  desktop entry is `Utility;Office;` and the macOS bundle sets no category.

This is a documented reading of the published text, not regulatory advice. A
borderline determination can be put to the MHRA directly if one is ever wanted.

### 4.2 Internationalisation

English only in version 1. Dates print in the day-month order of the source
example (`22 Sep 17:12`).

### 4.3 Risk

Two harms matter: loss or silent corruption of the record; a receipt that
misstates it to a doctor. Rather than a full FMEA, three
requirements carry it: FR-024 (time never silently changed), FR-042 (receipt
holds nothing but the record) and NFR-REL-001 (durability). Judged
proportionate for a single-user record keeper.

### 4.4 What version 1 commits to

A first release is where a promise starts, so the promises are named rather
than left to be inferred from the code as it happens to stand.

- **The export format is a contract, not a snapshot.** Every file SymDiary has
  ever written stays readable by every later SymDiary. The envelope (`format`
  and an integer `version`) never changes shape; a new version is added beside
  the old readers rather than in place of them, with the old version's real
  bytes frozen under `testdata` and read by the suite on every run. This is
  what makes an export a file the user owns: it is worth nothing if the program
  that wrote it is the only one that will ever read it, then worth nothing again
  if next year's SymDiary refuses last year's file.
- **The printed sheet's wording is a public claim** (FR-045). It goes into
  filing cabinets and is read by people who will never see the application, so
  the line saying the record is not a diagnosis is held by a test asserting it
  whole; changing it is a decision about what SymDiary tells a clinician rather
  than a wording tidy-up.
- **The record's own schema** is versioned by the store and a newer schema is
  refused rather than guessed at, so a record written by a later SymDiary is
  never silently misread by this one.

What version 1 does NOT commit to: the window's layout, the wording of anything on
screen other than the sheet, the log's shape, the setup program's screens or
any internal structure. Those change whenever they are improved.

## 5. Appendices

### Appendix A: Feasibility measurements

| ID | Question | Measured |
|---|---|---|
| M-1 | Can the application draw the record itself, so that it reads the same on every desktop? | Yes; it had to. Measured 2026-09-22 in the built window, `window.print()` opened the Windows print dialog and printed the receipt alone; measured again on all three desktops, the same page printed three different ways, with the browser's own header and footer reaching the paper unless the reader had turned them off. Since Amendment 16 the document is drawn in Go. Measured 2026-09-23 by writing a record and reading the text positions back out of it: A4, `Page N of M` on every page in order, the first line 19.3mm to 19.8mm from the top of every page, no event split across two sheets and nothing on the paper the record does not hold. |
| M-2 | Does an installed recorder open and record with the network off, on iPhone and on Android? | **Measured 2026-11-16, Android, Chrome 157: yes.** Installed through the page's own install offer (`beforeinstallprompt`), reopened in airplane mode, an event was recorded with the note "Offline, installed, probe-2", shared later as `.txt`, then imported into the desktop: the second import of the same file reported "Imported 0 events. Skipped 2 events already in your record." (FR-093's acceptance, end to end). iPhone untested (no device). |
| M-3 | Is persistent storage granted on each? What does iPhone do to an installed web app's storage after weeks unused? | **Measured 2026-11-16, Android, Chrome 157:** the installed probe reads "Persisted: true", 7,789,913 bytes used of 10,745,208,153. Not measured: whether events survive weeks unused. That half is assumption A-5 rather than a blocker; FR-081 and FR-082 bound the loss to unsent events either way. |
| M-4 | Can the recorder hand a JSON file to the share sheet on each? Which routes reach a Windows desktop, a Mac and a Linux desktop? | **Partly measured 2026-11-16, Android, Chrome 157:** `navigator.share` and `canShare` both present; `canShare` answered true for every variant, yet the share of a `.json` file was refused within 8 ms with `NotAllowedError: Permission denied`, whether declared `application/json` or `text/plain`. The same contents as `.txt`, declared `text/plain`, opened the share sheet. The extension decides; `canShare` does not predict it. Still open: which routes reach each desktop; iPhone untested (no device). |
| M-5 | Does the share call tell a completed share from a cancelled one on each? | **Measured 2026-11-16, Android (Galaxy Z Flip 5), Chrome 157: yes.** A share completed through Gmail resolved after 9,646 ms; a share dismissed without choosing an app rejected after 1,914 ms with `AbortError: Share canceled`. FR-090 and FR-091 stand as written. Resolving means the receiving app took the file, not that it reached the desktop, so FR-095 stays the recovery for a file lost after that. iPhone untested (no device). |
| M-6 | Does `internal/domain` build for WebAssembly? At what download size? | **Measured 2026-11-16: yes.** `internal/domain` and `internal/application` build for `js/wasm` under Go 1.26.3. Linked together: 2,765,198 bytes, 791,305 at gzip -9, plus Go's 16,992 byte `wasm_exec.js`. Loaded as base64 inside a script file, since `connect-src 'none'` forbids fetching it, which needs `'wasm-unsafe-eval'` in `script-src`. Started in 130 ms in desktop Chromium; 496 ms on the Galaxy Z Flip 5 in Chrome 157, where `KeyOf("  Tired  ")` answered `"tired"` and an empty record was refused with Go's own wording. |

### Appendix B: Open questions

Decided by the owner on 2026-09-22:

| ID | Question | Decision | Lives in |
|---|---|---|---|
| Q-1 | Severity list | `Mild`, `Moderate`, `Severe`. | FR-004 |
| Q-2 | `tired` typed when `Tired` exists | Same symptom; the existing label is kept. | FR-014 |
| Q-3 | Rename a definition | Yes, Should. | FR-028 |
| Q-4 | Receipt order | Groups by first occurrence, oldest first; events oldest to newest; never alphabetical. | FR-044 |
| Q-5 | Import in version 1 | Yes, Should. | FR-053 |
| Q-6 | Encryption at rest | No; the non-claim is stated. | NFR-PRIV-003 |
| Q-7 | Edit history | Not in version 1. | 3.9 |
| Q-9 | Light, dark or following Windows? | Neither: SymDiary opens dark and carries a button that moves it to light, remembered for next time. Answered first as "follow the Windows app mode", then superseded the same day by Amendment 12. | FR-073 |
| Q-8 | Does the MHRA reading in 4.1 hold? | Yes. Ruled 2026-09-22, then confirmed the same day against SI 2002/618 reg. 2 and MHRA guidance v1.10f, page 21, which lists a replacement for a written symptom diary among the examples unlikely to be devices. | 4.1 |
| Q-10 | Identifying details on the receipt | None. | FR-041 |

Nothing is still open. Every question raised while this specification was
written has been answered; each answer lives in the requirement or the file
named beside it.

Decided by the owner on 2026-11-16 (Amendment 21):

| ID | Question | Decision | Owner |
|---|---|---|---|
| Q-11 | What does a send carry? | **Decided 2026-11-16 (D-6):** only the events not yet sent, plus `Send again` for the last batch. Reason, read from the source the same day: the desktop matches events by all five fields (`internal/domain/event.go:47`), so with D-2 a whole-record send after a desktop correction adds the phone's original copy back as a second event, while a deletion on the desktop is undone. The residual risk sits on `Send again` alone, which warns first (FR-096). Lives in FR-089, FR-095 to FR-097. | Oliver |
| Q-12 | One home for the recording rules: the Go domain compiled to WebAssembly versus a TypeScript copy held equal by shared case tables? | **Decided 2026-11-16 (D-10): the Go domain and application compiled to WebAssembly.** The house rule settles it: a second copy of the rules is a defect, not a trade-off. M-6 shows it works on the phone. Its 496 ms start becomes a constraint on the design: the form appears at once and Go loads behind it (FR-100). | Oliver |
| Q-13 | Where is the recorder served from? | **Reopened 2026-11-16.** Answered "under symdiary.com", then a measurement contradicted it: every page of the site (`docs/index.html`, `features.html`, `why.html`, `download.html`) loads GoatCounter's `count.js` from `gc.zgo.at`. Browser storage belongs to an origin rather than a path (general platform knowledge, not measured here), so a recorder at `symdiary.com/recorder/` shares its events with pages running a third-party script, which NFR-PRIV-005 forbids. **Decided 2026-11-17: `symdiary.com/app/`, with GoatCounter removed from the site** (Oliver's ruling: the recorder holds personal data). One repository on GitHub Pages alone rules out a subdomain: GitHub's documentation gives a Pages site one custom domain and lists an apex domain paired with a custom subdomain as unsupported. | Oliver |
| Q-14 | Minimum phone versions? | **Decided 2026-11-17: no browser floor.** Any Android browser passing the recorder's own check (FR-104) is supported; the site names no browser as a requirement. Chrome 157 is the one seen working end to end. Supersedes the 2026-11-16 answer of Chrome 157. | Oliver |
| Q-15 | Does a phone recorder change 4.1's MHRA reading? | **Decided 2026-11-16:** no. The intended purpose is unchanged (a written symptom diary's replacement, MHRA v1.10f page 21). The standing rules carry over: no store listing, no health category, nothing that enhances the data presented. | Oliver |
| Q-17 | Chrome on Android shares `.txt` but not `.json` (M-4). How does the file reach the desktop? | **Decided 2026-11-16 (D-9): share as `.txt`.** Lives in FR-094, FR-098, FR-099. As proposed: the recorder shares `SymDiary-phone-<date>-<time>.txt` holding format version 1 unchanged; the desktop's Import dialog lists `*.txt` beside `*.json`. The reader already decides by contents (the `format` and `version` envelope), so a `.txt` that is not a SymDiary record is refused as any foreign file is. Cost: a desktop change, plus a name that no longer says what the file holds, which the desktop has avoided before (`app.go:25`). Alternative: keep `.json` through `Download the file instead` only, leaving the user to move it from the phone's Downloads folder by hand. | Oliver |
| Q-16 | Same repository or a new one? | **Decided 2026-11-16:** the same. FR-093 and NFR-MAINT-007 tie the two sides together, so one gate sees both. | Oliver |

### Appendix C: Build order

1. Domain: event, symptom definition, severity, occurrence rules (FR-002,
   FR-006), suggestion matching, filters, receipt model.
2. Application: record, edit, delete, history, receipt, export, with fakes.
3. Infrastructure: SQLite store with schema versions, export file, run log,
   single-instance lock.
4. UI: Wails facade plus the React frontend; M-1 probe before the receipt page.
5. Setup program, ported from ED Voyage Companion.
6. Docs: README, ARCHITECTURE, TESTING, DEVELOPMENT.

The phone recorder (Amendment 21), in the same order:

1. Domain and application: unchanged, built for `js/wasm` on every gate run,
   so a change the phone cannot run fails at once.
2. Desktop: FR-098 and FR-099, the Import dialog listing `.txt`.
3. Infrastructure on the phone: the browser storage behind the application's
   store port, with its schema version (FR-102); the file the share sheet is
   given; one copy at a time (FR-103).
4. UI on the phone: the recording form, the list, `Send` and `Send again`.
5. Deploy to `symdiary.com/app/`.

## 6. Amendments

| No. | Date | Requirement | Change | Reason |
|---|---|---|---|---|
| 23 | 2026-11-17 | 2.3, NFR-PORT-001, Q-14, new FR-104 | The recorder no longer requires Chrome. It checks on opening for each thing it needs; where the browser lacks any, it starts nothing and names what is missing. | Owner's ruling: a product cannot force a browser on people. Only Chrome 157 was ever measured, which made it the stated requirement; the check moves that judgement onto each phone, so a browser nobody here has tried is told plainly whether it can run the recorder. |
| 22 | 2026-11-17 | FR-103 | A copy of the recorder that a newer one paused reloads itself when the user comes back to it, taking the recorder back; while paused it offers a Reload button. | Owner's report from the phone: reopening the app showed the paused copy's message again. Only pulling the page down to reload brought the recorder back; nothing told the user so. Measured by the owner: a reload restores it. The page now does that reload itself the moment it is back on screen. |
| 21 | 2026-11-16 | 1.3, 1.4, 2.1, 2.3, C-9, A-5, new 3.10 (FR-076 to FR-103), new NFR-PRIV-004 to 006, NFR-REL-002, NFR-USE-005 to 007, NFR-PORT-001, NFR-MAINT-006 and 007, 3.9, Appendices A to C | A phone recorder for Android joins the product. It records offline, keeps every event, edits only what it has not sent and hands unsent events to the desktop as a `.txt` file holding export format version 1 through the share sheet; the desktop's Import lists `.txt` beside `.json`. It runs the desktop's own Go rules compiled to WebAssembly, served from `symdiary.com/app/`; the site stops loading GoatCounter. The 3.9 row "Phone or web version" goes; iPhone joins 3.9. Synchronisation between machines stays out of scope. | Owner's request: symptoms are noticed away from the desk, so the recorder belongs where they are noticed. Every behaviour the phone's browser owns was measured on the owner's Galaxy Z Flip 5 in Chrome 157 before it was written (M-2 to M-6). Two of those measurements changed the design: Chrome refuses to share `.json` whatever it is declared as, so the file is `.txt` (D-9); the desktop matches events on all five fields, so a send carries only unsent events or a correction made on the desktop would come back (D-6). A Wi-Fi link was rejected because it would break NFR-PRIV-001 and needs a certificate the phone trusts. The counter goes because any script on a page of symdiary.com can read the recorder's storage, which holds personal data. A subdomain would have kept them apart; GitHub Pages gives one repository one custom domain. |
| 20 | 2026-10-03 | FR-040, FR-041, FR-053 | Five corrections from an audit of the shipped 1.2.0. (1) An event too tall for any page continues onto the next, which opens with the event's time again; an event that fits is still never split. (2) A note's line breaks reach the sheet. (3) A word wider than the line is broken at the line. (4) Save PDF refuses, naming them, characters the carried typeface cannot print, before asking where to save. (5) An import refuses a symptom kept with no events whose name is blank, as it already refused a blank event. Separately, an import made while the record or the history is showing now refreshes it. | Each was measured before it was changed. A single note of 5,059 characters ran off the foot of its page while the document reported success; a note's line breaks printed as one line; a 220 character address ran 375mm wide on a 210mm sheet; the typeface has no glyph for Chinese, Japanese, Korean, Arabic, Hebrew, Devanagari or emoji, which printed as empty boxes; a blank symptom name was imported as a blank entry. The refresh closes the path Amendment 19 left: an import from the band under a record on screen. Owner's rulings the same day: continue rather than shrink, refuse rather than carry a font for every script, break rather than leave whole; an import keeps future times as the file states them (DECISIONS-TRADEOFFS.md). Verified by `internal/infrastructure/pdf/layout_test.go::TestAnEventTallerThanAPageContinuesWithItsTimeRepeated`, `TestANoteKeepsItsLineBreaks`, `TestAWordWiderThanTheLineIsBrokenAtTheLine`; `sheet_test.go::TestCharactersTheTypefaceCannotPrintAreRefusedByName`, `TestTheTypefacePrintsTheEuropeanScripts`; `pdf_test.go::TestSavePDFRefusesTextItCannotPrintBeforeAskingWhereToSave`; `export_test.go::TestReadRefusals` ("blank definition"); `frontend/src/App.test.tsx`, "shows an imported event in the record already on screen" and "... in the history already on screen". |
| 19 | 2026-10-03 | FR-040, FR-041, FR-042, FR-046 | The sheet names the last appointment, `Last appointment: 02 Sep 2026`, under the date range whenever the range covers it; a `Show on the sheet` box, ticked by default, lets the reader leave it off. Once the record is showing, changing the range refreshes it. | Owner's request, superseding Amendment 18's ruling that the sheet never mentions the appointment. The reader of the sheet benefits from knowing where the last consultation falls: what has already been heard and what is new is the first thing a doctor asks of the record; the comparison is the doctor's to make. The rule is the range rather than how it was chosen, so a custom range spanning the appointment names it too, while a range that does not cover it names nothing, where the line would be noise. The line sits beside the range and never among the events, so the software arranges nothing around it (4.1). The opt-out is for a reader who would rather the date not reach paper. The refresh answers a defect the owner found: the record on screen stayed on the old range while Save PDF wrote the new one, so the two could differ. |
| 18 | 2026-10-03 | New FR-046, FR-068 | The receipt pane holds the date of the last appointment and offers `Since last appointment` beside `Custom dates`. The Guide says so. | Owner's request. The product exists for the walk from recording to the appointment; the date of the last one is the boundary a person reaches for. Typing it again each time is friction with nothing to show for it. Owner's rulings the same day: the date lives in the window's storage rather than the record, so the record stays observations alone and the export format stays at version 1; the appointment's own day is included, since an event seen twice costs the reader nothing while an afternoon event left off costs something; one date, today or earlier, so it stays a boundary rather than becoming a calendar. The sheet is unchanged: the MHRA caveat in 4.1 concerns features that enhance the data presented; the data presented is identical. |
| 17 | 2026-10-02 | Section 2.3 | Windows 10 joins Windows 11 as a supported system, 64-bit on both. Where the WebView2 runtime is missing or older than 94.0.992.31, the application and the setup program each offer to download and install it through Microsoft's bootstrapper before their window opens; declining leaves them unopened. | Owner's decision: the download page says Windows 10 or 11. Windows 10 does not always carry WebView2, which Windows 11 does. The behaviour above is Wails' default `download` strategy, read from the v2.12.0 source (`internal/wv2installer`); `build.ps1` chooses no other. It has not yet been run on a Windows 10 machine. |
| 16 | 2026-09-23 | FR-040, FR-045 | The record is no longer printed through the window's browser. The application draws it as a PDF and the reader chooses where it goes; the button reads Save PDF. | Owner's decision after the browser path was measured on all three desktops and gave three different answers. A page belongs to whichever engine the desktop ships, so what came off the paper depended on that engine, on whether the reader had "Headers and footers" ticked in their print dialog and on which of the CSS the sheet leaned on that engine had implemented. Windows was made correct and measured; Linux and macOS were, in the owner's words, a mess. The record is the product, so it is drawn once, in Go, where the same bytes reach every reader. What that buys beyond consistency: pagination that never splits an event across two sheets, a page number on every page, a document the suite can measure rather than one only a printer can. The typeface is the Go fonts, carried inside the binary, because a PDF's built-in fonts can say nothing outside Latin-1 and a note holding a curly quote or an accented name would have reached a doctor with the user's own words mangled. |
| 15 | 2026-09-23 | FR-045 | Every printed page carries `Page N of M` at its foot. | Owner's request. A page counter can live nowhere but an @page margin box, so the foot of the sheet takes the one page margin there is while the top and sides keep none. Measured through the engine WebView2 runs, driving the print with the browser's own header and footer switched ON, which the command line cannot ask for and the devtools protocol can: a ten page record printed `Page 1 of 10` through `Page 10 of 10` and carried none of the browser's date, address or page count. Declaring a margin box is what suppresses them; the same record with no margin at all prints neither. Also measured: `counter(page)` resolves to 0 anywhere but a margin box, in the table foot and in a fixed element alike, so there is no way to number the pages of a sheet with no margin. The paper size is still left alone: that ten page print came off US Letter. |
| 14 | 2026-09-23 | FR-045 | The line naming the program and its address prints once, above the title; the copy below the last event goes. The printed sheet carries nothing of the browser's: no page margin, with the band the paper needs reserved by the sheet's own repeating head and foot. The record is set across the full width of the page rather than the reading column the screen uses. | Owner's ruling after a real test print. The letterhead at the top already says what produced the sheet, so repeating it at the foot said nothing twice. The same print carried the date, the window title, `wails.localhost` and a page number, none of them ours; a browser draws that furniture inside the page margin and a page can reach it no other way, so the margin goes. Padding cannot replace it: padding is applied once to the element, not once per sheet; a print made that way came off with the first line of page two sliced through. A thead and a tfoot ARE laid out again on every page, so two empty rows hold the record clear of the top and bottom edges of every sheet. Measured against the engine WebView2 runs, by printing a record to PDF and reading back where the text landed: over ten pages the first baseline sits between 17.7mm and 18.3mm from the top of every one and the last no closer than 17.2mm to the foot. Proved by planting: with the gutter removed the same record prints 3.4mm from the top, which is the sliced line that was photographed; at 40mm it prints 43.4mm down on all three pages. |
| 13 | 2026-09-23 | FR-045, FR-072, new FR-074, new FR-075 | The printed sheet carries the application's icon beside the first provenance line. The setup program's licence screen explains the licence in ordinary words before showing its text in full, in a pane that reads itself down. The licence text gets one home in `internal/licence`, guarded against the published LICENSE byte for byte. FR-072's promise is covered by tests rather than by a manual check alone. | Owner's request on all three. The licence screen named GPL-3.0 and left it there, which tells somebody installing a program nothing about what they may do with it. The mark on the sheet identifies the producer and presents no data, so it stays inside the reading Q-8 closed against. A copy of the licence now travels with the program on all three platforms, which it did not before this pass: the install folder held the executable alone; neither the Flatpak nor the bundle carried the text, so nothing conveyed gave the recipient the copy section 4 asks for. Showing it on a setup screen alone would not have answered that either, which is why the text is displayed and shipped rather than linked to. Measured: the pane's descent was timed at 4007ms of travel against the 50px the house pace owes; the pane drew no ring at rest, after Tab focus or after a click. |
| 12 | 2026-09-22 | New FR-073, Q-9, 3.9, NFR-USE-003 | SymDiary opens dark and carries a light or dark button in the bar, left of Guide between two rules; the setup program carries the same in its header. Neither reads the Windows app mode any more, so the Windows theme read and the `prefersDark` field on setup's state are gone. | Owner's request, superseding the Q-9 answer of the same day. The window following the desktop means it changes under the reader at dusk; a button is one press away and what it chooses is remembered. Dark is what it opens in because that is what the owner wants to see first. Measured: the setup package's coverage rose from 56.9% to 59.2% once the registry read no test could exercise was gone; the floor moved with it. |
| 11 | 2026-09-22 | NFR-USE-002 | The window hands the page the keyboard as it opens, through a focuser seam wired at the composition root and called on DOM ready. | A defect the owner found in the built window: no Tab ever stepped the ring, while a single click on the page fixed it for the rest of the run. WebView2 keeps DOM focus and keyboard focus apart, so the sink held the first while the webview held none of the second and no keydown reached the listener at all. Asking the Wails runtime to show the window is NOT enough, which is how the first attempt missed: that focuses the MAIN window, while WebView2 hosts the page in a child window of class `Chrome_WidgetWin_1` and the keys follow the child. The answer is ported from PigeonPost, where it is measured and works: find this process's visible window by its title, enumerate its children for that class, attach to that window's input thread, bring the window to the foreground and set focus on the child. The page cannot fix this from its own side, since a DOM focus call cannot make the webview the thing keys are sent to. |
| 10 | 2026-09-22 | FR-041, FR-042, new FR-045 | The printed record is framed: the program and its address above the title and below the last event, then a statement under the date range saying it is not a diagnosis but the person's own notes for a healthcare professional. | Owner's request. It also answers the surface Amendment 9 identified as the one a regulator reads: a sheet that leaves the recording program's hand carries what it is and what it is not, rather than relying on a reader who has seen the application. The sheet says notes rather than guidance, since guidance names a purpose SymDiary does not have (FR-045). |
| 9 | 2026-09-22 | 4.1, Q-8 | The MHRA reading is confirmed against SI 2002/618 reg. 2 and MHRA guidance v1.10f, so Q-8 closes. The guidance's own caveat about features that enhance the data presented replaces the narrower list as the standing constraint. Two further rules are recorded: a disclaimer carries no weight on its own; the store category is promotional material. | The ruling had been made against a summary of the guidance rather than its text, while intended purpose is fixed by what the manufacturer publishes, so the text was the only thing that could settle it. Measured: page 21 lists a replacement for a written symptom diary among the examples unlikely to be devices, which is SymDiary's intended purpose. |
| 8 | 2026-09-22 | NFR-USE-002, new NFR-USE-004, FR-066 | The house keyboard model and its three-state focus ring are stated as requirements rather than left to the page; About names the copyright holder and year. | Owner's request. The ring was a single blue outline on keyboard focus alone, which said nothing about what could be used and nothing about what could not: Print sat inert beside the button that fills it with no way to tell it apart from a control waiting to be pressed. |
| 7 | 2026-09-22 | 1.3 scope, 2.3 operating environment, FR-060, FR-065 | Linux and macOS leave the deferred list and become part of this version: a Flatpak and a signed DMG, ported from PigeonPost's. Only the bespoke setup program stays Windows-only. | Owner's decision, superseding Amendment 3, which had made them a later version. Measured: one package stopped the module building elsewhere, the run log, whose Windows handle work is now behind a build tag and whose folder rule is a pure function taking the platform as an argument, so all three answers are exercised on every platform. The gate now builds and vets for Linux and macOS on every run, which is what keeps this true; a planted Windows-only import was refused by name. |
| 6 | 2026-09-22 | New FR-069 | The bar carries a Donate button, last in its right-hand group. | Owner's request. It takes a seat in the bar the window already has rather than a band of its own, as AudioDeck's does, since a whole new strip of chrome carrying one control costs more than it buys. The address lives once, in Go's product package; the page asks for the donation page rather than naming one, so nothing arriving from the page has to be checked before it is opened. |
| 5 | 2026-09-22 | New FR-068 | The application carries a Guide, reached from the bar. | Owner's request, in line with PigeonPost and ClearBudget. Its words are one document (`frontend/src/guideContent.ts`) and the dialog only draws them, as PigeonPost's does. |
| 4 | 2026-09-22 | FR-050, FR-053 | Both file dialogs open in the user's Downloads folder, which the user may leave. | Owner's request: it is where a person already looks for files they have saved. Measured: Wails' dialog options carry a default directory; the folder is checked before it is named, so a machine without one falls back to the dialog's own choice. |
| 3 | 2026-09-22 | 1.3 scope | Linux and macOS leave the permanent out-of-scope list and become planned for a later version: a Flatpak and a DMG builder, following PigeonPost's. | Owner's decision. Version 1 stays Windows only; nothing in the domain, application or page is Windows-specific, so the work is packaging plus the run log's standard-handle code. |
| 2 | 2026-09-22 | FR-041 | Receipt events name their year when the range crosses a year; they leave it out otherwise. | Found while writing the tests: the source example's `22 Sep 17:12` is unambiguous inside one year and ambiguous across two, which a doctor reading the record cannot resolve. |
| 1 | 2026-09-22 | FR-011 | The acceptance example's second case is corrected: typing `a` lists `Headache` and `Back pain`, not all three. | The baselined example was wrong: `Tired` holds no letter `a`. Found by the test that was written from it. |
