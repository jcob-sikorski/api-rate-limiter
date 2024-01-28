#!/bin/bash

# Replace 'your_uid' with the actual user ID
uid='your_uid'

for i in {1..100}; do
    echo "
    
Calling API: $i"
    
    # Make the API request using curl
    curl -X POST http://localhost:3000/api/$uid -H "Content-Type: application/json" -d '{"example": "data"}'
done
