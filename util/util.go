package util

import (
	"github.com/gen2brain/beeep"
)

func Notify(title, message, level string) {
	switch level {
	case "info":
		beeep.Notify(title, message, "assets/information.png")
	case "warning":
		beeep.Alert(title, message, "assets/warning.png")
	case "error":
		beeep.Alert(title, message, "assets/error.png")
	default:
		beeep.Notify(title, message, "assets/information.png")
	}
}
