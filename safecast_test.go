package safecast_test

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"fortio.org/safecast"
)

// TODO: steal the tests from https://github.com/ccoVeille/go-safecast

const all64bitsOne = ^uint64(0)

// Interesting part is the "true" for the first line, which is why we have to change the
// code in Convert to handle that 1 special case.
// safecast_test.go:22: bits 64: 1111111111111111111111111111111111111111111111111111111111111111
// : 18446744073709551615 -> float64 18446744073709551616 true.
func FindNumIntBits[T safecast.Float](t *testing.T) int {
	var v T
	for i := 0; i < 64; i++ {
		bits := all64bitsOne >> i
		v = T(bits)
		t.Logf("bits %02d: %b : %d -> %T %.0f %t", 64-i, bits, bits, v, v, uint64(v) == bits)
		if v != v-1 {
			return 64 - i
		}
	}
	panic("bug... didn't fine num bits")
}

// https://en.wikipedia.org/wiki/Double-precision_floating-point_format
const expectedFloat64Bits = 53

// https://en.wikipedia.org/wiki/Single-precision_floating-point_format#IEEE_754_standard:_binary32
const expectedFloat32Bits = 24

func TestFloat32Bounds(t *testing.T) {
	float32bits := FindNumIntBits[float32](t)
	t.Logf("float32: %d bits", float32bits)
	if float32bits != expectedFloat32Bits {
		t.Errorf("unexpected number of bits: %d", float32bits)
	}
	float32int := uint64(1<<(float32bits) - 1) // 24 bits
	for i := 0; i <= 64-float32bits; i++ {
		t.Logf("float32int %b %d", float32int, float32int)
		f := safecast.MustConvert[float32](float32int)
		t.Logf("float32int -> %.0f", f)
		float32int <<= 1
	}
}

func TestFloat64Bounds(t *testing.T) {
	float64bits := FindNumIntBits[float64](t)
	t.Logf("float64: %d bits", float64bits)
	float64int := uint64(1<<(float64bits) - 1) // 53 bits
	if float64bits != expectedFloat64Bits {
		t.Errorf("unexpected number of bits: %d", float64bits)
	}
	for i := 0; i <= 64-float64bits; i++ {
		t.Logf("float64int %b %d", float64int, float64int)
		f := safecast.MustConvert[float64](float64int)
		t.Logf("float64int -> %.0f", f)
		float64int <<= 1
	}
}

func TestNonIntegerFloat(t *testing.T) {
	_, err := safecast.Convert[int](math.Pi)
	if err == nil {
		t.Errorf("expected error")
	}
	truncPi := math.Trunc(math.Pi) // math.Trunc returns a float64
	i, err := safecast.Convert[int](truncPi)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i != 3 {
		t.Errorf("unexpected value: %v", i)
	}
	i, err = safecast.Truncate[int](math.Pi)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i != 3 {
		t.Errorf("unexpected value: %v", i)
	}
	i, err = safecast.Round[int](math.Phi)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if i != 2 {
		t.Errorf("unexpected value: %v", i)
	}
}

// MaxUint64 special case and also MaxInt64+1.
func TestMaxInt64(t *testing.T) {
	f32, err := safecast.Convert[float32](all64bitsOne)
	if err == nil {
		t.Errorf("expected error, got %d -> %.0f", all64bitsOne, f32)
	}
	f64, err := safecast.Convert[float64](all64bitsOne)
	if err == nil {
		t.Errorf("expected error, got %d -> %.0f", all64bitsOne, f64)
	}
	minInt64p1 := int64(math.MinInt64) + 1 // not a power of 2
	t.Logf("minInt64p1 %b %d", minInt64p1, minInt64p1)
	_, err = safecast.Convert[float64](minInt64p1)
	f64 = float64(minInt64p1)
	int2 := int64(f64)
	t.Logf("minInt64p1 -> %.0f %d", f64, int2)
	if err == nil {
		t.Errorf("expected error, got %d -> %.0f", minInt64p1, f64)
	}
}

