package util

import (
	"github.com/gen2brain/beeep"
	"github.com/stenstromen/tunnelvision/types"
)

const (
	InfoLevel    types.Level = "info"
	WarningLevel types.Level = "warning"
	ErrorLevel   types.Level = "error"
)

func Notify(title, message string, level types.Level) {
	switch level {
	case InfoLevel:
		beeep.Notify(title, message, "assets/information.png")
	case WarningLevel:
		beeep.Alert(title, message, "assets/warning.png")
	case ErrorLevel:
		beeep.Alert(title, message, "assets/error.png")
	}
}
