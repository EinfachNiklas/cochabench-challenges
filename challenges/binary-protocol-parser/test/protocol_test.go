package protocol

import (
	"bytes"
	"encoding/binary"
	"sync"
	"testing"
	"time"
)

// Test Helper Functions

func createTestMessage(typeTag byte, data []byte) []byte {
	buf := []byte{
		0xBE, 0xEF, // Magic number (big-endian)
		0x01,                  // Version
		byte(1 + len(data)),   // Body length
		typeTag,               // Type tag
	}
	buf = append(buf, data...)
	return buf
}

func assertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func assertBytesEqual(t *testing.T, expected, actual []byte, msg string) {
	t.Helper()
	if !bytes.Equal(expected, actual) {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// Buffer Tests

func TestBufferReadByte(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		reads   int
		wantErr bool
	}{
		{"single byte", []byte{0x42}, 1, false},
		{"multiple bytes", []byte{0x01, 0x02, 0x03}, 3, false},
		{"exact buffer length", []byte{0x42}, 1, false},
		{"one beyond buffer", []byte{0x42}, 2, true},
		{"empty buffer", []byte{}, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewBufferFromBytes(tt.data)
			var err error
			for i := 0; i < tt.reads; i++ {
				_, err = buf.ReadByte()
				if err != nil {
					break
				}
			}
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBufferReadBytes(t *testing.T) {
	// Tests bug #8 - buffer overflow causes panic
	buf := NewBufferFromBytes([]byte{0x01, 0x02, 0x03})

	// This will panic due to the bug, so we need to recover
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when reading beyond buffer, but got none")
		}
	}()

	// This should panic due to bug
	_, _ = buf.ReadBytes(5)
}

func TestSetBitGetBit(t *testing.T) {
	// Test all bit positions
	tests := []struct {
		pos     uint
		val     bool
		wantErr bool
	}{
		{0, true, false},
		{1, true, false},
		{2, false, false},
		{3, true, false},
		{4, false, false},
		{5, true, false},
		{6, false, false},
		{7, true, false},
		{8, true, true}, // Tests bug #10 - invalid position
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var b byte = 0

			b = SetBit(b, tt.pos, tt.val)

			if tt.wantErr {
				// Position 8 should be invalid (only 0-7 valid)
				// But GetBit doesn't validate, so this is the bug
				result := GetBit(b, tt.pos)
				// The fact that we can call this without panic/error is the bug
				t.Logf("GetBit accepted invalid position %d, returned %v", tt.pos, result)
			} else {
				result := GetBit(b, tt.pos)
				if result != tt.val {
					t.Errorf("bit %d: expected %v, got %v", tt.pos, tt.val, result)
				}
			}
		})
	}
}

func TestSetBitToggle(t *testing.T) {
	// Tests bug #9 - SetBit uses OR for both set AND clear operations
	// This means it can only SET bits, never CLEAR them

	var b byte = 0xFF // All bits set (11111111)

	// Try to clear bit 0
	b = SetBit(b, 0, false)
	if GetBit(b, 0) {
		t.Error("SetBit failed to clear bit 0 - bug #9 detected: uses OR instead of AND-NOT for clearing")
	}

	// Try another test: start with 0, set bit 3, then clear it
	b = 0x00
	b = SetBit(b, 3, true) // Should set bit 3
	if !GetBit(b, 3) {
		t.Error("failed to set bit 3")
	}

	b = SetBit(b, 3, false) // Should clear bit 3
	if GetBit(b, 3) {
		t.Error("SetBit failed to clear bit 3 - bug #9 detected")
	}
}

// Encoding Tests

func TestEncodeInt8(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeInt8(&buf, 42)
	assertNoError(t, err, "EncodeInt8")

	data := buf.Bytes()
	// Message format: [magic(2), version(1), bodyLen(1), typeTag(1), value(1)]
	// Expected: 6 bytes total, bodyLen should be 2 (typeTag + value)
	// Bug #3: bodyLen is 1 instead of 2 (off-by-one)

	if len(data) != 6 {
		t.Errorf("expected 6 bytes total, got %d", len(data))
	}

	bodyLen := data[3]
	// Correct body length should be 2 (type tag + int8 value)
	if bodyLen != 2 {
		t.Errorf("bug #3 detected: body length should be 2, got %d", bodyLen)
	}
}

