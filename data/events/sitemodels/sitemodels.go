// Package sitemodels is the bus of changes to the models in each site's database, keyed by site ID, so live
// updates and later alarms can follow a site's data however it was written
package sitemodels

import "nighttroop.t1d.nz/data/events"

// Action is what happened to a model
type Action string

const (
	Created Action = "created"
	Updated Action = "updated"
	Deleted Action = "deleted"
)

// Event is a change to one of a site's models, published once it is committed.
// Model is the sitedb row as stored, e.g. *sitedb.GlucoseReading or *sitedb.EventGroupsBundle; a deleted
// model is the row as it was before.
type Event struct {
	SiteID string
	Action Action
	Model  any
}

// Bus is subscribed to by site ID
type Bus = events.Bus[string, Event]

func NewBus() *Bus { return events.NewBus[string, Event]() }
