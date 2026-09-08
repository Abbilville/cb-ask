#!/usr/bin/env node

/**
 * cb-ask — Enterprise Codebase RAG & Knowledge Graph Agent Setup
 * Zero-dependency modern interactive terminal wizard and Stdio MCP bridge.
 */

import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import http from 'node:http';
import https from 'node:https';
import readline from 'node:readline';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const PKG_ROOT = path.resolve(__dirname, '..');

// ANSI formatting helpers
const c = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  cyan: '\x1b[36m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  magenta: '\x1b[35m',
  gray: '\x1b[90m',
  white: '\x1b[37m',
  bgCyan: '\x1b[46m\x1b[30m',
  clearLine: '\x1b[2K\r',
  cursorUp: (n = 1) => `\x1b[${n}A`,
  cursorDown: (n = 1) => `\x1b[${n}B`,
  hideCursor: '\x1b[?25l',
  showCursor: '\x1b[?25h',
};

// Standard OS directories
function getUserConfigDir() {
  if (process.platform === 'win32') {
    return process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
  }
  return process.env.XDG_CONFIG_HOME || path.join(os.homedir(), '.config');
}

function getGlobalConfigFilePath() {
  return path.join(getUserConfigDir(), 'cb-ask', 'config.json');
}

function loadSavedConfig() {
  const cfgPath = getGlobalConfigFilePath();
  try {
    if (fs.existsSync(cfgPath)) {
      return JSON.parse(fs.readFileSync(cfgPath, 'utf8'));
    }
  } catch {
    // Ignore corrupt config
  }
  return null;
}

function saveConfig(config) {
  const cfgPath = getGlobalConfigFilePath();
  const dir = path.dirname(cfgPath);
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(cfgPath, JSON.stringify(config, null, 2), 'utf8');
  return cfgPath;
}

// All 21 AI Agents in Popularity Order
const AGENT_CATALOG = [
  {
    id: 'vscode',
    name: 'VS Code (Cline / Roo Code / Copilot)',
    badge: 'IDE',
    configFolder: '.vscode/mcp.json',
    skillFolder: '.agents/skills/ & .vscode/',
    defaultSelected: false,
    getTargets: (ws, isGlobal) => {
      const targets = [path.join(ws, '.vscode', 'mcp.json')];
      if (isGlobal) {
        const appData = getUserConfigDir();
        targets.push(
          path.join(appData, 'Code', 'User', 'globalStorage', 'saoudrizwan.claude-dev', 'settings', 'cline_mcp_settings.json'),
          path.join(appData, 'Code', 'User', 'globalStorage', 'rooveterinaryinc.roo-cline', 'settings', 'cline_mcp_settings.json')
        );
      }
      return targets;
    },
  },
  {
    id: 'cursor',
    name: 'Cursor IDE',
    badge: 'IDE',
    configFolder: '.cursor/mcp.json',
    skillFolder: '.cursorrules & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.cursor', 'mcp.json')],
  },
  {
    id: 'claude-code',
    name: 'Claude Code (CLI)',
    badge: 'CLI',
    configFolder: '~/.claude.json',
    skillFolder: 'CLAUDE.md (workspace root)',
    defaultSelected: false,
    getTargets: () => [path.join(os.homedir(), '.claude.json')],
  },
  {
    id: 'claude-desktop',
    name: 'Claude Desktop',
    badge: 'Desktop',
    configFolder: 'Claude/claude_desktop_config.json',
    skillFolder: 'CLAUDE.md (workspace root)',
    defaultSelected: false,
    getTargets: () => {
      if (process.platform === 'win32') {
        return [path.join(getUserConfigDir(), 'Claude', 'claude_desktop_config.json')];
      }
      if (process.platform === 'darwin') {
        return [path.join(os.homedir(), 'Library', 'Application Support', 'Claude', 'claude_desktop_config.json')];
      }
      return [path.join(getUserConfigDir(), 'Claude', 'claude_desktop_config.json')];
    },
  },
  {
    id: 'copilot-cli',
    name: 'GitHub Copilot CLI',
    badge: 'CLI',
    configFolder: '.github/mcp.json',
    skillFolder: '.github/copilot-instructions.md',
    defaultSelected: false,
    getTargets: (ws, isGlobal) => {
      if (isGlobal) {
        return [path.join(getUserConfigDir(), 'github-copilot', 'mcp.json')];
      }
      return [path.join(ws, '.github', 'mcp.json')];
    },
  },
  {
    id: 'omp',
    name: 'Oh My Pi (omp)',
    badge: 'Harness',
    configFolder: 'omp.json & .agents/mcp_config.json',
    skillFolder: '.agents/skills/codebase-rag/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, 'omp.json'), path.join(ws, '.agents', 'mcp_config.json')],
  },
  {
    id: 'antigravity',
    name: 'Google Antigravity',
    badge: 'Agent',
    configFolder: '.agents/mcp_config.json',
    skillFolder: '.agents/skills/codebase-rag/',
    defaultSelected: false,
    getTargets: (ws, isGlobal) => {
      if (isGlobal) {
        return [path.join(os.homedir(), '.gemini', 'config', 'mcp_config.json')];
      }
      return [path.join(ws, '.agents', 'mcp_config.json')];
    },
  },
  {
    id: 'antigravity-ide',
    name: 'Antigravity IDE',
    badge: 'IDE',
    configFolder: '.antigravity/mcp.json',
    skillFolder: '.agents/skills/codebase-rag/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.antigravity', 'mcp.json')],
  },
  {
    id: 'gemini-cli',
    name: 'Gemini CLI',
    badge: 'CLI',
    configFolder: '~/.gemini/mcp.json',
    skillFolder: '.agents/skills/codebase-rag/',
    defaultSelected: false,
    getTargets: () => [path.join(os.homedir(), '.gemini', 'mcp.json')],
  },
  {
    id: 'windsurf',
    name: 'Windsurf (Codeium)',
    badge: 'IDE',
    configFolder: '.codeium/windsurf/mcp_config.json',
    skillFolder: '.windsurfrules (workspace root)',
    defaultSelected: false,
    getTargets: () => [path.join(os.homedir(), '.codeium', 'windsurf', 'mcp_config.json')],
  },
  {
    id: 'codex',
    name: 'OpenAI Codex CLI',
    badge: 'CLI',
    configFolder: '.codex/config.json',
    skillFolder: '.agents/skills/ & CODEX.md',
    defaultSelected: false,
    getTargets: (ws, isGlobal) => {
      if (isGlobal) {
        return [path.join(os.homedir(), '.codex', 'config.json')];
      }
      return [path.join(ws, '.codex', 'config.json')];
    },
  },
  {
    id: 'visual-studio',
    name: 'Visual Studio (IDE)',
    badge: 'IDE',
    configFolder: '.vs/mcp.json',
    skillFolder: '.vs/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.vs', 'mcp.json')],
  },
  {
    id: 'zed',
    name: 'Zed Editor',
    badge: 'Editor',
    configFolder: '~/.config/zed/settings.json',
    skillFolder: '.agents/skills/codebase-rag/',
    defaultSelected: false,
    formatKind: 'zed',
    getTargets: () => [path.join(getUserConfigDir(), 'zed', 'settings.json')],
  },
  {
    id: 'gitlab-duo',
    name: 'GitLab Duo CLI',
    badge: 'CLI',
    configFolder: '~/.gitlab/duo_mcp.json',
    skillFolder: '.gitlab/duo/ & .agents/skills/',
    defaultSelected: false,
    getTargets: () => [path.join(os.homedir(), '.gitlab', 'duo_mcp.json')],
  },
  {
    id: 'qwen-code',
    name: 'Qwen Code',
    badge: 'CLI',
    configFolder: '.qwen/mcp.json',
    skillFolder: '.qwen/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.qwen', 'mcp.json')],
  },
  {
    id: 'kimi-code',
    name: 'Kimi Code CLI',
    badge: 'CLI',
    configFolder: '.kimi/mcp.json',
    skillFolder: '.kimi/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.kimi', 'mcp.json')],
  },
  {
    id: 'grok-build',
    name: 'Grok Build',
    badge: 'Agent',
    configFolder: '.grok/mcp.json',
    skillFolder: '.grok/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.grok', 'mcp.json')],
  },
  {
    id: 'opencode',
    name: 'OpenCode',
    badge: 'Open',
    configFolder: '.opencode/mcp.json',
    skillFolder: '.opencode/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.opencode', 'mcp.json')],
  },
  {
    id: 'openclaw',
    name: 'OpenClaw',
    badge: 'Open',
    configFolder: '.openclaw/mcp.json',
    skillFolder: '.openclaw/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.openclaw', 'mcp.json')],
  },
  {
    id: 'kilocode',
    name: 'KiloCode',
    badge: 'IDE',
    configFolder: '.kilocode/mcp.json',
    skillFolder: '.kilocode/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.kilocode', 'mcp.json')],
  },
  {
    id: 'devin',
    name: 'Devin CLI',
    badge: 'Autonomous',
    configFolder: '.devin/mcp.json',
    skillFolder: '.devin/ & .agents/skills/',
    defaultSelected: false,
    getTargets: (ws) => [path.join(ws, '.devin', 'mcp.json')],
  },
];

