package arguments

import (
	//"fmt"
	prompt "github.com/c-bata/go-prompt"
	. "p4click/commands"
	. "p4click/context"
	. "p4click/types"
)


func ArgumentsCompleter(args []string, context []ContextType, cd *DeploymentConfig, ad []DeploymentConfig,
	cm *MachineConfig, cc *ControllerConfig, cw *SwitchPipeline) []prompt.Suggest {

	if context[len(context)-1].Ctype == TOOL{

		var cmds []prompt.Suggest

		if len(args) <= 1 {

			cmds = ToolCommands
			for i := range AllCommands{
				cmds = append(cmds, AllCommands[i])
			}

			return prompt.FilterHasPrefix(cmds, args[0], true)
		}else
		if args[0] == "create" && len(args) <= 2{
			return prompt.FilterHasPrefix(ToolCreateCommands, args[1], true)
		}else
		if args[0] == "use"{

			if len(args) <= 2{
				return prompt.FilterHasPrefix(ToolUseCommands, args[1], true)
			}

			if len(args) == 3 && args[1] == "deployment"{
				suggestions := []prompt.Suggest{}

				for i := range ad {
					entry := prompt.Suggest{Text: ad[i].Name, Description: "Test that connection works"}
					suggestions = append(suggestions, entry)
				}
				return prompt.FilterHasPrefix(suggestions, args[2], true)
			}
		}else
		if args[0] == "default"{

			if len(args) <= 2{
				return prompt.FilterHasPrefix(ToolDefaultCommands, args[1], true)
			}
		}
	}else
	if context[len(context)-1].Ctype == DEPLOYMENT{

		var cmds []prompt.Suggest

		if len(args) <= 1 {

			cmds = DeploymentCommands
			for i := range AllCommands{
				cmds = append(cmds, AllCommands[i])
			}

			return prompt.FilterHasPrefix(cmds, args[0], true)
		}else
		if args[0] == "create" && len(args) <= 2{
			return prompt.FilterHasPrefix(DeploymentCreateCommands, args[1], true)
		}else
		if args[0] == "use"{

			if len(args) <= 2{
				return prompt.FilterHasPrefix(DeploymentUseCommands, args[1], true)
			}

			if len(args) <= 3 && args[1] == "machine"{
				suggestions := []prompt.Suggest{}

				for i := range cd.ControllerMachines{
					entry := prompt.Suggest{Text: cd.ControllerMachines[i].Name, Description: "Name of a machine"}
					suggestions = append(suggestions, entry)
				}
				return prompt.FilterHasPrefix(suggestions, args[2], true)
			}else
			if len(args) <= 3 && args[1] == "switch"{
				suggestions := []prompt.Suggest{}

				for i := range cd.Switches{
					entry := prompt.Suggest{Text: cd.Switches[i].Name, Description: "Name of a switch"}
					suggestions = append(suggestions, entry)
				}
				return prompt.FilterHasPrefix(suggestions, args[2], true)
			}
		}else
		if args[0] == "add"{
			if len(args) <= 2{
				return prompt.FilterHasPrefix(DeploymentAddCommands, args[1], true)
			}
		}
	}else
	if context[len(context)-1].Ctype == MACHINE{

 		var cmds []prompt.Suggest

		if len(args) <= 1 {
			cmds = MachineCommands
			for i := range AllCommands{
				cmds = append(cmds, AllCommands[i])
			}
			return prompt.FilterHasPrefix(cmds, args[0], true)
		}else
		if args[0] == "create"{

			if len(args) <= 2 {
				return prompt.FilterHasPrefix(MachineCreateCommands, args[1], true)
			}

			if len(args) <= 3 && args[1] == "controller" {
				return prompt.FilterHasPrefix(MachineControllerCommands, args[2], true)
			}

		}else
		if args[0] == "set"{
			if len(args) > 2 && args[1] == "ssh"{
				if len(args) > 3 && args[2] == "auth"{
					return prompt.FilterHasPrefix(MachineSshAuthCommands, args[3], true)
				}
				if len(args) < 4 {
					return prompt.FilterHasPrefix(MachineSshCommands, args[2], true)
				}
			}
			return prompt.FilterHasPrefix(MachineSetCommands, args[1], true)
		}else
		if args[0] == "use"{

			if len(args) <= 2{
				return prompt.FilterHasPrefix(MachineUseCommands, args[1], true)
			}

			if len(args) == 3 && args[1] == "controller"{
				suggestions := []prompt.Suggest{}

				for i := range cm.ControllersConfig{
					entry := prompt.Suggest{Text: cm.ControllersConfig[i].ControllerName, Description: "Controller name"}
					suggestions = append(suggestions, entry)
				}
				return prompt.FilterHasPrefix(suggestions, args[2], true)
			}
		}
	}else
	if context[len(context)-1].Ctype == CONTROLLER{

		var cmds []prompt.Suggest

		if len(args) <= 1 {
			cmds = ControllerCommands
			for i := range AllCommands{
				cmds = append(cmds, AllCommands[i])
			}
			return prompt.FilterHasPrefix(cmds, args[0], true)
		}else
		if args[0] == "add"{

			if len(args) <= 1{
				return prompt.FilterHasPrefix(ControllerAddCommands, args[1], true)
			}

			if len(args) <= 2 && args[1] == "app"{
				//TODO: Get apps that can be installed
				return []prompt.Suggest{}
			}

			return prompt.FilterHasPrefix(ControllerAddCommands, args[0], true)
		}else
		if args[0] == "disable" || args[0] == "enable"{
			if len(args) == 2 {
				return prompt.FilterHasPrefix(ControllerEnableCommands, args[0], true)
			}
		}
	}else
	if context[len(context)-1].Ctype == SWITCH{

		var cmds []prompt.Suggest

		if len(args) <= 1 {
			cmds = SwitchCommands
			for i := range AllCommands{
				cmds = append(cmds, AllCommands[i])
			}
			return prompt.FilterHasPrefix(cmds, args[0], true)
		}else
		if args[0] == "add"{
			if len(args) <= 2{
				return prompt.FilterHasPrefix(SwitchAddCommands, args[1], true)
			}

		}else
		if args[0] == "set"{
			if len(args) <= 2{
				return prompt.FilterHasPrefix(SwitchSetCommands, args[1], true)
			}

		}else
		if args[0] == "build"{
			if len(args) <= 2{
				return prompt.FilterHasPrefix(SwitchBuildCommands, args[1], true)
			}

		}else
		if args[0] == "show"{
			if len(args) <= 2{
				return prompt.FilterHasPrefix(SwitchShowCommands, args[1], true)
			}

		}
	}

	return []prompt.Suggest{}
}