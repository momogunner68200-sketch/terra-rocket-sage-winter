package main

import (
	"image/png"
	"math"
	"os"
	"testing"
)

func paintScene(letter byte, markerDeg, redFrom, redTo int) []byte {
	const w, h = 240, 240
	pix := make([]byte, w*h*4)
	for i := 0; i < len(pix); i += 4 {
		pix[i], pix[i+1], pix[i+2], pix[i+3] = 18, 22, 20, 255
	}
	cx, cy := 120.0, 120.0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Hypot(dx, dy)
			if dist < 72 || dist > 96 {
				continue
			}
			deg := int(math.Atan2(dy, dx)*180/math.Pi+180) % 360
			i := (y*w + x) * 4
			if angleInside(deg, redFrom, redTo) {
				pix[i], pix[i+1], pix[i+2] = 40, 40, 210
			} else {
				pix[i], pix[i+1], pix[i+2] = 70, 74, 72
			}
		}
	}
	drawTrait(pix, w, h, 120, 120, markerDeg)
	stampLetter(pix, w, letter)
	return pix
}

func angleInside(deg, from, to int) bool {
	if from == to {
		return false
	}
	if from <= to {
		return deg >= from && deg <= to
	}
	return deg >= from || deg <= to
}

func drawTrait(pix []byte, w, h int, cx, cy float64, deg int) {
	for delta := -6; delta <= 6; delta++ {
		ang := float64(deg+delta-180) * math.Pi / 180
		c, s := math.Cos(ang), math.Sin(ang)
		for rad := 76.0; rad <= 98.0; rad += 1.2 {
			x := int(cx + c*rad)
			y := int(cy + s*rad)
			if x < 0 || y < 0 || x >= w || y >= h {
				continue
			}
			i := (y*w + x) * 4
			pix[i], pix[i+1], pix[i+2] = 25, 18, 210
		}
	}
}

func stampLetter(pix []byte, w int, key byte) {
	var bits string
	for _, letter := range letters {
		if letter.key == key {
			bits = letter.bits
		}
	}
	size := 36
	ox := 120 - size/2
	oy := 120 - size/2
	for gy := 0; gy < 16; gy++ {
		for gx := 0; gx < 16; gx++ {
			if bits[gy*16+gx] != '1' {
				continue
			}
			for yy := 0; yy < size/16+1; yy++ {
				for xx := 0; xx < size/16+1; xx++ {
					x := ox + gx*(size/16) + xx
					y := oy + gy*(size/16) + yy
					i := (y*w + x) * 4
					pix[i], pix[i+1], pix[i+2] = 240, 236, 228
				}
			}
		}
	}
}

func TestMarkerInsideRedZone(t *testing.T) {
	var tr Track
	pressed := false
	for deg := 0; deg <= 70; deg += 2 {
		got := Analyze(paintScene('Z', deg, 20, 80), 240, 240)
		if got.Key != 'Z' || !got.Seen {
			t.Fatalf("deg %d: %+v", deg, got)
		}
		if tr.Press(got) {
			pressed = true
		}
	}
	if !pressed {
		t.Fatal("le trait est entré dans le rouge sans touche")
	}
}

func TestMarkerOutsideRedZone(t *testing.T) {
	pix := paintScene('Q', 200, 20, 80)
	got := Analyze(pix, 240, 240)
	if !got.Seen || !got.HasNeedle || got.InZone || got.Key != 'Q' {
		t.Fatalf("outside zone: %+v", got)
	}
	var tr Track
	if tr.Press(got) {
		t.Fatal("touche alors que le trait est hors zone")
	}
}

func TestLineEntersRedZoneOnce(t *testing.T) {
	var tr Track
	var hits []int
	for deg := 0; deg < 360; deg += 4 {
		got := Analyze(paintScene('Z', deg, 30, 90), 240, 240)
		if got.Key != 'Z' || !got.Seen {
			t.Fatalf("deg %d: %+v", deg, got)
		}
		if tr.Press(got) {
			hits = append(hits, deg)
		}
	}
	if len(hits) != 1 || hits[0] < 28 || hits[0] > 92 {
		t.Fatalf("la touche partirait aux degrés %v", hits)
	}
}

func TestRedWithoutRingIsIgnored(t *testing.T) {
	const w, h = 240, 240
	pix := make([]byte, w*h*4)
	for i := 0; i < len(pix); i += 4 {
		pix[i], pix[i+1], pix[i+2], pix[i+3] = 40, 50, 60, 255
	}
	for y := 80; y < 150; y++ {
		for x := 30; x < 70; x++ {
			i := (y*w + x) * 4
			pix[i], pix[i+1], pix[i+2] = 30, 30, 220
		}
	}
	got := Analyze(pix, w, h)
	if got.Seen || got.Key != 0 {
		t.Fatalf("faux cercle: %+v", got)
	}
}

