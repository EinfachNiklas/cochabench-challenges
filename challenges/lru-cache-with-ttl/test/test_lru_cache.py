import unittest
import time
import threading
from src.lru_cache import LRUCache


class TestLRUCacheBasics(unittest.TestCase):
    """Grundlegende Funktionalität des LRU Cache"""

    def test_init_with_valid_capacity(self):
        """Cache wird mit gültiger Kapazität initialisiert"""
        cache = LRUCache(capacity=5)
        self.assertEqual(cache.size(), 0)

    def test_init_with_invalid_capacity(self):
        """Wirft ValueError bei ungültiger Kapazität"""
        with self.assertRaises(ValueError):
            LRUCache(capacity=0)
        with self.assertRaises(ValueError):
            LRUCache(capacity=-1)

    def test_put_and_get(self):
        """Einfaches Speichern und Abrufen"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        self.assertEqual(cache.get("key1"), "value1")
        self.assertEqual(cache.size(), 1)

    def test_get_nonexistent_key(self):
        """Get gibt None zurück für nicht existierende Keys"""
        cache = LRUCache(capacity=3)
        self.assertIsNone(cache.get("nonexistent"))

    def test_overwrite_existing_key(self):
        """Überschreiben eines existierenden Keys"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key1", "value2")
        self.assertEqual(cache.get("key1"), "value2")
        self.assertEqual(cache.size(), 1)


class TestLRUEviction(unittest.TestCase):
    """LRU Eviction Policy Tests"""

    def test_eviction_when_full(self):
        """Ältester Eintrag wird entfernt wenn Cache voll"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")
        cache.put("key4", "value4")  # key1 sollte entfernt werden

        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.get("key2"), "value2")
        self.assertEqual(cache.get("key3"), "value3")
        self.assertEqual(cache.get("key4"), "value4")
        self.assertEqual(cache.size(), 3)

    def test_get_updates_recency(self):
        """Get aktualisiert die Recency"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")

        # key1 zugreifen -> wird "recent"
        cache.get("key1")

        # key4 hinzufügen -> key2 sollte entfernt werden (nicht key1)
        cache.put("key4", "value4")

        self.assertEqual(cache.get("key1"), "value1")
        self.assertIsNone(cache.get("key2"))
        self.assertEqual(cache.get("key3"), "value3")
        self.assertEqual(cache.get("key4"), "value4")

    def test_put_updates_recency(self):
        """Put auf existierenden Key aktualisiert Recency"""
        cache = LRUCache(capacity=3)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")

        # key1 überschreiben -> wird "recent"
        cache.put("key1", "updated")

        # key4 hinzufügen -> key2 sollte entfernt werden
        cache.put("key4", "value4")

        self.assertEqual(cache.get("key1"), "updated")
        self.assertIsNone(cache.get("key2"))


class TestTTLFunctionality(unittest.TestCase):
    """Time-to-Live Funktionalität Tests"""

    def test_entry_expires_after_ttl(self):
        """Eintrag läuft nach TTL ab"""
        cache = LRUCache(capacity=5, default_ttl=0.1)  # 100ms
        cache.put("key1", "value1")

        # Sofort abrufbar
        self.assertEqual(cache.get("key1"), "value1")

        # Nach TTL nicht mehr abrufbar
        time.sleep(0.15)
        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.size(), 0)

    def test_custom_ttl_overrides_default(self):
        """Custom TTL überschreibt default_ttl"""
        cache = LRUCache(capacity=5, default_ttl=1.0)
        cache.put("key1", "value1", ttl=0.1)  # Kürzere TTL

        time.sleep(0.15)
        self.assertIsNone(cache.get("key1"))

    def test_no_expiry_with_none_ttl(self):
        """Einträge ohne TTL laufen nicht ab"""
        cache = LRUCache(capacity=5, default_ttl=None)
        cache.put("key1", "value1")

        time.sleep(0.1)
        self.assertEqual(cache.get("key1"), "value1")

    def test_cleanup_expired_removes_only_expired(self):
        """cleanup_expired entfernt nur abgelaufene Einträge"""
        cache = LRUCache(capacity=5, default_ttl=0.1)
        cache.put("key1", "value1")  # Wird ablaufen
        cache.put("key2", "value2", ttl=None)  # Läuft nicht ab
        cache.put("key3", "value3")  # Wird ablaufen

        time.sleep(0.15)
        removed = cache.cleanup_expired()

        self.assertEqual(removed, 2)
        self.assertEqual(cache.size(), 1)
        self.assertEqual(cache.get("key2"), "value2")


class TestCacheOperations(unittest.TestCase):
    """Tests für Cache-Operationen"""

    def test_delete_existing_key(self):
        """Löschen eines existierenden Keys"""
        cache = LRUCache(capacity=5)
        cache.put("key1", "value1")
        result = cache.delete("key1")

        self.assertTrue(result)
        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.size(), 0)

    def test_delete_nonexistent_key(self):
        """Löschen eines nicht existierenden Keys"""
        cache = LRUCache(capacity=5)
        result = cache.delete("nonexistent")
        self.assertFalse(result)

    def test_clear_removes_all(self):
        """Clear entfernt alle Einträge"""
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
    """Tests für Cache-Statistiken"""

    def test_stats_tracks_hits_and_misses(self):
        """Statistiken tracken Hits und Misses"""
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
        """Statistiken tracken LRU Evictions"""
        cache = LRUCache(capacity=2)
        cache.put("key1", "value1")
        cache.put("key2", "value2")
        cache.put("key3", "value3")  # Eviction

        stats = cache.get_stats()
        self.assertEqual(stats['evictions'], 1)
        self.assertEqual(stats['size'], 2)
        self.assertEqual(stats['capacity'], 2)

    def test_stats_tracks_expired_entries(self):
        """Statistiken tracken abgelaufene Einträge"""
        cache = LRUCache(capacity=5, default_ttl=0.1)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        time.sleep(0.15)
        cache.get("key1")  # Triggert Expired-Check
        cache.get("key2")  # Triggert Expired-Check

        stats = cache.get_stats()
        self.assertEqual(stats['expired'], 2)


class TestThreadSafety(unittest.TestCase):
    """Thread-Safety Tests"""

    def test_concurrent_puts_and_gets(self):
        """Concurrent Puts und Gets sind thread-safe"""
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
        """Concurrent Evictions funktionieren korrekt"""
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
    """Edge Cases und Spezialfälle"""

    def test_capacity_one(self):
        """Cache mit Kapazität 1"""
        cache = LRUCache(capacity=1)
        cache.put("key1", "value1")
        cache.put("key2", "value2")

        self.assertIsNone(cache.get("key1"))
        self.assertEqual(cache.get("key2"), "value2")
        self.assertEqual(cache.size(), 1)

    def test_none_values(self):
        """None als Wert speichern"""
        cache = LRUCache(capacity=5)
        cache.put("key1", None)

        # None-Wert sollte von "nicht gefunden" unterscheidbar sein
        # Implementierung kann hier variieren
        self.assertEqual(cache.size(), 1)

    def test_complex_objects(self):
        """Komplexe Objekte als Values"""
        cache = LRUCache(capacity=5)
        obj = {"nested": {"data": [1, 2, 3]}}
        cache.put("key1", obj)

        retrieved = cache.get("key1")
        self.assertEqual(retrieved, obj)


if __name__ == '__main__':
    unittest.main()
