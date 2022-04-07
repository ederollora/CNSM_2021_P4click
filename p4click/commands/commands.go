package commands

import "github.com/c-bata/go-prompt"

var ToolCommands = []prompt.Suggest{
	{Text: "create", Description: ""},
	{Text: "deploy", Description: ""},
	{Text: "use", Description: ""},
	{Text: "add", Description: ""},
	{Text: "default", Description: ""},
}

var DeploymentCommands = []prompt.Suggest{
	{Text: "create", Description: ""},
	{Text: "use", Description: ""},
	{Text: "add", Description: ""},
}

var MachineCommands = []prompt.Suggest{
	{Text: "create", Description: "Create a new entity within the machine"},
	{Text: "set", Description: "Configure a feature for the current machine"},
	{Text: "use", Description: "Switch context to another entity"},
	{Text: "test", Description: "Test that p4click can connect to the machine"},
}

var ControllerCommands = []prompt.Suggest{
	{Text: "add", Description: "Command to incorporate a module"},
	{Text: "disable", Description: "Turn to false a particular feature for the controller"},
	{Text: "enable", Description: "Turn to true a particular feature for the controller"},
	{Text: "test", Description: "Test that the system can add a controller"},
}

var SwitchCommands = []prompt.Suggest{
	{Text: "add", Description: "Include a new module"},
	{Text: "export", Description: "Extract some kind of asset from the switch"},
	{Text: "show", Description: "List added modules for current switch"},
	{Text: "build", Description: "Integrate modules"},
	{Text: "set", Description: "Configure a particular feature for a switch."},
}

var AllCommands = []prompt.Suggest{
	{Text: "exit", Description: "Move to a previous context stage"},
	{Text: "save", Description: "Preserve the current configuration"},
}

var ToolCreateCommands = []prompt.Suggest{
	{Text: "deployment", Description: "Create a new deployment"},
}

var ToolUseCommands = []prompt.Suggest{
	{Text: "deployment", Description: "Switch context to a deployment"},
}

var ToolDefaultCommands = []prompt.Suggest{
	{Text: "config", Description: "Use default configuration"},
}

var DeploymentAddCommands = []prompt.Suggest{
	{Text: "repository", Description: "The endpoint to retrieve modules"},
}

var DeploymentCreateCommands = []prompt.Suggest{
	{Text: "machine", Description: "Create a new machine"},
	{Text: "switch", Description: "Create a new machine"},
}

var DeploymentUseCommands = []prompt.Suggest{
	{Text: "machine", Description: "Move context to a machine"},
	{Text: "switch", Description: "Move context to a switch"},
}

var MachineCreateCommands = []prompt.Suggest{
	{Text: "controller", Description: "Create a new controller"},
}

var MachineControllerCommands = []prompt.Suggest{
	{Text: "onos", Description: "Create a new controller"},
}

var MachineSetCommands = []prompt.Suggest{
	{Text: "ssh", Description: "Set SSH parameters"},
	{Text: "address", Description: "Set address parameters"},
}

var MachineUseCommands = []prompt.Suggest{
	{Text: "controller", Description: "Switch context to controller"},
}

var MachineSshCommands = []prompt.Suggest{
	{Text: "auth", Description: "Create a new machine"},
	{Text: "keypath", Description: "Create a new machine"},
	{Text: "user", Description: "Create a new machine"},
}

var MachineSshAuthCommands = []prompt.Suggest{
	{Text: "pka", Description: "SSH auth is user/private-key based."},
	{Text: "password", Description: "SSH auth is user/password based"},
}

var ControllerAddCommands = []prompt.Suggest{
	{Text: "app", Description: "Create a new machine"},
}

var ControllerEnableCommands = []prompt.Suggest{
	{Text: "cluster", Description: "Create a new machine"},
}

var SwitchAddCommands = []prompt.Suggest{
	{Text: "module", Description: "Add a module to the data plane"},
}

var SwitchSetCommands = []prompt.Suggest{
	{Text: "model", Description: "Set a particular model associated to the target"},
	{Text: "target", Description: "Set a particular target for the switch"},
	{Text: "module-order", Description: "Set the order of modules"},
}

var SwitchBuildCommands = []prompt.Suggest{
	{Text: "data plane", Description: "Merge modules and build the data plane"},
}
var SwitchShowCommands = []prompt.Suggest{
	{Text: "modules", Description: "Show the modules that have been added to the switch"},
}