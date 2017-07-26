#!/bin/sh

if [ -z "$LLVM_DIR" ]; then
    export CGO_CFLAGS=$(llvm-config --cflags)
else
    export CGO_CFLAGS="-I$LLVM_DIR/include"
fi

if [ -z "$LLVM_LIB" ]; then
    export CGO_LDFLAGS=$(llvm-config --ldflags)
else
    export CGO_LDFLAGS="-L$LLVM_LIB"
fi
