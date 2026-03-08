package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupMCPConfig_NewFile(t *testing.T) {
	tmpDir := t.TempDir()

	if err := setupMCPConfig(tmpDir, "/usr/local/bin/multiterm"); err != nil {
		t.Fatalf("setupMCPConfig failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "mcp-config.json"))
	if err != nil {
		t.Fatalf("could not read config: %v", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	mt, ok := config["multiterm"].(map[string]any)
	if !ok {
		t.Fatal("missing 'multiterm' key in config")
	}

	if mt["type"] != "stdio" {
		t.Errorf("type = %v, want stdio", mt["type"])
	}
	if mt["command"] != "/usr/local/bin/multiterm" {
		t.Errorf("command = %v, want /usr/local/bin/multiterm", mt["command"])
	}
}

func TestSetupMCPConfig_PreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mcp-config.json")

	// Write existing config with another server
	existing := map[string]any{
		"other-server": map[string]any{
			"type":    "stdio",
			"command": "other",
		},
	}
	data, _ := json.Marshal(existing)
	os.WriteFile(configPath, data, 0644)

	if err := setupMCPConfig(tmpDir, "multiterm"); err != nil {
		t.Fatalf("setupMCPConfig failed: %v", err)
	}

	data, _ = os.ReadFile(configPath)
	var config map[string]any
	json.Unmarshal(data, &config)

	if _, ok := config["other-server"]; !ok {
		t.Error("existing 'other-server' entry was lost")
	}
	if _, ok := config["multiterm"]; !ok {
		t.Error("multiterm entry was not added")
	}
}

func TestSetupMCPConfig_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Run twice
	setupMCPConfig(tmpDir, "multiterm")
	if err := setupMCPConfig(tmpDir, "multiterm"); err != nil {
		t.Fatalf("second setupMCPConfig failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, "mcp-config.json"))
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("invalid JSON after second run: %v", err)
	}

	// Should still have exactly one top-level multiterm key
	count := 0
	for k := range config {
		if k == "multiterm" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 multiterm key, found %d", count)
	}
}

func TestSetupSkill(t *testing.T) {
	tmpDir := t.TempDir()

	if err := setupSkill(tmpDir); err != nil {
		t.Fatalf("setupSkill failed: %v", err)
	}

	skillPath := filepath.Join(tmpDir, "skills", "multiterm", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("skill file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: multiterm") {
		t.Error("skill missing name frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("skill missing description frontmatter")
	}
	if !strings.Contains(content, "create_pane") {
		t.Error("skill missing create_pane reference")
	}
}

func TestSetupAgent(t *testing.T) {
	tmpDir := t.TempDir()

	if err := setupAgent(tmpDir); err != nil {
		t.Fatalf("setupAgent failed: %v", err)
	}

	agentPath := filepath.Join(tmpDir, "agents", "multiterm.agent.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("agent file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: multiterm") {
		t.Error("agent missing name frontmatter")
	}
	if !strings.Contains(content, "create_pane") {
		t.Error("agent missing create_pane reference")
	}
	if strings.Contains(content, "You are inside a multiterm") {
		t.Error("agent still has old scoped language")
	}
}
