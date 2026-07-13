package main

import "syscall/js"

func registrationStart(
	this js.Value,
	args []js.Value,
) js.Value {
	return js.ValueOf(nil)
}

func main() {
	js.Global().Set(
		"opaqueRegistrationStart",
		js.FuncOf(registrationStart),
	)

	select {}
}