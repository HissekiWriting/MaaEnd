package protocolspace

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

var (
	_ maa.CustomActionRunner = &FightAxisCopyAction{}
	_ maa.CustomActionRunner = &FightAxisImportAction{}
	_ maa.CustomActionRunner = &FightAxisLoadDefaultAction{}
	_ maa.CustomActionRunner = &FightAxisLoadCustomAction{}
)

const (
	fightAxisDefaultFile = "assets/fight_axis/default.json"
	fightAxisUserFile    = "assets/fight_axis/user_axis.json"
)

type fightAxisActionParam struct {
	Source string `json:"source"`
	Input  string `json:"input"`
}

type fightAxisConfig struct {
	Version        int             `json:"version"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	BattleMode     string          `json:"battle_mode"`
	OperatorCount  int             `json:"operator_count"`
	ImmediateCast  []fightAxisCast `json:"immediate_cast"`
	SequentialCast []fightAxisCast `json:"sequential_cast"`
}

type fightAxisCast struct {
	Type       string `json:"type"`
	OperatorID int    `json:"operator_id,omitempty"`
}

// FightAxisCopyAction copies the custom fight axis file content to the clipboard.
type FightAxisCopyAction struct{}

// FightAxisLoadDefaultAction activates the default fight axis.
type FightAxisLoadDefaultAction struct{}

// FightAxisLoadCustomAction activates the custom fight axis or falls back to default if invalid.
type FightAxisLoadCustomAction struct{}

// Run copies user_axis.json content to clipboard and shows a confirmation message.
func (a *FightAxisCopyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	_ = arg
	content, err := os.ReadFile(resolveFightAxisPath(fightAxisUserFile))
	if err != nil {
		log.Error().Err(err).Str("file", fightAxisUserFile).Msg("failed to read custom fight axis file")
		return false
	}
	if err := writeClipboard(string(content)); err != nil {
		log.Error().Err(err).Str("file", fightAxisUserFile).Msg("failed to copy custom fight axis file to clipboard")
		return false
	}
	log.Info().Str("component", "FightAxisCopyAction").Msg("copied custom fight axis content to clipboard")
	return true
}

// Run restores the default fight axis into user_axis.json.
func (a *FightAxisLoadDefaultAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	_ = arg
	if err := restoreDefaultFightAxis(); err != nil {
		log.Error().Err(err).Str("component", "FightAxisLoadDefaultAction").Msg("failed to activate default fight axis")
		return false
	}
	log.Info().Str("component", "FightAxisLoadDefaultAction").Msg("activated default fight axis")
	return true
}

// Run validates the custom fight axis file and restores the default axis when the custom one is invalid.
func (a *FightAxisLoadCustomAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	_ = arg
	content, err := os.ReadFile(resolveFightAxisPath(fightAxisUserFile))
	if err != nil {
		log.Error().Err(err).Str("component", "FightAxisLoadCustomAction").Msg("failed to read custom fight axis")
		if fallbackErr := restoreDefaultFightAxis(); fallbackErr != nil {
			log.Error().Err(fallbackErr).Str("component", "FightAxisLoadCustomAction").Msg("failed to restore default fight axis")
			return false
		}
		log.Warn().Str("component", "FightAxisLoadCustomAction").Msg("failed to read custom fight axis, restored default axis")
		return false
	}
	if err := validateFightAxisContent(content); err != nil {
		log.Error().Err(err).Str("component", "FightAxisLoadCustomAction").Msg("custom fight axis is invalid")
		if fallbackErr := restoreDefaultFightAxis(); fallbackErr != nil {
			log.Error().Err(fallbackErr).Str("component", "FightAxisLoadCustomAction").Msg("failed to restore default fight axis")
			return false
		}
		log.Warn().Str("component", "FightAxisLoadCustomAction").Msg("custom fight axis invalid, restored default axis")
		return false
	}
	log.Info().Str("component", "FightAxisLoadCustomAction").Msg("activated custom fight axis")
	return true
}

// FightAxisImportAction imports a JSON file into user_axis.json.
type FightAxisImportAction struct{}

// Run opens a file picker, then replaces user_axis.json with the selected file content.
func (a *FightAxisImportAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	axisContent, err := resolveFightAxisContent(arg)
	if err != nil {
		log.Error().Err(err).Str("component", "FightAxisImportAction").Msg("failed to resolve fight axis input")
		if fallbackErr := restoreDefaultFightAxis(); fallbackErr != nil {
			log.Error().Err(fallbackErr).Str("component", "FightAxisImportAction").Msg("failed to restore default fight axis")
			return false
		}
		log.Warn().Str("component", "FightAxisImportAction").Msg("invalid fight axis input, restored default axis")
		return false
	}
	if err := validateFightAxisContent(axisContent); err != nil {
		log.Error().Err(err).Str("component", "FightAxisImportAction").Msg("fight axis validation failed")
		if fallbackErr := restoreDefaultFightAxis(); fallbackErr != nil {
			log.Error().Err(fallbackErr).Str("component", "FightAxisImportAction").Msg("failed to restore default fight axis")
			return false
		}
		log.Warn().Str("component", "FightAxisImportAction").Msg("fight axis validation failed, restored default axis")
		return false
	}
	if err := os.WriteFile(resolveFightAxisPath(fightAxisUserFile), axisContent, 0644); err != nil {
		log.Error().Err(err).Str("file", fightAxisUserFile).Msg("failed to overwrite custom fight axis file")
		if fallbackErr := restoreDefaultFightAxis(); fallbackErr != nil {
			log.Error().Err(fallbackErr).Str("component", "FightAxisImportAction").Msg("failed to restore default fight axis")
			return false
		}
		log.Warn().Str("component", "FightAxisImportAction").Msg("failed to write custom fight axis file, restored default axis")
		return false
	}
	log.Info().Str("component", "FightAxisImportAction").Msg("imported custom fight axis successfully")
	return true
}

func resolveFightAxisContent(arg *maa.CustomActionArg) ([]byte, error) {
	var params fightAxisActionParam
	if err := json.Unmarshal([]byte(arg.CustomActionParam), &params); err != nil {
		return nil, fmt.Errorf("parse fight axis action param: %w", err)
	}
	input := strings.TrimSpace(params.Input)
	source := strings.TrimSpace(strings.Trim(params.Source, "{}"))
	if source == "" {
		source = inferFightAxisSource(input)
	}
	switch strings.ToLower(source) {
	case "file_path":
		if input == "" {
			return nil, fmt.Errorf("file path is empty")
		}
		content, err := os.ReadFile(resolveFightAxisPath(input))
		if err != nil {
			return nil, fmt.Errorf("read fight axis file: %w", err)
		}
		return content, nil
	case "data_code":
		if input == "" {
			return nil, fmt.Errorf("data code is empty")
		}
		return []byte(input), nil
	default:
		return nil, fmt.Errorf("unsupported fight axis source: %s", params.Source)
	}
}

func inferFightAxisSource(input string) string {
	if input == "" {
		return ""
	}
	if strings.ContainsAny(input, `\\/`) || strings.HasSuffix(strings.ToLower(input), ".json") {
		return "file_path"
	}
	return "data_code"
}

func validateFightAxisContent(content []byte) error {
	var axis fightAxisConfig
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&axis); err != nil {
		return fmt.Errorf("decode fight axis json: %w", err)
	}
	if axis.Version <= 0 {
		return fmt.Errorf("version must be positive")
	}
	if strings.TrimSpace(axis.Name) == "" {
		return fmt.Errorf("name is empty")
	}
	if strings.TrimSpace(axis.BattleMode) == "" {
		return fmt.Errorf("battle_mode is empty")
	}
	if axis.OperatorCount <= 0 {
		return fmt.Errorf("operator_count must be positive")
	}
	for index, cast := range axis.ImmediateCast {
		if err := validateFightAxisCast(cast); err != nil {
			return fmt.Errorf("immediate_cast[%d]: %w", index, err)
		}
	}
	for index, cast := range axis.SequentialCast {
		if err := validateFightAxisCast(cast); err != nil {
			return fmt.Errorf("sequential_cast[%d]: %w", index, err)
		}
	}
	return nil
}

func validateFightAxisCast(cast fightAxisCast) error {
	if strings.TrimSpace(cast.Type) == "" {
		return fmt.Errorf("type is empty")
	}
	switch cast.Type {
	case "combo":
		if cast.OperatorID != 0 {
			return fmt.Errorf("combo must not set operator_id")
		}
	case "skill", "ultimate":
		if cast.OperatorID <= 0 {
			return fmt.Errorf("operator_id must be positive for %s", cast.Type)
		}
	default:
		return fmt.Errorf("unsupported type %q", cast.Type)
	}
	return nil
}

func restoreDefaultFightAxis() error {
	content, err := os.ReadFile(resolveFightAxisPath(fightAxisDefaultFile))
	if err != nil {
		return fmt.Errorf("read default fight axis: %w", err)
	}
	if err := validateFightAxisContent(content); err != nil {
		return fmt.Errorf("validate default fight axis: %w", err)
	}
	if err := os.WriteFile(resolveFightAxisPath(fightAxisUserFile), content, 0644); err != nil {
		return fmt.Errorf("write default fight axis to user file: %w", err)
	}
	return nil
}

func resolveFightAxisPath(relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	searchBases := []string{}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		searchBases = append(searchBases, exeDir, filepath.Dir(exeDir), filepath.Dir(filepath.Dir(exeDir)))
	}
	if cwd, err := os.Getwd(); err == nil {
		searchBases = append(searchBases, cwd, filepath.Dir(cwd), filepath.Dir(filepath.Dir(cwd)))
	}

	for _, base := range searchBases {
		candidate := filepath.Join(base, relativePath)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate
		}
	}
	return relativePath
}

func writeClipboard(content string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("clipboard is only supported on windows in this action")
	}
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard -Value ([Console]::In.ReadToEnd())")
	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

func pickJSONFile() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("file picker is only supported on windows in this action")
	}
	ps := `(Add-Type -AssemblyName System.Windows.Forms | Out-Null; $dialog = New-Object System.Windows.Forms.OpenFileDialog; $dialog.Filter = 'JSON Files (*.json)|*.json|All Files (*.*)|*.*'; $dialog.Multiselect = $false; if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($dialog.FileName) })`
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", fmt.Errorf("no file selected")
	}
	return path, nil
}

func escapePowerShellString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
