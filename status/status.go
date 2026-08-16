package status

type Status int

const (
	Passed Status = iota
	Failed
	Review  // manual assessment needed
	Skipped // premise doesn't hold (notApplicable)
	Error   // couldn't determine (open)
)