func TestNoRedZoneDoesNotArm(t *testing.T) {
	pix := paintScene('S', 40, 0, 0)
	got := Analyze(pix, 240, 240)
	if got.Seen || got.InZone {
		t.Fatalf("no zone: %+v", got)
	}
}

func TestPhotoQTE(t *testing.T) {
	f, err := os.Open("/workspace/attachments/image.png")
	if err != nil {
		t.Skip(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			i := (y*w + x) * 4
			pix[i], pix[i+1], pix[i+2], pix[i+3] = byte(bl>>8), byte(g>>8), byte(r>>8), 255
		}
	}
	pix, w, h = cropPreview(pix, w, h)
	got := Analyze(pix, w, h)
	if !got.Seen || got.Key != 'Q' {
		t.Fatalf("photo: %+v %dx%d", got, w, h)
	}
}

func cropPreview(pix []byte, w, h int) ([]byte, int, int) {
	minX, minY, maxX, maxY := w, h, 0, 0
	for y := h * 7 / 10; y < h; y++ {
		for x := 0; x < w; x++ {
			b, g, r := int(pix[(y*w+x)*4]), int(pix[(y*w+x)*4+1]), int(pix[(y*w+x)*4+2])
			if r > 180 && r > g+40 && r > b+40 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX <= minX {
		return pix, w, h
	}
	cx := (minX + maxX) / 2
	cy := minY + (maxY-minY)/3
	side := (maxY - minY) * 2
	if side < 80 {
		side = 80
	}
	x0, y0 := cx-side/2, cy-side/2
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x0+side > w {
		side = w - x0
	}
	if y0+side > h {
		side = h - y0
	}
	out := make([]byte, side*side*4)
	for y := 0; y < side; y++ {
		copy(out[y*side*4:(y+1)*side*4], pix[((y0+y)*w+x0)*4:((y0+y)*w+x0+side)*4])
	}
	return out, side, side
}

func TestThinWhiteLettersOnBlack(t *testing.T) {
	for _, key := range []byte{'Z', 'Q', 'S', 'D'} {
		pix := paintScene(key, 200, 30, 90)
		clearDisk(pix, 240, 120, 120, 58)
		drawThinLetter(pix, 240, 120, 120, key)
		got := Analyze(pix, 240, 240)
		if got.Key != key || !got.Seen {
			t.Fatalf("lettre blanche %c got %q seen %v", key, got.Key, got.Seen)
		}
	}
}

func clearDisk(pix []byte, w, cx, cy, radius int) {
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy > r2 {
				continue
			}
			i := (y*w + x) * 4
			pix[i], pix[i+1], pix[i+2] = 6, 6, 6
		}
	}
}

func drawThinLetter(pix []byte, w, cx, cy int, key byte) {
	put := func(x, y int) {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				xx, yy := x+dx, y+dy
				if xx < 0 || yy < 0 || xx >= w || yy >= w {
					continue
				}
				i := (yy*w + xx) * 4
				pix[i], pix[i+1], pix[i+2] = 245, 245, 245
			}
		}
	}
	line := func(x0, y0, x1, y1 int) {
		steps := abs(x1-x0) + abs(y1-y0)
		if steps < 1 {
			steps = 1
		}
		for i := 0; i <= steps; i++ {
			x := x0 + (x1-x0)*i/steps
			y := y0 + (y1-y0)*i/steps
			put(x, y)
		}
	}
	switch key {
	case 'Z':
		line(cx-16, cy-16, cx+16, cy-16)
		line(cx+16, cy-16, cx-16, cy+16)
		line(cx-16, cy+16, cx+16, cy+16)
	case 'S':
		line(cx+14, cy-16, cx-14, cy-16)
		line(cx-14, cy-16, cx-14, cy-2)
		line(cx-14, cy-2, cx+14, cy-2)
		line(cx+14, cy-2, cx+14, cy+16)
		line(cx+14, cy+16, cx-14, cy+16)
	case 'D':
		line(cx-14, cy-16, cx-14, cy+16)
		line(cx-14, cy-16, cx+4, cy-16)
		line(cx+4, cy-16, cx+16, cy)
		line(cx+16, cy, cx+4, cy+16)
		line(cx+4, cy+16, cx-14, cy+16)
	case 'Q':
		line(cx-12, cy-12, cx+10, cy-12)
		line(cx+10, cy-12, cx+10, cy+8)
		line(cx+10, cy+8, cx-12, cy+8)
		line(cx-12, cy+8, cx-12, cy-12)
		line(cx+2, cy+2, cx+16, cy+16)
	}
}
