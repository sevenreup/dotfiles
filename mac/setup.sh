#!/bin/bash

# macOS-specific setup script
# Installs and configures tools specific to macOS

set -e  # Exit on any error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up macOS-specific tools...${NC}"

# Check if Homebrew is installed
if ! command -v brew &> /dev/null; then
    echo -e "${YELLOW}Homebrew not found. Installing Homebrew...${NC}"
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    # Add Homebrew to PATH for Apple Silicon Macs
    if [[ $(uname -m) == "arm64" ]]; then
        echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/opt/homebrew/bin/brew shellenv)"
    else
        # Intel Macs
        echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/usr/local/bin/brew shellenv)"
    fi

    echo -e "${GREEN}✓ Homebrew installed successfully!${NC}"
else
    echo -e "${GREEN}✓ Homebrew is already installed${NC}"
fi

# Install Raycast
echo -e "${BLUE}Installing Raycast...${NC}"
if brew list --cask raycast &> /dev/null; then
    echo -e "${GREEN}✓ Raycast is already installed${NC}"
else
    echo -e "${YELLOW}Installing Raycast via Homebrew...${NC}"
    brew install --cask raycast
    echo -e "${GREEN}✓ Raycast installed successfully!${NC}"
fi

# Install Spicetify
echo -e "${BLUE}Installing Spicetify...${NC}"
if brew list spicetify-cli &> /dev/null; then
    echo -e "${GREEN}✓ Spicetify is already installed${NC}"
else
    echo -e "${YELLOW}Installing Spicetify via Homebrew...${NC}"
    brew install spicetify-cli
    echo -e "${GREEN}✓ Spicetify installed successfully!${NC}"
fi

echo -e "${GREEN}✓ macOS setup complete!${NC}"