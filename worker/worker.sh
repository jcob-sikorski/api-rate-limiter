#!/bin/bash

# Define the file names
FILE_NAME="/config/config.yaml"
TMP_FILE_NAME="config_tmp.yaml"

# Copy the file to a temporary location
cp $FILE_NAME ./$TMP_FILE_NAME

# Store the configuration in Redis
cat $TMP_FILE_NAME | redis-cli -h redis -p 6379 -x set config

# Remove the temporary file
rm $TMP_FILE_NAME
