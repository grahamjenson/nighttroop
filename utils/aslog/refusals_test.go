package aslog

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// summary is a flushed summary log's attributes
type summary struct {
	Reason string     `json:"reason"`
	Status int        `json:"status"`
	Total  int        `json:"total"`
	Keys   int        `json:"keys"`
	Other  int        `json:"other"`
	Top    []KeyCount `json:"top"`
}

// logToBuffer makes the default logger hold its logs in a buffer of its own, as the logs DB is written from, for the
// test
func logToBuffer(t *testing.T) *Buffer {
	buffer, before := &Buffer{Full: make(chan struct{}, 1)}, slog.Default()
	slog.SetDefault(slog.New(ContextHandler{BufferHandler{Buffer: buffer}}))
	t.Cleanup(func() { slog.SetDefault(before) })
	return buffer
}

// flushed flushes refusals, and is the summaries it logged, by reason then status
func flushed(t *testing.T, refusals *Refusals, buffer *Buffer) []summary {
	refusals.Flush(context.Background())
	records, _ := buffer.Take()
	summaries := []summary{}
	for _, record := range records {
		require.Equal(t, "info", record.Level)
		require.Equal(t, "Unauthenticated requests refused", record.Message)
		var s summary
		require.NoError(t, json.Unmarshal([]byte(record.Attributes), &s))
		summaries = append(summaries, s)
	}
	slices.SortFunc(summaries, func(a, b summary) int {
		return cmp.Or(cmp.Compare(a.Reason, b.Reason), cmp.Compare(a.Status, b.Status))
	})
	return summaries
}

// Refusals are counted by reason and status, then by key, the most refused first; a key-less one counts only in the
// total. Past maxRefusalKeys keys of a reason the rest count as other, so a sweep of every subdomain is fixed memory.
// A flush logs one summary for each, at info, and starts again, and logs nothing when nothing was refused.
func TestRefusalsSummariseEachMinute(t *testing.T) {
	buffer, refusals := logToBuffer(t), &Refusals{}
	require.Empty(t, flushed(t, refusals, buffer), "nothing refused, nothing logged")

	for i := range maxRefusalKeys + 50 {
		refusals.Add("site does not exist", 410, fmt.Sprintf("site%03d", i))
	}
	for range 3 {
		refusals.Add("site does not exist", 410, "site099")
		refusals.Add("site does not exist", 410, "site150")
	}
	for range 2 {
		refusals.Add("site does not exist", 410, "site007")
		refusals.Add("site does not exist", 404, "site007")
	}
	refusals.Add("rate limited", 429, "")
	refusals.Add("rate limited", 429, strings.Repeat("k", 100))

	summaries := flushed(t, refusals, buffer)
	require.Len(t, summaries, 3)
	require.Equal(t, summary{Reason: "rate limited", Status: 429, Total: 2, Keys: 1,
		Top: []KeyCount{{strings.Repeat("k", maxRefusalKeyBytes-3) + "...", 1}}}, summaries[0])
	require.Equal(t, summary{Reason: "site does not exist", Status: 404, Total: 2, Keys: 1, Top: []KeyCount{{"site007", 2}}}, summaries[1])

	missing := summaries[2]
	require.Equal(t, 410, missing.Status)
	require.Equal(t, maxRefusalKeys+50+8, missing.Total)
	require.Equal(t, maxRefusalKeys, missing.Keys)
	require.Equal(t, 50+3, missing.Other, "keys past the cap, site150 among them, are other")
	require.Len(t, missing.Top, topRefusalKeys)
	require.Equal(t, []KeyCount{{"site099", 4}, {"site007", 3}, {"site000", 1}}, missing.Top[:3])

	require.Empty(t, flushed(t, refusals, buffer), "a flush starts counting again")
}

// Requests are refused at once, so refusals are counted at once
func TestRefusalsCountConcurrently(t *testing.T) {
	buffer, refusals := logToBuffer(t), &Refusals{}
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			for j := range 100 {
				refusals.Add("site does not exist", 410, fmt.Sprintf("site%d", (i*100+j)%150))
			}
		})
	}
	wg.Wait()
	summaries := flushed(t, refusals, buffer)
	require.Len(t, summaries, 1)
	require.Equal(t, 2000, summaries[0].Total)
	require.Equal(t, maxRefusalKeys, summaries[0].Keys)
	require.Positive(t, summaries[0].Other, "the keys past the first 100 seen are other")
}

// Refuse logs at debug, which the logs DB does not keep, and counts the refusal by its message
func TestRefuseLogsAtDebug(t *testing.T) {
	buffer := logToBuffer(t)
	before := RefusedRequests
	RefusedRequests = &Refusals{}
	t.Cleanup(func() { RefusedRequests = before })

	Refuse(context.Background(), 401, "/api/v1/entries", "No token", "error", "missing")
	records, _ := buffer.Take()
	require.Empty(t, records)
	summaries := flushed(t, RefusedRequests, buffer)
	require.Equal(t, []summary{{Reason: "No token", Status: 401, Total: 1, Keys: 1, Top: []KeyCount{{"/api/v1/entries", 1}}}}, summaries)
}
