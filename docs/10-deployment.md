# GasTrack 线上部署指南

> **最后更新**: 2026-03-30

使用 Docker Compose 一键部署 GasTrack（PostgreSQL + Go 后端 + Nginx），默认启用 HTTPS。

---

## 1. 前置要求

| 项目 | 要求 |
|------|------|
| 服务器 | 1C1G 起步（Go 运行仅 ~30 MB），推荐 2C2G |
| 系统 | Ubuntu 22.04 / Debian 12 |
| 软件 | Docker + Docker Compose |
| 网络 | 公网 IP，域名 DNS A 记录指向服务器，防火墙开放 80/443 |

```bash
# 安装 Docker（如未安装）
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER && newgrp docker
```

---

## 2. 架构

```
用户 ──HTTPS──► Nginx (80/443)
                  ├─ /api/*  ──► Go 后端 (:8098) ──► PostgreSQL (:5432)
                  └─ /*      ──► 前端静态文件 (SPA)
```

三个容器均在同一 Docker 网络内通信，仅 Nginx 对外暴露 80/443。

---

## 3. 文件清单

| 文件 | 说明 |
|------|------|
| `server/Dockerfile` | 后端多阶段构建（最终 ~14 MB） |
| `Dockerfile.web` | 前端多阶段构建（最终 ~83 MB） |
| `nginx/default.conf` | HTTPS 配置（TLS 1.2/1.3 + HSTS + ACME + API 反代） |
| `nginx/default-http-only.conf` | 纯 HTTP 备用配置 |
| `docker-compose.prod.yaml` | 三服务编排 |
| `.env.production.example` | 环境变量模板 |
| `scripts/init-ssl.sh` | SSL 证书初始化脚本 |

---

## 4. 首次部署

### 4.1 配置环境变量

```bash
cp .env.production.example .env.production
```

编辑 `.env.production`：

```bash
DB_PASSWORD=<强密码>                       # 必填
JWT_SECRET=<openssl rand -hex 32 的输出>   # 必填
DOMAIN=gas.example.com                     # 必填，你的域名
EMAIL=you@email.com                        # 必填，Let's Encrypt 邮箱
CORS_ORIGINS=https://gas.example.com       # 与域名一致
```

### 4.2 构建镜像

```bash
docker compose -f docker-compose.prod.yaml --env-file .env.production build
```

首次构建需要拉取基础镜像 + 编译，预计 3-5 分钟。构建完成后产出两个镜像：

| 镜像 | 内容 | 大小 |
|------|------|------|
| `gastrack-backend` | Go 二进制 + alpine | ~14 MB |
| `gastrack-nginx` | 前端 dist + Nginx | ~83 MB |

> 也可以在本地/CI 构建后 `docker save` 打包上传到服务器再 `docker load`，跳过服务器端编译。

### 4.3 启动（含 HTTPS 证书申请）

```bash
chmod +x scripts/init-ssl.sh
DOMAIN=gas.example.com EMAIL=you@email.com ./scripts/init-ssl.sh
```

脚本会自动：
1. 生成临时自签名证书 → 启动所有容器（PostgreSQL → 后端 → Nginx）
2. 用 Certbot webroot 模式申请 Let's Encrypt 真实证书
3. 替换证书并重载 Nginx

> 如果已经在 4.2 构建过镜像，脚本启动时会直接使用已有镜像，不再重复构建。

### 4.4 验证

```bash
curl https://gas.example.com/api/v1/health
# → {"code":0,"message":"success","data":{"status":"ok"}}
```

---

## 5. 不使用 HTTPS

外层有 CDN（Cloudflare）/ 负载均衡做 SSL 终止时：

```bash
cp nginx/default-http-only.conf nginx/default.conf
docker compose -f docker-compose.prod.yaml --env-file .env.production up -d
```

---

## 6. 日常运维

### 6.1 更新代码

