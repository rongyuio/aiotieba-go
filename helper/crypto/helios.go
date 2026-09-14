package crypto

import "hash/crc32"

// heliosHash 是 tbc_heliosHash 的逐字节移植。
//
// 它在 40 位宽的 5 字节累加器（sec）上混合 CRC32 与 XXH32。C 代码操作 64 位字；位移宽度被精确保留，因此绝不会越过第 39 位。
func heliosHash(src []byte) [HeliosHashSize]byte {
	const stepSize = 5

	sec := uint64(1)<<40 - 1
	var buffer [4 * stepSize]byte
	for i := range stepSize {
		buffer[i] = 0xFF
	}

	// 第一次哈希使用 CRC32。
	crc := crc32.Update(0, crc32.IEEETable, src)
	crc = crc32.Update(crc, crc32.IEEETable, buffer[0:stepSize])
	sec = updateSection(sec, uint64(crc), 8, false)
	writeSection(buffer[stepSize:2*stepSize], sec)

	// 第二次哈希使用 XXH32。
	xx := newXXH32(0)
	xx.update(src)
	xx.update(buffer[0 : stepSize*2])
	xxCopy := xx.copy()
	val := xx.digest()
	sec = updateSection(sec, uint64(val), 0, true)
	writeSection(buffer[stepSize*2:stepSize*3], sec)

	// 第三次哈希使用 XXH32。
	xxCopy.update(buffer[stepSize*2 : stepSize*3])
	val = xxCopy.digest()
	sec = updateSection(sec, uint64(val), 1, true)
	writeSection(buffer[stepSize*3:stepSize*4], sec)

	// 第四次哈希使用 CRC32。
	crc = crc32.Update(crc, crc32.IEEETable, buffer[stepSize:stepSize*4])
	sec = updateSection(sec, uint64(crc), 7, true)

	var dst [HeliosHashSize]byte
	writeSection(dst[:], sec)
	return dst
}

// updateSection 是 __tbc_update 的移植。
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

// writeSection 是 __tbc_writeBuffer 的移植：以 little-endian 顺序写入 sec 的低 5 字节。
func writeSection(dst []byte, sec uint64) {
	for i := range 5 {
		dst[i] = byte(sec & 0xFF)
		sec >>= 8
	}
}
