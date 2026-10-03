# Decisions and trade-offs

The deliberate choices SymDiary rests on: what was chosen, what was given up
for it and why. Each entry is the decision as the product makes it today.
How the code carries each one is in [ARCHITECTURE.md](ARCHITECTURE.md) and the
specification ([REQUIREMENTS.md](REQUIREMENTS.md)); [TECH_DEBT.md](TECH_DEBT.md)
records what was weighed and deliberately left alone.

## The product as a whole

### Local first, one person, one machine

Everything SymDiary keeps is one file in the place the platform gives a
program for its own settings, for the one person whose record it is.

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
change. Where SymDiary is filed under a category at all, it is a utility, never
health.

- **Rather than:** a disclaimer doing the work, which the same guidance says
  carries no weight where claims are made elsewhere.
- **Gains:** the product stays outside medical device regulation for as long as
  it interprets nothing.
- **Costs:** every new feature, word on the site or line on the sheet has to be
  read against that reading first.

### Go and Wails, with the record and the document in Go alone

The application is Go behind a Wails window over a React page. The record and
the document both go through libraries written in Go alone; on Windows nothing
in the build needs a C compiler.

- **Rather than:** a C database driver; a toolkit drawn natively on each
  desktop.
- **Gains:** one toolchain builds the Windows release; the delivery recipe, the
  gate and the setup program were ported from earlier projects on the same
  stack.
- **Costs:** the page is rendered by a different browser engine on each
  desktop, which is what later moved the document out of it. On Linux and macOS
  Wails' own window code needs a C compiler, so those packages are built on
  their own machines.

### Specification before code

Every behaviour has a numbered requirement with acceptance criteria and the
test that verifies it. A change to behaviour amends the specification first,
with a numbered amendment giving the reason.

- **Rather than:** building first and describing afterwards.
- **Gains:** a reversal is recorded with its reason; a requirement that names a
  test nobody wrote is caught by reading the tree.
- **Costs:** keeping the specification true is work of its own.

## Privacy and the network

### No network connection of its own

SymDiary's own code opens no connection, for any purpose. The rule is enforced
by tests over the code and by the page's own security policy, not left as a
promise. That includes the one feature most desktop applications reach for:
nothing asks whether a newer release exists. The single fetch that can happen
belongs to the window toolkit, on a Windows machine missing the component the
window is drawn with (see the Windows 10 entry below).

- **Rather than:** a promise that the network is used sparingly; a daily update
  check.
- **Gains:** no account, telemetry, advertising or analytics is possible
  without a test failing first; nothing from the record ever leaves the
  machine.
- **Costs:** every feature that would need the network is ruled out; a person
  finds a new release by visiting the site.

### The Linux sandbox is given no network either

The finished Flatpak runs without network access. Only the build is given the
network, to fetch what it compiles from.

- **Rather than:** the network permission most desktop applications are given.
- **Gains:** on Linux the claim is enforced by the system as well as by the
  tests; an application that started talking to something would fail rather
  than quietly work.
- **Costs:** two permission lists, for the build and for the application, that
  must not be confused for each other.

### Donations go through the browser

The Donate button asks the program for the donation page; the program holds
the only copy of the address and hands it to the desktop, whose browser does
the asking.

- **Rather than:** the page naming an address; SymDiary making the request
  itself.
- **Gains:** nothing arrives from the page that has to be checked before it is
  opened; SymDiary itself still opens no connection.
- **Costs:** the address changes only with a new release.

### No encryption at rest

The record is an ordinary file protected by the user's own account. The README
says so in plain words.

- **Rather than:** a passphrase over the record.
- **Gains:** no passphrase to lose, which would lose the record with it; a file
  the person can copy and inspect.
- **Costs:** anyone who can sign in as the user can read the record.

### The log holds no part of the record

The run log takes the program's errors and any crash, never a symptom or a
note. Every place that writes to it is declared and checked, so a new one has to
be justified before it can exist.

- **Rather than:** trusting that nobody logs a symptom.
- **Gains:** the promise stays checkable; adding a writer is the moment
  somebody has to ask what it puts in the file.
- **Costs:** a log that says little about what the person was doing when
  something failed.

## The record

### Durable writes, all or nothing

The record is written for durability rather than speed, one complete change
at a time; recording an event and creating its symptom happen together or not
at all.

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
- **Costs:** a stand-in record whose every action refuses, which looks like
  repetition and is not.

### The person's words are kept exactly

A symptom label and a note are stored exactly as typed: no trimming, no
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

An edit carries a time only when the person changed it; otherwise the time held
stays.