// Helper to write JSON config safely
function writeJsonConfig(filePath, serverConfig, formatKind = 'standard') {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });

  let data = {};
  if (fs.existsSync(filePath)) {
    try {
      data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
    } catch {
      data = {};
    }
  }

  if (formatKind === 'zed') {
    if (!data.context_servers || typeof data.context_servers !== 'object') {
      data.context_servers = {};
    }
    data.context_servers['cb-ask'] = {
      command: serverConfig.command,
      args: serverConfig.args,
    };
  } else {
    if (!data.mcpServers || typeof data.mcpServers !== 'object') {
      data.mcpServers = {};
    }
    // Clean legacy server names if present
    delete data.mcpServers['oss-mcp'];
    delete data.mcpServers['oss-query'];

    data.mcpServers['cb-ask'] = serverConfig;
  }

  fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + '\n', 'utf8');
  return filePath;
}

// Universal Codebase RAG Skill content
const SKILL_MARKDOWN = `# Skill: Codebase Knowledge Graph & RAG Navigation

## Intent
Use this skill whenever the user asks questions about:
- Microservice architecture, dependencies, communication routes, or service boundaries.
- Cross-service request flows, API endpoints, or event triggers.
- Code definitions, method signatures, or implementations across repositories.

## Priority & Tool Selection
- **ALWAYS use cb-ask tools FIRST** for cross-service, microservice, or architectural inquiries:
  - \`cb-ask:trace_cross_service_flow(query="...")\`
  - \`cb-ask:get_architecture_overview()\`
  - \`cb-ask:query_codebase_symbols(query="...")\`
  - \`cb-ask:get_symbol_context(repo_name="...", file_path="...")\`
- **DO NOT call codebase-memory-mcp:list_projects or trace_path** for cross-service communication or topology queries. codebase-memory-mcp is single-repo and lacks network ports and inter-service routing. cb-ask already merges AST knowledge graphs with macro topology.
- **NEVER start with broad whole-workspace grep or blind file searches.**

## Step-by-Step Execution
1. **STEP 1 — Identify Boundaries & Topology**:
   - Tool: \`cb-ask:get_architecture_overview()\`
   - Discovers which services exist, tech stacks, and listening ports.
2. **STEP 2 — Trace Cross-Service Flow**:
   - Tool: \`cb-ask:trace_cross_service_flow(query="<feature or endpoint>")\`
   - Returns ordered hops, source/target ports, and caller/callee AST code symbols.
3. **STEP 3 — Query Remote Symbols & Context**:
   - Tool: \`cb-ask:query_codebase_symbols(query="<symbol_pattern>", repo_name="<optional>")\`
   - Tool: \`cb-ask:get_symbol_context(repo_name="...", file_path="...", start_line=N, end_line=M)\`
4. **STEP 4 — Read Only Identified Code**:
   - Use the knowledge graph's exact file paths and line ranges to read only the critical slices of code.
5. **Structure the final response with**:
   - Multi-hop request flow summary (Caller -> Gateway -> Callee -> DB).
   - Exact file locations and function signatures (\`path/to/file:line\`).
`;

