# Autonomous Repository Architecture & Topology Workflow

This workflow guides you through multi-repository and cross-service tasks:

1. **Step 1: Check Topology & Index Status**
   - Call `check_project_status(project?)` on `oss-indexer`.
   - Review total repositories, tech stacks, configured ports, and index coverage.

2. **Step 2: Stale / Unindexed Repositories**
   - **IMPORTANT**: If any repository shows `is_indexed: false` or outdated timestamps, **inform the user and ask for confirmation** before calling `trigger_index(project="...")` on `oss-indexer`.
   - Indexing can take CPU time and bandwidth; always obtain user consent.

3. **Step 3: Cross-Service Flow Tracing**
   - For inquiries about request lifecycle, API calls, auth flow, or database relationships across services:
   - Call `trace_cross_service_flow(query="...", source_repo="...", target_repo="...")` on `oss-ask`.
   - Synthesize findings with clear endpoint contracts, parameter breakdown, and a concise Mermaid sequence diagram if explaining a multi-hop flow.

4. **Step 4: Deep AST Code Inspection**
   - For specific method implementation, query `codebase-memory-mcp` scoped by repo (`search_graph`, `trace_path`, `get_code_snippet`).