func TestConvert(t *testing.T) {
	var inp uint32 = 42
	out, err := safecast.Convert[int8](inp)
	t.Logf("Out is %T: %v", out, out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out != 42 {
		t.Errorf("unexpected value: %v", out)
	}
	inp = 129
	_, err = safecast.Convert[int8](inp)
	t.Logf("Got err: %v", err)
	if err == nil {
		t.Errorf("expected error")
	}
	inp2 := int32(-1)
	_, err = safecast.Convert[uint8](inp2)
	t.Logf("Got err: %v", err)
	if err == nil {
		t.Errorf("expected error")
	}
	out, err = safecast.Convert[int8](inp2)
	t.Logf("Out is %T: %v", out, out)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out != -1 {
		t.Errorf("unexpected value: %v", out)
	}
	inp2 = -129
	_, err = safecast.Convert[uint8](inp2)
	t.Logf("Got err: %v", err)
	if err == nil {
		t.Errorf("expected error")
	}
	var a uint16 = 65535
	x, err := safecast.Convert[int16](a)
	if err == nil {
		t.Errorf("expected error, %d %d", a, x)
	}
	b := int8(-1)
	y, err := safecast.Convert[uint](b)
	if err == nil {
		t.Errorf("expected error, %d %d", b, y)
	}
	up := uintptr(42)
	b, err = safecast.Convert[int8](up)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if b != 42 {
		t.Errorf("unexpected value: %v", b)
	}
	b = -1
	_, err = safecast.Convert[uintptr](b)
	if err == nil {
		t.Errorf("expected err")
	}
	ub := safecast.MustTruncate[uint8](255.6)
	if ub != 255 {
		t.Errorf("unexpected value: %v", ub)
	}
	ub = safecast.MustConvert[uint8](int64(255)) // shouldn't panic
	if ub != 255 {
		t.Errorf("unexpected value: %v", ub)
	}
}

func TestPanicMustRound(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic")
		} else {
			expected := "safecast: out of range for 255.5 (float32) to uint8"
			if r != expected {
				t.Errorf("unexpected panic: %q wanted %q", r, expected)
			}
		}
	}()
	safecast.MustRound[uint8](float32(255.5))
}

func TestPanicMustTruncate(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic")
		} else {
			expected := "safecast: out of range for -1.5 (float32) to uint8"
			if r != expected {
				t.Errorf("unexpected panic: %q wanted %q", r, expected)
			}
		}
	}()
	safecast.MustTruncate[uint8](float32(-1.5))
}

func TestPanicMustConvert(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic")
		} else {
			expected := "safecast: out of range for 256 (int) to uint8"
			if r != expected {
				t.Errorf("unexpected panic: %q wanted %q", r, expected)
			}
		}
	}()
	safecast.MustConvert[uint8](256)
}

func Example() {
	var in int16 = 256
	// will error out
	out, err := safecast.Convert[uint8](in)
	fmt.Println(out, err)
	// will be fine
	out = safecast.MustRound[uint8](255.4)
	fmt.Println(out)
	// Also fine
	out = safecast.MustTruncate[uint8](255.6)
	fmt.Println(out)
	// Output: 0 out of range
	// 255
	// 255
}

func ExampleMustRound() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic:", r)
		}
	}()
	out := safecast.MustRound[int8](-128.6)
	fmt.Println("not reached", out) // not reached
	// Output: panic: safecast: out of range for -128.6 (float64) to int8
}

// We need this function, because 'f' format does not work as expected.
// Consider the following code:
//
//	src := 72057594037927936 // 2^56
//	dst := safecast.MustConvert[float32](src)
//	t.Logf("%d (0x%x) -> %s (%s)", src, src,
//		strconv.FormatFloat(float64(dst), 'f', -1, 32),
//		strconv.FormatFloat(float64(dst), 'x', -1, 32))
//
// It prints
//
//	72057594037927936 (0x100000000000000) -> 72057594000000000 (0x1p+56)
//
// So despite the actual number (float hex format is lossless)
// being stored in float is 72057594037927936,
// formatter is printed it as 72057594000000000.
func strconvAppendGoodFloat(w []byte, v float64) ([]byte, bool) {
	bits := math.Float64bits(v)
	// Get the raw IEEE 754 bit representation of the float64
	// Extract the sign bit
	sign := (bits >> 63) & 1
	// Extract the exponent bits
	// Mask the exponent bits (bits 52-62) and shift them to the right
	exponentBits := (bits >> 52) & 0x7FF // 0x7FF is 11 ones in binary (2^11 - 1)
	// Extract the mantissa bits
	// Mask the mantissa bits (bits 0-51)
	mantissaBits := bits & 0xFFFFFFFFFFFFF // 52 ones in binary (2^52 - 1)

	if exponentBits == 0 && mantissaBits == 0 { // +0 -0
		return append(w, '0'), true
	}
	if exponentBits == 0 { // denormalized
		return w, false
	}
	if exponentBits == 0x7FF { // +inf, -inf, all nans
		return w, false
	}

	u := mantissaBits | (1 << 52) // put implicit 1.

	e := int(exponentBits) - 0x3FF - 52

	for e > 0 {
		if u&(1<<63) != 0 {
			return w, false // would overflow
		}
		u <<= 1
		e--
	}
	for e < 0 {
		if u&1 != 0 {
			return w, false // would underflow
		}
		u >>= 1
		e++
	}
	if sign != 0 {
		w = append(w, '-')
	}
	return strconv.AppendUint(w, u, 10), true
}

