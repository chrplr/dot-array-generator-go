// dot-array-gen: Generate dot array stimuli for numerical cognition research.
//
// Usage: dot-array-gen [flags]
// Run with -help to see all flags.

// This is a port to [Go](https://go.dev) of Lauren S. Aulet's [dot-array-generator](https://github.com/laurenaulet/dot-array-stimulus-toolbox) described in her PsyArXiv preprint: 
// [Dot Array Stimulus Toolbox: An Open-Source Solution for Generating and Analyzing Non-Symbolic Number Stimuli](https://osf.io/preprints/psyarxiv/uhsv6_v1)
//
// The port was performed using [Claude Code](https://code.claude.com/docs/en/overview) by [Christophe Pallier](http://www.pallier.org) on March 17, 2026
//
// Distributed under an MIT License


package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------
// Data types
// ---------------------------------------------------------------------------

type dot struct {
	x, y, radius float64
}

func (d dot) area() float64 { return math.Pi * d.radius * d.radius }

type groundTruth struct {
	filename                   string
	number                     int
	cumulativeArea             float64
	averageElementSize         float64
	sizeSD                     float64
	minElementSize             float64
	maxElementSize             float64
	totalContourLength         float64
	convexHullArea             float64
	convexHullPerimeter        float64
	fieldArea                  int
	density                    float64
	occupancy                  float64
	avgNearestNeighborDistance float64
	imageWidth                 int
	imageHeight                int
}

// ---------------------------------------------------------------------------
// Radius generation
// ---------------------------------------------------------------------------

func generateRadii(n int, avgRadius, sizeVariability, minRadius float64,
	controlArea bool, targetArea float64, rng *rand.Rand) []float64 {

	radii := make([]float64, n)
	if sizeVariability <= 0 {
		for i := range radii {
			radii[i] = avgRadius
		}
	} else {
		for i := range radii {
			r := rng.NormFloat64()*sizeVariability + avgRadius
			if r < minRadius {
				r = minRadius
			}
			radii[i] = r
		}
	}

	if controlArea && targetArea > 0 {
		currentArea := 0.0
		for _, r := range radii {
			currentArea += math.Pi * r * r
		}
		if currentArea > 0 {
			scale := math.Sqrt(targetArea / currentArea)
			for i, r := range radii {
				if r*scale < minRadius {
					radii[i] = minRadius
				} else {
					radii[i] = r * scale
				}
			}
		}
	}
	return radii
}

// ---------------------------------------------------------------------------
// Dot placement
// ---------------------------------------------------------------------------

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func placeDots(n int, radii []float64, width, height, margin int,
	minSpacing float64, rng *rand.Rand) []dot {

	dots := make([]dot, 0, n)
	const maxAttempts = 1000

	for i := 0; i < n; i++ {
		r := radii[i]
		xMin := float64(margin) + r
		xMax := float64(width) - float64(margin) - r
		yMin := float64(margin) + r
		yMax := float64(height) - float64(margin) - r

		if xMin > xMax {
			xMin, xMax = float64(width)/2, float64(width)/2
		}
		if yMin > yMax {
			yMin, yMax = float64(height)/2, float64(height)/2
		}

		placed := false

		// Pass 1: respect min spacing
		for attempt := 0; attempt < maxAttempts; attempt++ {
			x := rng.Float64()*(xMax-xMin) + xMin
			y := rng.Float64()*(yMax-yMin) + yMin
			if noOverlap(dots, x, y, r, minSpacing) {
				dots = append(dots, dot{x, y, r})
				placed = true
				break
			}
		}

		// Pass 2: allow touching but not overlapping
		if !placed {
			for attempt := 0; attempt < maxAttempts; attempt++ {
				x := rng.Float64()*(xMax-xMin) + xMin
				y := rng.Float64()*(yMax-yMin) + yMin
				if noOverlap(dots, x, y, r, 0) {
					dots = append(dots, dot{x, y, r})
					placed = true
					break
				}
			}
		}

		// Last resort: place anywhere
		if !placed {
			x := rng.Float64()*(xMax-xMin) + xMin
			y := rng.Float64()*(yMax-yMin) + yMin
			dots = append(dots, dot{x, y, r})
		}
	}
	return dots
}

