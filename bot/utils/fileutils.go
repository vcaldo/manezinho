package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vcaldo/manezinho/jonatas/redisutils"
)

func CompressAndSplitDownload(ctx context.Context, source, destination string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return err
	}

	// Prepare 7za command with volume size parameter
	cmd := exec.Command("7zz",
		"a",                               // add to archive
		"-v2000m",                         // split into 2gb volumes
		"-t7z",                            // use 7z format
		fmt.Sprintf("%s.7z", destination), // output file
		source,                            // input file/directory
	)

	// Capture command output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compression failed: %v\nOutput: %s", err, output)
	}

	log.Printf("Compression completed: %s\n", output)
	return nil
}

func RemoveUploadedFiles(ctx context.Context, download redisutils.Download) error {
	if err := os.RemoveAll(download.UploadPath); err != nil {
		return fmt.Errorf("failed to remove uploaded files: %v", err)
	}
	return nil
}
