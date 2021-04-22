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
	)
}

func main() {
	if err := Prototype().Run(); err != nil {
		panic(err)
	}
}
