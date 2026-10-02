# Decisions and trade-offs

The deliberate choices SymDiary rests on: what was chosen, what was given up
for it and why. Each entry is the decision as the product makes it today.
The detail behind each one, with the tests that hold it, lives in
[ARCHITECTURE.md](ARCHITECTURE.md) and the specification
([REQUIREMENTS.md](REQUIREMENTS.md)); [TECH_DEBT.md](TECH_DEBT.md) records what
was weighed and deliberately left alone.

## The product as a whole

### Local first, one person, one machine

Everything SymDiary keeps is one SQLite file in the folder the platform gives a
program for its own configuration, for the one person whose record it is.

- **Rather than:** an account, a server or anything synchronised between
  machines.
- **Gains:** nothing to sign in to; nothing to run; it works with the network
  switched off.
- **Costs:** the record belongs to that one machine. Moving it means exporting
  it and importing it on the other.

### A recorder, not a diagnostician

SymDiary keeps what the person observed and when, then hands it back. It does
not diagnose, interpret, score, chart, remind or rewrite anything into medical
terms. The only arithmetic anywhere is a count of events.

- **Rather than:** the trends, scores and nudges a health application is
  usually assumed to carry.
- **Gains:** a small surface that can be held to a high bar; a record a doctor
  can read as the person's own words.
- **Costs:** anyone wanting insight from their symptoms has to look elsewhere.

### Built to stay a written symptom diary in law

The intended purpose is held to the published MHRA guidance's own example of
software unlikely to be a medical device: a replacement for a written diary of
symptoms used when seeing a doctor. The guidance's caveat, that features which
enhance the data presented may change that, is a standing constraint on every
change. The Linux desktop entry lists it under utilities and office, never
health.

- **Rather than:** a disclaimer doing the work, which the same guidance says
  carries no weight where claims are made elsewhere.
- **Gains:** the product stays outside medical device regulation for as long as
  it interprets nothing.
- **Costs:** every new feature, word on the site or line on the sheet has to be
  read against that reading first.

### Go and Wails, with pure Go libraries

The application is Go behind a Wails window over a React page. The record goes
through a pure Go SQLite driver and the document through a pure Go PDF library;
nothing in the build graph uses cgo.

- **Rather than:** a C SQLite driver; a toolkit drawn natively on each desktop.
- **Gains:** neither the record nor the document needs a C toolchain; the
  delivery recipe, the gate and the setup program were ported from earlier
  projects on the same stack.
- **Costs:** the page is rendered by a different browser engine on each
  desktop, which is what later moved the document out of it.

### Specification before code

Every behaviour has a numbered requirement with acceptance criteria and the
test that verifies it. A change to behaviour amends the specification first,
with a numbered amendment giving the reason.

- **Rather than:** building first and describing afterwards.
- **Gains:** a reversal is recorded with its reason; a requirement that names a
  test nobody wrote is caught by reading the tree.
- **Costs:** keeping the specification true is work of its own.

## Privacy and the network

### No network connection at all

SymDiary opens no connection. A structural test forbids every networking
package in the repository's own Go code; the page carries a content security
policy that allows it no connection; a second test reads that policy.

- **Rather than:** a promise that the network is used sparingly.
- **Gains:** no account, telemetry, advertising or analytics is possible
  without a test failing first.
- **Costs:** every feature that would need the network is ruled out, update
  checks among them.

### No update check

Nothing asks whether a newer release exists.

- **Rather than:** a daily check against the releases page.
- **Gains:** the no-network rule has no exception.
- **Costs:** a person finds a new release by visiting the site.

### The Flatpak is given no network permission

The finished Linux application runs without network access in its sandbox.
Only the build gets the network, to fetch Go modules and npm packages.

- **Rather than:** the network permission most desktop applications are given.
- **Gains:** on Linux the claim is enforced by the system as well as by a test;
  an application that started talking to something would fail rather than
  quietly work.
- **Costs:** the manifest carries two permission lists that must not be
  confused for each other.

