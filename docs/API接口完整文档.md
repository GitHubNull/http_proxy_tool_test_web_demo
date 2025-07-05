# HTTP/WebSocket代理测试工具 - API接口完整文档

## 📋 概述

本文档详细描述了HTTP/WebSocket代理测试工具提供的所有API接口。项目采用模块化架构，共提供60+个API接口，按功能分为5大类：

## 🏗️ 模块化架构

### 路由管理器
- **文件**: `routes/router.go`
- **功能**: 类似Python Flask蓝图的路由注册机制
- **接口**: `RouteModule` - 统一的模块实现标准

### 公共类型
- **文件**: `routes/types.go`
- **功能**: 定义所有模块共享的数据类型和结构

## 📚 API接口分类

### 1. 基础API模块 (`routes/api/basic.go`) - 8个接口

#### 1.1 通用测试接口
```
GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS /api/test
```
**功能**: 支持所有HTTP方法的通用测试接口
**参数**: 无特殊参数，支持任意请求头和请求体
**响应**: 
```json
{
  "code": 200,
  "message": "请求成功",
  "data": {
    "method": "GET",
    "url": "/api/test",
    "headers": {...},
    "query": {...},
    "body": "..."
  },
  "timestamp": 1642435200
}
```

#### 1.2 HTTP状态码测试
```
GET/POST /api/status/{code}
```
**功能**: 返回指定的HTTP状态码
**参数**: 
- `{code}`: HTTP状态码 (100-599)
**响应**: 返回指定状态码和相应的状态消息

#### 1.3 延迟测试
```
GET/POST /api/delay/{seconds}
```
**功能**: 模拟网络延迟
**参数**: 
- `{seconds}`: 延迟秒数 (1-30)
**响应**: 延迟指定时间后返回成功响应

#### 1.4 重定向测试
```
GET /api/redirect/{times}
```
**功能**: 多次重定向测试
**参数**: 
- `{times}`: 重定向次数 (1-10)
**响应**: 根据次数进行多级重定向

#### 1.5 重定向到指定URL
```
GET /api/redirect-to?url=<target>
```
**功能**: 重定向到指定URL
**参数**: 
- `url`: 目标URL (query参数)
**响应**: 302重定向到指定地址

#### 1.6 错误模拟
```
GET /api/error
```
**功能**: 模拟服务器内部错误
**响应**: 500错误状态码

#### 1.7 超时测试
```
GET /api/timeout
```
**功能**: 长时间无响应测试
**响应**: 60秒后超时

### 2. 格式处理模块 (`routes/format/formats.go`) - 14个接口

#### 2.1 基础格式响应接口

##### JSON格式响应
```
GET /api/json
```
**功能**: 返回标准JSON格式数据
**响应**: 
```json
{
  "code": 200,
  "data": {
    "string": "Hello World",
    "number": 12345,
    "boolean": true,
    "array": [1, 2, 3],
    "object": {"key": "value"}
  }
}
```

##### XML格式响应
```
GET /api/xml
```
**功能**: 返回XML格式数据
**响应**: 
```xml
<?xml version="1.0" encoding="UTF-8"?>
<response>
    <code>200</code>
    <message>XML响应成功</message>
    <data>
        <item>Item 1</item>
        <item>Item 2</item>
    </data>
</response>
```

##### HTML格式响应
```
GET /api/html
```
**功能**: 返回HTML格式页面
**响应**: 完整的HTML文档

##### 文本格式响应
```
GET /api/text
```
**功能**: 返回纯文本数据
**响应**: `text/plain` 格式的文本内容

##### 二进制格式响应
```
GET /api/binary
```
**功能**: 返回二进制数据
**响应**: `application/octet-stream` 格式的二进制数据

#### 2.2 数据解析接口

##### JSON数据解析
```
POST /api/parse/json
Content-Type: application/json
```
**功能**: 解析和验证JSON格式数据
**请求体**: 任意JSON数据
**响应**: 解析结果和数据结构信息

##### XML数据解析
```
POST /api/parse/xml
Content-Type: application/xml
```
**功能**: 解析和验证XML格式数据
**请求体**: 任意XML数据
**响应**: 解析结果和XML结构信息

##### Multipart数据解析
```
POST /api/parse/multipart
Content-Type: multipart/form-data
```
**功能**: 解析multipart表单数据
**请求体**: multipart/form-data格式
**响应**: 解析的表单字段和文件信息

##### 二进制数据解析
```
POST /api/parse/binary
Content-Type: application/octet-stream
```
**功能**: 处理原始二进制数据
**请求体**: 任意二进制数据
**响应**: 数据大小和类型信息

