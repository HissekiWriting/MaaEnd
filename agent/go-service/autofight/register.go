package autofight

import (
	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

var (
	_ maa.CustomRecognitionRunner = &AutoFightEntryRecognition{}
	_ maa.CustomRecognitionRunner = &AutoFightExitRecognition{}
	_ maa.CustomRecognitionRunner = &AutoFightPauseRecognition{}
	_ maa.CustomRecognitionRunner = &AutoFightExecuteRecognition{}
	_ maa.CustomActionRunner      = &AutoFightExecuteAction{}
)

// Register registers all custom recognition and action components for autofight package
func Register() {
	const axisPath = "assets/resource/skill_axis/DefaultAxis.json"
	if err := loadSkillAxis(axisPath); err != nil {
		log.Warn().
			Err(err).
			Str("path", axisPath).
			Msg("failed to load skill axis, using defaults")
	}
	maa.AgentServerRegisterCustomRecognition("AutoFightEntryRecognition", &AutoFightEntryRecognition{})
	maa.AgentServerRegisterCustomRecognition("AutoFightExitRecognition", &AutoFightExitRecognition{})
	maa.AgentServerRegisterCustomRecognition("AutoFightPauseRecognition", &AutoFightPauseRecognition{})
	maa.AgentServerRegisterCustomRecognition("AutoFightExecuteRecognition", &AutoFightExecuteRecognition{})
	maa.AgentServerRegisterCustomAction("AutoFightExecuteAction", &AutoFightExecuteAction{})
}
