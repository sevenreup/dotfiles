#!/bin/bash

# Generic dotfiles setup script
# This script creates symbolic links from all folders in configs/.config/ to ~/.config/

set -e  # Exit on any error

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define paths
SOURCE_CONFIG_DIR="$SCRIPT_DIR/configs/.config"
TARGET_CONFIG_DIR="$HOME/.config"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up dotfiles...${NC}"

# Check if source .config directory exists
if [ ! -d "$SOURCE_CONFIG_DIR" ]; then
    echo -e "${RED}Error: Source directory '$SOURCE_CONFIG_DIR' does not exist.${NC}"
    exit 1
fi

# Create ~/.config directory if it doesn't exist
if [ ! -d "$TARGET_CONFIG_DIR" ]; then
    echo -e "${YELLOW}Creating ~/.config directory...${NC}"
    mkdir -p "$TARGET_CONFIG_DIR"
fi

# Function to handle existing files/directories
handle_existing() {
    local target_path="$1"
    local name="$2"

    if [ -L "$target_path" ]; then
        echo -e "${YELLOW}  Removing existing symlink for '$name'...${NC}"
        rm "$target_path"
    elif [ -d "$target_path" ]; then
        echo -e "${YELLOW}  Backing up existing directory '$name' to '$name.backup'...${NC}"
        mv "$target_path" "$target_path.backup"
    elif [ -f "$target_path" ]; then
        echo -e "${YELLOW}  Backing up existing file '$name' to '$name.backup'...${NC}"
        mv "$target_path" "$target_path.backup"
    fi
}

# Counter for successful links
success_count=0
total_count=0

# Loop through all items in the source .config directory
for item in "$SOURCE_CONFIG_DIR"/*; do
    # Skip if no items found (glob didn't match anything)
    [ -e "$item" ] || continue

    # Get the basename (folder/file name)
    name=$(basename "$item")

    # Define source and target paths
    source_path="$item"
    target_path="$TARGET_CONFIG_DIR/$name"

    echo -e "${BLUE}Processing: $name${NC}"

    # Increment total counter
    ((total_count++))

    # Handle existing files/directories/symlinks
    if [ -e "$target_path" ]; then
        handle_existing "$target_path" "$name"
    fi

    # Create the symbolic link
    echo -e "${GREEN}  Creating symbolic link for '$name'...${NC}"
    if ln -s "$source_path" "$target_path"; then
        # Verify the link was created successfully
        if [ -L "$target_path" ] && [ -e "$target_path" ]; then
            echo -e "${GREEN}  ✓ Successfully linked '$name'${NC}"
            ((success_count++))
        else
            echo -e "${RED}  ✗ Failed to verify link for '$name'${NC}"
        fi
    else
        echo -e "${RED}  ✗ Failed to create link for '$name'${NC}"
    fi

    echo
done

# Summary
echo -e "${BLUE}Setup Summary:${NC}"
echo -e "  Successfully linked: ${GREEN}$success_count${NC} out of ${BLUE}$total_count${NC} items"

if [ $success_count -eq $total_count ] && [ $total_count -gt 0 ]; then
    echo -e "${GREEN}✓ All dotfiles setup complete!${NC}"
elif [ $total_count -eq 0 ]; then
    echo -e "${YELLOW}No configuration folders found in '$SOURCE_CONFIG_DIR'${NC}"
else
    echo -e "${YELLOW}Some items may have failed. Check the output above.${NC}"
fi
