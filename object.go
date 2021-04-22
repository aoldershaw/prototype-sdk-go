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

type Object interface {
	Icon() string
}

type objectWrapper struct {
	object   Object
	messages []message
}

func (o objectWrapper) Messages() []MessageInfo {
	var msgs []MessageInfo

	// TODO: need a prepare function to give the inputs/outputs/caches
	for _, msg := range o.messages {
		msgs = append(msgs, MessageInfo{Name: msg.name})
	}

	return msgs
}

type ObjectOption func(*objectWrapper)

type message struct {
	name        string
	requestType reflect.Type
	execute     func(Object, interface{}) ([]MessageResponse, error)
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
// executeFunc must have one of the following signatures:
//
// * func(ConcreteObject) ([]MessageResponse, error)
// * func(ConcreteObject, ConcreteRequest) ([]MessageResponse, error)
//
// ...where ConcreteObject must match the Object the message is for, and
// ConcreteRequest may be any type.
func WithMessage(name string, executeFunc interface{}) ObjectOption {
	return func(o *objectWrapper) {
		rt := reflect.TypeOf(executeFunc)
		if rt.NumIn() != 1 && rt.NumIn() != 2 {
			panic("the function must have 1 or 2 arguments")
		}
		//if rt.NumOut() != 2 ||
		//	!rt.Out(0).AssignableTo(reflect.TypeOf([]MessageResponse(nil))) ||
		//	!rt.Out(1).AssignableTo(reflect.TypeOf(error(nil))) {
		//	panic("the function must have 2 return values ([]MessageResponse, error)")
		//}
		if rt.NumOut() != 2 ||
			!rt.Out(0).AssignableTo(reflect.TypeOf([]MessageResponse(nil))) {
			panic("the function must have 2 return values ([]MessageResponse, error)")
		}
		if !rt.In(0).AssignableTo(reflect.TypeOf(o.object)) {
			panic("the first argument must be of type " + reflect.TypeOf(o.object).String())
		}
		var requestType reflect.Type
		if rt.NumIn() == 2 {
			requestType = rt.In(1)
		}

		execute := func(object Object, request interface{}) ([]MessageResponse, error) {
			var args []reflect.Value
			if rt.NumIn() == 1 {
				args = []reflect.Value{reflect.ValueOf(object)}
			} else {
				args = []reflect.Value{reflect.ValueOf(object), reflect.ValueOf(request)}
			}

			result := reflect.ValueOf(executeFunc).Call(args)

			response := result[0].Interface().([]MessageResponse)
			err, _ := result[1].Interface().(error)
			return response, err
		}

		o.messages = append(o.messages, message{name: name, requestType: requestType, execute: execute})
	}
}

type MessageInfo struct {
	Name string `json:"name"`
}

func invoke(object objectWrapper, msg string, request interface{}) ([]MessageResponse, error) {
	for _, m := range object.messages {
		if m.name == msg {
			return m.execute(object.object, request)
		}
	}
	return nil, unsupportedMessageError{msg: msg, object: object.object}
}
