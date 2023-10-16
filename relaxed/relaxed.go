package relaxed

// Flags specifies flags for kdl-go's relaxed mode which violates the KDL spec but allows extra flexibility in
// parsing documents
type Flags int

const (
	// NGINXSyntax loosens KDL's grammar to allow nginx-style configuration syntax
	NGINXSyntax Flags = 1 << iota
	// YAMLTOMLAssignments loosens KDL's grammar to allow `=` and `:` between a node name and its first argument
	YAMLTOMLAssignments
	MultiplierSuffixes
)

// Permit indicates whether a given flag is set
func (f Flags) Permit(q Flags) bool {
	return (f & q) != 0
}