func noOverlap(dots []dot, x, y, r, minSpacing float64) bool {
	for _, d := range dots {
		dx := x - d.x
		dy := y - d.y
		if math.Sqrt(dx*dx+dy*dy) < r+d.radius+minSpacing {
			return false
		}
	}
	return true
}

func placeDotsHullControlled(n int, radii []float64, width, height, margin int,
	minSpacing, targetHullArea float64, rng *rand.Rand) []dot {

	targetR := math.Sqrt(targetHullArea / math.Pi)
	cx := float64(width) / 2
	cy := float64(height) / 2

	var bestDots []dot
	bestDiff := math.Inf(1)

	for trial := 0; trial < 50; trial++ {
		dots := make([]dot, 0, n)

		for i := 0; i < n; i++ {
			r := radii[i]
			xMin := float64(margin) + r
			xMax := float64(width) - float64(margin) - r
			yMin := float64(margin) + r
			yMax := float64(height) - float64(margin) - r
			placed := false

			for attempt := 0; attempt < 100; attempt++ {
				angle := rng.Float64() * 2 * math.Pi
				dist := rng.Float64() * targetR
				x := clamp(cx+dist*math.Cos(angle), xMin, xMax)
				y := clamp(cy+dist*math.Sin(angle), yMin, yMax)
				if noOverlap(dots, x, y, r, minSpacing) {
					dots = append(dots, dot{x, y, r})
					placed = true
					break
				}
			}
			if !placed {
				x := clamp(cx+rng.Float64()*2*targetR-targetR, xMin, xMax)
				y := clamp(cy+rng.Float64()*2*targetR-targetR, yMin, yMax)
				dots = append(dots, dot{x, y, r})
			}
		}

		if len(dots) >= 3 {
			pts := dotsToPoints(dots)
			hullArea, _ := convexHullMetrics(pts)
			if diff := math.Abs(hullArea - targetHullArea); diff < bestDiff {
				bestDiff = diff
				bestDots = make([]dot, len(dots))
				copy(bestDots, dots)
			}
		}
	}

	if bestDots != nil {
		return bestDots
	}
	return placeDots(n, radii, width, height, margin, minSpacing, rng)
}

func dotsToPoints(dots []dot) [][2]float64 {
	pts := make([][2]float64, len(dots))
	for i, d := range dots {
		pts[i] = [2]float64{d.x, d.y}
	}
	return pts
}

// ---------------------------------------------------------------------------
// Convex hull (Graham scan) — area via shoelace, perimeter via edge lengths
// ---------------------------------------------------------------------------

func convexHullMetrics(pts [][2]float64) (area, perimeter float64) {
	n := len(pts)
	if n < 2 {
		return
	}
	if n == 2 {
		dx := pts[1][0] - pts[0][0]
		dy := pts[1][1] - pts[0][1]
		perimeter = 2 * math.Sqrt(dx*dx+dy*dy)
		return
	}

	hull := grahamScan(pts)
	m := len(hull)

	for i := 0; i < m; i++ {
		j := (i + 1) % m
		area += hull[i][0]*hull[j][1] - hull[j][0]*hull[i][1]
		dx := hull[j][0] - hull[i][0]
		dy := hull[j][1] - hull[i][1]
		perimeter += math.Sqrt(dx*dx + dy*dy)
	}
	area = math.Abs(area) / 2
	return
}

