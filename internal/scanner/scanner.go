package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// File represents a discovered YAML model file with its relative location
type File struct {
	AbsolutePath string
	RelativeDir  string // e.g. "ivet" if the file is in "semantic_models/ivet/some.yaml"
	BaseName     string // e.g. "some" (without extension)
}

// Scan finds all files with the given extension in the root directory,
// maintaining their relative directory structure for mirroring.
func Scan(rootDir string, extension string) ([]File, error) {
	var files []File

	// Ensure the root directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return files, err
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), extension) {
			// Calculate the relative path to maintain the mirror structure
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}

			// Get just the directory portion of the relative path
			relDir := filepath.Dir(relPath)
			if relDir == "." {
				relDir = "" // Root level files don't need a subdirectory
			}

			baseName := strings.TrimSuffix(info.Name(), extension)

			files = append(files, File{
				AbsolutePath: path,
				RelativeDir:  relDir,
				BaseName:     baseName,
			})
		}
		return nil
	})

	return files, err
}
