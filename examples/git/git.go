package main

import (
	"fmt"

	prototype "github.com/aoldershaw/prototype-sdk-go"
)

type Repository struct {
	URI        string `json:"uri" prototype:"required"`
	PrivateKey string `json:"private_key"`
}

type ListBranchesRequest struct {
	BranchFilter string `json:"branch_filter"`
}

func (r Repository) ListBranches(request ListBranchesRequest) ([]prototype.MessageResponse, error) {
	fmt.Println("listing branches...", request)
	return []prototype.MessageResponse{
		{Object: Branch{Repository: r, Branch: "master"}},
		{Object: Branch{Repository: r, Branch: "dev"}},
	}, nil
}

type Branch struct {
	Repository
	Branch string `json:"branch" prototype:"required"`
}

type ListCommitsRequest struct {
	Paths []string `json:"paths"`
}

func (b Branch) ListCommits(request ListCommitsRequest) ([]prototype.MessageResponse, error) {
	fmt.Println("listing commits in branch "+b.Branch+"...", request)
	return []prototype.MessageResponse{
		{Object: Commit{Branch: b, Ref: "abcdef"}},
		{Object: Commit{Branch: b, Ref: "ghijkl"}},
	}, nil
}

type PushRequest struct {
	Repository string `json:"repository" prototype:"required"`
}

func (b Branch) Push(request PushRequest) ([]prototype.MessageResponse, error) {
	fmt.Println("pushing a new commit...", request)
	return nil, nil
}

type Commit struct {
	Branch
	Ref string `json:"ref" prototype:"required"`
}
