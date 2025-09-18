#!/bin/bash

# Development Setup Script for API Gateway
# This script sets up the development environment with file watching

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 API Gateway Development Setup${NC}"
echo "=================================="

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi

# Check if make is available
if ! command -v make >/dev/null 2>&1; then
    echo -e "${RED}❌ Make is not installed. Please install make first.${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Checking development tools...${NC}"

# Install development tools
echo -e "${BLUE}Installing development tools...${NC}"
make install-tools

# Check development status
echo -e "${BLUE}Checking development environment status...${NC}"
make dev-status

echo ""
echo -e "${GREEN}✅ Development setup complete!${NC}"
echo ""
echo -e "${YELLOW}Available commands:${NC}"
echo "  make dev           - Start development environment with hot reload"
echo "  make dev-local     - Start local development (requires air)"
echo "  make watch-files   - Watch for file changes"
echo "  make compose-logs  - View container logs"
echo "  make dev-status    - Check development status"
echo ""
echo -e "${YELLOW}Quick start:${NC}"
echo "  make dev"
echo ""
echo -e "${BLUE}Happy coding! 🎉${NC}"
