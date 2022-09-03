package ftoa

import (
	"github.com/rarnu/goscript/ftoa/fast"
	"math"
	"strconv"
)

type FToStrMode int

const (
	ModeStandard FToStrMode = iota
	ModeStandardExponential
	ModeFixed
	ModeExponential
	ModePrecision
)

func insert(b []byte, p int, c byte) []byte {
	b = append(b, 0)
	copy(b[p+1:], b[p:])
	b[p] = c
	return b
}

func expand(b []byte, delta int) []byte {
	newLen := len(b) + delta
	if newLen <= cap(b) {
		return b[:newLen]
	}
	b1 := make([]byte, newLen)
	copy(b1, b)
	return b1
}

func FToStr(d float64, mode FToStrMode, precision int, buffer []byte) []byte {
	if math.IsNaN(d) {
		buffer = append(buffer, "NaN"...)
		return buffer
	}
	if math.IsInf(d, 0) {
		if math.Signbit(d) {
			buffer = append(buffer, '-')
		}
		buffer = append(buffer, "Infinity"...)
		return buffer
	}

	if mode == ModeFixed && (d >= 1e21 || d <= -1e21) {
		mode = ModeStandard
	}

	var decPt int
	var ok bool
	startPos := len(buffer)

	if d != 0 {
		if d < 0 {
			buffer = append(buffer, '-')
			d = -d
			startPos++
		}
		switch mode {
		case ModeStandard, ModeStandardExponential:
			buffer, decPt, ok = fast.Dtoa(d, fast.ModeShortest, 0, buffer)
		case ModeExponential, ModePrecision:
			buffer, decPt, ok = fast.Dtoa(d, fast.ModePrecision, precision, buffer)
		}
	} else {
		buffer = append(buffer, '0')
		decPt, ok = 1, true
	}
	if !ok {
		buffer, decPt = ftoa(d, dtoaModes[mode], mode >= ModeFixed, precision, buffer)
	}
	exponentialNotation := false
	minNDigits := 0
	nDigits := len(buffer) - startPos

	switch mode {
	case ModeStandard:
		if decPt < -5 || decPt > 21 {
			exponentialNotation = true
		} else {
			minNDigits = decPt
		}
	case ModeFixed:
		if precision >= 0 {
			minNDigits = decPt + precision
		} else {
			minNDigits = decPt
		}
	case ModeExponential:
		minNDigits = precision
		fallthrough
	case ModeStandardExponential:
		exponentialNotation = true
	case ModePrecision:
		minNDigits = precision
		if decPt < -5 || decPt > precision {
			exponentialNotation = true
		}
	}

	for nDigits < minNDigits {
		buffer = append(buffer, '0')
		nDigits++
	}

	if exponentialNotation {
		if nDigits != 1 {
			buffer = insert(buffer, startPos+1, '.')
		}
		buffer = append(buffer, 'e')
		if decPt-1 >= 0 {
			buffer = append(buffer, '+')
		}
		buffer = strconv.AppendInt(buffer, int64(decPt-1), 10)
	} else if decPt != nDigits {
		if decPt > 0 {
			buffer = insert(buffer, startPos+decPt, '.')
		} else {
			buffer = expand(buffer, 2-decPt)
			copy(buffer[startPos+2-decPt:], buffer[startPos:])
			buffer[startPos] = '0'
			buffer[startPos+1] = '.'
			for i := startPos + 2; i < startPos+2-decPt; i++ {
				buffer[i] = '0'
			}
		}
	}
	return buffer
}
