package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielgtaylor/openapi-cli-generator/cli"
	"github.com/spf13/viper"
)

func TestAgentFormatterFieldsLimitAndSaveFull(t *testing.T) {
	resetAgentTestState(t)

	var out bytes.Buffer
	cli.Stdout = &out
	viper.Set("output", "agent")
	viper.Set("fields", "uuid,username,status")
	viper.Set("limit", 1)

	fullPath := filepath.Join(t.TempDir(), "full.json")
	viper.Set("save-full", fullPath)

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"uuid": "u1", "username": "alice", "status": "ACTIVE", "secret": "hidden"},
			map[string]interface{}{"uuid": "u2", "username": "bob", "status": "DISABLED", "secret": "hidden"},
		},
	}

	if err := (&agentFormatter{}).Format(data); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	for _, want := range []string{
		"ok users total=2 shown=1 truncated=true full=" + fullPath,
		"uuid\tusername\tstatus",
		"u1\talice\tACTIVE",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("agent output missing %q in:\n%s", want, got)
		}
	}

	full, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(full), `"secret": "hidden"`) {
		t.Fatalf("full JSON was not written as expected:\n%s", string(full))
	}
}

func TestAgentFormatterNDJSONProjectsFields(t *testing.T) {
	resetAgentTestState(t)

	var out bytes.Buffer
	cli.Stdout = &out
	viper.Set("fields", "uuid,username")

	data := []interface{}{
		map[string]interface{}{"uuid": "u1", "username": "alice", "status": "ACTIVE"},
		map[string]interface{}{"uuid": "u2", "username": "bob", "status": "DISABLED"},
	}

	if err := formatNDJSON(data); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(out.String())
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 ndjson lines, got %d:\n%s", len(lines), got)
	}
	if !strings.Contains(lines[0], `"uuid":"u1"`) || !strings.Contains(lines[0], `"username":"alice"`) {
		t.Fatalf("first ndjson line not projected as expected: %s", lines[0])
	}
	if strings.Contains(lines[0], "status") {
		t.Fatalf("unexpected unselected field in ndjson: %s", lines[0])
	}
}

func resetAgentTestState(t *testing.T) {
	t.Helper()

	oldStdout := cli.Stdout
	t.Cleanup(func() {
		cli.Stdout = oldStdout
		viper.Reset()
	})

	viper.Reset()
	viper.Set("fields", "")
	viper.Set("limit", 50)
	viper.Set("save-full", "")
	viper.Set("query", "")
}
