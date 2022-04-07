package config

type Typedefs struct {
	TypedefList []Typedef `yaml:"typedefs"`
}

type Typedef struct {
	Name string `yaml:"name"`
	Size int    `yaml:"size"`
}