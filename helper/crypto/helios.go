package crypto

import "hash/crc32"

// heliosHash is a byte-for-byte port of tbc_heliosHash.
//
// It mixes CRC32 and XXH32 over a 5-byte accumulator (sec) that is 40 bits
// wide. The C code manipulates 64-bit words; the shift widths are preserved
// exactly so bit 39 is never crossed.
func heliosHash(src []byte) [HeliosHashSize]byte {
	const stepSize = 5

	sec := uint64(1)<<40 - 1
	var buffer [4 * stepSize]byte
	for i := range stepSize {
		buffer[i] = 0xFF
	}

	// 1st hash with CRC32.
	crc := crc32.Update(0, crc32.IEEETable, src)
	crc = crc32.Update(crc, crc32.IEEETable, buffer[0:stepSize])
	sec = updateSection(sec, uint64(crc), 8, false)
	writeSection(buffer[stepSize:2*stepSize], sec)

	// 2nd hash with XXH32.
	xx := newXXH32(0)
	xx.update(src)
	xx.update(buffer[0 : stepSize*2])
	xxCopy := xx.copy()
	val := xx.digest()
	sec = updateSection(sec, uint64(val), 0, true)
	writeSection(buffer[stepSize*2:stepSize*3], sec)

	// 3rd hash with XXH32.
	xxCopy.update(buffer[stepSize*2 : stepSize*3])
	val = xxCopy.digest()
	sec = updateSection(sec, uint64(val), 1, true)
	writeSection(buffer[stepSize*3:stepSize*4], sec)

	// 4th hash with CRC32.
	crc = crc32.Update(crc, crc32.IEEETable, buffer[stepSize:stepSize*4])
	sec = updateSection(sec, uint64(crc), 7, true)

	var dst [HeliosHashSize]byte
	writeSection(dst[:], sec)
	return dst
}

// updateSection is a port of __tbc_update.
func updateSection(sec, hashVal uint64, start int, flag bool) uint64 {
	end := start + 32
	secTemp := sec
	mask := (uint64(1) << uint(end)) - 1
	val := (mask & sec) >> uint(start)

	if flag {
		val ^= hashVal
	} else {
		val &= hashVal
	}

	for i := range 32 {
		opIdx := start + i
		if val&(uint64(1)<<uint(i)) != 0 {
			secTemp |= uint64(1) << uint(opIdx)
		} else {
			secTemp &^= uint64(1) << uint(opIdx)
		}
	}
	return secTemp
}

// writeSection is a port of __tbc_writeBuffer: it writes the low 5 bytes of sec
// in little-endian order.
func writeSection(dst []byte, sec uint64) {
	for i := range 5 {
		dst[i] = byte(sec & 0xFF)
		sec >>= 8
	}
}
