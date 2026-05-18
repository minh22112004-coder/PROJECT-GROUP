# PROMPT FOR AI CODE

## Python + Docker – UBER-like Artifact Simulation Extension

---

## ROLE & CONTEXT

You are a **senior security engineer and Python developer** working on **defensive malware analysis sandbox research**.

The sandbox environment:
- Runs inside **Docker containers**
- Already includes a **service simulation module** (DNS, HTTP, etc.) implemented in Python
- Uses **Docker volumes** for persistence

Your task is to **extend** the existing service simulation module with a **UBER-like artifact simulation layer**, inspired by the paper:

> *Enhancing malware analysis sandboxes with emulated user behavior*

### Constraints
- Defensive & academic use only
- Do NOT create real malware
- Do NOT simulate direct user actions (mouse/keyboard)
- Focus on **persistent system artifacts**
- Use **Python 3.10+**

---

## PROBLEM STATEMENT

Current limitation of the sandbox:
- Services (DNS, HTTP) respond correctly
- However, the system lacks **persistent system artifacts**
- Malware can detect the sandbox via **fingerprinting**, e.g.:
  - Missing DNS cache
  - No browser history or cache
  - No realistic file or log traces

---

## OBJECTIVE

Design and implement a **UBER-like Artifact Simulation Extension** that:

1. Does **NOT modify** the existing service simulation logic
2. Listens to **service-level events**
3. Generates **persistent system artifacts**
4. Injects artifacts **before malware/sample execution**
5. Uses **Docker volumes** to emulate an “always-on” user system

---

## REQUIRED ARCHITECTURE

### High-level flow

```
Service Simulation Module (existing)
            ↓
     Artifact Extension Layer (new)
            ↓
   Persistent Artifacts (Docker volume)
```

### Event-driven model

```
Service Event → Artifact Rule → Artifact Generation → Volume Injection
```

### Examples
- DNS query → DNS cache artifact
- HTTP request → browser history + cache artifact
- File download → file + realistic timestamps

---

## ARTIFACT PROFILE (UBER-STYLE)

Artifacts must be generated from a **statistical user profile**, NOT real user data.

Example profile:

```json
{
  "user_type": "normal",
  "dns": {
    "entries": 150
  },
  "browser": {
    "history_entries": 400,
    "cache_size_mb": 200
  },
  "filesystem": {
    "user_files": 1000,
    "recent_files": 30
  }
}
```

Rules:
- Only statistical properties (counts, distributions)
- No personal or raw user data
- Used to control artifact quantity and timestamp realism

---

## EXTENSION API REQUIREMENTS

Design a **plugin-based Python API** for artifact simulation.

Example interface:

```python
from abc import ABC, abstractmethod

class ArtifactExtension(ABC):
    @abstractmethod
    def on_service_event(self, event: dict) -> None:
        """Receive service events and record artifact intent"""
        pass

    @abstractmethod
    def inject(self, artifacts_path: str) -> None:
        """Generate and inject persistent artifacts before execution"""
        pass
```

Guidelines:
- One extension per service (e.g. DNSArtifactExtension, HTTPArtifactExtension)
- Extensions must be independent and easily pluggable

---

## DOCKER-SPECIFIC REQUIREMENTS

- Use a **mounted Docker volume** (e.g. `/artifacts`) to store all artifacts
- Artifacts must:
  - Persist across container restarts
  - Appear older than the container start time
- Artifact generation must occur:
  - In an **entrypoint script**, or
  - Before malware/sample execution begins
- Do NOT spawn suspicious runtime processes while the sample is running

---

## TESTING & EVALUATION (MANDATORY)

Design tests that demonstrate **behavioral impact**, not just correctness.

### Required test cases

| Case | Setup | Expected Result |
|----|----|----|
| A | Clean sandbox | PoC detects sandbox |
| B | Service simulation only | Still detected |
| C | Service + artifact extension | Appears as real system |

Testing rules:
- Use **safe Python PoC checks** instead of real malware
- Avoid any real malicious payloads
- Include freshness tests (artifacts differ across runs)

---

## REQUIRED DELIVERABLES

Generate the following:

1. Python module/repository structure
2. `ArtifactExtension` base class
3. `ArtifactManager` (central controller)
4. One concrete extension (DNS or HTTP)
5. Docker entrypoint logic
6. Explanation of artifact realism decisions
7. Testing and evaluation strategy

Code requirements:
- Pythonic and readable
- Well-commented
- Easy to extend
- Suitable for academic evaluation

---

## FINAL ACADEMIC SUMMARY

Also provide a short academic-style paragraph explaining:

- Why service simulation alone is insufficient
- How persistent artifacts improve sandbox stealth
- How this approach is inspired by the UBER system

---

**END OF PROMPT**

