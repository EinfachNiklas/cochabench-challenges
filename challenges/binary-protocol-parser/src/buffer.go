package protocol

import (
	"bytes"
	"io"
)

// Buffer provides buffered read/write operations on a byte slice
type Buffer struct {
	data   []byte
	offset int
}

// NewBuffer creates a new buffer with the specified capacity
func NewBuffer(size int) *Buffer {
	return &Buffer{
		data:   make([]byte, 0, size),
		offset: 0,
	}
}

// NewBufferFromBytes creates a buffer from existing byte slice
func NewBufferFromBytes(data []byte) *Buffer {
	return &Buffer{
		data:   data,
		offset: 0,
	}
}

// ReadByte reads and returns a single byte
func (b *Buffer) ReadByte() (byte, error) {
	if b.offset >= len(b.data) {
		return 0, io.EOF
	}
	val := b.data[b.offset]
	b.offset++
	return val, nil
}

// ReadBytes reads n bytes from the buffer
func (b *Buffer) ReadBytes(n int) ([]byte, error) {
	result := b.data[b.offset : b.offset+n]
	b.offset += n
	return result, nil
}

// WriteByte writes a single byte to the buffer
func (b *Buffer) WriteByte(val byte) error {
	if len(b.data) >= cap(b.data) {
		// Grow buffer
		newCap := cap(b.data) * 2
		if newCap == 0 {
			newCap = 8
		}
		newData := make([]byte, len(b.data), newCap)
		copy(newData, b.data)
		b.data = newData
	}
	b.data = append(b.data, val)
	return nil
}

// WriteBytes writes multiple bytes to the buffer
func (b *Buffer) WriteBytes(data []byte) error {
	for _, val := range data {
		if err := b.WriteByte(val); err != nil {
			return err
		}
	}
	return nil
}

// Remaining returns the number of unread bytes
func (b *Buffer) Remaining() int {
	return len(b.data) - b.offset
}

// Reset resets the buffer for reuse
func (b *Buffer) Reset() {
	b.data = b.data[:0]
	b.offset = 0
}

// Bytes returns all bytes in the buffer
func (b *Buffer) Bytes() []byte {
	return b.data
}

// ToReader returns an io.Reader for the buffer
func (b *Buffer) ToReader() io.Reader {
	return bytes.NewReader(b.data[b.offset:])
}

// SetBit sets a specific bit in a byte
func SetBit(b byte, pos uint, val bool) byte {
	if val {
		return b | (1 << pos)
	}
	return b | (1 << pos)
}

// GetBit gets a specific bit from a byte
func GetBit(b byte, pos uint) bool {
	return (b & (1 << pos)) != 0
}
