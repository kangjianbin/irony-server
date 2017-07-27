#!/bin/bash

llvm_config="llvm-config"
has_llvm_config=0
hash $llvm_config 2>/dev/null
if [ $? = "0" ]; then
    has_llvm_config=1
fi
os=$(uname)


if [ -z "$LLVM_DIR" ]; then
    if [[ $has_llvm_config == 0 ]]; then
        echo "Can't find $llvm_config. Please set environment variable LLVM_DIR"
        exit 1
    fi
    LLVM_DIR=$($llvm_config --prefix)
    export CGO_CFLAGS=$($llvm_config --cflags)
    export CGO_LDFLAGS=$($llvm_config --ldflags)
else
    export CGO_CFLAGS="-I$LLVM_DIR/include"
    if [ -z "$LLVM_LIB" ]; then
        if [ "$os" = "Linux" ]; then
            export CGO_LDFLAGS="-L$LLVM_DIR/lib"
        else
            export CGO_LDFLAGS="-L$LLVM_DIR/bin"
        fi
    else
        export CGO_LDFLAGS="-L$LLVM_LIB"
    fi
fi

clang_header="$LLVM_DIR/lib/clang/*/include"
[ -e $clang_header ] && export LIBCLANG_HEADER="$clang_header"
