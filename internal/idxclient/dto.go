package idxclient

// RelationDTO represents a relationship between services returned by oss-indexer.
type RelationDTO struct {
	Source      string         `json:"source"`
	Target      string         `json:"target"`
	Type        string         `json:"type"`
	Description string         `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// RepoDTO represents repository details returned by oss-indexer.
type RepoDTO struct {
	Name         string   `json:"name"`
	Owner        string   `json:"owner,omitempty"`
	LocalPath    string   `json:"local_path"`
	Description  string   `json:"description,omitempty"`
	TechStack    []string `json:"tech_stack,omitempty"`
	EntryPoint   string   `json:"entry_point,omitempty"`
	Port         *int     `json:"port,omitempty"`
	IsIndexed    bool     `json:"is_indexed,omitempty"`
	IndexNodes   *int     `json:"index_nodes,omitempty"`
	IndexEdges   *int     `json:"index_edges,omitempty"`
	IndexedAt    *string  `json:"indexed_at,omitempty"`
	IndexProject string   `json:"index_project,omitempty"`
}

// ArchitectureOverviewDTO is a tolerant DTO for get_architecture_overview responses.
type ArchitectureOverviewDTO struct {
	ProjectID          string        `json:"project_id"`
	ProjectName        string        `json:"project_name"`
	Description        string        `json:"description,omitempty"`
	SourcePath         string        `json:"source_path,omitempty"`
	TotalRepos         int           `json:"total_repos"`
	IndexedRepos       int           `json:"indexed_repos"`
	UnindexedRepos     int           `json:"unindexed_repos"`
	TotalRelationships int           `json:"total_relationships"`
	Repos              []RepoDTO     `json:"repos"`
	Relationships      []RelationDTO `json:"relationships,omitempty"`
}

// RepoDetailsDTO is a tolerant DTO for get_repo_details responses.
type RepoDetailsDTO struct {
	ProjectID     string                   `json:"project_id"`
	Repo          RepoDTO                  `json:"repo"`
	Relationships map[string][]RelationDTO `json:"relationships"`
}

// RelatedReposDTO is a tolerant DTO for get_related_repos responses.
type RelatedReposDTO struct {
	Repo             string        `json:"repo"`
	Direction        string        `json:"direction"`
	TotalConnections int           `json:"total_connections"`
	Relationships    []RelationDTO `json:"relationships"`
}

// ProjectSummaryDTO represents a project catalog entry from oss-indexer.
type ProjectSummaryDTO struct {
	ProjectID    string `json:"project_id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	RegistryPath string `json:"registry_path,omitempty"`
	RootPath     string `json:"root_path,omitempty"`
}

// ProjectListDTO is a tolerant DTO for list_projects responses.
type ProjectListDTO struct {
	RegisteredProjects          []ProjectSummaryDTO `json:"registered_projects"`
	IndexedCodebaseMemoryGraphs []map[string]any    `json:"indexed_codebase_memory_graphs"`
	TotalRegisteredProjects     int                 `json:"total_registered_projects"`
	TotalIndexedGraphs          int                 `json:"total_indexed_graphs"`
}