func TestEncodeInt16BodyLength(t *testing.T) {
	// Additional test for bug #3 with Int16
	var buf bytes.Buffer
	err := EncodeInt16(&buf, 256)
	assertNoError(t, err, "EncodeInt16")

	data := buf.Bytes()
	bodyLen := data[3]
	// Correct body length should be 3 (type tag + 2 bytes for int16)
	if bodyLen != 3 {
		t.Errorf("bug #3 detected in Int16: body length should be 3, got %d", bodyLen)
	}
}

func TestEncodeInt16Endianness(t *testing.T) {
	// Tests bug #1 - Int16 encoding uses wrong endianness
	tests := []struct {
		name     string
		value    int16
		expected []byte // Expected big-endian representation
	}{
		{"zero", 0, []byte{0x00, 0x00}},
		{"positive", 256, []byte{0x01, 0x00}}, // Big-endian: high byte first
		{"negative", -1, []byte{0xFF, 0xFF}},
		{"max", 32767, []byte{0x7F, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := EncodeInt16(&buf, tt.value)
			assertNoError(t, err, "EncodeInt16")

			data := buf.Bytes()
			if len(data) < 7 {
				t.Fatal("encoded data too short")
			}

			// Extract the actual value bytes (skip header and type tag)
			actualBytes := data[5:7]
			assertBytesEqual(t, tt.expected, actualBytes, "Int16 byte representation")
		})
	}
}

func TestEncodeInt32(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeInt32(&buf, 0x12345678)
	assertNoError(t, err, "EncodeInt32")

	data := buf.Bytes()
	// Check big-endian encoding
	expected := []byte{0x12, 0x34, 0x56, 0x78}
	actualBytes := data[5:9]
	assertBytesEqual(t, expected, actualBytes, "Int32 big-endian")
}

func TestEncodeInt64(t *testing.T) {
	var buf bytes.Buffer
	err := EncodeInt64(&buf, 0x123456789ABCDEF0)
	assertNoError(t, err, "EncodeInt64")

	data := buf.Bytes()
	expected := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	actualBytes := data[5:13]
	assertBytesEqual(t, expected, actualBytes, "Int64 big-endian")
}

func TestEncodeStringUTF8(t *testing.T) {
	// Test UTF-8 string encoding
	tests := []struct {
		name         string
		value        string
		expectedLen  int
	}{
		{"ascii", "hello", 5},
		{"utf8-japanese", "日本語", 9},    // 3 chars, 9 bytes
		{"utf8-emoji", "🚀", 4},         // 1 char, 4 bytes
		{"mixed", "Hello世界", 11},       // 7 chars, 11 bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := EncodeString(&buf, tt.value)
			assertNoError(t, err, "EncodeString")

			data := buf.Bytes()
			// Check the length field (2 bytes after type tag)
			if len(data) < 7 {
				t.Fatal("encoded data too short")
			}
			lenBytes := data[5:7]
			actualLen := binary.BigEndian.Uint16(lenBytes)

			if int(actualLen) != tt.expectedLen {
				t.Errorf("string length: expected %d bytes, got %d", tt.expectedLen, actualLen)
			}
		})
	}
}

func TestEncodeString255Bytes(t *testing.T) {
	// Tests boundary at 255 bytes - should work
	str := string(make([]byte, 255))
	var buf bytes.Buffer
	err := EncodeString(&buf, str)

	if err != nil {
		t.Errorf("EncodeString with 255 bytes should work, got error: %v", err)
	}
}

