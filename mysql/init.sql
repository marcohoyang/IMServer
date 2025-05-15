-- 确保使用 utf8mb4 字符集（支持 emoji 等特殊字符）
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库（若不存在）
CREATE DATABASE IF NOT EXISTS mydb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户（% 表示允许任意主机连接，生产环境建议限制 IP）
CREATE USER IF NOT EXISTS 'hoyang'@'%' IDENTIFIED BY '123456';

-- 授予权限（授予 mydb 数据库的所有权限）
GRANT ALL PRIVILEGES ON mydb.* TO 'hoyang'@'%' WITH GRANT OPTION;

-- 刷新权限
FLUSH PRIVILEGES;