### Donations go through the browser; the address lives in Go

The Donate button asks Go for the donation page. Go holds the only copy of the
address and hands it to the desktop; the browser does the asking.

- **Rather than:** the page naming an address; SymDiary making the request
  itself.
- **Gains:** nothing arrives from the page that has to be checked before it is
  opened; no connection is opened by SymDiary.
- **Costs:** one more bound method, with a seam over the opener so no test
  opens a browser.

### No encryption at rest

The record is an ordinary file protected by the user's own account. The README
says so in plain words.

- **Rather than:** a passphrase over the record.
- **Gains:** no passphrase to lose, which would lose the record with it; a file
  the person can copy and inspect.
- **Costs:** anyone who can sign in as the user can read the record.

### The log holds no part of the record

The run log takes the process's error output and any crash. Every place that
writes to it is on a declared list, each a constant sentence plus an error or a
stack; a new writer fails a test until it is declared.

- **Rather than:** trusting that nobody logs a symptom.
- **Gains:** the promise stays checkable; adding a writer is the moment
  somebody has to ask what it puts in the file.
- **Costs:** a log that says little about what the person was doing when
  something failed.

## The record

### Durable writes, one transaction each

The store runs SQLite in write-ahead mode with full synchronisation. Every
write is one transaction; recording an event and creating its symptom happen
together or not at all.

- **Rather than:** the faster defaults.
- **Gains:** an event reported as recorded survives a power cut moments later;
  a failure leaves neither half behind.
- **Costs:** each write waits for the disk.

### A record that will not open still opens the window

If the file exists but cannot be read, SymDiary opens anyway, says what is
wrong and names the file. The file is never overwritten; every action answers
with the same reason.

- **Rather than:** a program that never appears; replacing a damaged file with
  an empty one.
- **Gains:** the person sees the problem; the damaged file is still there to
  recover.
- **Costs:** a stand-in store whose every method refuses, which looks like
  repetition and is not.

### The person's words are kept exactly

A symptom label and a note are stored byte for byte as typed: no trimming, no
spelling correction, no change of case.

- **Rather than:** tidying the text on the way in.
- **Gains:** what reaches a doctor is what the person wrote.
- **Costs:** a typing slip stays until the person corrects it.

### One symptom however it is typed

A typed symptom that differs from one already held only in letter case or
surrounding spaces is that symptom; the label first typed is kept. Suggestions
match anywhere in the label regardless of case, most recently used first.

- **Rather than:** a new symptom for every spelling; a fixed list of symptoms.
- **Gains:** a record of tiredness is one group however it was typed that
  morning; the person names symptoms in their own words.
- **Costs:** two symptoms the person means to keep apart cannot differ only by
  case.

### Two times on every event

An event carries when it happened and when it was recorded. The first defaults
to the moment of recording and can be set earlier; the second is set once and
never changes. A time in the future is refused.

- **Rather than:** a single timestamp.
- **Gains:** a symptom remembered hours later is recorded at the right time
  without hiding when it was written down.
- **Costs:** two fields to explain.

### An edit never moves the time by accident

An edit sends no time unless the person changed it; an empty time means keep
the one held.

- **Rather than:** sending the form's whole state on every save.
- **Gains:** correcting a note cannot move an occurrence to the moment of the
  edit. The rule is structural rather than remembered.
- **Costs:** the edit form compares before it sends.

### Times keep their offset

Times are stored with the offset they were written in. Ordering happens in Go
rather than in the database.

- **Rather than:** converting everything to one zone.
- **Gains:** an instant reads back as the same instant; an export carries the
  offset it was written in.
- **Costs:** the database cannot sort by time on its own.

### Severity is optional and never computed

Severity is a choice of Mild, Moderate or Severe, with none selected by
default.

- **Rather than:** a score or a scale.
- **Gains:** the record says only what the person chose to say.
- **Costs:** three steps are coarse.

### Deletion asks, naming what will go

