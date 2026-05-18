# Steins Gate 云存档后端

这是一个游戏云存档后端框架，当前提供：

- 邮箱验证码注册
- 邮箱验证码修改密码
- 邮箱密码登录
- Redis 登录 Token
- MySQL 玩家账号表与玩家数据表
- 登录后获取、创建/覆盖、修改、删除玩家存档
- 邮件发送接口抽象，当前支持 console 和 SMTP，预留阿里云邮件实现入口

## 技术栈

- Go + Fiber v3
- MySQL 8
- Redis 7
- SMTP 邮件发送

## 启动依赖

```bash
docker-compose up -d
```

默认 docker-compose 会创建：

- MySQL：`127.0.0.1:3306`
- 数据库：`game_db`
- 用户：`game_app`
- 密码：`123456`
- Redis：`127.0.0.1:6379`
- Redis 密码：`123456`

## 配置

配置文件：`config/config.yaml`

程序默认读取 `config/config.yaml`，也可以通过环境变量指定：

```bash
STEINS_GATE_CONFIG=/path/to/config.yaml go run .
```

邮件配置：

```yaml
mail:
  provider: "console"
  smtp:
    host: "smtp.example.com"
    port: 587
    username: "your-smtp-user"
    password: "your-smtp-password"
    from: "no-reply@example.com"
    from_name: "Steins Gate"
    use_tls: false
    start_tls: true
    skip_verify: false
```

`provider` 可选值：

- `console`：开发环境直接在服务日志打印验证码
- `smtp`：通过 SMTP 发送验证码邮件
- `aliyun`：预留，后续在 `mailer.AliyunSender` 中接阿里云邮件推送

验证码和 Token 配置：

```yaml
auth:
  token_ttl: "24h"
  email_code_ttl: "5m"
  email_code_resend_interval: "60s"
  email_code_send_limit_per_hour: 5
  email_code_max_verify_attempts: 5
  email_code_hash_secret: "change-me-in-production"
  password_min_length: 8
```

`email_code_hash_secret` 用于对 Redis 中的验证码做 HMAC 哈希，生产环境需要改成自己的随机密钥。

## 数据库表和初始化

服务启动时会自动补建缺失的业务表，不会删除已有数据。

当前表结构在 `dbschema/schema.sql` 中维护，现在只有两张业务表：

- `user_accounts`：用户账号表，保存邮箱、昵称、密码哈希、账号状态、登录次数、最后登录时间、创建时间、更新时间等
- `player_data`：玩家云存档表，保存游戏、槽位、JSON 存档数据、版本号、创建时间、更新时间等

`player_data.account_id` 外键关联 `user_accounts.id`。一个账号可以拥有多条存档数据，使用 `(account_id, game_key, slot_key)` 唯一约束实现同一游戏、同一槽位的覆盖保存。

如果需要按开发环境方式重置表，运行：

```bash
go run ./cmd/dbinit
```

这个命令默认会先删除旧表再创建新表，相当于：

```bash
go run ./cmd/dbinit -reset=true
```

只补建缺失表、不删除已有数据：

```bash
go run ./cmd/dbinit -reset=false
```

注意：`-reset=true` 会删除 `player_data`、`user_accounts`，也会清理旧版本遗留的 `user_infos` 和 `player_accounts`。

也可以在别的 MySQL 工具里直接执行 SQL 文件：

```bash
mysql -h 127.0.0.1 -P 3306 -u game_app -p game_db < dbschema/schema.sql
```

这个 SQL 文件包含删表和建表语句，直接完整执行会先删表再创建。服务启动时只会读取其中的建表分区，不会执行删表分区。

## 统一响应

成功：

```json
{
  "code": "ok",
  "message": "ok",
  "data": {}
}
```

失败：

```json
{
  "code": "invalid_input",
  "message": "请求参数不合法"
}
```

## 鉴权

登录、注册、改密成功后会返回 Token：

