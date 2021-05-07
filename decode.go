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

func decodePossibleInvocations(object map[string]interface{}, objects []objectWrapper, messageName string) ([]invokableMessage, error) {
	fullObjectJSON, payload, err := rawJSONObject(object)
	if err != nil {
		return nil, fmt.Errorf("re-marshal object: %w", err)
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
		subObjectJSON, _, err := rawJSONObject(object)
		if err != nil {
			return nil, fmt.Errorf("invalid sub-object: %w", err)
		}
		jsonWithoutObject := jsonDiff(fullObjectJSON, subObjectJSON)
		payloadWithoutObject, err := json.Marshal(jsonWithoutObject)
		if err != nil {
			return nil, fmt.Errorf("re-marshal sub-object: %w", err)
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
			requestJSON, _, err := rawJSONObject(request)
			if err != nil {
				return nil, fmt.Errorf("invalid request object: %w", err)
			}
			leftoverJSON := jsonDiff(jsonWithoutObject, requestJSON)
			if !isJSONObjectEmpty(leftoverJSON) {
				// skip over when there are unused entries in the JSON
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
	return invokableMessages, nil
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
	if err != nil {
		return nil, err
	}
	return dereference(req), nil
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

func rawJSONObject(obj interface{}) (map[string]json.RawMessage, []byte, error) {
	objPayload, err := json.Marshal(obj)
	if err != nil {
		return nil, nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(objPayload, &raw); err != nil {
		return nil, nil, err
	}
	return raw, objPayload, nil
}

func jsonDiff(full, subtractKeys map[string]json.RawMessage) map[string]json.RawMessage {
	diff := map[string]json.RawMessage{}
	for k, v := range full {
		if _, ok := subtractKeys[k]; !ok {
			diff[k] = v
		}
	}
	return diff
}

func isJSONObjectEmpty(obj map[string]json.RawMessage) bool {
	for _, v := range obj {
		var dst interface{}
		if err := json.Unmarshal([]byte(v), &dst); err != nil {
			return false
		}
		if !reflect.ValueOf(dst).IsZero() {
			return false
		}
	}
	return true
}

func dereference(i interface{}) interface{} {
	// TODO: lol how are you supposed to avoid getting a pointer to the type?
	return reflect.ValueOf(i).Elem().Interface()
}