- **Rather than:** sending the form's whole state on every save.
- **Gains:** correcting a note cannot move an occurrence to the moment of the
  edit. The rule is structural rather than remembered.
- **Costs:** the edit form compares before it sends.

### Times keep their offset

Times are kept with the offset they were written in.

- **Rather than:** converting everything to one zone.
- **Gains:** an instant reads back as the same instant; an export carries the
  offset it was written in.
- **Costs:** the program orders events itself rather than leaving it to the
  database.

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

### The record carries its own version

The record says which shape it is in. An older shape is upgraded in one step;
a record written by a newer SymDiary is refused rather than guessed at.

- **Rather than:** an unversioned file.
- **Gains:** a record written by a later version is never silently misread.
- **Costs:** none recorded; only one shape exists so far.

### Room for other kinds of event

Every event carries a kind, which today is always a symptom.

- **Rather than:** a record that could only ever hold symptoms.
- **Gains:** medication, meals or sleep could follow without rewriting the
  events already held.
- **Costs:** a field that holds one value.

## The symptom record

### The words on the sheet are written in one place

Every line of the record is composed by the product's own rules, each with a
kind attached. The window and the document only style by kind; neither
composes a sentence of its own.

- **Rather than:** the page building its own text from the events.
- **Gains:** the one thing SymDiary must never do, add words to a medical
  record, is settled in one place a test can read.
- **Costs:** the page cannot reflow a line; a new kind of line touches every
  place that draws one.

### Counts and nothing else

The record holds a title, the date range, each symptom with its number of
events and each event's time, severity and note, plus two fixed framing lines.
The one other line is the date of the last appointment the person marked,
where the range covers it. Nothing else may appear on it.

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

### The last appointment is a date, not an event

The person can mark one date, today or earlier, as their last appointment. The
record then opens on the range from that day to today, with custom dates beside
it. The date is kept by the window, as the theme is, not in the record.

- **Rather than:** an appointment event in the record; a calendar of past and
  future appointments.
- **Gains:** the boundary the person reaches for before every appointment is
  one click; the record and its export stay the person's observations alone,
  with no change to either format.
- **Costs:** the date does not travel in an export or to another machine; only
  the most recent appointment is known.

### The sheet names the last appointment where the range covers it

When the range covers the marked appointment, its date prints under the date
range, whichever way the range was chosen. A range that does not cover it names
nothing. The person can untick a box to leave it off. The date sits beside the
range and never among the events.

- **Rather than:** a sheet that never mentions it; a line on every sheet;
  dividing the events into before and after.
- **Gains:** the doctor sees where the last consultation falls and can tell
  what is new; the comparison stays the doctor's, so the sheet still
  interprets nothing.
- **Costs:** one more line a test holds as public wording; a range spanning two
  appointments names only the later one.

### An empty range gives no record

A range holding no events is reported as empty; no document is written.

- **Rather than:** a blank sheet.
- **Gains:** nobody hands a doctor a page that says nothing.
- **Costs:** none recorded.

### Drawn as a PDF, not printed by the browser

SymDiary draws the record itself as an A4 PDF in a typeface it carries, then
saves it where the person chooses. It was first printed through the
window's browser engine; that engine is a different one on each desktop, so
the same record came off the paper three different ways.

- **Rather than:** printing the page, along with the print stylesheet and the
  margin tricks it needed; the PDF's built-in fonts, which cannot write an
  accented name or a curly quote.
- **Gains:** the same document on every desktop; the person's words reach the
  doctor as they were typed; a document the suite can measure rather than one
  only a printer can.
- **Costs:** a PDF library, a typeface and a layout of SymDiary's own to
  maintain; a larger program; on other paper the printer scales the sheet.

### An event is never split; every page is numbered

The layout keeps each event whole on one sheet and every page says which page
it is of how many. The one exception is an event too tall for any page: its note
continues onto the next page, which opens with the event's time again.

- **Rather than:** filling each page to the foot; for a note longer than a page,
  moving it to a fresh page and letting it run off the foot, which is what the
  layout once did, so the end of the note never reached the paper.
- **Gains:** a dropped sheaf can be put back in order; no observation that fits
  on a page is read half on one page and half on the next; nothing the person
  wrote is lost to make the layout tidy. The repeated time is a recorded field,
  so the sheet still carries no words of SymDiary's own.
- **Costs:** a page may end with space left over; a very long note does span
  pages.

### A word wider than the line is broken at the line

A run of characters with no space in it (a long address, say) is cut where the
line ends and carried on beneath.

