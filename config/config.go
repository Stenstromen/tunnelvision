package config

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/stenstromen/tunnelvision/types"
	"github.com/stenstromen/tunnelvision/util"
	"gopkg.in/yaml.v2"
)

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

func SaveSettings(addr, cacert, tlsServerName, pass string) {
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

func LoadSettings() *types.Settings {
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

func GetHostsFilePath() (string, error) {
	appSupportDir, err := getAppSupportDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appSupportDir, "hosts.yaml"), nil
}

func updateHostsList(hostsList *widget.List) {
	hostsFile, err := GetHostsFilePath()
	if err != nil {
		fmt.Println("Error getting hosts file path:", err)
		return
	}

	hosts, err := LoadHostsFromFile(hostsFile)
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

func SaveHostsToFile(hosts []types.Host, filename string) error {
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

func LoadHostsFromFile(filename string) ([]types.Host, error) {
	var hosts []types.Host
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			err = SaveHostsToFile(hosts, filename)
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