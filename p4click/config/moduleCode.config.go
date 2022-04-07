package config

type ModuleCodeConfig struct {
	Include   Include     `yaml:"include"`
	Integrate []Integrate `yaml:"integrate"`
}
type Include struct {
	Name []string `yaml:"name"`
}
type Arguments struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
}
type Integrate struct {
	Logic       string      `yaml:"logic"`
	Block       string      `yaml:"block"`
	ControlName string      `yaml:"controlName"`
	CallControl bool        `yaml:"callControl"`
	Arguments   []Arguments `yaml:"arguments,omitempty"`
	Merge       bool        `yaml:"merge,omitempty"`
}