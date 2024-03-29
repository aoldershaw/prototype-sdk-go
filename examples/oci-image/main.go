package main

import (
	prototype "github.com/aoldershaw/prototype-sdk-go"
)

func Prototype() prototype.Prototype {
	return prototype.New(
		prototype.WithObject(OCIImage{},
			prototype.WithMessage("build", (OCIImage).Build),
			prototype.WithMessage("run-stage", (OCIImage).RunStage),
		),
		prototype.WithIcon("mdi:docker"),
	)
}

func main() {
	if err := Prototype().Execute(); err != nil {
		panic(err)
	}
}
