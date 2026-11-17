//go:build !(js && wasm)

// Command recorder is the phone recorder, which runs only in a browser. Built
// for any other platform it says so and stops; the gate still builds and vets
// it there, which is how a desktop-only import is caught.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "the SymDiary recorder runs only in a phone's browser; build it with build-recorder.ps1")
	os.Exit(1)
}
