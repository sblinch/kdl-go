package relaxed

// Flags specifies flags for kdl-go's relaxed mode which violates the KDL spec but allows extra flexibility in
// parsing documents
type Flags int

const (
	// NGINXSyntax loosens KDL's grammar to allow nginx-style configuration syntax
	NGINXSyntax Flags = 1 << iota
	// YAMLTOMLAssignments loosens KDL's grammar to allow `=` and `:` between a node name and its first argument
	YAMLTOMLAssignments
	// MultiplierSuffixes allows bare numeric values to have a suffix indicating a multiplier; for unmarshaling
	// time.Duration values, this may include any suffix accepted by time.ParseDuration (such as `15s`); for other
	// numeric values, this may include [kKMgGtTpP]?[Bb]? (indicating kilo, mega, giga, tera, peta, respectively); a
	// single-character suffix such as `k` uses a decimal multiplier (so `32k` unmarshals as 32x1000=32000), whereas a
	// suffix followed by a `b` or `B` uses a binary multiplier (so `32kb` unmarshals as 32x1024=32768).
	MultiplierSuffixes
)

// Permit indicates whether a given flag is set
func (f Flags) Permit(q Flags) bool {
	return (f & q) != 0
}
