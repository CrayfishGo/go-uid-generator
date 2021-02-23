package bits

import "errors"

type BitsAllocator struct {
	maxDeltaSeconds int64
	maxSequence     int64
	maxWorkerId     int64
	timestampShift  int64
	workerIdShift   int64
	workerIdBits    int64
	sequenceBits    int64
	timestampBits   int64
	signBits        int64
}

const TOTAL_BITS = 1 << 6

func NewBitsAllocator(timestampBits int64, workerIdBits int64, sequenceBits int64) (*BitsAllocator, error) {
	var bitsAllocator = &BitsAllocator{
		signBits: 1,
	}
	var allocateTotalBits = 1 + timestampBits + workerIdBits + sequenceBits
	if allocateTotalBits != TOTAL_BITS {
		return bitsAllocator, errors.New("allocate not enough 64 bits")
	}

	bitsAllocator.timestampShift = workerIdBits + sequenceBits
	bitsAllocator.workerIdShift = sequenceBits
	bitsAllocator.sequenceBits = sequenceBits
	bitsAllocator.workerIdBits = workerIdBits
	bitsAllocator.maxDeltaSeconds = -1 ^ (-1 << timestampBits)
	bitsAllocator.maxWorkerId = -1 ^ (-1 << workerIdBits)
	bitsAllocator.maxSequence = -1 ^ (-1 << sequenceBits)
	return bitsAllocator, nil
}

func NewDefaultBitAllocator() *BitsAllocator {
	return &BitsAllocator{
		maxDeltaSeconds: -1 ^ (-1 << 28),
		maxSequence:     -1 ^ (-1 << 13),
		maxWorkerId:     -1 ^ (-1 << 22),
		timestampShift:  22 + 13,
		workerIdShift:   13,
		workerIdBits:    22,
		sequenceBits:    13,
		timestampBits:   28,
	}
}

func (idBits *BitsAllocator) Allocate(deltaSeconds int64, workerId int64, sequence int64) int64 {
	return (deltaSeconds << idBits.timestampShift) | (workerId << idBits.workerIdShift) | sequence
}

// get max Delta Seconds
func (idBits *BitsAllocator) GetMaxDeltaSeconds() int64 {
	return idBits.maxDeltaSeconds
}

// get max worknode id
func (idBits *BitsAllocator) GetMaxWorkerId() int64 {
	return idBits.maxWorkerId
}

// get max sequence
func (idBits *BitsAllocator) GetMaxSequence() int64 {
	return idBits.maxSequence
}

func (idBits *BitsAllocator) GetSignBits() int64 {
	return idBits.signBits
}
