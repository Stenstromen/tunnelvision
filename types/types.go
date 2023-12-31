package types

import "os/exec"

type Host struct {
	Name         string        `yaml:"name"`
	Username     string        `yaml:"username"`
	TargetID     string        `yaml:"target_id"`
	PortForwards []PortForward `yaml:"port_forwards"`
}

type PortForward struct {
	SourcePort      string `yaml:"source_port"`
	DestinationHost string `yaml:"destination_host"`
	DestinationPort string `yaml:"destination_port"`
}

type Settings struct {
	BoundaryBinary         string `yaml:"boundary_binary"`
	BoundaryAddr           string `yaml:"boundary_addr"`
	BoundaryCACert         string `yaml:"boundary_cacert"`
	BoundaryCAPath         string `yaml:"boundary_capath"`
	BoundaryClientCertPath string `yaml:"boundary_client_cert_path"`
	BoundaryClientKeyPath  string `yaml:"boundary_client_key_path"`
	BoundaryTLSInsecure    bool   `yaml:"boundary_tls_insecure"`
	BoundaryTLSServerName  string `yaml:"boundary_tls_server_name"`
}

type TunnelConfig struct {
	BoundaryPath      string
	Username          string
	TargetID          string
	PortForwards      []string
	HostName          string
	ActiveTunnels     map[string]bool
	BoundaryProcesses map[string]*exec.Cmd
}

type Level string
