#!/bin/bash

port="800"$1
logRoot="./log/"
logPath="./log/logfile_"$1"/"
configFile="../../bin/config/config.json"

go build -o logicserver_$1
if [ $? -eq 0 ];then
    echo "logicserver编译成功"
    # 日志根目录
    if [ ! -d "$logRoot" ];then
        mkdir "$logRoot"
    else
        echo "$logRoot""已存在"
    fi
    # 日志文件夹
    if [ ! -d "$logPath" ];then
        mkdir "$logPath"
    else
        echo "$logPath""已存在"
    fi
    # 程序执行
    ./logicserver_$1 -config "$configFile" -port "$port" -log_dir="$logPath" -alsologtostderr
else
    echo "logicserver编译失败"
fi