```bash
git pull
docker compose -f docker-compose.prod.yaml --env-file .env.production up -d --build

# 仅更新后端 / 仅更新前端
docker compose -f docker-compose.prod.yaml --env-file .env.production up -d --build backend
docker compose -f docker-compose.prod.yaml --env-file .env.production up -d --build nginx
```

### 6.2 查看日志

```bash
# 所有服务
docker compose -f docker-compose.prod.yaml logs -f

# 单个服务
docker logs -f gastrack-backend
docker logs -f gastrack-nginx
```

### 6.3 数据库备份

```bash
# 手动备份
docker exec gastrack-postgres pg_dump -U gastrack -d gastrack | gzip > backup_$(date +%Y%m%d).sql.gz

# 恢复
gunzip -c backup_20260330.sql.gz | docker exec -i gastrack-postgres psql -U gastrack -d gastrack
```

自动备份（crontab）：

```bash
# 每天凌晨 3 点备份，保留 30 天
(crontab -l 2>/dev/null; echo '0 3 * * * docker exec gastrack-postgres pg_dump -U gastrack -d gastrack | gzip > /opt/gastrack/backups/gastrack_$(date +\%Y\%m\%d).sql.gz && find /opt/gastrack/backups -name "*.sql.gz" -mtime +30 -delete') | crontab -
```

### 6.4 SSL 证书续期

```bash
# 手动续期
docker run --rm \
  -v certbot-etc:/etc/letsencrypt \
  -v certbot-var:/var/lib/letsencrypt \
  -v $(pwd)/certbot/www:/var/www/certbot \
  certbot/certbot renew --quiet
cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ./ssl/
cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ./ssl/
docker exec gastrack-nginx nginx -s reload

# 自动续期（加入 crontab，每天 4 点）
(crontab -l 2>/dev/null; echo '0 4 * * * cd /opt/gastrack && docker run --rm -v certbot-etc:/etc/letsencrypt -v certbot-var:/var/lib/letsencrypt -v $(pwd)/certbot/www:/var/www/certbot certbot/certbot renew --quiet && cp /etc/letsencrypt/live/$(grep DOMAIN .env.production | cut -d= -f2)/fullchain.pem ./ssl/ && cp /etc/letsencrypt/live/$(grep DOMAIN .env.production | cut -d= -f2)/privkey.pem ./ssl/ && docker exec gastrack-nginx nginx -s reload') | crontab -
```

---

## 7. 环境变量参考

`.env.production` 中可配置：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_PASSWORD` | 数据库密码 | **必填** |
| `JWT_SECRET` | JWT 签名密钥 | **必填** |
| `DOMAIN` | 站点域名 | **必填** |
| `EMAIL` | Let's Encrypt 邮箱 | **必填** |
| `CORS_ORIGINS` | 允许的跨域来源 | `https://yourdomain.com` |
| `REGISTRATION_MODE` | `open` / `invite_only` / `closed` | `invite_only` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `HTTP_PORT` | HTTP 端口 | `80` |
| `HTTPS_PORT` | HTTPS 端口 | `443` |
| `DEFAULT_LOCALE` | 前端默认语言 | `zh-CN` |

后端也支持 `GASTRACK_` 前缀的环境变量覆盖任意 YAML 配置（如 `GASTRACK_DATABASE_HOST`）。

---

## 8. 防火墙

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
# 确保 5432 和 8098 不对外暴露（Docker 内网通信，无需开放）
```

---

## 9. 安全清单

- [x] `JWT_SECRET` 使用 `openssl rand -hex 32` 随机生成
- [x] `DB_PASSWORD` 使用强密码
- [x] PostgreSQL 不暴露到宿主机（仅 Docker 内网）
- [x] Go 后端不暴露到宿主机（仅 Nginx 反代）
- [x] HTTPS + HSTS + TLS 1.2/1.3
- [x] 防火墙仅开放 22/80/443
- [x] CORS 限制为 `https://你的域名`
- [x] 默认邀请制注册（`invite_only`）
- [x] 定时数据库备份
