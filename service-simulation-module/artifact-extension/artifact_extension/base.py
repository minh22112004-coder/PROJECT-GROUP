"""Base class for all artifact simulation extensions."""

from abc import ABC, abstractmethod


class ArtifactExtension(ABC):
    """Abstract base for all artifact simulation extensions.

    Each extension handles one service type (DNS, HTTP, etc.) and is
    responsible for both receiving service events and generating the
    corresponding persistent system artifacts.

    Plugin contract:
    - ``on_service_event`` is called for every event the ArtifactManager
      dispatches; extensions must silently ignore events they do not own.
    - ``inject`` is called once, before the malware/sample starts, and
      must write all accumulated artifacts to the given path.
    """

    @abstractmethod
    def on_service_event(self, event: dict) -> None:
        """Record a service-level event for later artifact generation.

        Args:
            event: Dictionary with at minimum:
                   - ``"type"`` (str): e.g. ``"dns_query"``, ``"http_request"``
                   - ``"timestamp"`` (float): Unix epoch seconds
                   Additional keys depend on the event type.
        """
        ...

    @abstractmethod
    def inject(self, artifacts_path: str) -> None:
        """Generate and write persistent artifacts to the target path.

        Args:
            artifacts_path: Absolute path to the Docker volume mount root
                            (e.g. ``"/artifacts"``).  Each extension is
                            responsible for creating its own subdirectory.
        """
        ...
