package unary

type T uint16

const (
	A = 1
	B = ^A

	cA = T(1)
	cB = ^cA

	u8   = uint8(1)
	cu8  = ^u8
	u16  = uint16(1)
	cu16 = ^u16
	u32  = uint32(1)
	cu32 = ^u32
	u64  = uint64(1)
	cu64 = ^u64
	u    = uint(1)
	cu   = ^u
	up   = uintptr(1)
	cp   = ^up

	i8   = int8(1)
	ci8  = ^i8
	i16  = int16(1)
	ci16 = ^i16
	i32  = int32(1)
	ci32 = ^i32
	i64  = int64(1)
	ci64 = ^i64
	i    = int(1)
	ci   = ^i

	r0  = 'a'
	cr0 = ^r0
	r1  = rune('a')
	cr1 = ^r1
)

var (
	vA T
	vB = ^vA
)

const (
	z0 = ^i
	z1
)
