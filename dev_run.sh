#!/bin/bash

go run main.go --text-logging=true --jwt-secret=secret --backend "provider=simple,bob=secret"
