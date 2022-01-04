#!/usr/bin/env sh

set -eu

cd $(dirname $0)
git branch -M github
git checkout -b main
find . \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -exec grep 'github.com/cnk3x' {} \;
find . \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -exec sed -i '' 's|github.com/cnk3x|gitee.com/k3x|g' {} \;
rm -f gitee.sh
git add .
git commit -m 'change to gitee'
git push -f -u gitee main
git checkout github
git branch -D main
git branch -M main
