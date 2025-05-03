package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"log"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type Boundary struct {
	Left, Right, Bottom, Top float32
}

type PosNode struct {
	X, Y  float32
	Edges []int
}

type PosGraph []PosNode

var testgraph = PosGraph{
	{X: 0, Y: 0, Edges: []int{1, 2, 3}},
	{X: 1, Y: 0, Edges: []int{0, 3}},
	{X: 0, Y: 1, Edges: []int{0, 3}},
	{X: 1, Y: 1, Edges: []int{1, 2}},
}

var edgeColor = color.RGBA{0, 0, 0, 255}
var arrowColor = color.RGBA{32, 32, 255, 255}
var nodeColor = color.RGBA{0, 0, 255, 255}
var nodeRadius = 12

/***** Rendering subroutines *****/

/* Bresenham's line algorithm, translated from [1] */
func lineOctant0or3(img *image.RGBA, x1, y1, absdx, dy, xDirection int, color color.RGBA) {
	deltaYx2 := 2 * dy
	DeltaYx2MinusDeltaXx2 := deltaYx2 - 2*absdx
	// Error actually represents how far off we are from the top of the current pixel, where -2*absdx is the bottom,
	// and 0 is the top. So its misnamed, but the point is that algorithm only needs integer math.
	Error := -absdx + deltaYx2
	img.Set(x1, y1, color)
	x2 := x1 + xDirection*absdx
	for x1 != x2 {
		if Error > 0 {
			y1 += 1
			Error += DeltaYx2MinusDeltaXx2
		} else {
			Error += deltaYx2
		}
		x1 += xDirection
		img.Set(x1, y1, color)
	}
}

func lineOctant1or2(img *image.RGBA, x1, y1, absdx, dy, xDirection int, color color.RGBA) {
	DeltaXx2 := 2 * absdx
	DeltaXx2MinusDeltaYx2 := DeltaXx2 - 2*dy
	// Here -2*dy is the left, and 0 is the right(if we're moving to the right).
	Error := DeltaXx2 - dy
	img.Set(x1, y1, color)
	y2 := y1 + dy
	for y1 != y2 {
		if Error > 0 {
			x1 += xDirection
			Error += DeltaXx2MinusDeltaYx2
		} else {
			Error += DeltaXx2
		}
		y1 += 1
		img.Set(x1, y1, color)
	}
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 int, color color.RGBA) {
	if y1 > y2 {
		x1, x2, y1, y2 = x2, x1, y2, y1
	}
	dx := x2 - x1
	dy := y2 - y1
	if dx > 0 {
		if dx > dy {
			lineOctant0or3(img, x1, y1, dx, dy, 1, color)
		} else {
			lineOctant1or2(img, x1, y1, dx, dy, 1, color)
		}
	} else {
		absdx := -dx
		if absdx > dy {
			lineOctant0or3(img, x1, y1, absdx, dy, -1, color)
		} else {
			lineOctant1or2(img, x1, y1, absdx, dy, -1, color)
		}
	}
}

func round64(x float64) int {
	return int(math.Round(x))
}

func roundAndClamp(x float64, min, max int) int {
	xi := int(math.Round(x))
	if xi < min {
		return min
	} else if xi > max {
		return max
	} else {
		return xi
	}
}

type Mat2d = struct{ xx, xy, yx, yy float64 }

var arrowLeftRotMatrix = Mat2d{xx: -.866, xy: -.5, yx: .5, yy: -.866}
var arrowRightRotMatrix = Mat2d{xx: -.866, xy: .5, yx: -.5, yy: -.866}

func drawDirectedLine(img *image.RGBA, x1, y1, x2, y2 int, color color.RGBA, arrowcolor color.RGBA) {
	imgW, imgH := img.Bounds().Max.X, img.Bounds().Max.Y
	drawLine(img, x1, y1, x2, y2, color)
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	r := math.Sqrt(dx*dx + dy*dy)
	R := float64(20.0)
	rx := dx / r
	ry := dy / r
	tipX := float64(x2) - float64(nodeRadius)*rx
	tipY := float64(y2) - float64(nodeRadius)*ry
	// Apply a rotation matrix to find the location of the "arrowhead" points...
	// This won't draw the arrows quite right if they go off the screen: the proper thing
	// todo would be to compute the intersection with the boundary lines.
	arrowLeftX := roundAndClamp(tipX+arrowLeftRotMatrix.xx*R*rx+arrowLeftRotMatrix.xy*R*ry, 0, imgW-1)
	arrowLeftY := roundAndClamp(tipY+arrowLeftRotMatrix.yx*R*rx+arrowLeftRotMatrix.yy*R*ry, 0, imgH-1)
	drawLine(img, round64(tipX), round64(tipY), arrowLeftX, arrowLeftY, arrowcolor)

	arrowRightX := roundAndClamp(tipX+arrowRightRotMatrix.xx*R*rx+arrowRightRotMatrix.xy*R*ry, 0, imgW-1)
	arrowRightY := roundAndClamp(tipY+arrowRightRotMatrix.yx*R*rx+arrowRightRotMatrix.yy*R*ry, 0, imgH-1)
	drawLine(img, round64(tipX), round64(tipY), arrowRightX, arrowRightY, arrowcolor)
}

