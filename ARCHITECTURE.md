# SymDiary: architecture

SymDiary records observations and retrieves them. Everything below follows from
one rule: **record what happened; do not decide what it means.**

## The invariants

Each is enforced by a test, named beside it. Every one was proved to bite by
planting a violation and reading the exit code.

| # | Invariant | Enforced by |
|---|---|---|
| 1 | The domain depends on nothing: no other layer, no framework. | `tests/structural/boundary_test.go::TestDomainHasNoOutwardImports` |
| 2 | The domain performs no IO and never reads the wall clock. Time arrives as an argument. | `TestDomainIsPure` |
| 3 | The application layer depends on the domain and on the ports it declares, never on infrastructure or Wails. | `TestApplicationDoesNotImportInfrastructure` |
| 4 | Only `main.go` wires infrastructure to the application. | `TestCompositionRootIsWhitelisted` |
| 5 | No source file exceeds 400 lines; none sits in the 381 to 400 danger band. | `TestNoFileExceedsLineLimit`, `TestNoFileInDangerBand` |
| 6 | SymDiary's own code opens no network connection: no `net` package in any of its packages, `connect-src 'none'` on the page. The one fetch is Wails' own, before the window opens, on a Windows machine missing WebView2 (REQUIREMENTS.md, Amendment 17). | `TestNoNetworkImports`, `TestThePageOpensNoConnection` |
| 7 | The wire is stated twice, in Go and in TypeScript; the two agree field for field; the receipt's line kinds agree name for name with the page's union and with the document's styles. | `tests/structural/wire_test.go`, `tests/structural/linekind_test.go` |
| 8 | Every colour lives in `frontend/src/theme.css`; every text pairing meets WCAG 2.2 AA in both modes. | `tests/structural/colours_test.go` |
| 9 | Every exported type carries a doc comment. | `TestEveryExportedTypeIsDocumented` |
| 10 | The ring belongs to a control: no container carries a ring rule or a tabindex; a surface made to scroll turns the engine's own ring off. | `tests/structural/focus_test.go::TestNoRingRuleNamesAContainer`, `TestNoContainerTakesFocus`, `TestEveryScrollingSurfaceSuppressesTheNativeRing` |
| 11 | A dialog body that scrolls pins its action row beneath it and wears the self-reading cycle. | `TestEveryScrollingDialogPinsItsActionsAndReadsItself` |
| 12 | A control that rings while it is usable says so while it is not: the ring is green on hover or focus, permanently red while disabled. | `TestEveryRingedControlSaysWhenItIsInert` |
| 13 | Each ring colour meets the 3:1 WCAG asks of a non-text indicator, in both modes. | `tests/structural/colours_test.go::TestEveryRingIsVisibleAgainstWhatItIsDrawnOn` |
| 14 | Every place that writes to the log is declared, so FR-065's promise that the log holds no part of the record stays checkable. A new writer fails until it is declared, which is the moment to ask what it puts in the file. | `tests/structural/logscan_test.go::TestOnlyTheKnownPlacesWriteToTheLog` |

## The layers

```text
UI (root package, frontend/)  →  Application  →  Domain  ←  Infrastructure
```

### Domain: `internal/domain`

The rules, as pure types and functions. No IO, no clock, no framework.

| File | What it owns |
|---|---|
| `symptom.go` | The symptom label and its comparison key: how a typed name is matched to one already held, without ever changing the label. |
| `severity.go` | The three severities and the absence of one. Names live here once. |
| `event.go` | The event, the future-occurrence rule and what makes two events the same observation. |
| `definition.go` | Reusable symptoms: matching, suggestion order and the rename checks. |
| `dates.go` | Civil dates, the two wire time formats and the display format. |
| `filter.go` | The history filter and the newest-first order. |
| `receipt.go` | The receipt: grouping, ordering and the exact lines it holds. |