function installSkillFiles(workspaceRoot, selectedAgents = []) {
  const createdFiles = [];
  const agentIds = new Set(selectedAgents.map((a) => (typeof a === 'string' ? a : a.id)));
  const installAll = agentIds.size === 0;

  // 1. .agents/skills/codebase-rag/skill.md
  // Standard universal skill folder used by Antigravity, OMP, Codex, Zed, etc.
  const standardAgentIds = [
    'antigravity', 'antigravity-ide', 'omp', 'gemini-cli', 'codex',
    'zed', 'gitlab-duo', 'qwen-code', 'kimi-code', 'grok-build',
    'opencode', 'openclaw', 'kilocode', 'devin', 'vscode', 'cursor'
  ];
  const needsAgentsSkill = installAll || standardAgentIds.some((id) => agentIds.has(id));

  if (needsAgentsSkill) {
    const skillFile = path.join(workspaceRoot, '.agents', 'skills', 'codebase-rag', 'skill.md');
    fs.mkdirSync(path.dirname(skillFile), { recursive: true });
    fs.writeFileSync(skillFile, SKILL_MARKDOWN, 'utf8');
    createdFiles.push(skillFile);
  }

  // 2. Append to .cursorrules (Only when Cursor IDE is selected)
  if (installAll || agentIds.has('cursor')) {
    const cursorRulesPath = path.join(workspaceRoot, '.cursorrules');
    const ruleHeader = '# Skill: Codebase Knowledge Graph & RAG Navigation';
    let cursorRules = '';
    if (fs.existsSync(cursorRulesPath)) {
      cursorRules = fs.readFileSync(cursorRulesPath, 'utf8');
    }
    if (!cursorRules.includes(ruleHeader)) {
      const appended = (cursorRules ? cursorRules.trim() + '\n\n' : '') + SKILL_MARKDOWN;
      fs.writeFileSync(cursorRulesPath, appended, 'utf8');
      createdFiles.push(cursorRulesPath);
    }
  }

  // 3. Append to CLAUDE.md (Only when Claude Code or Claude Desktop is selected)
  if (installAll || agentIds.has('claude-code') || agentIds.has('claude-desktop')) {
    const claudeMdPath = path.join(workspaceRoot, 'CLAUDE.md');
    const ruleHeader = '# Skill: Codebase Knowledge Graph & RAG Navigation';
    let claudeMd = '';
    if (fs.existsSync(claudeMdPath)) {
      claudeMd = fs.readFileSync(claudeMdPath, 'utf8');
    }
    if (!claudeMd.includes(ruleHeader)) {
      const appended = (claudeMd ? claudeMd.trim() + '\n\n' : '') + SKILL_MARKDOWN;
      fs.writeFileSync(claudeMdPath, appended, 'utf8');
      createdFiles.push(claudeMdPath);
    }
  }

  // 4. .windsurfrules (Only when Windsurf is selected)
  if (installAll || agentIds.has('windsurf')) {
    const windsurfRulesPath = path.join(workspaceRoot, '.windsurfrules');
    if (!fs.existsSync(windsurfRulesPath)) {
      fs.writeFileSync(windsurfRulesPath, SKILL_MARKDOWN, 'utf8');
      createdFiles.push(windsurfRulesPath);
    }
  }

  // 5. .github/copilot-instructions.md (Only when Copilot or VS Code is selected)
  if (installAll || agentIds.has('copilot-cli') || agentIds.has('vscode')) {
    const copilotInstrPath = path.join(workspaceRoot, '.github', 'copilot-instructions.md');
    if (!fs.existsSync(copilotInstrPath)) {
      fs.mkdirSync(path.dirname(copilotInstrPath), { recursive: true });
      fs.writeFileSync(copilotInstrPath, SKILL_MARKDOWN, 'utf8');
      createdFiles.push(copilotInstrPath);
    }
  }

  return createdFiles;
}
function cleanEmptyParents(dir, maxDepth = 3) {
  let current = path.resolve(dir);
  const cwd = path.resolve(process.cwd());
  const home = path.resolve(os.homedir());

  for (let i = 0; i < maxDepth; i++) {
    if (!fs.existsSync(current)) break;
    // Safety: NEVER delete workspace root, home directory, or filesystem root
    if (current === cwd || current === home || path.dirname(current) === current) {
      break;
    }
    try {
      const entries = fs.readdirSync(current);
      if (entries.length === 0) {
        fs.rmdirSync(current);
        current = path.dirname(current);
      } else {
        break;
      }
    } catch {
      break;
    }
  }
}

