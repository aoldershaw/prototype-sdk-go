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

	invocations := decodePossibleInvocations(req.Object, p.objects, msg)
	if len(invocations) == 0 {
		return fmt.Errorf("no object satisfied payload")
	}
	if len(invocations) > 1 {
		var satisfiableTypes []reflect.Type
		for _, invocation := range invocations {
			satisfiableTypes = append(satisfiableTypes, reflect.TypeOf(invocation.object))
		}
		return fmt.Errorf("object is ambiguous - satisfies types %v", satisfiableTypes)
	}

	responses, err := invocations[0].invoke()
	if err != nil {
		return fmt.Errorf("invoke message %q: %w", msg, err)
	}

	responseFile, err := os.OpenFile(req.ResponsePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open response file: %w", err)
	}
	defer responseFile.Close()

	encoder := json.NewEncoder(responseFile)
	for _, response := range responses {
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
	return nil
}

func (p Prototype) runInfo() error {
	var req InfoRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return fmt.Errorf("decode message request: %w", err)
	}

	invocations := decodePossibleInvocations(req.Object, p.objects, "")

	messages := make([]MessageInfo, len(invocations))
	for i, invocation := range invocations {
		messages[i] = invocation.messageInfo()
	}

	response := InfoResponse{
		InterfaceVersion: InterfaceVersion,
		Icon:             p.Icon,
		Messages:         messages,
	}

	responseFile, err := os.OpenFile(req.ResponsePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
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
	Messages []MessageInfo `json:"messages"`
}

type MessageInfo struct {
	Name string `json:"name"`
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
