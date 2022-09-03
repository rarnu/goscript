package goscript

import (
	"errors"
	"time"
)

const (
	_                        = iota
	stdLongMonth             = iota + stdNeedDate  // "January"
	stdMonth                                       // "Jan"
	stdNumMonth                                    // "1"
	stdZeroMonth                                   // "01"
	stdLongWeekDay                                 // "Monday"
	stdWeekDay                                     // "Mon"
	stdDay                                         // "2"
	stdUnderDay                                    // "_2"
	stdZeroDay                                     // "02"
	stdHour                  = iota + stdNeedClock // "15"
	stdHour12                                      // "3"
	stdZeroHour12                                  // "03"
	stdMinute                                      // "4"
	stdZeroMinute                                  // "04"
	stdSecond                                      // "5"
	stdZeroSecond                                  // "05"
	stdLongYear              = iota + stdNeedDate  // "2006"
	stdYear                                        // "06"
	stdPM                    = iota + stdNeedClock // "PM"
	stdpm                                          // "pm"
	stdTZ                    = iota                // "MST"
	stdBracketTZ                                   // "(MST)"
	stdISO8601TZ                                   // "Z0700"  // UTC 下打印 Z
	stdISO8601SecondsTZ                            // "Z070000"
	stdISO8601ShortTZ                              // "Z07"
	stdISO8601ColonTZ                              // "Z07:00" // UTC 下打印 Z
	stdISO8601ColonSecondsTZ                       // "Z07:00:00"
	stdNumTZ                                       // "-0700"  // 总是数字
	stdNumSecondsTz                                // "-070000"
	stdNumShortTZ                                  // "-07"    // 总是数字
	stdNumColonTZ                                  // "-07:00" // 总是数字
	stdNumColonSecondsTZ                           // "-07:00:00"
	stdFracSecond0                                 // ".0", ".00", ... , 包括尾数 0
	stdFracSecond9                                 // ".9", ".99", ..., 省略尾数 0
	stdNeedDate              = 1 << 8              // 需要 月，日，年
	stdNeedClock             = 2 << 8              // 需要 时，分，秒
	stdArgShift              = 16                  // 高位的额外参数，高于低位 stdArgShift 偏移
	stdMask                  = 1<<stdArgShift - 1
)

var errBad = errors.New("bad value for field") // 占位符，不会传递给用户

