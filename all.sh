#! /bin/sh
PKGS="
lib/sort
golang/scanner
golang/parser
golang/build
golang/pdb
golang/typecheck
"

PROGS="
scanfilt
parsefilt
pkgdep
"

go test -cover ${PKGS} && go install ${PROGS}
