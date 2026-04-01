import unittest
import time
import threading
from src.lru_cache import LRUCache


class TestLRUCacheBasics(unittest.TestCase):
    """Basic functionality of the LRU Cache"""

    def test_init_with_valid_capacity(self):
        """Cache is initialized with valid capacity"""
        cache = LRUCache(capacity=5)
        self.assertEqual(cache.size(), 0)

    def test_init_with_invalid_capacity(self):
        """Raises ValueError for invalid capacity"""
        with self.assertRaises(ValueError):
            LRUCache(capacity=0)
        with self.assertRaises(ValueError):
            LRUCache(capacity=-1)

    def test_put_and_get(self):
        """Simple store and retrieve"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        self.assertEqual(cache.get("key1"), "value1")
        self.assertEqual(cache.size(), 1)

    def test_get_nonexistent_key(self):
        """Get returns None for non-existent keys"""
        cache = LRUCache(capacity=3)
        self.assertIsNone(cache.get("nonexistent"))

    def test_overwrite_existing_key(self):
        """Overwriting an existing key"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key1", "value2")
        self.assertEqual(cache.get("key1"), "value2")
        self.assertEqual(cache.size(), 1)


class TestLRUEviction(unittest.TestCase):
    """LRU Eviction Policy Tests"""

    def test_eviction_when_full(self):
        """Oldest entry is removed when cache is full"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")
        cache.put("key4", "value4")  # key1 should be evicted

        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.get("key2"), "value2")
        self.assertEqual(cache.get("key3"), "value3")
        self.assertEqual(cache.get("key4"), "value4")
        self.assertEqual(cache.size(), 3)

    def test_get_updates_recency(self):
        """Get updates the recency"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")

        # access key1 -> becomes "recent"
        cache.get("key1")

        # add key4 -> key2 should be evicted (not key1)
        cache.put("key4", "value4")

        self.assertEqual(cache.get("key1"), "value1")
        self.assertIsNone(cache.get("key2"))
        self.assertEqual(cache.get("key3"), "value3")
        self.assertEqual(cache.get("key4"), "value4")

    def test_put_updates_recency(self):
        """Put on existing key updates recency"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")

        # overwrite key1 -> becomes "recent"
        cache.put("key1", "updated")

        # add key4 -> key2 should be evicted
        cache.put("key4", "value4")

        self.assertEqual(cache.get("key1"), "updated")
        self.assertIsNone(cache.get("key2"))


class TestTTLFunctionality(unittest.TestCase):
    """Time-to-Live functionality tests"""

    def test_entry_expires_after_ttl(self):
        """Entry expires after TTL"""
        cache = LRUCache(capacity=5, default_ttl=0.1)  # 100ms
        cache.put("key1", "value1")

        # retrievable immediately
        self.assertEqual(cache.get("key1"), "value1")

        # no longer retrievable after TTL
        time.sleep(0.15)
        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.size(), 0)

    def test_custom_ttl_overrides_default(self):
        """Custom TTL overrides default_ttl"""
        cache = LRUCache(capacity=5, default_ttl=1.0)
        cache.put("key1", "value1", ttl=0.1)  # Shorter TTL

        time.sleep(0.15)
        self.assertIsNone(cache.get("key1"))

    def test_no_expiry_with_none_ttl(self):
        """Entries without TTL do not expire"""
        cache = LRUCache(capacity=5, default_ttl=None)
        cache.put("key1", "value1")

        time.sleep(0.1)
        self.assertEqual(cache.get("key1"), "value1")

    def test_cleanup_expired_removes_only_expired(self):
        """cleanup_expired removes only expired entries"""
        cache = LRUCache(capacity=5, default_ttl=0.1)
        cache.put("key1", "value1")  # will expire
        cache.put("key2", "value2", ttl=None)  # will not expire
        cache.put("key3", "value3")  # will expire

        time.sleep(0.15)
        removed = cache.cleanup_expired()

        self.assertEqual(removed, 2)
        self.assertEqual(cache.size(), 1)
        self.assertEqual(cache.get("key2"), "value2")


class TestCacheOperations(unittest.TestCase):
    """Tests for cache operations"""

    def test_delete_existing_key(self):
        """Deleting an existing key"""
        cache = LRUCache(capacity=5)
        cache.put("key1", "value1")
        result = cache.delete("key1")

        self.assertTrue(result)
        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.size(), 0)

    def test_delete_nonexistent_key(self):
        """Deleting a non-existent key"""
        cache = LRUCache(capacity=5)
        result = cache.delete("nonexistent")
        self.assertFalse(result)

    def test_clear_removes_all(self):
        """Clear removes all entries"""
        cache = LRUCache(capacity=5)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")

        cache.clear()

        self.assertEqual(cache.size(), 0)
        self.assertIsNone(cache.get("key1"))
        self.assertIsNone(cache.get("key2"))
        self.assertIsNone(cache.get("key3"))


class TestStatistics(unittest.TestCase):
    """Tests for cache statistics"""

    def test_stats_tracks_hits_and_misses(self):
        """Statistics track hits and misses"""
        cache = LRUCache(capacity=5)
        cache.put("key1", "value1")

        cache.get("key1")  # Hit
        cache.get("key2")  # Miss
        cache.get("key1")  # Hit
        cache.get("key3")  # Miss

        stats = cache.get_stats()
        self.assertEqual(stats['hits'], 2)
        self.assertEqual(stats['misses'], 2)

    def test_stats_tracks_evictions(self):
        """Statistics track LRU evictions"""
        cache = LRUCache(capacity=2)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")  # Eviction

        stats = cache.get_stats()
        self.assertEqual(stats['evictions'], 1)
        self.assertEqual(stats['size'], 2)
        self.assertEqual(stats['capacity'], 2)

    def test_stats_tracks_expired_entries(self):
        """Statistics track expired entries"""
        cache = LRUCache(capacity=5, default_ttl=0.1)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        time.sleep(0.15)
        cache.get("key1")  # triggers expired check
        cache.get("key2")  # triggers expired check

        stats = cache.get_stats()
        self.assertEqual(stats['expired'], 2)


class TestThreadSafety(unittest.TestCase):
    """Thread-safety tests"""

    def test_concurrent_puts_and_gets(self):
        """Concurrent puts and gets are thread-safe"""
        cache = LRUCache(capacity=100)
        errors = []

        def worker(thread_id):
            try:
                for i in range(50):
                    key = f"key_{thread_id}_{i}"
                    cache.put(key, f"value_{thread_id}_{i}")
                    value = cache.get(key)
                    if value != f"value_{thread_id}_{i}":
                        errors.append(f"Thread {thread_id}: Got wrong value")
            except Exception as e:
                errors.append(f"Thread {thread_id}: {str(e)}")

        threads = [threading.Thread(target=worker, args=(i,)) for i in range(5)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        self.assertEqual(len(errors), 0, f"Thread safety errors: {errors}")

    def test_concurrent_evictions(self):
        """Concurrent evictions work correctly"""
        cache = LRUCache(capacity=10)
        errors = []

        def worker(thread_id):
            try:
                for i in range(20):
                    cache.put(f"key_{thread_id}_{i}", f"value_{thread_id}_{i}")
                    time.sleep(0.001)
            except Exception as e:
                errors.append(f"Thread {thread_id}: {str(e)}")

        threads = [threading.Thread(target=worker, args=(i,)) for i in range(3)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        self.assertEqual(len(errors), 0)
        self.assertLessEqual(cache.size(), 10)


class TestEdgeCases(unittest.TestCase):
    """Edge cases and special scenarios"""

    def test_capacity_one(self):
        """Cache with capacity 1"""
        cache = LRUCache(capacity=1)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.get("key2"), "value2")
        self.assertEqual(cache.size(), 1)

    def test_none_values(self):
        """Store None as value"""
        cache = LRUCache(capacity=5)
        cache.put("key1", None)

        # None value should be distinguishable from "not found"
        # implementation may vary here
        self.assertEqual(cache.size(), 1)

    def test_complex_objects(self):
        """Complex objects as values"""
        cache = LRUCache(capacity=5)
        obj = {"nested": {"data": [1, 2, 3]}}
        cache.put("key1", obj)

        retrieved = cache.get("key1")
        self.assertEqual(retrieved, obj)


if __name__ == '__main__':
    unittest.main()
