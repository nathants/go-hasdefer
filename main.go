package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
)

// TODO is it worth using ast instead of strings, even with go fmt regularity?

func emptyMultilineStrings(data []byte) []byte {
outer:
	for {
		outs := regexp.MustCompile("(?s)`(.*?)`").FindAllStringIndex(string(data), -1)
		for _, out := range outs {
			head := data[:out[0]]
			mid := data[out[0]:out[1]]
			empty := true
			for _, b := range mid {
				if b != '\n' && b != '`' {
					empty = false
				}
			}
			if empty {
				continue
			}
			tail := data[out[1]:]
			count := bytes.Count(mid, []byte("\n"))
			mid = []byte("`")
			for i := 0; i < count; i++ {
				mid = append(mid, []byte("\n")...)
			}
			mid = append(mid, []byte("`")...)
			data = []byte{}
			data = append(data, head...)
			data = append(data, mid...)
			data = append(data, tail...)
			continue outer
		}
		return data
	}
}

func color(code int) func(...string) string {
	forced := os.Getenv("COLORS") != ""
	return func(xs ...string) string {
		s := strings.Join(xs, " ")
		if forced || isatty.IsTerminal(os.Stdout.Fd()) {
			return fmt.Sprintf("\033[%dm%s\033[0m", code, s)
		}
		return s
	}
}

var (
	Red     = color(31)
	Green   = color(32)
	Yellow  = color(33)
	Blue    = color(34)
	Magenta = color(35)
	Cyan    = color(36)
	White   = color(37)
)

func main() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && (os.Args[0] == "-h" || os.Args[0] == "--help" || os.Args[0] == "help")) {
		fmt.Println("\nlinter to check that all goroutines have a defer statement")
		fmt.Println("\nusage: go-hasdefer $(find -type f -name '*.go')")
		os.Exit(1)
	}
	lines := make(map[string][]string)
	for _, filePath := range os.Args {
		if strings.HasSuffix(filePath, ".go") {
			var stdout bytes.Buffer
			cmd := exec.Command("gofmt", filePath)
			cmd.Stdout = &stdout
			err := cmd.Run()
			if err != nil {
				fmt.Println("fatal: gofmt failed on:", filePath)
				os.Exit(1)
			}
			text := stdout.String()
			text = string(emptyMultilineStrings([]byte(text)))
			lns := strings.Split(text, "\n")
			lastImport := 0
			inImport := false
			for i, l := range lns {
				if strings.HasPrefix(l, "import") {
					inImport = true
				} else if inImport && !strings.HasPrefix(l, "\t") && l != "" {
					inImport = false
					lastImport = i
				}
			}
			for i, l := range lns {
				if i <= lastImport {
					l = ""
				}
				l = regexp.MustCompile(`"[^"]+"`).ReplaceAllString(l, `""`)
				if !strings.HasPrefix(strings.TrimLeft(l, "\t "), "// defer func() {}()") {
					l = regexp.MustCompile(`//.*`).ReplaceAllString(l, `//`)
				}
				lines[filePath] = append(lines[filePath], l)
			}
		}
	}
	var vals []string
	for _, filePath := range os.Args {
		if strings.HasSuffix(filePath, ".go") {
			for i, line := range lines[filePath] {
				if regexp.MustCompile(`( |\t)go `).FindAllString(line, -1) != nil {
					if regexp.MustCompile(`( |\t)go func\b`).FindAllString(line, -1) != nil {
						if strings.HasSuffix(line, "{") {
							if !strings.HasPrefix(strings.TrimLeft(lines[filePath][i+1], "\t/ "), "defer") {
								vals = append(vals, strings.Join([]string{"missing defer anon func multiliner:     ", Cyan(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
							}
						} else {
							if regexp.MustCompile(`\bdefer\b`).FindAllString(line, -1) == nil {
								vals = append(vals, strings.Join([]string{"missing defer anon func oneliner:       ", Cyan(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
							}
						}
					} else {
						parts := strings.Split(line, "go ")
						funcName := parts[len(parts)-1]
						parts = strings.Split(funcName, "(")
						funcName = parts[0]
						parts = strings.Split(funcName, ".")
						funcName = parts[len(parts)-1]
						found := false
						// check for named functions in the same file
						lns := lines[filePath]
						for j, l := range lns {
							if strings.HasPrefix(strings.TrimLeft(l, "\t"), funcName+" := func(") {
								found = true
								if strings.HasSuffix(l, "{") {
									if !strings.HasPrefix(strings.TrimLeft(lns[j+1], "\t/ "), "defer") {
										vals = append(vals, strings.Join([]string{"missing defer named func multiliner:    ", Cyan(filePath + ":" + fmt.Sprint(j+1)), l, Red("from:"), Green(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
									}
								} else {
									if regexp.MustCompile(`\bdefer\b`).FindAllString(l, -1) == nil {
										vals = append(vals, strings.Join([]string{"missing defer named func oneliner:      ", Cyan(filePath + ":" + fmt.Sprint(j+1)), l, Red("from:"), Green(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
									}
								}
							}
						}
						// check all source code for top level functions
						if !found {
							for fp, lns := range lines {
								for j, l := range lns {
									if strings.HasPrefix(l, "func ") && regexp.MustCompile(`\b`+funcName+`\b`).FindAllString(l, -1) != nil {
										found = true
										if strings.HasSuffix(l, "{") {
											if !strings.HasPrefix(strings.TrimLeft(lines[fp][j+1], "\t/ "), "defer") {
												vals = append(vals, strings.Join([]string{"missing defer top level func multiliner:", Cyan(fp + ":" + fmt.Sprint(j+1)), l, Red("from:"), Green(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
											}
										} else {
											if regexp.MustCompile(`\bdefer\b`).FindAllString(l, -1) == nil {
												vals = append(vals, strings.Join([]string{"missing defer top level func oneliner:  ", Cyan(fp + ":" + fmt.Sprint(j+1)), l, Red("from:"), Green(filePath + ":" + fmt.Sprint(i+1)), line}, " "))
											}
										}
									}
								}
							}
						}
						if !found {
							panic("failed to find definition of function: [" + funcName + "] " + line + " " + filePath)
						}
					}
				}
			}
		}
	}
	seen := make(map[string]interface{})
	for _, val := range vals {
		_, ok := seen[val]
		if !ok {
			fmt.Println(val)
			seen[val] = nil
		}
	}
	if len(vals) != 0 {
		os.Exit(1)
	}
}
