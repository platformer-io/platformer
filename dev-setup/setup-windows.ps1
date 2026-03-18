# =============================================================================
# PlatFormer - Dev Environment Setup (Windows)
# =============================================================================
# Run with PowerShell as Administrator:
#   Set-ExecutionPolicy Bypass -Scope Process -Force
#   .\setup-windows.ps1
# =============================================================================

$ErrorActionPreference = "Stop"

# Colors
function Write-Step($num, $msg)  { Write-Host "`n[$num] $msg" -ForegroundColor Blue }
function Write-Ok($msg)          { Write-Host "  [OK]  $msg" -ForegroundColor Green }
function Write-Skip($msg)        { Write-Host "  [!!]  Skipped: $msg (already installed)" -ForegroundColor Yellow }
function Write-Installing($msg)  { Write-Host "  -->  Installing $msg..." -ForegroundColor Cyan }
function Write-Note($msg)        { Write-Host "  [i]  $msg" -ForegroundColor Gray }

function Command-Exists($cmd) {
    return [bool](Get-Command $cmd -ErrorAction SilentlyContinue)
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Blue
Write-Host "   PlatFormer Dev Environment Setup (Windows)" -ForegroundColor Blue
Write-Host "=============================================" -ForegroundColor Blue
Write-Host ""

# Check for admin
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "  [XX]  Please run this script as Administrator (right-click PowerShell --> Run as administrator)" -ForegroundColor Red
    exit 1
}

# -----------------------------------------------------------------------------
# 1. WINGET CHECK
# -----------------------------------------------------------------------------

Write-Step "1/9" "Checking winget (Windows Package Manager)"

if (-not (Command-Exists winget)) {
    Write-Host "  [XX]  winget not found. Install 'App Installer' from the Microsoft Store first." -ForegroundColor Red
    Write-Host "     https://apps.microsoft.com/detail/9NBLGGH4NNS1" -ForegroundColor Gray
    exit 1
} else {
    Write-Ok "winget available"
}

# -----------------------------------------------------------------------------
# 2. GIT
# -----------------------------------------------------------------------------

Write-Step "2/9" "Git"

if (-not (Command-Exists git)) {
    Write-Installing "Git"
    winget install --id Git.Git -e --source winget --silent
    # Refresh PATH
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "Git installed"
} else {
    Write-Skip "Git ($(git --version))"
}

# -----------------------------------------------------------------------------
# 3. GO
# -----------------------------------------------------------------------------

Write-Step "3/9" "Go 1.21+"

if (-not (Command-Exists go)) {
    Write-Installing "Go"
    winget install --id GoLang.Go -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    $env:GOPATH = "$env:USERPROFILE\go"
    $env:PATH += ";$env:GOPATH\bin"
    Write-Ok "Go installed"
    Write-Note "You may need to restart your terminal for Go PATH to take effect"
} else {
    Write-Skip "Go ($(go version))"
}

