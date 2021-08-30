package node

// node.Type defines a type for Nodes to be differentiated.
type Type uint8

const (
	// node.Full is a full-featured Celestia Node.
	Full Type = iota + 1
	// node.Light is a stripped-down Celestia Node which aims to be lightweight while preserving highest possible
	// security guarantees.
	Light
)

// String converts Type to its string representation.
func (t Type) String() string {
	if !t.IsValid() {
		return "unknown"
	}
	return typeToString[t]
}

// IsValid reports whenever the Type is valid.
func (t Type) IsValid() bool {
	_, ok := typeToString[t]
	return ok
}

// typeToString keeps string representations of all valid Types.
var typeToString = map[Type]string{
	Full:  "Full",
	Light: "Light",
}