The receipt is domain code because its wording is a rule, not decoration: it is
what stops a printed record saying anything the user did not record. The page
receives lines with a kind and draws them; it composes no sentences of its own.

The sheet's framing (FR-045) is the one exception to that; it is handed in
rather than read. The words name the product and its address and say what the
sheet is; they live in `internal/product`, which the domain may not import
because the domain depends on nothing. So `Receipt.Lines` takes a `Framing` and
writes it where it belongs, while the facade fills that struct from `product`.
The domain stays pure, the words keep one home and the sheet still carries
nothing the domain did not write.

The application's mark is drawn beside the opening framing line and nowhere
else (FR-045). It is a presentation decision, not the domain's: the receipt
stays a list of lines with a kind; both the window and the document decide
for themselves that the first line, where it is a provenance one, is a
letterhead.

### Application: `internal/application`

One service per user-visible action, over the ports in `ports.go`.

| File | What it owns |
|---|---|
| `ports.go` | `Store`, `RecordFile` and `Clock`: everything the layer needs from outside. |
| `record.go` | Recording, including which instant an event gets. |
| `edit.go` | Editing, the deletion confirmation's wording and deletion. |
| `history.go` | Listing, suggesting, the symptom list, renaming and building the receipt. |
| `transfer.go` | Export; import too, skipping what is already held. |
| `about.go` | The About content and the credits. |
| `errors.go` | The refusals, each wrapped round the reason it happened. |

### Infrastructure: `internal/infrastructure`

| Package | What it implements |
|---|---|
| `store` | The SQLite record, one transaction per write, plus `Unavailable`, which stands in when the file could not be opened so that every action still answers. |
| `export` | The JSON export file: a versioned format, an atomic write and a distrustful read. |
| `runlog` | The run log, plus pointing the process's error output at it before anything can fail. Ported from Bridge Talk. |
| `pdf` | The document a reader takes to an appointment: a pure layout answering the rows of each page, then a renderer that turns a row into ink. |
| `windowfocus` | Handing the page the keyboard as the window opens: on Windows it finds the child window WebView2 draws the page in and focuses it; elsewhere it does nothing. |
| `setup` | The per-user install policy: the paths, the fenced payload extraction, the version comparison, the registry entry, the shortcuts and the process work. The setup program is a facade over it and owns no install logic. |

### The document: `internal/infrastructure/pdf`

What a reader takes to an appointment. It is handed the receipt's lines and a
path; it answers a file.

SymDiary used to hand the record to the window's browser engine to print. That
engine is a different one on each desktop, so the same record came off the
paper three different ways, with what the reader got depending also on a
tickbox in their own print dialog. The record is the product, so it is drawn once here
(Amendment 16).

The layout is a pure function in `layout.go`: given the lines, the page
geometry and a Measurer, it answers the rows of each page. Nothing in it opens
a file or knows what a PDF is, so what lands on which page is settled by tests
that draw nothing. `sheet.go` is the renderer around it and owns no layout
decisions beyond turning a row into ink. Two rules are worth stating because
they are what a browser could not be made to keep: an event is never split
across two sheets; every page says which page it is.

### UI: the root package and `frontend/`

`app.go` and `actions.go` are the facade: one bound method per action, each
converting shapes, calling one service and converting back. `dto.go` holds the
wire shapes. `window.go` holds the file dialogs, the browser opener, the keyboard handover
and the single-instance lock.
`main.go` is the composition root.

The page in `frontend/src` is a client of the facade through `api.ts` and
nothing else.

### The setup program: `installer/`

A second Wails application in the same module, carrying the built application
as an embedded zip, so one file is the whole distribution. `installer/main.go`
is its composition root and `installer/app.go` a facade over
`internal/infrastructure/setup`; its page is hand-written and holds no product
name of its own, because a page has no build step to catch a stale one. Its
licence screen is filled the same way, sentence for sentence, from
`internal/licence`.