func parseDate(layout, value string, defaultLocation *time.Location) (time.Time, error) {
	alayout, avalue := layout, value
	rangeErrString := "" // 如果值超出范围，则设置此信息
	amSet := false
	pmSet := false
	var (
		year       int
		month      int = 1 // January
		day        int = 1
		hour       int
		min        int
		sec        int
		nsec       int
		z          *time.Location
		zoneOffset int = -1
		zoneName   string
	)

	for {
		var err error
		prefix, std, suffix := nextStdChunk(layout)
		stdstr := layout[len(prefix) : len(layout)-len(suffix)]
		value, err = skip(value, prefix)
		if err != nil {
			return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, LayoutElem: prefix, ValueElem: value}
		}
		if std == 0 {
			if len(value) != 0 {
				return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, ValueElem: value, Message: ": extra text: " + value}
			}
			break
		}
		layout = suffix
		var p string
		switch std & stdMask {
		case stdYear:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			year, err = atoi(p)
			if year >= 69 {
				// 在某些时区，Unix 时间始于 1969.12.31
				year += 1900
			} else {
				year += 2000
			}
		case stdLongYear:
			if len(value) >= 7 && (value[0] == '-' || value[0] == '+') {
				neg := value[0] == '-'
				p, value = value[1:7], value[7:]
				year, err = atoi(p)
				if neg {
					year = -year
				}
			} else {
				if len(value) < 4 || !isDigit(value, 0) {
					err = errBad
					break
				}
				p, value = value[0:4], value[4:]
				year, err = atoi(p)
			}

		case stdMonth:
			month, value, err = lookup(longMonthNames, value)
			if err != nil {
				month, value, err = lookup(shortMonthNames, value)
			}
			month++
		case stdLongMonth:
			month, value, err = lookup(longMonthNames, value)
			month++
		case stdNumMonth, stdZeroMonth:
			month, value, err = getnum(value, std == stdZeroMonth)
			if month <= 0 || 12 < month {
				rangeErrString = "month"
			}
		case stdWeekDay:
			_, value, err = lookup(longDayNames, value)
			if err != nil {
				_, value, err = lookup(shortDayNames, value)
			}
		case stdLongWeekDay:
			_, value, err = lookup(longDayNames, value)
		case stdDay, stdUnderDay, stdZeroDay:
			if std == stdUnderDay && len(value) > 0 && value[0] == ' ' {
				value = value[1:]
			}
			day, value, err = getnum(value, false)
			if day < 0 {
				rangeErrString = "day"
			}
		case stdHour:
			hour, value, err = getnum(value, false)
			if hour < 0 || 24 <= hour {
				rangeErrString = "hour"
			}
		case stdHour12, stdZeroHour12:
			hour, value, err = getnum(value, std == stdZeroHour12)
			if hour < 0 || 12 < hour {
				rangeErrString = "hour"
			}
		case stdMinute, stdZeroMinute:
			min, value, err = getnum(value, std == stdZeroMinute)
			if min < 0 || 60 <= min {
				rangeErrString = "minute"
			}
		case stdSecond, stdZeroSecond:
			sec, value, err = getnum(value, std == stdZeroSecond)
			if sec < 0 || 60 <= sec {
				rangeErrString = "second"
				break
			}
			// 特殊情况，当时间内有小数秒，但是格式中没有小数秒
			if len(value) >= 2 && value[0] == '.' && isDigit(value, 1) {
				_, std, _ = nextStdChunk(layout)
				std &= stdMask
				if std == stdFracSecond0 || std == stdFracSecond9 {
					break
				}
				// 在布局中没有小数秒，但在输入中有
				n := 2
				for ; n < len(value) && isDigit(value, n); n++ {
				}
				nsec, rangeErrString, err = parseNanoseconds(value, n)
				value = value[n:]
			}
		case stdPM:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			switch p {
			case "PM":
				pmSet = true
			case "AM":
				amSet = true
			default:
				err = errBad
			}
		case stdpm:
			if len(value) < 2 {
				err = errBad
				break
			}
			p, value = value[0:2], value[2:]
			switch p {
			case "pm":
				pmSet = true
			case "am":
				amSet = true
			default:
				err = errBad
			}
		case stdISO8601TZ, stdISO8601ColonTZ, stdISO8601SecondsTZ, stdISO8601ShortTZ, stdISO8601ColonSecondsTZ, stdNumTZ, stdNumShortTZ, stdNumColonTZ, stdNumSecondsTz, stdNumColonSecondsTZ:
			if (std == stdISO8601TZ || std == stdISO8601ShortTZ || std == stdISO8601ColonTZ || std == stdISO8601SecondsTZ || std == stdISO8601ColonSecondsTZ) && len(value) >= 1 && value[0] == 'Z' {
				value = value[1:]
				z = time.UTC
				break
			}
			var sign, hour, min, seconds string
			if std == stdISO8601ColonTZ || std == stdNumColonTZ || std == stdNumTZ || std == stdISO8601TZ {
				if len(value) < 4 {
					err = errBad
					break
				}
				if value[3] != ':' {
					if std == stdNumColonTZ || std == stdISO8601ColonTZ || len(value) < 5 {
						err = errBad
						break
					}
					sign, hour, min, seconds, value = value[0:1], value[1:3], value[3:5], "00", value[5:]
				} else {
					if len(value) < 6 {
						err = errBad
						break
					}
					sign, hour, min, seconds, value = value[0:1], value[1:3], value[4:6], "00", value[6:]
				}
			} else if std == stdNumShortTZ || std == stdISO8601ShortTZ {
				if len(value) < 3 {
					err = errBad
					break
				}
				sign, hour, min, seconds, value = value[0:1], value[1:3], "00", "00", value[3:]
			} else if std == stdISO8601ColonSecondsTZ || std == stdNumColonSecondsTZ || std == stdISO8601SecondsTZ || std == stdNumSecondsTz {
				if len(value) < 7 {
					err = errBad
					break
				}
				if value[3] != ':' || value[6] != ':' {
					if std == stdISO8601ColonSecondsTZ || std == stdNumColonSecondsTZ || len(value) < 7 {
						err = errBad
						break
					}
					sign, hour, min, seconds, value = value[0:1], value[1:3], value[3:5], value[5:7], value[7:]
				} else {
					if len(value) < 9 {
						err = errBad
						break
					}
					sign, hour, min, seconds, value = value[0:1], value[1:3], value[4:6], value[7:9], value[9:]
				}
			}
			var hr, mm, ss int
			hr, err = atoi(hour)
			if err == nil {
				mm, err = atoi(min)
			}
			if err == nil {
				ss, err = atoi(seconds)
			}
			zoneOffset = (hr*60+mm)*60 + ss // 偏移量的单位：秒
			switch sign[0] {
			case '+':
			case '-':
				zoneOffset = -zoneOffset
			default:
				err = errBad
			}
		case stdTZ:
			if len(value) >= 3 && value[0:3] == "UTC" {
				z = time.UTC
				value = value[3:]
				break
			}
			n, ok := parseTimeZone(value)
			if !ok {
				err = errBad
				break
			}
			zoneName, value = value[:n], value[n:]
		case stdBracketTZ:
			if len(value) < 3 || value[0] != '(' {
				err = errBad
				break
			}
			i := 1
			for ; ; i++ {
				if i >= len(value) {
					err = errBad
					break
				}
				if value[i] == ')' {
					zoneName, value = value[1:i], value[i+1:]
					break
				}
			}

		case stdFracSecond0:
			// stdFracSecond0 要求包含布局中指定的精确数字
			ndigit := 1 + (std >> stdArgShift)
			if len(value) < ndigit {
				err = errBad
				break
			}
			nsec, rangeErrString, err = parseNanoseconds(value, ndigit)
			value = value[ndigit:]

		case stdFracSecond9:
			if len(value) < 2 || value[0] != '.' || value[1] < '0' || '9' < value[1] {
				// 省略小数秒
				break
			}
			i := 0
			for i < 9 && i+1 < len(value) && '0' <= value[i+1] && value[i+1] <= '9' {
				i++
			}
			nsec, rangeErrString, err = parseNanoseconds(value, 1+i)
			value = value[1+i:]
		}
		if rangeErrString != "" {
			return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, LayoutElem: stdstr, ValueElem: value, Message: ": " + rangeErrString + " out of range"}
		}
		if err != nil {
			return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, LayoutElem: stdstr, ValueElem: value}
		}
	}
	if pmSet && hour < 12 {
		hour += 12
	} else if amSet && hour == 12 {
		hour = 0
	}

	// 验证每月的日期
	if day < 1 || day > daysIn(time.Month(month), year) {
		return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, ValueElem: value, Message: ": day out of range"}
	}

	if z == nil {
		if zoneOffset == -1 {
			if zoneName != "" {
				if z1, err := time.LoadLocation(zoneName); err == nil {
					z = z1
				} else {
					return time.Time{}, &time.ParseError{Layout: alayout, Value: avalue, ValueElem: value, Message: ": unknown timezone"}
				}
			} else {
				z = defaultLocation
			}
		} else if zoneOffset == 0 {
			z = time.UTC
		} else {
			z = time.FixedZone("", zoneOffset)
		}
	}

	return time.Date(year, time.Month(month), day, hour, min, sec, nsec, z), nil
}

