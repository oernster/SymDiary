# SymDiary: testing

## The gate

```powershell
./test.ps1
```

It runs these in order, stopping at the first failure:

1. `gofmt -l` over the Go, ignoring the front end.
2. `go vet` over every package but `node_modules`.
3. A build and a vet for Linux and for macOS. SymDiary ships to all three, so a
   Windows-only import is a defect the day it is written rather than the day
   someone tries the Flatpak. This needs no Linux machine and no Mac; it fails
   naming the import and the platform: proved by planting
   `golang.org/x/sys/windows` in a file with no build tag.
4. `staticcheck`, fetched with `go run honnef.co/go/tools/cmd/staticcheck@latest`.
5. The whole Go suite.
6. Coverage of `internal/domain` and `internal/application`, which must be 100%.
7. Coverage of every other Go package against its measured floor.
8. The front end: `eslint`, `tsc --noEmit` and `vite build`.
9. The front-end suite: Vitest over jsdom.
10. `node --check` over every script in `installer/frontend/dist`. The setup
    page has no build step, so nothing else parses it and a typo there reaches a
    user as a window that draws no screen at all. It is neither a lint nor a
    type check: `tests/structural/setuppage_test.go` covers the rest, holding
    every element the scripts look up to the markup and every field they read
    to what the setup facade sends.

`./test.ps1 -SkipFrontend` runs the Go half alone while working on it.

`build.ps1` runs the whole gate first and stops on a failure. Its one escape is
`-Fast`, which skips the gate for a working loop and prints that it has done so;
a release is never cut with it.

## How to read a result

Read the exit code, never the last line of output. The coverage run prints a
table last, so a `tail` shows coverage rows rather than a verdict; a grep for `error` matches file names.

```powershell
./test.ps1; "EXIT=$LASTEXITCODE"
```

`0` means every check passed and every floor held. Anything else means read the
failure the script threw.

## What a first run needs from the machine

- Go 1.26 or later; Node 20 or later.
- `npm install` inside `frontend/` once. `test.ps1` does not install for you.
- Step 4 downloads staticcheck the first time; it needs the network once.
- On Windows, allow Go's scratch directory (`go env GOTMPDIR`) in your
  anti-virus. A quarantined test binary stops the suite running at all rather
  than failing a test.

## The floors

Each is the measured number at the time of writing, not a target. A floor at an
aspiration only teaches people to lower it; a floor at the measured number
fails the moment cover is lost, which is the only moment it is worth being
told.

| Package | Floor | Why not 100 |
|---|---|---|
| `internal/domain` | 100 | Nothing to excuse: pure rules. |
| `internal/application` | 100 | Every port is faked, so every path is reachable. |
| `internal/infrastructure/store` | 92 | The rest is SQLite write failures that cannot be forced without breaking the disk. |
| `internal/infrastructure/export` | 91 | The rest is operating-system write failures on a temporary file. |
| `internal/infrastructure/runlog` | 81 | The rest is Win32 standard-handle work, reachable only in a windowed process with no error output. It was 74 until the folder rule became a pure function taking the platform as an argument: all three platforms' answers are now exercised on whichever platform the suite runs on, rather than two of them waiting for a user to report the answer. |
| `internal/infrastructure/setup` | 64 | The portable half and the shortcut writing are tested, as is what an uninstall takes and the scheduled removal of the install folder. The registry writes and the process work change the machine, so a real install exercises them instead. It was 56 until the Windows theme read went: setup opens dark and carries its own toggle, so a registry lookup no test could exercise is gone rather than sitting there lowering the number. It reached 62 when the uninstall's choice of folders became a function that could be asserted on, then 64 when the install and removal sequences moved into the package behind the `Machine` seam and got tests of their own. |
| root package (the facade) | 77 | The facade itself is covered. `main`, the log handover, the file dialogs, the browser opener and the single-instance lock need a real window. It has been re-measured twice rather than estimated: it fell to 76 when the donate button added one more line of Wails runtime no test can reach, then returned to 77 when printing through the window was replaced by saving a PDF, because the bound method now decides something (the range, the name it suggests, the kind of file it asks for) where before it forwarded one call. |
| `./installer` | 69 | The setup program's facade. What is left is the Wails runtime beneath it: emitting a progress event to a window and quitting one. Everything the facade decides sits above that and is tested against `setup.Machine`. |
| `internal/infrastructure/pdf` | 99 | The document a reader takes away. It needs no window, no clock and no device, so it is gated at what it actually reaches; the one statement not covered is a font the library refuses to load, which cannot happen to a font compiled into the binary. |

Not gated at all: `internal/product` holds constants and no behaviour, though the
two that reach paper are asserted word for word from the facade's own tests;
`internal/licence` holds the embedded licence and its plain reading, compared
with the root LICENSE by `tests/structural`; `internal/infrastructure/windowfocus`
is Win32 work against a window the suite does not own, settled by pressing Tab
in the built window; `tests/structural` is itself the guard;
`internal/infrastructure/setup/setuptest` is the double the other two suites are
written against.

