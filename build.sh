#!/bin/sh

if [ -z "$LIBCLANG_DIR" ]; then
    LIBCLANG_DIR=$(llvm-config --prefix)/lib/clang/$(llvm-config --version)
fi

go build -ldflags "-X main.ClangHeaderDir=${LIBCLANG_DIR}/include"
