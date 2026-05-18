# UBER-like Artifact Simulation Extension
# ----------------------------------------
# Extends the service-simulation container with persistent system artifacts
# that are injected into the /artifacts volume before sample execution.

FROM python:3.11-slim

LABEL description="UBER-like artifact simulation extension for malware analysis sandbox"
LABEL usage="academic/defensive research only"

WORKDIR /app

# Install dependencies first (separate layer for better cache reuse)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the extension package and entrypoint
COPY artifact_extension/ ./artifact_extension/
COPY entrypoint.py .

# The volume where artifacts are written.
# Mount this volume into the analysis container so the sample sees
# the artifacts at a path you control (e.g. /artifacts).
VOLUME ["/artifacts"]

# Optional: supply a JSON profile via environment variable or --profile flag
ENV ARTIFACTS_PATH=/artifacts
ENV PROFILE_PATH=""

# The entrypoint injects artifacts and then exits (exit code 0).
# Compose with CMD in the analysis container to chain execution.
ENTRYPOINT ["python", "entrypoint.py"]
