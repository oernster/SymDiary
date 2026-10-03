# Verifies SymDiary: formatting, vet, staticcheck, the Go suite and its coverage
# floors, then the front end's lint, types and suite.
#
#   ./test.ps1                 run everything
#   ./test.ps1 -SkipFrontend   the Go half alone, while working on it
#   ./test.ps1 -Floor 95       a different floor for the gated layers, for a deliberate check
#
# build.ps1 runs this before it builds. Its one escape is -Fast, which is for a
# working loop and prints that it took it, so a release is never cut from a tree
# nobody verified.
#
# Floors are measured numbers, never targets. A floor picked from an aspiration
# only teaches people to lower it; a floor at the measured number fails the
# moment cover is lost, which is the only moment it is worth being told.
param(
    [double]$Floor = 100,
    [switch]$SkipFrontend
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

# The gate covers the layers that can be exercised with no filesystem, no clock
# and no window: the rules and the use cases. Anything short there is a decision
# nobody made. The rest of the tree is held at what it actually reaches, below.
$gated = './internal/domain/...', './internal/application/...'

$packages = go list ./... | Where-Object { $_ -notmatch '/node_modules/' }
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }

Write-Host 'Checking formatting...'
$unformatted = gofmt -l . | Where-Object { $_ -notmatch '^frontend' }
if ($unformatted) { throw "gofmt reports unformatted files:`n$($unformatted -join "`n")" }

Write-Host 'Vetting...'
go vet $packages
if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }

# SymDiary ships to Windows, Linux and macOS, so a Windows-only import is a
# defect the day it is written rather than the day someone tries the Flatpak.
# Building for the other two is the cheapest way to find one: it needs no
# Linux machine and no Mac; it fails by name. Vet runs with it, since a
# build says only that the package compiles.
Write-Host 'Building for Linux and macOS...'
# tests/structural holds test files alone, which go build refuses to be handed.
# It is still vetted below, where the test files are compiled.
$buildable = go list -f '{{if .GoFiles}}{{.ImportPath}}{{end}}' ./...
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }
$buildable = $buildable | Where-Object { $_ }
foreach ($target in 'linux', 'darwin') {
    $env:GOOS = $target
    try {
        go build $buildable
        if ($LASTEXITCODE -ne 0) { throw "the $target build failed with exit code $LASTEXITCODE" }
        go vet $packages
        if ($LASTEXITCODE -ne 0) { throw "go vet for $target failed with exit code $LASTEXITCODE" }
    } finally {
        Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    }
    Write-Host "  $target builds and vets clean."
}

Write-Host 'Running staticcheck...'
go run honnef.co/go/tools/cmd/staticcheck@latest $packages
if ($LASTEXITCODE -ne 0) { throw "staticcheck failed with exit code $LASTEXITCODE" }

Write-Host 'Running the whole Go suite...'
go test $packages
if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }

Write-Host "Measuring coverage of $($gated -join ', ')..."
$profilePath = Join-Path ([System.IO.Path]::GetTempPath()) 'symdiary-coverage.out'
try {
    go test "-coverprofile=$profilePath" @gated
    if ($LASTEXITCODE -ne 0) { throw "the coverage run failed with exit code $LASTEXITCODE" }

    $summary = go tool cover "-func=$profilePath"
    if ($LASTEXITCODE -ne 0) { throw "go tool cover failed with exit code $LASTEXITCODE" }

    # The last line is the total across the merged profile. Read the exit code
    # and this line, never the run's own output.
    $total = ($summary | Select-Object -Last 1)
    if ($total -notmatch '([0-9]+(?:\.[0-9]+)?)%\s*$') { throw "could not read a total from: $total" }
    $percent = [double]$Matches[1]

    if ($percent -lt $Floor) {
        Write-Host 'Not covered:'
        $summary | Where-Object { $_ -notmatch '100\.0%\s*$' -and $_ -notmatch '^total:' } |
            ForEach-Object { Write-Host "  $_" }
        throw "coverage is $percent%, below the floor of $Floor%"
    }
    Write-Host "Coverage $percent%, floor $Floor%."
} finally {
    if (Test-Path $profilePath) { Remove-Item $profilePath -Force }
}

# The rest of the tree, each package at the number it actually reaches.
#
# The root package holds the facade, which is tested, plus main, the log
# handover, the file dialogs, the browser opener and the single-instance lock,
# which need a window and a platform. Its floor moved from 77 to 76 when the
# donate button landed: the opener is one more line of Wails runtime that no
# test can reach, of exactly the same kind as the file dialogs beside it, so the
# blend fell rather than the facade going untested. Re-measured, not estimated.
# It moved 76 to 77 when printing through the window was replaced by saving a
# PDF: the bound method now decides something (the range, the name it suggests,
# the order it asks in) where before it forwarded one call to the runtime.
# The run log moved the other way, 74 to 81, when its folder rule became a pure
# function taking the platform as an argument: all three platforms' answers are
# now exercised on whichever platform the suite runs on. The run log needs a windowed process with no error output, so
# its crash tests start a child; what remains uncovered there is the Win32
# handle work. The store and the export file reach everything but a handful of
# operating-system write failures that cannot be forced without breaking the
# disk. The setup package moved 56 to 59 when the Windows theme read went: setup
# opens dark and carries its own toggle, so the registry lookup nothing could
# exercise is gone rather than sitting there lowering the number. It moved 59
# to 62 when what an uninstall takes became a function with tests of its own,
# then 62 to 64 when the install and removal SEQUENCES moved into it behind a
# Machine seam and got tests of their own. The setup program itself enters the
# table at the same moment and for the same reason: it had no test files at all
# while its facade owned the sequences; once they moved out what is left is
# a facade with a seam, so what it decides (the route it opens on, the choices
# it hands over) is reachable. What is not is the Wails runtime underneath:
# emitting a progress event and quitting the window.
# The document the reader takes away is drawn in internal/infrastructure/pdf,
# which needs no window, no clock and no device: it is handed a record and a
# path and answers a file, so it is gated at what it actually reaches. The one
# statement it does not cover is a font the library refuses to load, which
# cannot happen to a font compiled into the binary.
# TESTING.md names each shortfall.
$measured = [ordered]@{
    '.'                                      = 77
    './installer'                            = 69
    './internal/infrastructure/store'        = 92
    './internal/infrastructure/export'       = 91
    './internal/infrastructure/runlog'       = 81
    './internal/infrastructure/pdf'          = 99
    './internal/infrastructure/setup'        = 64
}

Write-Host 'Measuring the rest of the tree...'
foreach ($package in $measured.Keys) {
    $floor = $measured[$package]
    $reported = go test -cover $package
    if ($LASTEXITCODE -ne 0) { throw "$package failed with exit code $LASTEXITCODE" }

    $line = $reported | Where-Object { $_ -match 'coverage: ' } | Select-Object -First 1
    if ($line -notmatch 'coverage: ([0-9]+(?:\.[0-9]+)?)%') {
        throw "could not read a coverage figure for ${package}: $line"
    }
    $reached = [double]$Matches[1]
    if ($reached -lt $floor) {
        throw "$package is at $reached%, below its floor of $floor%"
    }
    Write-Host ("  {0,-42} {1,5}%  floor {2}%" -f $package, $reached, $floor)
}

# Not gated at all, deliberately: internal/product holds nine constants,
# tests/structural is itself the guard and setup/setuptest is the double the
# other two suites are written against. A floor over any of them asserts
# nothing.

if ($SkipFrontend) {
    Write-Host 'All green (the front end was skipped).'
    return
}

Write-Host 'Checking the front end: lint, types, build...'
npm --prefix frontend run build
if ($LASTEXITCODE -ne 0) { throw "the front-end build failed with exit code $LASTEXITCODE" }

Write-Host 'Running the front-end suite...'
npm --prefix frontend test
if ($LASTEXITCODE -ne 0) { throw "the front-end suite failed with exit code $LASTEXITCODE" }

# The setup program's page has no build step, so nothing compiles it and a
# typo there reaches a user as a window that draws no screen at all. This is
# the cheapest check that exists: node parses each file without running it.
# It is not a lint and it is not a type check; tests/structural/setuppage_test.go
# checks the ids and state fields the scripts read against the markup and the
# setup facade.
Write-Host 'Parsing the setup page...'
foreach ($script in Get-ChildItem -Path 'installer/frontend/dist' -Filter '*.js') {
    node --check $script.FullName
    if ($LASTEXITCODE -ne 0) { throw "$($script.Name) does not parse, exit code $LASTEXITCODE" }
}

Write-Host 'All green.'
