// Ported from flying-saucer-pdf/src/main/java/org/xhtmlrenderer/pdf/PagePosition.java

package pdf

type PagePosition struct {
	id     string
	pageNo int
	x      float32
	width  float32
	y      float32
	height float32
}

func NewPagePosition(id string, pageNo int, x float32, width float32, y float32, height float32) *PagePosition {
	return &PagePosition{
		id:     id,
		pageNo: pageNo,
		x:      x,
		width:  width,
		y:      y,
		height: height,
	}
}

func (p *PagePosition) GetPageNo() int {
	return p.pageNo
}

func (p *PagePosition) GetX() float32 {
	return p.x
}

func (p *PagePosition) GetWidth() float32 {
	return p.width
}

func (p *PagePosition) GetY() float32 {
	return p.y
}

func (p *PagePosition) GetHeight() float32 {
	return p.height
}

func (p *PagePosition) GetId() string {
	return p.id
}
