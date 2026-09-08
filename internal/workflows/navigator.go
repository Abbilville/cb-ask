package workflows

import (
	"context"
	"fmt"
	"strings"

	"cb-ask/internal/cbmclient"
	"cb-ask/internal/idxclient"
)

// CrossServiceHop represents a single hop in a cross-service communication path.
type CrossServiceHop struct {
	HopNumber   int                   `json:"hop_number"`
	Source      string                `json:"source"`
	Target      string                `json:"target"`
	Type        string                `json:"type"`
	Description string                `json:"description"`
	Port        *int                  `json:"port,omitempty"`
	CallerCode  []cbmclient.CbmSymbol `json:"caller_symbols,omitempty"`
	CalleeCode  []cbmclient.CbmSymbol `json:"callee_symbols,omitempty"`
}

// CrossServiceFlowReport is the composite response returned by trace_cross_service_flow.
type CrossServiceFlowReport struct {
	Query         string                  `json:"query"`
	ProjectID     string                  `json:"project_id"`
	TopologyEdges []idxclient.RelationDTO `json:"topology_edges"`
	Hops          []CrossServiceHop       `json:"hops"`
	Summary       string                  `json:"summary"`
}

// TraceCrossServiceFlow merges topology data from cb-indexer with AST graph symbols from codebase-memory-mcp.
func TraceCrossServiceFlow(
	ctx context.Context,
	idx *idxclient.IndexerClient,
	cbm *cbmclient.CbmClient,
	query, sourceRepo, targetRepo, project string,
) (*CrossServiceFlowReport, error) {
	// 1. Fetch architecture overview from cb-indexer
	overview, err := idx.GetArchitectureOverview(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch topology from cb-indexer: %w", err)
	}

	var relevantRels []idxclient.RelationDTO
	for _, rel := range overview.Relationships {
		matchSource := sourceRepo == "" || strings.EqualFold(rel.Source, sourceRepo)
		matchTarget := targetRepo == "" || strings.EqualFold(rel.Target, targetRepo)
		if matchSource && matchTarget {
			relevantRels = append(relevantRels, rel)
		}
	}

	// If no specific source/target provided and no exact match, filter by query terms
	if len(relevantRels) == 0 && (sourceRepo == "" || targetRepo == "") {
		qLower := strings.ToLower(query)
		for _, rel := range overview.Relationships {
			if strings.Contains(qLower, strings.ToLower(rel.Source)) ||
				strings.Contains(qLower, strings.ToLower(rel.Target)) ||
				strings.Contains(strings.ToLower(rel.Description), qLower) {
				relevantRels = append(relevantRels, rel)
			}
		}
	}

	if len(relevantRels) == 0 && len(overview.Relationships) > 0 {
		relevantRels = overview.Relationships
	}

	repoPortMap := make(map[string]*int)
	for _, r := range overview.Repos {
		repoPortMap[strings.ToLower(r.Name)] = r.Port
	}

	var hops []CrossServiceHop
	for i, rel := range relevantRels {
		hop := CrossServiceHop{
			HopNumber:   i + 1,
			Source:      rel.Source,
			Target:      rel.Target,
			Type:        rel.Type,
			Description: rel.Description,
			Port:        repoPortMap[strings.ToLower(rel.Target)],
		}

		// 1. First try querying AST graph symbols from local codebase-memory-mcp if available
		var callerSyms []cbmclient.CbmSymbol
		var calleeSyms []cbmclient.CbmSymbol

		if cbm != nil {
			callerSyms, _ = cbm.SearchGraph(ctx, ".*api.*|.*client.*|.*fetch.*", rel.Source)
			calleeSyms, _ = cbm.SearchGraph(ctx, ".*handler.*|.*controller.*|.*route.*", rel.Target)
		}

		// 2. If local CBM is not installed or returned no symbols, query remote cb-indexer RAG
		if len(callerSyms) == 0 && idx != nil {
			if res, err := idx.QueryCodebaseSymbols(ctx, "api", rel.Source, "", 5); err == nil && len(res.Symbols) > 0 {
				for _, s := range res.Symbols {
					callerSyms = append(callerSyms, cbmclient.CbmSymbol{
						Name:          s.Name,
						QualifiedName: s.QualifiedName,
						Type:          s.Label,
						FilePath:      s.FilePath,
						LineNumber:    s.StartLine,
					})
				}
			}
		}

		if len(calleeSyms) == 0 && idx != nil {
			if res, err := idx.QueryCodebaseSymbols(ctx, "handler", rel.Target, "", 5); err == nil && len(res.Symbols) > 0 {
				for _, s := range res.Symbols {
					calleeSyms = append(calleeSyms, cbmclient.CbmSymbol{
						Name:          s.Name,
						QualifiedName: s.QualifiedName,
						Type:          s.Label,
						FilePath:      s.FilePath,
						LineNumber:    s.StartLine,
					})
				}
			}
		}

		if len(callerSyms) > 5 {
			callerSyms = callerSyms[:5]
		}
		if len(calleeSyms) > 5 {
			calleeSyms = calleeSyms[:5]
		}

		hop.CallerCode = callerSyms
		hop.CalleeCode = calleeSyms

		hops = append(hops, hop)
	}

	summary := fmt.Sprintf("Traced %d cross-service hop(s) across %d repositories in project '%s'",
		len(hops), len(overview.Repos), overview.ProjectID)

	return &CrossServiceFlowReport{
		Query:         query,
		ProjectID:     overview.ProjectID,
		TopologyEdges: relevantRels,
		Hops:          hops,
		Summary:       summary,
	}, nil
}
