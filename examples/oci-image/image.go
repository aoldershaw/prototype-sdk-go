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

func (o OCIImage) Build() error {
	fmt.Println("building an image!", o.Context)
	return nil
}

func build(o prototypesdk.Object) ([]prototypesdk.MessageResponse, error) {
	return nil, o.(*OCIImage).Build()
}
