# 📦 Production Packaging Guide

<div align="center">

**How NeuronIP is packaged for production**

[← Docker](docker.md) • [Production](production.md) • [Kubernetes](kubernetes.md)

</div>

---

## 📋 Table of Contents

- [Overview](#overview)
- [Docker Containerization](#docker-containerization)
- [Image Build Strategy](#image-build-strategy)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)
- [Image Registry](#image-registry)
- [Build Optimizations](#build-optimizations)

---

## Overview

NeuronIP is packaged as **two Docker containers** for production deployment:

1. **Backend API Container** (`neuronip/api`)
   - Go application compiled to static binary
   - Minimal Alpine Linux runtime
   - Image size: ~10-20MB

2. **Frontend Container** (`neuronip/frontend`)
   - Next.js production build
   - Node.js Alpine runtime
   - Image size: ~150-200MB

**Important**: External services (PostgreSQL, NeuronDB, NeuronMCP, NeuronAgent) are **not** packaged with NeuronIP and must be deployed separately.

---

## Docker Containerization

### Backend API Container

**Location:** [`api/Dockerfile`](../../api/Dockerfile)

**Multi-stage build pattern:**

#### Stage 1: Builder

**Complete Dockerfile:**
```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/neuronip-api ./cmd/server
```

**Line-by-line explanation:**

1. **`FROM golang:1.24-alpine AS builder`**
   - Base image: Official Go 1.24 Alpine Linux
   - Alpine: Minimal Linux distribution (~5MB base)
   - `AS builder`: Names this stage for multi-stage build
   - Includes Go compiler, build tools, and standard library

2. **`WORKDIR /app`**
   - Sets working directory inside container
   - All subsequent commands run from `/app`
   - Creates directory if it doesn't exist

3. **`COPY go.mod go.sum ./`**
   - Copies dependency manifest files first
   - **Critical optimization**: Separates dependency layer from source code
   - Docker caches this layer separately
   - Only rebuilds if dependencies change

4. **`RUN go mod download`**
   - Downloads all Go module dependencies
   - Creates cache layer for dependencies
   - Subsequent builds skip this if `go.mod`/`go.sum` unchanged
   - Downloads to `$GOPATH/pkg/mod` (cached in layer)

5. **`COPY . .`**
   - Copies all source code into container
   - Excludes files in `.dockerignore`
   - This layer invalidates on any source change
   - Dependencies already cached, so rebuild is fast

6. **`RUN CGO_ENABLED=0 GOOS=linux go build -o /app/neuronip-api ./cmd/server`**
   - **`CGO_ENABLED=0`**: Disables CGO (C bindings)
     - Creates fully static binary
     - No external C library dependencies
     - Binary runs on any Linux system
   - **`GOOS=linux`**: Targets Linux operating system
   - **`-o /app/neuronip-api`**: Output binary name and path
   - **`./cmd/server`**: Source package to build
   - Compiles Go code to machine code

**Build output:**
- Binary size: ~15-25MB (depends on dependencies)
- Architecture: Linux amd64 (or specified target)
- Type: Static ELF executable

#### Stage 2: Runtime

**Complete Dockerfile:**
```dockerfile
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/neuronip-api .

EXPOSE 8082

CMD ["./neuronip-api"]
```

**Line-by-line explanation:**

1. **`FROM alpine:latest`**
   - Minimal Alpine Linux base image (~5MB)
   - No build tools, no compilers, no package managers (except apk)
   - Minimal attack surface
   - Only essential system libraries

2. **`RUN apk --no-cache add ca-certificates`**
   - **`apk`**: Alpine package manager
   - **`--no-cache`**: Doesn't store package index (saves space)
   - **`ca-certificates`**: Root CA certificates bundle
   - Required for HTTPS/TLS connections to external services
   - Enables SSL certificate validation

3. **`WORKDIR /root/`**
   - Sets working directory to root user's home
   - Binary will execute from this directory
   - Standard location for single-binary applications

4. **`COPY --from=builder /app/neuronip-api .`**
   - **`--from=builder`**: Copies from named build stage
   - Copies only the compiled binary
   - No source code, no build tools, no dependencies
   - Final image contains only: Alpine base + CA certs + binary

5. **`EXPOSE 8082`**
   - Documents that container listens on port 8082
   - Doesn't actually open the port (that's done at runtime)
   - Documentation for Docker and orchestrators
   - Used by `docker run -P` to map ports

6. **`CMD ["./neuronip-api"]`**
   - Default command to run when container starts
   - Executes the binary directly
   - **`[]`**: Exec form (recommended, no shell wrapper)
   - Process becomes PID 1 in container

**Final image composition:**
- Base: Alpine Linux (~5MB)
- CA certificates: ~200KB
- Binary: ~15-25MB
- **Total: ~20-30MB** (vs ~300MB+ with full golang image)

**Security benefits:**
- No shell access (unless explicitly added)
- No package manager in runtime (can't install malware)
- Minimal base reduces CVE surface area
- Static binary reduces dependency vulnerabilities

**Optimizations:**
- ✅ Layer caching: Dependencies cached separately
- ✅ Static binary: No runtime dependencies
- ✅ Minimal base: Alpine reduces attack surface
- ✅ No build tools: Only binary in final image

### Frontend Container

**Location:** [`frontend/Dockerfile`](../../frontend/Dockerfile)

**Multi-stage build pattern:**

#### Stage 1: Builder

**Complete Dockerfile:**
```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build
```

**Line-by-line explanation:**

1. **`FROM node:20-alpine AS builder`**
   - Base image: Official Node.js 20 Alpine Linux
   - Includes Node.js runtime, npm, and build tools
   - Alpine variant: Smaller than standard Node image
   - `AS builder`: Names this stage for multi-stage build

2. **`WORKDIR /app`**
   - Sets working directory to `/app`
   - All commands execute from this directory
   - Creates directory structure

3. **`COPY package*.json ./`**
   - Copies `package.json` and `package-lock.json`
   - **Critical optimization**: Separates dependency layer
   - Docker caches this layer independently
   - Only rebuilds dependencies if package files change

4. **`RUN npm ci`**
   - **`npm ci`**: Clean install (vs `npm install`)
   - **Deterministic**: Installs exact versions from `package-lock.json`
   - **Faster**: Skips dependency resolution
   - **Reliable**: Fails if `package-lock.json` is out of sync
   - Removes `node_modules` first, then installs fresh
   - Production dependencies only (if `--production` flag used)

5. **`COPY . .`**
   - Copies all source code into container
   - Includes TypeScript files, React components, configs
   - Excludes files in `.dockerignore`
   - This layer invalidates on source changes

6. **`RUN npm run build`**
   - Executes Next.js production build
   - **Process:**
     1. TypeScript compilation (`tsc`)
     2. React component bundling
     3. Code minification and optimization
     4. Static page generation (SSG)
     5. Creates `.next` directory with optimized output
   - **Output:**
     - `.next/`: Optimized production build
     - Static assets: CSS, JS, images
     - Server-side code: API routes, middleware
     - Build manifest: For runtime optimization

**Build output:**
- `.next/`: Production build directory (~50-200MB)
- Optimized JavaScript bundles
- Minified CSS
- Static HTML pages (if SSG used)
- Source maps (optional, for debugging)

#### Stage 2: Runtime

**Complete Dockerfile:**
```dockerfile
FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

CMD ["npm", "start"]
```

**Line-by-line explanation:**

1. **`FROM node:20-alpine`**
   - Fresh Alpine Node.js image
   - No build artifacts from builder stage
   - Only Node.js runtime (no build tools needed)

2. **`WORKDIR /app`**
   - Sets working directory
   - Matches builder stage structure

3. **`COPY --from=builder /app/.next ./.next`**
   - Copies Next.js production build
   - Contains optimized server and client code
   - Required for `npm start` to work

4. **`COPY --from=builder /app/public ./public`**
   - Copies static public assets
   - Images, fonts, favicons, robots.txt
   - Served directly by Next.js
   - Not processed by build system

5. **`COPY --from=builder /app/package*.json ./`**
   - Copies package files
   - Required for `npm start` command
   - Contains scripts and metadata

6. **`COPY --from=builder /app/node_modules ./node_modules`**
   - Copies production dependencies
   - Only runtime dependencies (not devDependencies)
   - Required for Next.js to run
   - **Note**: Could be optimized with Next.js standalone mode

7. **`EXPOSE 3000`**
   - Documents container listens on port 3000
   - Next.js default port
   - Can be changed via `PORT` environment variable

8. **`CMD ["npm", "start"]`**
   - Runs `npm start` script
   - Executes: `next start` (production server)
   - Serves optimized build from `.next/`
   - Handles SSR, API routes, and static files

**Final image composition:**
- Base: Node.js Alpine (~40MB)
- Production dependencies: ~100-150MB
- `.next` build: ~50-200MB
- Public assets: ~5-50MB
- **Total: ~200-450MB** (depends on dependencies and build size)

**Optimization opportunities:**
- **Next.js Standalone Mode**: Reduces `node_modules` size
  - Only includes necessary dependencies
  - Can reduce image by 50-100MB
  - Configure in `next.config.js`: `output: 'standalone'`

**Optimizations:**
- ✅ Dependency layer caching
- ✅ Production-only artifacts
- ✅ No source code in final image

**Potential improvement:**
- Next.js standalone output mode (commented in `next.config.js`)
  - Would reduce image size further
  - Only includes necessary dependencies

---

## Image Build Strategy

### Build Context Optimization

Both containers use `.dockerignore` files to exclude unnecessary files:

**Backend exclusions:**
- Test files: `*_test.go`, `testdata/`
- Build artifacts: `*.exe`, `*.dll`, `*.so`
- IDE files: `.vscode/`, `.idea/`
- Documentation: `*.md`, `docs/`
- CI/CD: `.github/`, `.circleci/`
- Environment: `.env`, `.env.local`

**Frontend exclusions:**
- Dependencies: `node_modules/` (rebuilt in container)
- Test files: `*.test.ts`, `*.test.tsx`
- Build outputs: `.next/`, `out/`, `dist/`
- IDE files: `.vscode/`, `.idea/`
- Documentation: `*.md`, `docs/`

### Build Benefits

1. **Smaller Images**
   - Go backend: ~15MB vs ~300MB+ (if using full golang image)
   - Node.js frontend: Only production artifacts

2. **Faster Builds**
   - Layer caching speeds up subsequent builds
   - Only changed layers are rebuilt
   - Dependencies cached separately from source

3. **Security**
   - Minimal base images reduce attack surface
   - No build tools in runtime images
   - Only necessary files included

4. **Reproducible Builds**
   - Deterministic dependency installation
   - Consistent build environment
   - Same source code produces identical images

---

## CI/CD Pipeline

### GitHub Actions Workflows

#### Build Workflow (`.github/workflows/build.yml`)

**Location:** [`.github/workflows/build.yml`](../../.github/workflows/build.yml)

**Complete Workflow:**
```yaml
name: Build and Push Docker Images

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

jobs:
  build:
    name: Build and Push
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to Docker Hub
        if: github.event_name != 'pull_request'
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: |
            neuronip/api
            neuronip/frontend
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-
      
      - name: Build and push API image
        uses: docker/build-push-action@v5
        with:
          context: ./api
          push: ${{ github.event_name != 'pull_request' }}
          tags: neuronip/api:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}
      
      - name: Build and push Frontend image
        uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: ${{ github.event_name != 'pull_request' }}
          tags: neuronip/frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**Detailed Step-by-Step Explanation:**

1. **Trigger Configuration**
   ```yaml
   on:
     push:
       branches: [ main ]
       tags:
         - 'v*'
   ```
   - **Triggers on:**
     - Push to `main` branch (continuous integration)
     - Push of version tags matching `v*` (e.g., `v1.2.3`)
   - **Does NOT trigger on:**
     - Pull requests (handled by separate CI workflow)
     - Other branches (unless explicitly configured)

2. **Checkout Code**
   ```yaml
   - uses: actions/checkout@v4
   ```
   - Checks out repository code
   - Makes source available to workflow
   - Version: v4 (latest stable)

3. **Docker Buildx Setup**
   ```yaml
   - name: Set up Docker Buildx
     uses: docker/setup-buildx-action@v3
   ```
   - **Buildx**: Advanced Docker build features
   - **Features enabled:**
     - Multi-platform builds (amd64, arm64)
     - Build cache management
     - Parallel builds
     - Advanced caching strategies

4. **Docker Hub Authentication**
   ```yaml
   - name: Login to Docker Hub
     if: github.event_name != 'pull_request'
     uses: docker/login-action@v3
     with:
       username: ${{ secrets.DOCKER_USERNAME }}
       password: ${{ secrets.DOCKER_PASSWORD }}
   ```
   - **Conditional**: Only runs if not a pull request
   - **Secrets**: Stored in GitHub repository settings
   - **Purpose**: Authenticate to push images
   - **Security**: Secrets are masked in logs

5. **Metadata Extraction**
   ```yaml
   - name: Extract metadata
     id: meta
     uses: docker/metadata-action@v5
     with:
       images: |
         neuronip/api
         neuronip/frontend
       tags: |
         type=ref,event=branch
         type=ref,event=pr
         type=semver,pattern={{version}}
         type=semver,pattern={{major}}.{{minor}}
         type=sha,prefix={{branch}}-
   ```
   - **Purpose**: Generate image tags automatically
   - **Tag strategies:**
     - `type=ref,event=branch`: Tags with branch name
     - `type=ref,event=pr`: Tags with PR number
     - `type=semver`: Tags with semantic version
     - `type=sha`: Tags with commit SHA
   - **Example tags for `v1.2.3` on `main`:**
     - `neuronip/api:v1.2.3`
     - `neuronip/api:v1.2`
     - `neuronip/api:main-abc123`
     - `neuronip/api:latest` (if configured)

6. **Build and Push API Image**
   ```yaml
   - name: Build and push API image
     uses: docker/build-push-action@v5
     with:
       context: ./api
       push: ${{ github.event_name != 'pull_request' }}
       tags: neuronip/api:${{ github.sha }}
       cache-from: type=gha
       cache-to: type=gha,mode=max
       build-args: |
         VERSION=${{ github.sha }}
         BUILD_DATE=${{ github.event.head_commit.timestamp }}
   ```
   - **Context**: `./api` directory (build context)
   - **Push**: Only if not a pull request
   - **Tags**: Single tag with commit SHA
   - **Cache:**
     - `cache-from: type=gha`: Restores from GitHub Actions cache
     - `cache-to: type=gha,mode=max`: Saves all layers to cache
     - **Benefits**: Faster subsequent builds
   - **Build args**: Pass metadata to Dockerfile
     - `VERSION`: Commit SHA
     - `BUILD_DATE`: Timestamp of commit

7. **Build and Push Frontend Image**
   - Same process as API image
   - Context: `./frontend` directory
   - Tags: `neuronip/frontend:${{ github.sha }}`

**Image Tagging Strategy:**

| Event | Tag Format | Example |
|-------|-----------|---------|
| Push to main | `neuronip/api:main-abc123` | Commit SHA |
| Version tag | `neuronip/api:v1.2.3` | Semantic version |
| Version tag | `neuronip/api:v1.2` | Major.minor |
| PR #123 | `neuronip/api:pr-123` | PR number |

**Caching Strategy:**

- **GitHub Actions Cache (`type=gha`):**
  - Stores Docker layer cache
  - Persists between workflow runs
  - **Mode `max`**: Caches all layers (not just final)
  - **Benefits:**
    - Faster builds (reuses layers)
    - Reduced bandwidth (doesn't re-download)
    - Cost savings (faster CI runs)

**Build Performance:**

- **First build**: ~5-10 minutes (no cache)
- **Cached build**: ~2-5 minutes (with cache)
- **Cache hit rate**: ~70-90% (if dependencies unchanged)

#### Release Workflow (`.github/workflows/release.yml`)

**Location:** [`.github/workflows/release.yml`](../../.github/workflows/release.yml)

**Complete Workflow:**
```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    permissions:
      contents: write
    
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Get version from tag
        id: get_version
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          echo "Version: $VERSION"
      
      - name: Generate changelog
        id: changelog
        run: |
          VERSION="${{ steps.get_version.outputs.version }}"
          echo "changelog=See CHANGELOG.md" >> $GITHUB_OUTPUT
      
      - name: Build API
        working-directory: ./api
        run: |
          go build -ldflags "-X main.version=${{ steps.get_version.outputs.version }} -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.gitCommit=${GITHUB_SHA}" -o neuronip-api cmd/server/main.go
      
      - name: Build Frontend
        working-directory: ./frontend
        run: |
          npm ci
          npm run build
      
      - name: Build Docker images
        uses: docker/build-push-action@v5
        with:
          context: ./api
          push: true
          tags: |
            neuronip/api:${{ steps.get_version.outputs.version }}
            neuronip/api:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build Frontend Docker image
        uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: true
          tags: |
            neuronip/frontend:${{ steps.get_version.outputs.version }}
            neuronip/frontend:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Create GitHub Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ steps.get_version.outputs.version }}
          body: ${{ steps.changelog.outputs.changelog }}
          draft: false
          prerelease: false
      
      - name: Upload release artifacts
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ steps.create_release.outputs.upload_url }}
          asset_path: ./api/neuronip-api
          asset_name: neuronip-api-${{ steps.get_version.outputs.version }}
          asset_content_type: application/octet-stream
```

**Detailed Step-by-Step Explanation:**

1. **Trigger Configuration**
   ```yaml
   on:
     push:
       tags:
         - 'v*.*.*'
   ```
   - **Triggers on:** Version tags matching semantic versioning
   - **Examples:** `v1.0.0`, `v2.3.1`, `v10.5.2`
   - **Does NOT trigger on:** `v1`, `v1.0`, `beta-1.0.0`

2. **Permissions**
   ```yaml
   permissions:
     contents: write
   ```
   - Required to create GitHub Releases
   - Required to upload release artifacts
   - Minimal permissions (security best practice)

3. **Checkout with Full History**
   ```yaml
   - uses: actions/checkout@v4
     with:
       fetch-depth: 0
   ```
   - **`fetch-depth: 0`**: Fetches full git history
   - Required for changelog generation
   - Required for version comparison

4. **Version Extraction**
   ```bash
   VERSION=${GITHUB_REF#refs/tags/v}
   ```
   - **Input:** `refs/tags/v1.2.3`
   - **Process:** Removes `refs/tags/v` prefix
   - **Output:** `1.2.3`
   - **Stored in:** `steps.get_version.outputs.version`

5. **Go Binary Build with Metadata**
   ```bash
   go build -ldflags "-X main.version=$VERSION -X main.buildDate=$DATE -X main.gitCommit=$SHA" \
     -o neuronip-api cmd/server/main.go
   ```
   - **`-ldflags`**: Linker flags for variable injection
   - **`-X main.version`**: Injects version string
   - **`-X main.buildDate`**: Injects build timestamp (UTC)
   - **`-X main.gitCommit`**: Injects commit SHA
   - **Runtime access:** Binary can display version info
   - **Example usage in Go:**
     ```go
     var version = "unknown"
     var buildDate = "unknown"
     var gitCommit = "unknown"
     
     func main() {
         fmt.Printf("Version: %s\nBuild Date: %s\nCommit: %s\n", 
                    version, buildDate, gitCommit)
     }
     ```

6. **Docker Image Tagging**
   ```yaml
   tags: |
     neuronip/api:${{ steps.get_version.outputs.version }}
     neuronip/api:latest
   ```
   - **Version tag:** `neuronip/api:v1.2.3` (specific version)
   - **Latest tag:** `neuronip/api:latest` (most recent)
   - **Both tags point to same image**
   - **Latest tag updates on each release**

7. **GitHub Release Creation**
   ```yaml
   - name: Create GitHub Release
     uses: actions/create-release@v1
     with:
       tag_name: ${{ github.ref }}
       release_name: Release ${{ steps.get_version.outputs.version }}
       body: ${{ steps.changelog.outputs.changelog }}
       draft: false
       prerelease: false
   ```
   - **Creates:** GitHub Release for the tag
   - **Includes:** Release notes from changelog
   - **Status:** Published (not draft, not prerelease)
   - **URL:** `https://github.com/owner/repo/releases/tag/v1.2.3`

8. **Binary Artifact Upload**
   ```yaml
   - name: Upload release artifacts
     uses: actions/upload-release-asset@v1
     with:
       asset_path: ./api/neuronip-api
       asset_name: neuronip-api-${{ steps.get_version.outputs.version }}
   ```
   - **Uploads:** Compiled Go binary
   - **Purpose:** Direct binary download (no Docker required)
   - **Name:** `neuronip-api-v1.2.3`
   - **Location:** Attached to GitHub Release

**Release Artifacts:**

1. **Docker Images:**
   - `neuronip/api:v1.2.3`
   - `neuronip/api:latest`
   - `neuronip/frontend:v1.2.3`
   - `neuronip/frontend:latest`

2. **GitHub Release:**
   - Release notes
   - Source code archive
   - Binary artifact (`neuronip-api-v1.2.3`)

**Version Metadata in Go Binary:**

The binary contains embedded metadata accessible at runtime:

```bash
# View version information
./neuronip-api --version
# Output:
# Version: 1.2.3
# Build Date: 2024-01-15T10:30:00Z
# Git Commit: abc123def456
```

**Image Tags Created:**

| Tag | Purpose | Example |
|-----|---------|---------|
| `v1.2.3` | Specific version | `neuronip/api:v1.2.3` |
| `latest` | Most recent release | `neuronip/api:latest` |
| Both | Point to same image | Immutable version, mutable latest |

#### CI Workflow (`.github/workflows/ci.yml`)

**Purpose:** Validate builds without pushing

**Actions:**
- Runs tests for backend and frontend
- Builds Docker images (without pushing)
- Validates build process
- Uses GitHub Actions cache

---

## Release Process

### Manual Release

**Script:** [`scripts/release.sh`](../../scripts/release.sh)

**Usage:**
```bash
./scripts/release.sh patch   # v1.2.3 → v1.2.4
./scripts/release.sh minor   # v1.2.3 → v1.3.0
./scripts/release.sh major   # v1.2.3 → v2.0.0
```

**Process:**
1. Accepts version type (major/minor/patch)
2. Calculates new version from current git tag
3. Updates CHANGELOG.md (placeholder)
4. Creates git tag with version
5. Pushes tag to trigger release workflow

**Versioning:** Semantic versioning (vMAJOR.MINOR.PATCH)

### Automated Release

When a version tag is pushed:
1. GitHub Actions release workflow triggers
2. Docker images are built and pushed
3. GitHub Release is created
4. Binary artifacts are uploaded

**Example:**
```bash
# Create and push tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# GitHub Actions automatically:
# - Builds images
# - Tags as neuronip/api:v1.2.3 and neuronip/api:latest
# - Creates GitHub Release
```

---

## Image Registry

### Docker Hub

**Registry:** `docker.io`

**Images:**
- `neuronip/api`
- `neuronip/frontend`

### Tagging Strategy

**Latest Release:**
- `neuronip/api:latest`
- `neuronip/frontend:latest`

**Version Tags:**
- `neuronip/api:v1.2.3`
- `neuronip/frontend:v1.2.3`

**Branch Tags:**
- `neuronip/api:main-abc123` (branch-commit SHA)

**PR Tags:**
- `neuronip/api:pr-123`

### Pulling Images

**Docker Compose:**
```yaml
services:
  neuronip-api:
    image: neuronip/api:v1.2.3
    # or
    image: neuronip/api:latest
```

**Kubernetes:**
```yaml
containers:
  - name: api
    image: neuronip/api:v1.2.3
    imagePullPolicy: IfNotPresent
```

**Helm:**
```yaml
image:
  api:
    repository: neuronip/api
    tag: v1.2.3
    pullPolicy: IfNotPresent
```

---

## Build Optimizations

### Current Optimizations

1. **Multi-stage Builds**
   - Separate build and runtime stages
   - Minimal final images

2. **Layer Caching**
   - Dependencies in separate layers
   - GitHub Actions cache for faster builds

3. **Static Binary (Backend)**
   - No runtime dependencies
   - Single executable file

4. **Production Builds (Frontend)**
   - Only production artifacts
   - Optimized Next.js build

### Potential Improvements

1. **Next.js Standalone Mode**
   ```javascript
   // next.config.js
   output: 'standalone'
   ```
   - Reduces frontend image size
   - Only includes necessary dependencies

2. **BuildKit Cache Mounts**
   ```dockerfile
   RUN --mount=type=cache,target=/go/pkg/mod \
       go mod download
   ```
   - Faster dependency downloads
   - Better cache utilization

3. **Multi-architecture Builds**
   - Support ARM64 (Apple Silicon, ARM servers)
   - Use Docker Buildx for multi-platform builds

4. **Image Compression**
   - Use `docker-squash` or similar tools
   - Reduce image layers
   - Smaller image sizes

5. **Distroless Images**
   - Consider Google's distroless images
   - Even smaller attack surface
   - No shell, no package manager

---

## Build Commands

### Local Build

#### Backend API

**Basic build:**
```bash
cd api
docker build -t neuronip-api:local .
```

**Build with specific platform:**
```bash
docker build --platform linux/amd64 -t neuronip-api:local .
docker build --platform linux/arm64 -t neuronip-api:local-arm64 .
```

**Build with build arguments:**
```bash
docker build \
  --build-arg VERSION=1.2.3 \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t neuronip-api:local \
  .
```

**Build with cache from registry:**
```bash
docker build \
  --cache-from neuronip/api:latest \
  -t neuronip-api:local \
  .
```

**Build with progress output:**
```bash
docker build --progress=plain -t neuronip-api:local .
```

#### Frontend

**Basic build:**
```bash
cd frontend
docker build -t neuronip-frontend:local .
```

**Build with Node.js version override:**
```bash
docker build \
  --build-arg NODE_VERSION=20 \
  -t neuronip-frontend:local \
  .
```

**Build with npm registry override:**
```bash
docker build \
  --build-arg NPM_REGISTRY=https://registry.npmjs.org/ \
  -t neuronip-frontend:local \
  .
```

### Build with Docker Compose

**Build all services:**
```bash
docker compose build
```

**Build specific service:**
```bash
docker compose build neuronip-api
docker compose build neuronip-frontend
```

**Build without cache:**
```bash
docker compose build --no-cache
```

**Build with progress:**
```bash
docker compose build --progress=plain
```

**Build and start:**
```bash
docker compose up -d --build
```

**Parallel build (Docker BuildKit):**
```bash
DOCKER_BUILDKIT=1 docker compose build
```

### Build for Production

#### Single Image Build

**Backend:**
```bash
# Build with version tag
docker build -t neuronip/api:v1.2.3 ./api

# Tag as latest
docker tag neuronip/api:v1.2.3 neuronip/api:latest

# Push both tags
docker push neuronip/api:v1.2.3
docker push neuronip/api:latest
```

**Frontend:**
```bash
# Build with version tag
docker build -t neuronip/frontend:v1.2.3 ./frontend

# Tag as latest
docker tag neuronip/frontend:v1.2.3 neuronip/frontend:latest

# Push both tags
docker push neuronip/frontend:v1.2.3
docker push neuronip/frontend:latest
```

#### Multi-Architecture Build

**Using Docker Buildx:**

```bash
# Create buildx builder
docker buildx create --name multiarch --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t neuronip/api:v1.2.3 \
  -t neuronip/api:latest \
  --push \
  ./api

# Build frontend for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t neuronip/frontend:v1.2.3 \
  -t neuronip/frontend:latest \
  --push \
  ./frontend
```

**Benefits:**
- Single image supports multiple architectures
- Automatic platform selection
- Works on Intel, AMD, and ARM systems

#### Build with Secrets

**Using Docker BuildKit secrets:**

```bash
# Build with secret file
docker build \
  --secret id=npm_token,src=./.npmrc \
  -t neuronip/frontend:local \
  ./frontend
```

**In Dockerfile:**
```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=secret,id=npm_token \
  npm ci
```

### Advanced Build Options

#### Build with Custom Dockerfile

```bash
docker build -f Dockerfile.prod -t neuronip-api:prod ./api
```

#### Build with Target Stage

```bash
# Build only builder stage
docker build --target builder -t neuronip-api:builder ./api

# Build only runtime stage (default)
docker build --target runtime -t neuronip-api:runtime ./api
```

#### Build with Resource Limits

```bash
docker build \
  --memory=4g \
  --cpuset-cpus=0-3 \
  -t neuronip-api:local \
  ./api
```

#### Build with Network Mode

```bash
# Build with host network (for private registries)
docker build --network=host -t neuronip-api:local ./api
```

### Build Performance Optimization

**Enable BuildKit:**
```bash
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

**Use BuildKit cache mounts:**
```dockerfile
# In Dockerfile
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
```

**Parallel builds:**
```bash
# Build multiple images in parallel
docker compose build --parallel
```

**Build time comparison:**
- **Without cache:** ~10-15 minutes
- **With cache:** ~2-5 minutes
- **With BuildKit cache mounts:** ~1-3 minutes

---

## Verification

### Verify Image Contents

#### Inspect Image Metadata

**Basic inspection:**
```bash
docker inspect neuronip/api:v1.2.3
```

**Inspect specific fields:**
```bash
# Get image size
docker inspect -f '{{.Size}}' neuronip/api:v1.2.3

# Get creation date
docker inspect -f '{{.Created}}' neuronip/api:v1.2.3

# Get architecture
docker inspect -f '{{.Architecture}}' neuronip/api:v1.2.3

# Get environment variables
docker inspect -f '{{.Config.Env}}' neuronip/api:v1.2.3

# Get exposed ports
docker inspect -f '{{.Config.ExposedPorts}}' neuronip/api:v1.2.3
```

#### Check Image Size

**List all images:**
```bash
docker images neuronip/api
docker images neuronip/frontend
```

**Detailed size information:**
```bash
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" \
  | grep neuronip
```

**Compare image sizes:**
```bash
# Compare different tags
docker images neuronip/api --format "{{.Tag}}: {{.Size}}"

# Show size breakdown
docker system df -v
```

#### Analyze Image Layers

**View layer history:**
```bash
docker history neuronip/api:v1.2.3
```

**Detailed layer information:**
```bash
docker history --human --format "table {{.ID}}\t{{.CreatedBy}}\t{{.Size}}" \
  neuronip/api:v1.2.3
```

**Analyze layer sizes:**
```bash
# Show layer sizes
docker history --no-trunc neuronip/api:v1.2.3 | head -20
```

**Export and analyze:**
```bash
# Export image to tar
docker save neuronip/api:v1.2.3 -o neuronip-api.tar

# Analyze tar contents
tar -tf neuronip-api.tar | head -20
```

#### Verify Image Contents

**List files in image:**
```bash
docker run --rm neuronip/api:v1.2.3 ls -la /
```

**Check binary:**
```bash
docker run --rm neuronip/api:v1.2.3 file /root/neuronip-api
docker run --rm neuronip/api:v1.2.3 ldd /root/neuronip-api 2>&1 || echo "Static binary"
```

**Check frontend build:**
```bash
docker run --rm neuronip/frontend:v1.2.3 ls -la /app/.next
```

### Test Image Locally

#### Run API Container

**Basic run:**
```bash
docker run -p 8082:8082 \
  -e DB_HOST=postgres \
  -e DB_USER=neuronip \
  -e DB_PASSWORD=password \
  -e DB_NAME=neuronip \
  neuronip/api:v1.2.3
```

**Run with volume mounts:**
```bash
docker run -p 8082:8082 \
  -v $(pwd)/config:/config \
  -e DB_HOST=postgres \
  neuronip/api:v1.2.3
```

**Run with network:**
```bash
docker run -p 8082:8082 \
  --network neurondb-network \
  -e DB_HOST=postgres \
  neuronip/api:v1.2.3
```

**Run in background:**
```bash
docker run -d --name neuronip-api \
  -p 8082:8082 \
  -e DB_HOST=postgres \
  neuronip/api:v1.2.3
```

**Check logs:**
```bash
docker logs neuronip-api
docker logs -f neuronip-api  # Follow logs
```

#### Run Frontend Container

**Basic run:**
```bash
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8082/api/v1 \
  neuronip/frontend:v1.2.3
```

**Run with custom port:**
```bash
docker run -p 3001:3000 \
  -e PORT=3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8082/api/v1 \
  neuronip/frontend:v1.2.3
```

### Health Check Verification

**Test API health endpoint:**
```bash
# From host
curl http://localhost:8082/health

# From container
docker exec neuronip-api wget -qO- http://localhost:8082/health
```

**Test frontend:**
```bash
# Check if frontend is serving
curl http://localhost:3000

# Check API connectivity from frontend
docker exec neuronip-frontend wget -qO- http://neuronip-api:8082/health
```

### Security Scanning

**Scan for vulnerabilities:**
```bash
# Using Trivy
trivy image neuronip/api:v1.2.3

# Using Docker Scout
docker scout cves neuronip/api:v1.2.3

# Using Snyk
snyk test --docker neuronip/api:v1.2.3
```

**Check for secrets:**
```bash
# Using TruffleHog
trufflehog filesystem --directory=./api

# Using GitGuardian
ggshield secret scan docker neuronip/api:v1.2.3
```

### Performance Testing

**Measure container startup time:**
```bash
time docker run --rm neuronip/api:v1.2.3 --version
```

**Check resource usage:**
```bash
docker stats neuronip-api --no-stream
```

**Profile container:**
```bash
docker run --rm \
  --cap-add=SYS_PTRACE \
  -p 8082:8082 \
  neuronip/api:v1.2.3
```

---

## Troubleshooting

### Build Issues

#### Build Fails with "go: module not found"

**Problem:** Go module dependencies not found

**Solutions:**
```bash
# Verify go.mod exists
ls -la api/go.mod

# Download dependencies locally
cd api && go mod download

# Verify go.sum is up to date
go mod tidy

# Rebuild
docker build -t neuronip-api:local ./api
```

#### Build Fails with "npm ERR!"

**Problem:** npm install fails

**Solutions:**
```bash
# Clear npm cache
docker build --no-cache -t neuronip-frontend:local ./frontend

# Check package.json syntax
cd frontend && npm install --dry-run

# Verify package-lock.json
npm ci --dry-run
```

#### Build Takes Too Long

**Problem:** Builds are slow

**Solutions:**
```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Use cache from registry
docker build --cache-from neuronip/api:latest -t neuronip-api:local ./api

# Build in parallel
docker compose build --parallel

# Use BuildKit cache mounts (in Dockerfile)
RUN --mount=type=cache,target=/go/pkg/mod go mod download
```

#### Image Size Too Large

**Problem:** Final image is larger than expected

**Solutions:**
```bash
# Analyze image layers
docker history neuronip/api:v1.2.3

# Check for unnecessary files
docker run --rm neuronip/api:v1.2.3 du -sh /*

# Use multi-stage build (already implemented)
# Remove unnecessary COPY commands
# Use .dockerignore effectively
```

### Runtime Issues

#### Container Exits Immediately

**Problem:** Container starts then exits

**Solutions:**
```bash
# Check logs
docker logs <container-id>

# Run interactively
docker run -it neuronip/api:v1.2.3 sh

# Check if binary exists
docker run --rm neuronip/api:v1.2.3 ls -la /root/

# Verify binary is executable
docker run --rm neuronip/api:v1.2.3 file /root/neuronip-api
```

#### "Permission Denied" Errors

**Problem:** Binary cannot execute

**Solutions:**
```dockerfile
# In Dockerfile, ensure binary is executable
RUN chmod +x /root/neuronip-api

# Or build with proper permissions
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/neuronip-api ./cmd/server && \
    chmod +x /app/neuronip-api
```

#### Port Already in Use

**Problem:** Cannot bind to port 8082 or 3000

**Solutions:**
```bash
# Find process using port
lsof -i :8082
netstat -tulpn | grep 8082

# Use different port
docker run -p 8083:8082 neuronip/api:v1.2.3

# Stop conflicting container
docker stop <container-id>
```

#### Database Connection Errors

**Problem:** Cannot connect to external database

**Solutions:**
```bash
# Test network connectivity
docker run --rm --network neurondb-network \
  neuronip/api:v1.2.3 nc -zv postgres 5432

# Verify environment variables
docker run --rm neuronip/api:v1.2.3 env | grep DB_

# Check DNS resolution
docker run --rm --network neurondb-network \
  neuronip/api:v1.2.3 nslookup postgres
```

### CI/CD Issues

#### GitHub Actions Build Fails

**Problem:** Workflow fails during build

**Solutions:**
- Check workflow logs in GitHub Actions
- Verify secrets are configured:
  - `DOCKER_USERNAME`
  - `DOCKER_PASSWORD`
- Check Docker Hub rate limits
- Verify Dockerfile syntax
- Check for syntax errors in workflow YAML

#### Images Not Pushed to Registry

**Problem:** Build succeeds but images not in registry

**Solutions:**
```yaml
# Verify push is enabled
push: ${{ github.event_name != 'pull_request' }}

# Check authentication
- name: Login to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}

# Verify tags are correct
tags: neuronip/api:${{ github.sha }}
```

#### Release Workflow Not Triggered

**Problem:** Tag push doesn't trigger release

**Solutions:**
- Verify tag format: `v1.2.3` (not `1.2.3` or `v1.2`)
- Check workflow trigger: `tags: - 'v*.*.*'`
- Ensure tag is pushed: `git push origin v1.2.3`
- Check workflow permissions in repository settings

### Performance Issues

#### Slow Image Pulls

**Problem:** Pulling images takes too long

**Solutions:**
```bash
# Use image registry closer to deployment
# Use CDN for image distribution
# Compress images (already using Alpine)
# Use multi-stage builds (already implemented)

# Pull specific platform
docker pull --platform linux/amd64 neuronip/api:v1.2.3
```

#### High Memory Usage

**Problem:** Containers use too much memory

**Solutions:**
```bash
# Set memory limits
docker run -m 512m neuronip/api:v1.2.3

# Monitor memory usage
docker stats neuronip-api

# Optimize Go binary
# Use -ldflags="-s -w" to strip debug info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/neuronip-api ./cmd/server
```

#### Slow Container Startup

**Problem:** Containers take long to start

**Solutions:**
```bash
# Use smaller base images (already using Alpine)
# Minimize layers (combine RUN commands)
# Use health checks for faster readiness
# Pre-warm containers
```

---

## Advanced Optimizations

### Next.js Standalone Mode

**Enable in `next.config.js`:**
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
}

module.exports = nextConfig
```

**Update Dockerfile:**
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production

# Copy standalone build
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000
CMD ["node", "server.js"]
```

**Benefits:**
- Reduces `node_modules` size by 50-70%
- Only includes necessary dependencies
- Faster container startup
- Smaller image size

### BuildKit Cache Mounts

**For Go dependencies:**
```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/neuronip-api ./cmd/server
```

**For npm dependencies:**
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build
```

**Benefits:**
- Faster builds (cached dependencies)
- Persistent cache across builds
- Reduced network usage

### Multi-Stage Build Optimization

**Combine RUN commands:**
```dockerfile
# Before (multiple layers)
RUN apk update
RUN apk add ca-certificates
RUN rm -rf /var/cache/apk/*

# After (single layer)
RUN apk --no-cache add ca-certificates
```

**Use .dockerignore effectively:**
```
# Exclude unnecessary files
node_modules/
.git/
*.md
.env*
.DS_Store
coverage/
.nyc_output/
```

### Security Hardening

**Use non-root user:**
```dockerfile
FROM alpine:latest
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
RUN apk --no-cache add ca-certificates
WORKDIR /home/appuser
COPY --from=builder /app/neuronip-api .
RUN chown appuser:appuser /home/appuser/neuronip-api
USER appuser
CMD ["./neuronip-api"]
```

**Scan for vulnerabilities:**
```bash
# Regular scanning
trivy image neuronip/api:v1.2.3

# In CI/CD
- name: Scan image
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: neuronip/api:${{ github.sha }}
    format: 'sarif'
    output: 'trivy-results.sarif'
```

**Use distroless images (advanced):**
```dockerfile
FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/neuronip-api /neuronip-api
EXPOSE 8082
CMD ["/neuronip-api"]
```

**Benefits:**
- Even smaller attack surface
- No shell, no package manager
- Only application binary

---

## 📚 Related Documentation

- [Docker Deployment](docker.md) - Docker setup and configuration
- [Production Deployment](production.md) - Production deployment guide
- [Kubernetes Deployment](kubernetes.md) - Kubernetes deployment guide
- [CI/CD Reference](../../.github/workflows/) - GitHub Actions workflows

---

<div align="center">

[← Back to Documentation](../README.md)

</div>
