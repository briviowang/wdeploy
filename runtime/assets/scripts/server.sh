#!/bin/bash
SERVER_PATH='{{.config.ServerPath}}'
action='{{.action}}'
upload_zip_name='dist.zip'
encrypt_php_dir='{{.config.EncryptPHPDirectory}}'
encrypt_exclude_dir='{{.config.EncryptExcludeDir}}'
tp_version='tp{{.config.TpVersion}}'

{{if .logColor}}
COLOR_RED="\033[31m"
COLOR_GREEN="\033[32m"
COLOR_YELLOW="\033[33m"
COLOR_BLUE="\033[34m"
COLOR_END="\033[0m"
{{end}}

cd $SERVER_PATH

_echo(){
    printf "$1\n"
}

_error_log(){
    _echo "$COLOR_RED$1$COLOR_END"
}

_system_name(){
    cat /etc/*-release|grep '^ID='|awk -F= '{printf $2}'|sed 's/"//g'
}

_system_version(){
    cat /etc/*-release|grep '^VERSION_ID='|awk -F= '{printf $2}'|sed 's/"//g'
}

_rm () {
    while [[ $# > 0 ]]; do
        if [[ -r $1 ]]; then
            rm -rf $1
        fi
        shift
    done
}
_install_package () {
    while [[ $# > 0 ]]; do
        if ! type $1 >&/dev/null; then
            if type yum >&/dev/null; then
                yum install -y $1
            else
                if type apt-get >&/dev/null; then
                    apt-get install $1
                else
                    if type brew >&/dev/null; then
                        brew install $1
                    fi
                fi
            fi
        fi
        shift
    done
}

_command_exist () {
    type "$1" >&/dev/null
}

_get_extension(){
    filename=$(basename "$1")
    printf ${filename##*.}|tr 'A-Z' 'a-z'
}

_check_remote () {
    if [[ -f ./$upload_zip_name ]]; then
        if [[ $action = 'remove' ]]; then
            _echo "*Delete ./_core ./app ./static"
            rm -rf ./_core ./app ./static
        fi
        _echo "*Unzip $SERVER_PATH/$upload_zip_name"

        unzip -q -o ./$upload_zip_name -d ./

        # 设置相关文件权限
        files=$(unzip -l ./$upload_zip_name|head -n-2|tail -n+4|awk '{printf $4"\n"}')
        assets_ext_list=("jpg",'jpeg','webp','png','bmp','gif','mp4','avi','mov','rm','rmvb','mpeg','mkv')
        for i in $files
        do
            ext=$(_get_extension "$i")
            if [[ "${assets_ext_list[@]}" =~ "${ext}" ]]; then
                chmod u=rw,go=r $i
            fi
        done
        rm ./$upload_zip_name
    else
        _echo "${COLOR_YELLOW}Working directory: $SERVER_PATH${COLOR_END}"
    fi

    sdk_dir=./tools/sdk
    if [[ -d $sdk_dir ]]; then
        cd $sdk_dir
        doc_zip_list=$(find . -name "*-doc.zip" -type f)
        for i in $doc_zip_list
        do
            _echo "*Unzip api document($i)"
            file_name=$(basename $i)
            unzip -q -o $i -d ./${file_name%%.*}
        done
        find . -name "*-doc.zip" -type f -delete
        cd ../..
    fi

    # _echo "*检查shell脚本权限"
    shell_dir='./tools'
    if [[ -d $shell_dir ]];then
        chmod -R 755 $shell_dir
    fi

    # _echo "*删除冗余文件"
    _rm ./static/node_modules
    # find . -name "*.bat" -o -name "*.less" -o -name ".*" ! -name "." ! -name ".." ! -name ".htaccess" -type f -exec rm -rf {} \; > /dev/null
    # if [[ -d ./data/runtime ]]; then
    #     find ./data/runtime/ -type f -exec rm -rf {} \; > /dev/null
    # fi

    if [[ $tp_version = 'tp6' ]];then
        echo "">/dev/null
    else
        _echo "*Checking Directory"
        for i in ./data/upload ./data/runtime
        do
            if [[ ! -d $i ]]; then
                mkdir -p $i
            fi
        done

        for i in ./data/upload ./data/runtime
        do
            if [[ -d $i ]];then
                chmod -R 777 $i &>/dev/null
            fi
        done
    fi

    if [[ -d ./app ]];then
        model_list=$(find ./app -type d -name "model")
        for i in $model_list
        do
            if [[ -d $i ]];then
                chmod -R 777 $i
            fi
        done
    fi

    if [[ -n $encrypt_php_dir ]]; then
        php_exist_code="var_export(function_exists('beast_encode_file'));"
        if [[ $(php -r "$php_exist_code") = 'true' ]]; then
            _echo "*encrypt PHP files..."
            for d in $encrypt_php_dir
            do
                files=$(eval "find $d -type f -name \"*.php\" $encrypt_exclude_dir")
                for i in $files
                do
                    if [[ -n $(cat $i|grep '<?php') ]]; then
                        echo $i
                        php -r "beast_encode_file('$i','$i',0,BEAST_ENCRYPT_TYPE_AES);"
                    fi
                done
            done
        else
            _error_log "*mbeast没有配置，无法加密PHP代码\n"
        fi
    fi
}

{{if eq .action "lnmp"}}
    if [[ $(_system_name) != 'centos' ]];then
        echo 'need centos'
        exit
    fi

    echo "*Checking php"
    if ! _command_exist php;then
        php=php72w
        rpm -Uvh https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
        rpm -Uvh https://mirror.webtatic.com/yum/el7/webtatic-release.rpm

        yum -y install $php ${php}-cli ${php}-common ${php}-devel ${php}-embedded ${php}-fpm ${php}-gd ${php}-mbstring ${php}-mysqlnd ${php}-opcache ${php}-pdo ${php}-xml
    fi

    echo "*Checking mysql"
    if ! _command_exist mysql;then
        rpm -Uvh http://dev.mysql.com/get/mysql-community-release-el7-5.noarch.rpm
        yum -y install mysql-community-server
        systemctl start  mysqld
        systemctl enable  mysqld
    fi

    echo "*Checking apache"
    if ! _command_exist httpd;then
        yum -y install httpd httpd-manual mod_ssl mod_perl
        systemctl start  httpd
        systemctl enable  httpd
    fi
    exit
{{end}}

{{if eq .action "update_mysql_conf"}}
    my_cnf=/etc/my.cnf

    if [[ ! -f $my_cnf ]];then
        echo $my_cnf" not exist"
        exit
    else
        if [[ ! -f $my_cnf.backup ]];then
            cp $my_cnf $my_cnf.backup
        fi

        python <<EOF
import ConfigParser

config_path = '$my_cnf'
config = ConfigParser.RawConfigParser()
config.read(config_path)

mysqld_section = 'mysqld'
config.set(mysqld_section, 'character-set-server', 'utf8mb4')
config.set(mysqld_section, 'default-storage-engine', 'MYISAM')
config.set(mysqld_section, 'sql_mode',
           'ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION')

with open(config_path, 'wb') as configfile:
    config.write(configfile)
EOF
        for db in mysqld mariadb
        do
            if [[ $(systemctl list-units|grep "$db.service"|wc -l) -gt 0 ]];then
                systemctl restart $db
            fi
        done
    fi
    exit
{{end}}

{{if eq .action "update_php_conf"}}
    my_cnf=/etc/php.ini

    if [[ ! -f $my_cnf ]];then
        echo $my_cnf" not exist"
        exit
    else
        if [[ ! -f $my_cnf.backup ]];then
            cp $my_cnf $my_cnf.backup
        fi

        python <<EOF
import ConfigParser

config_path = '$my_cnf'
config = ConfigParser.RawConfigParser()
config.read(config_path)

config.set('PHP', 'short_open_tag', 'On')
config.set('Date', 'date.timezone', 'Asia/Shanghai')
config.set('Session', 'session.cookie_lifetime', '2592000')

with open(config_path, 'wb') as configfile:
    config.write(configfile)
EOF
    fi
    exit
{{end}}

{{if eq .action "install_beast"}}
    wget https://github.com/liexusong/php-beast/archive/master.zip
    unzip master.zip
    cd php-beast-master
    phpize
    ./configure
    make
    make install
    cd ..
    rm -rf ./master.zip ./php-beast-master
    exit
{{end}}

echo "*Checking system environment"

# 判断是否需要输入root密码
if [[ $(id -u) != 0 ]];then
    _install_package zip unzip
fi

_check_remote

cd $SERVER_PATH
{{.config.AfterUploadScript}}