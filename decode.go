package prototype

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
	return fmt.Sprintf("prototype: required field %q is unset", e.name)
}

func decodePossibleInvocations(objectJSON map[string]json.RawMessage, objects []objectWrapper, messageName string) []invokableMessage {
	payload, err := json.Marshal(objectJSON)
	if err != nil {
		panic(err)
	}

	var invokableMessages []invokableMessage
	for _, wrapper := range objects {
		rt := reflect.TypeOf(wrapper.object)
		object := reflect.New(rt).Interface()
		err := decodeSingle(payload, object)
		if err != nil {
			// skip over when fail to decode object
			continue
		}
		jsonWithoutObject := jsonDiff(objectJSON, object)
		payloadWithoutObject, err := json.Marshal(jsonWithoutObject)
		if err != nil {
			panic(fmt.Sprintf("marshal JSON: %v", err))
		}
		for _, msg := range wrapper.messages {
			if messageName != "" && msg.name != messageName {
				// we are invoking a specific message, and it doesn't match the current message, so skip
				continue
			}
			request, err := decodeRequest(payloadWithoutObject, msg)
			if err != nil {
				// skip over when fail to decode request
				continue
			}
			leftoverJSON := jsonDiff(jsonWithoutObject, request)
			if len(leftoverJSON) > 0 {
				// skip over when there are unused keys
				// TODO: is this what we want for the info endpoint? when used
				// with the run step, it's fine, since it'll have the request
				// as well - but not sure how else concourse will use it
				continue
			}
			invokableMessages = append(invokableMessages, invokableMessage{
				msg:     msg,
				object:  dereference(object).(Object),
				request: request,
			})
		}
	}
	return invokableMessages
}

func decodeSingle(payload []byte, dst interface{}) error {
	if err := json.Unmarshal(payload, dst); err != nil {
		return err
	}

	return reflectwalk.Walk(dst, requiredTagWalker{})
}

func decodeRequest(payload []byte, message message) (interface{}, error) {
	if message.requestType == nil {
		// no request type for this message
		return nil, nil
	}
	req := reflect.New(message.requestType).Interface()
	err := decodeSingle(payload, req)
	return dereference(req), err
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

func jsonDiff(full map[string]json.RawMessage, obj interface{}) map[string]json.RawMessage {
	objPayload, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	var subtractKeys map[string]json.RawMessage
	if err := json.Unmarshal(objPayload, &subtractKeys); err != nil {
		panic(err)
	}
	diff := map[string]json.RawMessage{}
	for k, v := range full {
		if _, ok := subtractKeys[k]; !ok {
			diff[k] = v
		}
	}
	return diff
}

func dereference(i interface{}) interface{} {
	// TODO: lol how are you supposed to avoid getting a pointer to the type?
	return reflect.ValueOf(i).Elem().Interface()
}
