package main

import (
	prototypesdk "github.com/aoldershaw/prototype-sdk-go"
)

func prototype() prototypesdk.Prototype {
	return prototypesdk.New(
		prototypesdk.WithObject(OCIImage{},
			prototypesdk.WithMessage("build", (OCIImage).Build),
			prototypesdk.WithMessage("run-stage", (OCIImage).RunStage),
		),
	)
}

func main() {
	if err := prototype().Run(); err != nil {
		panic(err)
	}
}
