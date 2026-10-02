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
`tests/structural` is itself the guard; `internal/infrastructure/setup/setuptest`
is the double the other two suites are written against.

## What each suite proves

| Suite | What it settles |
|---|---|
| `internal/domain` | The rules: blank and future occurrences refused, case variants matched, filters combined, the receipt's grouping, ordering and exact lines, including a range that crosses a year. The framing too: it opens the record and appears exactly once, the statement prints under the range, then two different records are shown to carry framing identical word for word. |
| `internal/application` | The use cases over fakes: recording reads the clock once, an edit leaves recorded-at alone, a failed write loses nothing the user typed, an import skips what is held. |
| `internal/infrastructure/store` | Real SQLite in a temporary folder: notes and symptoms read back byte for byte, a deletion is all or nothing, a garbage file is never overwritten, a newer schema is refused, the durability pragmas are set, plus a planted trigger proving a failed write leaves nothing behind. |
| `internal/infrastructure/export` | The format against a committed sample, a refusal for anything that is not a SymDiary export, plus a size cap read before the file is. The version contract as well: every version ever written still reads its own frozen sample, no reader claims a version nothing wrote, a file naming no version is refused rather than read as the current one, plus a gap in the table refusing rather than guessing. |
| `internal/infrastructure/runlog` | Crash reporting, by starting this test binary again as a child and making it panic, plus each platform's rule for where the log lives, all three exercised wherever the suite runs. |
| root package | The facade end to end over a real record: every conversion, every refusal, plus a panic in a bound method becoming an error rather than a dead window. `window_test.go` holds the window's own seams: the Downloads folder both dialogs open in, plus the keyboard being asked for once on DOM ready; that asking returns at once because the Win32 work happens on a goroutine of its own. `receipt_test.go` holds the sheet on its own, with the framing's exact words written out a second time, so an edit to the line saying the record is not a diagnosis fails rather than ships. |
| `internal/infrastructure/setup` | The install policy: the payload fence refusing an entry that climbs out of the install folder, the paths, the version comparison that picks the route, plus real shortcuts written into temporary folders with plain paths rather than doubled separators. What an uninstall takes is covered separately, because it is the one part that cannot be undone: the record's folder is never among the folders cleared without asking, nor inside one of them; a folder the machine cannot place is left out rather than turned into a relative path. The scheduled removal of the install folder is run for real on Windows and waited for. |
| `internal/infrastructure/pdf` | The document: that a record of nothing is refused rather than written blank, that a path that cannot be written is reported, that a long record runs to the pages the layout said it would and that the file begins `%PDF-`. The layout is settled separately and without drawing anything, against a measurer that gives every character the same width: a note wraps without losing a word, a word wider than the page is left whole, an event is never split across two sheets, every page opens at the same height and no row reaches into the foot margin the page number sits in. |
| `tests/structural` | The invariants in ARCHITECTURE.md, plus one that is not an invariant of the code at all: the licence embedded in `internal/licence` is compared with the repository's own LICENSE byte for byte, because a setup program stating terms nobody granted is worse than one that says nothing. |
| `frontend` | The page over a fake facade: the keyboard path to a recorded event, the form keeping everything when a save is refused, an edit sending no time unless it changed, the confirmation naming what will go, plus the receipt showing exactly the lines it was given and carrying exactly one mark, in its first line. The self-reading cycle is covered twice: the pure machine tick by tick, then the hook under jsdom for what suspends it and what freezes it. The focus ring is covered the same way: the rules alone, then the ring against a real page, where Tab and Right agree, Shift+Tab and Left agree, both wrap, a disabled control is passed over and a text field keeps its arrows. |

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

These were measured by hand on the reference machines, against builds run with a
sandboxed `APPDATA` on Windows and against the Flatpak and the disk image on
their own machines. They need doing again when the window changes.

Where a row says it was measured in Chromium against the built page or
stylesheet, that was the reading available before a build existed; the owner has
since confirmed it in the running application, which the row says. A reading
taken against a stylesheet is not a reading of the program, so both are kept
rather than the second quietly replacing the first.