Deleting an event shows a confirmation naming its symptom and its time; a bulk
deletion says how many events will go.

- **Rather than:** deleting on one press with an undo.
- **Gains:** nothing leaves the record without the person seeing what it is.
- **Costs:** one more press per deletion.

### Renaming a symptom renames it everywhere

Events refer to a symptom rather than holding a copy of its name, so a rename
reaches every event at once.

- **Rather than:** a label copied onto each event.
- **Gains:** a better name for a symptom is one change.
- **Costs:** the old name is gone from every event, past ones included.

### No edit history

An edit replaces what was held; nothing keeps the earlier version.

- **Rather than:** an audit trail of every change.
- **Gains:** a simpler record holding only what the person currently says.
- **Costs:** an edited event cannot show what it said before.

### The schema carries its version

The store records which schema a file holds and upgrades an older one in a
single transaction. A file written by a newer SymDiary is refused rather than
guessed at.

- **Rather than:** an unversioned file.
- **Gains:** a record written by a later version is never silently misread.
- **Costs:** none recorded; only one schema version exists so far.

### Room for other kinds of event

Every event row carries a kind, which today is always a symptom.

- **Rather than:** a table that could only ever hold symptoms.
- **Gains:** medication, meals or sleep could follow without rewriting the
  rows already held.
- **Costs:** a column that holds one value.

## The symptom record

### The words on the sheet are written by the domain

Every line of the record is composed by the pure rules with a kind attached.
The window and the document style by kind; neither composes a sentence of its
own. Tests compare the kinds against the page's list and the document's styles.

- **Rather than:** the page building its own text from the events.
- **Gains:** the one thing SymDiary must never do, add words to a medical
  record, is settled in one place a test can read.
- **Costs:** the page cannot reflow a line; a new kind of line is several
  edits.

### Counts and nothing else

The record holds a title, the date range, each symptom with its number of
events and each event's time, severity and note. A test asserts every line is
one of those or one of the two fixed framing lines.

- **Rather than:** summaries, comparisons or trends.
- **Gains:** a doctor reads the person's observations and not the program's
  opinion of them.
- **Costs:** the doctor does any arithmetic beyond a count.

### Grouped by first occurrence, oldest first

Symptom groups run in the order each first appeared in the range; events within
a group run oldest to newest. Names play no part in the order.

- **Rather than:** alphabetical groups; the newest-first order of the history.
- **Gains:** the record reads in the order things happened.
- **Costs:** the history on screen and the record run in opposite directions.

### The year only where it is needed

Event times leave out the year while the range lies inside one year and name
it on every event when the range crosses one.

- **Rather than:** the year on every line; never the year.
- **Gains:** short lines in the common case; no line a doctor could read as the
  wrong year.
- **Costs:** none recorded.

### The sheet says what it is and what it is not

The first line names SymDiary and its address, once, beside the application's
mark. Under the date range the sheet says it is not a diagnosis but one
person's own notes, printed for a healthcare professional to read. It says
notes rather than guidance on purpose: guidance names a purpose SymDiary does
not have.

- **Rather than:** a bare list; a second copy of the address at the foot, which
  a real test print showed saying the same thing twice.
- **Gains:** a sheet that outlives the window still says where it came from and
  what it is.
- **Costs:** the wording is a public claim once a sheet is in a filing cabinet,
  so a test holds it whole.

### The sheet does not say whose it is

The record carries no name, date of birth or other identifying detail.

- **Rather than:** a header naming the person.
- **Gains:** SymDiary holds no identity to leak; a mislaid sheet names nobody.
- **Costs:** the person writes their name on it by hand if a clinic wants it.

### An empty range gives no record

A range holding no events is reported as empty; no document is written.

- **Rather than:** a blank sheet.
- **Gains:** nobody hands a doctor a page that says nothing.
- **Costs:** none recorded.

### Drawn as a PDF, not printed by the browser

