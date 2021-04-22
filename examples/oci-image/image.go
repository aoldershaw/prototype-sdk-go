package main

import (
	"fmt"
	"path/filepath"

	prototype "github.com/aoldershaw/prototype-sdk-go"
)

type OCIImage struct {
	Context        string            `json:"context" prototype:"required"`
	ContextInputs  map[string]string `json:"context_inputs,omitempty"`
	DockerfilePath string            `json:"dockerfile,omitempty"`
	Cache          bool              `json:"cache,omitempty"`
	Output         string            `json:"output" prototype:"required"`
}

func (o OCIImage) Icon() string { return "mdi:docker" }

func (o OCIImage) Build() ([]prototype.MessageResponse, error) {
	fmt.Println("building an image!", o.Context)
	return nil, nil
}

func (o OCIImage) Config() prototype.Config {
	var config prototype.Config

	config.Inputs = append(config.Inputs, prototype.Input{Name: o.Context})
	for name, path := range o.ContextInputs {
		config.Inputs = append(config.Inputs, prototype.Input{
			Name: name,
			Path: filepath.Join(o.Context, path),
		})
	}

	config.Outputs = []prototype.Output{{Name: o.Output}}

	if o.Cache {
		config.Caches = []prototype.Cache{{Path: "cache"}}
	}

	return config
}

type RunStageRequest struct {
	Stage string `json:"stage" prototype:"required"`
}

func (o OCIImage) RunStage(request RunStageRequest) ([]prototype.MessageResponse, error) {
	fmt.Println("running stage", request.Stage)
	return nil, nil
}
