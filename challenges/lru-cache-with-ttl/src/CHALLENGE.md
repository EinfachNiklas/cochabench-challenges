# LRU Cache with Time-to-Live

## Task

Implement a complete `LRU-Cache` class with time-to-live (TTL) support and thread safety.

Implement the following public API:

```python
class LRUCache:
    def __init__(self, capacity: int, default_ttl: Optional[float] = None)
    def get(self, key: str) -> Optional[Any]
    def put(self, key: str, value: Any, ttl: Optional[float] = None) -> None
    def delete(self, key: str) -> bool
    def clear(self) -> None
    def size(self) -> int
    def cleanup_expired(self) -> int
    def get_stats(self) -> dict
```

Detailed behavior:

### `__init__(capacity, default_ttl=None)`

- `capacity`: maximum number of entries, must be greater than `0`
- `default_ttl`: default TTL in seconds, `None` means no expiration
- Raise `ValueError` for an invalid capacity

### `get(key) -> Optional[Any]`

- Return the value stored for `key` or `None`
- Remove expired entries automatically
- Mark the entry as recently used on a cache hit
- Increase hit and miss counters accordingly

### `put(key, value, ttl=None)`

- Insert or update an entry
- `ttl`: custom TTL for this entry, `None` means use `default_ttl`
- If the cache is full, evict the least recently used entry
- Mark the inserted or updated entry as recently used

### `delete(key) -> bool`

- Remove the entry explicitly
- Return `True` if the key existed and was removed, otherwise `False`

### `clear()`

- Remove all entries from the cache

### `size() -> int`

- Return the number of currently valid entries

### `cleanup_expired() -> int`

- Remove all expired entries
- Return the number of removed entries

### `get_stats() -> dict`

Return a dictionary with the following keys:

- `'hits'`: number of successful `get()` calls
- `'misses'`: number of failed `get()` calls
- `'evictions'`: number of LRU evictions
- `'expired'`: number of TTL expirations
- `'size'`: current number of entries
- `'capacity'`: maximum cache capacity

## Context

An LRU cache is a common data structure for caching frequently accessed values. When the cache reaches capacity, the least recently used entry must be removed first.

This challenge extends a basic LRU cache with additional requirements:

- TTL support so entries can expire automatically
- Thread safety for multi-threaded environments
- Statistics tracking for hits, misses, evictions, and expirations

The challenge is intended to exercise data-structure design, expiration handling, synchronization, and API correctness.

## Dependencies

- Python 3
- The provided challenge requirements file lists the local test dependencies:
  `pytest>=7.4.0` and `pytest-cov>=4.1.0`

Typical local setup:

```bash
pip install -r src/requirements.txt
pytest
```

## Constraints

- Do not change the provided public API
- Do not modify the tests
- Preserve the expected LRU eviction behavior
- Preserve TTL semantics for default and per-entry expiration
- All public methods must remain safe for concurrent use

Implementation guidance:

- A doubly linked list plus hash map is a valid `O(1)` approach for `get()` and `put()`
- `collections.OrderedDict` is also acceptable if it preserves the required behavior
- Use synchronization such as `threading.Lock()` to protect critical sections
- `get_stats()` should return a snapshot of the current counters and capacity

Expected complexity targets:

- `get()`: average `O(1)`
- `put()`: average `O(1)`
- `cleanup_expired()`: `O(n)` in the worst case

## Edge Cases

- `capacity <= 0`
- Reading a missing key
- Overwriting an existing key
- Capacity `1`
- Expired entries being accessed through `get()`
- A custom TTL overriding the default TTL
- Entries with no expiration
- Cleanup removing only expired entries
- Concurrent reads and writes
- Concurrent evictions under load
