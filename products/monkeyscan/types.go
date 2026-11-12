package monkeyscan

import "time"

const (
	productName        = "monkeyscan"
	defaultURL         = "https://monkeyscan-ai.com"
	envAPIKeyName      = "MONKEYSCAN_API_KEY"
	envURLName         = "MONKEYSCAN_URL"
	reviewsDir         = ".monkeyscan/reviews"
	maxReviewBodySize  = 5 * 1024 * 1024
	maxArchiveFileBody = 100 * 1024 * 1024
	maxArchiveEntries  = 50000
	maxArchiveDepth    = 100
	maxArchiveSize     = 1024 * 1024 * 1024
	maxArchiveFileSize = 1024 * 1024 * 1024
)

type Config struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type keySource string

const (
	keySourceNone   keySource = "none"
	keySourceEnv    keySource = "environment"
	keySourceConfig keySource = "config"
)

type statusResponse struct {
	Authenticated         bool        `json:"authenticated"`
	Account               accountInfo `json:"account"`
	GitHubBound           bool        `json:"github_bound"`
	GitHubAccount         string      `json:"github_account"`
	SyncedRepositoryCount int         `json:"synced_repository_count"`
	Ready                 bool        `json:"ready"`
	ReviewReady           bool        `json:"review_ready"`
	FullScanReady         bool        `json:"full_scan_ready"`
	Messages              []string    `json:"messages"`
}

type accountInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type diffReviewRequest struct {
	ClientRunID  string        `json:"client_run_id,omitempty"`
	Command      string        `json:"command,omitempty"`
	Scope        string        `json:"scope,omitempty"`
	GitRemoteURL string        `json:"git_remote_url"`
	BaseBranch   string        `json:"base_branch,omitempty"`
	BaseSHA      string        `json:"base_sha,omitempty"`
	HeadBranch   string        `json:"head_branch,omitempty"`
	HeadSHA      string        `json:"head_sha,omitempty"`
	Diff         string        `json:"diff"`
	Files        []changedFile `json:"files,omitempty"`
}

type changedFile struct {
	Path      string  `json:"path"`
	OldPath   *string `json:"old_path,omitempty"`
	Status    string  `json:"status,omitempty"`
	Additions int     `json:"additions,omitempty"`
	Deletions int     `json:"deletions,omitempty"`
	Changes   int     `json:"changes,omitempty"`
	Patch     string  `json:"patch,omitempty"`
}

type diffReviewResponse struct {
	RunID       string `json:"run_id"`
	TaskGroupID string `json:"task_group_id"`
	Status      string `json:"status"`
}

type scanCreateRequest struct {
	RepoURL  string `json:"repo_url"`
	Branch   string `json:"branch,omitempty"`
	TaskName string `json:"task_name,omitempty"`
}

type scanCreateResponse struct {
	TaskGroupID  string     `json:"task_group_id"`
	Status       string     `json:"status"`
	ScanTarget   scanTarget `json:"scan_target"`
	NextCommands []string   `json:"next_commands"`
}

type scanTarget struct {
	Type string `json:"type"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
	Ref  string `json:"ref,omitempty"`
}

type scanListResponse struct {
	Items    []scanListItem `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	HasNext  bool           `json:"has_next"`
}

