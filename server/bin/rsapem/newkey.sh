#!/bin/bash

openssl genrsa -out rsa_private_key.pem 1024
# -out 指定生成文件
# 1024 生成的密钥长度

openssl rsa -in rsa_private_key.pem -pubout -out rsa_public_key.pem
