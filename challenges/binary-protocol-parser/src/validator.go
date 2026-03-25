package protocol

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInvalidBodyLength = errors.New("invalid body length")
	ErrInvalidTypeTag    = errors.New("invalid type tag")
)

// ValidateHeader validates the message header fields
func ValidateHeader(magic uint16, version, bodyLen uint8) error {
	if magic != MagicNumber {
		return fmt.Errorf("invalid magic number: expected 0x%04X, got 0x%04X", MagicNumber, magic)
	}
	if version != Version {
		return fmt.Errorf("invalid version: expected 0x%02X, got 0x%02X", Version, version)
	}
	return nil
}

// ValidateTypeTag checks if a type tag is valid
func ValidateTypeTag(tag byte) error {
	switch tag {
	case TypeInt8, TypeInt16, TypeInt32, TypeInt64, TypeString, TypeBytes:
		return nil
	default:
		return fmt.Errorf("%w: 0x%02X", ErrInvalidTypeTag, tag)
	}
}

// ComputeChecksum computes a simple checksum for data integrity
func ComputeChecksum(data []byte) uint16 {
	var sum uint8
	for _, b := range data {
		sum += b
	}
	return uint16(sum)
}

// Stats tracks encoding and decoding statistics
type Stats struct {
	mu sync.RWMutex

	messagesEncoded int64
	messagesDecoded int64
	bytesEncoded    int64
	bytesDecoded    int64
	encodeErrors    int64
	decodeErrors    int64
	totalEncodeTime time.Duration
	totalDecodeTime time.Duration
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{}
}

// RecordEncode records a successful encoding operation
func (s *Stats) RecordEncode(byteCount int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messagesEncoded++
	s.bytesEncoded += int64(byteCount)
	s.totalEncodeTime += duration
}

// RecordDecode records a successful decoding operation
func (s *Stats) RecordDecode(byteCount int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messagesDecoded++
	s.bytesDecoded += int64(byteCount)
	s.totalDecodeTime += duration
}

// RecordEncodeError records an encoding error
func (s *Stats) RecordEncodeError() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.encodeErrors++
}

// RecordDecodeError records a decoding error
func (s *Stats) RecordDecodeError() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.decodeErrors++
}

// Snapshot returns a snapshot of current statistics
func (s *Stats) Snapshot() StatsSnapshot {
	s.mu.RLock()
	messagesEncoded := s.messagesEncoded
	messagesDecoded := s.messagesDecoded
	bytesEncoded := s.bytesEncoded
	bytesDecoded := s.bytesDecoded
	encodeErrors := s.encodeErrors
	decodeErrors := s.decodeErrors
	totalEncodeTime := s.totalEncodeTime
	totalDecodeTime := s.totalDecodeTime
	s.mu.RUnlock()

	var avgEncodeTime time.Duration
	if messagesEncoded > 0 {
		avgEncodeTime = totalEncodeTime / time.Duration(messagesEncoded)
	}

	var avgDecodeTime time.Duration
	if messagesDecoded > 0 {
		avgDecodeTime = totalDecodeTime / time.Duration(messagesDecoded)
	}

	return StatsSnapshot{
		MessagesEncoded: messagesEncoded,
		MessagesDecoded: messagesDecoded,
		BytesEncoded:    bytesEncoded,
		BytesDecoded:    bytesDecoded,
		EncodeErrors:    encodeErrors,
		DecodeErrors:    decodeErrors,
		AvgEncodeTime:   avgEncodeTime,
		AvgDecodeTime:   avgDecodeTime,
	}
}

// Reset resets all statistics
func (s *Stats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messagesEncoded = 0
	s.messagesDecoded = 0
	s.bytesEncoded = 0
	s.bytesDecoded = 0
	s.encodeErrors = 0
	s.decodeErrors = 0
	s.totalEncodeTime = 0
	s.totalDecodeTime = 0
}

// StatsSnapshot is a point-in-time snapshot of statistics
type StatsSnapshot struct {
	MessagesEncoded int64
	MessagesDecoded int64
	BytesEncoded    int64
	BytesDecoded    int64
	EncodeErrors    int64
	DecodeErrors    int64
	AvgEncodeTime   time.Duration
	AvgDecodeTime   time.Duration
}
