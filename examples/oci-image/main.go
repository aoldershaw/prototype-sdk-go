package main

import (
	prototypesdk "github.com/aoldershaw/prototype-sdk-go"
)

func prototype() *prototypesdk.Prototype {
	return prototypesdk.New(
		prototypesdk.WithObject(OCIImage{}, prototypesdk.WithMessage("build", build)),
	)
}

func main() {
	if err := prototype().Run(); err != nil {
		panic(err)
	}
}
