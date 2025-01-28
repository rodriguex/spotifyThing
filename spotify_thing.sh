#!/bin/bash

response=$(curl -X POST "https://accounts.spotify.com/api/token" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=client_credentials&client_id=d2493183072a47da84be16790faa88d9&client_secret=215d15e458e946ac85004e98d245da41")

echo "$response"