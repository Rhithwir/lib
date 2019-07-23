package rhithwirlib

import (
	"math"
	"math/bits"
	"time"

	"github.com/325gerbils/go-vector"
	"github.com/stianeikeland/go-rpio"
)

// QuadsToLinear converts a quad-formatted frame to a linear-formatted frame
func QuadsToLinear(frame [64]int) (out []int) {
	for i := 0; i < len(frame); i++ {
		ni := (4 * int(i/8)) + (imod(i, 4) + int(imod(i/4, 2))*16) + (16 * int(i/32)) // <- result of too much caffeine
		out = append(out, frame[ni])
	}
	return out
}

// DrivePuckLoc drives the puck (location only, no rotational correction)
func DrivePuckLoc(target vector.Vector) {
	data := GetHallEffect(0)
	puck := PuckLocation(data)
	err := puck.Sub(target).Normalize()
	//phase :=
	for sel := 0; sel < 8; sel++ { // yu yd xl xr yu yd xl xr
		// yu and yd both pwm.Y
		// xl and xr both pwm.X

		// just X

		//Write()
	}
}

// GetHallEffect returns a hall effect grid for a board
func GetHallEffect(board uint) (data [64]int) {
	for quad := 0; quad < 4; quad++ {
		for anadr := 0; anadr < 16; anadr++ {
			v := Read(board, uint(quad), uint(anadr))
			data[i] = v
		}
	}
	data = QuadsToLinear(data) // will this work?
	return
}

// i have no idea if this will work at all
func locToPhase(in vector.Vector) (out vector.Vector) {
	phaseLength := 0.5 // whats the real value? is each phase 1 inch? half a inch? more? less? if one iteration of all phases is X inches then each phase is X/4 inches
	out.X, out.Y = imod(in.X, phaseLength*4), imod(in.Y, phaseLength*4)
	return
}

// PuckLocation gets location of puck from hall effect data
func PuckLocation(data []int) vector.Vector {
	var pointCloud []vector.Vector
	for threshold := -15; threshold <= 15; threshold++ {
		for i := 0; i < len(data); i++ {
			x := imod(i, 8)
			y := int(i / 8)
			if 127+threshold == data[i] { // will this work?
				pointCloud = append(pointCloud, vector.New(x, y))
			}
		}
	}
	return centerOfPoints(pointCloud)
}

func centerOfPoints(data []vector.Vector) vector.Vector {
	var xs = make([]float64, len(data))
	var ys = make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		xs[i] = data[i].X
		ys[i] = data[i].Y
	}
	output := vector.New(mean(xs), mean(ys))
	return output
}

// helpers
func imod(a, b int) int {
	return int(math.Mod(float64(a), float64(b)))
}
func distsq(x1, y1, x2, y2 float64) float64 {
	return (y2-y1)*(y2-y1) + (x2-x1)*(x2-x1)
}
func sum(data []float64) float64 {
	out := 0.0
	for i := 0; i < len(data); i++ {
		out += data[i]
	}
	return out
}
func mean(data []float64) float64 {
	return sum(data) / float64(len(data))
}

// Driver

var pins = []rpio.Pin{
	rpio.Pin(4),
	rpio.Pin(25),
	rpio.Pin(24),
	rpio.Pin(23),
	rpio.Pin(22),
	rpio.Pin(27),
	rpio.Pin(18),
	rpio.Pin(2),
	rpio.Pin(3),
	rpio.Pin(8),
	rpio.Pin(7),
	rpio.Pin(10),
	rpio.Pin(9),
	rpio.Pin(11),
	rpio.Pin(6),
	rpio.Pin(13),
	rpio.Pin(19),
	rpio.Pin(26),
	rpio.Pin(12),
	rpio.Pin(16),
}
var strobe = rpio.Pin(5)
var rw = rpio.Pin(20)

// Init does the RPI setup
func Init() {
	rpio.Open()
	go func() {
		// hopefully this will fire rpio.Close when I exit main and main force-exits this infinite loop
		defer rpio.Close()
		for {
			time.Sleep(1 * time.Millisecond) // no resource hogging!
		}
	}()
	strobe.Output()
	rw.Output()

	strobe.High()
	rw.High()
}

// Write to Rpi
func Write(pwm, phase, board, sel uint) {
	word := sel + board<<6 + uint(bits.Reverse8(uint8(pwm)<<4))<<16 + phase<<14 // bitbanging, ouch
	for i, p := range pins {
		p.Output()
		switch word >> uint(len(pins)-1-i) & 1 { // more bitbanging ouch
		case 0:
			p.Low()
		case 1:
			p.High()
		}
	}
	strobe.Low()
	strobe.High()
}

// Read from Rpi
func Read(board, quad, anadr uint) (data int) {
	// write command
	word := anadr + quad<<4 + board<<6 // yay more bitbanging
	for i, p := range pins {
		p.Output()
		switch word >> uint(len(pins)-1-i) & 1 { // -.-
		case 0:
			p.Low()
		case 1:
			p.High()
		}
	}
	strobe.Low()
	strobe.High()
	rw.Low()
	// get data
	for i := 0; i < 8; i++ {
		p := pins[i]
		p.Input()
		data += int(p.Read()) << uint(8-i-1) // is this done yet?
	}
	// return
	return
}