# Set GOPATH if not set
if (-not $env:GOPATH) {
    $env:GOPATH = "$env:USERPROFILE\go"
    [System.Environment]::SetEnvironmentVariable("GOPATH", $env:GOPATH, "User")
}
$goBin = "$env:GOPATH\bin"
if ($env:PATH -notlike "*$goBin*") {
    [System.Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$goBin", "User")
    $env:PATH += ";$goBin"
}

# -----------------------------------------------------------------------------
# 4. NODE.JS 20 LTS
# -----------------------------------------------------------------------------

Write-Step "4/9" "Node.js 20 LTS"

if (-not (Command-Exists node)) {
    Write-Installing "Node.js 20 LTS"
    winget install --id OpenJS.NodeJS.LTS -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "Node.js installed"
} else {
    $nodeVersion = node --version
    Write-Skip "Node.js ($nodeVersion)"
}

# -----------------------------------------------------------------------------
# 5. PNPM
# -----------------------------------------------------------------------------

Write-Step "5/9" "pnpm (frontend package manager)"

if (-not (Command-Exists pnpm)) {
    Write-Installing "pnpm"
    npm install -g pnpm
    Write-Ok "pnpm installed"
} else {
    Write-Skip "pnpm ($(pnpm --version))"
}

# -----------------------------------------------------------------------------
# 6. DOCKER DESKTOP
# -----------------------------------------------------------------------------

Write-Step "6/9" "Docker Desktop"

if (-not (Command-Exists docker)) {
    Write-Installing "Docker Desktop"
    winget install --id Docker.DockerDesktop -e --source winget --silent
    Write-Ok "Docker Desktop installed"
    Write-Host ""
    Write-Host "  [!!]  IMPORTANT: Docker Desktop requires a restart and manual setup:" -ForegroundColor Yellow
    Write-Host "     1. Restart your computer after this script finishes" -ForegroundColor Yellow
    Write-Host "     2. Open Docker Desktop and complete the setup wizard" -ForegroundColor Yellow
    Write-Host "     3. Enable WSL 2 backend in Docker Desktop settings" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Skip "Docker ($(docker --version))"
}

# -----------------------------------------------------------------------------
# 7. KUBECTL
# -----------------------------------------------------------------------------

Write-Step "7/9" "kubectl"

if (-not (Command-Exists kubectl)) {
    Write-Installing "kubectl"
    winget install --id Kubernetes.kubectl -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "kubectl installed"
} else {
    Write-Skip "kubectl"
}

# Install kind
if (-not (Command-Exists kind)) {
    Write-Installing "kind (Kubernetes in Docker)"
    $kindVersion = "v0.22.0"
    $kindUrl = "https://kind.sigs.k8s.io/dl/$kindVersion/kind-windows-amd64"
    $kindDest = "$env:USERPROFILE\AppData\Local\Microsoft\WinGet\Packages\kind.exe"
    New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.local\bin" | Out-Null
    Invoke-WebRequest -Uri $kindUrl -OutFile "$env:USERPROFILE\.local\bin\kind.exe"
    $localBin = "$env:USERPROFILE\.local\bin"
    if ($env:PATH -notlike "*$localBin*") {
        [System.Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$localBin", "User")
        $env:PATH += ";$localBin"
    }
    Write-Ok "kind installed"
} else {
    Write-Skip "kind"
}

# -----------------------------------------------------------------------------
# 8. KUBEBUILDER
# -----------------------------------------------------------------------------

Write-Step "8/9" "kubebuilder (operator scaffolding)"

if (-not (Command-Exists kubebuilder)) {
    Write-Installing "kubebuilder"
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $kbVersion = "v3.14.0"
    $kbUrl = "https://github.com/kubernetes-sigs/kubebuilder/releases/download/$kbVersion/kubebuilder_windows_amd64.exe"
    $kbDest = "$env:USERPROFILE\.local\bin\kubebuilder.exe"
    New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.local\bin" | Out-Null
    Invoke-WebRequest -Uri $kbUrl -OutFile $kbDest -UseBasicParsing
    Write-Ok "kubebuilder installed"
} else {
    Write-Skip "kubebuilder"
}

# -----------------------------------------------------------------------------
# 9. AWS CLI v2
# -----------------------------------------------------------------------------

Write-Step "9/9" "AWS CLI v2"

if (-not (Command-Exists aws)) {
    Write-Installing "AWS CLI v2"
    winget install --id Amazon.AWSCLI -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "AWS CLI installed"
} else {
    Write-Skip "AWS CLI ($(aws --version))"
}

# -----------------------------------------------------------------------------
# CLAUDE CODE
# -----------------------------------------------------------------------------

Write-Step "+" "Claude Code"

if (-not (Command-Exists claude)) {
    Write-Installing "Claude Code"
    npm install -g @anthropic-ai/claude-code
    Write-Ok "Claude Code installed"
} else {
    Write-Skip "Claude Code"
}

# -----------------------------------------------------------------------------
# GO TOOLS
# -----------------------------------------------------------------------------

Write-Step "+" "Go development tools"

Write-Installing "controller-gen"
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
Write-Ok "controller-gen installed"

Write-Installing "goimports"
go install golang.org/x/tools/cmd/goimports@latest
Write-Ok "goimports installed"

# golangci-lint for Windows
Write-Installing "golangci-lint"
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | Out-Null
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
Write-Ok "golangci-lint installed"

# -----------------------------------------------------------------------------
# NOTION MCP CONFIG
# -----------------------------------------------------------------------------

Write-Step "+" "Notion MCP server for Claude Code"

$claudeConfigDir = "$env:USERPROFILE\.claude"
$claudeConfigFile = "$claudeConfigDir\claude_desktop_config.json"

if (-not (Test-Path $claudeConfigDir)) {
    New-Item -ItemType Directory -Force -Path $claudeConfigDir | Out-Null
}

if (-not (Test-Path $claudeConfigFile)) {
    $config = @{
        mcpServers = @{
            notion = @{
                command = "npx"
                args = @("-y", "@notionhq/notion-mcp-server")
                env = @{
                    OPENAPI_MCP_HEADERS = '{"Authorization": "Bearer YOUR_NOTION_TOKEN", "Notion-Version": "2022-06-28"}'
                }
            }
        }
    } | ConvertTo-Json -Depth 10

    Set-Content -Path $claudeConfigFile -Value $config
    Write-Ok "Created $claudeConfigFile"
    Write-Note "Replace YOUR_NOTION_TOKEN with your token from https://notion.so/my-integrations"
} else {
    Write-Skip "Claude config file (already exists at $claudeConfigFile)"
    Write-Note "Manually add Notion MCP server to $claudeConfigFile if not already present"
}

# -----------------------------------------------------------------------------
# LOCAL KIND CLUSTER
# -----------------------------------------------------------------------------

Write-Step "+" "Creating local kind cluster"

if (Command-Exists kind) {
    try {
        $clusters = kind get clusters 2>&1
        if ($clusters -notlike "*platformer-dev*") {
            Write-Installing "kind cluster: platformer-dev"
            kind create cluster --name platformer-dev
            Write-Ok "kind cluster 'platformer-dev' created"
        } else {
            Write-Skip "kind cluster 'platformer-dev' (already exists)"
        }
    } catch {
        Write-Host "  [!!]  Could not create kind cluster - make sure Docker Desktop is running" -ForegroundColor Yellow
        Write-Note "Run manually after starting Docker: kind create cluster --name platformer-dev"
    }
}

# -----------------------------------------------------------------------------
# SUMMARY
# -----------------------------------------------------------------------------

Write-Host ""
Write-Host "=============================================" -ForegroundColor Green
Write-Host "   Setup Complete!" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""
Write-Host "Installed tools:" -ForegroundColor White

if (Command-Exists go)          { Write-Ok "Go:           $(go version)" }
if (Command-Exists node)        { Write-Ok "Node.js:      $(node --version)" }
if (Command-Exists pnpm)        { Write-Ok "pnpm:         $(pnpm --version)" }
if (Command-Exists docker)      { Write-Ok "Docker:       $(docker --version)" }
if (Command-Exists kubectl)     { Write-Ok "kubectl:      installed" }
if (Command-Exists kind)        { Write-Ok "kind:         installed" }
if (Command-Exists kubebuilder) { Write-Ok "kubebuilder:  installed" }
if (Command-Exists aws)         { Write-Ok "AWS CLI:      $(aws --version)" }
if (Command-Exists claude)      { Write-Ok "Claude Code:  installed" }

Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "  1. Restart your terminal (or computer if Docker was just installed)"
Write-Host "  2. Configure AWS:     aws configure"
Write-Host "  3. Update Notion MCP: edit $env:USERPROFILE\.claude\claude_desktop_config.json"
Write-Host "  4. Copy CLAUDE.md:    copy CLAUDE.md to your project root"
Write-Host "  5. Initialize repo:   cd ~/projects/platformer; go mod init github.com/platformer-io/platformer"
Write-Host "  6. Start coding:      claude"
Write-Host ""
# =============================================================================
# PlatFormer - Dev Environment Setup (Windows)
# =============================================================================
# Run with PowerShell as Administrator:
#   Set-ExecutionPolicy Bypass -Scope Process -Force
#   .\setup-windows.ps1
# =============================================================================

$ErrorActionPreference = "Stop"

# Colors
function Write-Step($num, $msg)  { Write-Host "`n[$num] $msg" -ForegroundColor Blue }
function Write-Ok($msg)          { Write-Host "  [OK]  $msg" -ForegroundColor Green }
function Write-Skip($msg)        { Write-Host "  [!!]  Skipped: $msg (already installed)" -ForegroundColor Yellow }
function Write-Installing($msg)  { Write-Host "  -->  Installing $msg..." -ForegroundColor Cyan }
function Write-Note($msg)        { Write-Host "  [i]  $msg" -ForegroundColor Gray }

function Command-Exists($cmd) {
    return [bool](Get-Command $cmd -ErrorAction SilentlyContinue)
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Blue
Write-Host "   PlatFormer Dev Environment Setup (Windows)" -ForegroundColor Blue
Write-Host "=============================================" -ForegroundColor Blue
Write-Host ""

# Check for admin
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "  [XX]  Please run this script as Administrator (right-click PowerShell --> Run as administrator)" -ForegroundColor Red
    exit 1
}

# -----------------------------------------------------------------------------
# 1. WINGET CHECK
# -----------------------------------------------------------------------------

Write-Step "1/9" "Checking winget (Windows Package Manager)"

if (-not (Command-Exists winget)) {
    Write-Host "  [XX]  winget not found. Install 'App Installer' from the Microsoft Store first." -ForegroundColor Red
    Write-Host "     https://apps.microsoft.com/detail/9NBLGGH4NNS1" -ForegroundColor Gray
    exit 1
} else {
    Write-Ok "winget available"
}

# -----------------------------------------------------------------------------
# 2. GIT
# -----------------------------------------------------------------------------

Write-Step "2/9" "Git"

if (-not (Command-Exists git)) {
    Write-Installing "Git"
    winget install --id Git.Git -e --source winget --silent
    # Refresh PATH
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "Git installed"
} else {
    Write-Skip "Git ($(git --version))"
}

# -----------------------------------------------------------------------------
# 3. GO
# -----------------------------------------------------------------------------

Write-Step "3/9" "Go 1.21+"

if (-not (Command-Exists go)) {
    Write-Installing "Go"
    winget install --id GoLang.Go -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    $env:GOPATH = "$env:USERPROFILE\go"
    $env:PATH += ";$env:GOPATH\bin"
    Write-Ok "Go installed"
    Write-Note "You may need to restart your terminal for Go PATH to take effect"
} else {
    Write-Skip "Go ($(go version))"
}

# Set GOPATH if not set
if (-not $env:GOPATH) {
    $env:GOPATH = "$env:USERPROFILE\go"
    [System.Environment]::SetEnvironmentVariable("GOPATH", $env:GOPATH, "User")
}
$goBin = "$env:GOPATH\bin"
if ($env:PATH -notlike "*$goBin*") {
    [System.Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$goBin", "User")
    $env:PATH += ";$goBin"
}

# -----------------------------------------------------------------------------
# 4. NODE.JS 20 LTS
# -----------------------------------------------------------------------------

Write-Step "4/9" "Node.js 20 LTS"

if (-not (Command-Exists node)) {
    Write-Installing "Node.js 20 LTS"
    winget install --id OpenJS.NodeJS.LTS -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "Node.js installed"
} else {
    $nodeVersion = node --version
    Write-Skip "Node.js ($nodeVersion)"
}

# -----------------------------------------------------------------------------
# 5. PNPM
# -----------------------------------------------------------------------------

Write-Step "5/9" "pnpm (frontend package manager)"

if (-not (Command-Exists pnpm)) {
    Write-Installing "pnpm"
    npm install -g pnpm
    Write-Ok "pnpm installed"
} else {
    Write-Skip "pnpm ($(pnpm --version))"
}

# -----------------------------------------------------------------------------
# 6. DOCKER DESKTOP
# -----------------------------------------------------------------------------

Write-Step "6/9" "Docker Desktop"

if (-not (Command-Exists docker)) {
    Write-Installing "Docker Desktop"
    winget install --id Docker.DockerDesktop -e --source winget --silent
    Write-Ok "Docker Desktop installed"
    Write-Host ""
    Write-Host "  [!!]  IMPORTANT: Docker Desktop requires a restart and manual setup:" -ForegroundColor Yellow
    Write-Host "     1. Restart your computer after this script finishes" -ForegroundColor Yellow
    Write-Host "     2. Open Docker Desktop and complete the setup wizard" -ForegroundColor Yellow
    Write-Host "     3. Enable WSL 2 backend in Docker Desktop settings" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Skip "Docker ($(docker --version))"
}

# -----------------------------------------------------------------------------
# 7. KUBECTL
# -----------------------------------------------------------------------------

Write-Step "7/9" "kubectl"

if (-not (Command-Exists kubectl)) {
    Write-Installing "kubectl"
    winget install --id Kubernetes.kubectl -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "kubectl installed"
} else {
    Write-Skip "kubectl"
}

# Install kind
if (-not (Command-Exists kind)) {
    Write-Installing "kind (Kubernetes in Docker)"
    $kindVersion = "v0.22.0"
    $kindUrl = "https://kind.sigs.k8s.io/dl/$kindVersion/kind-windows-amd64"
    $kindDest = "$env:USERPROFILE\AppData\Local\Microsoft\WinGet\Packages\kind.exe"
    New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.local\bin" | Out-Null
    Invoke-WebRequest -Uri $kindUrl -OutFile "$env:USERPROFILE\.local\bin\kind.exe"
    $localBin = "$env:USERPROFILE\.local\bin"
    if ($env:PATH -notlike "*$localBin*") {
        [System.Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$localBin", "User")
        $env:PATH += ";$localBin"
    }
    Write-Ok "kind installed"
} else {
    Write-Skip "kind"
}

# -----------------------------------------------------------------------------
# 8. KUBEBUILDER
# -----------------------------------------------------------------------------

Write-Step "8/9" "kubebuilder (operator scaffolding)"

if (-not (Command-Exists kubebuilder)) {
    Write-Installing "kubebuilder"
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $kbVersion = "v3.14.0"
    $kbUrl = "https://github.com/kubernetes-sigs/kubebuilder/releases/download/$kbVersion/kubebuilder_windows_amd64.exe"
    $kbDest = "$env:USERPROFILE\.local\bin\kubebuilder.exe"
    New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.local\bin" | Out-Null
    Invoke-WebRequest -Uri $kbUrl -OutFile $kbDest -UseBasicParsing
    Write-Ok "kubebuilder installed"
} else {
    Write-Skip "kubebuilder"
}

# -----------------------------------------------------------------------------
# 9. AWS CLI v2
# -----------------------------------------------------------------------------

Write-Step "9/9" "AWS CLI v2"

if (-not (Command-Exists aws)) {
    Write-Installing "AWS CLI v2"
    winget install --id Amazon.AWSCLI -e --source winget --silent
    $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
    Write-Ok "AWS CLI installed"
} else {
    Write-Skip "AWS CLI ($(aws --version))"
}

# -----------------------------------------------------------------------------
# CLAUDE CODE
# -----------------------------------------------------------------------------

Write-Step "+" "Claude Code"

if (-not (Command-Exists claude)) {
    Write-Installing "Claude Code"
    npm install -g @anthropic-ai/claude-code
    Write-Ok "Claude Code installed"
} else {
    Write-Skip "Claude Code"
}

# -----------------------------------------------------------------------------
# GO TOOLS
# -----------------------------------------------------------------------------

Write-Step "+" "Go development tools"

Write-Installing "controller-gen"
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
Write-Ok "controller-gen installed"

Write-Installing "goimports"
go install golang.org/x/tools/cmd/goimports@latest
Write-Ok "goimports installed"

# golangci-lint for Windows
Write-Installing "golangci-lint"
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | Out-Null
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
Write-Ok "golangci-lint installed"

# -----------------------------------------------------------------------------
# NOTION MCP CONFIG
# -----------------------------------------------------------------------------

Write-Step "+" "Notion MCP server for Claude Code"

$claudeConfigDir = "$env:USERPROFILE\.claude"
$claudeConfigFile = "$claudeConfigDir\claude_desktop_config.json"

if (-not (Test-Path $claudeConfigDir)) {
    New-Item -ItemType Directory -Force -Path $claudeConfigDir | Out-Null
}

if (-not (Test-Path $claudeConfigFile)) {
    $config = @{
        mcpServers = @{
            notion = @{
                command = "npx"
                args = @("-y", "@notionhq/notion-mcp-server")
                env = @{
                    OPENAPI_MCP_HEADERS = '{"Authorization": "Bearer YOUR_NOTION_TOKEN", "Notion-Version": "2022-06-28"}'
                }
            }
        }
    } | ConvertTo-Json -Depth 10

    Set-Content -Path $claudeConfigFile -Value $config
    Write-Ok "Created $claudeConfigFile"
    Write-Note "Replace YOUR_NOTION_TOKEN with your token from https://notion.so/my-integrations"
} else {
    Write-Skip "Claude config file (already exists at $claudeConfigFile)"
    Write-Note "Manually add Notion MCP server to $claudeConfigFile if not already present"
}

# -----------------------------------------------------------------------------
# LOCAL KIND CLUSTER
# -----------------------------------------------------------------------------

Write-Step "+" "Creating local kind cluster"

if (Command-Exists kind) {
    try {
        $clusters = kind get clusters 2>&1
        if ($clusters -notlike "*platformer-dev*") {
            Write-Installing "kind cluster: platformer-dev"
            kind create cluster --name platformer-dev
            Write-Ok "kind cluster 'platformer-dev' created"
        } else {
            Write-Skip "kind cluster 'platformer-dev' (already exists)"
        }
    } catch {
        Write-Host "  [!!]  Could not create kind cluster - make sure Docker Desktop is running" -ForegroundColor Yellow
        Write-Note "Run manually after starting Docker: kind create cluster --name platformer-dev"
    }
}

# -----------------------------------------------------------------------------
# SUMMARY
# -----------------------------------------------------------------------------

Write-Host ""
Write-Host "=============================================" -ForegroundColor Green
Write-Host "   Setup Complete!" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""
Write-Host "Installed tools:" -ForegroundColor White

if (Command-Exists go)          { Write-Ok "Go:           $(go version)" }
if (Command-Exists node)        { Write-Ok "Node.js:      $(node --version)" }
if (Command-Exists pnpm)        { Write-Ok "pnpm:         $(pnpm --version)" }
if (Command-Exists docker)      { Write-Ok "Docker:       $(docker --version)" }
if (Command-Exists kubectl)     { Write-Ok "kubectl:      installed" }
if (Command-Exists kind)        { Write-Ok "kind:         installed" }
if (Command-Exists kubebuilder) { Write-Ok "kubebuilder:  installed" }
if (Command-Exists aws)         { Write-Ok "AWS CLI:      $(aws --version)" }
if (Command-Exists claude)      { Write-Ok "Claude Code:  installed" }

Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "  1. Restart your terminal (or computer if Docker was just installed)"
Write-Host "  2. Configure AWS:     aws configure"
Write-Host "  3. Update Notion MCP: edit $env:USERPROFILE\.claude\claude_desktop_config.json"
Write-Host "  4. Copy CLAUDE.md:    copy CLAUDE.md to your project root"
Write-Host "  5. Initialize repo:   cd ~/projects/platformer; go mod init github.com/platformer-io/platformer"
Write-Host "  6. Start coding:      claude"
Write-Host ""
