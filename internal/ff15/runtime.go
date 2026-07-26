package ff15

import (
	"fmt"
	"os/exec"
)

func executeInstallStep(platform Platform, step Step) error {
	for _, command := range step.Commands {
		if err := runShell(platform, command); err != nil {
			return fmt.Errorf("install command failed: %s: %w", command, err)
		}
	}
	if step.Verify != "" {
		if err := runShell(platform, step.Verify); err != nil {
			if step.ManualHint != "" {
				return fmt.Errorf("verification failed for %q: %w; %s", step.Verify, err, step.ManualHint)
			}
			return fmt.Errorf("verification failed for %q: %w", step.Verify, err)
		}
	}
	return nil
}

func runShell(platform Platform, command string) error {
	var cmd *exec.Cmd
	if platform == PlatformWindows {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", command)
	} else {
		cmd = exec.Command("bash", "-lc", command)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w (%s)", err, string(output))
	}
	return nil
}
