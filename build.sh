#!/usr/bin/env bash

x=$(git rev-list -1 HEAD)
gv="$(echo ${x:0:10}) <$(git rev-parse --abbrev-ref HEAD)>"
now=$(date "+%Y-%m-%d %H:%M:%S")
cwd=$PWD
go build -ldflags="-X 'main.Built=${now}' -X 'main.GitVersion=${gv}' -X 'main.Prefix=${cwd}'"