function uninstallServerConfig(filePath, formatKind = 'standard') {
  if (!fs.existsSync(filePath)) return false;
  try {
    const raw = fs.readFileSync(filePath, 'utf8');
    const data = JSON.parse(raw);
    let modified = false;

    if (formatKind === 'zed') {
      if (data.context_servers && data.context_servers['cb-ask']) {
        delete data.context_servers['cb-ask'];
        modified = true;
      }
    } else {
      if (data.mcpServers) {
        for (const name of ['cb-ask', 'oss-ask', 'cb-indexer', 'oss-indexer', 'oss-mcp', 'oss-query']) {
          if (data.mcpServers[name]) {
            delete data.mcpServers[name];
            modified = true;
          }
        }
      }
    }

    let isEmpty = false;
    if (formatKind === 'zed') {
      const remaining = data.context_servers ? Object.keys(data.context_servers).length : 0;
      const topKeys = Object.keys(data);
      isEmpty = topKeys.length === 0 || (topKeys.length === 1 && topKeys[0] === 'context_servers' && remaining === 0);
    } else {
      const remaining = data.mcpServers ? Object.keys(data.mcpServers).length : 0;
      const topKeys = Object.keys(data);
      isEmpty = topKeys.length === 0 || (topKeys.length === 1 && topKeys[0] === 'mcpServers' && remaining === 0);
    }

    if (isEmpty) {
      fs.unlinkSync(filePath);
      cleanEmptyParents(path.dirname(filePath));
      return true;
    }

    if (modified) {
      fs.writeFileSync(filePath, JSON.stringify(data, null, 2) + '\n', 'utf8');
      return true;
    }
  } catch {
    return false;
  }
  return false;
}
function uninstallSkillFiles(workspaceRoot) {
  const removedFiles = [];
  const ruleHeader = '# Skill: Codebase Knowledge Graph & RAG Navigation';

  // 1. .agents/skills/codebase-rag/skill.md
  const skillFile = path.join(workspaceRoot, '.agents', 'skills', 'codebase-rag', 'skill.md');
  if (fs.existsSync(skillFile)) {
    try {
      fs.unlinkSync(skillFile);
      removedFiles.push(skillFile);
      cleanEmptyParents(path.dirname(skillFile));
    } catch {}
  }
  // 2. .cursorrules
  const cursorRulesPath = path.join(workspaceRoot, '.cursorrules');
  if (fs.existsSync(cursorRulesPath)) {
    try {
      let text = fs.readFileSync(cursorRulesPath, 'utf8').replace(/\r\n/g, '\n');
      const normalizedSkill = SKILL_MARKDOWN.replace(/\r\n/g, '\n');
      if (text.includes(ruleHeader)) {
        text = text.replace(normalizedSkill, '').trim();
        if (text.length === 0 || text === ruleHeader) {
          fs.unlinkSync(cursorRulesPath);
          removedFiles.push(cursorRulesPath);
          cleanEmptyParents(path.dirname(cursorRulesPath));
        } else {
          fs.writeFileSync(cursorRulesPath, text + '\n', 'utf8');
          removedFiles.push(cursorRulesPath + ' (cleaned)');
        }
      }
    } catch {}
  }

  // 3. CLAUDE.md
  const claudeMdPath = path.join(workspaceRoot, 'CLAUDE.md');
  if (fs.existsSync(claudeMdPath)) {
    try {
      let text = fs.readFileSync(claudeMdPath, 'utf8').replace(/\r\n/g, '\n');
      const normalizedSkill = SKILL_MARKDOWN.replace(/\r\n/g, '\n');
      if (text.includes(ruleHeader)) {
        text = text.replace(normalizedSkill, '').trim();
        if (text.length === 0 || text === ruleHeader) {
          fs.unlinkSync(claudeMdPath);
          removedFiles.push(claudeMdPath);
          cleanEmptyParents(path.dirname(claudeMdPath));
        } else {
          fs.writeFileSync(claudeMdPath, text + '\n', 'utf8');
          removedFiles.push(claudeMdPath + ' (cleaned)');
        }
      }
    } catch {}
  }

  // 4. .windsurfrules
  const windsurfRulesPath = path.join(workspaceRoot, '.windsurfrules');
  if (fs.existsSync(windsurfRulesPath)) {
    try {
      const text = fs.readFileSync(windsurfRulesPath, 'utf8').replace(/\r\n/g, '\n').trim();
      const normalizedSkill = SKILL_MARKDOWN.replace(/\r\n/g, '\n').trim();
      if (text === normalizedSkill || text.includes(ruleHeader)) {
        fs.unlinkSync(windsurfRulesPath);
        removedFiles.push(windsurfRulesPath);
        cleanEmptyParents(path.dirname(windsurfRulesPath));
      }
    } catch {}
  }

  // 5. .github/copilot-instructions.md
  const copilotInstrPath = path.join(workspaceRoot, '.github', 'copilot-instructions.md');
  if (fs.existsSync(copilotInstrPath)) {
    try {
      const text = fs.readFileSync(copilotInstrPath, 'utf8').replace(/\r\n/g, '\n').trim();
      const normalizedSkill = SKILL_MARKDOWN.replace(/\r\n/g, '\n').trim();
      if (text === normalizedSkill || text.includes(ruleHeader)) {
        fs.unlinkSync(copilotInstrPath);
        removedFiles.push(copilotInstrPath);
        cleanEmptyParents(path.dirname(copilotInstrPath));
      }
    } catch {}
  }

  return removedFiles;
}

// Test connectivity to cb-indexer
function testIndexerConnection(indexerUrl, authToken) {
  return new Promise((resolve) => {
    try {
      const parsed = new URL(indexerUrl);
      const isHttps = parsed.protocol === 'https:';
      const client = isHttps ? https : http;

      const basePath = parsed.pathname.replace(/\/mcp\/?$/, '').replace(/\/$/, '');
      const testPath = (basePath || '') + '/health';
      const headers = {
        'User-Agent': 'cb-ask-setup/0.1.0',
        'Accept': 'application/json, text/event-stream',
      };
      if (authToken) {
        headers['Authorization'] = 'Bearer ' + authToken;
        headers['X-API-Key'] = authToken;
      }

      const req = client.request(
        {
          hostname: parsed.hostname,
          port: parsed.port || (isHttps ? 443 : 80),
          path: testPath,
          method: 'GET',
          headers,
          timeout: 4000,
        },
        (res) => {
          resolve({ ok: res.statusCode >= 200 && res.statusCode < 400, statusCode: res.statusCode });
        }
      );

      req.on('error', (err) => {
        resolve({ ok: false, error: err.message });
      });
      req.on('timeout', () => {
        req.destroy();
        resolve({ ok: false, error: 'Connection timed out' });
      });
      req.end();
    } catch (err) {
      resolve({ ok: false, error: err.message });
    }
  });
}

function fetchProjects(indexerUrl, authToken) {
  return new Promise((resolve) => {
    try {
      const parsed = new URL(indexerUrl);
      const isHttps = parsed.protocol === 'https:';
      const client = isHttps ? https : http;
      const basePath = parsed.pathname.replace(/\/mcp\/?$/, '').replace(/\/$/, '');
      const apiPath = (basePath || '') + '/api/projects';
      const headers = { 'User-Agent': 'cb-ask-setup/0.1.0' };

      const req = client.request(
        {
          hostname: parsed.hostname,
          port: parsed.port || (isHttps ? 443 : 80),
          path: apiPath,
          method: 'GET',
          headers,
          timeout: 4000,
        },
        (res) => {
          let data = '';
          res.on('data', (c) => (data += c));
          res.on('end', () => {
            try {
              const parsedData = JSON.parse(data);
              resolve(parsedData.registered_projects || []);
            } catch {
              resolve([]);
            }
          });
        }
      );
      req.on('error', () => resolve([]));
      req.on('timeout', () => { req.destroy(); resolve([]); });
      req.end();
    } catch {
      resolve([]);
    }
  });
}

// Terminal interactive prompt engine
class InteractivePrompter {
  constructor() {
    readline.emitKeypressEvents(process.stdin);
    this.wasRaw = process.stdin.isRaw;
    if (process.stdin.setRawMode) {
      process.stdin.setRawMode(true);
    }
    process.stdin.resume();
  }

  close() {
    if (process.stdin.setRawMode) {
      process.stdin.setRawMode(this.wasRaw || false);
    }
    process.stdin.pause();
  }

