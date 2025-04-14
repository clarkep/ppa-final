package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
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

type PosGraph []PosNode;

var testgraph = PosGraph {
	{X: 0, Y: 0, Edges: []int{1, 2, 3}},
	{X: 1, Y: 0, Edges: []int{0, 3}},
	{X: 0, Y: 1, Edges: []int{0, 3}},
	{X: 1, Y: 1, Edges: []int{1, 2}},
}

var edgeColor = color.RGBA{0, 0, 0, 255}
var nodeColor = color.RGBA{0, 0, 255, 255}
var nodeRadius = 12

/***** Rendering subroutines *****/

/* Bresenham's line algorithm, translated from [1] */
func lineOctant0or3(img *image.RGBA, x1, y1, absdx, dy, xDirection int, color color.RGBA) {
	deltaYx2 := 2 * dy
	DeltaYx2MinusDeltaXx2 := deltaYx2 - 2*absdx
	// Error actually represents how far off we are from the top of the current pixel, where -2*absdx is the bottom,
	// and 0 is the top. So its misnamed, but the point is that algorithm only needs integer math. -Paul
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
				break;
			}
		}
	}
}

/***********/

func getBoundary(graph []PosNode) Boundary {
	boundary := Boundary{}
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

func drawEdges(img *image.RGBA, graph []PosNode, boundary Boundary) {
	imgW, imgH := img.Bounds().Max.X, img.Bounds().Max.Y
	for _, node := range graph {
		for _, edge := range node.Edges {
			x1p, y1p := translateCoords(node.X, node.Y, boundary, imgW, imgH)
			x2p, y2p := translateCoords(graph[edge].X, graph[edge].Y, boundary, imgW, imgH)
			drawLine(img, x1p, y1p, x2p, y2p, edgeColor)
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

/* Todo: Rendering the graph can actually be done in parallel because we don't mind
   drawing things on top of each other as long as the edge phase and the node phase are
   separate. */
func drawGraph(img *image.RGBA, graph PosGraph) {
	boundary := getBoundary(graph)
	drawEdges(img, graph, boundary)
	drawNodes(img, graph, boundary)
}

func run(window *app.Window, graph PosGraph) error {
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			img := image.NewRGBA(image.Rect(0, 0, e.Size.X, e.Size.Y))
			draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
			drawGraph(img, graph)

			// see https://gioui.org/doc/architecture/drawing
			paint.NewImageOp(img).Add(&ops)
			paint.PaintOp{}.Add(&ops)
			e.Frame(&ops)
		}
	}
}

func RenderGUI (graph PosGraph) {
	fmt.Println("Starting ui...")
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Graphs"))
		err := run(window, graph)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func RenderPNG (graph PosGraph) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 800))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	drawGraph(img, graph)
	out, _ := os.Create("output.png")
	png.Encode(out, img)
	out.Close()
}

/* Refs:
   [1] https://www.phatcode.net/res/224/files/html/ch35/35-03.html
*/
