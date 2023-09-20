#!/bin/bash

# Colors
RED='\033[0;31m'
NC='\033[0m' # No Color

current_dir=$(pwd)
original_dir=$current_dir
# Determine the "github.com" or equivalent directory for cloning
base_repo_dir=$(basename $(dirname $(dirname "$original_dir")))

# Define the directories for the search
parent_dir=$(dirname "$original_dir")
grandparent_dir=$(dirname "$parent_dir")

# We'll also define the use_directory function to avoid repetition:
use_directory() {
    local dir=$1
    cd "$dir"
    make bindata
    make swagger
    echo "$dir" > "$original_dir/.found-swagger-path"
}

# Search in the immediate parent directory
if [ -d "$parent_dir/go-swagger" ]; then
    read -p "Use this path: $parent_dir/go-swagger? [y/N]: " response
    if [ "$response" = "y" ]; then
        use_directory "$parent_dir/go-swagger"
    fi
    exit 0
fi

# Check the grandparent directory
if [ -d "$grandparent_dir/go-swagger" ]; then
    read -p "Use this path: $grandparent_dir/go-swagger? [y/N]: " response
    if [ "$response" = "y" ]; then
        use_directory "$grandparent_dir/go-swagger"
    fi
    exit 0
fi

# If not found in either, loop through each subdirectory of the grandparent directory
for subdir in "$grandparent_dir"/*; do
    if [ -d "$subdir/go-swagger" ]; then
        read -p "Use this path: $subdir/go-swagger? [y/N]: " response
        if [ "$response" = "y" ]; then
            use_directory "$subdir/go-swagger"
            exit 0
        fi
    fi
done

# If not found, attempt to clone
repo_base=$(basename $(dirname "$original_dir"))
clone_dest=$(dirname "$original_dir")/go-swagger
read -p "Attempt to clone from https://$base_repo_dir/$repo_base/go-swagger? [y/N]: " response
if [ "$response" = "y" ]; then
    git clone "https://$base_repo_dir/$repo_base/go-swagger" "$clone_dest"
    if [ ! -d "$clone_dest" ]; then
        printf "${RED}Error:${NC} Failed to clone go-swagger!\n"
        exit 1
    fi
    cd "$clone_dest"
    make bindata
    make swagger
    echo "$clone_dest" > .found-swagger-path
else
    printf "${RED}Error:${NC} go-swagger not found!\n"
    exit 1
fi