  askText(questionText, defaultValue = '') {
    return new Promise((resolve) => {
      const prompt = defaultValue
        ? `${c.cyan}?${c.reset}  ${c.bold}${questionText}${c.reset} ${c.dim}(default: ${defaultValue})${c.reset}: `
        : `${c.cyan}?${c.reset}  ${c.bold}${questionText}${c.reset}: `;
      process.stdout.write(prompt);
      let text = '';

      const onKeypress = (str, key) => {
        if (!key) return;
        if (key.ctrl && key.name === 'c') {
          process.stdout.write('\n');
          process.exit(130);
        }
        if (key.name === 'return' || key.name === 'enter') {
          process.stdin.removeListener('keypress', onKeypress);
          const val = text.trim() || defaultValue;
          process.stdout.write(`\r${c.clearLine}${c.green}✓${c.reset}  ${c.bold}${questionText}:${c.reset} ${c.cyan}${val}${c.reset}\n\n`);
          resolve(val);
        } else if (key.name === 'backspace') {
          if (text.length > 0) {
            text = text.slice(0, -1);
            process.stdout.write('\b \b');
          }
        } else if (str && !key.ctrl && !key.meta) {
          text += str;
          process.stdout.write(str);
        }
      };

      process.stdin.on('keypress', onKeypress);
    });
  }

  askPassword(questionText) {
    return new Promise((resolve) => {
      process.stdout.write(`${c.clearLine}${c.cyan}?${c.reset}  ${c.bold}${questionText}${c.reset} ${c.dim}(leave empty if none)${c.reset}: `);
      let password = '';

      const onKeypress = (str, key) => {
        if (!key) return;
        if (key.ctrl && key.name === 'c') {
          process.stdout.write('\n');
          process.exit(130);
        }
        if (key.name === 'return' || key.name === 'enter') {
          process.stdin.removeListener('keypress', onKeypress);
          const display = password ? '••••••••••••' : '(none)';
          process.stdout.write(`\r${c.clearLine}${c.green}✓${c.reset}  ${c.bold}${questionText}:${c.reset} ${c.dim}${display}${c.reset}\n\n`);
          resolve(password.trim());
        } else if (key.name === 'backspace') {
          if (password.length > 0) {
            password = password.slice(0, -1);
            process.stdout.write('\b \b');
          }
        } else if (str && str.length === 1 && !key.ctrl && !key.meta) {
          password += str;
          process.stdout.write('*');
        }
      };

      process.stdin.on('keypress', onKeypress);
    });
  }

  askChoice(questionText, options, defaultIndex = 0) {
    return new Promise((resolve) => {
      let selected = defaultIndex;
      process.stdout.write(c.hideCursor);

      let lastLinesCount = 0;

      const render = () => {
        let frame = '';
        if (lastLinesCount > 0) {
          frame += c.cursorUp(lastLinesCount);
        }
        let lines = 0;
        frame += `${c.clearLine}${c.cyan}?${c.reset}  ${c.bold}${questionText}${c.reset}\n`;
        lines++;

        options.forEach((opt, idx) => {
          const prefix = idx === selected ? `${c.cyan}(•)${c.reset}` : `${c.gray}( )${c.reset}`;
          const label = idx === selected ? `${c.bold}${c.cyan}${opt.label}${c.reset}` : opt.label;
          const hint = opt.hint ? ` ${c.dim}(${opt.hint})${c.reset}` : '';
          frame += `${c.clearLine}   ${prefix} ${label}${hint}\n`;
          lines++;
        });

        lastLinesCount = lines;
        process.stdout.write(frame);
      };

      render();

      const onKeypress = (str, key) => {
        if (!key) return;
        if (key.ctrl && key.name === 'c') {
          process.stdout.write(c.showCursor + '\n');
          process.exit(130);
        }
        if (key.name === 'up' || key.name === 'k') {
          selected = (selected - 1 + options.length) % options.length;
          render();
        } else if (key.name === 'down' || key.name === 'j') {
          selected = (selected + 1) % options.length;
          render();
        } else if (key.name === 'return' || key.name === 'enter') {
          process.stdin.removeListener('keypress', onKeypress);
          process.stdout.write(c.showCursor);

          let wipe = '';
          if (lastLinesCount > 0) {
            wipe += c.cursorUp(lastLinesCount);
            for (let i = 0; i < lastLinesCount; i++) {
              wipe += `${c.clearLine}\n`;
            }
            wipe += c.cursorUp(lastLinesCount);
          }
          wipe += `${c.clearLine}${c.green}✓${c.reset}  ${c.bold}${questionText}:${c.reset} ${c.cyan}${options[selected].label}${c.reset}\n\n`;
          process.stdout.write(wipe);

          resolve(options[selected].value);
        }
      };

      process.stdin.on('keypress', onKeypress);
    });
  }