func grahamScan(points [][2]float64) [][2]float64 {
	n := len(points)
	pts := make([][2]float64, n)
	copy(pts, points)

	// Find bottom-most (then left-most) point
	pivot := 0
	for i := 1; i < n; i++ {
		if pts[i][1] < pts[pivot][1] ||
			(pts[i][1] == pts[pivot][1] && pts[i][0] < pts[pivot][0]) {
			pivot = i
		}
	}
	pts[0], pts[pivot] = pts[pivot], pts[0]
	p0 := pts[0]

	sort.Slice(pts[1:], func(i, j int) bool {
		a, b := pts[1:][i], pts[1:][j]
		cross := (a[0]-p0[0])*(b[1]-p0[1]) - (a[1]-p0[1])*(b[0]-p0[0])
		if cross != 0 {
			return cross > 0
		}
		da := (a[0]-p0[0])*(a[0]-p0[0]) + (a[1]-p0[1])*(a[1]-p0[1])
		db := (b[0]-p0[0])*(b[0]-p0[0]) + (b[1]-p0[1])*(b[1]-p0[1])
		return da < db
	})

	stack := []([2]float64){pts[0], pts[1]}
	for i := 2; i < n; i++ {
		for len(stack) > 1 {
			a, b, c := stack[len(stack)-2], stack[len(stack)-1], pts[i]
			if (b[0]-a[0])*(c[1]-a[1])-(b[1]-a[1])*(c[0]-a[0]) <= 0 {
				stack = stack[:len(stack)-1]
			} else {
				break
			}
		}
		stack = append(stack, pts[i])
	}
	return stack
}

// ---------------------------------------------------------------------------
// Rendering
// ---------------------------------------------------------------------------

func renderStimulus(dots []dot, width, height int, bg, dotCol color.RGBA, aa bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, bg)
		}
	}
	for _, d := range dots {
		if aa {
			drawCircleAA(img, d.x, d.y, d.radius, dotCol)
		} else {
			drawCircle(img, d.x, d.y, d.radius, dotCol)
		}
	}
	return img
}

// drawCircleAA draws an antialiased filled circle using coverage estimation.
func drawCircleAA(img *image.RGBA, cx, cy, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	rInt := int(radius) + 2
	icx, icy := int(cx), int(cy)

	for py := icy - rInt; py <= icy+rInt; py++ {
		if py < bounds.Min.Y || py >= bounds.Max.Y {
			continue
		}
		for px := icx - rInt; px <= icx+rInt; px++ {
			if px < bounds.Min.X || px >= bounds.Max.X {
				continue
			}
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			var alpha float64
			if dist <= radius-0.5 {
				alpha = 1.0
			} else if dist >= radius+0.5 {
				continue
			} else {
				alpha = radius + 0.5 - dist
			}

			if alpha >= 1.0 {
				img.SetRGBA(px, py, col)
			} else {
				blendPixel(img, px, py, col, alpha)
			}
		}
	}
}

// drawCircle draws a non-antialiased filled circle.
func drawCircle(img *image.RGBA, cx, cy, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	rInt := int(radius)
	icx := int(math.Round(cx))
	icy := int(math.Round(cy))
	r2 := radius * radius

	for py := icy - rInt; py <= icy+rInt; py++ {
		if py < bounds.Min.Y || py >= bounds.Max.Y {
			continue
		}
		for px := icx - rInt; px <= icx+rInt; px++ {
			if px < bounds.Min.X || px >= bounds.Max.X {
				continue
			}
			dx := float64(px) + 0.5 - cx
			dy := float64(py) + 0.5 - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(px, py, col)
			}
		}
	}
}

func blendPixel(img *image.RGBA, px, py int, col color.RGBA, alpha float64) {
	src := img.RGBAAt(px, py)
	fa := alpha
	ea := float64(src.A) / 255.0
	outA := fa + ea*(1-fa)
	if outA <= 0 {
		return
	}
	r := (float64(col.R)*fa + float64(src.R)*ea*(1-fa)) / outA
	g := (float64(col.G)*fa + float64(src.G)*ea*(1-fa)) / outA
	b := (float64(col.B)*fa + float64(src.B)*ea*(1-fa)) / outA
	img.SetRGBA(px, py, color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(outA * 255),
	})
}

// ---------------------------------------------------------------------------
// Ground truth metrics
// ---------------------------------------------------------------------------

