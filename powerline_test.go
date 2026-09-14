package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	pwl "github.com/justjanne/powerline-go/powerline"
)

func Test_detectShell(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "",
			want: "bare",
		},
		{
			name: "/bin/sh",
			want: "bare",
		}, {
			name: "/bin/bash",
			want: "bash",
		}, {
			name: "/usr/local/bin/bash5",
			want: "bash",
		}, {
			name: "/usr/bin/zsh",
			want: "zsh",
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectShell(tt.name); got != tt.want {
				t.Errorf("detectShell(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func newTestPowerline(shell string, align alignment) *powerline {
	p := &powerline{
		cfg:     defaults,
		shell:   defaults.Shells[shell],
		theme:   defaults.Themes["default"],
		symbols: defaults.Modes["patched"],
		align:   align,
	}
	p.reset = fmt.Sprintf(p.shell.ColorTemplate, "[0m")
	p.Segments = [][]pwl.Segment{{{
		Name:                "exit",
		Content:             "SIGPIPE",
		Foreground:          p.theme.CmdFailedFg,
		Background:          p.theme.CmdFailedBg,
		Separator:           p.symbols.Separator,
		SeparatorForeground: p.theme.CmdFailedBg,
	}}}
	return p
}

// A row that soft-wraps on the last screen line makes the terminal scroll in
// a line filled with the background colour active at that moment. Every
// left-prompt row must therefore end by erasing to the end of the line once
// colours are reset, so that leftover cells get the default background.
func Test_drawRowClearsToEndOfLine(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		align alignment
		want  string // expected suffix; empty means the sequence must be absent
	}{
		{"bash", "bash", alignLeft, `\[\e[K\] `},
		{"zsh", "zsh", alignLeft, "%{\x1b[K%} "},
		{"bare", "bare", alignLeft, "\x1b[K "},
		{"zsh right prompt", "zsh", alignRight, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestPowerline(tt.shell, tt.align)
			var buf bytes.Buffer
			p.drawRow(0, &buf)
			got := buf.String()
			if tt.want == "" {
				if strings.Contains(got, "[K") {
					t.Errorf("drawRow() = %q, right prompt must not clear to end of line", got)
				}
			} else if !strings.HasSuffix(got, tt.want) {
				t.Errorf("drawRow() = %q, want suffix %q", got, tt.want)
			}
		})
	}
}