var errLeadingInt = errors.New("time: bad [0-9]*")

func signedLeadingInt(s string) (x int64, rem string, err error) {
	neg := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		s = s[1:]
	}
	x, rem, err = leadingInt(s)
	if err != nil {
		return
	}

	if neg {
		x = -x
	}
	return
}

// leadingInt 消费 s 前导的 [0-9]*
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// nextStdChunk 找到布局中第一个出现的 std 字符串，并返回之前的文本、std 字符串和之后的文本
func nextStdChunk(layout string) (prefix string, std int, suffix string) {
	for i := 0; i < len(layout); i++ {
		switch c := int(layout[i]); c {
		case 'J': // January, Jan
			if len(layout) >= i+3 && layout[i:i+3] == "Jan" {
				if len(layout) >= i+7 && layout[i:i+7] == "January" {
					return layout[0:i], stdLongMonth, layout[i+7:]
				}
				if !startsWithLowerCase(layout[i+3:]) {
					return layout[0:i], stdMonth, layout[i+3:]
				}
			}
		case 'M': // Monday, Mon, MST
			if len(layout) >= i+3 {
				if layout[i:i+3] == "Mon" {
					if len(layout) >= i+6 && layout[i:i+6] == "Monday" {
						return layout[0:i], stdLongWeekDay, layout[i+6:]
					}
					if !startsWithLowerCase(layout[i+3:]) {
						return layout[0:i], stdWeekDay, layout[i+3:]
					}
				}
				if layout[i:i+3] == "MST" {
					return layout[0:i], stdTZ, layout[i+3:]
				}
			}
		case '0': // 01, 02, 03, 04, 05, 06
			if len(layout) >= i+2 && '1' <= layout[i+1] && layout[i+1] <= '6' {
				return layout[0:i], std0x[layout[i+1]-'1'], layout[i+2:]
			}
		case '1': // 15, 1
			if len(layout) >= i+2 && layout[i+1] == '5' {
				return layout[0:i], stdHour, layout[i+2:]
			}
			return layout[0:i], stdNumMonth, layout[i+1:]
		case '2': // 2006, 2
			if len(layout) >= i+4 && layout[i:i+4] == "2006" {
				return layout[0:i], stdLongYear, layout[i+4:]
			}
			return layout[0:i], stdDay, layout[i+1:]
		case '_': // _2, _2006
			if len(layout) >= i+2 && layout[i+1] == '2' {
				if len(layout) >= i+5 && layout[i+1:i+5] == "2006" {
					return layout[0 : i+1], stdLongYear, layout[i+5:]
				}
				return layout[0:i], stdUnderDay, layout[i+2:]
			}
		case '3':
			return layout[0:i], stdHour12, layout[i+1:]
		case '4':
			return layout[0:i], stdMinute, layout[i+1:]
		case '5':
			return layout[0:i], stdSecond, layout[i+1:]
		case 'P': // PM
			if len(layout) >= i+2 && layout[i+1] == 'M' {
				return layout[0:i], stdPM, layout[i+2:]
			}
		case 'p': // pm
			if len(layout) >= i+2 && layout[i+1] == 'm' {
				return layout[0:i], stdpm, layout[i+2:]
			}
		case '-': // -070000, -07:00:00, -0700, -07:00, -07
			if len(layout) >= i+7 && layout[i:i+7] == "-070000" {
				return layout[0:i], stdNumSecondsTz, layout[i+7:]
			}
			if len(layout) >= i+9 && layout[i:i+9] == "-07:00:00" {
				return layout[0:i], stdNumColonSecondsTZ, layout[i+9:]
			}
			if len(layout) >= i+5 && layout[i:i+5] == "-0700" {
				return layout[0:i], stdNumTZ, layout[i+5:]
			}
			if len(layout) >= i+6 && layout[i:i+6] == "-07:00" {
				return layout[0:i], stdNumColonTZ, layout[i+6:]
			}
			if len(layout) >= i+3 && layout[i:i+3] == "-07" {
				return layout[0:i], stdNumShortTZ, layout[i+3:]
			}
		case 'Z': // Z070000, Z07:00:00, Z0700, Z07:00,
			if len(layout) >= i+7 && layout[i:i+7] == "Z070000" {
				return layout[0:i], stdISO8601SecondsTZ, layout[i+7:]
			}
			if len(layout) >= i+9 && layout[i:i+9] == "Z07:00:00" {
				return layout[0:i], stdISO8601ColonSecondsTZ, layout[i+9:]
			}
			if len(layout) >= i+5 && layout[i:i+5] == "Z0700" {
				return layout[0:i], stdISO8601TZ, layout[i+5:]
			}
			if len(layout) >= i+6 && layout[i:i+6] == "Z07:00" {
				return layout[0:i], stdISO8601ColonTZ, layout[i+6:]
			}
			if len(layout) >= i+3 && layout[i:i+3] == "Z07" {
				return layout[0:i], stdISO8601ShortTZ, layout[i+3:]
			}

		case '.': // .000 或 .999 - 小数秒的重复数字
			if i+1 < len(layout) && (layout[i+1] == '0' || layout[i+1] == '9') {
				ch := layout[i+1]
				j := i + 1
				for j < len(layout) && layout[j] == ch {
					j++
				}
				// 数字字符串必须在此结束，只有第二位小数之前是全部数字
				if !isDigit(layout, j) {
					std := stdFracSecond0
					if layout[i+1] == '9' {
						std = stdFracSecond9
					}
					std |= (j - (i + 1)) << stdArgShift
					return layout[0:i], std, layout[j:]
				}
			}
		case '(':
			if len(layout) >= i+5 && layout[i:i+5] == "(MST)" {
				return layout[0:i], stdBracketTZ, layout[i+5:]
			}
		}
	}
	return layout, 0, ""
}

