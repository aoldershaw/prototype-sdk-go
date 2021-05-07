package main

import (
	"fmt"

	prototype "github.com/aoldershaw/prototype-sdk-go"
)

type OCIImage struct {
	Context        string            `json:"context" prototype:"required"`
	ContextInputs  map[string]string `json:"context_inputs,omitempty"`
	DockerfilePath string            `json:"dockerfile,omitempty"`
}

func (o OCIImage) Build() ([]prototype.MessageResponse, error) {
	fmt.Println("building an image!", o.Context)
	return []prototype.MessageResponse{{
		Object: map[string]interface{}{
			"image": prototype.Artifact("./image"),
		}},
	}, nil
}

type RunStageRequest struct {
	Stage string `json:"stage" prototype:"required"`
}

func (o OCIImage) RunStage(request RunStageRequest) ([]prototype.MessageResponse, error) {
	fmt.Println("running stage", request.Stage)
	return nil, nil
}
