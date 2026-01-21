#!/bin/bash
# Security Scanning Script
# Runs secrets scanning, dependency scanning, and generates SBOMs

set -e

echo "=========================================="
echo "NeuronIP Security Scanning"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check for secrets
echo -e "${YELLOW}Scanning for secrets...${NC}"
if command -v gitleaks &> /dev/null; then
    gitleaks detect --source . --verbose || echo -e "${RED}Secrets found!${NC}"
else
    echo "Gitleaks not installed, skipping secrets scan"
fi

# Dependency scanning
echo ""
echo -e "${YELLOW}Scanning dependencies...${NC}"

# Go dependencies
if [ -f "api/go.mod" ]; then
    echo "Scanning Go dependencies..."
    cd api
    if command -v govulncheck &> /dev/null; then
        govulncheck ./... || echo -e "${RED}Vulnerabilities found!${NC}"
    else
        echo "govulncheck not installed, skipping"
    fi
    cd ..
fi

# Node.js dependencies
if [ -f "frontend/package.json" ]; then
    echo "Scanning Node.js dependencies..."
    cd frontend
    npm audit --audit-level=moderate || echo -e "${RED}Vulnerabilities found!${NC}"
    cd ..
fi

# Generate SBOMs
echo ""
echo -e "${YELLOW}Generating SBOMs...${NC}"

# Go SBOM
if command -v syft &> /dev/null; then
    syft packages dir:./api -o spdx-json > go-sbom.spdx.json
    echo -e "${GREEN}✓ Go SBOM generated${NC}"
else
    echo "Syft not installed, skipping SBOM generation"
fi

# Node.js SBOM
if [ -f "frontend/package.json" ] && command -v cyclonedx-npm &> /dev/null; then
    cd frontend
    cyclonedx-npm --output-file ../node-sbom.json
    cd ..
    echo -e "${GREEN}✓ Node.js SBOM generated${NC}"
fi

echo ""
echo -e "${GREEN}Security scanning complete!${NC}"
