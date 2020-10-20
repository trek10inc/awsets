package option

type StatusUpdate struct {
	Lister    string
	Region    string
	Error     error
	TotalJobs int
}
