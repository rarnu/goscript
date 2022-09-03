package fast

import (
	"fmt"
	"strconv"
)

const (
	kMinimalTargetExponent = -60
	kMaximalTargetExponent = -32

	kTen4 = 10000
	kTen5 = 100000
	kTen6 = 1000000
	kTen7 = 10000000
	kTen8 = 100000000
	kTen9 = 1000000000
)

type Mode int

const (
	ModeShortest Mode = iota
	ModePrecision
)

func roundWeed(buffer []byte, distance_too_high_w, unsafe_interval, rest, ten_kappa, unit uint64) bool {
	small_distance := distance_too_high_w - unit
	big_distance := distance_too_high_w + unit
	_DCHECK(rest <= unsafe_interval)
	for rest < small_distance && unsafe_interval-rest >= ten_kappa && (rest+ten_kappa < small_distance || small_distance-rest >= rest+ten_kappa-small_distance) {
		buffer[len(buffer)-1]--
		rest += ten_kappa
	}
	if rest < big_distance && unsafe_interval-rest >= ten_kappa && (rest+ten_kappa < big_distance || big_distance-rest > rest+ten_kappa-big_distance) {
		return false
	}
	return (2*unit <= rest) && (rest <= unsafe_interval-4*unit)
}

func roundWeedCounted(buffer []byte, rest, ten_kappa, unit uint64, kappa *int) bool {
	_DCHECK(rest < ten_kappa)
	if unit >= ten_kappa {
		return false
	}
	if ten_kappa-unit <= unit {
		return false
	}
	if (ten_kappa-rest > rest) && (ten_kappa-2*rest >= 2*unit) {
		return true
	}
	if (rest > unit) && (ten_kappa-(rest-unit) <= (rest - unit)) {
		buffer[len(buffer)-1]++
		for i := len(buffer) - 1; i > 0; i-- {
			if buffer[i] != '0'+10 {
				break
			}
			buffer[i] = '0'
			buffer[i-1]++
		}
		if buffer[0] == '0'+10 {
			buffer[0] = '1'
			*kappa += 1
		}
		return true
	}
	return false
}

func biggestPowerTen(number uint32, number_bits int) (power uint32, exponent int) {
	switch number_bits {
	case 32, 31, 30:
		if kTen9 <= number {
			power = kTen9
			exponent = 9
			break
		}
		fallthrough
	case 29, 28, 27:
		if kTen8 <= number {
			power = kTen8
			exponent = 8
			break
		}
		fallthrough
	case 26, 25, 24:
		if kTen7 <= number {
			power = kTen7
			exponent = 7
			break
		}
		fallthrough
	case 23, 22, 21, 20:
		if kTen6 <= number {
			power = kTen6
			exponent = 6
			break
		}
		fallthrough
	case 19, 18, 17:
		if kTen5 <= number {
			power = kTen5
			exponent = 5
			break
		}
		fallthrough
	case 16, 15, 14:
		if kTen4 <= number {
			power = kTen4
			exponent = 4
			break
		}
		fallthrough
	case 13, 12, 11, 10:
		if 1000 <= number {
			power = 1000
			exponent = 3
			break
		}
		fallthrough
	case 9, 8, 7:
		if 100 <= number {
			power = 100
			exponent = 2
			break
		}
		fallthrough
	case 6, 5, 4:
		if 10 <= number {
			power = 10
			exponent = 1
			break
		}
		fallthrough
	case 3, 2, 1:
		if 1 <= number {
			power = 1
			exponent = 0
			break
		}
		fallthrough
	case 0:
		power = 0
		exponent = -1
	}
	return
}

func digitGen(low, w, high diyfp, buffer []byte) (kappa int, buf []byte, res bool) {
	_DCHECK(low.e == w.e && w.e == high.e)
	_DCHECK(low.f+1 <= high.f-1)
	_DCHECK(kMinimalTargetExponent <= w.e && w.e <= kMaximalTargetExponent)
	unit := uint64(1)
	too_low := diyfp{f: low.f - unit, e: low.e}
	too_high := diyfp{f: high.f + unit, e: high.e}
	unsafe_interval := too_high.minus(too_low)
	one := diyfp{f: 1 << -w.e, e: w.e}
	integrals := uint32(too_high.f >> -one.e)
	fractionals := too_high.f & (one.f - 1)
	divisor, divisor_exponent := biggestPowerTen(integrals, diyFpKSignificandSize-(-one.e))
	kappa = divisor_exponent + 1
	buf = buffer
	for kappa > 0 {
		digit := int(integrals / divisor)
		buf = append(buf, byte('0'+digit))
		integrals %= divisor
		kappa--
		rest := uint64(integrals)<<-one.e + fractionals
		if rest < unsafe_interval.f {
			res = roundWeed(buf, too_high.minus(w).f,
				unsafe_interval.f, rest,
				uint64(divisor)<<-one.e, unit)
			return
		}
		divisor /= 10
	}
	_DCHECK(one.e >= -60)
	_DCHECK(fractionals < one.f)
	_DCHECK(0xFFFFFFFFFFFFFFFF/10 >= one.f)
	for {
		fractionals *= 10
		unit *= 10
		unsafe_interval.f *= 10
		digit := byte(fractionals >> -one.e)
		buf = append(buf, '0'+digit)
		fractionals &= one.f - 1
		kappa--
		if fractionals < unsafe_interval.f {
			res = roundWeed(buf, too_high.minus(w).f*unit, unsafe_interval.f, fractionals, one.f, unit)
			return
		}
	}
}

