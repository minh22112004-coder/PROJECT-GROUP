# Pack-A-Mal

Package Analysis for Malware - A comprehensive tool for analyzing packages from various open source ecosystems to detect potentially malicious behavior.

## Network Simulation Integration (NEW!)

Pack-A-Mal now features **INetSim integration** for safe analysis of packages with network-based malicious behavior!

### Key Features
✅ **Automatic URL liveness detection** - Checks if URLs are active before analysis  
✅ **Intelligent traffic redirection** - Routes dead URLs to INetSim simulation  
✅ **Safe C2 analysis** - Analyze malware C2 communication without real connections  
✅ **Complete network logging** - Capture all simulated network interactions  


## Project Overview: a dynamic malware analysis framwork for open-source packages


## Overview
This is comprehensive source code and deploy in production for project packamal. This project include 2 main components:

### 1. Dynamic Analysis Module

**Overview**: A comprehensive tool for analyzing packages from various open source ecosystems using dynamic analysis techniques. This module performs runtime analysis of packages in isolated sandbox environments to detect potentially malicious behavior, track system calls, monitor network activity, and analyze file operations.

**Key Features**:
- **Runtime Behavior Monitoring**: Analyzes packages during installation, import, and execution phases in isolated sandbox environments
- **System Call Tracking**: Monitors file operations, command executions, and system interactions via `strace`
- **Network Activity Analysis**: Captures DNS queries and network connections through packet capture
- **File Operations Monitoring**: Tracks reads, writes, and deletions during package execution
- **Execution Logging**: Records module execution and symbol tracking

**Supported Ecosystems**:
- PyPI (Python packages)
- npm (Node.js packages)
- RubyGems (Ruby packages)
- Packagist (PHP Composer packages)
- Crates.io (Rust packages)
- Maven (Java packages)
- Wolfi (Wolfi Linux packages)

**Architecture Components**:
- Analysis Image: Main analysis runner that orchestrates static and dynamic analysis
- Sandboxes: Containerized environments for safe package execution (dynamic and static analysis sandboxes)
- Scheduler: Kubernetes service that schedules analysis jobs from package feeds
- Worker: Processes analysis jobs from the queue

### 2. Web APIs

**Overview**: A Django-based web application that provides both a web interface and REST API for analyzing packages from multiple ecosystems. The platform performs dynamic and static analysis to detect security vulnerabilities, typosquatting attempts, and malicious behavior in software packages.

**Key Features**:

**Web Dashboard**:
- Package Discovery: Browse and search packages across multiple ecosystems
- Interactive Analysis: Submit packages for analysis through an intuitive web interface
- Real-time Status: Monitor analysis progress and queue position
- Report Visualization: View detailed analysis reports with security findings

**Analysis Capabilities**:
1. **Dynamic Analysis**: Runtime behavior monitoring, network activity tracking, file system operations, command execution logging, and domain/IP address tracking
2. **Static Analysis**: 
   - Bandit4Mal: Security vulnerability scanning
   - Malcontent: Malicious content detection
   - LastPyMile: Source package discrepancy identification
3. **Typosquatting Detection**: Identifies packages with similar names to popular packages
4. **Source Code Finder**: Locates source code repositories for packages (supports PyPI and npm ecosystems)

**REST APIs**:
- Full programmatic access to all analysis features
- API key authentication with rate limiting
- Queue management and task status tracking
- Package URL (PURL) format support for package identification

**Technology Stack**:
- Django 5.1.6 (Python web framework)
- PostgreSQL (database backend)
- Gunicorn (WSGI server)
- Docker (container-based analysis execution)
- Redis (optional caching and task queue)

**Supported Ecosystems**: PyPI, npm, Packagist, RubyGems, Maven, Rust, and Wolfi

## Contact: support@packguard.dev

## Papaer
- [Pack-A-Mal: A Malware Analysis Framework for Open-Source Packages](https://arxiv.org/pdf/2511.09957)
