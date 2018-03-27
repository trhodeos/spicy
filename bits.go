package spicy

const hi32Bit = 0x80000000
const signExtend64BitMask = 0xFFFFFFFF00000000

func SignExtend(in uint64) uint64 {

	if in&hi32Bit == hi32Bit {
		return in | signExtend64BitMask
	} else {
		return in
	}
}
