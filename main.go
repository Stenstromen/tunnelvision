package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/stenstromen/tunnelvision/types"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/beeep"
	"gopkg.in/yaml.v2"
)

var trayMenu *fyne.Menu
var mainWindow fyne.Window
var hostsList *widget.List

var activeTunnels map[string]bool // Keyed by host name

func init() {
	activeTunnels = make(map[string]bool)
}

func main() {
	app := app.New()

	mainWindow = app.NewWindow("Tunnelvision")
	hosts, _ := loadHostsFromFile(settingsFile2)
	updateTrayMenu(app, hosts)

	app.Run()
}

func updateHostsList() {
	hosts, err := loadHostsFromFile(settingsFile2)
	if err != nil {
		fmt.Println("Error loading hosts:", err)
		return
	}

	hostsList.Length = func() int {
		return len(hosts)
	}
	hostsList.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
		item.(*widget.Label).SetText(hosts[id].Name)
	}

	hostsList.Refresh()
}

func onHostAdded(app fyne.App, newHost types.Host) {
	hosts, _ := loadHostsFromFile(settingsFile2)
	hosts = append(hosts, newHost)
	_ = saveHostsToFile(hosts, settingsFile2)

	updateTrayMenu(app, hosts)
}

func updateTrayMenu(app fyne.App, hosts []types.Host) {
	dynamicMenuItems := make([]*fyne.MenuItem, 0)

	for _, host := range hosts {
		hostCopy := host
		tunnelStatus := ""
		if activeTunnels[hostCopy.Name] {
			tunnelStatus = " ✓" // Unicode checkmark
		}
		menuItemTitle := hostCopy.Name + tunnelStatus
		menuItem := fyne.NewMenuItem(menuItemTitle, func() {
			fmt.Println("Selected host:", hostCopy.Name)
			portForwards := make([]string, 0)
			for _, portForward := range hostCopy.PortForwards {
				portForwards = append(portForwards, portForward.SourcePort+":"+portForward.DestinationHost+":"+portForward.DestinationPort)
			}
			BoundaryTunnel("/opt/homebrew/bin/boundary", hostCopy.Username, hostCopy.TargetID, portForwards, hostCopy.Name)
			beeep.Notify("Boundary Tunnel", "The Boundary tunnel has been established successfully.", "")
		})
		dynamicMenuItems = append(dynamicMenuItems, menuItem)
	}

	dynamicMenuItems = append(dynamicMenuItems, fyne.NewMenuItemSeparator())

	standardMenuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Hosts", func() {
			showHostsWindow(app)
		}),
		fyne.NewMenuItem("Settings", func() {
			showSupportWindow(app)
		}),
		fyne.NewMenuItem("Quit", func() {
			app.Quit()
		}),
	}

	allMenuItems := append(dynamicMenuItems, standardMenuItems...)

	trayMenu = fyne.NewMenu("", allMenuItems...)
	if desk, ok := app.(desktop.App); ok {
		desk.SetSystemTrayMenu(trayMenu)
	}
}

func saveHostsToFile(hosts []types.Host, filename string) error {
	data, err := yaml.Marshal(hosts)
	if err != nil {
		fmt.Printf("Error marshalling YAML: %v\n", err)
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
	}
	return err
}

func loadHostsFromFile(filename string) ([]types.Host, error) {
	var hosts []types.Host
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			err = saveHostsToFile(hosts, filename)
			if err != nil {
				return nil, err
			}
			return hosts, nil
		}
		return nil, err
	}
	err = yaml.Unmarshal(data, &hosts)
	return hosts, err
}

