package acron

import "github.com/robfig/cron/v3"

type CronScheduler interface {
	AddFunc(name string, spec string, cmd func()) (cron.EntryID, error)
}

type ACron struct {
	*cron.Cron
	WrapperFunc func(name string, c func())
}

func NewAcron(wrapperFunc func(name string, c func())) *ACron {
	return &ACron{
		Cron: cron.New(
			cron.WithSeconds(),
			cron.WithChain(
				cron.SkipIfStillRunning(cron.DefaultLogger),
			),
			cron.WithChain(
				cron.Recover(cron.DefaultLogger),
			),
		),
		WrapperFunc: wrapperFunc,
	}
}

func (c *ACron) AddFunc(name string, spec string, cmd func()) (cron.EntryID, error) {
	wrappedCmd := cmd
	if c.WrapperFunc != nil {
		wrappedCmd = func() {
			c.WrapperFunc(name, cmd)
		}
	}
	return c.Cron.AddFunc(spec, wrappedCmd)
}