The facade owns no install policy. The ORDER of an install or a removal is
policy too. `setup.Machine` states every act either one performs on the computer it
runs on, `setup.Real` implements it with one call per method, while
`setup.Install` and `setup.Remove` state the sequences over it. That is what
makes it checkable that a removal takes the shortcuts before the registry
entry, that it refuses outright while the application is open and that the
record goes last and only when asked. `installer/app.go` holds a `Machine` and
a progress reporter rather than reaching for either, so its own decisions (the
screen setup opens on, the choices it hands over) are testable too; what is
left uncovered there is the Wails runtime itself.

That package is the one home for the licence: the published text plus a plain
reading of what it permits and requires (FR-074). The text sits there as well
as at the repository root because Go's embedding cannot reach above its own
package and the setup program is a separate main, so neither binary could
embed the root LICENSE. The second copy is held to the first byte for byte by
`tests/structural/licence_test.go`, which is what makes a second copy of a
legal text safe to keep.

The licence pane reads itself down at the house pace (FR-075). The cycle is a
copy of the application's own `frontend/src/autoScroll.ts` in plain JavaScript,
because the setup page has no build step and cannot reach a TypeScript module;
Fulcrum's installer carries its own copy for the same reason. What must never
differ is the pace, so the constants are stated in both and are the
application's. Nothing watches for the screen changing, since a screen stack
has no unmount: the one place that routes away is the one place that stops the
timer.

It follows the house setup model: work moves to a progress screen rather than
greying the options in place, the footer is rebuilt per screen, the progress
screen offers nothing, one reading of the machine decides the route; every path
ends in a verdict. Setup opens dark and carries the light or dark toggle in its
header, as the application carries one in its bar, remembering the choice in
its own storage.

## Decisions and why

The decisions as they bear on the code. The product-level record of what was
chosen, what was given up and what each costs is
[DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md).

