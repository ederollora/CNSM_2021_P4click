package config

type Defines struct {
	DefineList []Define `yaml:"defines"`
}
type Define struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}