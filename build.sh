#!/bin/bash

source env.sh

if [ -z "$LIBCLANG_HEADER" ]; then
    go build
else
    echo "Using builtin headers in ${LIBCLANG_HEADER}"
    go build -ldflags "-X main.ClangHeaderDir=${LIBCLANG_HEADER}/"
fi
