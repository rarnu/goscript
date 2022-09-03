package fast

import "math"

const (
	diyFpKSignificandSize           = 64
	kSignificandSize                = 53
	kUint64MSB               uint64 = 1 << 63
	kSignificandMask                = 0x000FFFFFFFFFFFFF
	kHiddenBit                      = 0x0010000000000000
	kExponentMask                   = 0x7FF0000000000000
	kPhysicalSignificandSize        = 52
	kExponentBias                   = 0x3FF + kPhysicalSignificandSize
	kDenormalExponent               = -kExponentBias + 1
)

type double float64

type diyfp struct {
	f uint64
	e int
}

func (f *diyfp) subtract(o diyfp) {
	_DCHECK(f.e == o.e)
	_DCHECK(f.f >= o.f)
	f.f -= o.f
}

func (f diyfp) minus(o diyfp) diyfp {
	res := f
	res.subtract(o)
	return res
}

func (f *diyfp) mul(o diyfp) {
	const kM32 uint64 = 0xFFFFFFFF
	a := f.f >> 32
	b := f.f & kM32
	c := o.f >> 32
	d := o.f & kM32
	ac := a * c
	bc := b * c
	ad := a * d
	bd := b * d
	tmp := (bd >> 32) + (ad & kM32) + (bc & kM32)
	tmp += 1 << 31
	result_f := ac + (ad >> 32) + (bc >> 32) + (tmp >> 32)
	f.e += o.e + 64
	f.f = result_f
}

func (f diyfp) times(o diyfp) diyfp {
	res := f
	res.mul(o)
	return res
}

func (f *diyfp) _normalize() {
	f_, e := f.f, f.e
	const k10MSBits uint64 = 0x3FF << 54
	for f_&k10MSBits == 0 {
		f_ <<= 10
		e -= 10
	}
	for f_&kUint64MSB == 0 {
		f_ <<= 1
		e--
	}
	f.f, f.e = f_, e
}

func normalizeDiyfp(f diyfp) diyfp {
	res := f
	res._normalize()
	return res
}

func (d double) toNormalizedDiyfp() diyfp {
	f, e := d.sigExp()
	for (f & kHiddenBit) == 0 {
		f <<= 1
		e--
	}
	f <<= diyFpKSignificandSize - kSignificandSize
	e -= diyFpKSignificandSize - kSignificandSize
	return diyfp{f, e}
}

func (d double) normalizedBoundaries() (m_minus, m_plus diyfp) {
	v := d.toDiyFp()
	significand_is_zero := v.f == kHiddenBit
	m_plus = normalizeDiyfp(diyfp{f: (v.f << 1) + 1, e: v.e - 1})
	if significand_is_zero && v.e != kDenormalExponent {
		m_minus = diyfp{f: (v.f << 2) - 1, e: v.e - 2}
	} else {
		m_minus = diyfp{f: (v.f << 1) - 1, e: v.e - 1}
	}
	m_minus.f <<= m_minus.e - m_plus.e
	m_minus.e = m_plus.e
	return
}

func (d double) toDiyFp() diyfp {
	f, e := d.sigExp()
	return diyfp{f: f, e: e}
}

func (d double) sigExp() (significand uint64, exponent int) {
	d64 := math.Float64bits(float64(d))
	significand = d64 & kSignificandMask
	if d64&kExponentMask != 0 {
		significand += kHiddenBit
		exponent = int((d64&kExponentMask)>>kPhysicalSignificandSize) - kExponentBias
	} else {
		exponent = kDenormalExponent
	}
	return
}
