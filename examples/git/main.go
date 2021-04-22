package main

import (
	prototypesdk "github.com/aoldershaw/prototype-sdk-go"
)

func prototype() prototypesdk.Prototype {
	return prototypesdk.New(
		prototypesdk.WithObject(Repository{},
			prototypesdk.WithMessage("list", (Repository).ListBranches),
		),
		prototypesdk.WithObject(Branch{},
			prototypesdk.WithMessage("list", (Branch).ListCommits),
			prototypesdk.WithMessage("put", (Branch).Push),
		),
		prototypesdk.WithObject(Commit{}),
	)
}

func main() {
	if err := prototype().Run(); err != nil {
		panic(err)
	}
}
