package ftoa

import (
	"math"
	"math/big"
)

const (
	exp_11     = 0x3ff00000
	frac_mask1 = 0xfffff
	bletch     = 0x10
	quick_max  = 14
	int_max    = 14
)

var (
	tens = [...]float64{
		1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
		1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
		1e20, 1e21, 1e22,
	}

	bigtens = [...]float64{1e16, 1e32, 1e64, 1e128, 1e256}

	big5  = big.NewInt(5)
	big10 = big.NewInt(10)

	p05       = []*big.Int{big5, big.NewInt(25), big.NewInt(125)}
	pow5Cache [7]*big.Int

	dtoaModes = []int{
		ModeStandard:            0,
		ModeStandardExponential: 0,
		ModeFixed:               3,
		ModeExponential:         2,
		ModePrecision:           2,
	}
)

/*
d 必须大于 0，并且不能是无穷大

模式:

	0 ==> 产生最短的字符串，并四舍五入为 d
	1 ==> 与 0 一样，但是有 Steele & White 停止规则
		例：在 IEEE P754 算术中 , 模式 0 输出 1e23，而模式 1 输出 9.999999999999999e22.
	2 ==> 从 1 和 ndigits 中取其大者作为有效数字. 它将给出一个与 ecvt 相似的输出，但是后面的 0 会受位数限制
	3 ==> 与 2 一样，但是此时 ndigits 可以是负数
	4,5 ==> 分别类似于 2 和 3 ，但在四舍五入模式通过对模式 0 的测试，可能会返回一个较短的字符串，四舍五入为 d
	6-9 ==> 调试模式，类似于模式 4 但是带有快速浮点估算（不要在正式开发中使用）

	如果模式值在 0-9 以外，将被当作 0 处理
*/
func ftoa(d float64, mode int, biasUp bool, ndigits int, buf []byte) ([]byte, int) {
	startPos := len(buf)
	dblBits := make([]byte, 0, 8)
	be, bbits, dblBits := d2b(d, dblBits)

	dBits := math.Float64bits(d)
	word0 := uint32(dBits >> 32)
	word1 := uint32(dBits)

	i := int((word0 >> exp_shift1) & (exp_mask >> exp_shift1))
	var d2 float64
	var denorm bool
	if i != 0 {
		d2 = setWord0(d, (word0&frac_mask1)|exp_11)
		i -= bias
		denorm = false
	} else {
		/* d 是补位的 */
		i = bbits + be + (bias + (p - 1) - 1)
		var x uint64
		if i > 32 {
			x = uint64(word0)<<(64-i) | uint64(word1)>>(i-32)
		} else {
			x = uint64(word1) << (32 - i)
		}
		d2 = setWord0(float64(x), uint32((x>>32)-31*exp_mask))
		i -= (bias + (p - 1) - 1) + 1
		denorm = true
	}
	/* 在这一点上，d = f*2^i，其中 1 <= f < 2，d2 是 f 的近似值 */
	ds := (d2-1.5)*0.289529654602168 + 0.1760912590558 + float64(i)*0.301029995663981
	k := int(ds)
	if ds < 0.0 && ds != float64(k) {
		k-- /* 预期 k = floor(ds) */
	}
	k_check := true
	if k >= 0 && k < len(tens) {
		if d < tens[k] {
			k--
		}
		k_check = false
	}
	/* 此时 floor(log10(d)) <= k <= floor(log10(d))+1，如果 k_check 为 0，就可以保证 k = floor(log10(d)) */
	j := bbits - i - 1
	var b2, s2, b5, s5 int
	/* 这里 d = b/2^j，其中 b 是一个奇数整数 */
	if j >= 0 {
		b2 = 0
		s2 = j
	} else {
		b2 = -j
		s2 = 0
	}
	if k >= 0 {
		b5 = 0
		s5 = k
		s2 += k
	} else {
		b2 -= k
		b5 = -k
		s5 = 0
	}
	/* 此时 d/10^k = (b * 2^b2 * 5^b5) / (2^s2 * 5^s5), 其中 b 是一个奇数整数，b2 >= 0, b5 >= 0, s2 >= 0, s5 >= 0 */
	if mode < 0 || mode > 9 {
		mode = 0
	}
	try_quick := true
	if mode > 5 {
		mode -= 4
		try_quick = false
	}
	leftright := true
	var ilim, ilim1 int
	switch mode {
	case 0, 1:
		ilim, ilim1 = -1, -1
		ndigits = 0
	case 2:
		leftright = false
		fallthrough
	case 4:
		if ndigits <= 0 {
			ndigits = 1
		}
		ilim, ilim1 = ndigits, ndigits
	case 3:
		leftright = false
		fallthrough
	case 5:
		i = ndigits + k + 1
		ilim = i
		ilim1 = i - 1
	}
	/* ilim 和 ilim1 是我们想要的最大有效位数，基于 k 和 ndigits。当发现 k 被计算得太高时，以 ilim1 限制的值为准 */
	fast_failed := false
	if ilim >= 0 && ilim <= quick_max && try_quick {
		/* 尝试用浮点算术来解决 */
		i = 0
		d2 = d
		k0 := k
		ilim0 := ilim
		ieps := 2 /* 保守值 */
		/* 用 d 除以 10^k, 保持跟踪舍入误差，避免溢出 */
		if k > 0 {
			ds = tens[k&0xf]
			j = k >> 4
			if (j & bletch) != 0 {
				/* 防止溢出 */
				j &= bletch - 1
				d /= bigtens[len(bigtens)-1]
				ieps++
			}
			for ; j != 0; i++ {
				if (j & 1) != 0 {
					ieps++
					ds *= bigtens[i]
				}
				j >>= 1
			}
			d /= ds
		} else if j1 := -k; j1 != 0 {
			d *= tens[j1&0xf]
			for j = j1 >> 4; j != 0; i++ {
				if (j & 1) != 0 {
					ieps++
					d *= bigtens[i]
				}
				j >>= 1
			}
		}
		/* 检查 k 的计算是否正确 */
		if k_check && d < 1.0 && ilim > 0 {
			if ilim1 <= 0 {
				fast_failed = true
			} else {
				ilim = ilim1
				k--
				d *= 10.
				ieps++
			}
		}
		/* 累积误差 eps 界限 */
		eps := float64(ieps)*d + 7.0
		eps = setWord0(eps, _word0(eps)-(p-1)*exp_msk1)
		if ilim == 0 {
			d -= 5.0
			if d > eps {
				buf = append(buf, '1')
				k++
				return buf, k + 1
			}
			if d < -eps {
				buf = append(buf, '0')
				return buf, 1
			}
			fast_failed = true
		}
		if !fast_failed {
			fast_failed = true
			if leftright {
				/* 使用 Steele & White 算法，只生成所需的数字 */
				eps = 0.5/tens[ilim-1] - eps
				for i = 0; ; {
					l := int64(d)
					d -= float64(l)
					buf = append(buf, byte('0'+l))
					if d < eps {
						return buf, k + 1
					}
					if 1.0-d < eps {
						buf, k = bumpUp(buf, k)
						return buf, k + 1
					}
					i++
					if i >= ilim {
						break
					}
					eps *= 10.0
					d *= 10.0
				}
			} else {
				/* 生成 ilim，然后修复它 */
				eps *= tens[ilim-1]
				for i = 1; ; i++ {
					l := int64(d)
					d -= float64(l)
					buf = append(buf, byte('0'+l))
					if i == ilim {
						if d > 0.5+eps {
							buf, k = bumpUp(buf, k)
							return buf, k + 1
						} else if d < 0.5-eps {
							buf = stripTrailingZeroes(buf, startPos)
							return buf, k + 1
						}
						break
					}
					d *= 10.0
				}
			}
		}
		if fast_failed {
			buf = buf[:startPos]
			d = d2
			k = k0
			ilim = ilim0
		}
	}

	/* 取一个小整数 */
	if be >= 0 && k <= int_max {
		ds = tens[k]
		if ndigits < 0 && ilim <= 0 {
			if ilim < 0 || d < 5*ds || (!biasUp && d == 5*ds) {
				buf = buf[:startPos]
				buf = append(buf, '0')
				return buf, 1
			}
			buf = append(buf, '1')
			k++
			return buf, k + 1
		}
		for i = 1; ; i++ {
			l := int64(d / ds)
			d -= float64(l) * ds
			buf = append(buf, byte('0'+l))
			if i == ilim {
				d += d
				if (d > ds) || (d == ds && (((l & 1) != 0) || biasUp)) {
					buf, k = bumpUp(buf, k)
				}
				break
			}
			d *= 10.0
			if d == 0 {
				break
			}
		}
		return buf, k + 1
	}

	m2 := b2
	m5 := b5
	var mhi, mlo *big.Int
	if leftright {
		if mode < 2 {
			if denorm {
				i = be + (bias + (p - 1) - 1 + 1)
			} else {
				i = 1 + p - bbits
			}
			/* i 为 1，加上 d 尾部的零位数，因此 (2^m2 * 5^m5) / (2^(s2+i) * 5^s5) = (1/2 lsb of d)/10^k */
		} else {
			j = ilim - 1
			if m5 >= j {
				m5 -= j
			} else {
				j -= m5
				s5 += j
				b5 += j
				m5 = 0
			}
			i = ilim
			if i < 0 {
				m2 -= i
				i = 0
			}
			/* (2^m2 * 5^m5) / (2^(s2+i) * 5^s5) = (1/2 * 10^(1-ilim))/10^k */
		}
		b2 += i
		s2 += i
		mhi = big.NewInt(1)
	}

	/* 我们仍有 d/10^k = (b * 2^b2 * 5^b5) / (2^s2 * 5^s5) 。在不改变等式的前提下，减少 b2, m2, s2 的公因数 */
	if m2 > 0 && s2 > 0 {
		if m2 < s2 {
			i = m2
		} else {
			i = s2
		}
		b2 -= i
		m2 -= i
		s2 -= i
	}

	b := new(big.Int).SetBytes(dblBits)
	/* 将 b5 折叠成 b m5 折叠成 mhi */
	if b5 > 0 {
		if leftright {
			if m5 > 0 {
				pow5mult(mhi, m5)
				b.Mul(mhi, b)
			}
			j = b5 - m5
			if j != 0 {
				pow5mult(b, j)
			}
		} else {
			pow5mult(b, b5)
		}
	}
	S := big.NewInt(1)
	if s5 > 0 {
		pow5mult(S, s5)
	}

	/* 检查 d 是 2 的归一化幂的特殊情况 */
	spec_case := false
	if mode < 2 {
		if (_word1(d) == 0) && ((_word0(d) & bndry_mask) == 0) && ((_word0(d) & (exp_mask & (exp_mask << 1))) != 0) {
			b2 += log2P
			s2 += log2P
			spec_case = true
		}
	}

	/* 为方便计算商数进行编排，必要时向左移，使除数有 4 个前导 0 位。
	此处有必要一劳永逸地计算 S 的前 28 位，并将它们和一个移位传递给 quorem，这样它就可以做移位和 ors 来计算 q 的分子。
	*/
	var zz int
	if s5 != 0 {
		S_bytes := S.Bytes()
		var S_hiWord uint32
		for idx := 0; idx < 4; idx++ {
			S_hiWord = S_hiWord << 8
			if idx < len(S_bytes) {
				S_hiWord |= uint32(S_bytes[idx])
			}
		}
		zz = 32 - hi0bits(S_hiWord)
	} else {
		zz = 1
	}
	i = (zz + s2) & 0x1f
	if i != 0 {
		i = 32 - i
	}
	/* i 是 S*2^s2 中前导 0 位的数量 */
	if i > 4 {
		i -= 4
		b2 += i
		m2 += i
		s2 += i
	} else if i < 4 {
		i += 28
		b2 += i
		m2 += i
		s2 += i
	}
	/* 现在 S*2^s2 中有了 4 个 0 位 */
	if b2 > 0 {
		b = b.Lsh(b, uint(b2))
	}
	if s2 > 0 {
		S.Lsh(S, uint(s2))
	}
	if k_check {
		if b.Cmp(S) < 0 {
			k--
			b.Mul(b, big10)
			if leftright {
				mhi.Mul(mhi, big10)
			}
			ilim = ilim1
		}
	}
	/* 此时 1 <= d/10^k = b/S < 10 */

	if ilim <= 0 && mode > 2 {
		/* 固定模式的输出，d 小于这个模式下的最小非零输出。输出零或最小非零输出，取决于哪个更接近于 d */
		if ilim >= 0 {
			i = b.Cmp(S.Mul(S, big5))
		}
		if ilim < 0 || i < 0 || i == 0 && !biasUp {
			/* 始终发出至少一个数字。如果在当前模式下，数字看起来是 0，则发射 '0'，并将解码器设置为 1 */
			buf = buf[:startPos]
			buf = append(buf, '0')
			return buf, 1
		}
		buf = append(buf, '1')
		k++
		return buf, k + 1
	}

	var dig byte
	if leftright {
		if m2 > 0 {
			mhi.Lsh(mhi, uint(m2))
		}

		/* 检查特殊情况，d 是 2 的归一化幂 */

		mlo = mhi
		if spec_case {
			mhi = mlo
			mhi = new(big.Int).Lsh(mhi, log2P)
		}

		var z, delta big.Int
		for i = 1; ; i++ {
			z.DivMod(b, S, b)
			dig = byte(z.Int64() + '0')
			j = b.Cmp(mlo)
			delta.Sub(S, mhi)
			var j1 int
			if delta.Sign() <= 0 {
				j1 = 1
			} else {
				j1 = b.Cmp(&delta)
			}
			if (j1 == 0) && (mode == 0) && ((_word1(d) & 1) == 0) {
				if dig == '9' {
					var flag bool
					buf = append(buf, '9')
					if buf, flag = roundOff(buf, startPos); flag {
						k++
						buf = append(buf, '1')
					}
					return buf, k + 1
				}
				if j > 0 {
					dig++
				}
				buf = append(buf, dig)
				return buf, k + 1
			}
			if (j < 0) || ((j == 0) && (mode == 0) && ((_word1(d) & 1) == 0)) {
				if j1 > 0 {
					/* 无论是 dig 还是 dig+1 都可以作为最小有效的小数位。使用哪个取决于哪个会产生更接近 d 的小数值 */
					b.Lsh(b, 1)
					j1 = b.Cmp(S)
					if (j1 > 0) || (j1 == 0 && (((dig & 1) == 1) || biasUp)) {
						dig++
						if dig == '9' {
							buf = append(buf, '9')
							buf, flag := roundOff(buf, startPos)
							if flag {
								k++
								buf = append(buf, '1')
							}
							return buf, k + 1
						}
					}
				}
				buf = append(buf, dig)
				return buf, k + 1
			}
			if j1 > 0 {
				if dig == '9' {
					buf = append(buf, '9')
					buf, flag := roundOff(buf, startPos)
					if flag {
						k++
						buf = append(buf, '1')
					}
					return buf, k + 1
				}
				buf = append(buf, dig+1)
				return buf, k + 1
			}
			buf = append(buf, dig)
			if i == ilim {
				break
			}
			b.Mul(b, big10)
			if mlo == mhi {
				mhi.Mul(mhi, big10)
			} else {
				mlo.Mul(mlo, big10)
				mhi.Mul(mhi, big10)
			}
		}
	} else {
		var z big.Int
		for i = 1; ; i++ {
			z.DivMod(b, S, b)
			dig = byte(z.Int64() + '0')
			buf = append(buf, dig)
			if i >= ilim {
				break
			}

			b.Mul(b, big10)
		}
	}
	b.Lsh(b, 1)
	j = b.Cmp(S)
	if (j > 0) || (j == 0 && (((dig & 1) == 1) || biasUp)) {
		var flag bool
		buf, flag = roundOff(buf, startPos)
		if flag {
			k++
			buf = append(buf, '1')
			return buf, k + 1
		}
	} else {
		buf = stripTrailingZeroes(buf, startPos)
	}

	return buf, k + 1
}

