# Muxi_ClassList(课表服务)

## 一、如何运行？

### 1、配置信息

将`configs/config-example.yaml`换成`configs/config.yaml`,并填充配置文件
### 2、构建镜像
在`DockerFile`所在目录下使用命令`docker build -t muxi_classlist:v1`构建镜像
### 3、运行
在`deploy`下执行`docker-compose up -d`即可



## 二、错误码

| 错误码 | 含义                         |
| ------ | ---------------------------- |
| 200    | 成功/课程信息未找到          |
| 450    | 数据库查找课程失败           |
| 300    | 课程更新失败                 |
| 301    | 入参错误                     |
| 302    | 课程保存失败                 |
| 303    | 课程删除失败                 |
| 304    | 爬取课表失败                 |
| 305    | 请求ccnu一站式登录服务错误   |
| 306    | 学号与课程ID的对应关系未找到 |

## 三、API文档

将文件中`openapi.yaml`导入到`apifox`中即可