var longDayNames = []string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

var shortDayNames = []string{
	"Sun",
	"Mon",
	"Tue",
	"Wed",
	"Thu",
	"Fri",
	"Sat",
}

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Nov",
	"Dec",
}

var longMonthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

// isDigit 判断 s[i] 是否在范围内并且是小数位
func isDigit(s string, i int) bool {
	if len(s) <= i {
		return false
	}
	c := s[i]
	return '0' <= c && c <= '9'
}

// getnum 将 s[0:1] 或 s[0:2]（固定为后者）解析为十进制整数，并返回该整数和字符串的剩余部分
func getnum(s string, fixed bool) (int, string, error) {
	if !isDigit(s, 0) {
		return 0, s, errBad
	}
	if !isDigit(s, 1) {
		if fixed {
			return 0, s, errBad
		}
		return int(s[0] - '0'), s[1:], nil
	}
	return int(s[0]-'0')*10 + int(s[1]-'0'), s[2:], nil
}

func cutspace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	return s
}

// 从值中删除给定的前缀，和多余的空格
func skip(value, prefix string) (string, error) {
	for len(prefix) > 0 {
		if prefix[0] == ' ' {
			if len(value) > 0 && value[0] != ' ' {
				return value, errBad
			}
			prefix = cutspace(prefix)
			value = cutspace(value)
			continue
		}
		if len(value) == 0 || value[0] != prefix[0] {
			return value, errBad
		}
		prefix = prefix[1:]
		value = value[1:]
	}
	return value, nil
}

