package generator

import (
	"errors"
	"go-uid-generator/bits"
	"strconv"
	"time"
)

type DefaultUidGenerator struct {
	TimeBits      int
	WorkerBits    int
	SeqBits       int
	EpochStr      string
	epochSeconds  int64
	workerId      int64
	bitsAllocator *bits.BitsAllocator

	lastSecond int64
	sequence   int64
}

func NewDefaultUidGenerator() DefaultUidGenerator {
	defaultuidgenerator := DefaultUidGenerator{
		TimeBits:      28,
		WorkerBits:    22,
		SeqBits:       13,
		EpochStr:      "2016-05-20",
		sequence:      0,
		lastSecond:    -1,
		bitsAllocator: bits.NewDefaultBitAllocator(),
	}

	epochTime, _ := time.Parse("2006-01-02", defaultuidgenerator.EpochStr)
	defaultuidgenerator.epochSeconds = epochTime.Unix()
	return defaultuidgenerator
}

func (generator *DefaultUidGenerator) GetUID() (int64, error) {
	nextId, e := generator.nextId()
	if e == nil {
		return nextId, nil
	}
	return 0, e
}

func (generator *DefaultUidGenerator) ParseUID(uid int64) string {
	// todo
	return ""
}

func (generator *DefaultUidGenerator) nextId() (int64, error) {
	var currentSecond = int64(0)
	currentSecond, e := generator.getCurrentSecond()
	if e != nil {
		return currentSecond, e
	}
	if currentSecond < generator.lastSecond {
		refusedSeconds := generator.lastSecond - currentSecond
		return 0, errors.New("TClock moved backwards. Refusing for seconds: " + strconv.FormatInt(refusedSeconds, 20))
	}

	if currentSecond == generator.lastSecond {
		generator.sequence = (generator.sequence + 1) & generator.bitsAllocator.GetMaxSequence()
		if generator.sequence == 0 {
			currentSecond, e = generator.getNextSecond(generator.lastSecond)
		}
	} else {
		generator.sequence = 0
	}
	generator.lastSecond = currentSecond
	allocate := generator.bitsAllocator.Allocate(currentSecond-generator.epochSeconds, generator.workerId, generator.sequence)
	return allocate, nil
}

func (generator *DefaultUidGenerator) getCurrentSecond() (int64, error) {
	currentSecond := time.Now().Unix()
	if currentSecond-generator.epochSeconds > generator.bitsAllocator.GetMaxDeltaSeconds() {
		return 0, errors.New("Timestamp bits is exhausted. Refusing UID generate. Now: " + strconv.FormatInt(currentSecond, 20))
	}
	return currentSecond, nil
}

func (generator *DefaultUidGenerator) getNextSecond(lastTimestamp int64) (int64, error) {
	var timestamp int64 = 0
	timestamp, e := generator.getCurrentSecond()
	if e != nil {
		return 0, e
	}
	for timestamp <= lastTimestamp {
		timestamp1, _ := generator.getCurrentSecond()
		timestamp = timestamp1
	}
	return timestamp, nil
}