func bumpUp(buf []byte, k int) ([]byte, int) {
	var lastCh byte
	stop := 0
	if len(buf) > 0 && buf[0] == '-' {
		stop = 1
	}
	for {
		lastCh = buf[len(buf)-1]
		buf = buf[:len(buf)-1]
		if lastCh != '9' {
			break
		}
		if len(buf) == stop {
			k++
			lastCh = '0'
			break
		}
	}
	buf = append(buf, lastCh+1)
	return buf, k
}

func setWord0(d float64, w uint32) float64 {
	dBits := math.Float64bits(d)
	return math.Float64frombits(uint64(w)<<32 | dBits&0xffffffff)
}

func _word0(d float64) uint32 {
	dBits := math.Float64bits(d)
	return uint32(dBits >> 32)
}

func _word1(d float64) uint32 {
	dBits := math.Float64bits(d)
	return uint32(dBits)
}

func stripTrailingZeroes(buf []byte, startPos int) []byte {
	bl := len(buf) - 1
	for bl >= startPos && buf[bl] == '0' {
		bl--
	}
	return buf[:bl+1]
}

/* b = b * 5^k.  k 不能是负数 */
func pow5mult(b *big.Int, k int) *big.Int {
	if k < (1 << (len(pow5Cache) + 2)) {
		i := k & 3
		if i != 0 {
			b.Mul(b, p05[i-1])
		}
		k >>= 2
		i = 0
		for {
			if k&1 != 0 {
				b.Mul(b, pow5Cache[i])
			}
			k >>= 1
			if k == 0 {
				break
			}
			i++
		}
		return b
	}
	return b.Mul(b, new(big.Int).Exp(big5, big.NewInt(int64(k)), nil))
}

func roundOff(buf []byte, startPos int) ([]byte, bool) {
	i := len(buf)
	for i != startPos {
		i--
		if buf[i] != '9' {
			buf[i]++
			return buf[:i+1], false
		}
	}
	return buf[:startPos], true
}

func init() {
	p := big.NewInt(625)
	pow5Cache[0] = p
	for i := 1; i < len(pow5Cache); i++ {
		p = new(big.Int).Mul(p, p)
		pow5Cache[i] = p
	}
}
