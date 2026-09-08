# Cross-Service Flow Navigation & Query Strategy

When answering questions that span multiple repositories or services:

1. **Topology Discovery**:
   - Query `trace_cross_service_flow(query="...")` on `cb-ask` or `get_related_repos(repo_name="...")` on `cb-indexer`.
   - Identify the source/caller service, target/callee service, communication type (`api_call`, `depends_on`, `shared_resource`), and destination port.

2. **AST Graph Tracing**:
   - Inspect caller outbound HTTP/client invocations in the source repo.
   - Inspect callee controller/router handlers in the destination repo.
   - Read exact code signatures with `get_symbol_context` on `cb-ask`.

3. **Synthesis Styles**:
   - **Targeted Code Inquiry**: Provide exact file links, function signatures, validation logic, and DTO structures.
   - **End-to-End Flow**: Provide step-by-step hops (Client -> Route -> Auth Middleware -> Controller -> Database / Event).
   - **Topology Overview**: Display a clear matrix showing service names, tech stacks, listening ports, and connection types.
