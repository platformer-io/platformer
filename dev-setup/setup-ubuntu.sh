#!/bin/bash

# =============================================================================
# PlatFormer - Dev Environment Setup (Ubuntu / WSL2)
# =============================================================================
# Run with: chmod +x setup-ubuntu.sh && ./setup-ubuntu.sh
# Tested on: Ubuntu 22.04 LTS, Ubuntu 24.04 LTS, WSL2
#
# WSL2 note:
#   Docker Desktop on Windows exposes Docker to WSL2 automatically.
#   If running native Ubuntu, Docker Engine is installed directly.
# =============================================================================

set -e

GREEN='\033[0;32m' ; BLUE='\033[0;34m' ; YELLOW='\033[1;33m' ; CYAN='\033[0;36m' ; GRAY='\033[0;37m' ; NC='\033[0m'
print_step()       { echo -e "\n${BLUE}[$1]${NC} $2"; }
print_ok()         { echo -e "  ${GREEN}[OK]${NC}   $1"; }
print_skip()       { echo -e "  ${YELLOW}[!!]${NC}   Skipped: $1 (already installed)"; }
print_installing() { echo -e "  ${CYAN}[-->]${NC}  Installing $1..."; }
print_note()       { echo -e "  ${GRAY}[i]${NC}    $1"; }
print_warn()       { echo -e "  ${YELLOW}[!!]${NC}   $1"; }
command_exists()   { command -v "$1" &>/dev/null; }
is_wsl()           { grep -qEi "(microsoft|wsl)" /proc/version 2>/dev/null; }

ARCH=$(dpkg --print-architecture)

echo -e "${BLUE}==============================================${NC}"
echo    "   PlatFormer Dev Environment Setup"
is_wsl && echo "   Environment: WSL2" || echo "   Environment: Native Ubuntu"
echo -e "${BLUE}==============================================${NC}" ; echo ""

# 1. System deps
print_step "1/10" "System dependencies"
sudo apt-get update -qq
sudo apt-get install -y -qq curl wget git make build-essential \
    apt-transport-https ca-certificates gnupg lsb-release \
    software-properties-common unzip jq xdg-utils
print_ok "System dependencies ready"

# 2. Go 1.22
print_step "2/10" "Go 1.22"
GO_VERSION="1.22.3"
if command_exists go; then
    print_skip "Go ($(go version | awk '{print $3}'))"
else
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" -O /tmp/go.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz
    for RC in "$HOME/.bashrc" "$HOME/.zshrc"; do
        [ -f "$RC" ] && ! grep -q "/usr/local/go/bin" "$RC" && {
            echo '' >> "$RC"
            echo '# Go' >> "$RC"
            echo 'export PATH=$PATH:/usr/local/go/bin' >> "$RC"
            echo 'export GOPATH=$HOME/go' >> "$RC"
            echo 'export PATH=$PATH:$GOPATH/bin' >> "$RC"
        }
    done
    export PATH=$PATH:/usr/local/go/bin ; export GOPATH=$HOME/go ; export PATH=$PATH:$GOPATH/bin
    print_ok "Go $GO_VERSION installed" ; print_note "Run: source ~/.bashrc"
fi
export GOPATH="${GOPATH:-$HOME/go}" ; export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin ; mkdir -p "$GOPATH/bin"

# 3. Node.js 20 LTS via nvm
print_step "3/10" "Node.js 20 LTS (via nvm)"
export NVM_DIR="$HOME/.nvm"
[ ! -d "$NVM_DIR" ] && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
if command_exists node; then print_skip "Node.js ($(node --version))"
else nvm install 20 && nvm use 20 && nvm alias default 20 && print_ok "Node.js $(node --version) installed"; fi

# 4. pnpm
print_step "4/10" "pnpm"
if command_exists pnpm; then print_skip "pnpm ($(pnpm --version))"
else npm install -g pnpm && print_ok "pnpm installed"; fi

# 5. Docker
print_step "5/10" "Docker"
if is_wsl; then
    if command_exists docker && docker info &>/dev/null 2>&1; then
        print_skip "Docker (WSL2 - Docker Desktop)"
    else
        print_warn "Docker not accessible in WSL2. Fix:"
        print_note "  1. Open Docker Desktop on Windows"
        print_note "  2. Settings → Resources → WSL Integration"
        print_note "  3. Enable your Ubuntu distro → Apply & Restart"
        print_note "  4. In PowerShell: wsl --shutdown (then reopen)"
    fi
else
    if command_exists docker; then print_skip "Docker"
    else
        for pkg in docker docker-engine docker.io containerd runc; do sudo apt-get remove -y -qq "$pkg" 2>/dev/null || true; done
        sudo install -m 0755 -d /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        sudo chmod a+r /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        sudo apt-get update -qq && sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        sudo usermod -aG docker "$USER" && sudo systemctl enable docker && sudo systemctl start docker
        print_ok "Docker Engine installed" ; print_warn "Run: newgrp docker (or log out/in)"
    fi
fi

