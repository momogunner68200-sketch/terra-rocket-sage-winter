package main

import (
	"math"
	"sort"
)

const bins = 72

type Result struct {
	Key       byte
	InZone    bool
	Seen      bool
	Marker    bool
	HasNeedle bool
	Needle    int
	ZoneStart int
	ZoneLen   int
}

func Analyze(pix []byte, w, h int) Result {
	if w < 40 || h < 40 || len(pix) < w*h*4 {
		return Result{}
	}
	cx := float64(w) / 2
	cy := float64(h) / 2
	half := math.Min(float64(w), float64(h)) / 2
	rMin := half * 0.16
	rMax := half * 0.94

	redBins := make([]int, bins)
	var sum, sumsq float64
	nRed := 0
	const rays = 180
	for i := 0; i < rays; i++ {
		ang := float64(i)/float64(rays)*2*math.Pi - math.Pi
		c, s := math.Cos(ang), math.Sin(ang)
		bin := i * bins / rays
		for rad := rMin; rad <= rMax; rad++ {
			x := int(cx + c*rad)
			y := int(cy + s*rad)
			if x < 0 || y < 0 || x >= w || y >= h {
				continue
			}
			o := (y*w + x) * 4
			bb, gg, rr := int(pix[o]), int(pix[o+1]), int(pix[o+2])
			if !isRed(rr, gg, bb) {
				continue
			}
			redBins[bin]++
			sum += rad
			sumsq += rad * rad
			nRed++
		}
	}
	if nRed < 12 {
		return Result{}
	}
	mean := sum / float64(nRed)
	if mean < 8 {
		return Result{}
	}
	variance := sumsq/float64(nRed) - mean*mean
	if variance < 0 {
		variance = 0
	}
	if math.Sqrt(variance) >= mean*0.40 {
		return Result{}
	}
	bandStart, bandLen := longestRun(activeBins(redBins, 2))
	if bandLen < 6 || bandLen > bins*2/3 {
		return Result{}
	}
	_, contrastN := ringScore(pix, w, h, cx, cy, mean)
	key := readKey(pix, w, h, int(mean*0.7))
	if contrastN < 14 {
		return Result{}
	}
	needle, okNeedle := outsidePeak(redBins, bandStart, bandLen)
	return Result{
		Key:       key,
		Seen:      true,
		Marker:    okNeedle,
		HasNeedle: okNeedle,
		Needle:    needle,
		ZoneStart: bandStart,
		ZoneLen:   bandLen,
		InZone:    okNeedle && distToArc(needle, bandStart, bandLen) == 0,
	}
}

type Track struct {
	have   bool
	live   bool
	bin    int
	dir    int
	step   int
	lastD  int
	zone0  int
	zoneN  int
	out    bool
	held   bool
}

func (t *Track) Press(res Result) bool {
	if !res.Seen || res.ZoneLen < 2 {
		*t = Track{}
		return false
	}
	if res.HasNeedle {
		d := distToArc(res.Needle, res.ZoneStart, res.ZoneLen)
		if d >= 2 {
			t.zone0 = res.ZoneStart
			t.zoneN = res.ZoneLen
		}
		if t.have {
			delta := res.Needle - t.bin
			if delta > bins/2 {
				delta -= bins
			}
			if delta < -bins/2 {
				delta += bins
			}
			if delta != 0 {
				if delta > 0 {
					t.dir = 1
				} else {
					t.dir = -1
				}
				t.step = abs(delta)
			}
		}
		inside := t.zoneN >= 6 && deepIn(res.Needle, t.zone0, t.zoneN)
		t.bin = res.Needle
		t.have = true
		t.live = true
		t.lastD = d
		if inside && !t.held {
			t.held = true
			t.out = false
			return true
		}
		t.out = !inside
		if d > 2 {
			t.held = false
		}
		return false
	}
	if t.live && t.have && t.out && !t.held && t.zoneN >= 6 && t.dir != 0 && t.lastD <= 5 {
		adv := t.step
		if adv < 1 {
			adv = 1
		}
		pred := t.bin
		for i := 0; i < adv; i++ {
			pred = (pred + t.dir + bins) % bins
		}
		t.bin = pred
		t.lastD = distToArc(pred, t.zone0, t.zoneN)
		if deepIn(pred, t.zone0, t.zoneN) {
			t.held = true
			t.live = false
			t.out = false
			return true
		}
		return false
	}
	t.live = false
	return false
}