  askMultiSelect(questionText, items) {
    return new Promise((resolve) => {
      const selections = new Set();
      let cursor = 0;
      const totalDisplay = items.length + 1; // Row 0 is [Select All], 1..N are items
      const PAGE_SIZE = 8;

      process.stdout.write(c.hideCursor);

      let lastLinesCount = 0;

      const render = () => {
        let frame = '';
        if (lastLinesCount > 0) {
          frame += c.cursorUp(lastLinesCount);
        }
        let lines = 0;

        // Line 1: Question
        frame += `${c.clearLine}${c.cyan}?${c.reset}  ${c.bold}${questionText}${c.reset}\n`;
        lines++;

        // Line 2: Instructions
        frame += `${c.clearLine}   ${c.dim}Space to toggle • a for all • Enter to confirm${c.reset}\n`;
        lines++;

        // Calculate window bounds around cursor
        let start = 0;
        if (cursor >= PAGE_SIZE) {
          start = cursor - PAGE_SIZE + 1;
        }
        const end = Math.min(start + PAGE_SIZE, totalDisplay);

        for (let rowIdx = start; rowIdx < end; rowIdx++) {
          const isCursor = cursor === rowIdx;
          const pointer = isCursor ? `${c.cyan}>${c.reset}` : ' ';

          if (rowIdx === 0) {
            const allSelected = selections.size === items.length;
            const box = allSelected ? `${c.green}[x]${c.reset}` : `${c.gray}[ ]${c.reset}`;
            const label = isCursor
              ? `${c.bold}${c.cyan}[ Select All (${items.length} Agents) ]${c.reset}`
              : `[ Select All (${items.length} Agents) ]`;
            frame += `${c.clearLine} ${pointer} ${box} ${label}\n`;
            lines++;
          } else {
            const itemIdx = rowIdx - 1;
            const item = items[itemIdx];
            const isSelected = selections.has(itemIdx);
            const box = isSelected ? `${c.green}[x]${c.reset}` : `${c.gray}[ ]${c.reset}`;
            const badge = item.badge ? ` ${c.gray}[${item.badge}]${c.reset}` : '';
            const name = isCursor ? `${c.bold}${c.cyan}${item.name}${c.reset}` : item.name;
            frame += `${c.clearLine} ${pointer} ${box} ${name}${badge}\n`;
            lines++;
          }
        }

        // Scroll indicator footer
        const hiddenAbove = start;
        const hiddenBelow = totalDisplay - end;
        let scrollHint = '';
        if (hiddenAbove > 0 && hiddenBelow > 0) {
          scrollHint = `↑ ${hiddenAbove} more above | ↓ ${hiddenBelow} more below`;
        } else if (hiddenAbove > 0) {
          scrollHint = `↑ ${hiddenAbove} more above`;
        } else if (hiddenBelow > 0) {
          scrollHint = `↓ ${hiddenBelow} more below`;
        }
        if (scrollHint) {
          frame += `${c.clearLine}   ${c.dim}${scrollHint}${c.reset}\n`;
          lines++;
        }

        // Active item path preview panel
        let cfgHint = 'Configures all 21 agent files';
        let sklHint = '.agents/skills/, .cursorrules, CLAUDE.md, etc.';
        if (cursor > 0) {
          const active = items[cursor - 1];
          cfgHint = active.configFolder || 'Auto-detected path';
          sklHint = active.skillFolder || '.agents/skills/codebase-rag/';
        }
        frame += `${c.clearLine}   ${c.dim}--------------------------------------------------${c.reset}\n`;
        frame += `${c.clearLine}   ${c.dim}Target:${c.reset} ${c.cyan}${cfgHint}${c.reset}\n`;
        frame += `${c.clearLine}   ${c.dim}Skills:${c.reset} ${c.dim}${sklHint}${c.reset}\n`;
        lines += 3;

        lastLinesCount = lines;
        process.stdout.write(frame);
      };

      render();

      const onKeypress = (str, key) => {
        if (!key) return;
        if (key.ctrl && key.name === 'c') {
          process.stdout.write(c.showCursor + '\n');
          process.exit(130);
        }

        if (key.name === 'up' || key.name === 'k') {
          cursor = (cursor - 1 + totalDisplay) % totalDisplay;
          render();
        } else if (key.name === 'down' || key.name === 'j') {
          cursor = (cursor + 1) % totalDisplay;
          render();
        } else if (key.name === 'space') {
          if (cursor === 0) {
            if (selections.size === items.length) {
              selections.clear();
            } else {
              items.forEach((_, i) => selections.add(i));
            }
          } else {
            const targetIdx = cursor - 1;
            if (selections.has(targetIdx)) {
              selections.delete(targetIdx);
            } else {
              selections.add(targetIdx);
            }
          }
          render();
        } else if (str === 'a' || str === 'A' || key.name === 'a') {
          if (selections.size === items.length) {
            selections.clear();
          } else {
            items.forEach((_, i) => selections.add(i));
          }
          render();
        } else if (key.name === 'return' || key.name === 'enter') {
          process.stdin.removeListener('keypress', onKeypress);
          process.stdout.write(c.showCursor);

          const chosen = Array.from(selections).sort((a, b) => a - b).map((idx) => items[idx]);

          let wipe = '';
          if (lastLinesCount > 0) {
            wipe += c.cursorUp(lastLinesCount);
            for (let i = 0; i < lastLinesCount; i++) {
              wipe += `${c.clearLine}\n`;
            }
            wipe += c.cursorUp(lastLinesCount);
          }
          wipe += `${c.clearLine}${c.green}✓${c.reset}  ${c.bold}${questionText}:${c.reset} ${c.cyan}(${chosen.length} agents selected)${c.reset}\n\n`;
          process.stdout.write(wipe);

          resolve(chosen);
        }
      };

      process.stdin.on('keypress', onKeypress);
    });
  }
}

