package blocks

type Ownership string

const (
	OwnershipManaged Ownership = "managed"
	OwnershipEdited  Ownership = "edited"
	OwnershipForeign Ownership = "foreign"
)