## What each suite proves

| Suite | What it settles |
|---|---|
| `internal/domain` | The rules: blank and future occurrences refused, case variants matched, filters combined, the receipt's grouping, ordering and exact lines, including a range that crosses a year. The framing too: it opens the record and appears exactly once, the statement prints under the range, then two different records are shown to carry framing identical word for word. The last appointment is named under the range exactly where the range covers it (ends included); anywhere else it is refused. |
| `internal/application` | The use cases over fakes: recording reads the clock once, an edit leaves recorded-at alone, a failed write loses nothing the user typed, an import skips what is held. |
| `internal/infrastructure/store` | Real SQLite in a temporary folder: notes and symptoms read back byte for byte, a deletion is all or nothing, a garbage file is never overwritten, a newer schema is refused, the durability pragmas are set, plus a planted trigger proving a failed write leaves nothing behind. |
| `internal/infrastructure/export` | The format against a committed sample, a refusal for anything that is not a SymDiary export, plus a size cap read before the file is. The version contract as well: every version ever written still reads its own frozen sample, no reader claims a version nothing wrote, a file naming no version is refused rather than read as the current one, plus a gap in the table refusing rather than guessing. |
| `internal/infrastructure/runlog` | Crash reporting, by starting this test binary again as a child and making it panic, plus each platform's rule for where the log lives, all three exercised wherever the suite runs. |
| root package | The facade end to end over a real record: every conversion, every refusal, plus a panic in a bound method becoming an error rather than a dead window. `window_test.go` holds the window's own seams: the Downloads folder both dialogs open in, plus the keyboard being asked for once on DOM ready; that asking returns at once because the Win32 work happens on a goroutine of its own. `receipt_test.go` holds the sheet on its own, with the framing's exact words written out a second time, so an edit to the line saying the record is not a diagnosis fails rather than ships; it also holds the appointment line the window is sent, word for word. |
| `internal/infrastructure/setup` | The install policy: the payload fence refusing an entry that climbs out of the install folder, the paths, the version comparison that picks the route, plus real shortcuts written into temporary folders with plain paths rather than doubled separators. What an uninstall takes is covered separately, because it is the one part that cannot be undone: the record's folder is never among the folders cleared without asking, nor inside one of them; a folder the machine cannot place is left out rather than turned into a relative path. The scheduled removal of the install folder is run for real on Windows and waited for. |
| `internal/infrastructure/pdf` | The document: that a record of nothing is refused rather than written blank, that a path that cannot be written is reported, that a long record runs to the pages the layout said it would and that the file begins `%PDF-`. The layout is settled separately and without drawing anything, against a measurer that gives every character the same width: a note wraps without losing a word, a word wider than the page is left whole, an event is never split across two sheets, every page opens at the same height and no row reaches into the foot margin the page number sits in. |
| `tests/structural` | The invariants in ARCHITECTURE.md, plus one that is not an invariant of the code at all: the licence embedded in `internal/licence` is compared with the repository's own LICENSE byte for byte, because a setup program stating terms nobody granted is worse than one that says nothing. |
| `frontend` | The page over a fake facade: the keyboard path to a recorded event, the form keeping everything when a save is refused, an edit sending no time unless it changed, the confirmation naming what will go, plus the receipt showing exactly the lines it was given and carrying exactly one mark, in its first line. The receipt pane is held to its range: once shown, the record follows every change of dates and a slow answer for an earlier range is thrown away; a marked appointment opens the range on its own day, is named on the sheet only while the range covers it and the box is ticked. Window storage that refuses to be read or written leaves the pane working. The self-reading cycle is covered twice: the pure machine tick by tick, then the hook under jsdom for what suspends it and what freezes it. The focus ring is covered the same way: the rules alone, then the ring against a real page, where Tab and Right agree, Shift+Tab and Left agree, both wrap, a disabled control is passed over and a text field keeps its arrows. |

## What the tests never do

- They never reach the network. Nothing in the suite may.
- They never touch your real record: every store test opens a file under `t.TempDir()`; the facade tests do the same.
- They never mock SQLite. The store is tested against the real engine.
- They never assert against Wails' generated bindings. The wire contract is
  `frontend/src/api.ts`, compared with the Go DTOs by a structural test; its
  receipt-line kinds are compared with the domain's by a second one. The first
  sees that a line carries a kind; only the second sees which kinds exist. Both
  arms were proved by planting a violation: a kind dropped from the page's union
  and a kind the page was never told about each failed by name, with the tree
  put back afterwards.

## What only a person can settle

Some behaviour can only be settled in a real build on a real machine: what the window and the setup program look like, how a saved PDF reads, keyboard focus inside WebView2, the install, update and uninstall routes and the Linux and macOS packages. Those are checked by hand before a release rather than by the suite.

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): what the structural tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [REQUIREMENTS.md](REQUIREMENTS.md): each requirement names the test that
  verifies it.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions the product rests on and what each costs.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks like debt.
