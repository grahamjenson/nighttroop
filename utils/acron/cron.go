package acron

import "github.com/robfig/cron/v3"

type CronScheduler interface {
	AddFunc(name string, spec string, cmd func()) (cron.EntryID, error)
}

type CronWrapper struct {
	*cron.Cron
	WrapperFunc func(name string, c func())
}

func (c *CronWrapper) AddFunc(name string, spec string, cmd func()) (cron.EntryID, error) {
	wrappedCmd := cmd
	if c.WrapperFunc != nil {
		wrappedCmd = func() {
			c.WrapperFunc(name, cmd)
		}
	}
	return c.Cron.AddFunc(spec, wrappedCmd)
}
