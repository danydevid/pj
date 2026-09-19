package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRootCmd_FlagsAndExecution(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantList     bool
		wantWhich    bool
		wantErr      bool
		expectOutput string
	}{
		{
			name:      "Default execution without flags",
			args:      []string{},
			wantList:  false,
			wantWhich: false,
			wantErr:   false,
		},
		{
			name:      "Set list flag short (-l)",
			args:      []string{"-l"},
			wantList:  true,
			wantWhich: false,
			wantErr:   false,
		},
		{
			name:      "Set which flag long (--which)",
			args:      []string{"--which"},
			wantList:  false,
			wantWhich: true,
			wantErr:   false,
		},
		{
			name:    "Invalid flag",
			args:    []string{"--invalid-flag"},
			wantErr: true,
		},
		{
			name:         "Version flag",
			args:         []string{"--version"},
			wantErr:      false,
			expectOutput: "0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listFlag = false
			whichFlag = false

			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if listFlag != tt.wantList {
				t.Errorf("listFlag = %v, want %v", listFlag, tt.wantList)
			}

			if whichFlag != tt.wantWhich {
				t.Errorf("whichFlag = %v, want %v", whichFlag, tt.wantWhich)
			}

			if tt.expectOutput != "" && !bytes.Contains(buf.Bytes(), []byte(tt.expectOutput)) {
				t.Errorf("expected output to contain %q, got %q", tt.expectOutput, buf.String())
			}
		})
	}
}

// ==========================================
// 1. TESTS FOR RESOLVEPLUGIN FUNCTION
// ==========================================

func TestResolvePlugin_Success(t *testing.T) {
	tmpDir := t.TempDir()

	binName := "pj-testcmd"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	mockBinPath := filepath.Join(tmpDir, binName)

	// Create executable mock binary
	//nolint:gosec // 0755 permission is required for the executable bit test in isExecutable()
	err := os.WriteFile(mockBinPath, []byte("#!/bin/sh\necho test"), 0755)
	if err != nil {
		t.Fatalf("failed to create mock plugin file: %v", err)
	}

	// Set PJ_PLUGIN_DIR so ResolvePlugin resolves the target mock file first
	t.Setenv("PJ_PLUGIN_DIR", tmpDir)

	got, err := ResolvePlugin("testcmd")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if got != mockBinPath {
		t.Errorf("expected path %s, got %s", mockBinPath, got)
	}
}

func TestResolvePlugin_NotFound(t *testing.T) {
	// Point PJ_PLUGIN_DIR to an empty temp directory
	t.Setenv("PJ_PLUGIN_DIR", t.TempDir())
	// Clear PATH variable to ensure exec.LookPath does not find system-wide binaries
	t.Setenv("PATH", "")

	got, err := ResolvePlugin("non-existent-plugin-xyz")

	if err == nil {
		t.Fatal("expected error when plugin is missing, got nil")
	}

	if got != "" {
		t.Errorf("expected empty path string, got %s", got)
	}

	expectedErrSubstr := "not found"
	if !strings.Contains(err.Error(), expectedErrSubstr) {
		t.Errorf("expected error message to contain %q, got: %v", expectedErrSubstr, err)
	}
}

// ==========================================
// 2. TESTS FOR DISPATCH FUNCTION
// ==========================================

func TestDispatch_PluginNotFound(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Mock os.Args with subcommand and parameters matching expected CLI input
	os.Args = []string{"pj", "jump.sh", "arg1"}

	// Set PJ_PLUGIN_DIR to an empty temp directory
	t.Setenv("PJ_PLUGIN_DIR", t.TempDir())

	// Clear PATH to force ResolvePlugin("jump.sh") to fail completely
	t.Setenv("PATH", "")

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	err := Dispatch()
	if err == nil {
		t.Error("expected error due to missing plugin, got", err)
	}
}

func TestDispatch_InvalidExecution(t *testing.T) {
	tmpDir := t.TempDir()

	// Dispatch currently resolves "jump.sh" hardcoded -> "pj-jump.sh"
	binName := "pj-jump.sh"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	mockBinPath := filepath.Join(tmpDir, binName)

	// Create non-executable content or invalid script binary
	//nolint:gosec // 0755 permission is required for the executable bit test in isExecutable()
	_ = os.WriteFile(mockBinPath, []byte("invalid binary content"), 0755)

	t.Setenv("PJ_PLUGIN_DIR", tmpDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"pj", "jump.sh", "dummy-arg"}

	err := Dispatch()
	if err == nil {
		t.Error("expected execution error, got nil")
	}
}
