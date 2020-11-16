package context

type StatusType string

const (
	StatusLogInfo           StatusType = "logInfo"
	StatusLogDebug          StatusType = "logDebug"
	StatusLogError          StatusType = "logError"
	StatusProcessing        StatusType = "processing"
	StatusComplete          StatusType = "complete"
	StatusCompleteWithError StatusType = "completeWithError"
)

type StatusUpdate struct {
	Type      StatusType
	Lister    string
	Region    string
	Message   string
	WorkerId  int
	TotalJobs int
}
