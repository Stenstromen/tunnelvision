package boundary

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/stenstromen/tunnelvision/types"
)

func Tunnel(cfg *types.TunnelConfig) (bool, error) {
	BOUNDARY_PATH, USERNAME, TARGETID := cfg.BoundaryPath, cfg.Username, cfg.TargetID
	portForwards, hostName := cfg.PortForwards, cfg.HostName
	activeTunnels, boundaryProcesses := cfg.ActiveTunnels, cfg.BoundaryProcesses

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
		} else {
			fmt.Println("Boundary command finished successfully")
		}
	}()

	return true, nil
}
