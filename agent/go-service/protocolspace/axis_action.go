package protocolspace

import (
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
)

const (
	fightAxisDefaultFile = "assets/fight_axis/default.json"
	fightAxisUserFile    = "assets/fight_axis/user_axis.json"
)

// FightAxisCopyAction copies the custom fight axis file content to the clipboard.
type FightAxisCopyAction struct{}

// Run copies user_axis.json content to clipboard and shows a confirmation message.
func (a *FightAxisCopyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	_ = arg
	content, err := os.ReadFile(resolveFightAxisPath(fightAxisUserFile))
	if err != nil {
		log.Error().Err(err).Str("file", fightAxisUserFile).Msg("failed to read custom fight axis file")
		showMessageBox("提示", "读取自定义排轴文件失败")
		return false
	}
	if err := writeClipboard(string(content)); err != nil {
		log.Error().Err(err).Str("file", fightAxisUserFile).Msg("failed to copy custom fight axis file to clipboard")
		showMessageBox("提示", "复制自定义排轴文件失败")
		return false
	}
	showMessageBox("提示", "已复制自定义排轴文件内容")
	return true
}

// FightAxisImportAction imports a JSON file into user_axis.json.
type FightAxisImportAction struct{}

// Run opens a file picker, then replaces user_axis.json with the selected file content.
func (a *FightAxisImportAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_ = ctx
	_ = arg
	selectedPath, err := pickJSONFile()
	if err != nil {
		log.Error().Err(err).Msg("failed to select fight axis json file")
		showMessageBox("提示", "未选择可导入的 JSON 文件")
		return false
	}
	content, err := os.ReadFile(selectedPath)
	if err != nil {
		log.Error().Err(err).Str("path", selectedPath).Msg("failed to read selected fight axis file")
		showMessageBox("提示", "读取选中的 JSON 文件失败")
		return false
	}
	if err := os.WriteFile(resolveFightAxisPath(fightAxisUserFile), content, 0644); err != nil {
		log.Error().Err(err).Str("path", selectedPath).Msg("failed to overwrite custom fight axis file")
		showMessageBox("提示", "写入自定义排轴文件失败")
		return false
	}
	showMessageBox("提示", "导入自定义排轴文件成功")
	return true
}

func resolveFightAxisPath(relativePath string) string {
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, relativePath)
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

func showMessageBox(title, message string) {
	if runtime.GOOS != "windows" {
		log.Info().Str("title", title).Str("message", message).Msg("message box fallback")
		return
	}
	ps := fmt.Sprintf("Add-Type -AssemblyName PresentationFramework | Out-Null; [System.Windows.MessageBox]::Show('%s','%s') | Out-Null", escapePowerShellString(message), escapePowerShellString(title))
	_ = exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps).Run()
}

func escapePowerShellString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