// CLI entry point
async function runInteractiveWizard() {
  console.clear();
  console.log(`${c.bold}cb-ask${c.reset} ${c.dim}— Codebase RAG & Knowledge Graph Agent Setup${c.reset}\n`);

  const prompter = new InteractivePrompter();

  // 1. Where is cb-indexer running?
  const locationChoice = await prompter.askChoice(
    'cb-indexer location:',
    [
      { label: 'Localhost', value: 'local', hint: 'http://127.0.0.1:43770' },
      { label: 'Remote server', value: 'remote', hint: 'custom URL' },
    ],
    0
  );

  let indexerUrl = 'http://127.0.0.1:43770';
  if (locationChoice === 'remote') {
    indexerUrl = await prompter.askText(
      'Server URL',
      'https://cb-indexer.internal.corp:43770'
    );
  }

  // Ensure /mcp format
  indexerUrl = indexerUrl.replace(/\/+$/, '');
  if (!indexerUrl.endsWith('/mcp')) {
    indexerUrl += '/mcp';
  }

  // 2. Authentication Password / Secret Token
  const authToken = await prompter.askPassword('Authentication token/password');

  // Check connection and fetch available projects
  process.stdout.write(`${c.dim}Checking connection to ${indexerUrl}...${c.reset}\r`);
  const conn = await testIndexerConnection(indexerUrl, authToken);
  let availableProjects = [];
  if (conn.ok) {
    availableProjects = await fetchProjects(indexerUrl, authToken);
    process.stdout.write(`${c.clearLine}${c.green}✓ Connected to cb-indexer.${c.reset}\n\n`);
  } else {
    process.stdout.write(`${c.clearLine}${c.yellow}! Warning: Could not reach cb-indexer (${conn.error || 'HTTP ' + conn.statusCode}).${c.reset}\n\n`);
  }

  // 3. Project Selection (Only prompted if 2+ projects exist on server!)
  let chosenProject = '';
  if (availableProjects.length === 1) {
    chosenProject = availableProjects[0].project_id;
  } else if (availableProjects.length > 1) {
    const cwd = process.cwd().toLowerCase();
    let defaultProjIdx = 0;
    const projectOptions = availableProjects.map((p, idx) => {
      const isMatch = (p.registry_path && cwd.includes(path.dirname(p.registry_path).toLowerCase())) ||
                      (p.description && cwd.includes(p.project_id.toLowerCase()));
      if (isMatch) defaultProjIdx = idx;
      return {
        label: p.project_id,
        value: p.project_id,
        hint: p.name || p.description || '',
      };
    });

    chosenProject = await prompter.askChoice('Select active project ecosystem:', projectOptions, defaultProjIdx);
  }

  // 4. Select AI Agents
  const selectedAgents = await prompter.askMultiSelect(
    'Select AI coding agents to configure:',
    AGENT_CATALOG
  );

  // 5. Install Skill
  const installSkill = await prompter.askChoice(
    'Install Codebase RAG skill for agents?',
    [
      { label: 'Yes', value: true, hint: 'adds .cursorrules, CLAUDE.md, .agents/skills/' },
      { label: 'No', value: false, hint: 'configure MCP connection only' },
    ],
    0
  );

  // 6. Scope
  const isGlobal = await prompter.askChoice(
    'Configuration scope:',
    [
      { label: 'Workspace local', value: false, hint: 'current project only' },
      { label: 'Global profile', value: true, hint: 'all projects on this machine' },
    ],
    0
  );

  prompter.close();

  // Persist zero-.env user configuration
  const globalConfig = {
    indexerUrl,
    authToken: authToken || undefined,
    project: chosenProject || undefined,
    installedAt: new Date().toISOString(),
    agents: selectedAgents.map((a) => a.id),
  };
  const savedConfigPath = saveConfig(globalConfig);
  console.log(`${c.green}✓ Credentials saved to: ${c.dim}${savedConfigPath}${c.reset}`);

  // Build the MCP command for agents
  const serverArgs = ['-y', 'cb-ask', 'run', '--indexer-url', indexerUrl];
  if (chosenProject) {
    serverArgs.push('--project', chosenProject);
  }
  if (authToken) {
    serverArgs.push('--auth-token', authToken);
  }

  const serverConfig = {
    command: 'npx',
    args: serverArgs,
  };

  const workspaceRoot = process.cwd();
  console.log(`\n${c.bold}Writing agent configurations:${c.reset}`);

  let writtenCount = 0;
  for (const agent of selectedAgents) {
    const targets = agent.getTargets(workspaceRoot, isGlobal);
    for (const target of targets) {
      try {
        writeJsonConfig(target, serverConfig, agent.formatKind || 'standard');
        console.log(`  ${c.green}✓${c.reset} ${agent.name} ${c.gray}-> ${path.relative(workspaceRoot, target) || target}${c.reset}`);
        writtenCount++;
      } catch (err) {
        console.log(`  ${c.red}!${c.reset} ${agent.name}: Failed to write ${target} (${err.message})`);
      }
    }
  }

  if (installSkill) {
    console.log(`\n${c.bold}Installing Codebase RAG skills:${c.reset}`);
    const createdSkills = installSkillFiles(workspaceRoot, selectedAgents);
    for (const file of createdSkills) {
      console.log(`  ${c.green}✓${c.reset} ${path.relative(workspaceRoot, file) || file}`);
    }
  }

  console.log(`\n${c.green}✓ Setup complete.${c.reset} Configured ${writtenCount} agent file(s).\n
${c.bold}Ask your AI assistant:${c.reset}
  "Explain our microservice architecture and listening ports."
  "Trace how payment-service communicates with account-service."
  "Find all handlers for user checkout across the repositories."\n`);
}


