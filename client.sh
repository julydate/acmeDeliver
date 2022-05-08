#!/bin/bash

DEBUG=true
WORKDIR="/tmp/acme" #工作目录，默认为/tmp/acme

# Variables corresponding to the deployment type
# 以下为自动安装证书的目的地设置
# manual
manual_cert_file=""
manual_key_file=""
manual_fullchain_file=""
#manual_reloadcmd="" # To execute commands after updating the certificate, uncomment and configure the content yourself 若要更新证书后执行命令，请取消注释并自行配置内容

# apache
apache_cert_file=""
apache_key_file=""
apache_fullchain_file=""
#apache_reloadcmd="service apache2 force-reload" # To execute commands after updating the certificate, uncomment and configure the content yourself 若要更新证书后执行命令，请取消注释并自行配置内容

# nginx
nginx_cert_file=""
nginx_key_file=""
#nginx_reloadcmd="service nginx force-reload" # To execute commands after updating the certificate, uncomment and configure the content yourself 若要更新证书后执行命令，请取消注释并自行配置内容

getTimestamp(){
  timestamp=$(date '+%s')
}

generateRandom(){
  checkSum=$(echo -n "$RANDOM$timestamp"|md5sum|cut -d ' ' -f1) # 将checksum改为随机性更高的md5(random+timestamp)
}

calculateToken(){
  token=$(echo -n "${domain}$1${password}${timestamp}${checkSum}"|md5sum|cut -d ' ' -f1)
  if $DEBUG; then echo "token:$token"; fi
}

requestServer(){
  # $1:server address, $2:filename
  generateRandom #每次请求生成一个新的随机值
  calculateToken "$2"
  url=$1'?domain='$domain'&file='$2'&t='$timestamp'&sign='$token'&checksum='$checkSum
  requestRes=$(curl -s -f -o "${WORKDIR}/${domain}/temp" -w %{http_code} "$url")
  if $DEBUG; then echo "requestUrl: $url"; echo "requestResult: $requestRes"; fi
}

checkUpdate(){
  if [ ! -d "${WORKDIR}" ]; then #判断工作目录是否存在
    mkdir "${WORKDIR}"
  fi
  if [ ! -d "${WORKDIR}/${domain}" ]; then #判断域名目录是否存在
    mkdir "${WORKDIR}/${domain}"
  fi
  if [ -f "${WORKDIR}/${domain}/timestamp.txt" ]; then
    #检测服务器时间戳是否有更新
    requestServer "${server}" "time.log" #向服务端请求服务端更新时间戳
    if [ "$requestRes" != '200' ]; then
      echo "请求服务端时间戳失败！"
      #exit 1 #取消注释以在服务端响应时间戳失败时退出脚本
    fi

    # shellcheck disable=SC2162
    read ts1 < "${WORKDIR}/${domain}/temp" && read ts2 < "${WORKDIR}/${domain}/timestamp.txt"
    if [ "$ts1" = "$ts2" ]; then
        echo "时间戳相同，无需更新"
        return
    else
        echo "时间戳不同，将会开始下载"
    fi
  else
    echo "本地不存在时间戳文件，将会开始下载"
    requestServer "${server}" "time.log" #向服务端请求服务端更新时间戳
    if [ "$requestRes" != '200' ]; then
      echo "请求服务端时间戳失败！"
      #exit 1 #取消注释以在服务端响应时间戳失败时退出脚本
    else
      cp -f "${WORKDIR}/${domain}/temp" "${WORKDIR}/${domain}/timestamp.tmp" #将刚对比的时间戳保存
    fi
  fi
  mv -f "${WORKDIR}/${domain}/temp" "${WORKDIR}/${domain}/timestamp.tmp" #将刚对比的时间戳临时保存

  for file_name_d in "cert.pem" "key.pem" "fullchain.pem"
  do
    echo "下载文件名：${file_name_d}"
    requestServer "${server}" "${file_name_d}" #向服务端请求服务端更新时间戳
    if [ "$requestRes" != '200' ]; then
      printf "请求文件失败！文件名：%s,状态码：%s \n" "$file_name_d" "$requestRes"
      flag_fail=true
      #exit 1 #取消注释以在服务端响应文件失败时退出脚本
    fi
    mv -f "${WORKDIR}/${domain}/temp" "${WORKDIR}/${domain}/${file_name_d}"
  done
  if [ ${flag_fail} ]; then
    echo "下载过程中有文件下载失败！"
    return 1
  fi
  mv -f "${WORKDIR}/${domain}/timestamp.tmp" "${WORKDIR}/${domain}/timestamp.txt" #将刚对比的时间戳作为保存的时间戳
  # 删除临时文件
  if [ -f "${WORKDIR}/${domain}/temp" ]; then
    rm -f "${WORKDIR}/${domain}/temp"
  fi
  return 0
}

