package blocks

import "github.com/ronz204/beacon/source/kernel"

type ManagedCode string

const (
	ManagedFrontmatterMissing ManagedCode = "managed_frontmatter_missing"
)

type ManagedError = kernel.Error[ManagedCode]

var managedErrors = kernel.ErrorCatalog[ManagedCode]{
	ManagedFrontmatterMissing: "a managed file must open with a " +
		"frontmatter fenced by two --- lines",
}

func failManagedFrontmatter() ManagedError {
	return managedErrors.Fail(ManagedFrontmatterMissing, nil)
}
