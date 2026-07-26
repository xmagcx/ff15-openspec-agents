package ff15

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const managedBlockName = "FF15:MANAGED"

func ExecutePlan(plan Plan, stdout io.Writer) error {
	for index, step := range plan.Steps {
		fmt.Fprintf(stdout, "[%d/%d] %s\n", index+1, len(plan.Steps), step.Title)
		switch step.Kind {
		case StepInstall:
			if err := executeInstallStep(plan.Platform, step); err != nil {
				fmt.Fprintf(stdout, "  FAIL: %v\n", err)
				return err
			}
			fmt.Fprintln(stdout, "  OK")
		case StepPatch:
			if err := applyManagedPatch(step.FilePath, step.ManagedText); err != nil {
				fmt.Fprintf(stdout, "  FAIL: %v\n", err)
				return err
			}
			fmt.Fprintln(stdout, "  OK")
		}
	}
	return nil
}

func applyManagedPatch(path string, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	block := managedBlock(body)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(defaultDocument(path, block)), 0o644)
		}
		return err
	}
	updated := upsertManagedBlock(string(content), block)
	return os.WriteFile(path, []byte(updated), 0o644)
}

func managedBlock(body string) string {
	return fmt.Sprintf("<!-- %s:START -->\n%s\n<!-- %s:END -->", managedBlockName, strings.TrimSpace(body), managedBlockName)
}

func upsertManagedBlock(content, block string) string {
	pattern := regexp.MustCompile(`(?s)<!-- ` + regexp.QuoteMeta(managedBlockName) + `:START -->.*?<!-- ` + regexp.QuoteMeta(managedBlockName) + `:END -->`)
	if pattern.MatchString(content) {
		return pattern.ReplaceAllString(content, block)
	}
	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return block + "\n"
	}
	return trimmed + "\n\n" + block + "\n"
}

func defaultDocument(path, block string) string {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "" || name == "." {
		name = "GUIDE"
	}
	return fmt.Sprintf("# %s\n\n%s\n", strings.ToUpper(name), block)
}
