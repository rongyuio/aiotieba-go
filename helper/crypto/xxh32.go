package crypto

import (
	"encoding/binary"
	"math/bits"
)

// 流式 XXH32，是 csrc/xxHash/xxhash.h 中内置的 xxHash 0.8.x 的 XXH32 系列的忠实移植。仅提供 tbc_heliosHash 用到的操作：reset、update、copyState 与 digest。

const (
	prime32_1 uint32 = 0x9E3779B1
	prime32_2 uint32 = 0x85EBCA77
	prime32_3 uint32 = 0xC2B2AE3D
	prime32_4 uint32 = 0x27D4EB2F
	prime32_5 uint32 = 0x165667B1
)

type xxh32State struct {
	totalLen uint64
	largeLen bool
	v1       uint32
	v2       uint32
	v3       uint32
	v4       uint32
	mem      [16]byte
	memSize  int
}

func newXXH32(seed uint32) *xxh32State {
	return &xxh32State{
		v1: seed + prime32_1 + prime32_2,
		v2: seed + prime32_2,
		v3: seed,
		v4: seed - prime32_1,
	}
}

func xxh32Round(acc, input uint32) uint32 {
	acc += input * prime32_2
	acc = bits.RotateLeft32(acc, 13)
	acc *= prime32_1
	return acc
}

func (s *xxh32State) update(input []byte) {
	s.totalLen += uint64(len(input))
	if len(input) >= 16 || s.totalLen >= 16 {
		s.largeLen = true
	}

	p := input
	if s.memSize+len(p) < 16 {
		copy(s.mem[s.memSize:], p)
		s.memSize += len(p)
		return
	}

	if s.memSize > 0 {
		n := 16 - s.memSize
		copy(s.mem[s.memSize:], p[:n])
		s.round(s.mem[:])
		p = p[n:]
		s.memSize = 0
	}

	for len(p) >= 16 {
		s.round(p[:16])
		p = p[16:]
	}

	if len(p) > 0 {
		copy(s.mem[:], p)
		s.memSize = len(p)
	}
}

func (s *xxh32State) round(block []byte) {
	s.v1 = xxh32Round(s.v1, binary.LittleEndian.Uint32(block[0:4]))
	s.v2 = xxh32Round(s.v2, binary.LittleEndian.Uint32(block[4:8]))
	s.v3 = xxh32Round(s.v3, binary.LittleEndian.Uint32(block[8:12]))
	s.v4 = xxh32Round(s.v4, binary.LittleEndian.Uint32(block[12:16]))
}

func (s *xxh32State) copy() *xxh32State {
	copied := *s
	return &copied
}

func (s *xxh32State) digest() uint32 {
	var h32 uint32
	if s.largeLen {
		h32 = bits.RotateLeft32(s.v1, 1) + bits.RotateLeft32(s.v2, 7) +
			bits.RotateLeft32(s.v3, 12) + bits.RotateLeft32(s.v4, 18)
	} else {
		h32 = s.v3 + prime32_5
	}
	h32 += uint32(s.totalLen)

	p := s.mem[:s.memSize]
	for len(p) >= 4 {
		h32 += binary.LittleEndian.Uint32(p[:4]) * prime32_3
		h32 = bits.RotateLeft32(h32, 17) * prime32_4
		p = p[4:]
	}
	for _, b := range p {
		h32 += uint32(b) * prime32_5
		h32 = bits.RotateLeft32(h32, 11) * prime32_1
	}

	h32 ^= h32 >> 15
	h32 *= prime32_2
	h32 ^= h32 >> 13
	h32 *= prime32_3
	h32 ^= h32 >> 16
	return h32
}

// XXH32 以给定种子返回 input 的一次性 XXH32 哈希。
func XXH32(input []byte, seed uint32) uint32 {
	s := newXXH32(seed)
	s.update(input)
	return s.digest()
}
