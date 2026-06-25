"""Linux artifact extension.

Generates plausible Linux host artifacts that help a sample see a more
realistic desktop environment before execution. The artifacts are
inspired by the kinds of system fingerprints and filesystem checks used
by sandbox-evasion tools such as fake-sandbox, pafish, and al-khaser.

Artifacts are written under::

    <artifacts_path>/linux/

The extension does not try to emulate a full Linux installation. It
creates a focused set of host identity, hardware, and user-environment
files that are commonly checked by malware or analysis tooling.
"""

from __future__ import annotations

import random
import secrets
import time
from dataclasses import dataclass
from pathlib import Path

from artifact_extension.base import ArtifactExtension
from artifact_extension.profile import UserProfile


@dataclass(frozen=True)
class _LinuxIdentity:
    hostname: str
    distro_id: str
    version_id: str
    pretty_name: str
    vendor: str
    product_name: str
    bios_vendor: str
    bios_version: str
    username: str


_IDENTITY_POOL: list[_LinuxIdentity] = [
    _LinuxIdentity(
        hostname="workstation-07",
        distro_id="ubuntu",
        version_id="22.04",
        pretty_name="Ubuntu 22.04.4 LTS",
        vendor="Dell Inc.",
        product_name="OptiPlex 7090",
        bios_vendor="Dell Inc.",
        bios_version="1.25.0",
        username="alex",
    ),
    _LinuxIdentity(
        hostname="devbox-12",
        distro_id="debian",
        version_id="12",
        pretty_name="Debian GNU/Linux 12 (bookworm)",
        vendor="Lenovo",
        product_name="ThinkPad T14 Gen 3",
        bios_vendor="Lenovo",
        bios_version="N3BET42W (1.18)",
        username="sam",
    ),
    _LinuxIdentity(
        hostname="studio-04",
        distro_id="fedora",
        version_id="40",
        pretty_name="Fedora Linux 40 (Workstation Edition)",
        vendor="HP",
        product_name="EliteBook 840 G9",
        bios_vendor="HP",
        bios_version="U91 Ver. 01.12.01",
        username="maria",
    ),
]


class LinuxArtifactExtension(ArtifactExtension):
    """Generates a small set of realistic Linux system artifacts."""

    def __init__(self, profile: UserProfile) -> None:
        self._profile = profile

    def on_service_event(self, event: dict) -> None:
        """This extension is injection-only and ignores service events."""
        return

    def inject(self, artifacts_path: str) -> None:
        linux_dir = Path(artifacts_path) / "linux"
        linux_dir.mkdir(parents=True, exist_ok=True)

        identity = random.choice(_IDENTITY_POOL)
        machine_id = secrets.token_hex(16)

        self._write_text(linux_dir / "etc" / "hostname", f"{identity.hostname}\n")
        self._write_text(linux_dir / "etc" / "machine-id", f"{machine_id}\n")
        self._write_text(linux_dir / "var" / "lib" / "dbus" / "machine-id", f"{machine_id}\n")
        self._write_text(linux_dir / "etc" / "os-release", self._render_os_release(identity))
        self._write_text(linux_dir / "etc" / "issue", f"{identity.pretty_name}\n")
        self._write_text(linux_dir / "etc" / "hosts", self._render_hosts(identity))
        self._write_text(linux_dir / "sys" / "class" / "dmi" / "id" / "sys_vendor", f"{identity.vendor}\n")
        self._write_text(linux_dir / "sys" / "class" / "dmi" / "id" / "product_name", f"{identity.product_name}\n")
        self._write_text(linux_dir / "sys" / "class" / "dmi" / "id" / "bios_vendor", f"{identity.bios_vendor}\n")
        self._write_text(linux_dir / "sys" / "class" / "dmi" / "id" / "bios_version", f"{identity.bios_version}\n")
        self._write_text(linux_dir / "proc" / "cpuinfo", self._render_cpuinfo())

        home_dir = linux_dir / "home" / identity.username
        self._write_text(home_dir / ".bash_history", self._render_bash_history(identity))
        self._write_text(home_dir / ".config" / "gtk-3.0" / "settings.ini", self._render_gtk_settings())
        self._write_text(home_dir / ".local" / "share" / "recently-used.xbel", self._render_recent_files())
        self._write_text(linux_dir / "var" / "log" / "syslog", self._render_syslog(identity))

    @staticmethod
    def _write_text(path: Path, content: str) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

    @staticmethod
    def _render_os_release(identity: _LinuxIdentity) -> str:
        return (
            f'NAME="{identity.pretty_name.split()[0]}"\n'
            f'PRETTY_NAME="{identity.pretty_name}"\n'
            f'ID={identity.distro_id}\n'
            'ID_LIKE="debian"\n'
            f'VERSION_ID="{identity.version_id}"\n'
            'HOME_URL="https://www.linux.org/"\n'
        )

    @staticmethod
    def _render_hosts(identity: _LinuxIdentity) -> str:
        return (
            "127.0.0.1\tlocalhost\n"
            f"127.0.1.1\t{identity.hostname}\n"
            "::1\tlocalhost ip6-localhost ip6-loopback\n"
        )

    @staticmethod
    def _render_cpuinfo() -> str:
        cores = random.choice([4, 6, 8])
        siblings = cores * 2
        cpu_mhz = random.choice([2394.0, 2593.0, 2793.0, 3192.0])
        return (
            "processor\t: 0\n"
            "vendor_id\t: GenuineIntel\n"
            "cpu family\t: 6\n"
            "model\t\t: 142\n"
            "model name\t: Intel(R) Core(TM) i7-1185G7 CPU @ 3.00GHz\n"
            f"cpu MHz\t\t: {cpu_mhz:.1f}\n"
            "cache size\t: 12288 KB\n"
            f"siblings\t: {siblings}\n"
            f"cpu cores\t: {cores}\n"
            "flags\t\t: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush mmx sse sse2 ss ht syscall nx lm constant_tsc\n"
        )

    @staticmethod
    def _render_bash_history(identity: _LinuxIdentity) -> str:
        commands = [
            "ls -la",
            "pwd",
            "cat /etc/os-release",
            "python3 --version",
            "git status",
            "docker ps",
            "journalctl -n 20",
            f"whoami # {identity.username}",
        ]
        return "\n".join(commands) + "\n"

    @staticmethod
    def _render_gtk_settings() -> str:
        theme = random.choice(["Yaru", "Adwaita", "Breeze"])
        return (
            "[Settings]\n"
            f"gtk-theme-name={theme}\n"
            "gtk-icon-theme-name=Adwaita\n"
            "gtk-font-name=Sans 11\n"
        )

    @staticmethod
    def _render_recent_files() -> str:
        now = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        return (
            '<?xml version="1.0" encoding="UTF-8"?>\n'
            '<xbel version="1.0">\n'
            f'  <bookmark href="file:///home/user/Documents/report.pdf" added="{now}">\n'
            '    <title>report.pdf</title>\n'
            '  </bookmark>\n'
            '</xbel>\n'
        )

    @staticmethod
    def _render_syslog(identity: _LinuxIdentity) -> str:
        timestamp = time.strftime("%b %d %H:%M:%S", time.gmtime())
        return (
            f"{timestamp} {identity.hostname} systemd[1]: Started User Manager for UID 1000.\n"
            f"{timestamp} {identity.hostname} NetworkManager[512]: state change: connected\n"
            f"{timestamp} {identity.hostname} sudo:     {identity.username} : TTY=pts/0 ; PWD=/home/{identity.username} ; USER=root ; COMMAND=/usr/bin/apt update\n"
        )
