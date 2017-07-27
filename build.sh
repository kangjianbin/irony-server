#!/bin/bash

source env.sh

if [ -z "$LIBCLANG_HEADER" ]; then
    LIBCLANG_DIR=$(llvm-config --prefix)/lib/clang/$(llvm-config --version)
fi

go build -ldflags "-X main.ClangHeaderDir=${LIBCLANG_HEADER}/include"
