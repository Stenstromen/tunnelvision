package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/beeep"
)

func main() {
	a := app.New()

	if desk, ok := a.(desktop.App); ok {
		desk.SetSystemTrayMenu(buildMenu(desk, false, a))
	}

	a.Run()
}

func buildMenu(desk desktop.App, started bool, a fyne.App) *fyne.Menu {
	menu1 := fyne.NewMenuItem("start", func() {
		settings := loadSettings()

		if settings != nil {
			os.Setenv("BOUNDARY_ADDR", settings["BOUNDARY_ADDR"])
			os.Setenv("BOUNDARY_CACERT", settings["BOUNDARY_CACERT"])
			os.Setenv("BOUNDARY_TLS_SERVER_NAME", settings["BOUNDARY_TLS_SERVER_NAME"])
			os.Setenv("BOUNDARY_PASS", settings["BOUNDARY_PASS"])
		}

		w := a.NewWindow("New Window")
		w.Resize(fyne.NewSize(400, 300))

		usernameEntry := widget.NewEntry()
		usernameEntry.SetPlaceHolder("Enter USERNAME")

		targetIDEntry := widget.NewEntry()
		targetIDEntry.SetPlaceHolder("Enter TARGETID")

		localEntry := widget.NewEntry()
		localEntry.SetPlaceHolder("Enter LOCAL")

		remoteEntry := widget.NewEntry()
		remoteEntry.SetPlaceHolder("Enter REMOTE")

		runButton := widget.NewButton("Set Envs and Run SSH Command", func() {
			portForwards := []string{
				"8080:localhost:8080",
				"8081:localhost:8081",
			}

			BoundaryTunnel("/opt/homebrew/bin/boundary", usernameEntry.Text, targetIDEntry.Text, portForwards)
		})

		w.SetContent(container.NewVBox(
			usernameEntry,
			targetIDEntry,
			localEntry,
			remoteEntry,
			runButton,
		))

		w.Show()

		desk.SetSystemTrayMenu(buildMenu(desk, true, a))
	})

	menu2 := fyne.NewMenuItem("Settings", func() {
		showSupportWindow(a)
	})

	menu3 := fyne.NewMenuItem("Quit", func() {
		a.Quit()
	})

	return fyne.NewMenu("Tunnelvision", menu1, menu2, menu3)
}

func showSupportWindow(a fyne.App) {
	w := a.NewWindow("Support Settings")
	w.Resize(fyne.NewSize(400, 300))

	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("Enter BOUNDARY_ADDR")
	cacertEntry := widget.NewEntry()
	cacertEntry.SetPlaceHolder("Enter BOUNDARY_CACERT")
	tlsServerNameEntry := widget.NewEntry()
	tlsServerNameEntry.SetPlaceHolder("Enter BOUNDARY_TLS_SERVER_NAME")
	passEntry := widget.NewEntry()
	passEntry.SetPlaceHolder("Enter BOUNDARY_PASS")
	passEntry.Password = true

	settings, err := os.ReadFile(settingsFile)
	if err == nil {
		lines := strings.Split(string(settings), "\n")
		if len(lines) >= 4 {
			addrEntry.SetText(lines[0])
			cacertEntry.SetText(lines[1])
			tlsServerNameEntry.SetText(lines[2])
			passEntry.SetText(lines[3])
		}
	}

	saveButton := widget.NewButton("Save Settings", func() {
		saveSettings(addrEntry.Text, cacertEntry.Text, tlsServerNameEntry.Text, passEntry.Text)
		beeep.Notify("Settings Saved", "Your settings have been saved successfully.", "")
	})

	w.SetContent(container.NewVBox(
		addrEntry,
		cacertEntry,
		tlsServerNameEntry,
		passEntry,
		saveButton,
	))

	w.Show()
}

const settingsFile = "/Users/filip/boundary_settings.txt"

func saveSettings(addr, cacert, tlsServerName, pass string) {
	content := fmt.Sprintf("%s\n%s\n%s\n%s", addr, cacert, tlsServerName, pass)
	err := os.WriteFile(settingsFile, []byte(content), 0644)
	if err != nil {
		beeep.Alert("Error", "Failed to save settings: "+err.Error(), "")
	}
}

func loadSettings() map[string]string {
	content, err := os.ReadFile(settingsFile)
	if err != nil {
		return nil
	}

	settingsMap := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	if len(lines) >= 4 {
		settingsMap["BOUNDARY_ADDR"] = lines[0]
		settingsMap["BOUNDARY_CACERT"] = lines[1]
		settingsMap["BOUNDARY_TLS_SERVER_NAME"] = lines[2]
		settingsMap["BOUNDARY_PASS"] = lines[3]
	}

	return settingsMap
}

func BoundaryTunnel(BOUNDARY_PATH string, USERNAME string, TARGETID string, portForwards []string) {
	fmt.Println("BOUNDARY_PATH: " + BOUNDARY_PATH)
	var stdoutBuf, stderrBuf bytes.Buffer

	args := []string{"connect", "ssh", "-username", USERNAME, "-target-id", TARGETID, "--"}
	for _, portForward := range portForwards {
		args = append(args, "-L", portForward)
	}

	cmd := exec.Command(BOUNDARY_PATH, args...)
	fmt.Println("cmd: " + strings.Join(cmd.Args, " "))
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fullOutput := "STDOUT:\n" + stdoutBuf.String() + "\nSTDERR:\n" + stderrBuf.String()
		beeep.Alert("Boundary Command Error", fullOutput, "")
		return
	}

	message := stdoutBuf.String()
	fmt.Println(message)
	beeep.Notify("Boundary Command Output", message, "assets/information.png")
}
