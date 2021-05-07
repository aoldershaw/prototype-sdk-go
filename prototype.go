package prototype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

const InterfaceVersion = "1.0"

type Prototype struct {
	objects []objectWrapper
	Icon    string
}

type Option func(*Prototype)

func New(options ...Option) Prototype {
	p := Prototype{}
	for _, opt := range options {
		opt(&p)
	}
	return p
}

func WithIcon(icon string) Option {
	return func(p *Prototype) {
		p.Icon = icon
	}
}

func (p Prototype) Execute() error {
	var responsePath string
	var encode func(*json.Encoder) error
	if len(os.Args) > 1 {
		var request struct {
			MessageRequest
			ResponsePath string `json:"response_path"`
		}
		if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
			return fmt.Errorf("invalid json request: %w", err)
		}
		responsePath = request.ResponsePath

		message := os.Args[1]
		responses, err := p.Run(message, request.MessageRequest)
		if err != nil {
			return fmt.Errorf("run %q: %w", message, err)
		}
		encode = func(encoder *json.Encoder) error {
			for _, response := range responses {
				if err := encoder.Encode(response); err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		var request struct {
			InfoRequest
			ResponsePath string `json:"response_path"`
		}
		if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
			return fmt.Errorf("invalid json request: %w", err)
		}
		responsePath = request.ResponsePath

		response, err := p.Info(request.InfoRequest)
		if err != nil {
			return fmt.Errorf("info: %w", err)
		}
		encode = func(encoder *json.Encoder) error {
			return encoder.Encode(response)
		}
	}

	responseFile, err := os.OpenFile(responsePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open response file: %w", err)
	}
	defer responseFile.Close()

	encoder := json.NewEncoder(responseFile)
	if err := encode(encoder); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func (p Prototype) Run(message string, request MessageRequest) ([]MessageResponse, error) {
	invocations, err := decodePossibleInvocations(request.Object, p.objects, message)
	if err != nil {
		return nil, err
	}
	if len(invocations) == 0 {
		return nil, fmt.Errorf("no object satisfied payload")
	}
	if len(invocations) > 1 {
		var satisfiableTypes []reflect.Type
		for _, invocation := range invocations {
			satisfiableTypes = append(satisfiableTypes, reflect.TypeOf(invocation.object))
		}
		return nil, fmt.Errorf("object is ambiguous - satisfies types %v", satisfiableTypes)
	}

	responses, err := invocations[0].invoke()
	if err != nil {
		return nil, fmt.Errorf("invoke: %w", err)
	}
	return responses, nil
}

func (p Prototype) Info(request InfoRequest) (InfoResponse, error) {
	invocations, err := decodePossibleInvocations(request.Object, p.objects, "")
	if err != nil {
		return InfoResponse{}, err
	}

	messages := make([]string, len(invocations))
	for i, invocation := range invocations {
		messages[i] = invocation.name()
	}

	return InfoResponse{
		InterfaceVersion: InterfaceVersion,
		Icon:             p.Icon,
		Messages:         messages,
	}, nil
}

// InfoRequest is the payload written to stdin for the default CMD.
type InfoRequest struct {
	// The object to act on.
	Object map[string]interface{} `json:"object"`
}

// InfoResponse is the payload written to the `response_path` in response to an
// InfoRequest.
type InfoResponse struct {
	// The version of the prototype interface that this prototype conforms to.
	InterfaceVersion string `json:"interface_version"`

	// An optional icon to show to the user.
	//
	// Icons must be namespaced by in order to explicitly reference an icon set
	// supported by Concourse, e.g. 'mdi:' for Material Design Icons.
	Icon string `json:"icon,omitempty"`

	// The messages supported by the object.
	Messages []string `json:"messages"`
}

// MessageRequest is the payload written to stdin for a message.
type MessageRequest struct {
	// The object to act on.
	Object map[string]interface{} `json:"object"`
}

// MessageResponse is written to the `response_path` for each object returned
// by the message. Multiple responses may be written to the same file,
// concatenated as a JSON stream.
type MessageResponse struct {
	// The object to return. May contain literal data and/or artifacts (using
	// the Artifact type).
	Object Object `json:"object"`

	// Metadata to associate with the object. Shown to the user.
	Metadata []MetadataField `json:"metadata,omitempty"`
}

// Artifact is a relative path relative to the working directory. It cannot go
// up a directory (i.e. must be within the working directory or any child
// directories of the working directory). If emitted in the Object of the
// MessageResponse, Artifacts may be used in pipelines as inputs to other
// prototypes/tasks/resources.
type Artifact string

func (a Artifact) MarshalJSON() ([]byte, error) {
	path := filepath.Clean(string(a))
	return json.Marshal(map[string]string{"artifact": path})
}

func (a *Artifact) UnmarshalJSON(payload []byte) error {
	var dst struct {
		Artifact string `json:"artifact"`
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dst); err != nil {
		return err
	}
	*a = Artifact(dst.Artifact)
	return nil
}

// MetadataField represents a named bit of metadata associated to an object.
type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
