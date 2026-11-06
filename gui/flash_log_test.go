package gui

import "testing"

func TestLastNonJSONLogLine(t *testing.T) {
	tests := []struct {
		name string
		logs []string
		want string
	}{
		{
			name: "fyne locale noise only",
			logs: []string{
				"2026/11/06 17:01:09 Fyne error:  Failed to load user locales",
				"2026/11/06 17:01:09   Cause: exit status 1",
				"2026/11/06 17:01:09   At: /Users/runner/go/pkg/mod/fyne.io/fyne/v2@v2.7.2/lang/lang.go:205",
			},
			want: "",
		},
		{
			name: "error line wins over fyne noise and progress",
			logs: []string{
				"2026/11/06 17:01:09 Fyne error:  Failed to load user locales",
				"2026/11/06 17:01:09   At: /x/lang.go:205",
				`{"phase":"extracting","processed":0}`,
				"Error: gzip: invalid header",
			},
			want: "gzip: invalid header",
		},
		{
			name: "error line wins over later output",
			logs: []string{"Error: device busy", "Warning: failed to eject device: x"},
			want: "device busy",
		},
		{
			name: "falls back to last plain line",
			logs: []string{"something odd", `{"phase":"writing"}`},
			want: "something odd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{lastLogs: tt.logs}
			if got := a.lastNonJSONLogLine(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