func showHostsWindow(a fyne.App) {
	w := a.NewWindow("Hosts")
	w.Resize(fyne.NewSize(600, 600))

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter Name")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Enter Username")

	targetIDEntry := widget.NewEntry()
	targetIDEntry.SetPlaceHolder("Enter TargetID")

	portForwardContainer := container.NewVBox()
	addPortForward := func() *fyne.Container {
		sourcePortEntry := widget.NewEntry()
		sourcePortEntry.SetPlaceHolder("Source Port")
		destinationHostEntry := widget.NewEntry()
		destinationHostEntry.SetPlaceHolder("Destination Host")
		destinationPortEntry := widget.NewEntry()
		destinationPortEntry.SetPlaceHolder("Destination Port")

		return container.NewHBox(sourcePortEntry, destinationHostEntry, destinationPortEntry)
	}

	portForwardContainer.Add(addPortForward())

	addPortForwardButton := widget.NewButton("Add Another Port Forward", func() {
		portForwardContainer.Add(addPortForward())
		portForwardContainer.Refresh()
	})

	collectPortForwards := func() []types.PortForward {
		var portForwards []types.PortForward
		for _, obj := range portForwardContainer.Objects {
			box := obj.(*fyne.Container)
			portForwards = append(portForwards, types.PortForward{
				SourcePort:      box.Objects[0].(*widget.Entry).Text,
				DestinationHost: box.Objects[1].(*widget.Entry).Text,
				DestinationPort: box.Objects[2].(*widget.Entry).Text,
			})
		}
		return portForwards
	}

	if hostsList == nil {
		hostsList = widget.NewList(
			func() int { return 0 }, // This will be updated in updateHostsList
			func() fyne.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, obj fyne.CanvasObject) {},
		)
	}

	updateHostsList()

	saveButton := widget.NewButton("Save Host", func() {
		portForwards := collectPortForwards()
		newHost := types.Host{
			Name:         nameEntry.Text,
			Username:     usernameEntry.Text,
			TargetID:     targetIDEntry.Text,
			PortForwards: portForwards,
		}

		beeep.Notify("Host Saved", "The new host has been saved successfully.", "")
		updateHostsList()
		onHostAdded(a, newHost)
	})

	w.SetContent(container.NewVBox(
		nameEntry,
		usernameEntry,
		targetIDEntry,
		portForwardContainer,
		addPortForwardButton,
		saveButton,
		hostsList,
	))

	w.Show()
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

	settings := loadSettings()
	if settings != nil {
		addrEntry.SetText(settings.BoundaryAddr)
		cacertEntry.SetText(settings.BoundaryCACert)
		tlsServerNameEntry.SetText(settings.BoundaryTLSServerName)
		passEntry.SetText(settings.BoundaryPass)
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

const settingsFile = "/Users/filip/Documents/tunnelvision/settings.yaml"
const settingsFile2 = "/Users/filip/Documents/tunnelvision/hosts.yaml"

func saveSettings(addr, cacert, tlsServerName, pass string) {
	settings := types.Settings{
		BoundaryAddr:          addr,
		BoundaryCACert:        cacert,
		BoundaryTLSServerName: tlsServerName,
		BoundaryPass:          pass,
	}

	data, err := yaml.Marshal(&settings)
	if err != nil {
		beeep.Alert("Error", "Failed to marshal settings: "+err.Error(), "")
		return
	}

	err = os.WriteFile(settingsFile, data, 0644)
	if err != nil {
		beeep.Alert("Error", "Failed to save settings: "+err.Error(), "")
	}
}

func loadSettings() *types.Settings {
	var settings types.Settings

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		fmt.Println("Error reading settings file:", err)
		return nil
	}

	err = yaml.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println("Error unmarshalling settings:", err)
		return nil
	}

	return &settings
}

func BoundaryTunnel(BOUNDARY_PATH string, USERNAME string, TARGETID string, portForwards []string, hostName string) {
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
		activeTunnels[hostName] = false
		return
	} else {
		activeTunnels[hostName] = true
	}

	updateTrayMenu(app.New(), nil) // FIX ME

	message := stdoutBuf.String()
	fmt.Println(message)
	beeep.Notify("Boundary Command Output", message, "assets/information.png")
}
