# Pack-a-mal Web Interface

A comprehensive web application for analyzing packages from multiple ecosystems (PyPI, npm, Packagist, RubyGems, Maven, Rust, Wolfi) for security vulnerabilities, typosquatting, and malicious behavior.

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Installation](#installation)
4. [Quick Start](#quick-start)
5. [Usage](#usage)
6. [API Documentation](#api-documentation)
7. [Architecture](#architecture)
8. [Configuration](#configuration)
9. [Development](#development)
10. [Troubleshooting](#troubleshooting)
11. [Additional Resources](#additional-resources)

---

## Overview

Pack-a-mal is a security analysis platform that provides both a web interface and REST API for analyzing packages from various package ecosystems. The platform performs dynamic and static analysis to detect security vulnerabilities, typosquatting attempts, and malicious behavior in software packages.

### Supported Ecosystems

- **PyPI** (Python packages)
- **npm** (Node.js packages)
- **Packagist** (PHP Composer packages)
- **RubyGems** (Ruby packages)
- **Maven** (Java packages)
- **Rust** (Cargo packages)
- **Wolfi** (Linux packages)

---

## Features

### Web Dashboard

- **Package Discovery**: Browse and search packages across multiple ecosystems
- **Interactive Analysis**: Submit packages for analysis through an intuitive web interface
- **Real-time Status**: Monitor analysis progress and queue position
- **Report Visualization**: View detailed analysis reports with security findings

### Analysis Capabilities

1. **Dynamic Analysis**
   - Runtime behavior monitoring
   - Network activity tracking
   - File system operations
   - Command execution logging
   - Domain and IP address tracking

2. **Static Analysis**
   - **Bandit4Mal**: Security vulnerability scanning
   - **Malcontent**: Malicious content detection
   - **LastPyMile**: Source package discrepancy identification

3. **Typosquatting Detection**
   - Identifies packages with similar names to popular packages
   - Helps detect potential typosquatting attacks

4. **Source Code Finder**
   - Locates source code repositories for packages
   - Supports PyPI and npm ecosystems

### REST API

- Full programmatic access to all analysis features
- API key authentication
- Rate limiting
- Queue management
- Task status tracking

---

## Installation

### Requirements

- **Operating System**: Linux (Ubuntu 22.04 recommended) or WSL on Windows
- **Python**: 3.x
- **Docker**: For running analysis containers
- **PostgreSQL**: Database backend
- **Redis**: For caching and task queue (optional)
- **Nginx**: Web server (production)
- **Gunicorn**: WSGI server (production)

### System Dependencies

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y python3 python3-pip python3-venv postgresql postgresql-contrib redis-server nginx docker.io
```

### Quick Installation

For a complete production setup with security best practices, see [SETUP.md](SETUP.md).

For a quick development setup:

```bash
# Clone the repository
git clone https://github.com/pakaremon/packamal.git
cd packamal/web/packamal

# Run automated setup script
./setup.sh

# Or manual setup:
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Environment Configuration

Create a `.env` file in `web/packamal/`:

```bash
# Django Settings
SECRET_KEY=your-secret-key-here
DEBUG=False
ALLOWED_HOSTS=localhost,127.0.0.1,your-domain.com

# Database Settings
DB_NAME=packamal
DB_USER=packamal_db
DB_PASSWORD=your-strong-password
DB_HOST=localhost
DB_PORT=5432

# Security Settings
SECURE_SSL_REDIRECT=True
SESSION_COOKIE_SECURE=True
CSRF_COOKIE_SECURE=True

# Redis Settings (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# Analysis Docker Image
ANALYSIS_IMAGE=docker.io/pakaremon/analysis
```

Generate a secure Django secret key:

```bash
python -c 'from django.core.management.utils import get_random_secret_key; print(get_random_secret_key())'
```

### Database Setup

```bash
# Create database and user
sudo -u postgres psql
```

```sql
CREATE DATABASE packamal;
CREATE USER packamal_db WITH PASSWORD 'your-strong-password';
ALTER ROLE packamal_db SET client_encoding TO 'utf8';
ALTER ROLE packamal_db SET default_transaction_isolation TO 'read committed';
ALTER ROLE packamal_db SET timezone TO 'UTC';
GRANT ALL PRIVILEGES ON DATABASE packamal TO packamal_db;
\q
```

### Initialize Database

```bash
source venv/bin/activate
cd web/packamal
python manage.py migrate
python manage.py createsuperuser
python manage.py collectstatic --noinput
```

---

## Quick Start

### Development Server

```bash
source venv/bin/activate
cd web/packamal
python manage.py runserver
```

Access the web interface at `http://localhost:8000`

### Production Server

See [SETUP.md](SETUP.md) for complete production setup with Gunicorn and Nginx.

---

## Usage

### Web Dashboard

#### Homepage

![Homepage](https://bit.ly/packguard-dev)

The homepage provides an overview of available features:
- Package analysis tools
- Static analysis capabilities
- Source code finder
- Contact information

#### Dashboard


[Dashboard](https://bit.ly/packguard-dev)

The dashboard allows you to:

1. **Select Ecosystem**: Choose from PyPI, npm, Packagist, RubyGems, Maven, Rust, or Wolfi
2. **Search Packages**: Use autocomplete to find packages
3. **Select Version**: Choose a specific package version
4. **Submit Analysis**: Queue the package for analysis
5. **Monitor Progress**: View real-time analysis status and queue position

#### Reports

[Report Detail](https://bit.ly/packguard-dev)

Analysis reports include:

- **Files Analyzed**: Complete list of files processed
- **Commands Executed**: All commands run during analysis
- **Network Activity**: Domains accessed and IP addresses contacted
- **Security Findings**: Detected vulnerabilities and threats
- **Behavioral Analysis**: Runtime behavior patterns

### REST API

The web interface is backed by a comprehensive REST API. See [API_DOCUMENTATION.md](packamal/API_DOCUMENTATION.md) for complete API documentation.

#### Quick API Example

```bash
# Submit package for analysis
curl -X POST https://packguard.dev/api/v1/analyze/ \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "purl": "pkg:pypi/requests@2.28.1"
  }'

# Check task status
curl https://packguard.dev/api/v1/task/123/
```

### Package URL (PURL) Format

The API uses Package URLs (PURL) to identify packages:

- `pkg:pypi/requests@2.28.1` - Python package
- `pkg:npm/lodash@4.17.21` - Node.js package
- `pkg:gem/rails@7.0.0` - Ruby gem
- `pkg:maven/org.apache.commons:commons-lang3@3.12.0` - Maven package
- `pkg:packagist/monolog/monolog@3.0.0` - PHP Composer package

---

## API Documentation

For complete API documentation, see [API_DOCUMENTATION.md](packamal/API_DOCUMENTATION.md).

### Key Endpoints

- `POST /api/v1/analyze/` - Submit package for analysis (requires API key)
- `GET /api/v1/task/<id>/` - Get task status (public)
- `GET /api/v1/reports/` - List your analysis tasks (requires API key)
- `GET /api/v1/queue/status/` - Get queue status (public)
- `GET /get_pypi_packages/` - Discover PyPI packages
- `GET /get_npm_packages/` - Discover npm packages

### Authentication

API keys can be created through the Django admin interface. Use the key in requests:

```http
Authorization: Bearer YOUR_API_KEY
```

or

```http
X-API-Key: YOUR_API_KEY
```

---

## Architecture

For detailed architecture documentation, see:
- [ARCHITECTURE_DIAGRAM.md](packamal/ARCHITECTURE_DIAGRAM.md)
- [k8s/ARCHITECTURE.md](packamal/k8s/ARCHITECTURE.md) (Kubernetes deployment)

### System Components

1. **Web Server Layer**: Gunicorn + Django
2. **API Layer**: RESTful endpoints with authentication and rate limiting
3. **Queue Management**: Priority-based task queue system
4. **Container Management**: Docker-based analysis execution
5. **Analysis Tools**: Dynamic and static analysis engines
6. **Data Storage**: PostgreSQL database and file-based reports

### Task Lifecycle

```
pending → queued → running → completed/failed
```

Tasks are processed sequentially (one container at a time) to prevent resource conflicts.

---

## Configuration

### Django Settings

Key settings in `packamal/settings.py`:

- `DEBUG`: Set to `False` in production
- `ALLOWED_HOSTS`: List of allowed hostnames
- `SECURE_SSL_REDIRECT`: Force HTTPS in production
- `DATABASES`: PostgreSQL configuration
- `MEDIA_ROOT`: Location for analysis reports

### Docker Configuration

The analysis runs in Docker containers. Ensure Docker is installed and the `packamal` user (or your application user) has Docker access:

```bash
sudo usermod -aG docker packamal
```

### Queue Configuration

- Default timeout: 30 minutes per task
- Priority system: Higher priority tasks processed first
- Smart caching: Completed results are cached indefinitely

---

## Development

### Project Structure

```
web/
├── packamal/              # Django project
│   ├── packamal/         # Django settings and configuration
│   ├── package_analysis/  # Main application
│   │   ├── views.py      # View handlers
│   │   ├── models.py     # Database models
│   │   ├── urls.py       # URL routing
│   │   └── templates/    # HTML templates
│   ├── requirements.txt  # Python dependencies
│   └── manage.py         # Django management script
├── images/                # Screenshots and images
├── README.md             # This file
└── SETUP.md              # Production setup guide
```

### Running Tests

```bash
source venv/bin/activate
cd web/packamal
python manage.py test
```


---

## Troubleshooting

### Debug Dynamic Analysis Container

```bash
docker run --rm --privileged -it --entrypoint /bin/sh -v "$PWD":/app pakaremon/dynamic-analysis:latest
```

### Run Commands in Existing Container

```bash
docker exec -it <container_id> /bin/sh
```

### Fix Node.js Path Issues

```bash
export NODE_PATH=/usr/lib/node_modules
```

### Common Issues

#### Permission Denied Errors

```bash
# Check file ownership
ls -la /opt/packamal

# Fix ownership
sudo chown -R packamal:packamal /opt/packamal
```

#### Database Connection Issues

```bash
# Test database connection
sudo -u packamal psql -h localhost -U packamal_db -d packamal

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*-main.log
```

#### Service Not Starting

```bash
# Check Gunicorn status
sudo systemctl status gunicorn

# View logs
sudo journalctl -u gunicorn -f
```

### Build Melange APK

```bash
docker run --rm --privileged -v "${PWD}:/work" cgr.dev/chainguard/melange build solana_web3.yaml
```

---

## Additional Resources

### Documentation

- [SETUP.md](SETUP.md) - Complete production setup guide with security best practices
- [API_DOCUMENTATION.md](packamal/API_DOCUMENTATION.md) - Complete REST API reference
- [ARCHITECTURE_DIAGRAM.md](packamal/ARCHITECTURE_DIAGRAM.md) - System architecture details
- [k8s/ARCHITECTURE.md](packamal/k8s/ARCHITECTURE.md) - Kubernetes deployment guide

### External Services

- **Micro Services**: [packamal_micro_services](https://github.com/pakaremon/packamal_micro_services.git)

### Useful Tutorials

- [Django + PostgreSQL Setup](https://www.digitalocean.com/community/tutorials/how-to-use-postgresql-with-your-django-application-on-ubuntu-22-04)
- [Nginx Installation](https://www.digitalocean.com/community/tutorials/how-to-install-nginx-on-ubuntu-22-04)
- [Nginx Configuration](https://www.digitalocean.com/community/tutorials/understanding-nginx-server-and-location-block-selection-algorithms)
- [Swap Space Configuration](https://www.digitalocean.com/community/tutorials/how-to-add-swap-space-on-ubuntu-20-04)

---

## License



## Contributing



## Support

For issues, questions, or feature requests, please contact the development team or refer to the project repository.

---

**Last Updated**: 2024
**Version**: 1.0