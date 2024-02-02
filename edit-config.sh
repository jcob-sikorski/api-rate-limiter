#!/bin/bash

# Define the volume and file names
VOLUME_NAME="go_rate_limiter_config"
FILE_NAME="config.yaml"
TMP_FILE_NAME="config_tmp.yaml"

# Use Docker to copy the file out of the volume
docker cp $(docker create --rm -v $VOLUME_NAME:/vol busybox):/vol/$FILE_NAME ./$TMP_FILE_NAME

# Open the temporary file with nano for editing
nano $TMP_FILE_NAME

# Once editing is done, move the temporary file back into the volume
docker cp $TMP_FILE_NAME $(docker create --rm -v $VOLUME_NAME:/vol busybox):/vol/$FILE_NAME

# Remove the temporary file
rm $TMP_FILE_NAME
