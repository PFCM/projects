// binary projects manages projects.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var (
	dirFlag = flag.String("dir", ".", "`path` to project directory")
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fatalf("no command specified\n")
	}
	switch cmd := flag.Arg(0); cmd {
	case "overview":
		if err := overview(*dirFlag); err != nil {
			fatalf("%v\n", err)
		}
	default:
		fatalf("unknown command %q\n", cmd)
	}
}

func fatalf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func overview(dir string) error {
	tasks, err := readTasks(dir)
	if err != nil {
		return err
	}

	var (
		byID   = make(map[string]Task, len(tasks))
		blocks = make(map[string][]string) // id -> ids of tasks it blocks
		roots  []string                    // tasks with no blockers
	)
	for _, t := range tasks {
		byID[t.ID] = t
		for _, b := range t.BlockedBy {
			blocks[b] = append(blocks[b], t.ID)
		}
		if len(t.BlockedBy) == 0 {
			roots = append(roots, t.ID)
		}
	}
	slices.SortFunc(roots, compareIDs)

	for _, root := range roots {
		writeTasksFrom(os.Stdout, root, byID, blocks)
	}
	return nil
}

func writeTasksFrom(w io.Writer, root string, tasks map[string]Task, blocks map[string][]string) {
	type frame struct {
		id     string
		line   string
		indent string
	}
	var (
		todo = []frame{{id: root}}
		seen = make(map[string]bool)
	)
	for len(todo) > 0 {
		cur := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		fmt.Fprint(w, cur.line)
		if seen[cur.id] {
			fmt.Fprintln(w, cur.id)
			continue
		}
		seen[cur.id] = true
		fmt.Fprintf(w, "%s: %s\n", cur.id, tasks[cur.id].Title)
		// Backwards, because they're pushed to a stack and popped in reverse.
		children := blocks[cur.id]
		slices.SortFunc(children, func(a, b string) int {
			return -compareIDs(a, b)
		})
		for i, child := range children {
			branch, indent := "├── ", "│   "
			if i == 0 {
				branch, indent = "└── ", "    "
			}
			todo = append(todo, frame{
				id:     child,
				line:   cur.indent + branch,
				indent: cur.indent + indent,
			})
		}
	}
}

func compareIDs(a, b string) int {
	ap, an := splitID(a)
	bp, bn := splitID(b)
	if c := strings.Compare(ap, bp); c != 0 {
		return c
	}
	return an - bn
}

func splitID(id string) (string, int) {
	// TODO: big assumption that isn't super clearly documented
	i := strings.IndexFunc(id, unicode.IsDigit)
	if i < 0 {
		return id, 0
	}
	n, err := strconv.Atoi(id[i:])
	if err != nil {
		panic(err)
	}
	return id[:i], n
}