The record is drawn in Go as a PDF and saved where the person chooses. It was
first printed through the window's browser engine; that engine is a different
one on each desktop, so the same record came off the paper three different
ways. On Windows it also carried the browser's own header and footer unless
the reader's print dialog said otherwise.

- **Rather than:** printing the page, along with the print stylesheet and the
  margin tricks it needed.
- **Gains:** the same bytes on every desktop; a document the suite can measure
  rather than one only a printer can.
- **Costs:** a PDF library, a typeface and a layout of SymDiary's own to
  maintain.

### The Go fonts, carried in the binary

The document is set in the Go fonts, embedded in the program.

- **Rather than:** a PDF's built-in fonts, which can say nothing outside
  Latin-1.
- **Gains:** a note holding a curly quote or an accented name reaches the
  doctor as it was written.
- **Costs:** the fonts add to the size of the program.

### An event is never split; every page is numbered

The layout keeps each event whole on one sheet and every page says Page N of
M. The layout is a pure function of the lines, so what lands on which page is
settled by tests that draw nothing.

- **Rather than:** filling each page to the foot.
- **Gains:** a dropped sheaf can be put back in order; no observation is read
  half on one page and half on the next.
- **Costs:** a page may end with space left over.

### A4 paper

The document is laid out on A4, the paper a record printed in the United
Kingdom is read on.

- **Rather than:** following the printer's own paper size.
- **Gains:** one layout to test; a PDF names its own size, so a printer on
  other paper scales it rather than cutting it off.
- **Costs:** on other paper the printer scales the sheet to fit.

## Export and import

### A JSON file the person owns

The whole record exports to a JSON file at a place the person chooses and can
be read back.

- **Rather than:** synchronisation; a format only SymDiary can read.
- **Gains:** a copy anything can read; the way to move a record to another
  machine.
- **Costs:** a copy taken today does not follow later changes.

### The export format is a contract

Every file carries an envelope that never changes shape: a format name and an
integer version. The version chooses its own reader from a table. Every version
ever written keeps a frozen sample in the tests; a file naming no version is
refused, as is one written by a newer SymDiary.

- **Rather than:** one structure for every version; reading an unversioned
  file as though it were the current one, which the first format did.
- **Gains:** every export SymDiary ever wrote stays readable by every later
  one; raising the version without a reader fails the suite.
- **Costs:** each new version is three changes: the constant, a reader and a
  committed sample.

### Written whole or not at all; read with suspicion

An export is written to a temporary file and moved into place in one step. A
read takes no more than a capped size and refuses anything that is not a
SymDiary export.

- **Rather than:** writing in place; reading whatever is offered.
- **Gains:** a failed export leaves no partial file; a stray file cannot decide
  what the record says.
- **Costs:** none recorded.

### Import skips what is already held

An event is the same observation when its symptom, its two times, its severity
and its note all agree. Import adds only what is new, repeats within the file
included.

- **Rather than:** adding everything; replacing the record.
- **Gains:** importing the same file twice changes nothing.
- **Costs:** an event edited on one machine arrives as a second event on the
  other.

## The interface

### Opens dark, with its own switch

SymDiary opens dark. A button in the bar moves it to light and back; the choice
is kept in the window's own storage. Storage that refuses leaves the default
showing. The setup program carries the same button. The Windows setting was
followed at first and dropped the same day.

- **Rather than:** following the desktop's light or dark setting.
- **Gains:** the window never changes under the reader because the desktop
  reached dusk; the choice is one press away.
- **Costs:** the page owns a preference, so the colours hang off an attribute
  rather than the system setting.

### A switch shows what a press will do

The theme button shows the mode it moves to, in its picture and its words: the
sun means light is a press away.

- **Rather than:** showing the current state.
- **Gains:** one convention, shared with the setup program.
- **Costs:** learned once.

### One home for every colour

Every colour is defined once, in light and dark sets. Tests hold every text
pairing to WCAG AA in both modes and every ring to the 3:1 a non-text indicator
needs.

