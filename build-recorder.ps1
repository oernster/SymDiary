# Builds the phone recorder (Amendment 21): one folder holding everything
# app.symdiary.com serves.
#
#   ./build-recorder.ps1        the gate, then the recorder
#   ./build-recorder.ps1 -Fast  no gate, for a working loop
#
# Output:
#   dist-recorder/   the whole site: index.html, its bundle, SymDiary's Go rules
#                    as symdiary-wasm.js, Go's wasm_exec.js, the manifest, the
#                    icons and sw.js, which keeps all of them for offline use
#
# The gate runs first, as build.ps1's does; -Fast is for a working loop and
# says so in its output.
#
# The rules arrive as a script holding the WebAssembly in base64, because the
# page's connect-src 'none' forbids fetching a .wasm file. The version reaches
# the binary through -ldflags -X, which only writes to a var.
param(
    [switch]$Fast
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$version = (Get-Content (Join-Path $root 'VERSION') -Raw).Trim()
if (-not $version) { throw 'VERSION is empty' }

if ($Fast) {
    Write-Host "Building the recorder $version without the gate (working build, not a release)..."
} else {
    & (Join-Path $root 'test.ps1')
    if ($LASTEXITCODE -ne 0) { throw "the gate failed with exit code $LASTEXITCODE" }
    Write-Host "Building the recorder $version..."
}

$templates = Join-Path $root 'build\recorder'
$site = Join-Path $root 'dist-recorder'
$utf8 = New-Object System.Text.UTF8Encoding $false

function Write-Text([string]$path, [string]$text) {
    [System.IO.File]::WriteAllText($path, $text, $utf8)
}

# SymDiary's rules, for the browser.
Write-Host 'Building the rules for WebAssembly...'
$wasm = Join-Path ([System.IO.Path]::GetTempPath()) 'symdiary-recorder.wasm'
$env:GOOS = 'js'
$env:GOARCH = 'wasm'
try {
    go build -ldflags "-s -w -X main.appVersion=$version" -o $wasm ./cmd/recorder
    if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }
} finally {
    Remove-Item Env:GOOS, Env:GOARCH
}
if (-not (Test-Path $wasm)) { throw "the build reported success but $wasm is not there" }

# Go's own loader for that binary, from the toolchain that built it.
$goRoot = go env GOROOT
$wasmExec = Join-Path $goRoot 'lib\wasm\wasm_exec.js'
if (-not (Test-Path $wasmExec)) { throw "this Go keeps no wasm_exec.js at $wasmExec" }

Write-Host 'Building the page...'
npm --prefix frontend run build:recorder
if ($LASTEXITCODE -ne 0) { throw "the page build failed with exit code $LASTEXITCODE" }
$page = Join-Path $root 'frontend\dist-recorder'
if (-not (Test-Path (Join-Path $page 'recorder.html'))) { throw "the page build left no recorder.html in $page" }

Write-Host 'Collecting the site...'
if (Test-Path $site) { Remove-Item $site -Recurse -Force }
Copy-Item $page $site -Recurse
# The site's root serves index.html.
Move-Item (Join-Path $site 'recorder.html') (Join-Path $site 'index.html')

$encoded = [Convert]::ToBase64String([System.IO.File]::ReadAllBytes($wasm))
Write-Text (Join-Path $site 'symdiary-wasm.js') "self.SYMDIARY_WASM_B64 = `"$encoded`";`n"
Copy-Item $wasmExec (Join-Path $site 'wasm_exec.js')
Remove-Item $wasm

Get-ChildItem $templates -Filter 'icon-*.png' | Copy-Item -Destination $site
if (-not (Get-ChildItem $site -Filter 'icon-*.png')) { throw "no icons in ${templates}: run python tools/genicons.py" }

# The manifest takes the page's own background, from its one home in the theme.
$theme = Get-Content (Join-Path $root 'frontend\src\theme.css') -Raw
if ($theme -notmatch ':root\s*\{[^}]*?--bg:\s*(#[0-9a-fA-F]+)') { throw 'theme.css holds no --bg under :root' }
$background = $Matches[1]
$manifest = (Get-Content (Join-Path $templates 'manifest.webmanifest') -Raw).Replace('__BACKGROUND__', $background)
Write-Text (Join-Path $site 'manifest.webmanifest') $manifest

# The worker keeps every file shipped, plus the root the home screen opens.
$files = @('./') + (Get-ChildItem $site -Recurse -File |
    ForEach-Object { $_.FullName.Substring($site.Length + 1).Replace('\', '/') } |
    Sort-Object)
$listed = '[' + (($files | ForEach-Object { "`"$_`"" }) -join ', ') + ']'
$worker = (Get-Content (Join-Path $templates 'sw.js') -Raw).Replace('__CACHE__', "symdiary-recorder-$version").Replace('__FILES__', $listed)
Write-Text (Join-Path $site 'sw.js') $worker

$bytes = (Get-ChildItem $site -Recurse -File | Measure-Object Length -Sum).Sum
Write-Host ("Built {0}: {1} files, {2:N0} bytes" -f $site, ($files.Count), $bytes)
