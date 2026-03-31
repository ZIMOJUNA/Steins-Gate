# -
这里是《创世纪》项目组


MongoDB Compass（官方免费，图形界面）
1. 下载安装
   https://www.mongodb.com/products/compass
2. 连接你的数据库
   填入： mongodb://admin:123456@localhost:27017/


### 后台启动（-d 表示守护进程）
docker-compose up -d

### 查看服务运行状态
docker-compose ps

### 停止服务（不删除数据）
docker-compose stop

### 停止并删除容器（数据卷保留）
docker-compose down

### 查看日志
docker-compose logs -f  # 实时日志
docker-compose logs redis  # 只看 Redis 日志
docker-compose logs mongodb  # 只看 MongoDB 日志