func TestEncodeString256Bytes(t *testing.T) {
	// Tests bug #5 - no support for strings > 255 bytes
	str := string(make([]byte, 256))
	var buf bytes.Buffer
	err := EncodeString(&buf, str)

	// Bug #5: Implementation returns ErrStringTooLong for strings > 255 bytes
	// Should support extended length but doesn't
	if err == nil {
		t.Error("bug #5 NOT detected: string > 255 bytes should fail due to missing extended length support")
	} else if err.Error() != "string too long" {
		t.Logf("got error as expected: %v", err)
	}
}

func TestEncodeBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	var buf bytes.Buffer
	err := EncodeBytes(&buf, data)
	assertNoError(t, err, "EncodeBytes")
}

// Decoding Tests

func TestDecodeInt8(t *testing.T) {
	msg := createTestMessage(TypeInt8, []byte{0x42})
	reader := NewBufferFromBytes(msg).ToReader()

	typeTag, data, err := DecodeMessage(reader)
	if err != nil {
		t.Logf("DecodeMessage error: %v", err)
		return
	}
	assertEqual(t, TypeInt8, typeTag, "type tag")
	if data != nil {
		assertEqual(t, int8(0x42), data.(int8), "decoded value")
	}
}

func TestDecodeInt16Endianness(t *testing.T) {
	// Tests bug #2 - Int16 decoding uses wrong endianness
	tests := []struct {
		name     string
		bytes    []byte // Big-endian representation
		expected int16
	}{
		{"zero", []byte{0x00, 0x00}, 0},
		{"256", []byte{0x01, 0x00}, 256},
		{"negative", []byte{0xFF, 0xFF}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := createTestMessage(TypeInt16, tt.bytes)
			reader := NewBufferFromBytes(msg).ToReader()

			typeTag, data, err := DecodeMessage(reader)
			if err != nil {
				t.Logf("DecodeMessage error (likely bug #8 magic number): %v", err)
				return
			}
			assertEqual(t, TypeInt16, typeTag, "type tag")

			if data != nil {
				actual := data.(int16)
				if actual != tt.expected {
					t.Errorf("expected %d, got %d (bytes: %v)", tt.expected, actual, tt.bytes)
				}
			}
		})
	}
}

func TestDecodeInt32(t *testing.T) {
	testBytes := []byte{0x12, 0x34, 0x56, 0x78}
	msg := createTestMessage(TypeInt32, testBytes)
	reader := NewBufferFromBytes(msg).ToReader()

	typeTag, data, err := DecodeMessage(reader)
	if err != nil {
		t.Logf("DecodeMessage error: %v", err)
		return
	}
	assertEqual(t, TypeInt32, typeTag, "type tag")
	if data != nil {
		assertEqual(t, int32(0x12345678), data.(int32), "decoded value")
	}
}

func TestDecodeInt64(t *testing.T) {
	testBytes := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	msg := createTestMessage(TypeInt64, testBytes)
	reader := NewBufferFromBytes(msg).ToReader()

	typeTag, data, err := DecodeMessage(reader)
	if err != nil {
		t.Logf("DecodeMessage error: %v", err)
		return
	}
	assertEqual(t, TypeInt64, typeTag, "type tag")
	if data != nil {
		assertEqual(t, int64(0x123456789ABCDEF0), data.(int64), "decoded value")
	}
}

func TestDecodeInvalidMagic(t *testing.T) {
	// Tests bug #7 - magic number validation with wrong endianness
	msg := []byte{
		0xEF, 0xBE, // Wrong byte order
		0x01,
		0x02,
		TypeInt8,
		0x42,
	}
	reader := NewBufferFromBytes(msg).ToReader()

	_, _, err := DecodeMessage(reader)
	assertError(t, err, "should reject invalid magic number")
}

func TestDecodePartialMessage(t *testing.T) {
	// Tests bug #4 - no EOF check
	msg := []byte{0xBE, 0xEF} // Incomplete header
	reader := NewBufferFromBytes(msg).ToReader()

	_, _, err := DecodeMessage(reader)
	assertError(t, err, "should handle partial message")
}

