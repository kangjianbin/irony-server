#!/bin/bash

source env.sh

if [ -z "$LIBCLANG_HEADER" ]; then
    go build
else
    go build -ldflags "-X main.ClangHeaderDir=${LIBCLANG_HEADER}/"
fi
