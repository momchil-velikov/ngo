#! /bin/sh

go test golang/scanner golang/parser golang/constexpr golang/build
go install scanfilt
go install parsefilt
go install pkgdep
