package structs

type Material struct {
	Type     string
	URI      string `yaml:"uri"`
	Metadata map[string]string
}

type Artifact struct {
	Path string
	Name string
}

type Command struct {
	Command string
	Cwd     string
}

type Pipeline struct {
	Name      string
	Materials []Material
	Artifacts []Artifact
	Scripts   []Command
	Container string
}
