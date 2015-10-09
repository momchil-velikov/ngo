#! /bin/sh

go test golang/scanner golang/parser golang/constexpr golang/build golang/resolve
go install scanfilt
go install parsefilt
go install pkgdep
