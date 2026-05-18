"""ArtifactManager — central controller for all artifact extensions.

Responsibilities
----------------
* Route service events to every registered :class:`ArtifactExtension`.
* Coordinate artifact injection in a deterministic order.
* Guarantee the Docker volume directory exists before injection.
* Isolate extension failures so one bad extension cannot block others.
"""

from __future__ import annotations

import logging
from pathlib import Path

from artifact_extension.base import ArtifactExtension
from artifact_extension.profile import UserProfile

logger = logging.getLogger(__name__)


class ArtifactManager:
    """Orchestrates all registered :class:`ArtifactExtension` instances.

    Usage::

        profile = UserProfile.default_normal()
        manager = ArtifactManager(artifacts_path="/artifacts", profile=profile)
        manager.register(DNSArtifactExtension(profile))
        manager.register(HTTPArtifactExtension(profile))

        # Feed events from the service simulation layer
        manager.dispatch({"type": "dns_query", "domain": "example.com", ...})

        # Before sample execution: write everything to the volume
        manager.inject_all()
    """

    def __init__(
        self,
        artifacts_path: str,
        profile: UserProfile | None = None,
    ) -> None:
        """
        Args:
            artifacts_path: Root directory of the Docker volume mount
                            (e.g. ``"/artifacts"``).
            profile:        Statistical user profile.  Defaults to
                            :meth:`UserProfile.default_normal` when omitted.
        """
        self._artifacts_path = Path(artifacts_path)
        self._profile: UserProfile = profile or UserProfile.default_normal()
        self._extensions: list[ArtifactExtension] = []

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def register(self, extension: ArtifactExtension) -> None:
        """Add an extension to the manager.

        Extensions are called in registration order during both
        :meth:`dispatch` and :meth:`inject_all`.

        Args:
            extension: A concrete :class:`ArtifactExtension` instance.
        """
        self._extensions.append(extension)
        logger.info("Registered extension: %s", type(extension).__name__)

    def dispatch(self, event: dict) -> None:
        """Broadcast a service-level event to every registered extension.

        Each extension silently ignores event types it does not own.
        Exceptions raised by individual extensions are caught, logged,
        and do not propagate — so one broken extension cannot stall others.

        Args:
            event: Dictionary with at minimum ``"type"`` and
                   ``"timestamp"`` keys.
        """
        for ext in self._extensions:
            try:
                ext.on_service_event(event)
            except Exception:
                logger.exception(
                    "Extension %s raised an exception handling event type=%r",
                    type(ext).__name__,
                    event.get("type"),
                )

    def inject_all(self) -> None:
        """Write all accumulated artifacts to the Docker volume.

        Creates the root artifact directory if it does not exist, then
        calls :meth:`~ArtifactExtension.inject` on each registered
        extension.  Failures are isolated per extension.
        """
        self._artifacts_path.mkdir(parents=True, exist_ok=True)
        logger.info(
            "Starting artifact injection into %s (%d extension(s))",
            self._artifacts_path,
            len(self._extensions),
        )
        for ext in self._extensions:
            try:
                ext.inject(str(self._artifacts_path))
                logger.info("  [OK] %s", type(ext).__name__)
            except Exception:
                logger.exception("  [FAIL] %s", type(ext).__name__)

    # ------------------------------------------------------------------
    # Properties
    # ------------------------------------------------------------------

    @property
    def profile(self) -> UserProfile:
        """The statistical user profile used by this manager."""
        return self._profile

    @property
    def extensions(self) -> list[ArtifactExtension]:
        """Read-only view of the registered extensions list."""
        return list(self._extensions)
