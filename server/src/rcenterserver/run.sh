#!/bin/bash

logPath="./log/"
configFile="../../bin/config/config.json"

go build -o rcenterserver
if [ $? -eq 0 ];then
    echo "rcenterserver编译成功"
    if [ ! -d "$logRoot" ];then
        mkdir "$logRoot"
    else
        echo "$logRoot""已存在"
    fi
    ./rcenterserver -config "$configFile" -log_dir="$logPath" -alsologtostderr
else
    echo "rcenterserver编译失败"
fi