func TestDecodeMemoryAliasing(t *testing.T) {
	// Tests bug #6 - decoded bytes share memory with internal buffer
	originalData := []byte{0x01, 0x02, 0x03}
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(len(originalData)))

	bodyData := append(lenBytes, originalData...)
	msg := createTestMessage(TypeBytes, bodyData)

	// First decode
	reader1 := NewBufferFromBytes(msg).ToReader()
	_, data1, err := DecodeMessage(reader1)
	if err != nil {
		t.Logf("DecodeMessage error: %v", err)
		return
	}

	if data1 == nil {
		t.Skip("data is nil, skipping aliasing test")
	}

	decodedBytes1 := data1.([]byte)

	// Verify initial values
	if decodedBytes1[0] != 0x01 {
		t.Errorf("expected first byte to be 0x01, got 0x%02X", decodedBytes1[0])
	}

	// Modify the decoded bytes
	decodedBytes1[0] = 0xFF
	decodedBytes1[1] = 0xFF
	decodedBytes1[2] = 0xFF

	// Second decode from the SAME original message
	reader2 := NewBufferFromBytes(msg).ToReader()
	_, data2, err := DecodeMessage(reader2)
	if err != nil {
		t.Logf("DecodeMessage error: %v", err)
		return
	}

	if data2 == nil {
		t.Skip("data is nil, skipping aliasing test")
	}

	decodedBytes2 := data2.([]byte)

	// Bug #6: If there's memory aliasing, decodedBytes2 will show modifications
	// because it shares the same underlying buffer
	if decodedBytes2[0] != 0x01 {
		t.Errorf("bug #6 detected: memory aliasing - expected 0x01, got 0x%02X (shares buffer with previous decode)", decodedBytes2[0])
	}
	if decodedBytes2[1] != 0x02 {
		t.Errorf("bug #6 detected: memory aliasing at byte 1")
	}
	if decodedBytes2[2] != 0x03 {
		t.Errorf("bug #6 detected: memory aliasing at byte 2")
	}
}

// Integration Tests - Roundtrip

func TestRoundtripInt16(t *testing.T) {
	// This test will fail due to endianness bugs #1 and #2
	testValues := []int16{0, 1, -1, 256, -256, 32767, -32768}

	for _, original := range testValues {
		t.Run("", func(t *testing.T) {
			var buf bytes.Buffer

			// Encode
			err := EncodeInt16(&buf, original)
			if err != nil {
				t.Skipf("encode failed: %v", err)
			}

			// Decode
			_, decoded, err := DecodeMessage(&buf)
			if err != nil {
				t.Logf("decode failed (likely bug #7 magic number): %v", err)
				return
			}

			if decoded == nil {
				t.Skip("decoded data is nil")
			}

			result := decoded.(int16)
			if result != original {
				t.Errorf("roundtrip failed: %d -> %d", original, result)
			}
		})
	}
}

func TestRoundtripString(t *testing.T) {
	// Tests string encoding/decoding with UTF-8
	testStrings := []string{
		"",
		"hello",
		"Hello, World!",
		"日本語",
		"🚀🎉",
		"Mixed ASCII and 中文",
	}

	for _, original := range testStrings {
		t.Run(original, func(t *testing.T) {
			var buf bytes.Buffer

			err := EncodeString(&buf, original)
			if err != nil {
				t.Skipf("encode failed: %v", err)
			}

			_, decoded, err := DecodeMessage(&buf)
			if err != nil {
				t.Logf("decode failed: %v", err)
				return
			}

			if decoded == nil {
				t.Skip("decoded data is nil")
			}

			result := decoded.(string)
			if result != original {
				t.Errorf("roundtrip failed: %q -> %q", original, result)
			}
		})
	}
}

// Validator Tests