| Decision | Why | What it costs |
|---|---|---|
| The receipt's lines are built in the domain, not the page. | The one thing SymDiary must never do is add words to a medical record. A test asserts every line is a title, a range, the last appointment the user marked, a heading, a count, a recorded field or one of the two framing lines. | The page cannot reflow a line; it styles by kind. |
| The last appointment lives in the window's storage, beside the theme. | It is a date boundary the user reaches for, not an observation, so the record and its export stay observations alone and neither the schema nor the export format moves. `frontend/src/storage.ts` is the one home for reading and writing that storage; a refusal reads as nothing held. | It does not travel in an export or to another machine; cleared storage means marking it again. |
| The page decides whether the sheet names the appointment; the domain refuses one outside the range. | The page owns the choice (the range covers it, the box is ticked) because it must also enable the box. The sheet is a public claim, so `Receipt.WithAppointment` does not trust the page with it: an appointment the range does not cover is refused rather than printed. | The in-range comparison is stated twice, once as the page's choice and once as the domain's guard; each is held by its own tests. |
| Once shown, the record follows its range. | Save PDF writes the dates in the fields, so the record on screen has to be the one those dates give. Each change asks again; an answer for a range since left is discarded, so a slow reply cannot overwrite the newer one. | A call per change of date while the record is showing. |
| The saved document says what made it and what it is not. | A sheet outlives the window it came from: the reader is a doctor who has never seen SymDiary and cannot be assumed to know that the notes are the patient's own. The framing is fixed text that names no event, so it says nothing about what was recorded. | Two more line kinds; the words become a promise once a sheet is in a filing cabinet. |
| The line kinds are compared against the page's union by a test. | The wire test sees that a line carries a kind; it cannot see which kinds exist, while nothing in either build compares the two lists. A kind added in Go alone renders with a class no stylesheet knows, which reads correctly on screen and prints wrong. | A second scan; a new kind is two edits rather than one. |
| The record is drawn as a PDF rather than printed by the browser. | A page belongs to whichever engine the desktop ships, so the same record came off the paper three different ways, with what the reader got depending also on a tickbox in their own print dialog. The record is the product. Drawing it ourselves also buys what no browser would keep: an event never split across two sheets, a page number on every page, a document the suite can measure rather than one only a printer can. | A PDF library, a typeface carried in the binary and a layout to maintain, in exchange for the print stylesheet and the page-margin tricks it replaces. |
| The store resolves a symptom by key, creating it when absent. | Recording an event and creating its symptom is one transaction, so a failure leaves neither. | The store holds a key column the domain computes. |
| An edit sends no time unless the user changed it. | An empty time means "keep what is held", so an edit cannot move an occurrence to the moment of the edit. The rule is structural rather than remembered. | The edit form compares before sending. |
| A record that will not open becomes `store.Unavailable`. | The window opens and says what is wrong, instead of a program that never appears. Every action answers with the same reason. | Eight one-line methods that refuse. |
| Times are stored as RFC 3339 with their offset. | The instant reads back as the same instant; an export carries the offset it was written in. | Ordering happens in Go rather than in SQL. |
| No encryption at rest. | A passphrase is a thing to lose; the user's own account already guards the file. Stated in the README rather than assumed. | Anyone who can sign in as the user can read the record. |
| SymDiary opens dark and carries its own switch, rather than following Windows. | A window that changes under the reader because the desktop reached dusk is a surprise; a button in the bar is one press away and what it chooses is remembered. The palette is held to AA in both modes by a test either way. | The page owns a preference, so the tokens hang off an attribute rather than a media query; the setup program carries the same button so the two cannot disagree. |
| The Guide and About read themselves down, gently, until the reader takes over. | Long help holds still on open, descends a pixel every second tick, holds at the tail and rewinds; any wheel, press, key or focus arrival suspends it for 2.5 seconds and it then resumes from wherever the reader left it. The pace belongs to the application rather than to either dialog. | A timer per open dialog, plus a pure state machine to keep the pacing testable without waiting. |
| The focus ring is answered by the page, not left to the browser. | The browser has an opinion about Tab and none about the arrows, so the house model (Tab and Right forward, Shift+Tab and Left back, wrapping at both ends) has to be stated. It is split in two: the rules are a pure module under test; one listener drives them against the page. | One key listener at the shell, plus a text field that has to be asked for its arrows back rather than assumed. |
| Three ring states and no more. | Nothing at rest, so the window is quiet until it is used; green while a control is hovered or focused, because both say "you can use this" and a reader should not have to learn two colours for one fact; permanently red while disabled, because the red IS the state and a ring that waited for the mouse would leave Save PDF looking like a button nobody had pressed yet. The accent is data meaning and never a ring. | A disabled control has to give up its fill as well; otherwise the ring it is meant to show disappears into it. |
| Only the run log and the keyboard handover know which platform they are on. | Everything else in the application is portable: the record's folder comes from `os.UserConfigDir`, the page is a page; SQLite is pure Go. The run log answers a Windows-only failure, a windowed run handed a standard error handle of 0; `windowfocus` answers a WebView2-only one and does nothing elsewhere. Each keeps its Windows half behind a build tag. The run log's folder rule is a pure function taking the platform as an argument, so all three answers are exercised wherever the suite runs. The `setup` package is Windows work by nature and serves only the Windows setup program. | Two small files per package instead of one, plus a rule stated rather than read from the machine it runs on. |
| The Flatpak is given no network permission. | SymDiary opens no connection; the sandbox is where that claim stops being a claim: an application that started talking to something would fail at run time rather than quietly working. The build gets the network, because it fetches Go modules and npm packages. | The manifest has two permission lists that must not be confused for each other. |
| The donate address lives in Go and the page never names one. | The page asks for the donation page; Go holds the only copy of the address and hands it to the desktop. Nothing arrives from the page, so there is no address to validate before opening; the no-network guarantee is untouched because SymDiary fetches nothing. | One more bound method, plus a seam over Wails' opener so no test opens a browser. |
| The Donate button takes a seat in the bar rather than a band of its own. | The window already has a tray of icon buttons and no footer, so a second strip carrying one control costs more than it buys. It is drawn at its neighbours' height: a member sized smaller than the row it sits in reads as a mistake. | The mark keeps its own width, so one rule sits beside the band's square icons. |
| The window hands the page the keyboard as it opens. | DOM focus and keyboard focus are two different things in a hosted webview. The page can hold the first while the webview holds none of the second; no key then reaches any listener: measured in the built window, where no Tab stepped the ring until the page had been clicked once. Showing the main window does not fix it either: WebView2 hosts the page in a child window of its own and the keys follow the child. The page cannot fix this from its own side, so the facade asks the window on DOM ready. | A `windowFocuser` seam wired at the composition root, over `internal/infrastructure/windowfocus`, which focuses the WebView2 child window through Win32 and does nothing off Windows. It runs on a goroutine of its own, with a recover behind it, so neither the wait nor a panic there can touch the opening window. |
| The licence is explained before it is shown. | Naming a licence explains nothing to the person installing the program; a setup screen that says "GNU General Public Licence, version 3" and stops has told them only that there is one. So the screen says what they may do and what they must do, in ordinary words, then shows the text in full for anyone who wants it. | A plain reading and the published text, both from `internal/licence`, in a pane that reads itself down and steps aside when touched. The pane is a text view, so it draws no ring in any state. |
| A scrolling dialog body stays a keyboard stop and paints nothing. | It carries no controls of its own, so a reader who never touches the mouse must be able to reach it and scroll it; a ring round a whole page of words marks nothing to act on. Measured in Chromium: an overflowing container is focusable with no tabindex and drew the engine's own ring, so the ring is turned off explicitly. | One suppression rule, held by a test, rather than the absence of a rule. |