func strconvAppend[Arg safecast.Number](w []byte, arg Arg) []byte {
	switch v := any(arg).(type) {
	case int:
		return strconv.AppendInt(w, int64(v), 10)
	case int8:
		return strconv.AppendInt(w, int64(v), 10)
	case int16:
		return strconv.AppendInt(w, int64(v), 10)
	case int32:
		return strconv.AppendInt(w, int64(v), 10)
	case int64:
		return strconv.AppendInt(w, v, 10)
	case uint:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint8:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(w, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(w, v, 10)
	case uintptr:
		return strconv.AppendUint(w, uint64(v), 10)
	case float32:
		wg, ok := strconvAppendGoodFloat(w, float64(v))
		if ok {
			return wg
		}
		return strconv.AppendFloat(w, float64(v), 'g', -1, 64)
	case float64:
		wg, ok := strconvAppendGoodFloat(w, v)
		if ok {
			return wg
		}
		return strconv.AppendFloat(w, v, 'g', -1, 64)
	default:
		panic("must be never")
	}
}

func testCast[Result safecast.Number, Arg safecast.Number](t *testing.T, arg Arg) {
	var scratch1 [64]byte
	var scratch2 [64]byte
	_, err := safecast.Convert[Result](arg)
	builtinCast := Result(arg)
	s1 := strconvAppend(scratch1[:0], arg)
	s2 := strconvAppend(scratch2[:0], builtinCast)
	good := string(s1) == string(s2)

	if (err == nil) != good {
		t.Errorf("%v %s -> %s %s\n", reflect.TypeOf(arg).String(), s1, reflect.TypeOf(builtinCast).String(), s2)
	}
}

func testCasts[Arg safecast.Number](t *testing.T, arg Arg) {
	testCast[int](t, arg)
	testCast[int8](t, arg)
	testCast[int16](t, arg)
	testCast[int32](t, arg)
	testCast[int64](t, arg)
	testCast[uint](t, arg)
	testCast[uint8](t, arg)
	testCast[uint16](t, arg)
	testCast[uint32](t, arg)
	testCast[uint64](t, arg)
	testCast[uintptr](t, arg)
	testCast[float32](t, arg)
	testCast[float64](t, arg)
}

func getPatterns(highBits uint64, shift byte) (uint64, uint64, uint64) {
	mask := uint64((1<<64)-1) >> (64 - shift)
	pattern1 := mask | (highBits << shift) // 00..00XFF..FF
	pattern2 := highBits << shift          // 00..00X00..00
	pattern3 := ^pattern1                  // FF..FFX00..00
	return pattern1, pattern2, pattern3
}

func testCastsFromInteger[Arg safecast.Number](t *testing.T) {
	for highBits := uint64(0); highBits < 16; highBits++ {
		for sh := 0; sh < 60; sh++ { // +4 bits in highBits
			p1, p2, p3 := getPatterns(highBits, byte(sh))
			testCasts(t, Arg(p1))
			testCasts(t, Arg(p2))
			testCasts(t, Arg(p3))
		}
	}
}

func testCastsFromFloat[Arg safecast.Number](t *testing.T) {
	for highBits := uint64(0); highBits < 16; highBits++ {
		for sh := 0; sh < 8; sh++ { // +4 bits in highBits
			p1, p2, p3 := getPatterns(highBits, byte(sh))
			testCastsFromFloatExp[Arg](t, p1)
			testCastsFromFloatExp[Arg](t, p2)
			testCastsFromFloatExp[Arg](t, p3)
		}
	}
}

func testCastsFromFloatExp[Arg safecast.Number](t *testing.T, expPattern uint64) {
	for sign := uint64(0); sign < 2; sign++ {
		for highBits := uint64(0); highBits < 16; highBits++ {
			for sh := 0; sh < 48; sh++ { // +4 bits in highBits
				p1, p2, p3 := getPatterns(highBits, byte(sh))
				testCasts(t, Arg(assembleFloat(sign, p1, expPattern)))
				testCasts(t, Arg(assembleFloat(sign, p2, expPattern)))
				testCasts(t, Arg(assembleFloat(sign, p3, expPattern)))
			}
		}
	}
}

func assembleFloat(sign, exp, mantissa uint64) float64 {
	bits := ((sign & 1) << 63) | ((exp & 0x7FF) << 52) | (mantissa & 0xFFFFFFFFFFFFF)
	ret := math.Float64frombits(bits)
	return ret
}

func TestPatterns(t *testing.T) {
	// some cases we triggered failure on and made fixes for
	testCast[byte](t, 9223372036854775807)
	testCast[float32](t, math.NaN())
	testCast[float64](t, float32(math.NaN()))
	testCast[float32](t, math.NaN())
	testCast[float64](t, float32(math.NaN()))
	testCast[byte](t, math.NaN())

	testCastsFromInteger[int](t)
	testCastsFromInteger[int8](t)
	testCastsFromInteger[int16](t)
	testCastsFromInteger[int32](t)
	testCastsFromInteger[int64](t)
	testCastsFromInteger[uint](t)
	testCastsFromInteger[uint8](t)
	testCastsFromInteger[uint16](t)
	testCastsFromInteger[uint32](t)
	testCastsFromInteger[uint64](t)
	testCastsFromInteger[uintptr](t)
	testCastsFromFloat[float32](t)
	testCastsFromFloat[float64](t)
}
