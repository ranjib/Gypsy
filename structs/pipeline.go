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

type Run struct {
	ID           string `json:"id"`
	PipelineName string `json:"pipeline_name"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	Status       bool   `json:"status"`
}
