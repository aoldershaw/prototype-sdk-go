package prototypesdk

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/reflectwalk"
)

type requiredFieldNotSetError struct {
	name string
}

func (e requiredFieldNotSetError) Error() string {
	return fmt.Sprintf("prototypesdk: required field %q is unset", e.name)
}

func decodeObjectAndRequest(objectJSON map[string]json.RawMessage, objects []objectWrapper, msg string) (objectWrapper, interface{}, bool) {
	payload, err := json.Marshal(objectJSON)
	if err != nil {
		panic(err)
	}

	// Loop in reverse with the assumption that the most highly specified
	// objects appear later.
	// TODO: is there a better way?
	for i := len(objects) - 1; i >= 0; i-- {
		rt := reflect.TypeOf(objects[i].object)
		object := reflect.New(rt).Interface()
		err := decodeSingle(payload, object)
		if err != nil {
			// skip over when fail to decode object
			continue
		}
		// TODO: disallow using the same field in both object and request?
		// disallow unused fields (at least when msg != "")?
		var request interface{}
		if msg != "" {
			request, err = decodeRequest(payload, objects[i].messages, msg)
			if err != nil {
				// skip over when fail to decode request
				// TODO: should this actually fail instead?
				continue
			}
		}
		return objectWrapper{
			// TODO: lol how are you supposed to avoid getting a pointer to the type?
			object:   reflect.ValueOf(object).Elem().Interface().(Object),
			messages: objects[i].messages,
		}, request, true
	}
	return objectWrapper{}, nil, false
}

func decodeObject(objectJSON map[string]json.RawMessage, objects []objectWrapper) (objectWrapper, bool) {
	object, _, ok := decodeObjectAndRequest(objectJSON, objects, "")
	return object, ok
}

func decodeSingle(payload []byte, dst interface{}) error {
	if err := json.Unmarshal(payload, dst); err != nil {
		return err
	}

	return reflectwalk.Walk(dst, requiredTagWalker{})
}

func decodeRequest(payload []byte, messages []message, messageName string) (interface{}, error) {
	for _, msg := range messages {
		if msg.Name == messageName {
			if msg.RequestType == nil {
				// no request type for this message
				return nil, nil
			}
			req := reflect.New(msg.RequestType).Interface()
			err := decodeSingle(payload, req)
			return reflect.ValueOf(req).Elem().Interface(), err
		}
	}
	// unsupported message - will be handled later on
	return nil, nil
}

type requiredTagWalker struct{}

func (requiredTagWalker) Struct(_ reflect.Value) error { return nil }
func (requiredTagWalker) StructField(field reflect.StructField, rv reflect.Value) error {
	if field.Tag.Get("prototype") != "required" {
		return nil
	}
	if rv.IsZero() {
		return requiredFieldNotSetError{name: field.Name}
	}
	return nil
}
