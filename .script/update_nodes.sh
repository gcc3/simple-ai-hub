#!/bin/bash

# Iterate over directories starting with "node"
for dir in node*/; do
    if [ -d "$dir/.git" ]; then
        echo "Updating repository in $dir"
        
        # Change to the directory
        cd "$dir" || continue

        # Fetch and pull the latest changes
        git fetch --all
        git pull

        # Go back to the parent directory
        cd ..
    fi
done
