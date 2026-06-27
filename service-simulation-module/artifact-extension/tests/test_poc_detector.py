"""PoC sandbox detector test suite.

Tests the three scenarios defined in the prompt:

+--------+-----------------------------------------------+---------------------------+
| Case   | Setup                                         | Expected result           |
+========+===============================================+===========================+
| A      | Clean sandbox — no artifacts                  | Sandbox detected          |
+--------+-----------------------------------------------+---------------------------+
| B      | Service simulation only (events, no inject)   | Sandbox still detected    |
+--------+-----------------------------------------------+---------------------------+
| C      | Service simulation + artifact injection       | Appears as a real system  |
+--------+-----------------------------------------------+---------------------------+

Additional tests
----------------
* Freshness — two independent injection runs must produce different artifacts
  (randomised TTLs and timestamps must not collide).
* Zone.Identifier — every ``file_download`` event must produce a companion
  ``.Zone.Identifier`` file with the required fields.

Run with::

    pytest tests/ -v
"""

from __future__ import annotations

import os
import sqlite3
import tempfile
import time
from pathlib import Path

import pytest

from artifact_extension.extensions.dns_extension import DNSArtifactExtension
from artifact_extension.extensions.http_extension import HTTPArtifactExtension
from artifact_extension.extensions.linux_extension import LinuxArtifactExtension
from artifact_extension.manager import ArtifactManager
from artifact_extension.profile import UserProfile


# ---------------------------------------------------------------------------
# Safe PoC sandbox detector
#
# Mimics the checks a real fingerprinting routine would perform.
# Returns True  → sandbox detected (bad for stealth)
#         False → environment looks real (good for stealth)
# ---------------------------------------------------------------------------

class SandboxDetector:
    """Safe PoC detector — checks artifact plausibility only.

    No network calls, no privilege escalation, no real malicious logic.
    """

    def __init__(self, artifacts_path: str) -> None:
        self._root = Path(artifacts_path)

    # ------------------------------------------------------------------
    # Individual checks
    # ------------------------------------------------------------------

    def check_dns_cache_missing(self) -> bool:
        """True if the DNS cache file is absent."""
        return not (self._root / "dns" / "cache.txt").exists()

    def check_dns_cache_sparse(self) -> bool:
        """True if the DNS cache has fewer than 10 entries — suspiciously bare."""
        cache = self._root / "dns" / "cache.txt"
        if not cache.exists():
            return True
        return cache.read_text().count("Record Name") < 10

    def check_dns_ttl_uniform(self) -> bool:
        """True if all DNS TTL values are identical — bulk injection fingerprint."""
        cache = self._root / "dns" / "cache.txt"
        if not cache.exists():
            return True
        ttls = [
            int(line.split(":")[-1].strip())
            for line in cache.read_text().splitlines()
            if "Time To Live" in line
        ]
        if not ttls:
            return True
        # If all TTLs are within a 60-second window it looks batch-injected
        return (max(ttls) - min(ttls)) < 60

    def check_browser_history_missing(self) -> bool:
        """True if the browser history database is absent."""
        return not (self._root / "browser" / "History").exists()

    def check_browser_history_sparse(self) -> bool:
        """True if the history DB has fewer than 20 URL rows."""
        db = self._root / "browser" / "History"
        if not db.exists():
            return True
        conn = sqlite3.connect(str(db))
        try:
            (count,) = conn.execute("SELECT COUNT(*) FROM urls").fetchone()
            return count < 20
        finally:
            conn.close()

    def check_transition_uniform(self) -> bool:
        """True if all visits share the same transition type (looks synthetic)."""
        db = self._root / "browser" / "History"
        if not db.exists():
            return True
        conn = sqlite3.connect(str(db))
        try:
            rows = conn.execute(
                "SELECT DISTINCT transition FROM visits"
            ).fetchall()
            return len(rows) < 2
        finally:
            conn.close()

    def is_sandbox(self) -> bool:
        """Aggregate gate — True if *any* check flags the environment."""
        return (
            self.check_dns_cache_missing()
            or self.check_dns_cache_sparse()
            or self.check_dns_ttl_uniform()
            or self.check_browser_history_missing()
            or self.check_browser_history_sparse()
            or self.check_transition_uniform()
        )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_manager(artifacts_path: str, profile: UserProfile) -> ArtifactManager:
    manager = ArtifactManager(artifacts_path=artifacts_path, profile=profile)
    manager.register(DNSArtifactExtension(profile))
    manager.register(HTTPArtifactExtension(profile))
    manager.register(LinuxArtifactExtension(profile))
    return manager


