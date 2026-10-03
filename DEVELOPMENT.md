# SymDiary: development

Every command here is PowerShell, run from the repository root.

## Tools

| Tool | What for | Where from |
|---|---|---|
| Go 1.26+ | The application | https://go.dev/dl/ |
| Node 20+ with npm | The page | https://nodejs.org/ |
| Wails v2 CLI | Building the window | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| WebView2 runtime | Running the window | Ships with Windows 11; where it is missing or too old, as it can be on Windows 10, the application and the setup program offer to download Microsoft's installer for it |
| Python 3 with Pillow | Regenerating the icons; stamping the site (Python alone) | `python -m pip install pillow` |

Versions measured on the reference machine on 2026-09-22: Go 1.26.3, Wails
v2.12.0, Node 24.11.1.

## Building

```powershell
./build.ps1
```

In order, it:

1. Reads `VERSION`.
2. Runs `test.ps1` and stops on failure, unless `-Fast` was given.
3. Runs `wails build -ldflags "-X main.appVersion=<version>"`, which installs
   the front-end dependencies, builds the page and compiles the application.
4. Checks the executable is there and prints its path and size.
5. Copies `LICENSE` beside the executable, so the licence lands in the install
   folder, then zips `build/bin` into `installer/payload.zip`.
6. Builds the setup program from `installer/`, with the same version flag.
7. Copies it to `dist-installer/SymDiarySetup.exe`, then puts the empty-zip
   placeholder back in `installer/payload.zip`, so a payload of megabytes never
   reaches a commit.

`./build.ps1 -SkipInstaller` stops after the application.

`./build.ps1 -Fast` skips the gate for a working loop. It says so in its output
and is never how a release is cut.

The version reaches the binary through `-ldflags -X`, which writes to a `var`
only. `main.appVersion` is declared `var` for that reason: against a `const` the
flag silently does nothing.

### Linux and macOS

Neither is built from Windows; each is built on its own machine, from the
repository root, with bash rather than PowerShell:

```bash
bash build_flatpak.sh
```

Ported from PigeonPost's. It installs the flatpak tooling where it is missing,
adds flathub, pulls the GNOME runtime (which is what supplies the
webkit2gtk-4.1 that Wails renders through, so the Go build carries
`-tags webkit2_41`), writes the desktop entry, the metainfo and the manifest,
builds inside the sandbox and exports `symdiary.flatpak`. `bash
cleanup_flatpak.sh` uninstalls it and removes what the build left; it never
touches the record.

The finished application is given no network permission at all. SymDiary opens no
connection; the sandbox is where that stops being a claim and becomes a
rule: a build that started talking to something would fail at run time rather
than quietly working. The build itself does get the network, since it fetches Go
modules and npm packages, which is why `--share=network` appears under
`build-args` and nowhere else.

```bash
bash builddmg.sh
```

Also ported from PigeonPost's, for Apple Silicon. It builds, stamps the bundle
version from `VERSION`, signs with a Developer ID, notarizes and staples both
the app and the DMG, then replays Gatekeeper's own check. Notarization is not
optional: since macOS 10.15 a signed but unnotarized app is refused on every
machine but the one that signed it; the failure is invisible at build time.
`ALLOW_UNNOTARIZED=1` exists for a local test build and for nothing else.

Both have now been run on their own platform: the Flatpak builds and its window
opens; the DMG builds and is notarized. What the gate adds on every run is
that the module compiles and vets for both, so nothing Windows-only can be
written without the next test run saying so.

## Running from source

```powershell
wails dev
```

The window opens with the page served by Vite, so an edit to `frontend/src`
appears at once. It reads your real record at `%APPDATA%\SymDiary\symdiary.db`.
To leave that alone, run the built executable with a sandboxed folder instead:

```powershell
$env:APPDATA = "$env:TEMP\symdiary-sandbox\Roaming"; $env:LOCALAPPDATA = "$env:TEMP\symdiary-sandbox\Local"; ./build/bin/SymDiary.exe
```

The log is at `%LOCALAPPDATA%\SymDiary\SymDiary.log`. It holds the run's start
line and any crash; it never holds a symptom, a note or anything else from the
record.

## The generated assets

The artwork masters live in `assets/`, plus `donate.png` at the root.

```powershell
python tools/genicons.py
```

It writes the band icons into `frontend/src/assets/icons`, the multi-size
`build/windows/icon.ico` that Wails puts on the executable, plus `build/appicon.png`, which Wails fills with its own logo when the file is
absent, plus `build/linux/icons`, the eight hicolor sizes a Linux desktop
chooses between, which the Flatpak manifest installs one by one. The output is
committed, so a clone needs neither Python nor Pillow; the Flatpak build needs
no Python inside its sandbox either.

The donate mark takes a path of its own. It is a picture rather than an icon, so
it is cropped to its artwork and scaled by height alone, keeping its width; the
square canvas every band icon is centred on would spend the difference on
nothing. It shares the icons' size constant, so the mark and the pictures beside
it in the bar cannot drift apart.

Run it whenever a master changes.

## The website

The site is plain HTML under `docs/`, served by GitHub Pages: a home page plus
Features, Why and Get it, sharing `styles.css` and `site.js`. Each page links
those two files with a hash of their content, so a browser fetches the new file
the moment it changes rather than pairing a new page with a cached old
stylesheet. After editing either file:

```powershell
python stamp_assets.py
```

It rewrites only the links whose hash has changed, says which pages it touched
and leaves an up-to-date site alone. The hash reads CRLF as LF, so a Windows
checkout and the copy GitHub serves agree.

No version is written into the site. Every download button points at the
newest release's files and `site.js` asks GitHub for that release's version
and file sizes as the page opens; without an answer the page still reads
correctly.

## Versioning

`VERSION` at the root is the only place a version is written. The build reads
it; nothing else holds one.

## Cutting a release

1. Clear `TECH_DEBT.md` if one is open.
2. Bump `VERSION` if a bump is owed against the newest tag.
3. `./build.ps1` and check the gate is green.
4. Launch the built executable and use it: record, save a PDF, export, import.
5. Run `dist-installer/SymDiarySetup.exe` and walk each route: install, reopen
   for manage, repair, then uninstall, then run the checks only a person can
   settle; [TESTING.md](TESTING.md) says which kinds of behaviour those cover.
6. Commit, tag and publish. Those are the owner's to run.

To try setup without installing onto your own machine, start it with the two
data folders pointed elsewhere. It still writes the real `HKCU` uninstall entry,
which its own uninstall removes again.

## The standing rules a first change has to know

- Structure before code. Every behaviour has a numbered requirement in
  [REQUIREMENTS.md](REQUIREMENTS.md) with acceptance criteria; a change to
  behaviour amends the specification first.
- The domain performs no IO and never reads the clock.
- Only `main.go` wires infrastructure to the application.
- No file over 400 lines; none left between 381 and 400.
- Every colour goes in `frontend/src/theme.css` and nowhere else.
- A shape crossing the window boundary is written twice, in `dto.go` and in
  `frontend/src/api.ts`. The structural test compares them.
- A call from the page takes a refusal handler as its last argument and answers
  null when refused, so a call that ignores a refusal does not compile.
- A new guard is not a guard until a planted violation has made it fail.

## Further reading

- [ARCHITECTURE.md](ARCHITECTURE.md): the invariants.
- [TESTING.md](TESTING.md): the gate and the floors.
- [REQUIREMENTS.md](REQUIREMENTS.md): what it must do.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions the product rests on and what each costs.
- [TECH_DEBT.md](TECH_DEBT.md): what is still open, what is deliberately left and what only looks like debt.
