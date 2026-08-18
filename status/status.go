package status

type Status int

const (
	Passed Status = iota
	Failed
	Review  // manual assessment needed
	Skipped // premise doesn't hold (notApplicable)
	Error   // couldn't determine (open)
)

func (s Status) String() string {
	switch s {
	case Passed:
		return "passed"
	case Failed:
		return "failed"
	case Review:
		return "review"
	case Skipped:
		return "skipped"
	case Error:
		return "error"
	}
	return "unknown"
}
