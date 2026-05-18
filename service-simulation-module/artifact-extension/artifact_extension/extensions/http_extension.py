"""HTTP artifact extension.

Generates two persistent artifact types:

1. **Browser History SQLite database** (Chrome-compatible schema)
   Written to ``<artifacts_path>/browser/History``.
   The schema mirrors Chrome's real ``History`` file (tables ``urls``
   and ``visits``) so that tools and malware that query it directly
   get plausible data.

2. **Zone.Identifier files** for ``file_download`` events
   Written alongside placeholder files under
   ``<artifacts_path>/browser/downloads/``.
   Each ``<filename>.Zone.Identifier`` file carries ``ZoneId=3``,
   ``ReferrerUrl``, and ``HostUrl`` — the fields that Mark-of-the-Web
   checkers look for.

Key implementation decisions
------------------------------
* Chrome stores timestamps as microseconds since **1601-01-01** (not the
  Unix epoch).  The helper ``_unix_to_chrome_time`` handles the conversion.
  Using the wrong epoch is one of the most common sandbox fingerprinting
  tells.
* Visit ``transition`` values are drawn from the BrowserProfile weight
  distribution so the visit table contains a realistic mix of LINK,
  TYPED, FORM_SUBMIT, etc. entries — not a uniform ``0`` column.
* Visit durations are randomised between 5 s and 5 min (in microseconds)
  to avoid the ``0``-duration tell present in many fake history DBs.
* History timestamps are spread across ``days_of_history`` days using an
  exponential distribution skewed toward recent activity.
"""

from __future__ import annotations

import random
import sqlite3
import time
from dataclasses import dataclass, field
from pathlib import Path

from artifact_extension.base import ArtifactExtension
from artifact_extension.profile import BrowserProfile, UserProfile

# ---------------------------------------------------------------------------
# Chrome epoch conversion
# ---------------------------------------------------------------------------

# Microseconds between 1601-01-01 (Chrome epoch) and 1970-01-01 (Unix epoch)
_CHROME_EPOCH_OFFSET_US: int = 11_644_473_600 * 1_000_000

# Chrome visit transition type constants (subset of net/base/page_transition_types.h)
_TRANSITION_TYPES: dict[str, int] = {
    "LINK":           0,
    "TYPED":          1,
    "AUTO_BOOKMARK":  2,
    "AUTO_SUBFRAME":  3,
    "FORM_SUBMIT":    7,
    "RELOAD":         8,
}


def _unix_to_chrome_time(unix_ts: float) -> int:
    """Convert a Unix timestamp (seconds) to Chrome's microsecond epoch."""
    return int(unix_ts * 1_000_000) + _CHROME_EPOCH_OFFSET_US


# ---------------------------------------------------------------------------
# Background URL pool — public, benign sites
# ---------------------------------------------------------------------------

_BACKGROUND_URLS: list[tuple[str, str]] = [
    ("https://www.google.com/",                   "Google"),
    ("https://mail.google.com/mail/u/0/",         "Gmail"),
    ("https://calendar.google.com/",              "Google Calendar"),
    ("https://drive.google.com/drive/my-drive",   "Google Drive"),
    ("https://github.com/",                       "GitHub"),
    ("https://github.com/explore",                "Explore · GitHub"),
    ("https://stackoverflow.com/",                "Stack Overflow"),
    ("https://stackoverflow.com/questions",       "Questions - Stack Overflow"),
    ("https://www.reddit.com/",                   "Reddit"),
    ("https://www.reddit.com/r/Python/",          "r/Python"),
    ("https://www.youtube.com/",                  "YouTube"),
    ("https://docs.python.org/3/",                "Python 3 Documentation"),
    ("https://docs.python.org/3/library/",        "The Python Standard Library"),
    ("https://pypi.org/",                         "PyPI"),
    ("https://pypi.org/simple/requests/",         "requests · PyPI"),
    ("https://www.microsoft.com/",                "Microsoft"),
    ("https://news.ycombinator.com/",             "Hacker News"),
    ("https://medium.com/",                       "Medium"),
    ("https://twitter.com/",                      "Twitter / X"),
    ("https://www.linkedin.com/feed/",            "LinkedIn"),
    ("https://en.wikipedia.org/wiki/Main_Page",   "Wikipedia"),
    ("https://cloudflare.com/",                   "Cloudflare"),
    ("https://www.amazon.com/",                   "Amazon"),
    ("https://outlook.live.com/mail/",            "Outlook"),
    ("https://notion.so/",                        "Notion"),
]


