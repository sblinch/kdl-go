package document

// Document is the top-level container for a KDL document
type Document struct {
	Nodes []*Node
}

// AddNode adds a Node to this document
func (d *Document) AddNode(child *Node) {
	d.Nodes = append(d.Nodes, child)
}

// New cerates a new Document
func New() *Document {
	return &Document{
		Nodes: make([]*Node, 0, 32),
	}
}
