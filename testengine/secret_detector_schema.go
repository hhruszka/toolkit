package testengine

type SecretDetectorResult struct {
	name       string
	value      string
	file       bool
	readable   bool
	confidence int
	sectype    string
}

func NewSecretDetectorResult(name string, value string, file bool, readable bool, confidence int, stype string) *SecretDetectorResult {
	return &SecretDetectorResult{
		name:       name,
		value:      value,
		file:       file,
		readable:   readable,
		confidence: confidence,
		sectype:    stype,
	}
}

func (cr *SecretDetectorResult) Name() string {
	return cr.name
}

func (cr *SecretDetectorResult) Value() string {
	return cr.value
}

func (cr *SecretDetectorResult) File() bool {
	return cr.file
}

func (cr *SecretDetectorResult) Readable() bool {
	return cr.readable
}

func (cr *SecretDetectorResult) Confidence() int {
	return cr.confidence
}

func (cr *SecretDetectorResult) Type() string {
	return cr.sectype
}
