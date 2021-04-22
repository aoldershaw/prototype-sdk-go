package main

import (
	prototype "github.com/aoldershaw/prototype-sdk-go"
)

func Prototype() prototype.Prototype {
	return prototype.New(
		prototype.WithObject(OCIImage{},
			prototype.WithMessage("build", (OCIImage).Build, (OCIImage).Config),
			prototype.WithMessage("run-stage", (OCIImage).RunStage, (OCIImage).Config),
		),
	)
}

func main() {
	if err := Prototype().Run(); err != nil {
		panic(err)
	}
}
