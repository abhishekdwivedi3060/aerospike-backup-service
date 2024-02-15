package model

type JobStatus string

const (
	JobStatusRunning JobStatus = "Running"
	JobStatusDone    JobStatus = "Done"
	JobStatusFailed  JobStatus = "Failed"
)

// RestoreJobStatus represents a restore job status.
// @Description RestoreJobStatus represents a restore job status.
type RestoreJobStatus struct {
	RestoreResult
	Status JobStatus `yaml:"status,omitempty" json:"status,omitempty" enums:"Running,Done,Failed"`
	Error  error     `yaml:"error,omitempty" json:"error,omitempty"`
}

// RestoreResult represents a single restore operation result.
type RestoreResult struct {
	TotalRecords    int `yaml:"total-records,omitempty" json:"total-records,omitempty" example:"10"`
	TotalBytes      int `yaml:"total-bytes,omitempty" json:"total-bytes,omitempty" example:"2000"`
	ExpiredRecords  int `yaml:"expired-records,omitempty" json:"expired-records,omitempty" example:"2"`
	SkippedRecords  int `yaml:"skipped-records,omitempty" json:"skipped-records,omitempty" example:"4"`
	IgnoredRecords  int `yaml:"ignored-records,omitempty" json:"ignored-records,omitempty" example:"12"`
	InsertedRecords int `yaml:"inserted-records,omitempty" json:"inserted-records,omitempty" example:"8"`
	ExistedRecords  int `yaml:"existed-records,omitempty" json:"existed-records,omitempty" example:"15"`
	FresherRecords  int `yaml:"fresher-records,omitempty" json:"fresher-records,omitempty" example:"5"`
	IndexCount      int `yaml:"index-count,omitempty" json:"index-count,omitempty" example:"3"`
	UDFCount        int `yaml:"udf-count,omitempty" json:"udf-count,omitempty" example:"1"`
}

// NewRestoreResult returns a new RestoreResult.
func NewRestoreResult() *RestoreResult {
	return &RestoreResult{}
}

// NewRestoreJobStatus returns a new RestoreJobStatus.
func NewRestoreJobStatus() *RestoreJobStatus {
	return &RestoreJobStatus{
		Status: JobStatusRunning,
	}
}
