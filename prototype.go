package prototypesdk

import (
	"encoding/json"
	"fmt"
	"os"
)

const InterfaceVersion = "1.0"

type Prototype struct {
	objects []objectWrapper
}

type Option func(*Prototype)

func New(options ...Option) Prototype {
	p := Prototype{}
	for _, opt := range options {
		opt(&p)
	}
	return p
}

func (p Prototype) Run() error {
	if len(os.Args) > 1 {
		return p.invokeMessage(os.Args[1])
	} else {
		return p.runInfo()
	}
}

func (p Prototype) invokeMessage(msg string) error {
	var req MessageRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode message request: %w", err)
	}

	object, request, ok := decodeObjectAndRequest(req.Object, p.objects, msg)
	if !ok {
		return fmt.Errorf("no object satisfied payload")
	}

	response, err := invoke(object, msg, request)
	if err != nil {
		return fmt.Errorf("invoke message %q: %w", msg, err)
	}

	responseFile, err := os.OpenFile(req.ResponsePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open response file: %w", err)
	}
	defer responseFile.Close()

	if err := json.NewEncoder(responseFile).Encode(response); err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func (p Prototype) runInfo() error {
	var req InfoRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode message request: %w", err)
	}

	object, ok := decodeObject(req.Object, p.objects)
	if !ok {
		return fmt.Errorf("no object satisfied payload")
	}

	response := InfoResponse{
		InterfaceVersion: InterfaceVersion,
		Icon:             object.object.Icon(),
		Messages:         object.Messages(),
	}

	responseFile, err := os.OpenFile(req.ResponsePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open response file: %w", err)
	}
	defer responseFile.Close()

	if err := json.NewEncoder(responseFile).Encode(response); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

// InfoRequest is the payload written to stdin for the default CMD.
type InfoRequest struct {
	// The object to act on.
	Object map[string]json.RawMessage `json:"object"`

	// Path to a file into which the prototype must write its InfoResponse.
	ResponsePath string `json:"response_path"`
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
	Messages []MessageInfo `json:"messages,omitempty"`
}

// MessageRequest is the payload written to stdin for a message.
type MessageRequest struct {
	// The object to act on.
	Object map[string]json.RawMessage `json:"object"`

	// Path to a file into which the prototype must write its InfoResponse.
	ResponsePath string `json:"response_path"`
}

// MessageResponse is written to the `response_path` for each object returned
// by the message. Multiple responses may be written to the same file,
// concatenated as a JSON stream.
type MessageResponse struct {
	// The object.
	Object Object `json:"object"`

	// Metadata to associate with the object. Shown to the user.
	Metadata []MetadataField `json:"metadata,omitempty"`
}

// MetadataField represents a named bit of metadata associated to an object.
type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
