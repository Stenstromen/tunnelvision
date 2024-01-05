package boundary

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/stenstromen/tunnelvision/config"
	"github.com/stenstromen/tunnelvision/types"
	"github.com/stenstromen/tunnelvision/util"
)

func Tunnel(cfg *types.TunnelConfig) (bool, error) {
	BOUNDARY_PATH, USERNAME, TARGETID := cfg.BoundaryPath, cfg.Username, cfg.TargetID
	portForwards, hostName := cfg.PortForwards, cfg.HostName
	activeTunnels, boundaryProcesses := cfg.ActiveTunnels, cfg.BoundaryProcesses

	settings := config.LoadSettings()

	env := os.Environ()
	env = append(env, "BOUNDARY_ADDR="+settings.BoundaryAddr)
	if settings.BoundaryCACert != "" {
		env = append(env, "BOUNDARY_CACERT="+settings.BoundaryCACert)
	}
	if settings.BoundaryCAPath != "" {
		env = append(env, "BOUNDARY_CAPATH="+settings.BoundaryCAPath)
	}
	if settings.BoundaryClientCertPath != "" {
		env = append(env, "BOUNDARY_CLIENT_CERT_PATH="+settings.BoundaryClientCertPath)
	}
	if settings.BoundaryClientKeyPath != "" {
		env = append(env, "BOUNDARY_CLIENT_KEY_PATH="+settings.BoundaryClientKeyPath)
	}
	//env = append(env, "BOUNDARY_TLS_INSECURE="+fmt.Sprintf("%t", settings.BoundaryTLSInsecure))
	if settings.BoundaryTLSServerName != "" {
		env = append(env, "BOUNDARY_TLS_SERVER_NAME="+settings.BoundaryTLSServerName)
	}

	if cmd, ok := boundaryProcesses[hostName]; ok {
		if err := cmd.Process.Kill(); err != nil {
			return false, fmt.Errorf("failed to kill boundary process: %v", err)
		}
		delete(boundaryProcesses, hostName)
		activeTunnels[hostName] = false
		return false, nil
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	args := []string{"connect", "ssh", "-username", USERNAME, "-target-id", TARGETID, "--"}
	for _, portForward := range portForwards {
		args = append(args, "-L", portForward)
	}

	cmd := exec.Command(BOUNDARY_PATH, args...)
	cmd.Env = env
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("failed to start boundary command: %v", err)
	}

	boundaryProcesses[hostName] = cmd
	activeTunnels[hostName] = true

	go func() {
		err := cmd.Wait()
		delete(boundaryProcesses, hostName)
		activeTunnels[hostName] = false

		if err != nil {
			fmt.Println("Boundary command finished with error:", err)
			fmt.Println("Standard Output:", stdoutBuf.String())
			fmt.Println("Standard Error:", stderrBuf.String())
			util.Notify("Boundary Error", stderrBuf.String(), util.ErrorLevel)
		} else {
			fmt.Println("Boundary command finished successfully")
			fmt.Println("Standard Output:", stdoutBuf.String())
			util.Notify("Boundary Success", stdoutBuf.String(), util.InfoLevel)
		}
	}()

	return true, nil
}
