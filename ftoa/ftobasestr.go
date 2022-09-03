package ftoa

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const (
	digits = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func FToBaseStr(num float64, radix int) string {
	var negative bool
	if num < 0 {
		num = -num
		negative = true
	}

	dfloor := math.Floor(num)
	ldfloor := int64(dfloor)
	var intDigits string
	if dfloor == float64(ldfloor) {
		if negative {
			ldfloor = -ldfloor
		}
		intDigits = strconv.FormatInt(ldfloor, radix)
	} else {
		floorBits := math.Float64bits(num)
		exp := int(floorBits>>exp_shiftL) & exp_mask_shifted
		var mantissa int64
		if exp == 0 {
			mantissa = int64((floorBits & frac_maskL) << 1)
		} else {
			mantissa = int64((floorBits & frac_maskL) | exp_msk1L)
		}

		if negative {
			mantissa = -mantissa
		}
		exp -= 1075
		x := big.NewInt(mantissa)
		if exp > 0 {
			x.Lsh(x, uint(exp))
		} else if exp < 0 {
			x.Rsh(x, uint(-exp))
		}
		intDigits = x.Text(radix)
	}

	if num == dfloor {
		return intDigits
	} else {
		var buffer strings.Builder
		buffer.WriteString(intDigits)
		buffer.WriteByte('.')
		df := num - dfloor

		dBits := math.Float64bits(num)
		word0 := uint32(dBits >> 32)
		word1 := uint32(dBits)

		dblBits := make([]byte, 0, 8)
		e, _, dblBits := d2b(df, dblBits)

		/* 此时 df = b * 2^e，e 必须小于 0，因为 0 < df < 1 */
		s2 := -int((word0 >> exp_shift1) & (exp_mask >> exp_shift1))
		if s2 == 0 {
			s2 = -1
		}
		s2 += bias + p
		/* 1/2^s2 = (nextDouble(d) - d)/2 */
		if -s2 >= e {
			panic(fmt.Errorf("-s2 >= e: %d, %d", -s2, e))
		}
		mlo := big.NewInt(1)
		mhi := mlo
		if (word1 == 0) && ((word0 & bndry_mask) == 0) && ((word0 & (exp_mask & (exp_mask << 1))) != 0) {
			s2 += log2P
			mhi = big.NewInt(1 << log2P)
		}

		b := new(big.Int).SetBytes(dblBits)
		b.Lsh(b, uint(e+s2))
		s := big.NewInt(1)
		s.Lsh(s, uint(s2))

		/* s = 2^s2
		 * 1 > df = b/2^s2 > 0
		 * (d - prevDouble(d))/2 = mlo/2^s2
		 * (nextDouble(d) - d)/2 = mhi/2^s2
		 */
		bigBase := big.NewInt(int64(radix))

		done := false
		m := &big.Int{}
		delta := &big.Int{}
		for !done {
			b.Mul(b, bigBase)
			b.DivMod(b, s, m)
			digit := byte(b.Int64())
			b, m = m, b
			mlo.Mul(mlo, bigBase)
			if mlo != mhi {
				mhi.Mul(mhi, bigBase)
			}
			j := b.Cmp(mlo)
			delta.Sub(s, mhi)
			var j1 int
			if delta.Sign() <= 0 {
				j1 = 1
			} else {
				j1 = b.Cmp(delta)
			}
			if j1 == 0 && (word1&1) == 0 {
				if j > 0 {
					digit++
				}
				done = true
			} else if j < 0 || (j == 0 && ((word1 & 1) == 0)) {
				if j1 > 0 {
					b.Lsh(b, 1)
					j1 = b.Cmp(s)
					if j1 > 0 {
						digit++
					}
				}
				done = true
			} else if j1 > 0 {
				digit++
				done = true
			}
			buffer.WriteByte(digits[digit])
		}
		return buffer.String()
	}
}