async function runUninstallCommand() {
  console.log(`${c.bold}cb-ask — Uninstall${c.reset}\n──────────────────\n`);

  const workspaceRoot = process.cwd();
  const removedConfigs = [];

  for (const agent of AGENT_CATALOG) {
    const wsTargets = agent.getTargets(workspaceRoot, false);
    const globalTargets = agent.getTargets(workspaceRoot, true);
    const allTargets = Array.from(new Set([...wsTargets, ...globalTargets]));

    for (const target of allTargets) {
      if (uninstallServerConfig(target, agent.formatKind || 'standard')) {
        removedConfigs.push({ agent: agent.name, path: target });
      }
    }
  }

  if (removedConfigs.length > 0) {
    console.log(`${c.bold}Removed agent configurations:${c.reset}`);
    for (const item of removedConfigs) {
      const displayPath = path.relative(workspaceRoot, item.path) || item.path;
      console.log(`  ${c.green}✓${c.reset} ${item.agent} ${c.gray}-> ${displayPath}${c.reset}`);
    }
    console.log();
  } else {
    console.log(`${c.dim}No agent configurations found to remove.${c.reset}\n`);
  }

  const removedSkills = uninstallSkillFiles(workspaceRoot);
  if (removedSkills.length > 0) {
    console.log(`${c.bold}Removed skills:${c.reset}`);
    for (const file of removedSkills) {
      const displayPath = path.relative(workspaceRoot, file) || file;
      console.log(`  ${c.green}✓${c.reset} ${displayPath}`);
    }
    console.log();
  }

  // Remove persistent user credentials
  const cfgPath = getGlobalConfigFilePath();
  if (fs.existsSync(cfgPath)) {
    try {
      fs.unlinkSync(cfgPath);
      console.log(`${c.green}✓${c.reset} Credentials removed from: ${c.dim}${cfgPath}${c.reset}\n`);
    } catch {}
  }

  console.log(`${c.green}✓ Uninstalled cb-ask. All configurations removed.${c.reset}\n`);
}
// Fallback pure-Node JSON-RPC MCP server over stdio
async function runStdioMcpProxy(indexerUrl, authToken, defaultProject = '') {
  // Try spawning compiled Go binary for the current OS/architecture
  let candidateBinaries = [];
  if (process.platform === 'win32') {
    candidateBinaries.push('cb-ask.exe', 'cb-ask');
  } else if (process.platform === 'darwin') {
    candidateBinaries.push('cb-ask-darwin-arm64', 'cb-ask-darwin-amd64', 'cb-ask');
  } else if (process.platform === 'linux') {
    candidateBinaries.push('cb-ask-linux-amd64', 'cb-ask');
  }

  for (const binName of candidateBinaries) {
    const binPath = path.join(PKG_ROOT, 'bin', binName);
    if (fs.existsSync(binPath)) {
      try {
        if (process.platform !== 'win32') {
          fs.chmodSync(binPath, 0o755);
        }
        const args = ['run', '--indexer-url', indexerUrl];
        if (defaultProject) {
          args.push('--project', defaultProject);
        }
        if (authToken) {
          args.push('--auth-token', authToken);
        }
        const proc = spawn(binPath, args, { stdio: 'inherit' });
        proc.on('exit', (code) => process.exit(code || 0));
        return;
      } catch {
        // Fallback to JS proxy if binary cannot be spawned
      }
    }
  }

  // Native lightweight Node.js Stdio MCP Proxy
  const rl = readline.createInterface({ input: process.stdin, output: process.stdout, terminal: false });

  function sendResult(id, result) {
    const msg = JSON.stringify({ jsonrpc: '2.0', id, result });
    process.stdout.write(msg + '\n');
  }

  function sendError(id, code, message) {
    const msg = JSON.stringify({ jsonrpc: '2.0', id, error: { code, message } });
    process.stdout.write(msg + '\n');
  }

  async function forwardToIndexer(toolName, args) {
    const parsed = new URL(indexerUrl);
    const client = parsed.protocol === 'https:' ? https : http;
    const body = JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/call',
      params: { name: toolName, arguments: args },
    });

    return new Promise((resolve, reject) => {
      const headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json, text/event-stream',
        'Content-Length': Buffer.byteLength(body),
      };
      if (authToken) {
        headers['Authorization'] = 'Bearer ' + authToken;
      }
      const req = client.request(
        {
          hostname: parsed.hostname,
          port: parsed.port || (parsed.protocol === 'https:' ? 443 : 80),
          path: parsed.pathname,
          method: 'POST',
          headers,
        },
        (res) => {
          let data = '';
          res.on('data', (chunk) => (data += chunk));
          res.on('end', () => {
            try {
              resolve(JSON.parse(data));
            } catch (err) {
              reject(err);
            }
          });
        }
      );
      req.on('error', reject);
      req.write(body);
      req.end();
    });
  }

  rl.on('line', async (line) => {
    if (!line.trim()) return;
    try {
      const req = JSON.parse(line);
      if (req.method === 'initialize') {
        sendResult(req.id, {
          protocolVersion: '2024-11-05',
          capabilities: { tools: {} },
          serverInfo: { name: 'cb-ask', version: '0.1.0' },
        });
        return;
      }
      if (req.method === 'tools/list') {
        sendResult(req.id, {
          tools: [
            {
              name: 'trace_cross_service_flow',
              description: 'Trace end-to-end request flow across microservices by merging topology with AST code symbols.',
              inputSchema: {
                type: 'object',
                properties: {
                  query: { type: 'string', description: 'Feature or endpoint description' },
                  source_repo: { type: 'string', description: 'Originating repository' },
                  target_repo: { type: 'string', description: 'Destination repository' },
                },
                required: ['query'],
              },
            },
            {
              name: 'get_architecture_overview',
              description: 'Fetch the complete microservice architecture topology, listening ports, tech stacks, and dependency edges.',
              inputSchema: { type: 'object', properties: { project: { type: 'string' } } },
            },
            {
              name: 'query_codebase_symbols',
              description: 'Search indexed AST symbols across microservices (functions, methods, classes, structs) via the knowledge graph.',
              inputSchema: {
                type: 'object',
                properties: {
                  query: { type: 'string' },
                  repo_name: { type: 'string' },
                  label: { type: 'string' },
                  limit: { type: 'number' },
                },
                required: ['query'],
              },
            },
            {
              name: 'get_symbol_context',
              description: 'Retrieve source code context snippet around a file and line range from the repository.',
              inputSchema: {
                type: 'object',
                properties: {
                  repo_name: { type: 'string' },
                  file_path: { type: 'string' },
                  start_line: { type: 'number' },
                  end_line: { type: 'number' },
                },
                required: ['repo_name', 'file_path'],
              },
            },
            {
              name: 'get_workflow_guide',
              description: 'Retrieve architectural navigation and multi-repository workflow instructions.',
              inputSchema: {
                type: 'object',
                properties: { guide_name: { type: 'string' } },
                required: ['guide_name'],
              },
            },
          ],
        });
        return;
      }
      if (req.method === 'tools/call') {
        const { name, arguments: args } = req.params;
        const callArgs = args || {};
        if (defaultProject && !callArgs.project) {
          callArgs.project = defaultProject;
        }

        if (name === 'get_workflow_guide') {
          sendResult(req.id, {
            content: [{ type: 'text', text: SKILL_MARKDOWN }],
          });
          return;
        }

        try {
          const resp = await forwardToIndexer(name, callArgs);
          if (resp.result) {
            sendResult(req.id, resp.result);
          } else if (resp.error) {
            sendError(req.id, resp.error.code || -32000, resp.error.message);
          } else {
            sendResult(req.id, { content: [{ type: 'text', text: JSON.stringify(resp) }] });
          }
        } catch (err) {
          sendResult(req.id, {
            content: [{ type: 'text', text: JSON.stringify({ error: err.message }) }],
            isError: true,
          });
        }
        return;
      }

      sendResult(req.id, {});
    } catch {
      // Ignore unparseable line
    }
  });
}

// Dispatch based on CLI arguments
async function main() {
  const args = process.argv.slice(2);
  const command = args[0] || 'init';

  if (command === 'run') {
    // Parse flags or load from user config
    let indexerUrl = '';
    let authToken = '';
    let project = '';

    for (let i = 1; i < args.length; i++) {
      if (args[i] === '--indexer-url' && args[i + 1]) {
        indexerUrl = args[i + 1];
        i++;
      } else if (args[i] === '--auth-token' && args[i + 1]) {
        authToken = args[i + 1];
        i++;
      } else if (args[i] === '--project' && args[i + 1]) {
        project = args[i + 1];
        i++;
      }
    }

    if (!indexerUrl || !project) {
      const saved = loadSavedConfig();
      if (saved) {
        indexerUrl = indexerUrl || saved.indexerUrl;
        authToken = authToken || saved.authToken;
        project = project || saved.project;
      }
    }

    indexerUrl = indexerUrl || 'http://127.0.0.1:43770/mcp';
    await runStdioMcpProxy(indexerUrl, authToken, project);
    return;
  }

  if (command === 'uninstall' || command === 'uninstall-agent' || command === 'remove') {
    await runUninstallCommand();
    return;
  }

  if (command === 'skill' || command === 'skills') {
    const created = installSkillFiles(process.cwd());
    console.log(`${c.green}✓ Installed Codebase RAG skills into:${c.reset}`);
    for (const f of created) {
      console.log(`  * ${f}`);
    }
    return;
  }

  // Default / setup / init: Interactive terminal wizard
  await runInteractiveWizard();
}

main().catch((err) => {
  console.error(`\n${c.red}[ERROR]${c.reset} ${err.message}`);
  process.exit(1);
});
