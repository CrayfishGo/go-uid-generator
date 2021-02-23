package buffer

import "sync"

type RingBuffer struct {
	tail   int64
	cursor int64

	slots []uint64
	flags []uint32

	//padding
	running uint64

	bufferSize uint64
	indexMask  int64

	paddingThreshold int64
	mu               sync.Mutex

	UidProvider func(uint64) []uint64

	lastSecond uint64
}
