package main

import (
	prototype "github.com/aoldershaw/prototype-sdk-go"
)

func Prototype() prototype.Prototype {
	return prototype.New(
		prototype.WithObject(Repository{},
			prototype.WithMessage("list", (Repository).ListBranches, prototype.EmptyConfig),
		),
		prototype.WithObject(Branch{},
			prototype.WithMessage("list", (Branch).ListCommits, prototype.EmptyConfig),
			prototype.WithMessage("put", (Branch).Push, (Branch).PushConfig),
		),
		prototype.WithObject(Commit{}),
	)
}

func main() {
	if err := Prototype().Run(); err != nil {
		panic(err)
	}
}