func isQteKey(key byte) bool {
	return key == 'Z' || key == 'Q' || key == 'S' || key == 'D'
}

func circDist(a, b int) int {
	d := abs(a - b)
	if d > bins/2 {
		d = bins - d
	}
	return d
}

func deepIn(bin, start, length int) bool {
	if length < 3 {
		return inArc(bin, start, length, 0)
	}
	return inArc(bin, (start+1)%bins, length-2, 0)
}

func outsidePeak(counts []int, start, length int) (int, bool) {
	bestBin, best := 0, 0
	for i, n := range counts {
		if n < 2 || inArc(i, start, length, 0) {
			continue
		}
		if n > best {
			best = n
			bestBin = i
		}
	}
	return bestBin, best >= 2
}

func distToArc(bin, start, length int) int {
	best := bins
	for d := 0; d < length; d++ {
		b := (start + d) % bins
		delta := abs(bin - b)
		if delta > bins/2 {
			delta = bins - delta
		}
		if delta < best {
			best = delta
		}
	}
	return best
}

func rgbAt(pix []byte, w, x, y int) (b, g, r int) {
	i := (y*w + x) * 4
	return int(pix[i]), int(pix[i+1]), int(pix[i+2])
}

func isRed(r, g, b int) bool {
	return r > 140 && r > g+40 && r > b+40
}

func angleBin(dx, dy float64) int {
	deg := math.Atan2(dy, dx)*180/math.Pi + 180
	bin := int(deg) * bins / 360
	if bin < 0 {
		bin = 0
	}
	if bin >= bins {
		bin = bins - 1
	}
	return bin
}

func activeBins(counts []int, min int) []bool {
	out := make([]bool, len(counts))
	for i, n := range counts {
		out[i] = n >= min
	}
	return out
}

func longestRun(flags []bool) (start, length int) {
	n := len(flags)
	curStart, cur := 0, 0
	for i := 0; i < n*2; i++ {
		if flags[i%n] {
			if cur == 0 {
				curStart = i
			}
			cur++
			if cur <= n && cur > length {
				length = cur
				start = curStart % n
			}
		} else {
			cur = 0
		}
	}
	return start, length
}

func inArc(bin, start, length, pad int) bool {
	for d := -pad; d < length+pad; d++ {
		if (start+d+bins)%bins == bin {
			return true
		}
	}
	return false
}

func ringScore(pix []byte, w, h int, cx, cy, mean float64) (redN, contrastN int) {
	for i := 0; i < bins; i++ {
		ang := float64(i)/float64(bins)*2*math.Pi - math.Pi
		c, s := math.Cos(ang), math.Sin(ang)
		inn := lumaAt(pix, w, h, cx+c*mean*0.5, cy+s*mean*0.5)
		out := lumaAt(pix, w, h, cx+c*mean*1.5, cy+s*mean*1.5)
		if inn < 0 || out < 0 {
			continue
		}
		best := 0
		red := false
		for rad := mean * 0.75; rad <= mean*1.25; rad += 1.5 {
			x := int(cx + c*rad)
			y := int(cy + s*rad)
			if x < 0 || y < 0 || x >= w || y >= h {
				continue
			}
			bb, gg, rr := rgbAt(pix, w, x, y)
			if isRed(rr, gg, bb) {
				red = true
				break
			}
			luma := rr + gg + bb
			d := abs(luma - inn)
			if o := abs(luma - out); o < d {
				d = o
			}
			if d > best {
				best = d
			}
		}
		if red {
			redN++
		} else if best > 24 {
			contrastN++
		}
	}
	return redN, contrastN
}