- **Rather than:** leaving it whole, which was the first choice, on the reasoning
  that running into the margin keeps the record intact. Measured, a long address
  did not run into the margin: it ran off the paper, so its end was not on the
  sheet.
- **Gains:** every character typed reaches the paper.
- **Costs:** a long word reads broken across two lines.

### Characters the typeface cannot print are refused

The typeface carried in the program covers Latin, Greek and Cyrillic. A record
holding a character outside it is refused when Save PDF is pressed (before the
person is asked where to save); the refusal names the characters.

- **Rather than:** printing the font's empty box in place of each, which is what
  happened before: the window showed the note whole while the sheet showed boxes;
  carrying a typeface for every script, which costs tens of megabytes in every
  download; borrowing a font from the operating system, a failure path that
  differs on every desktop.
- **Gains:** the sheet never silently loses the person's words.
- **Costs:** a person writing in a script outside the typeface cannot save a PDF
  of those entries. Should anyone need it, a fallback typeface is the next step.

## Export and import

### A JSON file the person owns

The whole record exports to a JSON file at a place the person chooses and can
be read back.

- **Rather than:** synchronisation; a format only SymDiary can read.
- **Gains:** a copy anything can read; the way to move a record to another
  machine.
- **Costs:** a copy taken today does not follow later changes.

### The export format is a contract

Every export names its format and its version. Every version ever written keeps
its own reader, checked against a real file of that version; a file naming no
version is refused, as is one written by a newer SymDiary.

- **Rather than:** one structure for every version; reading an unversioned
  file as though it were the current one, which the first format did.
- **Gains:** every export SymDiary ever wrote stays readable by every later
  one.
- **Costs:** each new version keeps its old readers forever.

### Written whole or not at all; read with suspicion

An export appears complete or not at all. A read takes no more than a capped
size and refuses anything that is not a SymDiary export.

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

### Import keeps the times a file states, future ones included

The recording form refuses a time in the future (FR-006). An import does not:
an event whose time lies ahead of this machine's clock is read as the file
states it. A blank symptom, on an event or on a symptom kept with none, is still
refused.

- **Rather than:** refusing a future time on import too.
- **Gains:** an export taken on a machine whose clock runs a few minutes fast
  reads back whole rather than refused.
- **Costs:** a hand-edited file can place an event in the future. Raised by the
  2026-10-03 audit and kept on the owner's ruling, so it is not reopened.

## The interface

### Opens dark, with its own switch

SymDiary opens dark. A button in the bar moves it to light and back and the
choice is remembered; a window that cannot remember simply opens dark again.
The button shows the mode a press moves to, in its picture and its words. The
setup program carries the same button. Following the desktop's setting was
tried first and dropped.

- **Rather than:** following the desktop's light or dark setting.
- **Gains:** the window never changes under the reader because the desktop
  reached dusk; the choice is one press away.
- **Costs:** the program owns a preference the desktop already has.

### One home for every colour

Every colour is defined once, in light and dark sets. Every text pairing is
held to WCAG AA in both modes and every focus ring to the contrast a non-text
indicator needs.

- **Rather than:** colours written where they are used.
- **Gains:** the two modes stay consistent; a colour that cannot be read fails
  the suite rather than shipping.
- **Costs:** a new colour has to earn its place in the palette.

### The keyboard reaches everything

Tab and Right step forward; Shift+Tab and Left step back; both wrap. Enter and
Space act; Escape closes a dialog. A text field keeps its arrows for its caret.

- **Rather than:** the browser's default, which has an opinion about Tab and
  none about the arrows.
- **Gains:** the whole window works without a mouse.
- **Costs:** the page answers every key itself, so a text field has to be
  asked for its arrows back.

### The page is given the keyboard as the window opens

On opening, the window hands the keyboard to the part of itself that draws the
page, so the first key pressed counts.

- **Rather than:** relying on the window being shown, which was tried first
  and left the page deaf until it was clicked.
- **Gains:** the first Tab reaches the ring without a click on the page first.
- **Costs:** platform code of its own on Windows, where the problem lives.

### Three ring states and no more

A control shows nothing at rest, a green ring while it is hovered or focused
and a permanent red ring while it is disabled. A disabled control gives up its
fill so the red can be seen. The accent colour is never a ring.

- **Rather than:** a single outline on keyboard focus.
- **Gains:** a reader can tell at a glance what can be used and what cannot.
- **Costs:** every disabled control loses its fill.

### Containers never take focus

A ring belongs to a control. No pane, list or dialog body carries a ring or a
stop of its own, except that a dialog body that scrolls stays reachable from
the keyboard while drawing nothing.

