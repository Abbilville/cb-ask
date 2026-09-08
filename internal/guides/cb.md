# Autonomous Repository Architecture & Topology Workflow

This workflow guides you through multi-repository and cross-service tasks:

1. **Step 1: Check Topology & Index Status**
   - Call `check_project_status(project?)` on `cb-indexer`.
   - Review total repositories, tech stacks, configured ports, and index coverage.

2. **Step 2: Stale / Unindexed Repositories**
   - **IMPORTANT**: If any repository shows `is_indexed: false` or outdated timestamps, **inform the user and ask for confirmation** before calling `trigger_index(project="...")` on `cb-indexer`.
   - Indexing can take CPU time and bandwidth; always obtain user consent.

3. **Step 3: Cross-Service Flow Tracing**
   - For inquiries about request lifecycle, API calls, auth flow, or database relationships across services:
   - Call `trace_cross_service_flow(query="...", source_repo="...", target_repo="...")` on `cb-ask`.
   - Synthesize findings with clear endpoint contracts, parameter breakdown, and hop sequences.

4. **Step 4: Deep AST Code Inspection**
   - For specific method implementation, query `query_codebase_symbols` and `get_symbol_context` on `cb-ask`.
