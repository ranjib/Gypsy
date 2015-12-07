// Copyright 2015 Ranjib Dey.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Holds common data types for gypsy
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
	ID           int    `json:"id"`
	PipelineName string `json:"pipeline_name"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	Success      bool   `json:"success"`
}
