# cb-ask

[![npm version](https://img.shields.io/npm/v/cb-ask.svg?style=flat-square&color=2563eb)](https://www.npmjs.com/package/cb-ask)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25-00ADD8.svg?style=flat-square&logo=go)](https://golang.org/)
[![Node.js](https://img.shields.io/badge/Node.js-%3E%3D18.0.0-brightgreen.svg?style=flat-square&logo=node.js)](https://nodejs.org/)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-ea580c.svg?style=flat-square)](https://modelcontextprotocol.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-10b981.svg?style=flat-square)](LICENSE)

Agent-facing query engine and local Stdio bridge connecting AI coding assistants (Cursor, Claude, VS Code, etc.) to your central [`cb-indexer`](https://github.com/Abbilville/cb-indexer) codebase knowledge graph.

---

## Overview

AI coding assistants typically operate with single-directory blinders. In polyrepo or microservice architectures, they lack cross-boundary context:
- They don't know which microservice handles which API route.
- They don't know runtime port assignments or inter-service dependencies.
- They resort to blind whole-workspace text searches that burn context tokens and miss dependencies.

`cb-ask` solves this. It runs locally on your machine over standard input/output (stdio), spawned directly by your AI coding assistant. Whenever your assistant needs to understand service architecture, trace cross-service requests, or find code definitions, `cb-ask` queries your team's remote (or local) `cb-indexer` server and supplies structured context directly into the conversation.

---

## Quick Start (Zero Install)

In any project workspace terminal, run:

```bash
npx cb-ask
```

The interactive terminal wizard will guide you through setup:

```text
cb-ask — Codebase RAG & Knowledge Graph Agent Setup
───────────────────────────────────────────────────

?  cb-indexer location:
   (•) Localhost (http://127.0.0.1:43770)
   ( ) Remote server (custom URL)

?  Authentication token/password (leave empty if none): 

?  Select AI coding agents to configure:
   Space to toggle • a for all • Enter to confirm
 > [ ] [ Select All (21 Agents) ]
   [ ] VS Code (Cline / Roo Code / Copilot) [IDE]
   [ ] Cursor IDE [IDE]
   [ ] Claude Code (CLI) [CLI]
   [ ] Claude Desktop [Desktop]
   [ ] Oh My Pi (omp) [Harness]
   ↓ 14 more below
   --------------------------------------------------
   Target: Configures all 21 agent files
   Skills: .agents/skills/, .cursorrules, CLAUDE.md, etc.

?  Install Codebase RAG skill for agents?
   (•) Yes (adds .cursorrules, CLAUDE.md, .agents/skills/)
   ( ) No (configure MCP connection only)

?  Configuration scope:
   (•) Workspace local (current project only)
   ( ) Global profile (all projects on this machine)
```

Done. Your AI tool is now wired up to your central codebase knowledge graph.

---

## What can you ask your AI?

Once configured, your AI assistant can answer questions like:

- *"What services communicate with payment-service and what ports do they use?"*
- *"Trace the checkout request flow from the web frontend to the database."*
- *"Find all API handlers for user checkout across our repos."*
- *"If I change the customer schema, which upstream services will break?"*

## Installation Options

### Option A: Run directly with npx (Recommended)
No pre-installation needed. Your AI tools invoke it via `npx -y cb-ask run`:
```bash
npx cb-ask
```

### Option B: Global npm Package
```bash
npm install -g cb-ask
cb-ask
```

### Option C: Standalone Go Binary
```bash
# Global install via Go
go install github.com/Abbilville/cb-ask/cmd/cb-ask@latest

# Or build from source
git clone https://github.com/Abbilville/cb-ask.git
cd cb-ask
go build -o bin/cb-ask.exe ./cmd/cb-ask
```

---

## Team & Remote Indexer Setup

In team environments, `cb-indexer` runs centrally on an internal server, and developers connect locally via `cb-ask`.

1. **Central Server (`cb-indexer`)**:
   An administrator runs `cb-indexer` on a server reachable by the team:
   ```bash
   cb-indexer daemon --port 43770 --auth-token "your-internal-secret"
   ```

2. **Developer Setup**:
   Developers run the setup wizard on their laptops:
   ```bash
   npx cb-ask
   ```
   Select **Remote Central Server**, enter the server URL (e.g. `https://cb-indexer.internal.corp:43770`), and input the authentication token.

3. **Zero-`.env` Storage**:
   `cb-ask` does not require `.env` files in your project repositories. Credentials and server URLs are saved to your user profile directory (`%APPDATA%\cb-ask\config.json` on Windows or `~/.config/cb-ask/config.json` on Linux/macOS) or inlined directly into your agent's configuration file.

---

## Supported AI Coding Agents

`cb-ask` supports 21 AI coding environments out of the box:

| Agent | Category | Configuration Target | Skill Target Folder |
| :--- | :---: | :--- | :--- |
| **VS Code** (Cline / Roo / Copilot) | IDE | `.vscode/mcp.json` | `.vscode/` & `.agents/skills/` |
| **Cursor IDE** | IDE | `.cursor/mcp.json` | `.cursorrules` & `.agents/skills/` |
| **Claude Code (CLI)** | CLI | `~/.claude.json` | `CLAUDE.md` (workspace root) |
| **Claude Desktop** | Desktop | `claude_desktop_config.json` | `CLAUDE.md` (workspace root) |
| **Oh My Pi (omp)** | Harness | `omp.json` & `.agents/mcp_config.json` | `.agents/skills/codebase-rag/` |
| **GitHub Copilot CLI** | CLI | `~/.config/github-copilot/mcp.json` | `.github/copilot-instructions.md` |
| **Google Antigravity** | Agent | `.agents/mcp_config.json` | `.agents/skills/codebase-rag/` |
| **Antigravity IDE** | IDE | `.antigravity/mcp.json` | `.agents/skills/codebase-rag/` |
| **Windsurf (Codeium)** | IDE | `~/.codeium/windsurf/mcp_config.json` | `.windsurfrules` (workspace root) |
| **Gemini CLI** | CLI | `~/.gemini/mcp.json` | `.agents/skills/codebase-rag/` |
| **OpenAI Codex CLI** | CLI | `.codex/config.json` | `.agents/skills/` & `CODEX.md` |
| **Visual Studio (IDE)** | IDE | `.vs/mcp.json` | `.vs/` & `.agents/skills/` |
| **Zed Editor** | Editor | `~/.config/zed/settings.json` | `.agents/skills/codebase-rag/` |
| **GitLab Duo CLI** | CLI | `~/.gitlab/duo_mcp.json` | `.gitlab/duo/` & `.agents/skills/` |
| **Qwen Code** | CLI | `.qwen/mcp.json` | `.qwen/` & `.agents/skills/` |
| **Kimi Code CLI** | CLI | `.kimi/mcp.json` | `.kimi/` & `.agents/skills/` |
| **Grok Build** | Agent | `.grok/mcp.json` | `.grok/` & `.agents/skills/` |
| **OpenCode** | Open | `.opencode/mcp.json` | `.opencode/` & `.agents/skills/` |
| **OpenClaw** | Open | `.openclaw/mcp.json` | `.openclaw/` & `.agents/skills/` |
| **KiloCode** | IDE | `.kilocode/mcp.json` | `.kilocode/` & `.agents/skills/` |
| **Devin CLI** | Autonomous | `.devin/mcp.json` | `.devin/` & `.agents/skills/` |

### Manual Agent Configuration Snippet
To configure any custom agent manually:

```json
{
  "mcpServers": {
    "cb-ask": {
      "command": "npx",
      "args": [
        "-y",
        "cb-ask",
        "run",
        "--indexer-url",
        "https://cb-indexer.internal.corp:43770",
        "--auth-token",
        "your-secret-token"
      ]
    }
  }
}
```

---

## Universal Codebase RAG Skill

When enabled during setup, `cb-ask` writes structured instructions to `.agents/skills/codebase-rag/skill.md`, `.cursorrules`, and `CLAUDE.md`.

This contract instructs the model to:
1. **Identify Service Boundaries**: Call `get_architecture_overview()` before guessing which repo owns a feature.
2. **Trace Cross-Service Paths**: Call `trace_cross_service_flow(query="...")` to identify exact network hops, listening ports, and caller/callee function signatures.
3. **Inspect Code Definitions**: Call `query_codebase_symbols()` and `get_symbol_context()` to read the relevant source code slice rather than performing broad grep searches.
4. **Render Sequence Flows**: Format multi-hop responses with exact `file:line` locations and port references.

---

## MCP Tools Specification

`cb-ask` exposes 5 standardized tools over Model Context Protocol:

### 1. `trace_cross_service_flow`
Merges macro topology from `cb-indexer` with AST symbol definitions into an ordered hop sequence.
- **Inputs**:
  - `query` (string, required): Endpoint or feature description (e.g. `"/api/v1/checkout"`).
  - `source_repo` (string, optional): Originating repository.
  - `target_repo` (string, optional): Destination repository.
  - `project` (string, optional): Target project ID.
- **Returns**: Ordered hop sequence with listening ports, caller code symbols, and callee route handler signatures.

### 2. `get_architecture_overview`
Returns the complete microservice architecture topology.
- **Inputs**:
  - `project` (string, optional): Target project ID.
- **Returns**: Repositories, tech stacks, listening ports, and dependency edges.

### 3. `query_codebase_symbols`
Searches indexed AST symbols across microservices.
- **Inputs**:
  - `query` (string, required): Symbol name or regex pattern.
  - `repo_name` (string, optional): Filter by repository.
  - `label` (string, optional): Node type (`Function`, `Method`, `Class`, `Struct`).
  - `limit` (integer, optional): Maximum results (default: 20).
- **Returns**: List of matching symbols with qualified names, file paths, and line numbers.

### 4. `get_symbol_context`
Retrieves a slice of source code around a file and line range from the remote repository.
- **Inputs**:
  - `repo_name` (string, required): Repository name.
  - `file_path` (string, required): Relative file path.
  - `start_line` (integer, optional): Start line number.
  - `end_line` (integer, optional): End line number.
- **Returns**: Line-numbered code snippet with surrounding context.

### 5. `get_workflow_guide`
Returns structured markdown reasoning guidelines for multi-repository code investigation.
- **Inputs**:
  - `guide_name` (string, required): `"cb"` or `"cb-navigator"` (legacy `"oss"` and `"oss-navigator"` also supported).
- **Returns**: Markdown guide text.

---

## CLI Reference

```text
Usage: cb-ask [command] [options]
```

| Command | Usage | Description |
| :--- | :--- | :--- |
| `setup` *(default)* | `cb-ask` or `npx cb-ask` | Launches interactive terminal setup wizard. |
| `run` | `cb-ask run [--indexer-url <url>] [--auth-token <secret>]` | Runs Stdio MCP server (invoked by AI agents). |
| `list` | `cb-ask list [-p project_id] [--indexer-url <url>] [--json]` | Displays architecture topology and dependencies in terminal. |
| `projects` | `cb-ask projects [--indexer-url <url>] [--json]` | Lists registered projects and indexed knowledge graphs. |
| `skill` | `cb-ask skill [--workspace <path>]` | Generates Codebase RAG skill files (`.cursorrules`, `CLAUDE.md`, `.agents/`). |
| `uninstall` | `cb-ask uninstall` or `npx cb-ask uninstall` | Removes cb-ask from all agent configs, cleans skills, and deletes credentials. |
### CLI Options

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--indexer-url <url>` | `http://127.0.0.1:43770` | HTTP endpoint of the running `cb-indexer` server. |
| `--auth-token <secret>` | *(empty)* | Bearer authentication token for secured indexer instances. |
| `--workspace <path>` | `.` | Target project directory for workspace-level configuration. |
| `--global` | `false` | Apply configuration to global user profile instead of current workspace. |
| `--json` | `false` | Output machine-readable JSON representation. |

---

## Troubleshooting

<details>
<summary><strong>Connection refused or timeout connecting to cb-indexer</strong></summary>

1. Verify `cb-indexer` is running on the target host:
   ```bash
   curl -i http://<indexer-host>:<port>/health
   ```
2. If running remotely, ensure firewall rules permit inbound TCP traffic on the configured port.
3. If running behind a reverse proxy, verify `proxy_read_timeout` is set to at least `300s`.
</details>

<details>
<summary><strong>Agent returns 401 Unauthorized</strong></summary>

1. Re-run `npx cb-ask` and enter the secret token at the password prompt.
2. Verify the server's `CB_INDEXER_AUTH_TOKEN` environment variable matches your configured token.
3. Test authentication directly:
   ```bash
   curl -H "Authorization: Bearer <your-token>" http://<indexer-host>:<port>/api/projects
   ```
</details>

<details>
<summary><strong>Stdio transport exits immediately in AI agent</strong></summary>

1. Ensure Node.js (`>=18.0.0`) is available on your system `PATH`.
2. Test the Stdio transport manually in your terminal:
   ```bash
   npx cb-ask run
   ```
   Type `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` and press Enter. It should return a JSON response containing server capabilities.
</details>

---

## License

MIT © [Abbilville](https://github.com/Abbilville)
