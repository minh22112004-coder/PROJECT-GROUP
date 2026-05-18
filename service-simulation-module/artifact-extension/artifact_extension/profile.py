"""Statistical user profile used to control artifact generation.

Profiles contain ONLY statistical properties (counts, distributions,
ratios) — never personal data or raw user activity.  They are used
to make generated artifacts plausible in size and timestamp spread.
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Literal


@dataclass
class DNSProfile:
    """Statistical DNS usage profile."""

    # Total number of cached entries to generate
    entries: int = 150
    # Fraction of entries that should be NXDOMAIN (negative cache)
    negative_ratio: float = 0.05
    # Maximum remaining-TTL spread across synthetic entries (seconds).
    # A large value means entries appear to have been inserted at different
    # times — key for avoiding the "all injected at once" fingerprint.
    ttl_jitter_seconds: int = 3600


@dataclass
class BrowserProfile:
    """Statistical browser usage profile."""

    # Number of distinct URL rows in the history database
    history_entries: int = 400
    # Rough cache size to simulate (informational; actual DB may be smaller)
    cache_size_mb: int = 200
    # How many days of history to spread timestamps across
    days_of_history: int = 14
    # Probability weights for Chrome visit transition types.
    # Must sum to 1.0 (or close enough for random.choices).
    transition_weights: dict[str, float] = field(
        default_factory=lambda: {
            "LINK": 0.45,
            "TYPED": 0.20,
            "AUTO_BOOKMARK": 0.05,
            "AUTO_SUBFRAME": 0.15,
            "FORM_SUBMIT": 0.10,
            "RELOAD": 0.05,
        }
    )


@dataclass
class UserProfile:
    """Top-level profile controlling artifact quantity and realism.

    Instantiate directly, or use :meth:`from_json` to load from a file.
    Example JSON::

        {
            "user_type": "normal",
            "dns":     { "entries": 150 },
            "browser": { "history_entries": 400, "days_of_history": 14 }
        }
    """

    user_type: Literal["normal", "power", "minimal"] = "normal"
    dns: DNSProfile = field(default_factory=DNSProfile)
    browser: BrowserProfile = field(default_factory=BrowserProfile)

    # ------------------------------------------------------------------
    # Factories
    # ------------------------------------------------------------------

    @classmethod
    def from_json(cls, path: str | Path) -> "UserProfile":
        """Load a profile from a JSON file.

        Args:
            path: Path to a JSON file.  Unknown keys are silently ignored
                  so that partial profiles work without errors.
        """
        data = json.loads(Path(path).read_text(encoding="utf-8"))
        dns = DNSProfile(**{
            k: v for k, v in data.get("dns", {}).items()
            if k in DNSProfile.__dataclass_fields__
        })
        browser = BrowserProfile(**{
            k: v for k, v in data.get("browser", {}).items()
            if k in BrowserProfile.__dataclass_fields__
        })
        return cls(
            user_type=data.get("user_type", "normal"),
            dns=dns,
            browser=browser,
        )

    @classmethod
    def default_normal(cls) -> "UserProfile":
        """Return a ready-to-use profile for a typical desktop user."""
        return cls(user_type="normal")

    @classmethod
    def default_minimal(cls) -> "UserProfile":
        """Return a lean profile suitable for quick test runs."""
        return cls(
            user_type="minimal",
            dns=DNSProfile(entries=20, ttl_jitter_seconds=600),
            browser=BrowserProfile(history_entries=30, days_of_history=3),
        )
