#!/bin/bash

# 定义要执行的文件夹列表
folders=("Reti_code_average_10" "Reti_code_average_20" "Reti_code_average_30" "Reti_code_average_33" "Reti_code_bankrun_10" "Reti_code_bankrun_20" "Reti_code_bankrun_30" "Reti_code_bankrun_33" "Reti_code_suicidal_10" "Reti_code_suicidal_20" "Reti_code_suicidal_30" "Reti_code_suicidal_33" "Reti_code_worst_10" "Reti_code_worst_20" "Reti_code_worst_30" "Reti_code_worst_33")

# 循环遍历文件夹并执行其中的start.sh文件
for folder in "${folders[@]}"; do
    if [ -d "$folder" ]; then
        # 添加执行权限
        chmod +x "$folder/start.sh"
        cd "$folder" || exit 1  # 切换到文件夹，如果切换失败则退出
        if [ -f "start.sh" ]; then
            echo "执行 $folder/start.sh"
            ./start.sh
        else
            echo "在 $folder 中找不到 start.sh 文件"
        fi
        cd ..  # 返回上级目录
    else
        echo "文件夹 $folder 不存在"
    fi
done
