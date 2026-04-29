# Rttys 架构说明

## 数据存储方式

### 持久化存储（SQLite）

**设备历史记录** 保存到 `db/rttys.db`（相对于可执行文件目录）

| 表名 | 用途 | 说明 |
|-----|------|------|
| `device_history` | 设备上下线历史 | 记录设备 ID、组、描述、IP、协议版本、上线/下线时间、在线时长 |

**表结构：**
```sql
CREATE TABLE device_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    group_name TEXT DEFAULT '',
    description TEXT DEFAULT '',
    ip_addr TEXT DEFAULT '',
    proto INTEGER DEFAULT 0,
    online_time DATETIME NOT NULL,
    offline_time DATETIME,
    duration INTEGER  -- 秒数
);
```

### 内存存储

| 数据类型 | 存储方式 | 位置 |
|---------|---------|------|
| 设备组 (groups) | `sync.Map` | `server.go:19` |
| 设备列表 (devices) | `sync.Map` | `server.go:25`, `device.go:50-53` |
| 用户会话 (sessions) | `cache.NewMemCache` | `api.go:41` |
| HTTP 代理会话 | `sync.Map` | `http.go:44` |
| 用户连接 (users) | `sync.Map` | `device.go:50` |

### 特点

- **设备历史持久化**：设备上下线记录永久保存，可通过 API 查询
- **实时数据内存存储**：当前在线设备、会话等运行时数据存在内存中
- **并发安全**：使用 `sync.Map` 和 `sync.RWMutex` 保证线程安全
- **自动清理**：
  - HTTP 代理会话：15 分钟过期 (`http.go:46`)
  - 用户会话：30 分钟过期 (`api.go:24`)

## 前端构建流程

### 目录结构

```
rttys/
  ui/              # 前端源码 (Vue + Vite)
    dist/          # 构建输出目录
  assets/          # Go embed 嵌入的静态资源
    dist/          # 前端构建产物 (复制自 ui/dist)
      index.html
      favicon.ico
      assets/      # JS/CSS 文件
    http-proxy-err.html
  embed.go         # Go embed 配置
  api.go           # HTTP 服务器 (提供静态文件)
```

### 构建步骤

```bash
# 1. 构建前端
cd ui
pnpm install
pnpm build

# 2. 复制前端产物到 assets 目录
cd ..
# 确保 ui/dist 的内容复制到 assets/dist/

# 3. 编译 Go
go build .

# 4. 运行
./rttys
```

### 关键代码

**embed.go:10**
```go
//go:embed assets
var staticFs embed.FS
```

**api.go:33**
```go
fs, err := fs.Sub(staticFs, "assets/dist")
```

## 网络端口

| 端口 | 用途 | 说明 |
|-----|------|------|
| 5912 | 设备监听 | 设备端连接到此端口 |
| 5913 | 用户监听 | Web 管理页面 |
| 动态 | HTTP 代理 | 用于设备 HTTP 代理功能 |

## 设备连接流程

1. 设备 TCP 连接到 5912 端口
2. 5 秒内发送注册消息 (`proto.MsgTypeRegister`)
3. 服务器验证 token、调用 hook URL
4. 注册成功后保持心跳连接
5. 超时或断开时清理设备信息

## API 路由

**无需认证:**
- `POST /signin` - 登录
- `GET /alive` - 检查会话状态

**需要认证:**
- `GET /connect/:devid` - 连接设备 (WebSocket)
- `GET /counts` - 获取设备数量
- `GET /groups` - 获取设备组列表
- `GET /devs` - 获取设备列表
- `GET /dev/:devid` - 获取单个设备信息
- `GET /history/:devid` - 查询设备历史上下线记录
- `GET /history` - 查询所有设备历史（当前在线）
  - 参数：`group` - 按组筛选，`limit` - 限制返回数量
- `POST /cmd/:devid` - 发送命令到设备
- `ANY /web/:devid/:proto/:addr/*path` - HTTP 代理
- `ANY /web2/:group/:devid/:proto/:addr/*path` - HTTP 代理 (带组)
- `GET /signout` - 登出
