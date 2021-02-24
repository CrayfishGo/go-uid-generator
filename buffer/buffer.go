package buffer

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	START_POINT             int64  = -1
	CAN_PUT_FLAG            uint32 = 0
	CAN_TAKE_FLAG           uint32 = 1
	DEFAULT_PADDING_PERCENT int64  = 50
	NOT_RUNNING             uint64 = 0
	RUNNING                 uint64 = 1
)

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

func NewRingBuffer(bufferSize uint64, paddingFactor uint64) *RingBuffer {
	buffer := &RingBuffer{
		bufferSize:       bufferSize,
		indexMask:        (int64)(bufferSize - 1),
		slots:            make([]uint64, bufferSize),
		flags:            make([]uint32, bufferSize),
		paddingThreshold: (int64)(bufferSize * paddingFactor / 100),
		tail:             START_POINT,
		cursor:           START_POINT,
		lastSecond:       uint64(time.Now().Unix()),
	}
	var i uint64
	for i = 0; i < bufferSize; i++ {
		buffer.flags[i] = CAN_PUT_FLAG
	}
	return buffer
}

func (buffer *RingBuffer) Put(uid uint64) bool {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()

	currentTail := atomic.LoadInt64(&buffer.tail)
	currentCursor := atomic.LoadInt64(&buffer.cursor)
	var distance uint64
	if currentCursor == START_POINT {
		distance = uint64(currentTail - 0)
	} else {
		distance = uint64(currentTail - currentCursor)
	}

	if distance == buffer.bufferSize-1 {
		log.Default().Printf("Rejected putting buffer for uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	nextTailIndex := buffer.calSlotIndex(currentTail + 1)
	if atomic.LoadUint32(&buffer.flags[nextTailIndex]) != CAN_PUT_FLAG {
		log.Default().Printf("Curosr not in can put status,rejected uid:%v,tail:%v,cursor:%v\n", uid, currentTail, currentCursor)
		return false
	}

	atomic.StoreUint64(&buffer.slots[nextTailIndex], uid)
	atomic.StoreUint32(&buffer.flags[nextTailIndex], CAN_TAKE_FLAG)
	atomic.AddInt64(&buffer.tail, 1)
	return true
}

func (buffer *RingBuffer) Take() (uint64, error) {
	currentCursor := atomic.LoadInt64(&buffer.cursor)
	nextCursor := Uint64UpdateAndGet(&buffer.cursor, func(old int64) int64 {
		if old == atomic.LoadInt64(&buffer.tail) {
			return old
		} else {
			return old + 1
		}
	})

	if nextCursor < currentCursor {
		panic("Curosr can't move back")
	}
	currentTail := atomic.LoadInt64(&buffer.tail)
	if currentTail-nextCursor < buffer.paddingThreshold {
		log.Default().Printf("Reach the padding threshold:%v, tail:%v, cursor:%v, rest:%v", buffer.paddingThreshold, currentTail, nextCursor, currentTail-nextCursor)
		go buffer.asyncPadding()
	}
	if currentTail == currentCursor {
		//拒绝
		return 0, errors.New("Rejected take uid")
	}

	nextCursorIndex := buffer.calSlotIndex(nextCursor)
	if atomic.LoadUint32(&buffer.flags[nextCursorIndex]) != CAN_TAKE_FLAG {
		panic("Curosr not in can take status")
	}

	uid := atomic.LoadUint64(&buffer.slots[nextCursorIndex])
	atomic.StoreUint32(&buffer.flags[nextCursorIndex], CAN_PUT_FLAG)
	return uid, nil
}

func (buffer *RingBuffer) calSlotIndex(sequence int64) int64 {
	return sequence & buffer.indexMask
}

func (buffer *RingBuffer) asyncPadding() {
	log.Default().Printf("Ready to padding buffer lastSecond:%v. %v", atomic.LoadUint64(&buffer.lastSecond), buffer)
	// is still running
	if !atomic.CompareAndSwapUint64(&buffer.running, NOT_RUNNING, RUNNING) {
		log.Default().Printf("Padding buffer is still running. %v", buffer)
		return
	}
	// fill the rest slots until to catch the cursor
	var isFullRingBuffer bool = false
	for !isFullRingBuffer {
		uidList := buffer.UidProvider(atomic.AddUint64(&buffer.lastSecond, 1))
		for _, uid := range uidList {
			isFullRingBuffer = !buffer.Put(uid)
			if isFullRingBuffer {
				break
			}
		}
	}
	// not running now
	atomic.CompareAndSwapUint64(&buffer.running, RUNNING, NOT_RUNNING)
	log.Default().Printf("End to padding buffer lastSecond:%v. %v", atomic.LoadUint64(&buffer.lastSecond), buffer)
}

func Uint64UpdateAndGet(v *int64, f func(int64) int64) int64 {
	var old, next int64
	for {
		old = atomic.LoadInt64(v)
		next = f(old)
		if atomic.CompareAndSwapInt64(v, old, next) {
			break
		}
	}
	return next
}
