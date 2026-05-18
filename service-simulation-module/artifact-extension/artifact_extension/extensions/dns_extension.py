"""DNS artifact extension.

Generates a realistic DNS resolver cache text file that mirrors the
output of ``ipconfig /displaydns`` (Windows), seeded with:

* Background entries drawn from a public-domain list of common sites
  (count and TTL spread controlled by the :class:`~artifact_extension.profile.DNSProfile`).
* NXDOMAIN (negative-cache) entries at a configurable ratio.
* Real query events captured via :meth:`on_service_event`.

Artifact written to::

    <artifacts_path>/dns/cache.txt
"""

from __future__ import annotations

import random
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

from artifact_extension.base import ArtifactExtension
from artifact_extension.profile import UserProfile

# ---------------------------------------------------------------------------
# Static background data — all public, benign, real-world hostnames
# ---------------------------------------------------------------------------

_COMMON_DOMAINS: list[tuple[str, str]] = [
    ("google.com",              "142.250.80.46"),
    ("www.google.com",          "142.250.80.68"),
    ("googleapis.com",          "142.250.80.74"),
    ("gstatic.com",             "142.250.80.84"),
    ("youtube.com",             "142.250.185.46"),
    ("bing.com",                "204.79.197.200"),
    ("microsoft.com",           "20.112.52.29"),
    ("github.com",              "140.82.112.3"),
    ("stackoverflow.com",       "151.101.193.69"),
    ("reddit.com",              "151.101.65.140"),
    ("cdn.jsdelivr.net",        "151.101.2.229"),
    ("fonts.googleapis.com",    "142.250.80.74"),
    ("cloudflare.com",          "104.16.132.229"),
    ("fastly.net",              "151.101.0.57"),
    ("akamaiedge.net",          "23.221.50.138"),
    ("wikipedia.org",           "208.80.154.224"),
    ("twitter.com",             "104.244.42.65"),
    ("linkedin.com",            "13.107.42.14"),
    ("amazon.com",              "176.32.103.205"),
    ("pypi.org",                "151.101.64.223"),
    ("docs.python.org",         "151.101.64.223"),
    ("npmjs.com",               "104.16.21.35"),
    ("update.googleapis.com",   "142.250.80.74"),
    ("accounts.google.com",     "142.250.80.84"),
    ("ocsp.digicert.com",       "93.184.220.29"),
    ("crl3.digicert.com",       "93.184.220.29"),
    ("ssl.gstatic.com",         "142.250.80.84"),
    ("play.google.com",         "142.250.80.68"),
    ("connectivitycheck.gstatic.com", "142.250.80.84"),
    ("www.gstatic.com",         "142.250.80.84"),
]

# Plausible synthetic NXDOMAIN entries (stale ad/tracker domains)
_NXDOMAIN_DOMAINS: list[str] = [
    "tracker.old-analytics.invalid",
    "ads.legacy-network.invalid",
    "telemetry.stale-service.invalid",
    "cdn.expired-asset.invalid",
    "pixel.retired-tracker.invalid",
]


# ---------------------------------------------------------------------------
# Internal data model
# ---------------------------------------------------------------------------

@dataclass
class _DNSEntry:
    domain: str
    ip: Optional[str]   # ``None`` → NXDOMAIN entry
    ttl: int            # remaining TTL in seconds (decayed, not original)
    record_type: str = "A"


# ---------------------------------------------------------------------------
# Extension
# ---------------------------------------------------------------------------

class DNSArtifactExtension(ArtifactExtension):
    """Generates a DNS resolver cache file artifact.

    Args:
        profile: Statistical user profile that controls entry count and
                 TTL distribution.
    """

    def __init__(self, profile: UserProfile) -> None:
        self._profile = profile
        # Entries captured from real service events
        self._event_entries: list[_DNSEntry] = []

    # ------------------------------------------------------------------
    # ArtifactExtension interface
    # ------------------------------------------------------------------

    def on_service_event(self, event: dict) -> None:
        """Capture ``dns_query`` events for later injection.

        All other event types are silently ignored.
        """
        if event.get("type") != "dns_query":
            return

        domain: str = event.get("domain", "")
        if not domain:
            return

        ip: Optional[str] = event.get("resolved_ip")
        # Assign a plausibly decayed TTL — not a fresh max value
        ttl = random.randint(30, min(3600, self._profile.dns.ttl_jitter_seconds))
        self._event_entries.append(_DNSEntry(domain=domain, ip=ip, ttl=ttl))

    def inject(self, artifacts_path: str) -> None:
        """Write the DNS cache file to ``<artifacts_path>/dns/cache.txt``.

        Args:
            artifacts_path: Root path of the Docker volume.
        """
        dns_dir = Path(artifacts_path) / "dns"
        dns_dir.mkdir(parents=True, exist_ok=True)

        entries = self._build_entries()
        cache_text = self._render_cache(entries)
        (dns_dir / "cache.txt").write_text(cache_text, encoding="utf-8")

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _build_entries(self) -> list[_DNSEntry]:
        """Combine statistical background entries with event-driven entries."""
        p = self._profile.dns
        entries: list[_DNSEntry] = []

        # 1. Background positive entries from the static pool
        # Use random.choices (with replacement) so we can fill up to p.entries
        # even when positive_count exceeds the pool size.
        positive_count = max(0, p.entries - len(self._event_entries))
        pool = list(_COMMON_DOMAINS)
        sampled = random.choices(pool, k=positive_count)
        for domain, ip in sampled:
            # Spread TTL values so entries appear inserted at different times.
            # Using a uniform distribution gives a visible "sawtooth" pattern
            # that is harder to fingerprint than all-max or all-min TTLs.
            ttl = random.randint(1, p.ttl_jitter_seconds)
            entries.append(_DNSEntry(domain=domain, ip=ip, ttl=ttl))

        # 2. NXDOMAIN (negative cache) entries
        neg_count = max(1, int(p.entries * p.negative_ratio))
        cycle = _NXDOMAIN_DOMAINS * (neg_count // len(_NXDOMAIN_DOMAINS) + 1)
        for domain in cycle[:neg_count]:
            entries.append(
                _DNSEntry(domain=domain, ip=None, ttl=random.randint(60, 900))
            )

        # 3. Event-driven entries last — highest realism, overrides nothing
        entries.extend(self._event_entries)

        return entries

    @staticmethod
    def _render_cache(entries: list[_DNSEntry]) -> str:
        """Render entries in ``ipconfig /displaydns``-compatible text format.

        The format is intentionally verbose and close to the real Windows
        output so that string-based parsers used by fingerprinting code
        find what they expect.
        """
        lines: list[str] = [
            "Windows IP Configuration",
            "",
            "    DNS Resolver Cache",
            "    -------------------",
            "",
        ]
        for e in entries:
            lines += [
                f"    Record Name . . . . . : {e.domain}",
                f"    Record Type . . . . . : 1",
                f"    Time To Live  . . . . : {e.ttl}",
                f"    Data Length . . . . . : 4",
                f"    Section . . . . . . . : Answer",
            ]
            if e.ip:
                lines.append(f"    A (Host) Record . . . : {e.ip}")
            else:
                lines.append(f"    (NXDOMAIN / No records found)")
            lines.append("")

        return "\n".join(lines)