#### 2.3 复杂格式处理接口

##### 复杂嵌套JSON处理
```
POST /api/complex/nested-json
```
**功能**: 处理深层嵌套的JSON结构
**请求体**: 复杂嵌套JSON数据
**响应**: 详细的结构分析结果

##### 大型XML处理
```
POST /api/complex/large-xml
```
**功能**: 处理大型XML文档
**请求体**: 大型XML数据（支持MB级别）
**响应**: XML文档分析结果

##### 混合Multipart处理
```
POST /api/complex/mixed-multipart
```
**功能**: 处理混合类型的multipart数据
**请求体**: 包含文件和表单的混合multipart数据
**响应**: 详细的数据分析结果

#### 2.4 压缩和编码接口

##### Gzip压缩响应
```
GET /api/gzip
```
**功能**: 返回Gzip压缩的响应数据
**响应头**: `Content-Encoding: gzip`

##### Deflate压缩响应
```
GET /api/deflate
```
**功能**: 返回Deflate压缩的响应数据
**响应头**: `Content-Encoding: deflate`

##### Base64编码数据
```
GET /api/base64
```
**功能**: 返回Base64编码的数据
**响应**: Base64编码格式的数据

### 3. 传输协议模块 (`routes/transfer/chunked.go`) - 12个接口

#### 3.1 分块传输编码接口

##### 分块数据发送
```
GET /api/transfer/chunked?chunks=<n>&delay=<ms>
```
**功能**: 以分块方式发送数据
**参数**:
- `chunks`: 分块数量 (1-10，默认3)
- `delay`: 每块间隔毫秒数 (0-1000，默认100)
**响应头**: `Transfer-Encoding: chunked`
**响应**: 分块传输的数据

##### 分块数据接收
```
POST /api/transfer/chunked
Transfer-Encoding: chunked
```
**功能**: 接收分块传输的请求数据
**请求体**: 分块传输格式的数据
**响应**: 接收到的数据分析结果

##### 流式分块传输
```
GET /api/transfer/chunked/stream?duration=<s>&interval=<ms>
```
**功能**: 实时流式分块数据传输
**参数**:
- `duration`: 传输持续时间秒数 (1-60，默认10)
- `interval`: 数据发送间隔毫秒数 (100-5000，默认1000)
**响应**: 持续的流式分块数据

##### 分块文件上传
```
POST /api/transfer/chunked/upload
Transfer-Encoding: chunked
```
**功能**: 支持大文件的分块上传
**请求体**: 分块传输的文件数据
**响应**: 文件上传处理结果

#### 3.2 传输编码接口

##### Identity传输
```
GET /api/transfer/identity
```
**功能**: 标准无压缩传输
**响应头**: `Transfer-Encoding: identity`

##### Deflate压缩传输
```
GET /api/transfer/deflate
```
**功能**: 使用Deflate算法压缩传输
**响应头**: `Content-Encoding: deflate`

##### Gzip压缩传输
```
GET /api/transfer/gzip
```
**功能**: 使用Gzip算法压缩传输
**响应头**: `Content-Encoding: gzip`

#### 3.3 大文件传输接口

##### 大文件生成
```
GET /api/transfer/large/{size}?format=<type>
```
**功能**: 生成指定大小的文件进行传输
**参数**:
- `{size}`: 文件大小，支持单位 (1K, 1M, 1G)
- `format`: 文件格式 (binary/text，默认binary)
**响应**: 指定大小的文件数据

##### 大文件接收
```
POST /api/transfer/large
```
**功能**: 接收和处理大文件上传
**请求体**: 大文件数据
**响应**: 文件接收处理结果

#### 3.4 流式传输接口

##### SSE流传输
```
GET /api/transfer/stream/sse?count=<n>&interval=<ms>
```
**功能**: Server-Sent Events实时推送
**参数**:
- `count`: 事件数量 (1-100，默认10)
- `interval`: 事件间隔毫秒数 (100-5000，默认1000)
**响应头**: `Content-Type: text/event-stream`
**响应**: SSE格式的实时事件流

##### WebSocket流传输
```
GET /api/transfer/stream/websocket
```
**功能**: WebSocket协议传输
**响应**: 升级为WebSocket连接

##### 流数据响应
```
GET /api/stream/{lines}?delay=<ms>
```
**功能**: 分行流式数据
**参数**:
- `{lines}`: 行数 (1-1000)
- `delay`: 每行间隔毫秒数 (0-1000，默认100)
**响应**: 逐行流式数据