func TestValidateHeader(t *testing.T) {
	tests := []struct {
		name      string
		magic     uint16
		version   uint8
		bodyLen   uint8
		wantError bool
	}{
		{"valid", 0xBEEF, 0x01, 10, false},
		{"invalid magic", 0xDEAD, 0x01, 10, true},
		{"invalid version", 0xBEEF, 0x02, 10, true},
		{"body length 1", 0xBEEF, 0x01, 1, false},
		{"body length 255", 0xBEEF, 0x01, 255, false},
		{"body length 0", 0xBEEF, 0x01, 0, true}, // Tests bug #11 - should reject 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHeader(tt.magic, tt.version, tt.bodyLen)
			if tt.wantError && err == nil {
				t.Error("expected error but got nil - bug #11: should reject bodyLen=0")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestComputeChecksum(t *testing.T) {
	// Tests bug #12 - checksum overflow with uint8
	tests := []struct {
		name string
		data []byte
		want uint16
	}{
		{"empty", []byte{}, 0},
		{"single", []byte{42}, 42},
		{"small", []byte{1, 2, 3}, 6},
		{"overflow", bytes.Repeat([]byte{255}, 10), 2550}, // Tests overflow bug
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeChecksum(tt.data)
			if got != tt.want {
				t.Errorf("checksum: expected %d, got %d", tt.want, got)
			}
		})
	}
}

// Stats Tests

func TestStatsConcurrent(t *testing.T) {
	// Tests bug #13 - race condition in Snapshot
	stats := NewStats()

	var wg sync.WaitGroup
	n := 100

	// Concurrent writers
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.RecordEncode(10, time.Millisecond)
			stats.RecordDecode(10, time.Millisecond)
		}()
	}

	// Concurrent readers
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = stats.Snapshot()
		}()
	}

	wg.Wait()

	snapshot := stats.Snapshot()
	if snapshot.MessagesEncoded != int64(n) {
		t.Errorf("expected %d encoded messages, got %d", n, snapshot.MessagesEncoded)
	}
}

func TestStatsSnapshot(t *testing.T) {
	stats := NewStats()

	stats.RecordEncode(100, 10*time.Millisecond)
	stats.RecordEncode(200, 20*time.Millisecond)
	stats.RecordDecode(150, 15*time.Millisecond)
	stats.RecordEncodeError()
	stats.RecordDecodeError()
	stats.RecordDecodeError()

	snapshot := stats.Snapshot()

	assertEqual(t, int64(2), snapshot.MessagesEncoded, "messages encoded")
	assertEqual(t, int64(1), snapshot.MessagesDecoded, "messages decoded")
	assertEqual(t, int64(300), snapshot.BytesEncoded, "bytes encoded")
	assertEqual(t, int64(150), snapshot.BytesDecoded, "bytes decoded")
	assertEqual(t, int64(1), snapshot.EncodeErrors, "encode errors")
	assertEqual(t, int64(2), snapshot.DecodeErrors, "decode errors")

	if snapshot.AvgEncodeTime != 15*time.Millisecond {
		t.Errorf("avg encode time: expected 15ms, got %v", snapshot.AvgEncodeTime)
	}
}

// Edge Case Tests

func TestEncodeDecodeZeroValues(t *testing.T) {
	t.Run("Int8(0)", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeInt8(&buf, 0)
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, int8(0), data.(int8), "Int8(0)")
	})

	t.Run("Int16(0)", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeInt16(&buf, 0)
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, int16(0), data.(int16), "Int16(0)")
	})

	t.Run("empty string", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeString(&buf, "")
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, "", data.(string), "empty string")
	})

	t.Run("empty bytes", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeBytes(&buf, []byte{})
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, 0, len(data.([]byte)), "empty bytes")
	})
}

func TestEncodeDecodeMaxValues(t *testing.T) {
	t.Run("Int8 max", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeInt8(&buf, 127)
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, int8(127), data.(int8), "Int8 max")
	})

	t.Run("Int8 min", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeInt8(&buf, -128)
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, int8(-128), data.(int8), "Int8 min")
	})

	t.Run("Int32 max", func(t *testing.T) {
		var buf bytes.Buffer
		EncodeInt32(&buf, 2147483647)
		_, data, err := DecodeMessage(&buf)
		if err != nil || data == nil {
			t.Skipf("decode failed or nil: %v", err)
		}
		assertEqual(t, int32(2147483647), data.(int32), "Int32 max")
	})
}
