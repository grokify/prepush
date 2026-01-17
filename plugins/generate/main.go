// Command generate creates Claude and Gemini plugins from canonical JSON specs.
//
// Usage:
//
//	go run ./plugins/generate
//
// This tool reads the canonical plugin specification from plugins/spec/ and
// generates tool-specific plugins in plugins/claude/ and plugins/gemini/.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/agentplexus/aiassistkit/agents"
	"github.com/agentplexus/aiassistkit/commands"
	"github.com/agentplexus/aiassistkit/plugins"
	"github.com/agentplexus/aiassistkit/skills"
)

func main() {
	// Determine paths
	baseDir := findBaseDir()
	specDir := filepath.Join(baseDir, "plugins", "spec")
	claudeDir := filepath.Join(baseDir, "plugins", "claude")
	geminiDir := filepath.Join(baseDir, "plugins", "gemini")

	fmt.Println("=== Release Agent Plugin Generator ===")
	fmt.Printf("Spec directory: %s\n", specDir)
	fmt.Printf("Claude output: %s\n", claudeDir)
	fmt.Printf("Gemini output: %s\n", geminiDir)
	fmt.Println()

	// Load canonical specs
	plugin, err := loadPlugin(filepath.Join(specDir, "plugin.json"))
	if err != nil {
		log.Fatalf("Failed to load plugin spec: %v", err)
	}

	cmds, err := loadCommands(filepath.Join(specDir, "commands"))
	if err != nil {
		log.Fatalf("Failed to load commands: %v", err)
	}

	skls, err := loadSkills(filepath.Join(specDir, "skills"))
	if err != nil {
		log.Fatalf("Failed to load skills: %v", err)
	}

	agts, err := loadAgents(filepath.Join(specDir, "agents"))
	if err != nil {
		log.Fatalf("Failed to load agents: %v", err)
	}

	fmt.Printf("Loaded: %d commands, %d skills, %d agents\n\n", len(cmds), len(skls), len(agts))

	// Generate Claude plugin
	fmt.Println("Generating Claude plugin...")
	if err := generateClaude(claudeDir, plugin, cmds, skls, agts); err != nil {
		log.Fatalf("Failed to generate Claude plugin: %v", err)
	}
	fmt.Printf("  Created: %s\n", claudeDir)

	// Generate Gemini extension
	fmt.Println("Generating Gemini extension...")
	if err := generateGemini(geminiDir, plugin, cmds); err != nil {
		log.Fatalf("Failed to generate Gemini extension: %v", err)
	}
	fmt.Printf("  Created: %s\n", geminiDir)

	fmt.Println("\nDone!")
}

func findBaseDir() string {
	// Try to find the project root by looking for go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// Fallback to current directory
	cwd, _ := os.Getwd()
	return cwd
}

func loadPlugin(path string) (*plugins.Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin plugins.Plugin
	if err := json.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

func loadCommands(dir string) ([]*commands.Command, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var cmds []*commands.Command
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var cmd commands.Command
		if err := json.Unmarshal(data, &cmd); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		cmds = append(cmds, &cmd)
	}

	return cmds, nil
}

func loadSkills(dir string) ([]*skills.Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var skls []*skills.Skill
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var skl skills.Skill
		if err := json.Unmarshal(data, &skl); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		skls = append(skls, &skl)
	}

	return skls, nil
}

func loadAgents(dir string) ([]*agents.Agent, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var agts []*agents.Agent
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var agt agents.Agent
		if err := json.Unmarshal(data, &agt); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		agts = append(agts, &agt)
	}

	return agts, nil
}

func generateClaude(dir string, plugin *plugins.Plugin, cmds []*commands.Command, skls []*skills.Skill, agts []*agents.Agent) error {
	// Get adapters
	pluginAdapter, ok := plugins.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude plugin adapter not found")
	}

	cmdAdapter, ok := commands.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude command adapter not found")
	}

	skillAdapter, ok := skills.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude skill adapter not found")
	}

	agentAdapter, ok := agents.GetAdapter("claude")
	if !ok {
		return fmt.Errorf("claude agent adapter not found")
	}

	// Write plugin structure
	if err := pluginAdapter.WritePlugin(plugin, dir); err != nil {
		return fmt.Errorf("write plugin: %w", err)
	}

	// Write commands
	commandsDir := filepath.Join(dir, "commands")
	if err := os.MkdirAll(commandsDir, 0700); err != nil {
		return err
	}
	for _, cmd := range cmds {
		path := filepath.Join(commandsDir, cmd.Name+".md")
		if err := cmdAdapter.WriteFile(cmd, path); err != nil {
			return fmt.Errorf("write command %s: %w", cmd.Name, err)
		}
	}

	// Write skills
	skillsDir := filepath.Join(dir, "skills")
	for _, skl := range skls {
		if err := skillAdapter.WriteSkillDir(skl, skillsDir); err != nil {
			return fmt.Errorf("write skill %s: %w", skl.Name, err)
		}
	}

	// Write agents
	agentsDir := filepath.Join(dir, "agents")
	if err := os.MkdirAll(agentsDir, 0700); err != nil {
		return err
	}
	for _, agt := range agts {
		path := filepath.Join(agentsDir, agt.Name+".md")
		if err := agentAdapter.WriteFile(agt, path); err != nil {
			return fmt.Errorf("write agent %s: %w", agt.Name, err)
		}
	}

	// Copy hooks from spec if present
	specHooksDir := filepath.Join(filepath.Dir(dir), "spec", "hooks")
	if _, err := os.Stat(specHooksDir); err == nil {
		hooksDir := filepath.Join(dir, "hooks")
		if err := copyDir(specHooksDir, hooksDir); err != nil {
			return fmt.Errorf("copy hooks: %w", err)
		}
	}

	return nil
}

func generateGemini(dir string, plugin *plugins.Plugin, cmds []*commands.Command) error {
	// Get adapters
	pluginAdapter, ok := plugins.GetAdapter("gemini")
	if !ok {
		return fmt.Errorf("gemini plugin adapter not found")
	}

	cmdAdapter, ok := commands.GetAdapter("gemini")
	if !ok {
		return fmt.Errorf("gemini command adapter not found")
	}

	// Write plugin structure
	if err := pluginAdapter.WritePlugin(plugin, dir); err != nil {
		return fmt.Errorf("write plugin: %w", err)
	}

	// Write commands (Gemini uses TOML)
	commandsDir := filepath.Join(dir, "commands")
	if err := os.MkdirAll(commandsDir, 0700); err != nil {
		return err
	}
	for _, cmd := range cmds {
		path := filepath.Join(commandsDir, cmd.Name+".toml")
		if err := cmdAdapter.WriteFile(cmd, path); err != nil {
			return fmt.Errorf("write command %s: %w", cmd.Name, err)
		}
	}

	// Copy hooks from spec if present
	specHooksDir := filepath.Join(filepath.Dir(dir), "spec", "hooks")
	if _, err := os.Stat(specHooksDir); err == nil {
		hooksDir := filepath.Join(dir, "hooks")
		if err := copyDir(specHooksDir, hooksDir); err != nil {
			return fmt.Errorf("copy hooks: %w", err)
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0700); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0600); err != nil {
				return err
			}
		}
	}

	return nil
}
