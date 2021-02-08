package position

import "fmt"

// 1-based.
type Point struct {
	line   uint
	column uint
}

type Pos struct {
	filename string
	start    Point
	end      Point
}

func NewPoint(line, column uint) *Point {
	return &Point{line: line, column: column}
}

func (p *Point) String() string {
	return fmt.Sprintf("%v.%v", p.line, p.column)
}

func (p *Point) Leftmost() bool {
	return p.column == 1
}
