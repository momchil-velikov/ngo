#! /bin/sh

go test lib/sort golang/scanner golang/parser golang/constexpr golang/build
go install scanfilt
go install parsefilt
go install pkgdep