- **Rather than:** colours written where they are used.
- **Gains:** the two modes stay consistent; a colour that cannot be read fails
  the suite rather than shipping.
- **Costs:** a new colour has to earn its place in the palette.

### The keyboard reaches everything

Tab and Right step forward; Shift+Tab and Left step back; both wrap. Enter and
Space act; Escape closes a dialog. A text field keeps its arrows for its caret.
The rules are a pure module under test with one listener driving them.

- **Rather than:** the browser's default, which has an opinion about Tab and
  none about the arrows.
- **Gains:** the whole window works without a mouse.
- **Costs:** one listener at the shell, plus a text field that has to be asked
  for its arrows back.

### The page is given the keyboard as the window opens

On opening, the window hands keyboard focus to the child window WebView2
draws the page in, through the Win32 calls that reach it. It runs on its own
goroutine behind a recover.

- **Rather than:** asking Wails to show the window, the first attempt, which
  focused the main window while the keys follow the child.
- **Gains:** the first Tab reaches the ring without a click on the page first.
- **Costs:** platform code of its own; nothing is done off Windows, where the
  page already has the keyboard.

### Three ring states and no more

A control shows nothing at rest, a green ring while it is hovered or focused
and a permanent red ring while it is disabled. A disabled control gives up its
fill so the red can be seen. The accent colour is never a ring.

- **Rather than:** a single outline on keyboard focus.
- **Gains:** a reader can tell at a glance what can be used and what cannot.
- **Costs:** every disabled control loses its fill.

### Containers never take focus

A ring belongs to a control. No pane, list or dialog body carries a ring rule
or a tab stop. A dialog body that scrolls stays reachable from the keyboard and
turns the engine's own ring off explicitly.

- **Rather than:** a ring round a whole page of words, which marks nothing to
  act on.
- **Gains:** the ring only ever lands on something that does something.
- **Costs:** one suppression rule held by a test rather than the absence of a
  rule.

### Long dialogs read themselves

The Guide, About and the setup program's licence pane hold still for five
seconds, then descend a pixel every second tick of forty milliseconds, hold at
the foot, rewind and repeat. Any wheel, press, key or focus arrival suspends
the cycle for two and a half seconds; it then resumes from where the reader
left it. The action row is pinned beneath the scrolling body.

- **Rather than:** static pages.
- **Gains:** long help can be read hands free; the reader is never fought for
  the scrollbar.
- **Costs:** a timer per open dialog, plus a state machine to keep the pace
  testable.

### The Guide names every control with its own picture

The Guide's words are one document; the dialog only draws them. Each control is
named with the picture it draws in the bar.

- **Rather than:** screenshots.
- **Gains:** the Guide shows exactly what the reader sees.
- **Costs:** the Guide has to be updated with the bar.

### The bar is grouped by purpose

Import comes before Export, the order the two are reached in. Rules separate
the file actions, the theme button, the help and Donate. A test pins the order
by label.

- **Rather than:** one undivided row of icons.
- **Gains:** related actions sit together; a later edit cannot quietly
  reshuffle them.
- **Costs:** none recorded.

### Donate takes a seat in the bar

The Donate button sits last in the bar, drawn at its neighbours' height.

- **Rather than:** a band of its own along the foot.
- **Gains:** no second strip of chrome for one control.
- **Costs:** its mark keeps its own width beside square icons.

### Each save dialog asks for the file it writes

A save dialog is told what kind of file it is for: its title, its filter and
the extension of the suggested name. Both file dialogs open in Downloads where
one exists.

- **Rather than:** one dialog shared by the export and the document, which on
  macOS saved a record as a PDF with a JSON extension.
- **Gains:** a saved file is named for what it is on every desktop.
- **Costs:** none recorded.

### One copy runs

Starting SymDiary while it is already running brings the open window forward.

- **Rather than:** two windows writing to one record.
- **Gains:** one writer for the record.
- **Costs:** no test covers it; it needs two real processes.

### The copyright year is the first release

About names the year of the first release, never the year the program runs
in.