# ---------------------------------------------------------------------------
# Internal data models
# ---------------------------------------------------------------------------

@dataclass
class _HistoryEntry:
    url: str
    title: str
    visit_time: float       # Unix timestamp (seconds)
    transition: int         # Chrome transition type integer


@dataclass
class _DownloadEntry:
    filename: str
    url: str
    referrer_url: str
    timestamp: float        # Unix timestamp (seconds)


# ---------------------------------------------------------------------------
# Extension
# ---------------------------------------------------------------------------

class HTTPArtifactExtension(ArtifactExtension):
    """Generates browser history and Zone.Identifier artifacts.

    Args:
        profile: Statistical user profile controlling history size,
                 timestamp spread, and transition distribution.
    """

    def __init__(self, profile: UserProfile) -> None:
        self._profile = profile
        self._http_events: list[_HistoryEntry] = []
        self._download_events: list[_DownloadEntry] = []

    # ------------------------------------------------------------------
    # ArtifactExtension interface
    # ------------------------------------------------------------------

    def on_service_event(self, event: dict) -> None:
        """Capture ``http_request`` and ``file_download`` events.

        All other event types are silently ignored.
        """
        etype = event.get("type")

        if etype == "http_request":
            url: str = event.get("url", "http://unknown/")
            ts: float = float(event.get("timestamp", time.time()))
            transition_name: str = event.get("transition", "LINK")
            self._http_events.append(
                _HistoryEntry(
                    url=url,
                    title=event.get("title", url),
                    visit_time=ts,
                    transition=_TRANSITION_TYPES.get(transition_name, 0),
                )
            )

        elif etype == "file_download":
            filename: str = event.get("filename", "file.bin")
            ts = float(event.get("timestamp", time.time()))
            self._download_events.append(
                _DownloadEntry(
                    filename=filename,
                    url=event.get("url", "http://unknown/"),
                    referrer_url=event.get("referrer", ""),
                    timestamp=ts,
                )
            )

    def inject(self, artifacts_path: str) -> None:
        """Write history DB and Zone.Identifier files.

        Args:
            artifacts_path: Root path of the Docker volume.
        """
        browser_dir = Path(artifacts_path) / "browser"
        browser_dir.mkdir(parents=True, exist_ok=True)

        self._inject_history(browser_dir)
        self._inject_zone_identifiers(browser_dir)

    # ------------------------------------------------------------------
    # History DB
    # ------------------------------------------------------------------

    def _inject_history(self, browser_dir: Path) -> None:
        """Create a Chrome-compatible History SQLite database."""
        db_path = browser_dir / "History"
        # Always start from a fresh DB so injection is idempotent
        db_path.unlink(missing_ok=True)

        conn = sqlite3.connect(str(db_path))
        try:
            self._create_schema(conn)
            entries = self._build_history_entries()
            self._write_entries(conn, entries)
            conn.commit()
        finally:
            conn.close()

    @staticmethod
    def _create_schema(conn: sqlite3.Connection) -> None:
        """Create the minimal Chrome History schema."""
        conn.executescript(
            """
            CREATE TABLE IF NOT EXISTS urls (
                id               INTEGER PRIMARY KEY AUTOINCREMENT,
                url              TEXT    NOT NULL,
                title            TEXT    DEFAULT '',
                visit_count      INTEGER DEFAULT 0,
                typed_count      INTEGER DEFAULT 0,
                last_visit_time  INTEGER NOT NULL,
                hidden           INTEGER DEFAULT 0
            );

            CREATE TABLE IF NOT EXISTS visits (
                id               INTEGER PRIMARY KEY AUTOINCREMENT,
                url              INTEGER NOT NULL,
                visit_time       INTEGER NOT NULL,
                from_visit       INTEGER DEFAULT 0,
                transition       INTEGER DEFAULT 0,
                segment_id       INTEGER DEFAULT 0,
                visit_duration   INTEGER DEFAULT 0
            );

            CREATE INDEX IF NOT EXISTS visits_url_index  ON visits (url);
            CREATE INDEX IF NOT EXISTS visits_time_index ON visits (visit_time);
            """
        )

    def _build_history_entries(self) -> list[_HistoryEntry]:
        """Combine statistical background traffic with event-driven entries."""
        bp: BrowserProfile = self._profile.browser
        now = time.time()
        entries: list[_HistoryEntry] = []

        # Background history spread over `days_of_history` days.
        # Exponential distribution creates a natural "more recent pages
        # visited more often" pattern instead of a flat uniform spread.
        target = max(0, bp.history_entries - len(self._http_events))
        mean_age = bp.days_of_history * 86400 / 3   # skew toward recent
        for _ in range(target):
            url, title = random.choice(_BACKGROUND_URLS)
            age = min(random.expovariate(1.0 / mean_age), bp.days_of_history * 86400)
            visit_ts = now - age
            transition = self._pick_transition(bp)
            entries.append(
                _HistoryEntry(url=url, title=title, visit_time=visit_ts, transition=transition)
            )

        # Append event-driven entries (highest realism signal)
        entries.extend(self._http_events)
        return entries

    @staticmethod
    def _pick_transition(bp: BrowserProfile) -> int:
        """Sample a Chrome transition integer from the profile weights."""
        names = list(bp.transition_weights.keys())
        weights = [bp.transition_weights[n] for n in names]
        chosen = random.choices(names, weights=weights, k=1)[0]
        return _TRANSITION_TYPES.get(chosen, 0)

    def _write_entries(
        self,
        conn: sqlite3.Connection,
        entries: list[_HistoryEntry],
    ) -> None:
        """Insert all history entries, maintaining URL deduplication."""
        # Maps URL → urls.id for deduplication
        url_id_map: dict[str, int] = {}

        for entry in sorted(entries, key=lambda x: x.visit_time):
            chrome_ts = _unix_to_chrome_time(entry.visit_time)
            # Visit duration: realistic random value between 5 s and 5 min
            duration_us = random.randint(5_000_000, 300_000_000)

            if entry.url not in url_id_map:
                cursor = conn.execute(
                    """
                    INSERT INTO urls (url, title, visit_count, last_visit_time)
                    VALUES (?, ?, 1, ?)
                    """,
                    (entry.url, entry.title, chrome_ts),
                )
                url_id_map[entry.url] = cursor.lastrowid  # type: ignore[assignment]
            else:
                uid = url_id_map[entry.url]
                conn.execute(
                    """
                    UPDATE urls
                    SET visit_count     = visit_count + 1,
                        last_visit_time = MAX(last_visit_time, ?)
                    WHERE id = ?
                    """,
                    (chrome_ts, uid),
                )

            conn.execute(
                """
                INSERT INTO visits (url, visit_time, transition, visit_duration)
                VALUES (?, ?, ?, ?)
                """,
                (url_id_map[entry.url], chrome_ts, entry.transition, duration_us),
            )

    # ------------------------------------------------------------------
    # Zone.Identifier files
    # ------------------------------------------------------------------

    def _inject_zone_identifiers(self, browser_dir: Path) -> None:
        """Write Zone.Identifier ADS simulation files for every download event.

        On Windows, the real NTFS ADS would be stored as ``filename:Zone.Identifier``.
        Inside a Linux Docker container we write a companion file named
        ``<filename>.Zone.Identifier`` so the artifact is volume-portable.

        Each file contains the fields that Mark-of-the-Web checkers look for:
        ``ZoneId=3`` (Internet zone), ``ReferrerUrl``, and ``HostUrl``.
        """
        downloads_dir = browser_dir / "downloads"
        downloads_dir.mkdir(exist_ok=True)

        for dl in self._download_events:
            fname = Path(dl.filename).name
            # Zero-byte placeholder simulates the presence of the downloaded file
            (downloads_dir / fname).touch()
            # Zone.Identifier companion file
            zone_path = downloads_dir / f"{fname}.Zone.Identifier"
            zone_path.write_text(
                "[ZoneTransfer]\n"
                "ZoneId=3\n"
                f"ReferrerUrl={dl.referrer_url}\n"
                f"HostUrl={dl.url}\n",
                encoding="utf-8",
            )
