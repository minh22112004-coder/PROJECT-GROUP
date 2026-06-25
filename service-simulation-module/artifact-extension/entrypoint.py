"""Docker entrypoint — inject UBER-like artifacts before sample execution.

This script is designed to run as the container ENTRYPOINT.  It:

1. Loads a statistical user profile (from ``--profile`` or uses the default).
2. Registers DNS, HTTP, and Linux artifact extensions with the ArtifactManager.
3. Emits a set of plausible warm-up service events (background traffic).
4. Calls ``inject_all()`` to write all artifacts to the Docker volume.
5. Exits with code 0 so the container CMD (the actual sample runner) starts.

Docker usage example::

    ENTRYPOINT ["python", "entrypoint.py", "--artifacts", "/artifacts"]
    CMD ["python", "run_sample.py"]

Standalone usage::

    python entrypoint.py --artifacts /artifacts --profile /config/profile.json
"""

from __future__ import annotations

import argparse
import logging
import os
import sys
import time

from artifact_extension.extensions.dns_extension import DNSArtifactExtension
from artifact_extension.extensions.http_extension import HTTPArtifactExtension
from artifact_extension.extensions.linux_extension import LinuxArtifactExtension
from artifact_extension.manager import ArtifactManager
from artifact_extension.profile import UserProfile

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s — %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger("entrypoint")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def _parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Inject UBER-like persistent artifacts before sandbox execution."
    )
    parser.add_argument(
        "--artifacts",
        default=os.environ.get("ARTIFACTS_PATH", "/artifacts"),
        help="Docker volume mount path for artifacts (default: /artifacts)",
    )
    parser.add_argument(
        "--profile",
        default=os.environ.get("PROFILE_PATH", None),
        help="Path to a JSON user profile file (optional)",
    )
    return parser.parse_args()


# ---------------------------------------------------------------------------
# Background event emission
# ---------------------------------------------------------------------------

def _emit_background_events(manager: ArtifactManager) -> None:
    """Emit plausible background service events to seed the extensions.

    These events represent activity that would have occurred on a real
    user machine *before* the malware/sample runs — not during execution.
    Timestamps are set in the past so artifacts appear aged.
    """
    now = time.time()
    hour = 3600.0

    # Background DNS queries — common sites a real user machine resolves
    background_dns = [
        ("google.com",        "142.250.80.46"),
        ("github.com",        "140.82.112.3"),
        ("stackoverflow.com", "151.101.193.69"),
        ("pypi.org",          "151.101.64.223"),
        ("docs.python.org",   "151.101.64.223"),
        ("bing.com",          "204.79.197.200"),
        ("microsoft.com",     "20.112.52.29"),
        ("youtube.com",       "142.250.185.46"),
    ]
    for domain, ip in background_dns:
        manager.dispatch({
            "type":        "dns_query",
            "domain":      domain,
            "resolved_ip": ip,
            "timestamp":   now - (4 * hour),   # queries happened ~4 h ago
        })

    # Background HTTP history — pages browsed before malware execution
    background_http = [
        ("https://www.google.com/",            "Google",           "TYPED"),
        ("https://github.com/",                "GitHub",           "TYPED"),
        ("https://pypi.org/simple/requests/",  "requests · PyPI",  "LINK"),
        ("https://stackoverflow.com/",         "Stack Overflow",   "TYPED"),
        ("https://news.ycombinator.com/",      "Hacker News",      "TYPED"),
    ]
    for url, title, transition in background_http:
        manager.dispatch({
            "type":       "http_request",
            "url":        url,
            "title":      title,
            "method":     "GET",
            "transition": transition,
            "timestamp":  now - (2 * hour),    # browsed ~2 h ago
        })

    # Background file downloads — files a real user would have downloaded
    background_downloads = [
        ("https://github.com/releases/download/v2.1/setup.exe",
         "github.com", "setup.exe"),
        ("https://dl.google.com/chrome/ChromeSetup.exe",
         "dl.google.com", "ChromeSetup.exe"),
        ("https://cdn.discordapp.com/apps/DiscordSetup.exe",
         "cdn.discordapp.com", "DiscordSetup.exe"),
        ("https://code.visualstudio.com/sha/download/VSCodeSetup-x64.exe",
         "code.visualstudio.com", "VSCodeSetup-x64.exe"),
        ("https://files.pythonhosted.org/packages/requests-2.31.0.tar.gz",
         "files.pythonhosted.org", "requests-2.31.0.tar.gz"),
    ]
    for download_url, referrer, filename in background_downloads:
        manager.dispatch({
            "type":         "file_download",
            "url":          download_url,
            "referrer_url": f"https://{referrer}/",
            "filename":     filename,
            "timestamp":    now - (3 * hour),  # downloaded ~3 h ago
        })


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    args = _parse_args()

    # Load or create user profile
    if args.profile and os.path.exists(args.profile):
        profile = UserProfile.from_json(args.profile)
        logger.info(
            "Loaded user profile from %s  (type=%s)", args.profile, profile.user_type
        )
    else:
        profile = UserProfile.default_normal()
        logger.info("Using default 'normal' user profile")

    # Build manager and plug in extensions
    manager = ArtifactManager(artifacts_path=args.artifacts, profile=profile)
    manager.register(DNSArtifactExtension(profile))
    manager.register(HTTPArtifactExtension(profile))
    manager.register(LinuxArtifactExtension(profile))

    # Seed with background events
    logger.info("Emitting background service events ...")
    _emit_background_events(manager)

    # Inject all artifacts into the Docker volume
    logger.info("Injecting artifacts into %s ...", args.artifacts)
    manager.inject_all()
    logger.info("Artifact injection complete.  Container is ready.")


if __name__ == "__main__":
    main()