- **Rather than:** a year that follows the clock.
- **Gains:** the notice claims no date nothing was published on.
- **Costs:** none recorded.

## Building and installing

### A setup program on Windows; native packages elsewhere

Windows gets a bespoke setup program. Linux gets a Flatpak and macOS a signed,
notarised disk image, each the way that platform installs things.

- **Rather than:** one bespoke installer on all three.
- **Gains:** each platform installs SymDiary the way its users expect.
- **Costs:** three delivery routes to maintain.

### Installed for one user, without administrator rights

The setup program installs into the user's own programs folder and registers
with the Apps list for that user alone.

- **Rather than:** a machine-wide install.
- **Gains:** no administrator prompt at any point.
- **Costs:** each account on a machine installs separately.

### One file is the whole distribution

The setup program carries the built application as an embedded archive. The
build puts an empty placeholder back afterwards.

- **Rather than:** a setup program that downloads or sits beside its payload.
- **Gains:** one file to hand over; a payload of megabytes never reaches a
  commit.
- **Costs:** the setup program is rebuilt for every release of the
  application.

### Uninstalling keeps the record

An uninstall removes the program, its shortcuts, its registry entry and its
log. It removes the record only where the person ticks a box naming the file.
A test fails if the record's folder is ever among the folders cleared without
asking.

- **Rather than:** removing everything SymDiary ever wrote.
- **Gains:** removing a program is never the same act as throwing away years
  of observations.
- **Costs:** a record left behind until the person deletes it.

### The order of a removal is policy

A removal refuses while SymDiary is open, takes the shortcuts before the
registry entry, takes the record only when asked and hands the install folder
to Windows to delete once setup has exited. Every act on the machine sits
behind one interface, so the sequence is tested.

- **Rather than:** a setup window that reaches for the registry and the disk
  directly.
- **Gains:** the order, which is where the decisions are, can be checked; a
  setup program running from inside the folder still leaves nothing behind.
- **Costs:** a thin wrapper with one call per act, which looks like a layer
  doing nothing.

### The payload cannot climb out

Every entry in the archive must land inside the install folder before it is
written.

- **Rather than:** trusting the archive.
- **Gains:** a malformed payload cannot write elsewhere on the machine.
- **Costs:** none recorded.

### The licence is explained before it is shown

The setup program's licence screen says in ordinary words what the person may
do and what they must do, then shows the published text in full. Every word
comes from Go. The text kept beside the code is held to the published file byte
for byte.

- **Rather than:** naming the licence and stopping there.
- **Gains:** somebody installing the program knows what they are being given.
- **Costs:** a second copy of a legal text, kept safe only by a test.

### The licence travels with the program

A copy of the licence is delivered on every platform: in the install folder on
Windows, inside the Flatpak and in the macOS bundle, copied there before
signing.

- **Rather than:** a licence shown on a setup screen or linked to.
- **Gains:** every recipient keeps the copy the licence asks for.
- **Costs:** none recorded.

### The setup page has no build step

The setup program's page is hand-written plain JavaScript. The gate parses it;
a structural test checks every element it looks up exists and every field it
reads is one Go sends.

- **Rather than:** a bundled, type-checked page.
- **Gains:** the setup program stays one Wails application with nothing more
  to build.
- **Costs:** no type checking; the hop from the record checkbox to the
  uninstall was confirmed by hand.

### Each platform builds on itself; the gate builds for all three

Windows, Linux and macOS packages are each built on their own platform. The
gate compiles and vets the module for Linux and macOS on every run from
Windows.

- **Rather than:** discovering a Windows-only import the day someone tries the
  Flatpak.
- **Gains:** a Windows-only import fails by name on the next run, with no Linux
  machine or Mac needed for that check.
- **Costs:** a machine of each kind to build the packages; an Apple developer
  account for the signing.

### Notarised and stapled, both bundle and image

The macOS build signs and notarises the application, staples the ticket to
both the bundle and the disk image, then replays Gatekeeper's own check. It
fails if any step does.

