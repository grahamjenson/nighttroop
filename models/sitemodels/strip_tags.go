package sitemodels

import "nighttroop.t1d.nz/utils/types"

// StripTags is text an app sent, a note, reason or notification, for the site, without its HTML tags or the contents
// of a script, so no app's text can put a script on the site's pages, whatever shows it. It is kept, and given back
// through the Nightscout API, as sent.
func StripTags(text *string) *string {
	if text == nil {
		return nil
	}
	return new(types.StripAllTags(*text))
}

// StripTagsValue is StripTags of text that is always set
func StripTagsValue(text string) string {
	return types.StripAllTags(text)
}