func digitGenCounted(w diyfp, requested_digits int, buffer []byte) (kappa int, buf []byte, res bool) {
	_DCHECK(kMinimalTargetExponent <= w.e && w.e <= kMaximalTargetExponent)
	w_error := uint64(1)
	one := diyfp{f: 1 << -w.e, e: w.e}
	integrals := uint32(w.f >> -one.e)
	fractionals := w.f & (one.f - 1)
	divisor, divisor_exponent := biggestPowerTen(integrals, diyFpKSignificandSize-(-one.e))
	kappa = divisor_exponent + 1
	buf = buffer
	for kappa > 0 {
		digit := byte(integrals / divisor)
		buf = append(buf, '0'+digit)
		requested_digits--
		integrals %= divisor
		kappa--
		if requested_digits == 0 {
			break
		}
		divisor /= 10
	}
	if requested_digits == 0 {
		rest := uint64(integrals)<<-one.e + fractionals
		res = roundWeedCounted(buf, rest, uint64(divisor)<<-one.e, w_error, &kappa)
		return
	}
	_DCHECK(one.e >= -60)
	_DCHECK(fractionals < one.f)
	_DCHECK(0xFFFFFFFFFFFFFFFF/10 >= one.f)
	for requested_digits > 0 && fractionals > w_error {
		fractionals *= 10
		w_error *= 10
		digit := byte(fractionals >> -one.e)
		buf = append(buf, '0'+digit)
		requested_digits--
		fractionals &= one.f - 1
		kappa--
	}
	if requested_digits != 0 {
		res = false
	} else {
		res = roundWeedCounted(buf, fractionals, one.f, w_error, &kappa)
	}
	return
}

func grisu3(f float64, buffer []byte) (digits []byte, decimal_exponent int, result bool) {
	v := double(f)
	w := v.toNormalizedDiyfp()
	boundary_minus, boundary_plus := v.normalizedBoundaries()
	ten_mk_minimal_binary_exponent := kMinimalTargetExponent - (w.e + diyFpKSignificandSize)
	ten_mk_maximal_binary_exponent := kMaximalTargetExponent - (w.e + diyFpKSignificandSize)
	ten_mk, mk := getCachedPowerForBinaryExponentRange(ten_mk_minimal_binary_exponent, ten_mk_maximal_binary_exponent)
	_DCHECK((kMinimalTargetExponent <= w.e+ten_mk.e+diyFpKSignificandSize) && (kMaximalTargetExponent >= w.e+ten_mk.e+diyFpKSignificandSize))
	scaled_w := w.times(ten_mk)
	_DCHECK(scaled_w.e == boundary_plus.e+ten_mk.e+diyFpKSignificandSize)
	scaled_boundary_minus := boundary_minus.times(ten_mk)
	scaled_boundary_plus := boundary_plus.times(ten_mk)
	var kappa int
	kappa, digits, result = digitGen(scaled_boundary_minus, scaled_w, scaled_boundary_plus, buffer)
	decimal_exponent = -mk + kappa
	return
}

func grisu3Counted(v float64, requested_digits int, buffer []byte) (digits []byte, decimal_exponent int, result bool) {
	w := double(v).toNormalizedDiyfp()
	ten_mk_minimal_binary_exponent := kMinimalTargetExponent - (w.e + diyFpKSignificandSize)
	ten_mk_maximal_binary_exponent := kMaximalTargetExponent - (w.e + diyFpKSignificandSize)
	ten_mk, mk := getCachedPowerForBinaryExponentRange(ten_mk_minimal_binary_exponent, ten_mk_maximal_binary_exponent)
	_DCHECK((kMinimalTargetExponent <= w.e+ten_mk.e+diyFpKSignificandSize) && (kMaximalTargetExponent >= w.e+ten_mk.e+diyFpKSignificandSize))
	scaled_w := w.times(ten_mk)
	var kappa int
	kappa, digits, result = digitGenCounted(scaled_w, requested_digits, buffer)
	decimal_exponent = -mk + kappa

	return
}

func Dtoa(v float64, mode Mode, requested_digits int, buffer []byte) (digits []byte, decimal_point int, result bool) {
	defer func() {
		if x := recover(); x != nil {
			if x == dcheckFailure {
				panic(fmt.Errorf("DCHECK assertion failed while formatting %s in mode %d", strconv.FormatFloat(v, 'e', 50, 64), mode))
			}
			panic(x)
		}
	}()
	var decimal_exponent int
	startPos := len(buffer)
	switch mode {
	case ModeShortest:
		digits, decimal_exponent, result = grisu3(v, buffer)
	case ModePrecision:
		digits, decimal_exponent, result = grisu3Counted(v, requested_digits, buffer)
	}
	if result {
		decimal_point = len(digits) - startPos + decimal_exponent
	} else {
		digits = digits[:startPos]
	}
	return
}