- **Rather than:** a ring round a whole page of words, which marks nothing to
  act on.
- **Gains:** the ring only ever lands on something that does something.
- **Costs:** the engine's own ring has to be turned off on purpose.

### Long dialogs read themselves

The Guide, About and the setup program's licence hold still on opening, then
read themselves down slowly, hold at the foot, rewind and repeat. Any touch by
the reader pauses the cycle, which then carries on from where they left it.
The buttons stay pinned beneath the moving text.

- **Rather than:** static pages.
- **Gains:** long help can be read hands free; the reader is never fought for
  the scrollbar.
- **Costs:** a timer per open dialog, with its pacing kept testable apart from
  the page.

### The Guide names every control with its own picture

The Guide's words are one document; the dialog only draws them. Each control is
named with the picture it draws in the bar.

- **Rather than:** screenshots.
- **Gains:** the Guide shows exactly what the reader sees.
- **Costs:** the Guide has to be updated with the bar.

### The bar is grouped by purpose

The screens come first, then Import before Export, the order the two are
reached in. Rules separate the file actions, the theme button, the help and
Donate. The order is fixed by a test.

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

A save dialog is told what kind of file it is for, so its title, its filter and
the name it suggests agree. Both file dialogs open in Downloads where one
exists.

- **Rather than:** one dialog shared by the export and the document, which on
  macOS saved a record as a PDF with a JSON extension.
- **Gains:** a saved file is named for what it is on every desktop.
- **Costs:** none recorded.

### One copy runs

Starting SymDiary while it is already running brings the open window forward.

- **Rather than:** two windows writing to one record.
- **Gains:** one writer for the record.
- **Costs:** a second window cannot be opened even on purpose.

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

### Windows 10 as well as Windows 11

SymDiary supports 64-bit Windows 10 and 11. Windows 11 always carries the
WebView2 component the window is drawn with; Windows 10 may not. Where it is
missing or too old, the application and the setup program ask before opening;
if the person agrees, they fetch and run Microsoft's installer for it. Declining
leaves them unopened.

- **Rather than:** Windows 11 alone.
- **Gains:** machines still on Windows 10 can run SymDiary.
- **Costs:** the one network fetch anything makes on SymDiary's behalf, on
  those machines alone and only after asking; the route has not yet been run
  on a Windows 10 machine.

### Installed for one user, without administrator rights

The setup program installs into the user's own programs folder and registers
with the Apps list for that user alone.

- **Rather than:** a machine-wide install.
- **Gains:** no administrator prompt at any point.
- **Costs:** each account on a machine installs separately.

### One file is the whole distribution

The setup program carries the built application inside itself and unpacks it
only into the install folder, refusing anything that would land elsewhere.

- **Rather than:** a setup program that downloads or sits beside its payload;
  trusting the archive.
- **Gains:** one file to hand over; a malformed payload cannot write elsewhere
  on the machine.
- **Costs:** the setup program is rebuilt for every release of the
  application.

### Uninstalling keeps the record

An uninstall removes the program, its shortcuts, its registration and its
log. It removes the record only where the person ticks a box naming the file.

- **Rather than:** removing everything SymDiary ever wrote.
- **Gains:** removing a program is never the same act as throwing away years
  of observations.
- **Costs:** a record left behind until the person deletes it.

### The order of a removal is policy

A removal refuses while SymDiary is open, takes the shortcuts before the
registration, takes the record only when asked and leaves the install folder
to be deleted once setup has exited. Every act on the machine sits behind one
interface, so the sequence is tested.

- **Rather than:** a setup window that reaches for the registry and the disk
  directly.
- **Gains:** the order, which is where the decisions are, can be checked; a
  setup program running from inside the folder still leaves nothing behind.
- **Costs:** a thin wrapper with one call per act, which looks like a layer
  doing nothing.

### The licence is explained before it is shown

The setup program's licence screen says in ordinary words what the person may
do and what they must do, then shows the published text in full. Every word
comes from the program rather than the page; the copy kept beside the code is
held to the published file exactly.

- **Rather than:** naming the licence and stopping there.
- **Gains:** somebody installing the program knows what they are being given.
- **Costs:** a second copy of a legal text, kept safe only by a test.

### The licence travels with the program

A copy of the licence is delivered on every platform, alongside the program
itself.

- **Rather than:** a licence shown on a setup screen or linked to.
- **Gains:** every recipient keeps the copy the licence asks for.
- **Costs:** none recorded.

### The setup page has no build step

The setup program's page is hand-written plain JavaScript. The gate parses it
and checks it against what the program sends it.

