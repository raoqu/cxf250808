#!/bin/sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o hbjsinfo && scp hbjsinfo root@10.6.0.1:/www/wwwroot/hbjs/
