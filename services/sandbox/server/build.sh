#!/bin/bash

go build -o botbox-sandbox

docker build ./ -t "botbox-sandbox" --no-cache