| Check | Last result |
|---|---|
| The window opens dark and the bar's button moves it to light and back, remembered across a restart. | Measured in Chromium against the built page on 2026-09-22: it opens dark with the sun and the words "Light mode", one press gives light with the moon and "Dark mode", then the choice is written to the window's storage. **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. |
| The setup program opens dark and its header toggle moves it to light and back. | Measured in Chromium against the setup page on 2026-09-22: dark with the sun, one press to light with the moon, the face and the words changing together and the choice stored. **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. |
| Wails' IPC works under the page's Content-Security-Policy. | Works: the severities shown came from Go. |
| Save PDF writes a document, where the reader chose, that another machine opens. | **A defect was found here on macOS 2026-09-23 and fixed.** Both file dialogs shared the one export filter; macOS applies a dialog filter to the name it is given, so a symptom record was saved as `record.pdf.json`: a name saying one thing, contents saying another and no viewer that would open it. Windows and Linux took the name as offered, which is why only macOS showed it. The dialog is now told which kind of file it is for, with the suite holding the suggested name, the filter and the title together; three planted violations each fail it. The folder each dialog opens in is held by `window_test.go`. **Confirmed by the owner on 2026-09-23**: on each platform a document of one to three pages was saved where it was asked for and opened correctly. |
| The document reads correctly: the letterhead, the statement, the groups, the severities and the notes, with `Page N of M` on every page. | Measured 2026-09-23 by writing a seven page record and reading the text positions back out of it: A4, every page numbered in order, the first line 19.3mm to 19.8mm from the top of every page, the left edge at exactly 16.0mm, the mark drawn once on page one and every page after the first opening with a complete event. Whether it LOOKS right is a reading only a person gives: **confirmed by the owner on 2026-09-23**, who saved documents of one to three pages on each platform and found them right. |
| A record long enough to run to several pages holds together. | Measured in the suite: no page begins with what was observed at an event whose time was on the page before; no row reaches into the foot margin the page number sits in. The owner read documents of one to three pages on each platform on 2026-09-23 and found them right, which is the length a record of a few weeks reaches. A document of forty pages or so has not been read by eye; the suite measures a seven page one, so what is unread is the long tail rather than the mechanism. |
| Recording, editing, deleting, exporting and importing through the real window. | All worked; the import skipped the event already held. |
| The Guide and About open, scroll and close. | They do. |
| The mark prints; a black and white copier makes something of it. | The document carries it: measured in the suite, then read back out of a written document as one image drawn on page one alone. Whether a printer and then a copier keep it legible is a reading only paper gives. The mark is covered in the suite, where exactly one appears and it is in the first line; whether a printer draws it, plus how a black and white copier treats it, is a reading only a print can give. |
| Setup's Licence screen explains the licence, then shows it in full and reads itself down. | Measured in Chromium against the built stylesheet first. The pane held still for the opening five seconds, then descended 50px in 4007ms of travel, which is the house pace exactly; a wheel event held it, then it carried on from where it stopped after the two and a half seconds of stillness rather than restarting. The pane drew no ring in any of the three states: at rest `outline none`, after focus with `hasFocus` confirmed `outline none`, after a click `outline none`, its border staying `rgb(58, 65, 76)` throughout. **Seen in the real setup window since, where it was wrong twice.** The wiring works and the pane reads itself; the window was two screens short of holding it: 720 had been measured in a browser, which has no title bar, while a real frame spends about forty pixels of it before the page sees any. The last statement was cut off and a scrollbar ran down the whole window. Side by side the licence also had half the width, while its own lines run to 78 characters, so every one of them wrapped a second time. The screen is stacked now and the window is 870. Measured at a client area of 840 by 820, which is what that gives the page: nothing overflows, each statement is one line at 27px, the pane's 739px of text width holds the licence's longest line in 549. The next build was the check that a real frame agrees, which it does: **confirmed in the setup window by the owner on 2026-09-23.** |
| The Rename buttons line up down the Symptoms screen. | Measured against the built stylesheet: laid out as a flex row the four buttons sat at x 78, 74, 228 and 34; as a grid they all sit at 162, including a name far longer than the column, which wraps inside it. **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. |
| The Guide and About hold still for five seconds, then read themselves down; they step aside the moment the reader scrolls. | **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. The cycle is covered by the suite and the pane's ring was measured in Chromium against the built stylesheet; neither on its own is a reading of the running application. |
| Tab reaches the Guide's text without drawing a ring round it; Close rings when it is reached. | Measured in Chromium against the built stylesheet: the pane draws nothing at rest, hovered, clicked or Tab-focused, while Close draws the 2px ring. **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. |
| The three-state ring: nothing at rest, green on hover or focus, permanent red while disabled. | Measured in Chromium against the built stylesheet, both modes, with `:hover` and `:focus-visible` confirmed each time. Dark: at rest `3px none` on the band, Show and an enabled Save PDF; hovered or Tab-focused `2px solid rgb(52,211,153)`; disabled Save PDF and Delete `2px solid rgb(255,138,128)` on a `rgb(42,48,58)` fill. Light: hovered `2px solid rgb(4,120,87)`, disabled `2px solid rgb(168,35,26)` on `rgb(228,232,237)`. The disabled Save PDF is skipped by Tab. **Confirmed in the running application by the owner on 2026-09-23.** What follows is how it was measured beforehand. |
| The window opens with nothing ringed; the first Tab enters the band. | **Found broken 2026-09-22, cause measured, fixed and confirmed 2026-09-23.** The owner reported that no Tab ever rang anything on a fresh build and install; one click on the page then made every Tab work. So the page never held the keyboard: WebView2 gives it DOM focus while the webview holds no keyboard focus, so no keydown reaches the listener. Chromium cannot show this and said so misleadingly: against the built page at 1100x760 the first Tab rang Record `rgb(52, 211, 153) solid 2px` correctly, while `document.hasFocus()` read false, which is the same defect wearing a pass. The first fix asked the Wails runtime to show the window and MISSED, because that focuses the main window while WebView2 hosts the page in a child window the keys follow. What is in the tree now is ported from PigeonPost: `internal/infrastructure/windowfocus` finds that child by its class and sets focus on it. **Measured in the built window by the owner on 2026-09-23: the first Tab rings Record in green without a click first, then every Tab after it steps the ring.** The fix holds. |
| Setup installs, with the options opening on what the machine already holds. | It does: files, registry entry, shortcuts, plus the application launched from the install folder. |
| Setup reopens on the manage screen when the versions match; a shortcut box applies immediately. | It does. |
| Setup started with `-uninstall`, as the Apps list starts it, opens on the removal screen. | It does. |
| Uninstall removes the program, its shortcuts, its log and the window's WebView2 folder, while keeping the record. | It does; the record was still there afterwards. |
| The Apps list's Uninstall and Modify point at a path that exists. | They do now. The first install wrote doubled separators; fixed and covered by a test. |
| The install folder is removed after setup exits. | The probe that found this is now a test rather than a memory: `setup.TestSchedulingARemovalActuallyRemovesTheFolder` schedules a removal over a real folder on Windows and waits for it to go, which it does in about two seconds. Proved to bite by putting the argument form back, where it removed nothing inside twenty. What no test reaches is the reason the removal is scheduled at all: on a real uninstall setup is running from inside the folder and holds its own executable open. **Confirmed by the owner on 2026-09-23** against the built setup program, with the record left alone: afterwards `%LOCALAPPDATA%\Programs\SymDiary` was gone and `%APPDATA%\SymDiary\symdiary.db` was still there. |
| Uninstall with "also delete my symptom record" ticked. | The decision is covered: `setup.Leftovers` is the list an uninstall clears without asking; `TestTheRecordIsNeverClearedWithTheLeftovers` fails if the record's folder is in it or inside anything in it, proved by planting it there. `TestTheRecordGoesOnlyWhenItIsAskedFor` writes a record where the screen says it is and removes it the way a tick does. The tick reaching the flag is read off `setup-routes.js`, which passes `read('record')` straight to `Uninstall`, since nothing type checks that page. **Confirmed by the owner on 2026-09-23** against the built setup program: a reinstall, then an uninstall with the box ticked, left both the install folder and `%APPDATA%\SymDiary\symdiary.db` gone. Read beside the run above it, where the box was left alone and the record survived, that is the pair: the tick decides it and nothing else does. |
| The Flatpak builds on a Linux machine and the window opens. | Built and run by the owner on 2026-09-22: it builds and the application runs. |
| The record lands under `~/.var/app/uk.codecrafter.SymDiary` inside the Flatpak. | Measured by the owner on 2026-09-23, after recording a symptom in the Flatpak build: the file is at `~/.var/app/uk.codecrafter.SymDiary/config/SymDiary/symdiary.db`, which is both the path the derivation predicted and the path `build_flatpak.sh` prints when it finishes. The derivation it confirms: the record is always the product's folder inside whatever the platform names as the place for a user's configuration, which `store.TestTheRecordSitsInTheFolderThePlatformNames` asserts on every platform; Flatpak redirects `XDG_CONFIG_HOME` into the sandbox and `os.UserConfigDir` reads it on Linux. |
| The DMG builds and notarizes on an Apple Silicon Mac. | Built and run by the owner on 2026-09-22, then notarized. |
| The notarisation ticket is stapled to the DMG and to the app inside it. | The build proves both or it fails. `builddmg.sh` runs `set -euo pipefail`, staples the bundle then validates it, then staples the DMG, validates it and replays Gatekeeper's own check with `spctl --assess`. So a DMG that reaches the repo root has been checked already; there is nothing left to confirm by hand. A notarized build with no ticket on it would still launch for whoever built it, then ask the network on someone else's machine. |

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): what the structural tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): building from source.
- [REQUIREMENTS.md](REQUIREMENTS.md): each requirement names the test that
  verifies it.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions the product rests on and what each costs.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks like debt.
