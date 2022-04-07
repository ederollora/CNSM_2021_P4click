package config


type MetadataConfig struct {
	Metadatas []Metadata `yaml:"metadata"`
}
type Metadata struct {
	Type       string   `yaml:"type"`
	Name       string   `yaml:"name"`
	Parent     string   `yaml:"parent"`
	Fields     []Field `yaml:"fields"`
	Statements []string `yaml:"statements"`
}
