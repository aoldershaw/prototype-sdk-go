package main

import (
	"fmt"

	prototypesdk "github.com/aoldershaw/prototype-sdk-go"
)

type OCIImage struct {
	Context        string            `json:"context" prototype:"required"`
	ContextInputs  map[string]string `json:"context_inputs,omitempty"`
	DockerfilePath string            `json:"dockerfile,omitempty"`
	Output         string            `json:"output" prototype:"required"`
}

func (o OCIImage) Icon() string { return "mdi:docker" }

func (o OCIImage) Build() ([]prototypesdk.MessageResponse, error) {
	fmt.Println("building an image!", o.Context)
	return nil, nil
}

type RunStageRequest struct {
	Stage string `json:"stage" prototype:"required"`
}

func (o OCIImage) RunStage(request RunStageRequest) ([]prototypesdk.MessageResponse, error) {
	fmt.Println("running stage", request.Stage)
	return nil, nil
}
