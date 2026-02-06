package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, yaml string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(f, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

func generateToTemp(t *testing.T, configPath string) KarabinerConfig {
	t.Helper()
	out := filepath.Join(t.TempDir(), "karabiner.json")
	if err := generateKarabinerConfig(configPath, out, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var kc KarabinerConfig
	if err := json.Unmarshal(data, &kc); err != nil {
		t.Fatal(err)
	}
	return kc
}

func findRule(rules []Rule, description string) *Rule {
	for i := range rules {
		if rules[i].Description == description {
			return &rules[i]
		}
	}
	return nil
}

func ruleDescriptions(rules []Rule) []string {
	descs := make([]string, len(rules))
	for i, r := range rules {
		descs[i] = r.Description
	}
	return descs
}

// --- v1 tests ---

func TestV1_HyperkeyAndLayersPresent(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 1
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
  layers:
    - key: 'o'
      type: 'app'
      sub:
        c: /Applications/Visual Studio Code.app
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	// Should have hyperkey rule
	if r := findRule(rules, "Hyper Key (caps_lock)"); r == nil {
		t.Error("expected Hyper Key rule in v1 output")
	}

	// Should have layer rule
	if r := findRule(rules, `Hyper Key sublayer "o"`); r == nil {
		t.Error("expected Hyper Key sublayer rule in v1 output")
	}

	// Should have option keybinding
	if r := findRule(rules, "Open TBD"); r == nil {
		t.Error("expected option keybinding rule in v1 output")
	}

	// Should have HJKL
	if r := findRule(rules, "Map Option + H/J/K/L to Arrow Keys"); r == nil {
		t.Error("expected HJKL rule in v1 output")
	}
}

func TestV1_LayerTypeInOptionErrors(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 1
keybindings:
  option:
    'o':
      type: layer
      sub:
        c:
          val: /Applications/Visual Studio Code.app
          type: app
`)
	out := filepath.Join(t.TempDir(), "karabiner.json")
	err := generateKarabinerConfig(cfg, out, true)
	if err == nil {
		t.Fatal("expected error for type: layer in v1 config")
	}
}

func TestV1_ExampleConfig(t *testing.T) {
	// Test the actual example config in the repo
	examplePath := filepath.Join("..", "exmaple", "config.yaml")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("example config not found")
	}
	out := filepath.Join(t.TempDir(), "karabiner.json")
	if err := generateKarabinerConfig(examplePath, out, true); err != nil {
		t.Fatalf("v1 example config should generate without error: %v", err)
	}
}

// --- v2 tests ---

func TestV2_NoHyperkeyRules(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	for _, r := range rules {
		if r.Description == "Hyper Key (caps_lock)" {
			t.Error("v2 should not have Hyper Key rule")
		}
	}
}

func TestV2_HHKBModeWorks(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
use_hhkb: true
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	if r := findRule(rules, "HHKB Mode (Caps Lock -> Left Control)"); r == nil {
		t.Error("v2 with use_hhkb should have HHKB mode rule")
	}
	// Should NOT have hyperkey rule even with HHKB
	if r := findRule(rules, "Hyper Key (caps_lock)"); r != nil {
		t.Error("v2 should not have Hyper Key rule even with HHKB enabled")
	}
}

func TestV2_OptionSublayer(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
    'o':
      type: layer
      sub:
        c:
          val: /Applications/Visual Studio Code.app
          type: app
        g:
          val: https://chatgpt.com/
          type: web
        e:
          val: "echo hello"
          type: shell
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	// Should have option sublayer rule
	r := findRule(rules, `option sublayer "o"`)
	if r == nil {
		t.Fatalf("expected option sublayer rule, got rules: %v", ruleDescriptions(rules))
	}

	// First manipulator should be the toggle
	toggle := r.Manipulators[0]
	if toggle.From.KeyCode != "o" {
		t.Errorf("toggle from key should be 'o', got %q", toggle.From.KeyCode)
	}
	if len(toggle.From.Modifiers.Mandatory) != 1 || toggle.From.Modifiers.Mandatory[0] != "left_option" {
		t.Errorf("toggle should have left_option mandatory, got %v", toggle.From.Modifiers.Mandatory)
	}
	if toggle.To[0].SetVariable == nil || toggle.To[0].SetVariable.Name != "option_sublayer_o" {
		t.Error("toggle should set option_sublayer_o variable")
	}
	if toggle.ToAfterKeyUp[0].SetVariable == nil || toggle.ToAfterKeyUp[0].SetVariable.Name != "option_sublayer_o" {
		t.Error("toggle should unset option_sublayer_o on key up")
	}

	// Remaining manipulators should be sub-key actions
	subManipulators := r.Manipulators[1:]
	if len(subManipulators) != 3 {
		t.Fatalf("expected 3 sub-key manipulators, got %d", len(subManipulators))
	}

	// Check that all sub-key manipulators have the correct condition
	for _, m := range subManipulators {
		if len(m.Conditions) != 1 {
			t.Fatalf("sub-key manipulator should have 1 condition, got %d", len(m.Conditions))
		}
		if m.Conditions[0].Name != "option_sublayer_o" || m.Conditions[0].Value != 1 {
			t.Errorf("sub-key condition should check option_sublayer_o==1, got %s==%d", m.Conditions[0].Name, m.Conditions[0].Value)
		}
	}

	// Verify each sub-binding type produces correct To action
	subByKey := map[string]Manipulator{}
	for _, m := range subManipulators {
		subByKey[m.From.KeyCode] = m
	}

	// app type
	if cm, ok := subByKey["c"]; ok {
		if cm.To[0].SoftwareFunction == nil || cm.To[0].SoftwareFunction.OpenApplication == nil {
			t.Error("app sub-binding should use SoftwareFunction.OpenApplication")
		}
	} else {
		t.Error("expected sub-key 'c'")
	}

	// web type
	if gm, ok := subByKey["g"]; ok {
		if gm.To[0].ShellCommand != "open https://chatgpt.com/" {
			t.Errorf("web sub-binding should use 'open URL', got %q", gm.To[0].ShellCommand)
		}
	} else {
		t.Error("expected sub-key 'g'")
	}

	// shell type
	if em, ok := subByKey["e"]; ok {
		if em.To[0].ShellCommand != "echo hello" {
			t.Errorf("shell sub-binding should use raw command, got %q", em.To[0].ShellCommand)
		}
	} else {
		t.Error("expected sub-key 'e'")
	}

	// Should NOT have hyper layer rules
	for _, rule := range rules {
		for _, desc := range []string{"Hyper Key (caps_lock)", `Hyper Key sublayer`} {
			if rule.Description == desc {
				t.Errorf("v2 should not have %q rule", desc)
			}
		}
	}

	// Should still have plain option binding
	if r := findRule(rules, "Open TBD"); r == nil {
		t.Error("expected plain option keybinding rule")
	}
}

func TestV2_MutualExclusion(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    'o':
      type: layer
      sub:
        c:
          val: /Applications/Visual Studio Code.app
          type: app
    'p':
      type: layer
      sub:
        d:
          val: /Applications/DBeaver.app
          type: app
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	ruleO := findRule(rules, `option sublayer "o"`)
	ruleP := findRule(rules, `option sublayer "p"`)
	if ruleO == nil || ruleP == nil {
		t.Fatalf("expected both sublayer rules, got: %v", ruleDescriptions(rules))
	}

	// Toggle of 'o' should have mutual exclusion condition for 'p'
	toggleO := ruleO.Manipulators[0]
	found := false
	for _, c := range toggleO.Conditions {
		if c.Name == "option_sublayer_p" && c.Value == 0 {
			found = true
		}
	}
	if !found {
		t.Error("sublayer 'o' toggle should have mutual exclusion condition for 'p'")
	}

	// Toggle of 'p' should have mutual exclusion condition for 'o'
	toggleP := ruleP.Manipulators[0]
	found = false
	for _, c := range toggleP.Conditions {
		if c.Name == "option_sublayer_o" && c.Value == 0 {
			found = true
		}
	}
	if !found {
		t.Error("sublayer 'p' toggle should have mutual exclusion condition for 'o'")
	}
}

func TestV2_IgnoresLayersField(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
  layers:
    - key: 'o'
      type: 'app'
      sub:
        c: /Applications/Visual Studio Code.app
`)
	kc := generateToTemp(t, cfg)
	rules := kc.Profiles[0].ComplexModifications.Rules

	for _, r := range rules {
		if r.Description == `Hyper Key sublayer "o"` {
			t.Error("v2 should ignore keybindings.layers field")
		}
	}
}

func TestV2_NestedLayerErrors(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    'o':
      type: layer
      sub:
        c:
          type: layer
          sub:
            x:
              val: /Applications/Safari.app
              type: app
`)
	out := filepath.Join(t.TempDir(), "karabiner.json")
	err := generateKarabinerConfig(cfg, out, true)
	if err == nil {
		t.Fatal("expected error for nested layer in v2 config")
	}
}

func TestV2_EmptySubErrors(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    'o':
      type: layer
      sub: {}
`)
	out := filepath.Join(t.TempDir(), "karabiner.json")
	err := generateKarabinerConfig(cfg, out, true)
	if err == nil {
		t.Fatal("expected error for empty sub in v2 config")
	}
}