var atoiError = errors.New("time: invalid number")

func atoi(s string) (x int, err error) {
	q, rem, err := signedLeadingInt(s)
	x = int(q)
	if err != nil || rem != "" {
		return 0, atoiError
	}
	return x, nil
}

// match 判断 s1 和 s2 是否匹配（无视大小写）
// 假设 s1 和 s2 的长度相等
func match(s1, s2 string) bool {
	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]
		if c1 != c2 {
			c1 |= 'a' - 'A'
			c2 |= 'a' - 'A'
			if c1 != c2 || c1 < 'a' || c1 > 'z' {
				return false
			}
		}
	}
	return true
}

func lookup(tab []string, val string) (int, string, error) {
	for i, v := range tab {
		if len(val) >= len(v) && match(val[0:len(v)], v) {
			return i, val[len(v):], nil
		}
	}
	return -1, val, errBad
}

// daysBefore 计算非闰年中 m 月开始前的天数
var daysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysIn(m time.Month, year int) int {
	if m == time.February && isLeap(year) {
		return 29
	}
	return int(daysBefore[m] - daysBefore[m-1])
}

func parseTimeZone(value string) (length int, ok bool) {
	if len(value) < 3 {
		return 0, false
	}
	if len(value) >= 4 && (value[:4] == "ChST" || value[:4] == "MeST") {
		return 4, true
	}
	if value[:3] == "GMT" {
		length = parseGMT(value)
		return length, true
	}
	if value[0] == '+' || value[0] == '-' {
		length = parseSignedOffset(value)
		return length, true
	}
	var nUpper int
	for nUpper = 0; nUpper < 6; nUpper++ {
		if nUpper >= len(value) {
			break
		}
		if c := value[nUpper]; c < 'A' || 'Z' < c {
			break
		}
	}
	switch nUpper {
	case 0, 1, 2, 6:
		return 0, false
	case 5:
		if value[4] == 'T' {
			return 5, true
		}
	case 4:
		if value[3] == 'T' || value[:4] == "WITA" {
			return 4, true
		}
	case 3:
		return 3, true
	}
	return 0, false
}

func parseGMT(value string) int {
	value = value[3:]
	if len(value) == 0 {
		return 3
	}

	return 3 + parseSignedOffset(value)
}

func parseSignedOffset(value string) int {
	sign := value[0]
	if sign != '-' && sign != '+' {
		return 0
	}
	x, rem, err := leadingInt(value[1:])
	if err != nil {
		return 0
	}
	if sign == '-' {
		x = -x
	}
	if x == 0 || x < -14 || 12 < x {
		return 0
	}
	return len(value) - len(rem)
}

func parseNanoseconds(value string, nbytes int) (ns int, rangeErrString string, err error) {
	if value[0] != '.' {
		err = errBad
		return
	}
	if ns, err = atoi(value[1:nbytes]); err != nil {
		return
	}
	if ns < 0 || 1e9 <= ns {
		rangeErrString = "fractional second"
		return
	}
	scaleDigits := 10 - nbytes
	for i := 0; i < scaleDigits; i++ {
		ns *= 10
	}
	return
}

var std0x = [...]int{stdZeroMonth, stdZeroDay, stdZeroHour12, stdZeroMinute, stdZeroSecond, stdYear}

func startsWithLowerCase(str string) bool {
	if len(str) == 0 {
		return false
	}
	c := str[0]
	return 'a' <= c && c <= 'z'
}
