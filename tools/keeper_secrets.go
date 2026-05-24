package tools

import (
	"os/exec"
	"strings"
)

func GetKeeperSecret(path string) (string, error) {
	// Assumes keeper-secrets-manager is installed and configured
	cmd := exec.Command("ksm", "secret", "get", "--path", path, "--field", "value")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
