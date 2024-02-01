#!/bin/bash

for i in {1..100000}
do
   curl "http://localhost:80/api/$RANDOM"
done