func lumaAt(pix []byte, w, h int, x, y float64) int {
	ix, iy := int(x), int(y)
	if ix < 0 || iy < 0 || ix >= w || iy >= h {
		return -1
	}
	b, g, r := rgbAt(pix, w, ix, iy)
	return r + g + b
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

type letterTpl struct {
	key  byte
	bits string
}

var letters = []letterTpl{
	{'Z', "1111111111111111" + "1111111111111110" + "0000000000011100" + "0000000000111000" + "0000000001110000" + "0000000011100000" + "0000000111000000" + "0000001110000000" + "0000011100000000" + "0000111000000000" + "0001110000000000" + "0011100000000000" + "0111000000000000" + "1110000000000000" + "1111111111111111" + "1111111111111111"},
	{'S', "0011111111111100" + "0111111111111110" + "0111000000000000" + "0111000000000000" + "0111000000000000" + "0111111111111000" + "0011111111111100" + "0000000000011100" + "0000000000011100" + "0000000000011100" + "0000000000011100" + "0111000000011100" + "0111111111111100" + "0011111111111000" + "0000000000000000" + "0000000000000000"},
	{'D', "1111111111100000" + "1111111111110000" + "1111000001111000" + "1111000000111100" + "1111000000011100" + "1111000000011100" + "1111000000011100" + "1111000000011100" + "1111000000011100" + "1111000000011100" + "1111000000111100" + "1111000001111000" + "1111111111110000" + "1111111111100000" + "0000000000000000" + "0000000000000000"},
	{'Q', "0001111111110000" + "0011111111111000" + "0111100000111100" + "0111000000011100" + "0111000000011100" + "0111000000011100" + "0111000000011100" + "0111000000011110" + "0111000000111110" + "0111100001111100" + "0011111111111000" + "0001111111110000" + "0000000001110000" + "0000000000111000" + "0000000000011100" + "0000000000000000"},
}

func readKey(pix []byte, w, h, inner int) byte {
	try := func(th int) byte {
		glyph, ok := glyphBitsTh(pix, w, h, inner, th)
		if !ok {
			return 0
		}
		if key := structuralKey(glyph); key != 0 {
			return key
		}
		return classifyGlyph(glyph)
	}
	if key := try(0); key != 0 {
		return key
	}
	for _, th := range []int{220, 170, 130, 90} {
		if key := try(th); key != 0 {
			return key
		}
	}
	return 0
}

func classifyGlyph(glyph string) byte {
	if len(glyph) != 256 {
		return 0
	}
	bestKey := byte(0)
	best, second := 0, 0
	for _, letter := range letters {
		n := 0
		for i := 0; i < 256; i++ {
			if glyph[i] == '1' && letter.bits[i] == '1' {
				n++
			}
		}
		if n > best {
			second = best
			best = n
			bestKey = letter.key
		} else if n > second {
			second = n
		}
	}
	if best >= 24 && best >= second+6 {
		return bestKey
	}
	return 0
}

func structuralKey(sample string) byte {
	if len(sample) != 256 {
		return 0
	}
	holes := countHoles(sample)
	top, bot, mid, diag, left, tail, botLeft := 0, 0, 0, 0, 0, 0, 0
	for x := 1; x <= 14; x++ {
		if sample[x] == '1' || sample[16+x] == '1' {
			top++
		}
		if sample[15*16+x] == '1' || sample[14*16+x] == '1' || sample[13*16+x] == '1' {
			bot++
		}
	}
	for y := 6; y <= 9; y++ {
		run := 0
		for x := 0; x < 16; x++ {
			if sample[y*16+x] == '1' {
				run++
				if run > mid {
					mid = run
				}
			} else {
				run = 0
			}
		}
	}
	for y := 2; y <= 13; y++ {
		x := 15 - y
		for dx := -1; dx <= 1; dx++ {
			xx := x + dx
			if xx >= 0 && xx < 16 && sample[y*16+xx] == '1' {
				diag++
				break
			}
		}
		if sample[y*16] == '1' || sample[y*16+1] == '1' || sample[y*16+2] == '1' {
			left++
		}
	}
	for y := 13; y <= 15; y++ {
		for x := 0; x <= 6; x++ {
			if sample[y*16+x] == '1' {
				botLeft++
			}
		}
		for x := 10; x <= 15; x++ {
			if sample[y*16+x] == '1' {
				tail++
			}
		}
	}
	if holes > 0 && tail >= 4 && botLeft <= 3 {
		return 'Q'
	}
	if holes > 0 && left >= 8 {
		return 'D'
	}
	if mid >= 8 && top >= 6 && bot >= 5 && diag <= 8 && holes == 0 {
		return 'S'
	}
	if diag >= 7 && top >= 6 && bot >= 6 && mid < 8 && holes == 0 {
		return 'Z'
	}
	return 0
}

func countHoles(sample string) int {
	seen := make([]bool, 256)
	var flood func(int)
	flood = func(i int) {
		if i < 0 || i >= 256 || seen[i] || sample[i] == '1' {
			return
		}
		seen[i] = true
		x := i % 16
		if x > 0 {
			flood(i - 1)
		}
		if x < 15 {
			flood(i + 1)
		}
		if i >= 16 {
			flood(i - 16)
		}
		if i < 240 {
			flood(i + 16)
		}
	}
	for i := 0; i < 256; i++ {
		x, y := i%16, i/16
		if x == 0 || y == 0 || x == 15 || y == 15 {
			flood(i)
		}
	}
	holes := 0
	for i := 0; i < 256; i++ {
		if sample[i] == '1' || seen[i] {
			continue
		}
		holes++
		flood(i)
	}
	return holes
}

func glyphBits(pix []byte, w, h, inner int) (string, bool) {
	return glyphBitsTh(pix, w, h, inner, 0)
}

func glyphBitsTh(pix []byte, w, h, inner, minLum int) (string, bool) {
	cx, cy := w/2, h/2
	if inner < 10 {
		inner = w / 5
	}
	if inner > w/2-2 {
		inner = w/2 - 2
	}
	x0, y0 := cx-inner, cy-inner
	x1, y1 := cx+inner, cy+inner
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > w {
		x1 = w
	}
	if y1 > h {
		y1 = h
	}
	bw, bh := x1-x0, y1-y0
	if bw < 8 || bh < 8 {
		return "", false
	}
	lums := make([]int, 0, bw*bh)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			bb, gg, rr := rgbAt(pix, w, x, y)
			if isRed(rr, gg, bb) {
				continue
			}
			lums = append(lums, rr+gg+bb)
		}
	}
	if len(lums) < 16 {
		return "", false
	}
	th := minLum
	if th == 0 {
		sort.Ints(lums)
		th = lums[len(lums)*40/100] + 40
		if th < 100 {
			th = 100
		}
	}
	mask := make([]byte, bw*bh)
	nOn := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			bb, gg, rr := rgbAt(pix, w, x, y)
			if isRed(rr, gg, bb) {
				continue
			}
			if rr+gg+bb >= th {
				mask[(y-y0)*bw+(x-x0)] = 1
				nOn++
			}
		}
	}
	if nOn < 8 {
		return "", false
	}
	bestN, best := 0, -1
	seen := make([]byte, bw*bh)
	for i := range mask {
		if mask[i] == 0 || seen[i] != 0 {
			continue
		}
		stack := []int{i}
		seen[i] = 1
		n := 0
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			n++
			x := p % bw
			for _, d := range []int{-1, 1, -bw, bw} {
				q := p + d
				if q < 0 || q >= len(mask) {
					continue
				}
				if (d == -1 && x == 0) || (d == 1 && x == bw-1) {
					continue
				}
				if mask[q] == 1 && seen[q] == 0 {
					seen[q] = 1
					stack = append(stack, q)
				}
			}
		}
		if n > bestN {
			bestN, best = n, i
		}
	}
	if best < 0 || bestN < 8 {
		return "", false
	}
	comp := make([]byte, len(mask))
	stack := []int{best}
	comp[best] = 1
	minX, minY, maxX, maxY := bw, bh, 0, 0
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		x, y := p%bw, p/bw
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
		for _, d := range []int{-1, 1, -bw, bw} {
			q := p + d
			if q < 0 || q >= len(mask) {
				continue
			}
			if (d == -1 && x == 0) || (d == 1 && x == bw-1) {
				continue
			}
			if mask[q] == 1 && comp[q] == 0 {
				comp[q] = 1
				stack = append(stack, q)
			}
		}
	}
	if maxX <= minX || maxY <= minY {
		return "", false
	}
	gw, gh := maxX-minX+1, maxY-minY+1
	out := make([]byte, 256)
	for i := range out {
		out[i] = '0'
	}
	for gy := 0; gy < 16; gy++ {
		for gx := 0; gx < 16; gx++ {
			sx := minX + gx*gw/16
			sy := minY + gy*gh/16
			ex := minX + (gx+1)*gw/16
			ey := minY + (gy+1)*gh/16
			if ex <= sx {
				ex = sx + 1
			}
			if ey <= sy {
				ey = sy + 1
			}
			on, tot := 0, 0
			for y := sy; y < ey && y <= maxY; y++ {
				for x := sx; x < ex && x <= maxX; x++ {
					tot++
					if comp[y*bw+x] == 1 {
						on++
					}
				}
			}
			if tot > 0 && on*2 >= tot {
				out[gy*16+gx] = '1'
			}
		}
	}
	return string(out), true
}
