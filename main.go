//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/firu11/git-calendar-core/gitcalendarcore"
)

func main() {
	api := gitcalendarcore.NewApi()

	js.Global().Set("GitCalendar", js.ValueOf(map[string]any{
		"initialize": js.FuncOf(func(this js.Value, args []js.Value) any {
			path := args[0].String()
			// Return a Promise so JS doesn't freeze
			handler := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) any {
				resolve := pArgs[0]
				go func() {
					err := api.Initialize(path)
					if err != nil {
						resolve.Invoke(err.Error())
					} else {
						resolve.Invoke(js.Null())
					}
				}()
				return nil
			}))
			return handler
		}),
		"addEvent": js.FuncOf(func(this js.Value, args []js.Value) any {
			data := args[0].String()
			handler := js.Global().Get("Promise").New(js.FuncOf(func(this js.Value, pArgs []js.Value) any {
				resolve := pArgs[0]
				go func() {
					err := api.AddEvent(data)
					if err != nil {
						resolve.Invoke(err.Error())
					} else {
						resolve.Invoke(js.Null())
					}
				}()
				return nil
			}))
			return handler
		}),
	}))

	select {}
}
