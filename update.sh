#!/bin/bash

# 自動起動の停止
cd /root
wget -O etos-mainunit.1 https://www.astina.co/etos/aa4d58cb3ba4c07dc49ae13a3f460abb/etos-mainunit

MD5_1=`md5sum etos-mainunit`
MD5_2=`md5sum etos-mainunit.1`
MD5_1=$(cut -d' ' -f 1 <<<${MD5_1})
MD5_2=$(cut -d' ' -f 1 <<<${MD5_2})

if [ $MD5_1 != $MD5_2 ]; then
    # アップデート処理
    echo $MD5_1
    echo $MD5_2

    # systemctl stop etos-mainunit.service
    # rm etos-mainunit
    # mv etos-mainunit.1 etos-mainunit
    # chmod +x etos-mainunit
    # systemctl start etos-mainunit.service
fi
