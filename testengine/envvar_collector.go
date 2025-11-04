package testengine

type EnvVarCollector struct {
	collectorFunc func() map[string]string
	EnvVars       map[string]string
}

type EnvVarCollectorOption func(*EnvVarCollector)

func NewEnvVarCollector(options ...EnvVarCollectorOption) *EnvVarCollector {
	ec := &EnvVarCollector{
		EnvVars: make(map[string]string),
		collectorFunc: func() map[string]string {
			return map[string]string{}
		},
	}
	for _, option := range options {
		option(ec)
	}
	return ec
}

func WithCollectorFunc(collectorFunc func() map[string]string) EnvVarCollectorOption {
	return func(ec *EnvVarCollector) {
		ec.collectorFunc = collectorFunc
	}
}

func (c *EnvVarCollector) CollectAllEnVars() map[string]string {
	c.EnvVars = c.collectorFunc()
	return c.EnvVars
}

func (c *EnvVarCollector) GetEnvAllVars() map[string]string {
	return c.EnvVars
}