func TestV2_SubBindingEmptyValErrors(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 2
keybindings:
  option:
    'o':
      type: layer
      sub:
        c:
          type: app
`)
	out := filepath.Join(t.TempDir(), "karabiner.json")
	err := generateKarabinerConfig(cfg, out, true)
	if err == nil {
		t.Fatal("expected error for sub-binding with empty val")
	}
}

// --- version validation ---

func TestUnsupportedVersionErrors(t *testing.T) {
	cfg := writeTestConfig(t, `
version: 3
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
`)
	out := filepath.Join(t.TempDir(), "karabiner.json")
	err := generateKarabinerConfig(cfg, out, true)
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

// --- round-trip ---

func TestRoundTrip_ValidJSON(t *testing.T) {
	configs := []string{
		// v1
		`
version: 1
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
  layers:
    - key: 'o'
      type: 'app'
      sub:
        c: /Applications/Visual Studio Code.app
`,
		// v2
		`
version: 2
keybindings:
  option:
    '1':
      val: /Applications/Safari.app
      type: app
    'o':
      type: layer
      sub:
        c:
          val: /Applications/Visual Studio Code.app
          type: app
`,
	}

	for i, yamlCfg := range configs {
		cfg := writeTestConfig(t, yamlCfg)
		out := filepath.Join(t.TempDir(), "karabiner.json")
		if err := generateKarabinerConfig(cfg, out, true); err != nil {
			t.Fatalf("config %d: generate failed: %v", i, err)
		}
		data, err := os.ReadFile(out)
		if err != nil {
			t.Fatalf("config %d: read failed: %v", i, err)
		}
		var kc KarabinerConfig
		if err := json.Unmarshal(data, &kc); err != nil {
			t.Fatalf("config %d: JSON round-trip failed: %v", i, err)
		}
		if len(kc.Profiles) == 0 {
			t.Fatalf("config %d: no profiles in output", i)
		}
	}
}
