package prototypesdk

import "fmt"

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
		msgs = append(msgs, MessageInfo{Name: msg.Name})
	}

	return msgs
}

type ObjectOption func(*objectWrapper)

type message struct {
	Name    string
	Execute func(Object) ([]MessageResponse, error)
}

func WithObject(object Object, options ...ObjectOption) Option {
	return func(p *Prototype) {
		wrapper := objectWrapper{object: object}
		for _, opt := range options {
			opt(&wrapper)
		}
		p.Objects = append(p.Objects, wrapper)
	}
}

func WithMessage(name string, executeFn func(Object) ([]MessageResponse, error)) ObjectOption {
	return func(o *objectWrapper) {
		o.messages = append(o.messages, message{Name: name, Execute: executeFn})
	}
}

type MessageInfo struct {
	Name string `json:"name"`
}

func invoke(object objectWrapper, msg string) ([]MessageResponse, error) {
	for _, m := range object.messages {
		if m.Name == msg {
			return m.Execute(object.object)
		}
	}
	return nil, unsupportedMessageError{msg: msg, object: object.object}
}
