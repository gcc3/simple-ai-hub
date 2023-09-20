#!/bin/bash

# Iterate over all sub-directories
for dir in */; do
    # Check if docker-compose.yml exists in the directory
    if [ -f "$dir/docker-compose.yaml" ]; then
        echo "Starting Docker Compose in $dir"

        # Change to the directory
        cd "$dir" || continue

        # Start the Docker Compose instance
        docker compose up --build -d

        # Go back to the parent directory
        cd ..
    fi
done
