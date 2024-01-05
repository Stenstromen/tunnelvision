package gui

import (
	"fmt"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/stenstromen/tunnelvision/boundary"
	"github.com/stenstromen/tunnelvision/config"
	"github.com/stenstromen/tunnelvision/types"
	"github.com/stenstromen/tunnelvision/util"
)

var trayMenu *fyne.Menu
var hostsList *widget.List

var boundaryProcesses map[string]*exec.Cmd
var activeTunnels map[string]bool

func init() {
	boundaryProcesses = make(map[string]*exec.Cmd)
	activeTunnels = make(map[string]bool)
}

func UpdateTrayMenu(app fyne.App, hosts []types.Host) {
	dynamicMenuItems := make([]*fyne.MenuItem, 0)

	for _, host := range hosts {
		hostCopy := host
		menuItemTitle := hostCopy.Name
		if active, exists := activeTunnels[hostCopy.Name]; exists && active {
			menuItemTitle += " ✓"
		}

		menuItem := fyne.NewMenuItem(menuItemTitle, func() {
			fmt.Println("Selected host:", hostCopy.Name)
			started, err := boundary.Tunnel(&types.TunnelConfig{
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
				if started {
					util.Notify("Success", "Boundary tunnel started successfully.", "info")
				} else {
					util.Notify("Success", "Boundary tunnel stopped successfully.", "info")
				}
				UpdateTrayMenu(app, hosts)
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
		fyne.NewMenuItemSeparator(),
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

	settings := config.LoadSettings()
	if settings != nil {
		addrEntry.SetText(settings.BoundaryAddr)
		cacertEntry.SetText(settings.BoundaryCACert)
		tlsServerNameEntry.SetText(settings.BoundaryTLSServerName)
		passEntry.SetText(settings.BoundaryPass)
	}

	saveButton := widget.NewButton("Save Settings", func() {
		config.SaveSettings(addrEntry.Text, cacertEntry.Text, tlsServerNameEntry.Text, passEntry.Text)
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

func updateHostsList() {
	hostsFile, err := config.GetHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}

	hosts, err := config.LoadHostsFromFile(hostsFile)
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
	hostsFile, err := config.GetHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}

	hosts, _ := config.LoadHostsFromFile(hostsFile)
	hosts = append(hosts, newHost)
	_ = config.SaveHostsToFile(hosts, hostsFile)

	UpdateTrayMenu(app, hosts)
}

func extractPortForwards(host types.Host) []string {
	portForwards := make([]string, 0)
	for _, pf := range host.PortForwards {
		portForwards = append(portForwards, pf.SourcePort+":"+pf.DestinationHost+":"+pf.DestinationPort)
	}
	return portForwards
}
