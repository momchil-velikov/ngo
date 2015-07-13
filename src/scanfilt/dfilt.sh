#! /bin/bash

dir=$1

for f in $(find ${dir} -name '*.go'); do
    echo "$f"
    ./bin/scanfilt < $f > ${f}.scanfilt.1
done

for f in $(find ${dir} -name '*.go.scanfilt.1'); do
    dst=$(dirname $f)/$(basename -s .scanfilt.1 ${f}).scanfilt.2
    echo "$f"
    ./bin/scanfilt < $f > ${dst}
done

for f1 in $(find ${dir} -name '*.go.scanfilt.1'); do
    f2=$(dirname ${f1})/$(basename -s .scanfilt.1 ${f1}).scanfilt.2
    if diff -qud ${f1} ${f2}
    then
        rm $f1 $f2
    fi
done
