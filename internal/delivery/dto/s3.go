package dto

type FileStatus struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Bucket   string `json:"bucket,omitempty"`
	Error    string `json:"error,omitempty"`
}

type IssueTemporaryURLRequest struct {
	Links []string `json:"filepath"`
}

type IssueTemporaryURLResponse struct {
	TemporaryURL string `json:"filepath,omitempty"`
	Error        string `json:"error,omitempty"`
}
