package id

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"
)

type Generator struct {
	random  io.Reader
	now     func() time.Time
	pid     int
	counter *uint64
}

func NewGenerator() *Generator {
	var counter uint64

	return &Generator{
		random:  rand.Reader,
		now:     func() time.Time { return time.Now().UTC() },
		pid:     os.Getpid(),
		counter: &counter,
	}
}

func (g *Generator) NewID() string {
	var value [16]byte
	if _, err := io.ReadFull(g.randomReader(), value[:]); err != nil {
		fillFallbackUUIDBytes(&value, g.nowFunc()().UTC(), g.processID(), g.nextCounter())
	}

	setUUIDVersionAndVariant(&value)
	return "shipment-" + formatUUID(value)
}

func (g *Generator) randomReader() io.Reader {
	if g != nil && g.random != nil {
		return g.random
	}

	return rand.Reader
}

func (g *Generator) nowFunc() func() time.Time {
	if g != nil && g.now != nil {
		return g.now
	}

	return func() time.Time { return time.Now().UTC() }
}

func (g *Generator) processID() int {
	if g != nil && g.pid != 0 {
		return g.pid
	}

	return os.Getpid()
}

func (g *Generator) nextCounter() uint64 {
	if g == nil || g.counter == nil {
		return 0
	}

	return atomic.AddUint64(g.counter, 1)
}

func fillFallbackUUIDBytes(target *[16]byte, now time.Time, pid int, counter uint64) {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%d:%d", now.UTC().UnixNano(), pid, counter)))
	copy(target[:], sum[:len(target)])
}

func setUUIDVersionAndVariant(target *[16]byte) {
	target[6] = (target[6] & 0x0f) | 0x40
	target[8] = (target[8] & 0x3f) | 0x80
}

func formatUUID(value [16]byte) string {
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(value[0:4]),
		binary.BigEndian.Uint16(value[4:6]),
		binary.BigEndian.Uint16(value[6:8]),
		binary.BigEndian.Uint16(value[8:10]),
		appendUint48(value[10:16]),
	)
}

func appendUint48(value []byte) uint64 {
	var result uint64
	for _, b := range value {
		result = (result << 8) | uint64(b)
	}

	return result
}
