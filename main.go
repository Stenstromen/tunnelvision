package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	var menu1 *fyne.MenuItem
	menu1 = fyne.NewMenuItem("start", func() {
		fmt.Println("started")
		w := a.NewWindow("New Window")
		//w.SetContent(widget.NewLabel("Hello, World!")) // Set the content of the window
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

		usernameEntry := widget.NewEntry()
		usernameEntry.SetPlaceHolder("Enter USERNAME")

		targetIDEntry := widget.NewEntry()
		targetIDEntry.SetPlaceHolder("Enter TARGETID")

		localEntry := widget.NewEntry()
		localEntry.SetPlaceHolder("Enter LOCAL")

		remoteEntry := widget.NewEntry()
		remoteEntry.SetPlaceHolder("Enter REMOTE")

		runButton := widget.NewButton("Set Envs and Run SSH Command", func() {
			// Set environment variables
			os.Setenv("BOUNDARY_ADDR", addrEntry.Text)
			os.Setenv("BOUNDARY_CACERT", cacertEntry.Text)
			os.Setenv("BOUNDARY_TLS_SERVER_NAME", tlsServerNameEntry.Text)
			os.Setenv("BOUNDARY_PASS", passEntry.Text)

			// Run SSH Command
			//runSSHCommand()
			BoundaryTunnel("/opt/homebrew/bin/boundary", usernameEntry.Text, targetIDEntry.Text, localEntry.Text, remoteEntry.Text)
		})

		// Layout
		w.SetContent(container.NewVBox(
			addrEntry,
			cacertEntry,
			tlsServerNameEntry,
			passEntry,
			usernameEntry,
			targetIDEntry,
			localEntry,
			remoteEntry,
			runButton,
		))

		w.Show() // Show the new window

		desk.SetSystemTrayMenu(buildMenu(desk, true, a)) // Rebuild menu with 'started' = true
	})

	menu2 := fyne.NewMenuItem("Settings", func() {
		showSupportWindow(a)
	})

	menu3 := fyne.NewMenuItem("show2", func() {
		fmt.Println("menu2")
	})

	return fyne.NewMenu("MyApp", menu1, menu2, menu3)
}

func showSupportWindow(a fyne.App) {
	w := a.NewWindow("Support Settings")

	// Entry widgets for each setting
	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("Enter BOUNDARY_ADDR")
	cacertEntry := widget.NewEntry()
	cacertEntry.SetPlaceHolder("Enter BOUNDARY_CACERT")
	tlsServerNameEntry := widget.NewEntry()
	tlsServerNameEntry.SetPlaceHolder("Enter BOUNDARY_TLS_SERVER_NAME")
	passEntry := widget.NewEntry()
	passEntry.SetPlaceHolder("Enter BOUNDARY_PASS")
	passEntry.Password = true

	// Load existing settings
	loadSettings(addrEntry, cacertEntry, tlsServerNameEntry, passEntry)

	// Save button
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

const settingsFile = "boundary_settings.txt"

func saveSettings(addr, cacert, tlsServerName, pass string) {
	content := fmt.Sprintf("%s\n%s\n%s\n%s", addr, cacert, tlsServerName, pass)
	err := ioutil.WriteFile(settingsFile, []byte(content), 0644)
	if err != nil {
		beeep.Alert("Error", "Failed to save settings: "+err.Error(), "")
	}
}

func loadSettings(addrEntry, cacertEntry, tlsServerNameEntry, passEntry *widget.Entry) {
	content, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return // File not found or other error, just return
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) >= 4 {
		addrEntry.SetText(lines[0])
		cacertEntry.SetText(lines[1])
		tlsServerNameEntry.SetText(lines[2])
		passEntry.SetText(lines[3])
	}
}

func BoundaryTunnel(BOUNDARY_PATH string, USERNAME string, TARGETID string, LOCAL string, REMOTE string) {
	fmt.Println("BOUNDARY_PATH: " + BOUNDARY_PATH)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(BOUNDARY_PATH, "connect", "ssh", "-username", USERNAME, "-target-id", TARGETID, "--", "-L", LOCAL+":localhost:"+REMOTE)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fullOutput := "STDOUT:\n" + stdoutBuf.String() + "\nSTDERR:\n" + stderrBuf.String()
		beeep.Alert("SSH Command Error", fullOutput, "")
		return
	}

	message := stdoutBuf.String()
	beeep.Notify("SSH Command Output", message, "assets/information.png")
}