```json
{
  "access_token": "token",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

访问需要登录的接口时带上：

```http
Authorization: Bearer <access_token>
```

## API

基础路径：`/api/v1`

### 健康检查

```http
GET /health
```

### 发送邮箱验证码

```http
POST /api/v1/auth/email-code
Content-Type: application/json
```

```json
{
  "email": "player@example.com",
  "scene": "register"
}
```

`scene` 可选：

- `register`：注册验证码
- `reset_password`：修改密码验证码

限制：

- 默认 60 秒内同一邮箱同一场景只能发送一次
- 默认 1 小时最多发送 5 次
- 默认验证码 5 分钟过期
- 默认验证码最多错误 5 次，超过后需要重新获取

### 注册

注册前需要先调用发送验证码接口，`scene` 使用 `register`。

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "email": "player@example.com",
  "password": "password123",
  "nickname": "player-one",
  "code": "123456"
}
```

返回：

```json
{
  "code": "ok",
  "message": "ok",
  "data": {
    "account": {
      "id": 1,
      "email": "player@example.com",
      "nickname": "player-one",
      "status": 1,
      "created_at": "2026-05-18T19:00:00+08:00",
      "updated_at": "2026-05-18T19:00:00+08:00"
    },
    "token": {
      "access_token": "token",
      "token_type": "Bearer",
      "expires_in": 86400
    }
  }
}
```

### 登录

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "email": "player@example.com",
  "password": "password123"
}
```

返回账号信息和登录 Token。

### 修改密码

修改密码前需要先调用发送验证码接口，`scene` 使用 `reset_password`。

```http
POST /api/v1/auth/password/reset
Content-Type: application/json
```

```json
{
  "email": "player@example.com",
  "code": "123456",
  "new_password": "new-password123"
}
```

修改成功后会返回新的登录 Token。

### 获取当前账号

需要登录。

```http
GET /api/v1/me
Authorization: Bearer <access_token>
```

### 退出登录

需要登录。会删除当前 Token。

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

成功返回 `204 No Content`。

## 玩家存档 API

以下接口都需要登录。

### 存档列表

```http
GET /api/v1/player-data
Authorization: Bearer <access_token>
```

可选按游戏过滤：

```http
GET /api/v1/player-data?game_key=steins-gate
```

### 获取单个存档

```http
GET /api/v1/player-data/{id}
Authorization: Bearer <access_token>
```

### 创建或覆盖存档

同一个账号下，相同 `game_key + slot_key` 会覆盖原存档并让 `version + 1`。

```http
POST /api/v1/player-data
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "game_key": "steins-gate",
  "slot_key": "slot-1",
  "data": {
    "chapter": 3,
    "play_time": 7200,
    "items": ["phone", "badge"]
  }
}
```

### 按 ID 修改存档

```http
PUT /api/v1/player-data/{id}
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "data": {
    "chapter": 4,
    "play_time": 8800,
    "items": ["phone", "badge", "lab-note"]
  }
}
```

### 删除存档

```http
DELETE /api/v1/player-data/{id}
Authorization: Bearer <access_token>
```

成功返回 `204 No Content`。

## 本地运行

```bash
go run .
```

默认监听：

```text
http://127.0.0.1:8080
```

## 接口测试程序

先启动后端服务：

```bash
go run .
```

再开一个终端运行接口测试：

```bash
go run ./cmd/apitest -email apitest@example.com -password password123
```

默认邮件 provider 是 `console`，验证码会打印在后端服务日志里。测试程序会提示输入注册验证码：

```text
enter register email code:
```

如果账号已经存在，可以跳过注册直接测登录和存档接口：

```bash
go run ./cmd/apitest -skip-register=true -email apitest@example.com -password password123
```

## Docker 常用指令

```bash
docker-compose up -d
docker-compose ps
docker-compose stop
docker-compose down
docker-compose logs -f
docker-compose logs redis
docker-compose logs mysql
```