### 4. 性能测试模块 (`routes/test/performance/concurrent.go`) - 8个接口

#### 4.1 并发测试
```
GET/POST /test/concurrent?concurrency=<n>&requests=<m>
```
**功能**: 并发请求测试
**参数**:
- `concurrency`: 并发数 (1-1000，默认10)
- `requests`: 总请求数 (1-10000，默认100)
**响应**: 并发测试结果和统计信息

#### 4.2 压力测试
```
GET/POST /test/stress?duration=<s>&concurrency=<n>
```
**功能**: 持续压力测试
**参数**:
- `duration`: 持续时间秒数 (1-300，默认30)
- `concurrency`: 并发数 (1-500，默认50)
**响应**: 压力测试结果和性能指标

#### 4.3 批量测试
```
POST /test/batch
Content-Type: application/json
```
**功能**: 批量请求测试
**请求体**: 
```json
{
  "endpoints": [
    {"method": "GET", "url": "/api/test"},
    {"method": "POST", "url": "/api/json"}
  ],
  "concurrent": true
}
```
**响应**: 批量测试执行结果

#### 4.4 负载测试
```
GET /test/load?workers=<n>&duration=<s>
```
**功能**: 负载均衡测试
**参数**:
- `workers`: 工作线程数 (1-100，默认10)
- `duration`: 测试持续时间 (1-120，默认60)
**响应**: 负载测试结果

#### 4.5 随机延迟测试
```
GET /test/random-delay?min=<ms>&max=<ms>&requests=<n>
```
**功能**: 随机延迟测试
**参数**:
- `min`: 最小延迟毫秒数 (0-1000，默认100)
- `max`: 最大延迟毫秒数 (100-5000，默认1000)
- `requests`: 请求数量 (1-100，默认10)
**响应**: 随机延迟测试结果

#### 4.6 测试统计
```
GET /test/stats
```
**功能**: 获取测试统计信息
**响应**: 详细的测试统计数据

#### 4.7 重置统计
```
POST /test/reset
```
**功能**: 重置所有测试统计信息
**响应**: 重置操作结果

#### 4.8 Keep-Alive测试
```
GET /test/keepalive?connections=<n>&duration=<s>
```
**功能**: 长连接测试
**参数**:
- `connections`: 连接数量 (1-100，默认5)
- `duration`: 保持时间秒数 (1-300，默认30)
**响应**: 长连接测试结果

### 5. 系统资源模块 (`routes/test/system/resources.go`) - 7个接口

#### 5.1 系统信息
```
GET /test/system
```
**功能**: 获取系统详细信息
**响应**: 
```json
{
  "code": 200,
  "data": {
    "os": "linux",
    "arch": "amd64",
    "go_version": "go1.23.0",
    "cpu_cores": 8,
    "memory": {
      "total": "16GB",
      "available": "8GB",
      "usage_percent": 50.0
    },
    "disk": {...},
    "network": {...}
  }
}
```

#### 5.2 内存测试
```
GET /test/memory?size=<mb>&duration=<s>
```
**功能**: 内存分配和使用测试
**参数**:
- `size`: 分配内存大小MB (1-1024，默认100)
- `duration`: 保持时间秒数 (1-120，默认10)
**响应**: 内存测试结果和使用情况

#### 5.3 CPU测试
```
GET /test/cpu?intensity=<level>&duration=<s>
```
**功能**: CPU密集型计算测试
**参数**:
- `intensity`: 强度等级 (1-10，默认5)
- `duration`: 测试持续时间秒数 (1-60，默认10)
**响应**: CPU测试结果和性能指标

#### 5.4 网络测试
```
GET /test/network?size=<kb>&speed=<kbps>
```
**功能**: 网络性能测试
**参数**:
- `size`: 传输数据大小KB (1-10240，默认1024)
- `speed`: 限制速度kbps (可选)
**响应**: 网络测试结果和带宽信息

#### 5.5 文件IO测试
```
GET /test/fileio?operations=<n>&size=<kb>
```
**功能**: 文件读写性能测试
**参数**:
- `operations`: 操作次数 (1-1000，默认100)
- `size`: 每次操作数据大小KB (1-1024，默认64)
**响应**: 文件IO测试结果

#### 5.6 数据库测试
```
GET /test/database?operations=<n>&concurrent=<c>
```
**功能**: 模拟数据库操作
**参数**:
- `operations`: 操作次数 (1-10000，默认1000)
- `concurrent`: 并发连接数 (1-100，默认10)
**响应**: 数据库操作测试结果

