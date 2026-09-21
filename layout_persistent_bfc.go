// Ported from flying-saucer-core/src/main/java/org/xhtmlrenderer/layout/PersistentBFC.java

package ufo

// PersistentBFC is the portion of a BlockFormattingContext which is saved
// with a box which defines a BFC.
//
// XXX This class can go away
type PersistentBFC struct {
	floatManager *FloatManager
}

func NewPersistentBFC(master BlockBoxI, c *LayoutContext) *PersistentBFC {
	p := &PersistentBFC{}
	master.SetPersistentBFC(p)
	p.floatManager = NewFloatManager(master)
	return p
}

func (p *PersistentBFC) GetFloatManager() *FloatManager {
	return p.floatManager
}
