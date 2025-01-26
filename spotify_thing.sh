#!/bin/bash

response=$(curl -s "https://api.spotify.com/v1/artists/4Z8W4fKeB5YxbusRsdQVPb" \
     -H "Authorization: Bearer BQAXm0Bx-X_Xt2DtlTxTcQNQ2h1y1JY_P4EgN0PFyZGIw5j9IH1XHYO-Iufk85HEnRxl-vlistsBr1Z44_cWdBCsk5MyOK3Q42iYUG5xhmfbeBizhMo"
)

echo "$response"

