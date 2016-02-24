package conv

type U8 uint8
type U16 uint16
type U32 uint32
type U64 uint64

type I8 int8
type I16 int16
type I32 int32
type I64 int64

const u8 = uint8(0xff)

const u16 = uint16(0xffff)
const u32 = uint32(0xffffffff)
const u64 = uint64(0xffffffffffffffff)
const u = uint(0xffffffffffffffff)
const uptr = uintptr(0xffffffffffffffff)

const w8 = U8(u8)
const w16 = U16(w8)
const w32 = U32(w16)
const w64 = U64(w32)

const i8 = int8(-128)

const i16 = int16(-32768)
const i32 = int32(-2147483648)
const i64 = int64(-9223372036854775808)

const i = int(-9223372036854775808)

const v8 = I8(i8)
const v16 = I16(v8)
const v32 = I32(v16)
const v64 = I64(v32)

const s = string('й')
const x = string(uptr)
const y = string(i)
const z0 = string(-18446744073709551616)
const z1 = string(18446744073709551616)

const f32 = float32(v32)
const f64 = float64(v64)
const g32 = float32(-v32)
const g64 = float64(-v64)
const h32 = float32(-128)
const h64 = float64(128)
const p32 = float32(0x1000001)
const p64 = float64(0x10000000000001)
const q32 = float32(uint32(0x80000001))
const q64 = float64(uint64(0x8000000000000001))

const c64 = complex64(v32)
const c128 = complex128(v64)
const d64 = complex64(-v32)
const d128 = complex128(-v64)
const e64 = complex64(-128)
const e128 = complex128(128)
