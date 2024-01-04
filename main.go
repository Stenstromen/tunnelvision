package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stenstromen/tunnelvision/boundary"
	"github.com/stenstromen/tunnelvision/types"
	"github.com/stenstromen/tunnelvision/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v2"
)

var trayMenu *fyne.Menu
var hostsList *widget.List

var boundaryProcesses map[string]*exec.Cmd
var activeTunnels map[string]bool

func init() {
	boundaryProcesses = make(map[string]*exec.Cmd)
	activeTunnels = make(map[string]bool)
}

func main() {
	myApp := app.New()

	hostsFile, err := getHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}
	hosts, _ := loadHostsFromFile(hostsFile)
	updateTrayMenu(myApp, hosts)

	myApp.Run()
}

func getAppSupportDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appSupportDir := filepath.Join(homeDir, "Library", "Application Support", "Tunnelvision")
	if err := os.MkdirAll(appSupportDir, 0700); err != nil {
		return "", err
	}

	return appSupportDir, nil
}

func getSettingsFilePath() (string, error) {
	appSupportDir, err := getAppSupportDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appSupportDir, "settings.yaml"), nil
}

func getHostsFilePath() (string, error) {
	appSupportDir, err := getAppSupportDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appSupportDir, "hosts.yaml"), nil
}

func updateHostsList() {
	hostsFile, err := getHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}

	hosts, err := loadHostsFromFile(hostsFile)
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
	hostsFile, err := getHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}

	hosts, _ := loadHostsFromFile(hostsFile)
	hosts = append(hosts, newHost)
	_ = saveHostsToFile(hosts, hostsFile)

	updateTrayMenu(app, hosts)
}

func updateTrayMenu(app fyne.App, hosts []types.Host) {
	dynamicMenuItems := make([]*fyne.MenuItem, 0)

	for _, host := range hosts {
		hostCopy := host
		menuItemTitle := hostCopy.Name
		if active, exists := activeTunnels[hostCopy.Name]; exists && active {
			menuItemTitle += " ✓"
		}

		menuItem := fyne.NewMenuItem(menuItemTitle, func() {
			fmt.Println("Selected host:", hostCopy.Name)
			err := boundary.Tunnel(&types.TunnelConfig{
				BoundaryPath:      "/opt/homebrew/bin/boundary",
				Username:          hostCopy.Username,
				TargetID:          hostCopy.TargetID,
				PortForwards:      extractPortForwards(hostCopy),
				HostName:          hostCopy.Name,
				ActiveTunnels:     activeTunnels,
				BoundaryProcesses: boundaryProcesses,
			})
			if err != nil {
				fmt.Println("Error running Boundary tunnel:", err)
				util.Notify("Error", "Failed to run Boundary tunnel: "+err.Error(), "error")
			} else {
				updateTrayMenu(app, hosts)
			}
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

func extractPortForwards(host types.Host) []string {
	portForwards := make([]string, 0)
	for _, pf := range host.PortForwards {
		portForwards = append(portForwards, pf.SourcePort+":"+pf.DestinationHost+":"+pf.DestinationPort)
	}
	return portForwards
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
			func() int { return 0 },
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

		util.Notify("Host Saved", "The new host has been saved successfully.", "info")
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
		util.Notify("Settings Saved", "Your settings have been saved successfully.", "info")
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

func saveSettings(addr, cacert, tlsServerName, pass string) {
	settings := types.Settings{
		BoundaryAddr:          addr,
		BoundaryCACert:        cacert,
		BoundaryTLSServerName: tlsServerName,
		BoundaryPass:          pass,
	}

	settingsFile, err := getSettingsFilePath()
	if err != nil {
		util.Notify("Error", "Failed to get settings file path: "+err.Error(), "error")
		return
	}

	data, err := yaml.Marshal(&settings)
	if err != nil {
		util.Notify("Error", "Failed to marshal settings: "+err.Error(), "error")
		return
	}

	err = os.WriteFile(settingsFile, data, 0644)
	if err != nil {
		util.Notify("Error", "Failed to save settings: "+err.Error(), "error")
	}
}

func loadSettings() *types.Settings {
	var settings types.Settings

	settingsFile, err := getSettingsFilePath()
	if err != nil {
		fmt.Println("Error getting settings file path:", err)
		return nil
	}

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
