// Package browserstore keeps the phone recorder's snapshot in the browser's
// IndexedDB (Amendment 21). It exists only in the WebAssembly build; on every
// other platform the package is this comment alone, so the gate's builds and
// vets for Windows, Linux and macOS still see a package here.
package browserstore
