"""Artifact extensions sub-package."""

from artifact_extension.extensions.dns_extension import DNSArtifactExtension
from artifact_extension.extensions.http_extension import HTTPArtifactExtension

__all__ = ["DNSArtifactExtension", "HTTPArtifactExtension"]
