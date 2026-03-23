package baseline

import (
	"os/exec"
	"runtime"
	"strings"
)

// Environment captures the system context for reproducibility.
type Environment struct {
	GoVersion     string `json:"go_version"`
	GOOS          string `json:"goos"`
	GOARCH        string `json:"goarch"`
	GOMAXPROCS    int    `json:"gomaxprocs"`
	NumCPU        int    `json:"num_cpu"`
	DockerVersion string `json:"docker_version,omitempty"`
}

// CaptureEnvironment collects runtime and system information.
func CaptureEnvironment() Environment {
	env := Environment{
		GoVersion:  runtime.Version(),
		GOOS:       runtime.GOOS,
		GOARCH:     runtime.GOARCH,
		GOMAXPROCS: runtime.GOMAXPROCS(0),
		NumCPU:     runtime.NumCPU(),
	}

	out, err := exec.Command("docker", "--version").Output()
	if err == nil {
		env.DockerVersion = strings.TrimSpace(string(out))
	}

	return env
}
