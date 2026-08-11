package verdict

type Verdict int

const (
	Passed Verdict = iota
	Failed
	Review  // manual assessment needed
	Skipped // premise doesn't hold (notApplicable)
	Error   // couldn't determine (open)
)
