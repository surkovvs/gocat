package catkafkaprod

import (
	"hash"
	"hash/crc32"

	"github.com/cespare/xxhash/v2"
)

type partitioner interface {
	PartHash([]byte) (uint32, error)
}

type prodOptions func(*producer)

type customPartitioner struct {
	customFunc func([]byte) (uint32, error)
}

func (cp customPartitioner) PartHash(key []byte) (uint32, error) {
	return cp.customFunc(key)
}

func PartWithFunc(partFunc func([]byte) (uint32, error)) prodOptions {
	return func(p *producer) {
		p.partitioner = customPartitioner{partFunc}
	}
}

func PartWithCrc32Hash(poly *uint32) prodOptions {
	var hasher hash.Hash32
	if poly != nil {
		hasher = crc32.New(crc32.MakeTable(*poly))
	} else {
		hasher = crc32.NewIEEE()
	}

	partFunc := func(p []byte) (uint32, error) {
		hasher.Reset()
		if _, err := hasher.Write(p); err != nil {
			return 0, err
		}
		return hasher.Sum32(), nil
	}

	return func(p *producer) {
		p.partitioner = customPartitioner{partFunc}
	}
}

func PartWithxxhash(seed *uint64) prodOptions {
	var hasher *xxhash.Digest
	if seed != nil {
		hasher = xxhash.NewWithSeed(*seed)
	} else {
		hasher = xxhash.New()
	}

	partFunc := func(p []byte) (uint32, error) {
		hasher.Reset()
		if _, err := hasher.Write(p); err != nil {
			return 0, err
		}
		return uint32(hasher.Sum64()), nil
	}

	return func(p *producer) {
		p.partitioner = customPartitioner{partFunc}
	}
}

func PartRR() prodOptions {
	var i uint32 = 0
	i--
	partFunc := func(p []byte) (uint32, error) {
		i++
		return i, nil
	}

	return func(p *producer) {
		p.partitioner = customPartitioner{partFunc}
	}
}
