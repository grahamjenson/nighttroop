package alog

import "log/slog"

func CaptureError(err error) {
	if err != nil {
		slog.Error("Error", "error", err)
	}
}
