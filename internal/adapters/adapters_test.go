package adapters_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/b2jant/bricktap/internal/adapters"
	"github.com/b2jant/bricktap/internal/core"
	"github.com/b2jant/bricktap/internal/dialects"
	"github.com/b2jant/bricktap/internal/parser"
	"github.com/b2jant/bricktap/internal/scanner"
)

var update = flag.Bool("update", false, "update golden files")

func TestAdapters(t *testing.T) {
	testName := "sales_orders"
	testDir := filepath.Join("testdata", testName)
	inputPath := filepath.Join(testDir, "input.yaml")

	// Parse the model directly from file
	model, err := parser.ParseModel(inputPath, core.GlobalRules{})
	if err != nil {
		t.Fatalf("Failed to parse input yaml: %v", err)
	}

	dialect := dialects.NewSnowflakeDialect()

	tests := []struct {
		name       string
		adapter    adapters.Generator
		goldenFile string
	}{
		{"DBT Output", adapters.NewDbtAdapter(), "expected_dbt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Generate the output to a temp directory to capture what the adapter creates
			tempDir := t.TempDir()
			fileInfo := scanner.File{
				RelativeDir: "models",
				BaseName:    "sales_orders",
			}

			err := tc.adapter.Generate(*model, fileInfo, dialect, tempDir)
			if err != nil {
				t.Fatalf("Adapter Generate failed: %v", err)
			}

			outDir := filepath.Join(tempDir, "models")

			// Bundle all generated files into a single string to compare against a golden file
			var actualOutput bytes.Buffer

			files, err := os.ReadDir(outDir)
			if err != nil {
				t.Fatalf("Failed to read output dir: %v", err)
			}

			for _, f := range files {
				content, err := os.ReadFile(filepath.Join(outDir, f.Name()))
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}
				actualOutput.WriteString("--- " + f.Name() + " ---\n")
				actualOutput.Write(content)
				actualOutput.WriteString("\n")
			}

			goldenPath := filepath.Join(testDir, tc.goldenFile+".txt")

			if *update {
				err := os.WriteFile(goldenPath, actualOutput.Bytes(), 0644)
				if err != nil {
					t.Fatalf("Failed to update golden file: %v", err)
				}
				return
			}

			expectedOutput, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			if !bytes.Equal(actualOutput.Bytes(), expectedOutput) {
				t.Errorf("Mismatch in generated output.\n=== ACTUAL ===\n%s\n=== EXPECTED ===\n%s\nRun with -update to overwrite.", actualOutput.String(), expectedOutput)
			}
		})
	}
}