def _dispatch_sample_events(manager: ArtifactManager) -> None:
    now = time.time()
    manager.dispatch({
        "type": "dns_query", "domain": "google.com",
        "resolved_ip": "142.250.80.46", "timestamp": now,
    })
    manager.dispatch({
        "type": "http_request", "url": "https://www.google.com/",
        "method": "GET", "transition": "TYPED", "timestamp": now,
    })


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def tmp_artifacts(tmp_path: Path) -> str:
    return str(tmp_path / "artifacts")


@pytest.fixture
def normal_profile() -> UserProfile:
    return UserProfile.default_normal()


# ---------------------------------------------------------------------------
# Case A — Clean sandbox
# ---------------------------------------------------------------------------

def test_case_a_clean_sandbox_detected(tmp_artifacts: str) -> None:
    """Case A: No artifacts at all → sandbox detector fires."""
    detector = SandboxDetector(tmp_artifacts)
    assert detector.is_sandbox(), (
        "A clean sandbox with no artifacts should be flagged as a sandbox"
    )


def test_case_a_individual_checks(tmp_artifacts: str) -> None:
    """Case A: Each individual check should independently fail on a clean path."""
    d = SandboxDetector(tmp_artifacts)
    assert d.check_dns_cache_missing()
    assert d.check_browser_history_missing()


# ---------------------------------------------------------------------------
# Case B — Service simulation only (events dispatched, inject NOT called)
# ---------------------------------------------------------------------------

