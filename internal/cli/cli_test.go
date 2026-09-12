package cli

import (
	"bytes"
	"io"
	"os"
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
			// Reset flag state before running each test case
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
// 1. TESTS FOR RESOLVE FUNCTION
// ==========================================

func TestResolve_Success(t *testing.T) {
	// Set temporary environment variable for testing
	envKey := "TEST_PJ_HOME"
	expectedVal := "/custom/path"
	t.Setenv(envKey, expectedVal)

	got, err := Resolve(envKey)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if got != expectedVal {
		t.Errorf("expected %s, got %s", expectedVal, got)
	}
}

func TestResolve_NotFound(t *testing.T) {
	// Capture stdout to prevent log prints from cluttering test output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	got, err := Resolve("NON_EXISTENT_ENV_VAR")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err == nil {
		t.Fatal("expected error when environment variable is missing, got nil")
	}

	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}

	if !strings.Contains(buf.String(), "cannot find") {
		t.Errorf("expected stdout to print error message, got: %s", buf.String())
	}
}

// ==========================================
// 2. TESTS FOR DISPATCH FUNCTION
// ==========================================

func TestDispatch_MissingEnv(t *testing.T) {
	// Mock os.Args state
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"pj", "my-plugin"}

	// Dispatch() currently calls Resolve(""), which should fail
	err := Dispatch()
	if err == nil {
		t.Error("expected error due to Resolve(\"\") failure, got nil")
	}
}

func TestDispatch_InvalidCommand(t *testing.T) {
	// Prepare mock environment if the env name in Resolve is fixed
	t.Setenv("PJ_PLUGIN_DIR", "/tmp") // Mocking empty string env to bypass temporarily

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"pj", "invalid-binary-command-xyz"}

	err := Dispatch()
	if err == nil {
		t.Error("expected error due to missing binary, got nil")
	}

	if !strings.Contains(err.Error(), "failed to start process") {
		t.Errorf("expected error 'failed to start process', got: %v", err)
	}
}
