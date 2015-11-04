#! /bin/sh
PKGS="
lib/sort
golang/scanner
golang/parser
golang/constexpr
golang/build
golang/resolve
"

PROGS="
scanfilt
parsefilt
pkgdep
"

go test ${PKGS}
go install ${PROGS}