deployCert(){
  # $1 deploy_type:manual,apache,nginx
  # if no need to deploy
  if [ "$1" = "0" ]; then
    return

  # manual
  elif [ "$1" = "m" ]; then
    if $DEBUG; then echo "deploy_type:manual"; fi
    if [ -z "$manual_cert_file" ] || [ -z "$manual_key_file" ] || [ -z "$manual_fullchain_file" ]; then
      echo "未配置证书目的地变量"
      exit 1
    fi
    cp -f "${WORKDIR}/${domain}/cert.pem" "$manual_cert_file"
    cp -f "${WORKDIR}/${domain}/key.pem" "$manual_key_file"
    cp -f "${WORKDIR}/${domain}/fullchain.pem" "$manual_fullchain_file"
    if [ -n "$manual_reloadcmd" ]; then
      cd "${WORKDIR}/${domain}" && eval "$manual_reloadcmd"
    fi

  # apache
  elif [ "$1" = "a" ]; then
    if $DEBUG; then echo "deploy_type:apache"; fi
    if [ -z "$apache_cert_file" ] || [ -z "$apache_key_file" ] || [ -z "$apache_fullchain_file" ]; then
      echo "未配置证书目的地变量"
      exit 1
    fi
    cp -f "${WORKDIR}/${domain}/cert.pem" "$apache_cert_file"
    cp -f "${WORKDIR}/${domain}/key.pem" "$apache_key_file"
    cp -f "${WORKDIR}/${domain}/fullchain.pem" "$apache_fullchain_file"
    if [ -n "$apache_reloadcmd" ]; then
      cd "${WORKDIR}/${domain}" && eval "$apache_reloadcmd"
    fi

  # nginx
  elif [ "$1" = "n" ]; then
    if $DEBUG; then echo "deploy_type:nginx"; fi
    if [ -z "$nginx_cert_file" ] || [ -z "$nginx_key_file" ]; then
      echo "未配置证书目的地变量"
      exit 1
    fi
    cp -f "${WORKDIR}/${domain}/cert.pem" "$nginx_cert_file"
    cp -f "${WORKDIR}/${domain}/key.pem" "$nginx_key_file"
    if [ -n "$nginx_reloadcmd" ]; then
      cd "${WORKDIR}/${domain}" && eval "$nginx_reloadcmd"
    fi
  fi
}

remove_workdir(){
  for dir_name_d in "/" "/boot" "/bin" "/dev" "/lib" "/lib64" "/proc" "/run" "/usr" "/usr/bin" "/etc" "/root"
  do
    if [ "${WORKDIR}" = "$dir_name_d" ]; then
      echo "ERROR! Your workdir is set to a dangerous path! The remove process will automatically stop!
  错误！您的workdir设置为危险路径！删除进程将自动停止！"
      exit 1
    fi
  done
  echo "Please be sure that ${WORKDIR} has no other important files!
请确保${WORKDIR}没有其他重要文件！"
  read -p "Are you sure?[Y/n]" sure_flag
  if [ "${sure_flag}" = "Y" ] || [ "${sure_flag}" = "y" ]; then
    rm -rf "${WORKDIR}" && echo "删除完成"
  fi

}

echo_help(){
  echo "Usage: [-c execute_check_update_job_type(m,a,n)] [-h help] [-d domain name] [-p password] [-s server address] [-n file name] [-w workdir(default:/tmp/acme)] [-r remove workdir files]
-c m    ------manually get cert files
   a    ------deploy cert files to apache
   n    ------deploy cert files to nginx
   0    ------only update, don't deploy
CAUTION! Variables corresponding to the deployment type must be defined
使用方法：[-c 执行自动更新任务类型(m,a,n)] [-h 帮助] [-d 域名] [-p 密码] [-s acmeDeliver服务器地址] [-n 要获取的文件名] [-w 工作目录(默认:/tmp/acme)] [-r 清除工作目录文件]
-c m    ------手动获得证书文件
   a    ------部署证书至apache
   n    ------部署证书至nginx
   0    ------仅更新证书，不部署
注意！ 必须定义部署类型对应的变量
"
}

#解析命令行参数
while getopts "rhc:p:s:d:n:w:f:" arg #选项后面的冒号表示该选项需要参数
do
  case $arg in
    h)
      echo_help
      exit 0
      ;;
    c)
      check_update_job=true
      deploy_type=$OPTARG
      ;;
    p)
      password=$OPTARG
      if $DEBUG; then echo "password:$password"; fi
      ;;
    s)
      server=$OPTARG
      if $DEBUG; then echo "server address:$server"; fi
      ;;
    d)
      domain=$OPTARG
      if $DEBUG; then echo "domain:$domain"; fi
      ;;
    n)
      filename=$OPTARG
      if $DEBUG; then echo "filename:$filename"; fi
      ;;
    w)
      WORKDIR=$OPTARG
      if $DEBUG; then echo "workdir:$WORKDIR"; fi
      ;;
    r)
      remove_workdir_job=true
      ;;
    f)
      cert_env_file=$OPTARG
      if $DEBUG; then echo "environment file:$cert_env_file"; fi
      ;;
    ?)  #当有不认识的选项的时候arg为?
      echo "unknown argument"
      echo_help
      exit 1
    ;;
  esac
done

main(){
  # 检测是否缺少必要参数
  if [ -z "$server" ] || [ -z "$domain" ] || [ -z "$password" ]; then
    echo "缺少必要参数"
    exit 1
  fi

  if [ -f "$cert_env_file" ]; then
    printf "发现配置文件， 将从 %s 内读取参数 \n" "$cert_env_file"
    set -o allexport
    source "$cert_env_file"
    set +o allexport
  fi

  getTimestamp #获取当前时间戳
  if test ${remove_workdir_job}; then remove_workdir; exit 0; fi
  if test ${check_update_job}; then checkUpdate; deployCert "$deploy_type"; exit 0; fi

  # 未设置工作模式时默认是获取指定文件
  if [ -z "$filename" ]; then
    echo "缺少文件名"
    exit 1
  fi

  requestServer "$server" "$filename" #请求服务器指定文件
  if [ "$requestRes" != '200' ]; then
    printf "请求文件失败！文件名：%s,状态码：%s \n" "$file_name_d" "$requestRes"
    exit 1
  fi

  mv -f "${WORKDIR}/${domain}/temp" "$filename" #默认存放在命令行工作目录下
  # shellcheck disable=SC2046
  printf "下载成功！文件保存在%s/%s \n" $(pwd) "${filename}"
  exit 0
}
main
