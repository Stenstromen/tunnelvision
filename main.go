package main

import (
	"fmt"

	"github.com/stenstromen/tunnelvision/config"
	"github.com/stenstromen/tunnelvision/gui"

	"fyne.io/fyne/v2/app"
)

func main() {
	tunnelVision := app.New()

	hostsFile, err := config.GetHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}
	hosts, _ := config.LoadHostsFromFile(hostsFile)
	gui.UpdateTrayMenu(tunnelVision, hosts)

	tunnelVision.Run()
}
