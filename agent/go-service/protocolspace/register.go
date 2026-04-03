package protocolspace

import maa "github.com/MaaXYZ/maa-framework-go/v4"

// Register registers custom actions for the ProtocolSpace task.
func Register() {
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisCopyAction", &FightAxisCopyAction{})
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisImportAction", &FightAxisImportAction{})
}
