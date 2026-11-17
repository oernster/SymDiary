//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"syscall/js"

	"github.com/oernster/symdiary/internal/facade"
)

// call is one bound method: it reads its arguments from the page and answers a
// value to hand back, else a refusal.
type call func(args []js.Value) (any, error)

// errMissingArgument refuses a call the page made with too few arguments.
var errMissingArgument = errors.New("the page sent too few arguments")

// install puts the facade where the page looks for the desktop's, with the
// same method names, so api.ts reaches either window the same way.
func install(core *facade.Core, outbox *facade.Outbox) {
	app := js.Global().Get("Object").New()
	methods := map[string]call{
		"State":  func([]js.Value) (any, error) { return core.State() },
		"Now":    func([]js.Value) (any, error) { return core.Now() },
		"About":  func([]js.Value) (any, error) { return core.About() },
		"Record": withForm(core.Record),
		"Suggest": func(args []js.Value) (any, error) {
			var typed string
			if err := decode(args, 0, &typed); err != nil {
				return nil, err
			}
			return core.Suggest(typed)
		},
		"Symptoms": func([]js.Value) (any, error) { return core.Symptoms() },
		"History": func(args []js.Value) (any, error) {
			var filter facade.FilterDTO
			if err := decode(args, 0, &filter); err != nil {
				return nil, err
			}
			return core.History(filter)
		},
		"Edit": func(args []js.Value) (any, error) {
			var id int64
			var form facade.EditFormDTO
			if err := errors.Join(decode(args, 0, &id), decode(args, 1, &form)); err != nil {
				return nil, err
			}
			return core.Edit(id, form)
		},
		"DeletionPrompt": withIDs(core.DeletionPrompt),
		"Delete": withIDs(func(ids []int64) (any, error) {
			return nil, core.Delete(ids)
		}),
		"Waiting": func([]js.Value) (any, error) { return outbox.Waiting() },
		"Prepare": func([]js.Value) (any, error) { return outbox.Prepare() },
		"Again":   func([]js.Value) (any, error) { return outbox.Again() },
		"Confirm": withIDs(func(ids []int64) (any, error) {
			return nil, outbox.Confirm(ids)
		}),
	}
	for name, method := range methods {
		app.Set(name, bind(method))
	}
	js.Global().Set("go", map[string]any{"main": map[string]any{"App": app}})
}

// withForm adapts a method taking the recording form.
func withForm(method func(facade.RecordFormDTO) (facade.EventDTO, error)) call {
	return func(args []js.Value) (any, error) {
		var form facade.RecordFormDTO
		if err := decode(args, 0, &form); err != nil {
			return nil, err
		}
		return method(form)
	}
}

// withIDs adapts a method taking a list of event ids.
func withIDs[T any](method func([]int64) (T, error)) call {
	return func(args []js.Value) (any, error) {
		var ids []int64
		if err := decode(args, 0, &ids); err != nil {
			return nil, err
		}
		return method(ids)
	}
}

// bind turns a call into a function the page can await. The work runs on a
// goroutine of its own, because the phone's storage answers through the
// browser's own thread, which a call made on that thread would block forever.
// A panic there is recovered, logged and refused, never left to end the page
// (house robustness rule 7).
func bind(method call) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		executor := js.FuncOf(func(_ js.Value, settle []js.Value) any {
			resolve, reject := settle[0], settle[1]
			go func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						fmt.Fprintf(os.Stderr, "panic in a bound method: %v\n%s", recovered, debug.Stack())
						reject.Invoke(js.Global().Get("Error").New(facade.ErrInternal.Error()))
					}
				}()
				answer, err := method(args)
				if err != nil {
					reject.Invoke(js.Global().Get("Error").New(err.Error()))
					return
				}
				resolve.Invoke(encode(answer))
			}()
			return nil
		})
		// The executor runs inside the constructor, so it can go at once.
		defer executor.Release()
		return js.Global().Get("Promise").New(executor)
	})
}

// decode reads argument at into target by way of JSON, as Wails does.
func decode(args []js.Value, at int, target any) error {
	if at >= len(args) {
		return errMissingArgument
	}
	text := js.Global().Get("JSON").Call("stringify", args[at])
	if text.IsUndefined() {
		return errMissingArgument
	}
	return json.Unmarshal([]byte(text.String()), target)
}

// encode hands an answer to the page by way of JSON; undefined for none.
func encode(answer any) js.Value {
	if answer == nil {
		return js.Undefined()
	}
	data, err := json.Marshal(answer)
	if err != nil {
		return js.Global().Get("Error").New(err.Error())
	}
	return js.Global().Get("JSON").Call("parse", string(data))
}
