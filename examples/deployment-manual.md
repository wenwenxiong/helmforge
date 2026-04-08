# 应用部署手册

## 应用名称
MyApp

## 应用版本
1.0.0

## 应用描述
这是一个示例应用，包含前端Web服务、后端API服务和数据库服务。

## 环境要求

### 硬件要求
- CPU: 2核及以上
- 内存: 4GB及以上
- 磁盘: 20GB及以上

### 软件要求
- Docker 20.10+  
- Docker Compose 1.29+
- PostgreSQL 13+
- Node.js 16+
- Nginx 1.20+

## 依赖组件

1. **前端Web服务**
   - 技术栈: Nginx
   - 端口: 80
   - 依赖: 无

2. **后端API服务**
   - 技术栈: Node.js
   - 端口: 3000
   - 依赖: PostgreSQL数据库

3. **数据库服务**
   - 技术栈: PostgreSQL
   - 端口: 5432
   - 依赖: 无

## 部署步骤

### 1. 准备环境
```bash
# 安装Docker和Docker Compose
sudo apt-get update
sudo apt-get install -y docker.io docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker
```

### 2. 配置数据库
```bash
# 创建数据库目录
mkdir -p postgres-data

# 启动PostgreSQL容器
docker run -d \
  --name postgres \
  -p 5432:5432 \
  -e POSTGRES_DB=app \
  -e POSTGRES_USER=app \
  -e POSTGRES_PASSWORD=secret \
  -v postgres-data:/var/lib/postgresql/data \
  postgres:13-alpine
```

### 3. 部署后端API服务
```bash
# 创建API目录
mkdir -p api

# 编写package.json
cat > api/package.json << 'EOF'
{
  "name": "api",
  "version": "1.0.0",
  "description": "Backend API service",
  "main": "index.js",
  "scripts": {
    "start": "node index.js"
  },
  "dependencies": {
    "express": "^4.17.1",
    "pg": "^8.7.1"
  }
}
EOF

# 编写index.js
cat > api/index.js << 'EOF'
const express = require('express');
const { Pool } = require('pg');

const app = express();
const port = process.env.PORT || 3000;

// 数据库连接
const pool = new Pool({
  host: process.env.DB_HOST || 'localhost',
  port: process.env.DB_PORT || 5432,
  database: process.env.DB_NAME || 'app',
  user: process.env.DB_USER || 'app',
  password: process.env.DB_PASSWORD || 'secret'
});

// 健康检查
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

// API端点
app.get('/api/data', async (req, res) => {
  try {
    const result = await pool.query('SELECT NOW() as timestamp');
    res.json({ data: result.rows[0] });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 启动服务
app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
EOF

# 启动API容器
docker run -d \
  --name api \
  -p 3000:3000 \
  -e NODE_ENV=production \
  -e PORT=3000 \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_NAME=app \
  -e DB_USER=app \
  -e DB_PASSWORD=secret \
  -v $(pwd)/api:/app \
  --working-dir /app \
  node:16-alpine \
  npm start
```

### 4. 部署前端Web服务
```bash
# 创建nginx配置文件
cat > nginx.conf << 'EOF'
events {
  worker_connections 1024;
}

http {
  server {
    listen 80;
    server_name localhost;

    location / {
      root /usr/share/nginx/html;
      index index.html;
    }

    location /api/ {
      proxy_pass http://localhost:3000/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
      return 200 '{"status":"ok"}';
      add_header Content-Type application/json;
    }
  }
}
EOF

# 创建index.html
mkdir -p html
cat > html/index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
  <title>MyApp</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      margin: 0;
      padding: 20px;
      background-color: #f0f0f0;
    }
    .container {
      max-width: 800px;
      margin: 0 auto;
      background-color: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    h1 {
      color: #333;
    }
    .status {
      padding: 10px;
      margin: 10px 0;
      border-radius: 4px;
    }
    .status.ok {
      background-color: #d4edda;
      color: #155724;
      border: 1px solid #c3e6cb;
    }
    .status.error {
      background-color: #f8d7da;
      color: #721c24;
      border: 1px solid #f5c6cb;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>MyApp</h1>
    <p>Welcome to MyApp!</p>
    
    <h2>Service Status</h2>
    <div id="health-status"></div>
    
    <h2>API Data</h2>
    <div id="api-data"></div>
  </div>
  
  <script>
    // 检查健康状态
    async function checkHealth() {
      try {
        const response = await fetch('/health');
        const data = await response.json();
        document.getElementById('health-status').innerHTML = 
          '<div class="status ok">Web Service: ' + data.status + '</div>';
      } catch (error) {
        document.getElementById('health-status').innerHTML = 
          '<div class="status error">Web Service: Error</div>';
      }
      
      try {
        const response = await fetch('/api/health');
        const data = await response.json();
        document.getElementById('health-status').innerHTML += 
          '<div class="status ok">API Service: ' + data.status + '</div>';
      } catch (error) {
        document.getElementById('health-status').innerHTML += 
          '<div class="status error">API Service: Error</div>';
      }
    }
    
    // 获取API数据
    async function fetchApiData() {
      try {
        const response = await fetch('/api/data');
        const data = await response.json();
        document.getElementById('api-data').innerHTML = 
          '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
      } catch (error) {
        document.getElementById('api-data').innerHTML = 
          '<div class="status error">Error fetching data: ' + error.message + '</div>';
      }
    }
    
    // 初始化
    checkHealth();
    fetchApiData();
  </script>
</body>
</html>
EOF

# 启动Nginx容器
docker run -d \
  --name web \
  -p 8080:80 \
  -v $(pwd)/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v $(pwd)/html:/usr/share/nginx/html:ro \
  nginx:latest
```

## 访问应用

应用启动后，可以通过以下地址访问：
- Web界面: http://localhost:8080
- API服务: http://localhost:3000/api/data
- 健康检查: http://localhost:8080/health

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查PostgreSQL容器是否正常运行
   - 检查数据库连接参数是否正确
   - 检查网络连接是否正常

2. **API服务启动失败**
   - 检查Node.js依赖是否安装完成
   - 检查数据库连接是否正常
   - 检查端口是否被占用

3. **Web服务无法访问**
   - 检查Nginx配置是否正确
   - 检查端口映射是否正确
   - 检查防火墙设置

### 日志查看
```bash
# 查看Web服务日志
docker logs web

# 查看API服务日志
docker logs api

# 查看数据库服务日志
docker logs postgres
```
