package geom

import (
	"math"
	"strconv"
)

// Point is java.awt.Point: a location with integer coordinates.
type Point struct {
	X int
	Y int
}

// NewPoint is new Point(x, y).
func NewPoint(x, y int) *Point {
	return &Point{X: x, Y: y}
}

// NewPointFromPoint is new Point(Point).
func NewPointFromPoint(p *Point) *Point {
	return &Point{X: p.X, Y: p.Y}
}

func (p *Point) GetX() float64 {
	return float64(p.X)
}

func (p *Point) GetY() float64 {
	return float64(p.Y)
}

// GetLocation returns a copy of the point.
func (p *Point) GetLocation() *Point {
	return &Point{X: p.X, Y: p.Y}
}

func (p *Point) SetLocation(x, y int) {
	p.X = x
	p.Y = y
}

// SetLocationPoint is setLocation(Point).
func (p *Point) SetLocationPoint(o *Point) {
	p.X = o.X
	p.Y = o.Y
}

// SetLocationDouble is setLocation(double, double), which rounds each
// coordinate with floor(v + 0.5).
func (p *Point) SetLocationDouble(x, y float64) {
	p.X = javaInt(math.Floor(x + 0.5))
	p.Y = javaInt(math.Floor(y + 0.5))
}

// Move is the same as SetLocation.
func (p *Point) Move(x, y int) {
	p.X = x
	p.Y = y
}

func (p *Point) Translate(dx, dy int) {
	p.X += dx
	p.Y += dy
}

// Distance returns the distance from this point to (px, py).
func (p *Point) Distance(px, py float64) float64 {
	px -= p.GetX()
	py -= p.GetY()
	return math.Sqrt(px*px + py*py)
}

// ToPoint2D returns the point as a Point2D. In the JDK a Point is a Point2D.
func (p *Point) ToPoint2D() *Point2D {
	return &Point2D{X: float64(p.X), Y: float64(p.Y)}
}

func (p *Point) Equals(o *Point) bool {
	return o != nil && p.X == o.X && p.Y == o.Y
}

func (p *Point) String() string {
	return "java.awt.Point[x=" + strconv.Itoa(p.X) + ",y=" + strconv.Itoa(p.Y) + "]"
}

// Point2D is java.awt.geom.Point2D: a location with floating point
// coordinates. It covers both Point2D.Float and Point2D.Double.
type Point2D struct {
	X float64
	Y float64
}

// NewPoint2DDouble is new Point2D.Double(x, y).
func NewPoint2DDouble(x, y float64) *Point2D {
	return &Point2D{X: x, Y: y}
}

// NewPoint2DFloat is new Point2D.Float(x, y).
func NewPoint2DFloat(x, y float32) *Point2D {
	return &Point2D{X: float64(x), Y: float64(y)}
}

func (p *Point2D) GetX() float64 {
	return p.X
}

func (p *Point2D) GetY() float64 {
	return p.Y
}

func (p *Point2D) SetLocation(x, y float64) {
	p.X = x
	p.Y = y
}

// SetLocationPoint2D is setLocation(Point2D).
func (p *Point2D) SetLocationPoint2D(o *Point2D) {
	p.X = o.X
	p.Y = o.Y
}

// Distance returns the distance from this point to (px, py).
func (p *Point2D) Distance(px, py float64) float64 {
	px -= p.X
	py -= p.Y
	return math.Sqrt(px*px + py*py)
}

// DistancePoint2D is distance(Point2D).
func (p *Point2D) DistancePoint2D(o *Point2D) float64 {
	return p.Distance(o.X, o.Y)
}

// DistanceSq returns the square of the distance from this point to (px, py).
func (p *Point2D) DistanceSq(px, py float64) float64 {
	px -= p.X
	py -= p.Y
	return px*px + py*py
}

func (p *Point2D) Clone() *Point2D {
	return &Point2D{X: p.X, Y: p.Y}
}

func (p *Point2D) Equals(o *Point2D) bool {
	return o != nil && p.X == o.X && p.Y == o.Y
}

func (p *Point2D) String() string {
	return "Point2D.Double[" + javaDoubleString(p.X) + ", " + javaDoubleString(p.Y) + "]"
}

// Dimension is java.awt.Dimension: a width and a height in integers.
type Dimension struct {
	Width  int
	Height int
}

// NewDimension is new Dimension(width, height).
func NewDimension(width, height int) *Dimension {
	return &Dimension{Width: width, Height: height}
}

// NewDimensionFromDimension is new Dimension(Dimension).
func NewDimensionFromDimension(d *Dimension) *Dimension {
	return &Dimension{Width: d.Width, Height: d.Height}
}

func (d *Dimension) GetWidth() float64 {
	return float64(d.Width)
}

func (d *Dimension) GetHeight() float64 {
	return float64(d.Height)
}

func (d *Dimension) SetSize(width, height int) {
	d.Width = width
	d.Height = height
}

// SetSizeDimension is setSize(Dimension).
func (d *Dimension) SetSizeDimension(o *Dimension) {
	d.Width = o.Width
	d.Height = o.Height
}

// SetSizeDouble is setSize(double, double), which rounds each value up.
func (d *Dimension) SetSizeDouble(width, height float64) {
	d.Width = javaInt(math.Ceil(width))
	d.Height = javaInt(math.Ceil(height))
}

// GetSize returns a copy of the dimension.
func (d *Dimension) GetSize() *Dimension {
	return &Dimension{Width: d.Width, Height: d.Height}
}

func (d *Dimension) Equals(o *Dimension) bool {
	return o != nil && d.Width == o.Width && d.Height == o.Height
}

func (d *Dimension) String() string {
	return "java.awt.Dimension[width=" + strconv.Itoa(d.Width) + ",height=" + strconv.Itoa(d.Height) + "]"
}

// Insets is java.awt.Insets: the space a container leaves at each of its
// edges.
type Insets struct {
	Top    int
	Left   int
	Bottom int
	Right  int
}

// NewInsets is new Insets(top, left, bottom, right).
func NewInsets(top, left, bottom, right int) *Insets {
	return &Insets{Top: top, Left: left, Bottom: bottom, Right: right}
}

func (i *Insets) Set(top, left, bottom, right int) {
	i.Top = top
	i.Left = left
	i.Bottom = bottom
	i.Right = right
}

func (i *Insets) Clone() *Insets {
	return &Insets{Top: i.Top, Left: i.Left, Bottom: i.Bottom, Right: i.Right}
}

func (i *Insets) Equals(o *Insets) bool {
	return o != nil && i.Top == o.Top && i.Left == o.Left && i.Bottom == o.Bottom && i.Right == o.Right
}

func (i *Insets) String() string {
	return "java.awt.Insets[top=" + strconv.Itoa(i.Top) + ",left=" + strconv.Itoa(i.Left) +
		",bottom=" + strconv.Itoa(i.Bottom) + ",right=" + strconv.Itoa(i.Right) + "]"
}
