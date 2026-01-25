from setuptools import setup, find_packages

setup(
    name="malicious-network-package",
    version="1.0.0",
    description="Sample package with network requests to dead URLs for testing network simulation",
    author="Pack-A-Mal Team",
    packages=find_packages(),
    install_requires=[
        "requests>=2.28.0",
    ],
    python_requires=">=3.7",
)
