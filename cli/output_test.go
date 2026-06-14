package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/render"
)

// TestOutDefaultFormat pins the one place linkedin differs from kit's default:
// with no -o, a terminal gets the readable list view and a pipe gets jsonl, but
// an explicit -o or --template always wins.
func TestOutDefaultFormat(t *testing.T) {
	cases := []struct {
		name string
		out  kit.OutputOptions
		want render.Format
	}{
		{"tty default is list", kit.OutputOptions{IsTTY: true}, render.List},
		{"piped default is jsonl", kit.OutputOptions{IsTTY: false}, render.JSONL},
		{"explicit table on a tty wins", kit.OutputOptions{Format: "table", IsTTY: true}, render.Table},
		{"explicit json piped wins", kit.OutputOptions{Format: "json", IsTTY: false}, render.JSON},
		{"template forces template", kit.OutputOptions{Template: "{{.url}}", IsTTY: true}, render.Template},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := &App{st: &kit.State{Output: c.out}}
			r, err := a.out()
			if err != nil {
				t.Fatalf("out(): %v", err)
			}
			if got := r.Format(); got != c.want {
				t.Errorf("Format() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestRenderContract spot-checks the shared renderer against the row shapes the
// commands feed it, so a future kit bump that changes formatting is caught here.
func TestRenderContract(t *testing.T) {
	row := Row{
		Cols:  []string{"slug", "name", "url"},
		Vals:  []string{"acme", "Acme", "https://example.com/acme"},
		Value: map[string]any{"slug": "acme", "name": "Acme", "url": "https://example.com/acme"},
	}

	t.Run("jsonl one line per record", func(t *testing.T) {
		out := renderRows(t, render.Options{Format: render.JSONL}, row, row)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(lines) != 2 {
			t.Fatalf("got %d lines, want 2: %q", len(lines), out)
		}
		if !strings.Contains(lines[0], `"slug":"acme"`) {
			t.Errorf("line 0 = %q", lines[0])
		}
	})

	t.Run("csv with header", func(t *testing.T) {
		out := renderRows(t, render.Options{Format: render.CSV}, row)
		if !strings.HasPrefix(out, "slug,name,url\n") {
			t.Errorf("csv header missing: %q", out)
		}
		if !strings.Contains(out, "acme,Acme,") {
			t.Errorf("csv body missing: %q", out)
		}
	})

	t.Run("url emits the url column", func(t *testing.T) {
		out := renderRows(t, render.Options{Format: render.URL}, row)
		if strings.TrimSpace(out) != "https://example.com/acme" {
			t.Errorf("url = %q", out)
		}
	})

	t.Run("template per record", func(t *testing.T) {
		out := renderRows(t, render.Options{Template: "{{.slug}}"}, row)
		if strings.TrimSpace(out) != "acme" {
			t.Errorf("template = %q", out)
		}
	})
}

func renderRows(t *testing.T, o render.Options, rows ...Row) string {
	t.Helper()
	var buf bytes.Buffer
	o.Writer = &buf
	r, err := render.New(o)
	if err != nil {
		t.Fatalf("render.New: %v", err)
	}
	for _, row := range rows {
		if err := r.Emit(row); err != nil {
			t.Fatalf("Emit: %v", err)
		}
	}
	if err := r.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}