// filled in circle at (x, y) with radius r
func drawCircle(img *image.RGBA, x, y, r int, color color.RGBA) {
	// not quite Bresenham quality, Todo: improve or antialias
	for i := 0; i <= r; i++ {
		for j := 0; j <= r; j++ {
			if i*i+j*j <= r*r+1 {
				img.Set(x-i, y-j, color)
				img.Set(x-i, y+j, color)
				img.Set(x+i, y-j, color)
				img.Set(x+i, y+j, color)
			} else {
				break
			}
		}
	}
}

/***********/

func getBoundary(graph []PosNode) Boundary {
	boundary := Boundary{Left: 1000000, Right: -1000000, Bottom: 1000000, Top: -1000000}
	for _, node := range graph {
		if node.X < boundary.Left {
			boundary.Left = node.X
		}
		if node.X > boundary.Right {
			boundary.Right = node.X
		}
		if node.Y < boundary.Bottom {
			boundary.Bottom = node.Y
		}
		if node.Y > boundary.Top {
			boundary.Top = node.Y
		}
	}
	if (boundary.Top == boundary.Bottom) {
		boundary.Top += 1
		boundary.Bottom -= 1
	}
	if (boundary.Left == boundary.Right) {
		boundary.Left -= 1
		boundary.Right += 1
	}
	return boundary
}

func translateCoords(x, y float32, boundary Boundary, imgW, imgH int) (int, int) {
	xScale := float32(imgW-2*nodeRadius) / (boundary.Right - boundary.Left)
	yScale := float32(imgH-2*nodeRadius) / (boundary.Top - boundary.Bottom)

	xOffset := nodeRadius + int(xScale*(x-boundary.Left))
	yOffset := nodeRadius + int(yScale*(boundary.Top-y))

	// xx remove
	// assert(xOffset >= 0 && xOffset < imgW && yOffset >= 0 && yOffset < imgH, "out of bounds")
	return xOffset, yOffset
}

func drawEdges(img *image.RGBA, graph []PosNode, boundary Boundary, directed bool) {
	imgW, imgH := img.Bounds().Max.X, img.Bounds().Max.Y
	for _, node := range graph {
		for _, edge := range node.Edges {
			x1p, y1p := translateCoords(node.X, node.Y, boundary, imgW, imgH)
			x2p, y2p := translateCoords(graph[edge].X, graph[edge].Y, boundary, imgW, imgH)
			if directed {
				drawDirectedLine(img, x1p, y1p, x2p, y2p, edgeColor, arrowColor)
			} else {
				drawLine(img, x1p, y1p, x2p, y2p, edgeColor)
			}
		}
	}
}

func drawNodes(img *image.RGBA, graph PosGraph, boundary Boundary) {
	imgW, imgH := img.Bounds().Max.X, img.Bounds().Max.Y
	for _, node := range graph {
		xp, yp := translateCoords(node.X, node.Y, boundary, imgW, imgH)
		drawCircle(img, xp, yp, nodeRadius, nodeColor)
	}
}

/*
Todo: Rendering the graph can actually be done in parallel because we don't mind

	drawing things on top of each other as long as the edge phase and the node phase are
	separate.
*/
func drawGraph(img *image.RGBA, graph PosGraph, directed bool) {
	boundary := getBoundary(graph)
	drawEdges(img, graph, boundary, directed)
	drawNodes(img, graph, boundary)
}

func run(window *app.Window, graph PosGraph, directed bool) error {
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			img := image.NewRGBA(image.Rect(0, 0, e.Size.X, e.Size.Y))
			draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
			drawGraph(img, graph, directed)

			// see https://gioui.org/doc/architecture/drawing
			paint.NewImageOp(img).Add(&ops)
			paint.PaintOp{}.Add(&ops)
			e.Frame(&ops)
		}
	}
}

func RenderGUI(graph PosGraph, directed bool) {
	fmt.Println("Starting ui...")
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Graphs"))
		err := run(window, graph, directed)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func RenderPNG(graph PosGraph, directed bool) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 800))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawGraph(img, graph, directed)
	out, _ := os.Create("output.png")
	png.Encode(out, img)
	out.Close()
}

/* Refs:
   [1] https://www.phatcode.net/res/224/files/html/ch35/35-03.html
*/
