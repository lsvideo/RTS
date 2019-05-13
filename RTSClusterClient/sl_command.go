// sl_command
package main

type CommandFunc func(task Task)

var mapCommands = make(map[string]CommandFunc, 20)

func RegisterCommand(cmd string, f CommandFunc) {
	mapCommands[cmd] = f
}

func RunCommand(cmd string, task Task) {
	f := mapCommands[cmd]
	f(task)
}
