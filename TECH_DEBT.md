# SymDiary: Technical Debt

A standing reference to the project's outstanding technical debt. It records what is still open, weighs whether each item is worth doing and gives the rationale. Every item is a behaviour-preserving internal concern: nothing here proposes reverting a feature or changing any UI or UX behaviour. Scope is the whole repository (the Go core, the React front end, the setup program with its hand-written page and the delivery scripts) read against `ARCHITECTURE.md`, `TESTING.md` and the structural tests in `tests/structural`.

The sections below the open items are the standing record of what was weighed and deliberately left alone, so the same ground is not covered twice. They carry no numbers, because a number here means an open item.

There is no open technical debt.

---

## Looks like debt, not worth touching

**The setup page living under `dist` with no bundler.** It reads as generated output and is not; the structural tests name it explicitly for that reason. Keeping it hand written is what makes the setup program a single file with no build step of its own, which is the whole of the distribution story. What that costs is checking, not bundling; the checking is now done from the outside: `tests/structural/setuppage_test.go` reads every `$('id')` the scripts ask for against the ids the markup declares, plus every field they read off the Go state against the json tags the setup facade sends. Both were proved to bite by planting the rename that used to pass everything.

**The setup package's coverage floor at 64.** The uncovered half is the registry writes, the process work and the Win32 handle calls, which change the machine they run on. A test double for them would assert that the double was called; the gate would rise on nothing. A real install exercises them and `TESTING.md` lists what a person has to look at.

**The setup program's floor at 69.** What is left uncovered there is the Wails runtime underneath the facade: emitting a progress event to a window and quitting one. Everything the facade decides sits above that and is tested against `setup.Machine`.

**`internal/infrastructure/windowfocus` having two tests that only prove it returns.** It is `EnumWindows`, `AttachThreadInput` and `SetFocus` against a window the suite does not own. The tests exist to pin the one property a harness can see, that a best-effort function stays best effort when there is no window to find. The behaviour that matters is settled by opening the built window and pressing Tab.

## Not debt (do not "fix" these)

**`Unavailable` repeating the same refusal across every `Store` method.** It looks like eleven copies of one line. Each one is a different method of the interface answering with the reason the record could not be opened, which is what keeps the window open and every action honest about why it cannot act (FR-062). Collapsing it would mean a smaller interface, not less code.

**`setup.Real` being a wrapper with one call per method.** It looks like a layer that does nothing. It is what lets the install and removal sequences be stated over an interface and tested against a recorder, while production still reaches the real registry, the real filesystem and the real process list. There is nothing in it for a test to hold, which is the point: the seam costs one line per operation and buys the order of operations being checkable at all.

**`quotedPath` and `literal` both quoting a path.** They are not a duplicate pair. `quotedPath` writes plain double quotes for the registry, because that is what every other entry in the Apps list carries. `literal` writes a PowerShell single-quoted string, doubling any apostrophe, because the shortcut is created through a PowerShell one-liner. Go's `%q` was used for both once and shipped a broken install: it escapes the separators, so every backslash reached the registry and the shortcut doubled. Two quoting rules need two functions.

**The frozen export samples, one per format version.** `internal/infrastructure/export` keeps a committed sample for every version ever written and reads each one on every run. They are near-identical files by design: the point is that a change to the reader that breaks an old export fails here rather than at a user who is importing a record they made last year.

**`appVersion` being a `var` rather than a `const`.** `VERSION` is read by `build.ps1` into `-ldflags -X` against a `var` and by `builddmg.sh` into the bundle's plist. A `const` would take the flag silently and keep the old value, which is why it is a `var` and why it looks like a mutable global that should be a constant.
