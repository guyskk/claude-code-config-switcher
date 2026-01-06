# Proposal: add-central-server

## 概述

构建中心服务器和 Web 管理界面，实现企业级 AI 员工管理系统。

## 动机

单机版的 ccc 虽然强大，但存在以下限制：
1. 每台机器需要单独配置
2. 无法统一管理多个 agent
3. 无法协作和共享
4. 无法监控和审计

中心服务器可以实现：

1. **统一管理**: 一个平台管理所有 agent
2. **岗位模板**: 批量创建相同配置的 AI 员工
3. **协作**: 多人协作和任务分配
4. **监控**: 实时状态和性能监控
5. **安全**: 统一认证和权限控制

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        中心服务器 (Cloud/VPS)                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ Web UI   │  │  API     │  │  WebSocket│  │  数据库   │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  岗位管理 (Post) 系统管理              │   │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐    │   │
│  │  │岗位1   │  │岗位2   │  │岗位3   │  │岗位N   │    │   │
│  │  │DevOps  │  │前端组  │  │后端组  │  │...     │    │   │
│  │  └────────┘  └────────┘  └────────┘  └────────┘    │   │
│  │      ↓            ↓            ↓            ↓          │   │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐    │   │
│  │  │AI员工1 │  │AI员工2 │  │AI员工3 │  │AI员工N │    │   │
│  │  │(复制)  │  │(复制)  │  │(复制)  │  │(复制)  │    │   │
│  │  └────────┘  └────────┘  └────────┘  └────────┘    │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                            ↕ HTTPS + WebSocket
┌─────────────────────────────────────────────────────────────────┐
│                    设备 A (开发机)                                  │
│  ccc 程序 ←→ 连接中心服务器                                        │
│  - 接收配置                                                     │
│  - 上报状态                                                     │
│  - 执行任务                                                     │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    设备 B (生产服务器)                              │
│  ccc 程序 ←→ 连接中心服务器                                        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    用户 (Web 浏览器)                              │
│  访问中心服务器 Web UI                                            │
│  - 创建岗位                                                     │
│  - 批量创建员工                                                 │
│  - 分配任务                                                     │
└─────────────────────────────────────────────────────────────────┘
```

## 核心概念

### 1. 岗位 (Post)

岗位是 AI 员工的模板，定义了：
- **名称**: DevOps 工程师、前端开发工程师等
- **技能集**: 可用的工具和权限
- **提供商配置**: 默认使用的模型提供商
- **系统提示词**: 角色定位和行为准则
- **资源限制**: 内存、CPU、超时等

### 2. AI 员工 (Agent)

AI 员工是岗位的实例化：
- **所属岗位**: 从岗位继承配置
- **分配设备**: 运行在哪个设备上
- **当前状态**: 空闲、忙碌、离线
- **任务队列**: 待处理任务
- **会话历史**: 工作记录

### 3. 任务 (Task)

分配给 AI 员工的工作：
- **类型**: 开发、测试、部署、监控等
- **优先级**: 紧急、高、中、低
- **状态**: 待处理、进行中、已完成、失败
- **协作**: 可以拉群，多个 AI 员工协作

## 技术实现

### 1. 中心服务器 (Go)

```go
// internal/central/server.go
type CentralServer struct {
    config      *Config
    db          *Database
    hub         *ConnectionHub
    agentMgr    *RemoteAgentManager
    postMgr     *PostManager
    taskMgr     *TaskManager
    authService *AuthService
}

type RemoteAgent struct {
    ID          string
    DeviceID    string
    PostID      string
    Status      AgentStatus
    LastSeen    time.Time
    Config      AgentConfig
}

type Post struct {
    ID          string
    Name        string
    Description string
    Template    AgentConfig
    CreatedBy   string
    CreatedAt   time.Time
}