func calcGroundTruth(dots []dot, width, height int, filename string) groundTruth {
	n := len(dots)
	gt := groundTruth{
		filename:    filename,
		number:      n,
		imageWidth:  width,
		imageHeight: height,
		fieldArea:   width * height,
	}
	if n == 0 {
		return gt
	}

	areas := make([]float64, n)
	for i, d := range dots {
		areas[i] = d.area()
		gt.cumulativeArea += areas[i]
		gt.totalContourLength += 2 * math.Pi * d.radius
	}

	gt.averageElementSize = gt.cumulativeArea / float64(n)

	if n > 1 {
		for _, a := range areas {
			diff := a - gt.averageElementSize
			gt.sizeSD += diff * diff
		}
		gt.sizeSD = math.Sqrt(gt.sizeSD / float64(n))
	}

	gt.minElementSize = areas[0]
	gt.maxElementSize = areas[0]
	for _, a := range areas[1:] {
		if a < gt.minElementSize {
			gt.minElementSize = a
		}
		if a > gt.maxElementSize {
			gt.maxElementSize = a
		}
	}

	if n >= 3 {
		pts := dotsToPoints(dots)
		gt.convexHullArea, gt.convexHullPerimeter = convexHullMetrics(pts)
	} else if n == 2 {
		dx := dots[0].x - dots[1].x
		dy := dots[0].y - dots[1].y
		gt.convexHullPerimeter = 2 * math.Sqrt(dx*dx+dy*dy)
	}

	if gt.convexHullArea > 0 {
		gt.density = float64(n) / gt.convexHullArea
	}
	gt.occupancy = gt.cumulativeArea / float64(gt.fieldArea)

	if n >= 2 {
		total := 0.0
		for i, a := range dots {
			minDist := math.Inf(1)
			for j, b := range dots {
				if i == j {
					continue
				}
				dx := a.x - b.x
				dy := a.y - b.y
				if d := math.Sqrt(dx*dx + dy*dy); d < minDist {
					minDist = d
				}
			}
			total += minDist
		}
		gt.avgNearestNeighborDistance = total / float64(n)
	}

	return gt
}

// ---------------------------------------------------------------------------
// CSV output
// ---------------------------------------------------------------------------

var csvHeaders = []string{
	"filename", "number",
	"cumulative_area", "average_element_size", "size_sd",
	"min_element_size", "max_element_size", "total_contour_length",
	"convex_hull_area", "convex_hull_perimeter",
	"field_area", "density", "occupancy",
	"avg_nearest_neighbor_distance",
	"image_width", "image_height",
}

func gtToRecord(gt groundTruth) []string {
	return []string{
		gt.filename,
		strconv.Itoa(gt.number),
		f2(gt.cumulativeArea),
		f2(gt.averageElementSize),
		f2(gt.sizeSD),
		f2(gt.minElementSize),
		f2(gt.maxElementSize),
		f2(gt.totalContourLength),
		f2(gt.convexHullArea),
		f2(gt.convexHullPerimeter),
		strconv.Itoa(gt.fieldArea),
		f6(gt.density),
		f6(gt.occupancy),
		f2(gt.avgNearestNeighborDistance),
		strconv.Itoa(gt.imageWidth),
		strconv.Itoa(gt.imageHeight),
	}
}

