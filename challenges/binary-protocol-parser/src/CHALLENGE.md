# Binary Protocol Parser - Bug Hunt Challenge

## Task

**Debug and fix** a faulty binary protocol parser implementation in Go. The code is already written but contains multiple bugs that prevent it from working correctly. Your task is to find and fix all bugs to make the implementation match the protocol specification.

The protocol is designed to encode and decode typed data in a space-efficient binary format, similar to Protocol Buffers or MessagePack.

The protocol supports the following data types:
- **Integers**: Int8, Int16, Int32, Int64
- **Strings**: UTF-8 encoded, length-prefixed
- **Byte Arrays**: Length-prefixed binary data

All multi-byte integers use **big-endian** (network byte order) encoding.

### Protocol Format

Each message consists of a **header** and a **body**:

```
Message Structure:
[Header: 4 bytes] [Body: variable length]

Header Format:
- Magic Number: 2 bytes (0xBEEF) - Protocol identifier
- Version: 1 byte (0x01) - Protocol version
- Body Length: 1 byte (0-255) - Length of the body in bytes

Body Format:
- Type Tag: 1 byte - Identifies the data type
- Data: Variable length based on type

Type Tags:
0x01 = Int8 (1 byte data)
0x02 = Int16 (2 bytes data, big-endian)
0x03 = Int32 (4 bytes data, big-endian)
0x04 = Int64 (8 bytes data, big-endian)
0x05 = String (2-byte length + UTF-8 bytes, big-endian length)
0x06 = Bytes (2-byte length + raw bytes, big-endian length)
```

### Example

Encoding the Int16 value `256`:
```
Header:
0xBE 0xEF        // Magic number
0x01             // Version
0x03             // Body length (type tag + 2 bytes data)

Body:
0x02             // Type tag for Int16
0x01 0x00        // 256 in big-endian
```

### Bug Categories

1. **Endianness Bugs** - Incorrect byte order in encoding/decoding
2. **Buffer Management** - Buffer overflows, missing bounds checks
3. **Bit Manipulation** - Incorrect bit operations
4. **Validation** - Missing or incorrect validation logic
5. **Concurrency** - Race conditions in concurrent access
6. **Integer Overflow** - Type overflow in calculations
7. **Memory Safety** - Memory aliasing issues

Your comprehensive test suite should catch all of these bugs. Run the tests to identify failures, then fix the bugs in the implementation.

### Components to Debug

The following files contain bugs that need to be fixed:

#### `protocol.go`
Contains encoding and decoding functions:
- `EncodeInt8(w io.Writer, val int8) error` - Encode an 8-bit integer
- `EncodeInt16(w io.Writer, val int16) error` - Encode a 16-bit integer
- `EncodeInt32(w io.Writer, val int32) error` - Encode a 32-bit integer
- `EncodeInt64(w io.Writer, val int64) error` - Encode a 64-bit integer
- `EncodeString(w io.Writer, val string) error` - Encode a UTF-8 string
- `EncodeBytes(w io.Writer, val []byte) error` - Encode a byte array
- `DecodeMessage(r io.Reader) (typeTag byte, data interface{}, error)` - Decode a complete message

#### `buffer.go`
Contains buffer operations:
- `Buffer` type with read/write operations
- `NewBuffer(size int) *Buffer` - Create a new buffer
- `ReadByte() (byte, error)` - Read a single byte
- `ReadBytes(n int) ([]byte, error)` - Read n bytes
- `WriteByte(b byte) error` - Write a single byte
- `WriteBytes(data []byte) error` - Write multiple bytes
- `Remaining() int` - Returns remaining bytes to read
- `SetBit(b byte, pos uint, val bool) byte` - Set a specific bit in a byte
- `GetBit(b byte, pos uint) bool` - Get a specific bit from a byte

#### `validator.go`
Contains validation and statistics tracking:
- `ValidateHeader(magic uint16, version, bodyLen uint8) error` - Validate header fields
- `ValidateTypeTag(tag byte) error` - Validate type tag
- `ComputeChecksum(data []byte) uint16` - Compute a simple checksum
- `Stats` type for tracking parse/encode statistics

## Context

This challenge tests your debugging skills in the context of binary protocols and low-level data processing. You need to understand:

- **Binary data encoding/decoding** - Converting Go values to/from byte representations
- **Endianness** - Understanding byte order (big-endian vs little-endian)
- **Buffer management** - Handling byte slices and buffer boundaries
- **Bit manipulation** - Working with individual bits in bytes
- **io.Reader and io.Writer interfaces** - Go's standard I/O abstractions
- **Error handling** - Proper error handling in data serialization
- **Memory management** - Avoiding memory aliasing issues with byte slices
- **Concurrency** - Thread-safe access to shared data structures

Don't assume any particular file is bug-free - bugs are distributed across all three implementation files.

Common bugs in binary data processing include:
- Confusing big-endian and little-endian byte order
- Off-by-one errors in buffer indexing and length calculations
- Buffer overruns and underruns when reading/writing data
- Integer overflow in calculations
- Memory aliasing when slicing byte arrays
- Incorrect bit manipulation operations
- Race conditions in concurrent access to shared state

## Dependencies

- Go 1.21 or later
- Go standard library only:
  - `encoding/binary` for byte order operations
  - `bytes` for buffer operations
  - `io` for Reader/Writer interfaces
  - `fmt` for error formatting

## Constraints

- **Do not modify the test files** - they correctly test the specification
- **Fix only the bugs** - don't refactor or add new features
- All multi-byte integers must use **big-endian encoding** (as per specification)
- Implement proper error handling for all edge cases
- The `Stats` type must be safe for concurrent use
- Do not use any external dependencies beyond the Go standard library
- Follow Go naming conventions and idiomatic patterns


## Edge Cases

Your bug fixes must correctly handle the following edge cases (all tested):

- **Empty data**: Encoding and decoding empty strings and byte arrays
- **Zero values**: All integer types with value 0
- **Maximum values**: Int8/16/32/64 with their maximum and minimum values
- **UTF-8 strings**: Strings containing multi-byte UTF-8 characters (e.g., "日本語", "🚀")
- **String length boundaries**: Strings with exactly 255 bytes, 256 bytes, and longer
- **Buffer boundaries**: Reading/writing at exact buffer capacity limits
- **Invalid magic numbers**: Rejecting messages with incorrect magic numbers
- **Invalid type tags**: Rejecting messages with unrecognized type tags
- **Partial data**: Handling incomplete messages when not enough bytes are available
- **Bit operations**: Correctly setting and getting all bit positions (0-7) in a byte
- **Integer overflow**: Checksum calculations that may exceed the data type range
- **Memory aliasing**: Ensuring decoded byte slices don't share memory with internal buffers
- **Concurrent access**: Multiple goroutines accessing `Stats` simultaneously