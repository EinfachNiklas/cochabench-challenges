package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	MagicNumber uint16 = 0xBEEF
	Version     uint8  = 0x01

	TypeInt8   byte = 0x01
	TypeInt16  byte = 0x02
	TypeInt32  byte = 0x03
	TypeInt64  byte = 0x04
	TypeString byte = 0x05
	TypeBytes  byte = 0x06
)

var (
	ErrInvalidMagic   = errors.New("invalid magic number")
	ErrInvalidVersion = errors.New("invalid version")
	ErrInvalidType    = errors.New("invalid type tag")
	ErrBufferTooSmall = errors.New("buffer too small")
	ErrStringTooLong  = errors.New("string too long")
)

// EncodeInt8 encodes an 8-bit integer with protocol header
func EncodeInt8(w io.Writer, val int8) error {
	buf := NewBuffer(8)

	// Write header
	if err := writeHeader(buf, 1); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeInt8)

	// Write value
	buf.WriteByte(byte(val))

	_, err := w.Write(buf.Bytes())
	return err
}

// EncodeInt16 encodes a 16-bit integer with protocol header
func EncodeInt16(w io.Writer, val int16) error {
	buf := NewBuffer(16)

	// Write header
	if err := writeHeader(buf, 2); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeInt16)

	// Write value
	valBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(valBytes, uint16(val))
	buf.WriteBytes(valBytes)

	_, err := w.Write(buf.Bytes())
	return err
}

// EncodeInt32 encodes a 32-bit integer with protocol header
func EncodeInt32(w io.Writer, val int32) error {
	buf := NewBuffer(16)

	// Write header
	if err := writeHeader(buf, 5); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeInt32)

	// Write value
	valBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(valBytes, uint32(val))
	buf.WriteBytes(valBytes)

	_, err := w.Write(buf.Bytes())
	return err
}

// EncodeInt64 encodes a 64-bit integer with protocol header
func EncodeInt64(w io.Writer, val int64) error {
	buf := NewBuffer(16)

	// Write header
	if err := writeHeader(buf, 9); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeInt64)

	// Write value
	valBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valBytes, uint64(val))
	buf.WriteBytes(valBytes)

	_, err := w.Write(buf.Bytes())
	return err
}

// EncodeString encodes a UTF-8 string with protocol header
func EncodeString(w io.Writer, val string) error {
	buf := NewBuffer(len(val) + 16)

	strLen := len(val)

	if strLen > 255 {
		return ErrStringTooLong
	}

	bodyLen := 1 + 2 + strLen // type + length + data

	// Write header
	if err := writeHeader(buf, uint8(bodyLen)); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeString)

	// Write string length (2 bytes, big-endian)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(strLen))
	buf.WriteBytes(lenBytes)

	// Write string data
	buf.WriteBytes([]byte(val))

	_, err := w.Write(buf.Bytes())
	return err
}

// EncodeBytes encodes a byte array with protocol header
func EncodeBytes(w io.Writer, val []byte) error {
	buf := NewBuffer(len(val) + 16)

	dataLen := len(val)
	if dataLen > 65535 {
		return errors.New("byte array too long")
	}

	bodyLen := 1 + 2 + dataLen // type + length + data
	if bodyLen > 255 {
		bodyLen = 255 // Cap at 255 for single-byte length
	}

	// Write header
	if err := writeHeader(buf, uint8(bodyLen)); err != nil {
		return err
	}

	// Write type tag
	buf.WriteByte(TypeBytes)

	// Write data length (2 bytes, big-endian)
	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, uint16(dataLen))
	buf.WriteBytes(lenBytes)

	// Write data
	buf.WriteBytes(val)

	_, err := w.Write(buf.Bytes())
	return err
}

// DecodeMessage decodes a complete message from the reader
func DecodeMessage(r io.Reader) (typeTag byte, data interface{}, err error) {
	// Read all available data
	headerBytes := make([]byte, 4)
	_, err = r.Read(headerBytes)
	if err != nil {
		return 0, nil, err
	}

	// Parse header
	magic := binary.LittleEndian.Uint16(headerBytes[0:2])
	version := headerBytes[2]
	bodyLen := headerBytes[3]

	// Validate header
	if magic != MagicNumber {
		return 0, nil, fmt.Errorf("%w: got 0x%04X", ErrInvalidMagic, magic)
	}
	if version != Version {
		return 0, nil, ErrInvalidVersion
	}

	// Read body
	bodyBytes := make([]byte, bodyLen)
	_, err = r.Read(bodyBytes)
	if err != nil {
		return 0, nil, err
	}

	buf := NewBufferFromBytes(bodyBytes)

	// Read type tag
	typeTag, err = buf.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	// Decode based on type
	switch typeTag {
	case TypeInt8:
		val, err := buf.ReadByte()
		return typeTag, int8(val), err

	case TypeInt16:
		bytes, err := buf.ReadBytes(2)
		if err != nil {
			return 0, nil, err
		}
		val := binary.LittleEndian.Uint16(bytes)
		return typeTag, int16(val), nil

	case TypeInt32:
		bytes, err := buf.ReadBytes(4)
		if err != nil {
			return 0, nil, err
		}
		val := binary.BigEndian.Uint32(bytes)
		return typeTag, int32(val), nil

	case TypeInt64:
		bytes, err := buf.ReadBytes(8)
		if err != nil {
			return 0, nil, err
		}
		val := binary.BigEndian.Uint64(bytes)
		return typeTag, int64(val), nil

	case TypeString:
		lenBytes, err := buf.ReadBytes(2)
		if err != nil {
			return 0, nil, err
		}
		strLen := binary.BigEndian.Uint16(lenBytes)

		strBytes, err := buf.ReadBytes(int(strLen))
		if err != nil {
			return 0, nil, err
		}
		return typeTag, string(strBytes), nil

	case TypeBytes:
		lenBytes, err := buf.ReadBytes(2)
		if err != nil {
			return 0, nil, err
		}
		dataLen := binary.BigEndian.Uint16(lenBytes)

		dataBytes, err := buf.ReadBytes(int(dataLen))
		if err != nil {
			return 0, nil, err
		}
		return typeTag, dataBytes, nil

	default:
		return 0, nil, ErrInvalidType
	}
}

// writeHeader writes the protocol header to the buffer
func writeHeader(buf *Buffer, bodyLen uint8) error {
	// Magic number (big-endian)
	magicBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(magicBytes, MagicNumber)
	buf.WriteBytes(magicBytes)

	// Version
	buf.WriteByte(Version)

	// Body length
	buf.WriteByte(bodyLen)

	return nil
}