type Task struct {
    ID          string
    PostID      string
    AgentID     string
    Prompt      string
    Status      TaskStatus
    Priority    int
    CreatedBy   string
    AssignedAt  time.Time
}
```

### 2. 设备连接协议

```go
// 设备注册
type DeviceRegisterRequest struct {
    Token     string // 连接 token
    DeviceID  string // 设备唯一标识
    Hostname  string
    IP        string
    OS        string
    Capabilities []string // ["claude", "opencode", "codex"]
}

// 心跳
type Heartbeat struct {
    DeviceID  string
    Timestamp time.Time
    Agents    []AgentStatus
}

// 任务分配
type TaskAssignment struct {
    TaskID    string
    AgentID   string
    Prompt    string
    Options   RunOptions
}
```

### 3. 数据库模型

```sql
-- 设备表
CREATE TABLE devices (
    id UUID PRIMARY KEY,
    device_id VARCHAR(255) UNIQUE NOT NULL,
    hostname VARCHAR(255),
    ip_address INET,
    os VARCHAR(255),
    status VARCHAR(50),
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 岗位表
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template JSONB,
    created_by UUID,
    created_at TIMESTAMP DEFAULT NOW()
);

-- AI 员工表
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    device_id UUID REFERENCES devices(id),
    post_id UUID REFERENCES posts(id),
    name VARCHAR(255),
    status VARCHAR(50),
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 任务表
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    post_id UUID REFERENCES posts(id),
    agent_id UUID REFERENCES agents(id),
    prompt TEXT,
    status VARCHAR(50),
    priority INT,
    result JSONB,
    created_by UUID,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 会话表
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    agent_id UUID REFERENCES agents(id),
    messages JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4. Web API

```
POST   /api/v1/devices/register          设备注册
GET    /api/v1/devices                   列出设备
GET    /api/v1/devices/{id}              设备详情

POST   /api/v1/posts                     创建岗位
GET    /api/v1/posts                     列出岗位
POST   /api/v1/posts/{id}/agents          批量创建员工
GET    /api/v1/posts/{id}                岗位详情

POST   /api/v1/agents/{id}/tasks          分配任务
GET    /api/v1/agents/{id}/tasks          任务列表
GET    /api/v1/agents/{id}/sessions       会话历史

POST   /api/v1/tasks                     创建任务（自动分配）
POST   /api/v1/tasks/{id}/group          创建群组任务

WS     /api/v1/agents/{id}/ws            实时通信
```

## 部署流程

### 1. 设备端 (ccc)

```bash
# 生成连接 token
ccc generate-token

# 连接到中心服务器
export CCC_CENTRAL_SERVER=wss://central.example.com
ccc connect --token=xxx

# 后台运行作为服务
ccc daemon
```

### 2. 中心服务器

```bash
# 启动服务器
central-server serve --listen-addr=:443 --tls-cert=cert.pem --tls-key=key.pem
```

## 安全考虑

1. **认证**: JWT token 认证
2. **授权**: 基于角色的访问控制 (RBAC)
3. **加密**: TLS 加密通信
4. **审计**: 完整的操作日志

## 实施计划

### Phase 1: 基础设施
1. 数据库设计和迁移
2. 基础 API 框架
3. 设备注册和心跳

### Phase 2: 岗位和员工管理
1. 岗位 CRUD
2. 员工 CRUD
3. 任务分配

### Phase 3: Web UI
1. 岗位管理界面
2. 员工监控界面
3. 任务管理界面

### Phase 4: 高级功能
1. 群组协作
2. 任务调度
3. 报表和分析

## 风险

| 风险 | 影响 | 缓解 |
|------|------|------|
| 复杂度高 | 高 | 分阶段实施 |
| 网络依赖 | 中 | 本地缓存降级 |
| 安全风险 | 高 | 完整的安全机制 |
| 性能问题 | 中 | 负载均衡和缓存 |

## 开放问题

1. 如何处理设备离线情况？
2. 如何实现任务优先级调度？
3. 如何支持跨设备任务迁移？
4. 如何实现 AI 员工的负载均衡？
