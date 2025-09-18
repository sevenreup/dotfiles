#!/bin/bash

# Windows-specific setup script
# Installs and configures tools specific to Windows

set -e  # Exit on any error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up Windows-specific tools...${NC}"

# Check if winget is available
if ! command -v winget &> /dev/null; then
    echo -e "${RED}winget not found. Please install App Installer from Microsoft Store first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ winget is available${NC}"

# Install Spicetify
echo -e "${BLUE}Installing Spicetify...${NC}"
if winget list --id Spicetify.Spicetify --exact &> /dev/null; then
    echo -e "${GREEN}✓ Spicetify is already installed${NC}"
else
    echo -e "${YELLOW}Installing Spicetify via winget...${NC}"
    winget install --id Spicetify.Spicetify --exact
    echo -e "${GREEN}✓ Spicetify installed successfully!${NC}"
fi

echo -e "${GREEN}✓ Windows setup complete!${NC}"