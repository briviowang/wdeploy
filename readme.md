# 已实现功能

  用法：
  aliyun 显示阿里云服务器配置信息
  api 生成api；参数module：区分不同分组
  db-check 检查数据库表
  db-compare 比较数据库,后可跟参数：server(默认同步远程)；local(同步本地)
  db-doc 生成文档
  db-info 获取数据库配置信息
  db-statics 统计
  doc 生成文档
  example-config 显示配置文件样例
  execute 远程执行一些操作,lnmp、update_mysql_conf
  help 显示帮助
  install 安装windows环境
  last-files 显示将要上传的文件
  last-upload 上传最近修改的,可传参数all:重新全部上传
  reset-upload 重制上传时间
  upload 上传文件到远程服务器，默认只上传代码
  upload-all 上传全部代码、图片到远程服务器,注意：该命令会预先把远程_core、app、static目录删除，再解压上传的代码
  xcx-upload 上传小程序
  zip 打包最近修改的,可传参数all