package config

type Constants struct {
	ConstantList []Constant `yaml:"constants"`
}
type Constant struct {
	Name  string `yaml:"name"`
	Size  string `yaml:"size"`
	Value int    `yaml:"value"`
}