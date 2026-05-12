package gen

import (
	"fmt"
	"os"
	"os/exec"
)

func BuildSQLC(folder string) error {
	if folder == "" {
		panic("Folder is required")
	}

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		panic("Folder does not exist: " + folder)
	}

	if fileInfo, err := os.Stat(folder); err != nil || !fileInfo.IsDir() {
		panic("Invalid folder: " + folder)
	}

	// execute `sqlc generate` in the folder
	cmd := exec.Command("sqlc", "generate")
	cmd.Dir = folder
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute sqlc generate: %w, output: %s", err, string(output))
	}

	return nil
}
