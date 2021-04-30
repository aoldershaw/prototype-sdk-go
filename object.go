package prototype

import (
	"fmt"
	"reflect"
)

type unsupportedMessageError struct {
	msg    string
	object Object
}

func (e unsupportedMessageError) Error() string {
	return fmt.Sprintf("message %q is not supported by object %+v", e.msg, e.object)
}

type Object interface{}
type Request interface{}

type objectWrapper struct {
	object   Object
	messages []message
}

type invokableMessage struct {
	msg     message
	object  Object
	request Request
}

func (i invokableMessage) invoke() ([]MessageResponse, error) {
	return i.msg.execute(i.object, i.request)
}

func (i invokableMessage) messageInfo() MessageInfo {
	return MessageInfo{Name: i.msg.name}
}

type ObjectOption func(*objectWrapper)

type message struct {
	name        string
	requestType reflect.Type
	execute     func(Object, Request) ([]MessageResponse, error)
}

func WithObject(object Object, options ...ObjectOption) Option {
	return func(p *Prototype) {
		wrapper := objectWrapper{object: object}
		for _, opt := range options {
			opt(&wrapper)
		}
		p.objects = append(p.objects, wrapper)
	}
}

// WithMessage takes the name of a message and an executeFunc to run upon the
// receipt of that message.
//
// executeFunc must have one of the following signatures:
//
// * func(ConcreteObject) []MessageResponse
// * func(ConcreteObject) ([]MessageResponse, error)
// * func(ConcreteObject, ConcreteRequest) []MessageResponse
// * func(ConcreteObject, ConcreteRequest) ([]MessageResponse, error)
//
// ...where ConcreteObject must match the Object the message is for, and
// ConcreteRequest may be any type.
func WithMessage(name string, executeFunc interface{}) ObjectOption {
	return func(o *objectWrapper) {
		objectType := reflect.TypeOf(o.object)

		execute, requestType, err := validateExecuteFunc(objectType, executeFunc)
		if err != nil {
			panic(err)
		}

		o.messages = append(o.messages, message{
			name:        name,
			requestType: requestType,
			execute:     execute,
		})
	}
}

func validateExecuteFunc(objectType reflect.Type, executeFunc interface{}) (func(Object, Request) ([]MessageResponse, error), reflect.Type, error) {
	rt := reflect.TypeOf(executeFunc)
	if (rt.NumIn() != 1 && rt.NumIn() != 2) ||
		!objectType.AssignableTo(rt.In(0)) {
		return nil, nil, fmt.Errorf("the function must have 1 or 2 arguments (%s, and optionally a request type)", objectType)
	}
	if (rt.NumOut() != 1 && rt.NumOut() != 2) ||
		!reflect.TypeOf([]MessageResponse(nil)).AssignableTo(rt.Out(0)) {
		return nil, nil, fmt.Errorf("the function must have 1 or 2 return types ([]prototype.MessageResponse, and optionally, error)")
	}
	var requestType reflect.Type
	if rt.NumIn() == 2 {
		requestType = rt.In(1)
	}

	return func(object Object, request Request) ([]MessageResponse, error) {
		var args []reflect.Value
		if rt.NumIn() == 1 {
			args = []reflect.Value{reflect.ValueOf(object)}
		} else {
			args = []reflect.Value{reflect.ValueOf(object), reflect.ValueOf(request)}
		}

		result := reflect.ValueOf(executeFunc).Call(args)

		response := result[0].Interface().([]MessageResponse)
		var err error
		if rt.NumOut() == 2 {
			err, _ = result[1].Interface().(error)
		}
		return response, err
	}, requestType, nil
}

func invoke(object objectWrapper, msg string, request interface{}) ([]MessageResponse, error) {
	for _, m := range object.messages {
		if m.name == msg {
			return m.execute(object.object, request)
		}
	}
	return nil, unsupportedMessageError{msg: msg, object: object.object}
}