## The export format

The file a user exports is their own copy of their record, so SymDiary has to go
on reading it after the format has moved on. That is a promise about every
version ever written, not only the current one.

Every file carries an envelope that never changes shape: `format`, always
`symdiary-record`, then `version`, an integer. The envelope is read on its own
first; the version chooses which reader reads the rest. Reading the whole
file into one struct and hoping it fits is the thing a versioned format exists
to avoid.

| The file says | What happens |
|---|---|
| A different `format` (or nothing that parses as JSON) | Refused: not a SymDiary export |
| No `version` (or one below 1) | Refused. An unversioned file is not one SymDiary wrote; a record is not the thing to be generous about |
| A version SymDiary knows | Read by that version's own reader |
| A version above the current one | Refused, naming the version, so the user knows a newer SymDiary wrote it |

Adding a version is three things, all of which the suite refuses the change
until you have done: raise `formatVersion`, add a reader to the table in
`versions.go`, then commit a sample of the **old** version under `testdata`. The
sample is the bytes that version really wrote, frozen and never edited again: a
reader tested against the current struct is only tested against itself.

Both halves of that guard were proved by planting. Raising the version alone
fails with "format 2 can be written but not read"; adding the reader without the
sample fails naming the file to commit.

## Execution

1. `main` points the error output at the log, so a crash leaves a record.
2. It opens the record, else carries the failure as a problem the window shows.
3. It builds the four services over the store, the export format and the clock.
4. It hands the facade to Wails, with the single-instance lock.
5. Once the page exists, the facade asks the window for the keyboard, so the
   first key pressed reaches the ring rather than nothing.
6. Every page action is one bound method: convert, call one service, convert
   back. A panic inside one becomes an error the page shows, with the stack in
   the log.
7. On shutdown the record is closed.

## Further reading

- [REQUIREMENTS.md](REQUIREMENTS.md): the requirements these invariants serve.
- [TESTING.md](TESTING.md): how it is verified.
- [DEVELOPMENT.md](DEVELOPMENT.md): how it is built.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions the product rests on and what each costs.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks like debt.