- **Rather than:** a bundled, type-checked page.
- **Gains:** the setup program stays one application with nothing more to
  build.
- **Costs:** no type checking; a little of it was confirmed by hand.

### Each platform builds on itself; the gate builds for all three

Windows, Linux and macOS packages are each built on their own platform. Every
test run on Windows also compiles the program for Linux and macOS. Line
endings are fixed per file type, so a checkout behaves the same on each.

- **Rather than:** discovering a Windows-only dependency the day someone tries
  the Flatpak; leaving line endings to each machine's setting.
- **Gains:** a Windows-only dependency fails by name on the next run, with no
  Linux machine or Mac needed for that check; a fresh checkout builds
  everywhere.
- **Costs:** a machine of each kind to build the packages; an Apple developer
  account for the signing.

### Notarised and stapled, both bundle and image

The macOS build signs and notarises the application, staples the ticket to
both the application and the disk image, then replays Gatekeeper's own check.
It fails if any step does.

- **Rather than:** stapling the image alone.
- **Gains:** the copy a person actually runs is the one proved to carry the
  ticket.
- **Costs:** a slower build that needs the network for notarisation.

### One home for the version

The version is written in one file; every build reads it from there. The site
names no version of its own and asks for the newest release's instead.

- **Rather than:** a version written in several places.
- **Gains:** a release is one edit; the site never claims a version that cannot
  be downloaded.
- **Costs:** the site shows no version where it cannot reach the release list.

### A website written for the person, not the programmer

The site explains what SymDiary does for somebody who wants to remember what
their body has been doing. It carries no dates and makes no claim of
monitoring or insight, since promotional material is what fixes an intended
purpose. Its stylesheet and script are linked by their content, so a browser
never pairs a new page with an old stylesheet.

- **Rather than:** a developer's project page.
- **Gains:** the people deciding whether to install it find what they need.
- **Costs:** developers go to the repository instead; the links have to be
  restamped after each edit to the stylesheet or script.

### GPL-3.0, plus a commercial licence

The source is GPL-3.0. A commercial licence for SymDiary's own code is offered
separately.

- **Rather than:** one licence only.
- **Gains:** free and open for everybody; a route for a closed-source product.
- **Costs:** none recorded.

## Engineering

### Layers with one place where they meet

The code is split into the rules, the use cases, the parts that touch the
outside world and the interface, each depending only inward, with one place
wiring them together. The rules perform no input or output and never read the
clock. Tests hold the boundaries.

- **Rather than:** convention alone.
- **Gains:** the rules about the record can be tested with no disk, clock or
  window.
- **Costs:** more packages and more explicit wiring.

### Complete coverage where it means something

The rules and the use cases must be fully covered. Every other part is held at
its measured figure; a floor moves only when it is measured again.

- **Rather than:** one figure over everything; floors set as targets.
- **Gains:** anything short in the pure layers is a decision nobody made; a
  floor at the measured number fails the moment cover is lost.
- **Costs:** window, registry and process code relies on targeted tests and
  hand checks.

### Small files

No source file may grow past a fixed limit or linger just below it. A file
that comes close is split at a seam.

- **Rather than:** letting files grow.
- **Gains:** files split where a concern leaves rather than where a line count
  says.
- **Costs:** many small files.

### The wire is stated twice and compared

Every shape crossing between Go and the page is written in both languages and
the two are compared field for field.

- **Rather than:** relying on the generated bindings.
- **Gains:** a field added on one side alone fails the suite.
- **Costs:** every shape is two edits.

### A refusal cannot go unseen

Every call from the page must say what happens when it is refused; a call that
does not will not compile. A crash inside one action becomes a message on screen, with the
detail in the log.

- **Rather than:** failures that may go unhandled.
- **Gains:** every failure reaches the person in words; a fault in one action
  never kills the window.
- **Costs:** a handler on every call.

### Platform code kept to two small corners

Everything else in the application is portable. Only the run log and the
handing of the keyboard to the page carry Windows-only work; the log's rule for
where it lives is written for all three platforms at once.

- **Rather than:** platform checks spread through the code.
- **Gains:** all three platforms' answers for the log are exercised wherever
  the suite runs; elsewhere the keyboard handover simply does nothing.
- **Costs:** each of the two is split in two.

### Tests with real parts; guards proved to bite

The record is tested against the real database in a temporary folder; the
suite never reaches the network and never touches the real record. Every
structural guard was proved by planting a violation and watching it fail.

- **Rather than:** mocks and assumed guards.
- **Gains:** a passing test means the real thing works; a guard is known to
  bite.
- **Costs:** fakes are written by hand.
