package config

type Model struct {
	Name              string     `yaml:"Name"`
	SwitchName        string     `yaml:"SwitchName"`
	Pipeline          []string   `yaml:"Pipeline"`
	DefaultLibraries  []string   `yaml:"DefaultLibraries"`
	IntrinsicMetadata struct {
	} `yaml:"IntrinsicMetadata"`
	UserMetadata []struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	} `yaml:"UserMetadata"`
	ProgrammableBlocks []ProgrammableBlock `yaml:"ProgrammableBlocks"`
	HeaderModelConfig HeaderModelConfig    `yaml:"HeaderModelConfig"`
	Main struct {
		Filename string `yaml:"filename"`
	} `yaml:"Main"`
}

type ProgrammableBlock struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Filename    string      `yaml:"filename"`
	Code        string      `yaml:"code,omitempty"`
	HasApply    bool        `yaml:"hasApply"`
	Abstraction string      `yaml:"abstraction"`
	Parameters  []Parameter `yaml:"parameters"`
}

type Parameter  struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Direction string `yaml:"direction,omitempty"`
}

type HeaderModelConfig struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
