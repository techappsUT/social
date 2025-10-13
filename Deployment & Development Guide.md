# SocialQueue - Complete Deployment & Development Guide

**Last Updated**: October 13, 2025  
**Version**: 1.0  
**Platform**: Linux, macOS, Windows (via WSL2)

---

## 📋 TABLE OF CONTENTS

1. [Prerequisites](#prerequisites)
2. [Development Setup](#development-setup)
3. [Production Deployment](#production-deployment)
4. [Environment Configuration](#environment-configuration)
5. [Database Management](#database-management)
6. [Troubleshooting](#troubleshooting)
7. [Maintenance & Operations](#maintenance--operations)

---

## 🛠️ PREREQUISITES

### Required Software

#### 1. Go (Backend)
```bash
# macOS (via Homebrew)
brew install go

# Linux
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Verify
go version  # Should show go1.21 or higher
```

#### 2. Node.js & pnpm (Frontend)
```bash
# macOS
brew install node
npm install -g pnpm

# Linux
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
npm install -g pnpm

# Verify
node --version  # Should be v18 or higher
pnpm --version
```

#### 3. Docker & Docker Compose
```bash
# macOS
brew install --cask docker

# Linux (Ubuntu/Debian)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker --version
docker-compose --version
```

#### 4. PostgreSQL Client (Optional - for direct DB access)
```bash
# macOS
brew install postgresql@15

# Linux
sudo apt install postgresql-client-15

# Verify
psql --version
```

#### 5. Migration Tool
```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$PATH:$HOME/go/bin"

# Verify
migrate -version
```

#### 6. SQLC (Optional - for code generation)
```bash
# macOS
brew install sqlc

# Linux
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify
sqlc version
```

#### 7. Make (Usually pre-installed)
```bash
# macOS
xcode-select --install

# Linux
sudo apt install build-essential

# Verify
make --version
```

---

## 🚀 DEVELOPMENT SETUP

### Step 1: Clone Repository

```bash
# Clone the repository
git clone https://github.com/techappsUT/social-queue.git
cd socialqueue

# Verify project structure
ls -la
# You should see: backend/, frontend/, docker-compose.yml, Makefile
```

### Step 2: Environment Configuration

#### Backend Environment Variables
```bash
# Copy example environment file
cd backend
cp .env.example .env

# Edit .env with your favorite editor
nano .env
```

**Required changes in `.env`**:
```bash
# Server Configuration
PORT=8000
HOST=localhost
BASE_URL=http://localhost:3000

# Database (Docker defaults - don't change for dev)
DB_HOST=localhost
DB_PORT=5432
DB_USER=socialqueue
DB_PASSWORD=socialqueue_dev_password
DB_NAME=socialqueue_dev
DB_SSLMODE=disable

# JWT Secrets (IMPORTANT: Change these!)
# Generate with: openssl rand -base64 32
JWT_ACCESS_SECRET=your-super-secret-access-key-min-32-characters
JWT_REFRESH_SECRET=your-super-secret-refresh-key-min-32-characters

# Security
# Generate with: openssl rand -hex 32
ENCRYPTION_KEY=your-64-character-hex-encryption-key-here

# Redis (Docker defaults)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Social Media Credentials (Get from platforms)
# Twitter/X - https://developer.twitter.com/
TWITTER_CLIENT_ID=your_twitter_client_id
TWITTER_CLIENT_SECRET=your_twitter_client_secret

# Facebook - https://developers.facebook.com/
FACEBOOK_APP_ID=your_facebook_app_id
FACEBOOK_APP_SECRET=your_facebook_app_secret

# LinkedIn - https://www.linkedin.com/developers/
LINKEDIN_CLIENT_ID=your_linkedin_client_id
LINKEDIN_CLIENT_SECRET=your_linkedin_client_secret

# Email (Use Mailtrap for dev - https://mailtrap.io/)
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your_mailtrap_user
SMTP_PASSWORD=your_mailtrap_password
SMTP_FROM_EMAIL=noreply@socialqueue.dev
SMTP_FROM_NAME=SocialQueue Dev
```

#### Frontend Environment Variables
```bash
cd ../frontend
cp .env.example .env.local

# Edit .env.local
nano .env.local
```

**Required changes in `.env.local`**:
```bash
# API endpoint
NEXT_PUBLIC_API_URL=http://localhost:8000

# App URL
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

### Step 3: Start Services with Docker

```bash
# Return to project root
cd ..

# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Verify services are running
docker-compose ps

# You should see:
# socialqueue-postgres    ... Up (healthy)
# socialqueue-redis       ... Up (healthy)
```

**Wait for health checks** (about 10-15 seconds):
```bash
# Check PostgreSQL
docker-compose exec postgres pg_isready -U socialqueue

# Check Redis
docker-compose exec redis redis-cli ping
```

### Step 4: Database Setup & Migrations

```bash
cd backend

# Verify database connection
psql -h localhost -U socialqueue -d socialqueue_dev
# Password: socialqueue_dev_password
# Type \q to exit

# Run migrations
make migrate-up

# You should see:
# 20240101000001/u initial_schema (xxx.xxxs)
# All migrations applied successfully!

# Verify migration status
make migrate-status

# Check tables were created
make db-shell
# In psql prompt:
\dt
# You should see all tables: users, teams, posts, etc.
\q
```

### Step 5: Start Backend API

```bash
# In backend directory
cd backend

# Install Go dependencies
go mod download

# Generate SQLC code (if you made SQL changes)
make sqlc

# Run the API server
make run

# You should see:
# 🚀 Starting SocialQueue API Server...
# ✓ Loaded .env file
# ✓ Dependencies initialized
# ✓ Router configured
# 🚀 Server starting on http://localhost:8000
```

**Test the API**:
```bash
# In another terminal
curl http://localhost:8000/health

# Expected response:
# {"status":"ok","database":"up","redis":"up","timestamp":"..."}
```

### Step 6: Start Frontend

```bash
# In another terminal
cd frontend

# Install dependencies
pnpm install

# Start development server
pnpm dev

# You should see:
# ▲ Next.js 15.x.x
# - Local:        http://localhost:3000
# - Environments: .env.local
```

**Open in browser**: http://localhost:3000

### Step 7: Verify Everything Works

#### Test Backend API
```bash
# Test signup
curl -X POST http://localhost:8000/api/v2/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "SecurePass123!",
    "firstName": "Test",
    "lastName": "User"
  }'

# Expected: 201 Created with user data and tokens
```

#### Test Frontend
1. Open http://localhost:3000
2. Navigate to `/signup`
3. Create an account
4. Check that you're redirected to dashboard

---

## 📦 PRODUCTION DEPLOYMENT

### Overview

Production deployment uses:
- **Application**: Docker containers
- **Database**: Managed PostgreSQL (AWS RDS, Google Cloud SQL, or self-hosted)
- **Cache**: Managed Redis (AWS ElastiCache, Redis Cloud, or self-hosted)
- **Reverse Proxy**: Nginx
- **SSL**: Let's Encrypt (Certbot)
- **Monitoring**: Prometheus + Grafana (optional)

### Option 1: Deploy with Docker Compose (Simple)

Best for: Small teams, MVP, staging environments

#### Step 1: Server Setup

```bash
# SSH into your server
ssh user@your-server.com

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Clone repository
git clone https://github.com/techappsUT/social-queue.git
cd socialqueue
```

#### Step 2: Production Environment

```bash
# Create production environment file
cp backend/.env.example backend/.env.prod

# Edit with production values
nano backend/.env.prod
```

**Production `.env.prod`**:
```bash
# Server
PORT=8000
HOST=0.0.0.0
BASE_URL=https://socialqueue.com

# Database (Use managed service)
DB_HOST=your-db-hostname.rds.amazonaws.com
DB_PORT=5432
DB_USER=socialqueue_prod
DB_PASSWORD=<strong-random-password>
DB_NAME=socialqueue_prod
DB_SSLMODE=require

# JWT Secrets (CRITICAL: Use strong, unique values)
JWT_ACCESS_SECRET=<64-character-random-string>
JWT_REFRESH_SECRET=<64-character-random-string>

# Security
ENCRYPTION_KEY=<64-character-random-string>
BCRYPT_COST=12

# Redis (Use managed service)
REDIS_HOST=your-redis.cache.amazonaws.com
REDIS_PORT=6379
REDIS_PASSWORD=<strong-random-password>

# Social Media (Production OAuth apps)
TWITTER_CLIENT_ID=<prod-twitter-id>
TWITTER_CLIENT_SECRET=<prod-twitter-secret>
FACEBOOK_APP_ID=<prod-facebook-id>
FACEBOOK_APP_SECRET=<prod-facebook-secret>
LINKEDIN_CLIENT_ID=<prod-linkedin-id>
LINKEDIN_CLIENT_SECRET=<prod-linkedin-secret>

# Email (Use SendGrid, AWS SES, or Mailgun)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=<sendgrid-api-key>
SMTP_FROM_EMAIL=noreply@socialqueue.com
SMTP_FROM_NAME=SocialQueue

# Security Settings
COOKIE_SECURE=true
COOKIE_SAMESITE=strict
RATE_LIMIT_ENABLED=true

# Environment
ENVIRONMENT=production
LOG_LEVEL=info
```

#### Step 3: Build Production Images

```bash
# Build backend
cd backend
docker build -f Dockerfile.prod -t socialqueue-backend:latest .

# Build frontend
cd ../frontend
docker build -f Dockerfile.prod -t socialqueue-frontend:latest .
```

#### Step 4: Setup Nginx with SSL

```bash
# Install Nginx
sudo apt install nginx certbot python3-certbot-nginx -y

# Create Nginx config
sudo nano /etc/nginx/sites-available/socialqueue
```

**Nginx configuration**:
```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name socialqueue.com www.socialqueue.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS configuration
server {
    listen 443 ssl http2;
    server_name socialqueue.com www.socialqueue.com;

    # SSL certificates (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/socialqueue.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/socialqueue.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';" always;

    # Client max body size (for media uploads)
    client_max_body_size 10M;

    # Backend API
    location /api/ {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint (bypass rate limiting)
    location /health {
        proxy_pass http://localhost:8000/health;
        access_log off;
    }

    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Static assets
    location /_next/static/ {
        proxy_pass http://localhost:3000;
        proxy_cache_valid 200 60m;
        add_header Cache-Control "public, immutable";
    }
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
```

**Enable the site**:
```bash
# Enable configuration
sudo ln -s /etc/nginx/sites-available/socialqueue /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Get SSL certificate
sudo certbot --nginx -d socialqueue.com -d www.socialqueue.com

# Restart Nginx
sudo systemctl restart nginx
```

#### Step 5: Run Database Migrations

```bash
cd backend

# Test connection to production database
psql -h your-db-hostname.rds.amazonaws.com -U socialqueue_prod -d socialqueue_prod

# IMPORTANT: Backup first!
pg_dump -h your-db-hostname.rds.amazonaws.com -U socialqueue_prod socialqueue_prod > backup_$(date +%Y%m%d).sql

# Run migrations
migrate -path ./migrations \
        -database "postgres://socialqueue_prod:password@your-db-hostname.rds.amazonaws.com:5432/socialqueue_prod?sslmode=require" \
        up

# Verify
migrate -path ./migrations \
        -database "postgres://socialqueue_prod:password@your-db-hostname.rds.amazonaws.com:5432/socialqueue_prod?sslmode=require" \
        version
```

#### Step 6: Start Production Services

```bash
# Create production docker-compose file
nano docker-compose.prod.yml
```

```yaml
version: '3.8'

services:
  backend:
    image: socialqueue-backend:latest
    restart: always
    ports:
      - "127.0.0.1:8000:8000"
    env_file:
      - backend/.env.prod
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  frontend:
    image: socialqueue-frontend:latest
    restart: always
    ports:
      - "127.0.0.1:3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=https://socialqueue.com
      - NEXT_PUBLIC_APP_URL=https://socialqueue.com
    depends_on:
      - backend
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  app-network:
    driver: bridge
```

**Start services**:
```bash
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

#### Step 7: Verify Deployment

```bash
# Test health endpoint
curl https://socialqueue.com/health

# Test API
curl https://socialqueue.com/api/v2/health

# Check SSL
curl -I https://socialqueue.com
# Should show: strict-transport-security header
```

### Option 2: Deploy to AWS (Scalable)

Best for: Production at scale, enterprise

#### Architecture

```
[Route 53] → [CloudFront CDN] → [ALB]
                                  ↓
                    ┌─────────────┼─────────────┐
                    ↓             ↓             ↓
                 [ECS Task 1] [ECS Task 2] [ECS Task 3]
                    ↓             ↓             ↓
                [RDS PostgreSQL Multi-AZ]
                    ↓
                [ElastiCache Redis]
                    ↓
                [S3 - Media Storage]
```

#### Step 1: Setup AWS Infrastructure

**Using AWS CLI**:
```bash
# Install AWS CLI
pip install awscli

# Configure credentials
aws configure

# Create VPC and subnets (or use default)
aws ec2 create-vpc --cidr-block 10.0.0.0/16

# Create RDS database
aws rds create-db-instance \
    --db-instance-identifier socialqueue-prod \
    --db-instance-class db.t3.medium \
    --engine postgres \
    --engine-version 15.4 \
    --master-username socialqueue \
    --master-user-password <strong-password> \
    --allocated-storage 100 \
    --storage-encrypted \
    --multi-az \
    --backup-retention-period 7

# Create ElastiCache Redis cluster
aws elasticache create-cache-cluster \
    --cache-cluster-id socialqueue-redis \
    --cache-node-type cache.t3.micro \
    --engine redis \
    --num-cache-nodes 1

# Create S3 bucket for media
aws s3 mb s3://socialqueue-media-prod

# Create ECR repositories
aws ecr create-repository --repository-name socialqueue-backend
aws ecr create-repository --repository-name socialqueue-frontend
```

#### Step 2: Build and Push Docker Images

```bash
# Get ECR login
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Tag images
docker tag socialqueue-backend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/socialqueue-backend:latest
docker tag socialqueue-frontend:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/socialqueue-frontend:latest

# Push images
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/socialqueue-backend:latest
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/socialqueue-frontend:latest
```

#### Step 3: Create ECS Cluster

```bash
# Create ECS cluster
aws ecs create-cluster --cluster-name socialqueue-prod

# Create task definition
# (See ecs-task-definition.json below)

# Register task definition
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json

# Create ECS service
aws ecs create-service \
    --cluster socialqueue-prod \
    --service-name socialqueue-api \
    --task-definition socialqueue-backend \
    --desired-count 2 \
    --launch-type FARGATE \
    --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}"
```

**ecs-task-definition.json**:
```json
{
  "family": "socialqueue-backend",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "backend",
      "image": "<account-id>.dkr.ecr.us-east-1.amazonaws.com/socialqueue-backend:latest",
      "portMappings": [
        {
          "containerPort": 8000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "PORT",
          "value": "8000"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:xxx:secret:socialqueue/db-password"
        },
        {
          "name": "JWT_ACCESS_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:xxx:secret:socialqueue/jwt-access"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/socialqueue-backend",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8000/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

#### Step 4: Setup Application Load Balancer

```bash
# Create ALB
aws elbv2 create-load-balancer \
    --name socialqueue-alb \
    --subnets subnet-xxx subnet-yyy \
    --security-groups sg-xxx \
    --scheme internet-facing

# Create target group
aws elbv2 create-target-group \
    --name socialqueue-backend \
    --protocol HTTP \
    --port 8000 \
    --vpc-id vpc-xxx \
    --target-type ip \
    --health-check-path /health

# Create listener
aws elbv2 create-listener \
    --load-balancer-arn arn:aws:elasticloadbalancing:xxx \
    --protocol HTTPS \
    --port 443 \
    --certificates CertificateArn=arn:aws:acm:xxx \
    --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:xxx
```

---

## ⚙️ ENVIRONMENT CONFIGURATION

### Configuration Priority

```
Command Line Args > Environment Variables > .env File > Defaults
```

### All Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| **Server** ||||
| `PORT` | Yes | `8000` | HTTP server port |
| `HOST` | Yes | `localhost` | Server host |
| `BASE_URL` | Yes | - | Frontend URL (for CORS) |
| `ENVIRONMENT` | No | `development` | Environment name |
| **Database** ||||
| `DB_HOST` | Yes | - | PostgreSQL host |
| `DB_PORT` | Yes | `5432` | PostgreSQL port |
| `DB_USER` | Yes | - | Database username |
| `DB_PASSWORD` | Yes | - | Database password |
| `DB_NAME` | Yes | - | Database name |
| `DB_SSLMODE` | No | `disable` | SSL mode (dev: disable, prod: require) |
| `DB_MAX_CONNECTIONS` | No | `25` | Connection pool size |
| **JWT** ||||
| `JWT_ACCESS_SECRET` | Yes | - | Access token secret (min 32 chars) |
| `JWT_REFRESH_SECRET` | Yes | - | Refresh token secret (min 32 chars) |
| `JWT_ACCESS_EXPIRY` | No | `15m` | Access token TTL |
| `JWT_REFRESH_EXPIRY` | No | `30d` | Refresh token TTL |
| **Redis** ||||
| `REDIS_HOST` | Yes | `localhost` | Redis host |
| `REDIS_PORT` | Yes | `6379` | Redis port |
| `REDIS_PASSWORD` | No | - | Redis password |
| `REDIS_DB` | No | `0` | Redis database number |
| **Security** ||||
| `ENCRYPTION_KEY` | Yes | - | OAuth token encryption key (64 chars) |
| `BCRYPT_COST` | No | `12` | Password hashing cost |
| `COOKIE_SECURE` | No | `false` | HTTPS-only cookies |
| `COOKIE_SAMESITE` | No | `strict` | SameSite cookie policy |
| **Social Media** ||||
| `TWITTER_CLIENT_ID` | No | - | Twitter OAuth client ID |
| `TWITTER_CLIENT_SECRET` | No | - | Twitter OAuth secret |
| `FACEBOOK_APP_ID` | No | - | Facebook app ID |
| `FACEBOOK_APP_SECRET` | No | - | Facebook app secret |
| `LINKEDIN_CLIENT_ID` | No | - | LinkedIn client ID |
| `LINKEDIN_CLIENT_SECRET` | No | - | LinkedIn client secret |
| **Email** ||||
| `SMTP_HOST` | No | - | SMTP server host |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_USER` | No | - | SMTP username |
| `SMTP_PASSWORD` | No | - | SMTP password |
| `SMTP_FROM_EMAIL` | No | - | From email address |
| `SMTP_FROM_NAME` | No | `SocialQueue` | From name |
| **Rate Limiting** ||||
| `RATE_LIMIT_ENABLED` | No | `true` | Enable rate limiting |
| `RATE_LIMIT_MAX_REQUESTS` | No | `100` | Max requests per window |
| `RATE_LIMIT_WINDOW` | No | `1m` | Time window |
| **Logging** ||||
| `LOG_LEVEL` | No | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | No | `json` | Log format (json, text) |

---

## 🗄️ DATABASE MANAGEMENT

### Common Operations

#### Connect to Database
```bash
# Development
psql -h localhost -U socialqueue -d socialqueue_dev

# Production (with SSL)
psql "postgresql://socialqueue_prod:password@your-db.rds.amazonaws.com:5432/socialqueue_prod?sslmode=require"
```

#### Backup Database
```bash
# Full backup
pg_dump -h localhost -U socialqueue socialqueue_dev > backup_$(date +%Y%m%d_%H%M%S).sql

# Schema only
pg_dump -h localhost -U socialqueue -s socialqueue_dev > schema.sql

# Data only
pg_dump -h localhost -U socialqueue -a socialqueue_dev > data.sql

# Compressed backup
pg_dump -h localhost -U socialqueue socialqueue_dev | gzip > backup.sql.gz
```

#### Restore Database
```bash
# From SQL file
psql -h localhost -U socialqueue socialqueue_dev < backup.sql

# From compressed file
gunzip -c backup.sql.gz | psql -h localhost -U socialqueue socialqueue_dev
```

#### Create Migration
```bash
# Create new migration
migrate create -ext sql -dir backend/migrations -seq add_new_feature

# This creates:
# - 000002_add_new_feature.up.sql
# - 000002_add_new_feature.down.sql
```

#### Run Migrations
```bash
# Up (apply all)
make migrate-up

# Down (rollback one)
make migrate-down

# Force version (if migrations table is corrupt)
migrate -path ./migrations -database "postgres://..." force 1

# Check status
make migrate-status
```

#### Reset Database (⚠️ DANGER - Development Only)
```bash
# Drop and recreate
make db-reset

# Manual reset
docker-compose exec postgres psql -U socialqueue -d socialqueue_dev -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
make migrate-up
```

---

## 🔧 TROUBLESHOOTING

### Common Issues

#### 1. Port Already in Use
```bash
# Error: "bind: address already in use"

# Find process using port 8000
lsof -i :8000

# Kill process
kill -9 <PID>

# Or use different port
PORT=8001 make run
```

#### 2. Database Connection Failed
```bash
# Error: "connection refused"

# Check if PostgreSQL is running
docker-compose ps postgres

# Check logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres

# Test connection manually
psql -h localhost -U socialqueue -d socialqueue_dev
```

#### 3. Migration Dirty State
```bash
# Error: "Dirty database version X. Fix and force version."

# Check current version
migrate -path ./migrations -database "postgres://..." version

# Force to correct version (be careful!)
migrate -path ./migrations -database "postgres://..." force 1

# Then run migrations again
make migrate-up
```

#### 4. SQLC Generation Errors
```bash
# Error: "unknown type uuid"

# Regenerate
cd backend
make sqlc

# If still fails, check sqlc.yaml configuration
cat sqlc.yaml
```

#### 5. Redis Connection Failed
```bash
# Error: "dial tcp: connect: connection refused"

# Check if Redis is running
docker-compose ps redis

# Test Redis connection
docker-compose exec redis redis-cli ping
# Expected: PONG

# Restart Redis
docker-compose restart redis
```

#### 6. Frontend Build Errors
```bash
# Error: "Module not found"

# Clear and reinstall
cd frontend
rm -rf node_modules .next
pnpm install
pnpm dev
```

#### 7. OAuth Redirect Mismatch
```bash
# Error: "redirect_uri_mismatch"

# Verify OAuth redirect URIs match in platform settings:
# Development: http://localhost:8000/api/v2/social/auth/{platform}/callback
# Production: https://socialqueue.com/api/v2/social/auth/{platform}/callback
```

---

## 🔄 MAINTENANCE & OPERATIONS

### Daily Operations

#### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend

# Last 100 lines
docker-compose logs --tail=100 backend

# Production (ECS)
aws logs tail /ecs/socialqueue-backend --follow
```

#### Monitor Health
```bash
# Health check
curl http://localhost:8000/health

# Detailed metrics (if implemented)
curl http://localhost:8000/metrics

# Database connections
psql -h localhost -U socialqueue -d socialqueue_dev -c "SELECT count(*) FROM pg_stat_activity;"
```

#### Restart Services
```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart backend

# Production (ECS)
aws ecs update-service --cluster socialqueue-prod --service socialqueue-api --force-new-deployment
```

### Weekly Maintenance

#### Database Maintenance
```bash
# Vacuum analyze (reclaim space, update statistics)
psql -h localhost -U socialqueue -d socialqueue_dev -c "VACUUM ANALYZE;"

# Check database size
psql -h localhost -U socialqueue -d socialqueue_dev -c "SELECT pg_size_pretty(pg_database_size('socialqueue_dev'));"

# Check table sizes
psql -h localhost -U socialqueue -d socialqueue_dev -c "
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

#### Backup Database
```bash
# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U socialqueue socialqueue_prod | gzip > $BACKUP_DIR/backup_$TIMESTAMP.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete
```

### Monthly Maintenance

#### Update Dependencies
```bash
# Backend
cd backend
go get -u ./...
go mod tidy

# Frontend
cd frontend
pnpm update

# Docker images
docker-compose pull
docker-compose up -d
```

#### Security Updates
```bash
# Check for vulnerabilities
cd backend
go list -json -m all | nancy sleuth

cd frontend
pnpm audit

# Apply security patches
pnpm audit fix
```

### Performance Monitoring

#### Database Performance
```bash
# Slow queries
psql -h localhost -U socialqueue -d socialqueue_prod -c "
SELECT 
    query, 
    calls, 
    total_time, 
    mean_time 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"

# Missing indexes
psql -h localhost -U socialqueue -d socialqueue_prod -c "
SELECT 
    schemaname, 
    tablename, 
    attname, 
    null_frac 
FROM pg_stats 
WHERE null_frac < 0.9 
ORDER BY null_frac DESC;"
```

#### Redis Performance
```bash
# Redis stats
docker-compose exec redis redis-cli INFO stats

# Memory usage
docker-compose exec redis redis-cli INFO memory

# Slowlog
docker-compose exec redis redis-cli SLOWLOG GET 10
```

---

## 📊 MONITORING SETUP (Optional but Recommended)

### Prometheus + Grafana

**docker-compose.monitoring.yml**:
```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    networks:
      - monitoring

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus-data:
  grafana-data:
```

---

## 🎓 SUMMARY

### Development Workflow
1. **Start services**: `docker-compose up -d`
2. **Run migrations**: `cd backend && make migrate-up`
3. **Start backend**: `make run`
4. **Start frontend**: `cd frontend && pnpm dev`
5. **Code, test, repeat**

### Deployment Workflow
1. **Test locally**: Ensure all tests pass
2. **Build images**: Docker build for backend + frontend
3. **Run migrations**: On production database (with backup!)
4. **Deploy**: Update ECS/Docker Compose
5. **Verify**: Health checks + smoke tests
6. **Monitor**: Check logs and metrics

### Key Commands Reference
```bash
# Development
make up          # Start all services
make down        # Stop all services
make logs        # View logs
make migrate-up  # Run migrations
make run         # Start backend

# Database
make db-shell    # PostgreSQL shell
make db-reset    # Reset database (DEV ONLY)
make migrate-status  # Check migration status

# Production
docker-compose -f docker-compose.prod.yml up -d
aws ecs update-service --cluster socialqueue-prod --force-new-deployment
```

---

**Need Help?**
- Check logs: `docker-compose logs -f`
- Review this guide's troubleshooting section
- Check GitHub Issues
- Contact: support@socialqueue.com

**Last Updated**: October 13, 2025