package protocolspace

import maa "github.com/MaaXYZ/maa-framework-go/v4"

// Register registers custom actions for the ProtocolSpace task.
func Register() {
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisLoadDefaultAction", &FightAxisLoadDefaultAction{})
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisLoadCustomAction", &FightAxisLoadCustomAction{})
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisCopyAction", &FightAxisCopyAction{})
	maa.AgentServerRegisterCustomAction("ProtocolSpaceFightAxisImportAction", &FightAxisImportAction{})
}
