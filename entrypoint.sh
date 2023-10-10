#!/bin/sh

# Any initialization or setup commands you want to perform
echo "Container is starting..."

# Check if the target directory exists
if [ -d "/var/lib/spotify_notifications_bot/configs" ]; then
  # If it exists, move it to a backup location
  mv /var/lib/spotify_notifications_bot/configs /var/lib/spotify_notifications_bot/configs_old
fi

# Recreate the target directory
mkdir -p /var/lib/spotify_notifications_bot/configs

# Copy configuration files to the desired location
cp -r /root/configs/* /var/lib/spotify_notifications_bot/configs/

# Execute your main application or command
./bot

# Optionally, you can add any cleanup or post-processing commands here
echo "Container is shutting down..."
