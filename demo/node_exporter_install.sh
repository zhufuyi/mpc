#!/bin/bash

# exporter名称
exporter_name="node_exporter"
# exporter安装目录
work_path="/opt/node_exporter"

check_md5(){
    verify_file=$1
    if [ ! -e ${verify_file} -a ! -e ${verify_file}.md5 ]; then
        echo "${verify_file} or ${verify_file}.md5 does not exist"
        exit 1;
    fi

    if [ "$(md5sum ${verify_file} | awk '{print $1}')" == "$(cat ${verify_file}.md5 | awk '{print $1}')" ]; then
        echo "${verify_file} MD5 verification succeeded"
    else
        echo "${verify_file} MD5 verification failed"
        exit 1;
    fi
}

# 获取到目标文件，如果有多个文件，指定过滤文件名
targetFile=""
function listFiles(){
    cd $1
    items=$(ls)

    for item in $items
    do
        if [ -d "$item" ]; then
            listFiles $item
        else
            if [ "${item}" == "${exporter_name}" ]; then
                targetFile=$(pwd)/${item}
            fi
        fi
    done
    cd ..
}

# 杀掉指定名称相关进程
killprocess(){
    local processName=$1

    if [ "${processName}"x = x ] ;then
        echo "killprocess() error, process name is empty."
        exit 1
    fi

    ID=`ps -ef | grep "$processName" | grep -v "$0" | grep -v "grep" | awk '{print $2}'`
    if [ -n "$ID" ]; then
        echo "find the ${processName} related process ID: ${ID}"

        for id in $ID
        do
           kill -9 $id
           echo "killed $id succeed."
        done
    fi
}

# 删除原始文件
removeUploadFiles(){
    rm -rf ${upload_path}/${compressed_filename}*
    rm -rf ${upload_path}/${tmpPath}
	rm -rf ${shell_file}*
}

main(){
    shell_file="$0"
    upload_path="$1"
    compressed_filename="$2"

    if [ $# != 2 ] ; then
       echo "USAGE: sh $0 <upload_path> <compressed_filename>"
       exit 1;
    fi

    # 切换到上传目录
    cd ${upload_path}

    # 检查文件md5
	check_md5 $(basename ${shell_file})
    check_md5 ${compressed_filename}

    # 解压文件到临时目录
    tmpPath="tmpPath_$(date "+%Y%m%d%H%M%S")"
    mkdir -p ${tmpPath}
    echo "unzip directory: ${upload_path}/${tmpPath}"
    if [ "${compressed_filename##*.}"x = "gz"x ];then
        tar zxvf ${compressed_filename} -C ${tmpPath}
    elif [ "${compressed_filename##*.}"x = "zip"x ];then
	    unzip -o ${compressed_filename} -d ${tmpPath}
	else
	    echo "${compressed_filename}文件类型不合法，只支持zip或tar.gz压缩文件类型"
		exit 1
    fi
    
    # 根据exporter名称获取目标文件
    listFiles ${tmpPath}
    echo "target file：${targetFile}"

    # 判断是否找到exporter
    if [ "${targetFile}x" = "x" ] ;then
        echo "not found ${exporter_name} file."
        exit 1
    fi
    chmod +x ${targetFile}

    # 如果需要备份就文件，先备份再替换

    # 替换旧的exporter
    mkdir -p ${work_path}
    /bin/cp -f ${targetFile} ${work_path}/

    # 杀掉旧进程
    killprocess ${exporter_name}

    # 启动exporter
    cd ${work_path}
    nohup ${work_path}/${exporter_name} >> out.log 2>&1 &

    sleep 1
    ID=`ps -ef | grep "${exporter_name}" | grep -v "$0" | grep -v "grep" | awk '{print $2}'`
    if [ -n "$ID" ]; then
        echo "ID=$ID, ${exporter_name} started successfully"
    else
        echo "${exporter_name} start failed"
		exit 1
    fi
}

main $1 $2
removeUploadFiles
