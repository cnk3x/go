#!/usr/bin/env sh

cd $(dirname $0)

rm -f go.work go.work.sum

go work init
IFS_OLD=IFS
IFS=$'\n'
for n in $(find . -name 'go.mod'); do
    n=$(echo ${n} | sed 's/\/go.mod//g')
    echo $n
    go work edit -use $n
done
IFS=${IFS_OLD}
rm -f go.work.sum
go work sync
