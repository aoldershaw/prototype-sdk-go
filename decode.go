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

func decodeObject(payload []byte, objects []objectWrapper) (objectWrapper, bool) {
	// Loop in reverse with the assumption that the most highly specified
	// objects appear later.
	// TODO: is there a better way?
	for i := len(objects) - 1; i >= 0; i-- {
		rt := reflect.TypeOf(objects[i].object)
		// TODO: proper pointer following?
		object := reflect.New(rt).Interface()
		err := decodeSingleObject(payload, object)
		if err == nil {
			return objectWrapper{
				object:   object.(Object),
				messages: objects[i].messages,
			}, true
		}
	}
	return objectWrapper{}, false
}

func decodeSingleObject(payload []byte, dst interface{}) error {
	if err := json.Unmarshal(payload, dst); err != nil {
		return err
	}

	return reflectwalk.Walk(dst, requiredTagWalker{})
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