#### 5.7 字节数据传输
```
GET/POST /api/bytes/{size}?format=<type>
```
**功能**: 指定大小数据传输测试
**参数**:
- `{size}`: 数据大小，支持单位 (1B, 1K, 1M)
- `format`: 数据格式 (random/zero/pattern，默认random)
**响应**: 指定大小的数据

### 6. WebSocket接口 - 8个接口

#### 6.1 基础连接测试
```
WebSocket: /ws/connect
```
**功能**: WebSocket连接测试
**消息格式**: JSON
**支持操作**: 连接、断开、状态查询

#### 6.2 回声测试
```
WebSocket: /ws/echo
```
**功能**: WebSocket回声测试
**消息格式**: 任意格式
**行为**: 服务器回传收到的消息

#### 6.3 广播测试
```
WebSocket: /ws/broadcast
```
**功能**: WebSocket广播测试
**消息格式**: JSON
**行为**: 消息广播到所有连接的客户端

#### 6.4 实时数据推送
```
WebSocket: /ws/realtime
```
**功能**: 实时数据推送
**消息格式**: JSON
**行为**: 定期推送实时数据

#### 6.5 心跳检测
```
WebSocket: /ws/heartbeat
```
**功能**: WebSocket心跳机制
**消息格式**: ping/pong
**行为**: 自动心跳检测和响应

#### 6.6 二进制传输
```
WebSocket: /ws/binary
```
**功能**: 二进制数据传输
**消息格式**: 二进制
**行为**: 处理二进制数据传输

#### 6.7 聊天测试
```
WebSocket: /ws/chat
```
**功能**: 聊天室模拟
**消息格式**: JSON
**行为**: 多用户聊天功能

#### 6.8 性能测试
```
WebSocket: /ws/performance
```
**功能**: WebSocket性能测试
**消息格式**: JSON
**行为**: 高频消息传输性能测试

## 📊 统一响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "请求成功",
  "data": {...},
  "timestamp": 1642435200,
  "request_id": "req_1642435200123456789"
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "详细错误信息",
  "timestamp": 1642435200,
  "request_id": "req_1642435200123456789"
}
```

## 🔧 使用示例

### cURL示例

#### 基础测试
```bash
# GET请求测试
curl -X GET http://localhost:8080/api/test

# POST请求测试
curl -X POST http://localhost:8080/api/test \
  -H "Content-Type: application/json" \
  -d '{"test": "data"}'
```

#### 分块传输测试
```bash
# 分块传输响应
curl -v http://localhost:8080/api/transfer/chunked?chunks=5&delay=200

# 分块传输请求
curl -X POST http://localhost:8080/api/transfer/chunked \
  -H "Transfer-Encoding: chunked" \
  --data-binary @largefile.dat
```

#### JSON数据解析
```bash
curl -X POST http://localhost:8080/api/parse/json \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test",
    "data": {
      "nested": {
        "value": 123
      }
    }
  }'
```

### JavaScript示例

#### 基础AJAX请求
```javascript
// 使用fetch API
fetch('/api/test', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({test: 'data'})
})
.then(response => response.json())
.then(data => console.log(data));
```

#### WebSocket连接
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/echo');

ws.onopen = function() {
  console.log('WebSocket连接已建立');
  ws.send('Hello Server');
};

ws.onmessage = function(event) {
  console.log('收到消息:', event.data);
};

ws.onclose = function() {
  console.log('WebSocket连接已关闭');
};
```

## 📈 性能指标

### 接口性能
- **平均响应时间**: < 10ms
- **最大并发连接**: 1000+
- **最大QPS**: 10,000+
- **内存使用**: 运行时约50MB

### 限制说明
- **请求体大小**: 最大100MB
- **WebSocket连接**: 最大1000个并发连接
- **文件上传**: 最大1GB
- **请求超时**: 默认30秒

## 🛡️ 安全说明

### 输入验证
- 所有参数都进行严格验证
- 防止SQL注入和XSS攻击
- 文件上传类型和大小限制

### 访问控制
- 本工具仅用于测试环境
- 建议不要在生产环境中使用
- 可配置IP白名单限制访问

### 数据保护
- 不保存敏感数据
- 临时文件自动清理
- 内存数据及时释放

## 📞 技术支持

如有问题或建议，请通过以下方式联系：
- GitHub Issues: [项目地址]/issues
- 文档网站: [在线文档地址]
- 用户手册: `docs/用户使用手册.md` 