# 6. kubectl
print_step "6/10" "kubectl"
if command_exists kubectl; then print_skip "kubectl"
else
    KUBECTL_VERSION=$(curl -sL https://dl.k8s.io/release/stable.txt)
    curl -sLO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl"
    curl -sLO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl.sha256"
    echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check --quiet
    sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && rm kubectl kubectl.sha256
    print_ok "kubectl $KUBECTL_VERSION installed"
fi

# 7. kind
print_step "7/10" "kind"
if command_exists kind; then print_skip "kind ($(kind version | awk '{print $2}'))"
else
    KIND_VERSION=$(curl -s https://api.github.com/repos/kubernetes-sigs/kind/releases/latest | jq -r '.tag_name')
    curl -sLo /tmp/kind "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-${ARCH}"
    chmod +x /tmp/kind && sudo mv /tmp/kind /usr/local/bin/kind
    print_ok "kind $KIND_VERSION installed"
fi

# 8. kubebuilder
print_step "8/10" "kubebuilder"
if command_exists kubebuilder; then print_skip "kubebuilder"
else
    KB_VERSION=$(curl -s https://api.github.com/repos/kubernetes-sigs/kubebuilder/releases/latest | jq -r '.tag_name')
    curl -sL "https://github.com/kubernetes-sigs/kubebuilder/releases/download/${KB_VERSION}/kubebuilder_linux_$(go env GOARCH)" -o /tmp/kubebuilder
    chmod +x /tmp/kubebuilder && sudo mv /tmp/kubebuilder /usr/local/bin/kubebuilder
    print_ok "kubebuilder $KB_VERSION installed"
fi

# 9. AWS CLI v2
print_step "9/10" "AWS CLI v2"
if command_exists aws; then print_skip "AWS CLI"
else
    [ "$ARCH" = "arm64" ] && AWS_URL="https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" || AWS_URL="https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip"
    curl -s "$AWS_URL" -o /tmp/awscliv2.zip && unzip -q /tmp/awscliv2.zip -d /tmp/aws-install
    sudo /tmp/aws-install/aws/install && rm -rf /tmp/awscliv2.zip /tmp/aws-install
    print_ok "AWS CLI installed"
fi

# 10. Claude Code
print_step "10/10" "Claude Code"
if command_exists claude; then print_skip "Claude Code"
else npm install -g @anthropic-ai/claude-code && print_ok "Claude Code installed"; fi

# Go dev tools
print_step "+" "Go development tools"
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest && print_ok "controller-gen installed"
go install golang.org/x/tools/cmd/goimports@latest && print_ok "goimports installed"
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" latest && print_ok "golangci-lint installed"

# Notion MCP config
print_step "+" "Notion MCP config"
CLAUDE_CONFIG="$HOME/.claude/claude_desktop_config.json"
mkdir -p "$HOME/.claude"
if [ ! -f "$CLAUDE_CONFIG" ]; then
    cat > "$CLAUDE_CONFIG" << 'MCPEOF'
{
  "mcpServers": {
    "notion": {
      "command": "npx",
      "args": ["-y", "@notionhq/notion-mcp-server"],
      "env": {
        "OPENAPI_MCP_HEADERS": "{\"Authorization\": \"Bearer YOUR_NOTION_TOKEN\", \"Notion-Version\": \"2022-06-28\"}"
      }
    }
  }
}
MCPEOF
    print_ok "Created $CLAUDE_CONFIG" ; print_note "Replace YOUR_NOTION_TOKEN → https://notion.so/my-integrations"
else print_skip "Claude config (already exists)"; fi

# kind cluster
print_step "+" "Local kind cluster: platformer-dev"
if command_exists kind && docker info &>/dev/null 2>&1; then
    if kind get clusters 2>/dev/null | grep -q "platformer-dev"; then print_skip "kind cluster 'platformer-dev'"
    else kind create cluster --name platformer-dev && print_ok "kind cluster created"; fi
else print_warn "Docker not running — run manually: kind create cluster --name platformer-dev"; fi

# VSCode extensions (WSL2 only)
if is_wsl && command_exists code; then
    print_step "+" "VSCode extensions"
    for EXT in ms-vscode-remote.remote-wsl golang.go ms-kubernetes-tools.vscode-kubernetes-tools ms-azuretools.vscode-docker redhat.vscode-yaml eamodio.gitlens; do
        code --install-extension "$EXT" --force &>/dev/null && print_ok "VSCode: $EXT" || print_note "Install manually: $EXT"
    done
fi

# Summary
echo "" ; echo -e "${GREEN}==============================================${NC}"
echo -e "${GREEN}   Setup Complete!${NC}" ; echo -e "${GREEN}==============================================${NC}" ; echo ""
command_exists go          && print_ok "Go:           $(go version | awk '{print $3}')"
command_exists node        && print_ok "Node.js:      $(node --version)"
command_exists pnpm        && print_ok "pnpm:         v$(pnpm --version)"
command_exists docker      && print_ok "Docker:       installed"
command_exists kubectl     && print_ok "kubectl:      installed"
command_exists kind        && print_ok "kind:         $(kind version | awk '{print $2}')"
command_exists kubebuilder && print_ok "kubebuilder:  installed"
command_exists aws         && print_ok "AWS CLI:      installed"
command_exists claude      && print_ok "Claude Code:  installed"
echo ""
echo "Next: source ~/.bashrc | aws configure | add Notion token | copy CLAUDE.md to repo root | claude"
is_wsl && echo "" && echo "  WSL2: open project with 'code .' | files at \\\\wsl\$\\Ubuntu\\home\\$USER"