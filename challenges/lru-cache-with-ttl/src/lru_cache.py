"""
LRU Cache mit Time-to-Live (TTL) Implementierung.

Implementiere eine Cache-Klasse, die folgende Features unterstützt:
- Least Recently Used (LRU) Eviction Policy
- Time-to-Live (TTL) für Cache-Einträge
- Thread-Safety
- Maximale Cache-Größe

Die Cache-Klasse soll die folgenden Methoden implementieren:
"""

from typing import Any, Optional
from datetime import datetime


class LRUCache:
    """
    Ein LRU-Cache mit Time-to-Live Unterstützung.

    Attributes:
        capacity: Maximale Anzahl der Einträge im Cache
        default_ttl: Standard Time-to-Live in Sekunden (None = kein Ablauf)
    """

    def __init__(self, capacity: int, default_ttl: Optional[float] = None):
        """
        Initialisiert den LRU Cache.

        Args:
            capacity: Maximale Anzahl der Cache-Einträge (muss > 0 sein)
            default_ttl: Standard TTL in Sekunden (None = unbegrenzt)

        Raises:
            ValueError: Wenn capacity <= 0
        """
        # TODO: Implementiere die Initialisierung
        pass

    def get(self, key: str) -> Optional[Any]:
        """
        Holt einen Wert aus dem Cache.

        Args:
            key: Der Schlüssel des gewünschten Eintrags

        Returns:
            Der gespeicherte Wert oder None wenn:
            - Der Key nicht existiert
            - Der Eintrag abgelaufen ist (TTL überschritten)

        Side Effects:
            - Abgelaufene Einträge werden entfernt
            - Bei Cache-Hit wird der Eintrag als "recently used" markiert
        """
        # TODO: Implementiere get
        pass

    def put(self, key: str, value: Any, ttl: Optional[float] = None) -> None:
        """
        Fügt einen Eintrag in den Cache ein oder aktualisiert ihn.

        Args:
            key: Der Schlüssel
            value: Der zu speichernde Wert
            ttl: Time-to-Live in Sekunden (None = default_ttl verwenden)

        Side Effects:
            - Wenn der Cache voll ist, wird der am längsten nicht verwendete
              Eintrag entfernt (LRU)
            - Existierende Keys werden überschrieben und als "recently used" markiert
        """
        # TODO: Implementiere put
        pass

    def delete(self, key: str) -> bool:
        """
        Entfernt einen Eintrag aus dem Cache.

        Args:
            key: Der zu entfernende Schlüssel

        Returns:
            True wenn der Eintrag existierte und entfernt wurde, sonst False
        """
        # TODO: Implementiere delete
        pass

    def clear(self) -> None:
        """
        Entfernt alle Einträge aus dem Cache.
        """
        # TODO: Implementiere clear
        pass

    def size(self) -> int:
        """
        Gibt die aktuelle Anzahl der Einträge im Cache zurück.

        Returns:
            Anzahl der Einträge (ohne abgelaufene Einträge)
        """
        # TODO: Implementiere size
        pass

    def cleanup_expired(self) -> int:
        """
        Entfernt alle abgelaufenen Einträge aus dem Cache.

        Returns:
            Anzahl der entfernten Einträge
        """
        # TODO: Implementiere cleanup_expired
        pass

    def get_stats(self) -> dict:
        """
        Gibt Statistiken über den Cache zurück.

        Returns:
            Dictionary mit folgenden Keys:
            - 'hits': Anzahl erfolgreicher get-Aufrufe
            - 'misses': Anzahl fehlgeschlagener get-Aufrufe
            - 'evictions': Anzahl der LRU-Evictions
            - 'expired': Anzahl der wegen TTL entfernten Einträge
            - 'size': Aktuelle Anzahl der Einträge
            - 'capacity': Maximale Kapazität
        """
        # TODO: Implementiere get_stats
        pass
