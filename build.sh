#!/usr/bin/env bash

set -x

# 获取源码最近一次 git commit log，包含 commit sha 值，以及 commit message
GitCommitLog=`git log --pretty=oneline -n 1`
# 将 log 原始字符串中的单引号替换成双引号
GitCommitLog=${GitCommitLog//\'/\"}
# 检查源码在git commit 基础上，是否有本地修改，且未提交的内容
GitStatus=`git status -s`
# 获取当前时间
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
# 获取 Go 的版本
BuildGoVersion=`go version`

# 将以上变量序列化至 LDFlags 变量中
LDFlags=" \
    -X 'github.com/bCoder778/qitmeer-sync/version.GitCommitLog=${GitCommitLog}' \
    -X 'github.com/bCoder778/qitmeer-sync/version.GitStatus=${GitStatus}' \
    -X 'github.com/bCoder778/qitmeer-sync/version.BuildTime=${BuildTime}' \
    -X 'github.com/bCoder778/qitmeer-sync/version.BuildGoVersion=${BuildGoVersion}' \
"

ROOT_DIR=`pwd`

# 如果可执行程序输出目录不存在，则创建
if [ ! -d ${ROOT_DIR}/bin ]; then
  mkdir bin
fi

# 编译多个可执行程序
cd ${ROOT_DIR} && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/linux/qitmeer-sync &&
cd ${ROOT_DIR} && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/darwin/qitmeer-sync &&
ls -lrt ${ROOT_DIR}/bin &&
echo 'build done.'