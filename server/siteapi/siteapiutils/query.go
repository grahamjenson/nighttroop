package siteapiutils

import "net/url"

type Query struct {
}

func QueryFromParams(values url.Values) *Query {
	return &Query{}
}