type scanListItem struct {
	TaskGroupID           string     `json:"task_group_id"`
	Status                string     `json:"status"`
	ScanTarget            scanTarget `json:"scan_target"`
	DefectCount           int        `json:"defect_count"`
	CriticalSeverityCount int        `json:"critical_severity_count"`
	HighSeverityCount     int        `json:"high_severity_count"`
	MediumSeverityCount   int        `json:"medium_severity_count"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type scanResultResponse struct {
	Task     scanListItem `json:"task"`
	Severity struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
	} `json:"severity"`
	Total     int          `json:"total"`
	Items     []scanDefect `json:"items"`
	Full      bool         `json:"full"`
	Truncated bool         `json:"truncated"`
}

type scanDefect struct {
	ID             string `json:"id"`
	RuleName       string `json:"ruleName"`
	DefectName     string `json:"defectName"`
	DefectNameZh   string `json:"defectNameZh,omitempty"`
	Severity       string `json:"severity"`
	FilePath       string `json:"filepath,omitempty"`
	Line           int    `json:"line,omitempty"`
	Message        string `json:"message,omitempty"`
	MessageZh      string `json:"messageZh,omitempty"`
	CheckerMessage string `json:"checkerMessage,omitempty"`
	AiVerification string `json:"aiVerification,omitempty"`
	AiSuggestion   string `json:"aiSuggestion,omitempty"`
	AiVerifyStatus int    `json:"aiVerifyStatus"`
	Pipeline       int    `json:"pipeline"`
	Confidence     string `json:"confidence,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
	CodeSnippet    string `json:"codeSnippet,omitempty"`
}

type reviewDetail struct {
	Run      reviewRun       `json:"run"`
	Findings []reviewFinding `json:"findings"`
	Comments []reviewComment `json:"comments"`
}

type reviewRun struct {
	ID                    string    `json:"id"`
	RepositoryName        string    `json:"repository_name"`
	TaskGroupID           string    `json:"task_group_id"`
	TaskGroupStatus       string    `json:"task_group_status"`
	Status                string    `json:"status"`
	StatusReason          string    `json:"status_reason"`
	BaseBranch            string    `json:"base_branch"`
	HeadBranch            string    `json:"head_branch"`
	HeadSHA               string    `json:"head_sha"`
	FindingCount          int       `json:"finding_count"`
	PublishedCommentCount int       `json:"published_comment_count"`
	FailedCommentCount    int       `json:"failed_comment_count"`
	PublishStatus         string    `json:"publish_status"`
	ErrorMessage          string    `json:"error_message"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type reviewFinding struct {
	ID               string          `json:"id"`
	Severity         string          `json:"severity"`
	Category         string          `json:"category"`
	Confidence       string          `json:"confidence"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Location         findingLocation `json:"location"`
	ProblemCode      string          `json:"problem_code"`
	Recommendation   string          `json:"recommendation"`
	SuggestedDiff    string          `json:"suggested_diff"`
	RecommendedDiff  string          `json:"recommended_diff"`
	FixDiff          string          `json:"fix_diff"`
	SuggestedPatch   string          `json:"suggested_patch"`
	ResolutionStatus string          `json:"resolution_status"`
}

type findingLocation struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Side      string `json:"side"`
}

type reviewComment struct {
	ID            string `json:"id"`
	CommentType   string `json:"comment_type"`
	PublishStatus string `json:"publish_status"`
	Body          string `json:"body"`
	ErrorMessage  string `json:"error_message"`
}

type localStatus struct {
	ClientRunID  string    `json:"client_run_id"`
	RunID        string    `json:"run_id,omitempty"`
	TaskGroupID  string    `json:"task_group_id,omitempty"`
	Scope        string    `json:"scope"`
	Status       string    `json:"status"`
	BaseBranch   string    `json:"base_branch,omitempty"`
	BaseSHA      string    `json:"base_sha,omitempty"`
	HeadBranch   string    `json:"head_branch,omitempty"`
	HeadSHA      string    `json:"head_sha,omitempty"`
	ReviewPath   string    `json:"review_path,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type reviewScope struct {
	Type       string
	All        bool
	Base       string
	BaseCommit string
}

type scanOptions struct {
	Path        string
	File        string
	Repo        string
	TaskGroupID string
	Branch      string
	Wait        bool
	Full        bool
	Output      string
}

type diffSnapshot struct {
	RepoRoot      string
	RemoteURL     string
	CurrentBranch string
	HeadSHA       string
	BaseBranch    string
	BaseSHA       string
	Scope         string
	Diff          string
	Files         []changedFile
}