def test_case_b_service_only_still_detected(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Case B: Receiving events without injecting artifacts → still detected."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    _dispatch_sample_events(manager)
    # Deliberately do NOT call manager.inject_all()

    detector = SandboxDetector(tmp_artifacts)
    assert detector.is_sandbox(), (
        "Service-only simulation (no inject) should still be detected"
    )


# ---------------------------------------------------------------------------
# Case C — Full artifact injection
# ---------------------------------------------------------------------------

def test_case_c_artifact_injection_bypasses_detection(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Case C: After full injection the environment must NOT be detected."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    _dispatch_sample_events(manager)
    manager.inject_all()

    detector = SandboxDetector(tmp_artifacts)
    assert not detector.is_sandbox(), (
        "After artifact injection the sandbox should NOT be detected"
    )


def test_case_c_dns_checks_pass(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Case C: All DNS-specific checks must pass individually."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    _dispatch_sample_events(manager)
    manager.inject_all()

    d = SandboxDetector(tmp_artifacts)
    assert not d.check_dns_cache_missing(),  "DNS cache file must exist"
    assert not d.check_dns_cache_sparse(),   "DNS cache must have ≥10 entries"
    assert not d.check_dns_ttl_uniform(),    "DNS TTL values must vary"


def test_case_c_browser_checks_pass(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Case C: All browser-history checks must pass individually."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    _dispatch_sample_events(manager)
    manager.inject_all()

    d = SandboxDetector(tmp_artifacts)
    assert not d.check_browser_history_missing(), "History DB must exist"
    assert not d.check_browser_history_sparse(),  "History DB must have ≥20 URLs"
    assert not d.check_transition_uniform(),       "Visit transitions must be diverse"


# ---------------------------------------------------------------------------
# Freshness test
# ---------------------------------------------------------------------------

def test_artifact_freshness_dns() -> None:
    """Two separate injection runs must produce different DNS cache content."""
    profile = UserProfile.default_normal()

    def run(d: str) -> str:
        mgr = ArtifactManager(artifacts_path=d, profile=profile)
        mgr.register(DNSArtifactExtension(profile))
        mgr.inject_all()
        return (Path(d) / "dns" / "cache.txt").read_text()

    with tempfile.TemporaryDirectory() as d1, tempfile.TemporaryDirectory() as d2:
        a1 = run(os.path.join(d1, "art"))
        a2 = run(os.path.join(d2, "art"))
        assert a1 != a2, (
            "DNS cache must differ across runs — randomised TTLs prevent "
            "the 'all injected at once' fingerprint"
        )


# ---------------------------------------------------------------------------
# Zone.Identifier test
# ---------------------------------------------------------------------------

def test_zone_identifier_created_for_download(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """A Zone.Identifier file must be created for every file_download event."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    manager.dispatch({
        "type":      "file_download",
        "filename":  "update.exe",
        "url":       "http://example.com/update.exe",
        "referrer":  "http://example.com/",
        "timestamp": time.time(),
    })
    manager.inject_all()

    zone_file = (
        Path(tmp_artifacts) / "browser" / "downloads" / "update.exe.Zone.Identifier"
    )
    assert zone_file.exists(), "Zone.Identifier file must exist for the download"
    content = zone_file.read_text()
    assert "ZoneId=3"   in content, "Zone must be Internet zone (ZoneId=3)"
    assert "HostUrl="   in content, "HostUrl field must be present"
    assert "ReferrerUrl=" in content, "ReferrerUrl field must be present"


def test_zone_identifier_fields_populated(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Zone.Identifier HostUrl and ReferrerUrl must reflect the event values."""
    url = "http://malicious-looking.example.com/payload.zip"
    ref = "http://malicious-looking.example.com/"

    manager = _make_manager(tmp_artifacts, normal_profile)
    manager.dispatch({
        "type": "file_download", "filename": "payload.zip",
        "url": url, "referrer": ref, "timestamp": time.time(),
    })
    manager.inject_all()

    content = (
        Path(tmp_artifacts) / "browser" / "downloads" / "payload.zip.Zone.Identifier"
    ).read_text()
    assert url in content
    assert ref in content


# ---------------------------------------------------------------------------
# Chrome epoch sanity test
# ---------------------------------------------------------------------------

def test_chrome_timestamps_in_correct_epoch(
    tmp_artifacts: str, normal_profile: UserProfile
) -> None:
    """Visit timestamps must use the Chrome epoch (not the Unix epoch).

    Chrome timestamps are microseconds since 1601-01-01.
    The minimum plausible value for a 2020+ timestamp is:
        (2020 - 1601) * 365.25 * 86400 * 1e6 ≈ 13_200_000_000_000_000
    """
    manager = _make_manager(tmp_artifacts, normal_profile)
    _dispatch_sample_events(manager)
    manager.inject_all()

    conn = sqlite3.connect(str(Path(tmp_artifacts) / "browser" / "History"))
    try:
        rows = conn.execute("SELECT visit_time FROM visits LIMIT 10").fetchall()
        for (ts,) in rows:
            assert ts > 13_200_000_000_000_000, (
                f"visit_time {ts} looks like a Unix timestamp — "
                "Chrome epoch offset was not applied"
            )
    finally:
        conn.close()


# ---------------------------------------------------------------------------
# Linux artifact checks
# ---------------------------------------------------------------------------

def test_linux_artifacts_created(tmp_artifacts: str, normal_profile: UserProfile) -> None:
    """Linux artifact injection must create a plausible host footprint."""
    manager = _make_manager(tmp_artifacts, normal_profile)
    manager.inject_all()

    linux_root = Path(tmp_artifacts) / "linux"
    assert (linux_root / "etc" / "hostname").exists()
    assert (linux_root / "etc" / "machine-id").exists()
    assert (linux_root / "etc" / "os-release").exists()
    assert (linux_root / "sys" / "class" / "dmi" / "id" / "product_name").exists()
    assert (linux_root / "home").exists()


def test_linux_artifacts_are_fresh_across_runs() -> None:
    """Machine IDs should vary across independent runs."""
    profile = UserProfile.default_normal()

    def run(root: str) -> str:
        manager = ArtifactManager(artifacts_path=root, profile=profile)
        manager.register(LinuxArtifactExtension(profile))
        manager.inject_all()
        return (Path(root) / "linux" / "etc" / "machine-id").read_text()

    with tempfile.TemporaryDirectory() as d1, tempfile.TemporaryDirectory() as d2:
        a1 = run(os.path.join(d1, "art"))
        a2 = run(os.path.join(d2, "art"))
        assert a1 != a2, "Linux machine-id should differ across runs"
