package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestWriteTasksFrom(t *testing.T) {
	tasks := map[string]Task{
		"T1": {ID: "T1", Title: "one"},
		"T2": {ID: "T2", Title: "two"},
		"T3": {ID: "T3", Title: "three"},
		"T4": {ID: "T4", Title: "four"},
		"T5": {ID: "T5", Title: "five"},
		"T6": {ID: "T6", Title: "six"},
	}
	for _, c := range []struct {
		name   string
		blocks map[string][]string
		want   string
	}{{
		name: "single",
		want: `T1: one
`,
	}, {
		name:   "chain",
		blocks: map[string][]string{"T1": {"T2"}, "T2": {"T3"}},
		want: `
T1: one
└── T2: two
    └── T3: three
`,
	}, {
		name:   "fan-out",
		blocks: map[string][]string{"T1": {"T3", "T2"}},
		want: `
T1: one
├── T2: two
└── T3: three
`,
	}, {
		name:   "continuation",
		blocks: map[string][]string{"T1": {"T2", "T3"}, "T2": {"T4"}},
		want: `
T1: one
├── T2: two
│   └── T4: four
└── T3: three
`,
	}, {
		name:   "shared",
		blocks: map[string][]string{"T1": {"T2", "T3"}, "T2": {"T4"}, "T3": {"T4"}},
		want: `
T1: one
├── T2: two
│   └── T4: four
└── T3: three
    └── T4
`,
	}, {
		name:   "nested-continuation",
		blocks: map[string][]string{"T1": {"T2", "T3"}, "T2": {"T4", "T5"}, "T4": {"T6"}},
		want: `
T1: one
├── T2: two
│   ├── T4: four
│   │   └── T6: six
│   └── T5: five
└── T3: three
`,
	}} {
		t.Run(c.name, func(t *testing.T) {
			var b strings.Builder
			writeTasksFrom(&b, "T1", tasks, c.blocks)

			want := strings.TrimPrefix(c.want, "\n")
			if got := b.String(); got != want {
				t.Errorf("writeTasksFrom wrote:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestCompareIDs(t *testing.T) {
	check := func(a, b string, want int) {
		t.Run(fmt.Sprintf("%s-%s", a, b), func(t *testing.T) {
			got := compareIDs(a, b)
			switch {
			case want < 0 && got >= 0:
				t.Errorf("compareIDs(%q, %q) = %d, want negative", a, b, got)
			case want == 0 && got != 0:
				t.Errorf("compareIDs(%q, %q) = %d, want 0", a, b, got)
			case want > 0 && got <= 0:
				t.Errorf("compareIDs(%q, %q) = %d, want positive", a, b, got)
			}
		})
	}
	for _, c := range []struct {
		a, b string
		want int
	}{
		{a: "T1", b: "T1", want: 0},
		{a: "T2", b: "T10", want: -1},
		{a: "T12", b: "T21", want: -1},
		{a: "T100", b: "T99", want: 1},
		{a: "A9", b: "T1", want: -1},
		{a: "AB1", b: "A2", want: 1},
		{a: "T", b: "T1", want: -1},
		{a: "A", b: "T", want: -1},
	} {
		check(c.a, c.b, c.want)
		if c.want == 0 {
			continue
		}
		check(c.b, c.a, -c.want)
	}
}