- **Rather than:** stapling the image alone.
- **Gains:** the copy a person actually runs is the one proved to carry the
  ticket.
- **Costs:** a slower build that needs the network for notarisation.

### One home for the version

The version is written in one file. The build reads it into the binary through
a linker flag, which is why the variable it sets cannot be a constant.

- **Rather than:** a version written in several places.
- **Gains:** a release is one edit.
- **Costs:** a variable that looks like it should be a constant.

### Line endings pinned

Go, shell, TypeScript, CSS, JSON, Markdown and YAML are stored and checked out
with LF; PowerShell with CRLF.

- **Rather than:** leaving it to each machine's setting, which broke the
  formatter on files nobody had edited and would have put a carriage return in
  every shell script's first line.
- **Gains:** a fresh checkout builds on every platform.
- **Costs:** none recorded.

### A website written for the person, not the programmer

The site explains what SymDiary does for somebody who wants to remember what
their body has been doing. It carries no dates and makes no claim of
monitoring or insight, since promotional material is what fixes an intended
purpose. Its stylesheet is linked by a hash of its content.

- **Rather than:** a developer's project page.
- **Gains:** the people deciding whether to install it find what they need; a
  changed stylesheet is never paired with a cached old one.
- **Costs:** developers go to the repository instead; the hashes have to be
  stamped after each edit.

### GPL-3.0, plus a commercial licence

The source is GPL-3.0. A commercial licence for SymDiary's own code is offered
separately.

- **Rather than:** one licence only.
- **Gains:** free and open for everybody; a route for a closed-source product.
- **Costs:** none recorded.

## Engineering

### Layers with one place where they meet

The code is split into domain, application, infrastructure and the interface,
each depending only inward, with one composition root wiring them together.
The domain performs no IO and never reads the clock. Structural tests hold the
boundaries.

- **Rather than:** convention alone.
- **Gains:** the rules about the record can be tested with no disk, clock or
  window.
- **Costs:** more packages and more explicit wiring.

### Complete coverage where it means something

The domain and application layers must be fully covered. Every other package
is held at its measured figure; a floor moves only when it is measured again.

- **Rather than:** one figure over everything; floors set as targets.
- **Gains:** anything short in the pure layers is a decision nobody made; a
  floor at the measured number fails the moment cover is lost.
- **Costs:** window, registry and process code relies on targeted tests and
  hand checks.

### Small files

No source file may exceed four hundred lines; none may sit just below it. A
file that comes close is split at a seam.

- **Rather than:** letting files grow.
- **Gains:** files split where a concern leaves rather than where a line count
  says.
- **Costs:** many small files.

### The wire is stated twice and compared

Every shape crossing between Go and the page is written in both languages. A
test compares them field for field.

- **Rather than:** relying on the generated bindings.
- **Gains:** a field added on one side alone fails the suite.
- **Costs:** every shape is two edits.

### A refusal cannot go unseen

Every call from the page takes a refusal handler as its last argument and
answers nothing when refused; a call without one does not compile. A panic in
a bound method becomes an error the page shows, with the stack in the log.

- **Rather than:** promises that reject and may go unhandled.
- **Gains:** every failure reaches the person in words; a fault in one action
  never kills the window.
- **Costs:** a handler on every call.

### Only the run log knows which platform it is on

Everything else is portable. The log's Windows-only handle work sits behind a
build tag; its rule for where the log lives takes the platform as an argument.

- **Rather than:** platform checks spread through the code.
- **Gains:** all three platforms' answers are exercised wherever the suite
  runs.
- **Costs:** two small files instead of one.

### Tests with real parts; guards proved to bite

The store is tested against real SQLite in a temporary folder; the suite never
reaches the network and never touches the real record. Every structural guard
was proved by planting a violation and watching it fail.

- **Rather than:** mocks and assumed guards.
- **Gains:** a passing test means the real thing works; a guard is known to
  bite.
- **Costs:** fakes are written by hand.