func f2(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
func f6(v float64) string { return strconv.FormatFloat(v, 'f', 6, 64) }

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// --- flags ---
	nFixed := flag.Int("n", 20, "Number of dots (fixed). Ignored if -n-min and -n-max are set differently.")
	nMin := flag.Int("n-min", 0, "Min dots when using a range (0 = use -n)")
	nMax := flag.Int("n-max", 0, "Max dots when using a range (0 = use -n)")
	count := flag.Int("count", 10, "Number of stimuli to generate")
	avgRadius := flag.Float64("avg-radius", 15.0, "Average dot radius in pixels")
	sizeVar := flag.Float64("size-variability", 0.0, "Size variability: SD of radius (0 = uniform)")
	minRadius := flag.Float64("min-radius", 5.0, "Minimum dot radius in pixels")
	controlArea := flag.Bool("control-area", false, "Scale dot sizes to reach target cumulative area")
	targetArea := flag.Float64("target-area", 5000.0, "Target cumulative area in px² (used with -control-area)")
	width := flag.Int("width", 400, "Image width in pixels")
	height := flag.Int("height", 400, "Image height in pixels")
	margin := flag.Int("margin", 20, "Margin from image edge in pixels")
	minSpacing := flag.Float64("min-spacing", 2.0, "Minimum gap between dot edges in pixels")
	whiteOnBlack := flag.Bool("white-on-black", false, "White dots on black background (default: black on white)")
	noAA := flag.Bool("no-aa", false, "Disable antialiasing")
	controlHull := flag.Bool("control-hull", false, "Attempt to constrain convex hull area (experimental)")
	targetHull := flag.Float64("target-hull", 50000.0, "Target convex hull area in px² (used with -control-hull)")
	seed := flag.Int64("seed", 0, "Random seed (0 = random)")
	prefix := flag.String("prefix", "stimulus", "Filename prefix for generated images")
	outDir := flag.String("output", ".", "Output directory for images and ground_truth.csv")

	flag.Parse()

	// Resolve dot count range
	lo, hi := *nFixed, *nFixed
	if *nMin > 0 || *nMax > 0 {
		if *nMin > 0 {
			lo = *nMin
		}
		if *nMax > 0 {
			hi = *nMax
		}
		if lo > hi {
			lo, hi = hi, lo
		}
	}

	// RNG
	var rng *rand.Rand
	if *seed == 0 {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		rng = rand.New(rand.NewSource(*seed))
	}

	// Colors
	var bg, dotCol color.RGBA
	if *whiteOnBlack {
		bg = color.RGBA{0, 0, 0, 255}
		dotCol = color.RGBA{255, 255, 255, 255}
	} else {
		bg = color.RGBA{255, 255, 255, 255}
		dotCol = color.RGBA{0, 0, 0, 255}
	}

	aa := !*noAA

	// Create output directory
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Open CSV
	csvPath := filepath.Join(*outDir, "ground_truth.csv")
	csvFile, err := os.Create(csvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating CSV: %v\n", err)
		os.Exit(1)
	}
	defer csvFile.Close()
	w := csv.NewWriter(csvFile)
	if err := w.Write(csvHeaders); err != nil {
		fmt.Fprintf(os.Stderr, "error writing CSV header: %v\n", err)
		os.Exit(1)
	}

	// Generate stimuli
	for i := 0; i < *count; i++ {
		n := lo
		if hi > lo {
			n = lo + rng.Intn(hi-lo+1)
		}

		radii := generateRadii(n, *avgRadius, *sizeVar, *minRadius,
			*controlArea, *targetArea, rng)

		var dots []dot
		if *controlHull {
			dots = placeDotsHullControlled(n, radii, *width, *height, *margin,
				*minSpacing, *targetHull, rng)
		} else {
			dots = placeDots(n, radii, *width, *height, *margin, *minSpacing, rng)
		}

		img := renderStimulus(dots, *width, *height, bg, dotCol, aa)

		filename := fmt.Sprintf("%s_%04d.png", *prefix, i+1)
		imgPath := filepath.Join(*outDir, filename)
		f, err := os.Create(imgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating image %s: %v\n", imgPath, err)
			os.Exit(1)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			fmt.Fprintf(os.Stderr, "error encoding PNG %s: %v\n", imgPath, err)
			os.Exit(1)
		}
		f.Close()

		gt := calcGroundTruth(dots, *width, *height, filename)
		if err := w.Write(gtToRecord(gt)); err != nil {
			fmt.Fprintf(os.Stderr, "error writing CSV row: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[%d/%d] %s  N=%d\n", i+1, *count, filename, n)
	}

	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "CSV flush error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done. Images and ground_truth.csv written to %s\n", *outDir)
}
