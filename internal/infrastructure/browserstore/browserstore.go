//go:build js && wasm

package browserstore

import (
	"errors"
	"syscall/js"
)

// The database, the object store inside it and the one key the snapshot sits
// under. One record, one value: a write replaces it whole (memstore).
const (
	databaseName    = "symdiary"
	databaseVersion = 1
	storeName       = "record"
	snapshotKey     = "snapshot"
)

// ErrNoStorage refuses a browser that offers no IndexedDB.
var ErrNoStorage = errors.New("this browser keeps no storage for SymDiary")

// Snapshots keeps the snapshot in IndexedDB. Every call blocks the goroutine
// that makes it until the browser answers, so it must never be called on the
// browser's own thread: the bridge runs every call on a goroutine of its own.
type Snapshots struct {
	db js.Value
}

// Open opens the database, creating it on the first run.
func Open() (*Snapshots, error) {
	factory := js.Global().Get("indexedDB")
	if factory.IsUndefined() || factory.IsNull() {
		return nil, ErrNoStorage
	}
	request := factory.Call("open", databaseName, databaseVersion)
	upgrade := js.FuncOf(func(js.Value, []js.Value) any {
		request.Get("result").Call("createObjectStore", storeName)
		return nil
	})
	defer upgrade.Release()
	request.Set("onupgradeneeded", upgrade)
	db, err := await(request)
	if err != nil {
		return nil, err
	}
	return &Snapshots{db: db}, nil
}

// Load answers the saved snapshot; nil when nothing has been saved yet.
func (s *Snapshots) Load() ([]byte, error) {
	transaction := s.db.Call("transaction", storeName, "readonly")
	value, err := await(transaction.Call("objectStore", storeName).Call("get", snapshotKey))
	if err != nil {
		return nil, err
	}
	if value.IsUndefined() || value.IsNull() {
		return nil, nil
	}
	return []byte(value.String()), nil
}

// Save replaces the snapshot, answering only once the transaction has
// committed, so an event reported as recorded is already kept (NFR-REL-002).
func (s *Snapshots) Save(data []byte) error {
	options := js.Global().Get("Object").New()
	options.Set("durability", "strict")
	transaction := s.db.Call("transaction", storeName, "readwrite", options)
	transaction.Call("objectStore", storeName).Call("put", string(data), snapshotKey)
	return committed(transaction)
}

// await waits for a request's success or error.
func await(request js.Value) (js.Value, error) {
	done := make(chan js.Value, 1)
	failed := make(chan error, 1)
	success := js.FuncOf(func(js.Value, []js.Value) any {
		done <- request.Get("result")
		return nil
	})
	failure := js.FuncOf(func(js.Value, []js.Value) any {
		failed <- reason(request.Get("error"))
		return nil
	})
	defer success.Release()
	defer failure.Release()
	request.Set("onsuccess", success)
	request.Set("onerror", failure)
	select {
	case value := <-done:
		return value, nil
	case err := <-failed:
		return js.Undefined(), err
	}
}

// committed waits for a transaction to complete, fail or be aborted.
func committed(transaction js.Value) error {
	finished := make(chan error, 1)
	complete := js.FuncOf(func(js.Value, []js.Value) any {
		finished <- nil
		return nil
	})
	failure := js.FuncOf(func(js.Value, []js.Value) any {
		finished <- reason(transaction.Get("error"))
		return nil
	})
	defer complete.Release()
	defer failure.Release()
	transaction.Set("oncomplete", complete)
	transaction.Set("onerror", failure)
	transaction.Set("onabort", failure)
	return <-finished
}

// reason turns a DOMException into an error naming it.
func reason(exception js.Value) error {
	if exception.IsUndefined() || exception.IsNull() {
		return errors.New("the browser's storage refused without saying why")
	}
	return errors.New(exception.Get("name").String() + ": " + exception.Get("message").String())
}
