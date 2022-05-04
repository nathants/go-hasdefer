package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// TODO is it worth using ast instead of strings, even with go fmt regularity?

func main() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && (os.Args[0] == "-h" || os.Args[0] == "--help" || os.Args[0] == "help")) {
		fmt.Println("\nlinter to check that all goroutines have a defer statement")
		fmt.Println("\nusage: go-hasdefer $(find -type f -name '*.go')")
		os.Exit(1)
	}
	fail := false
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
			text = regexp.MustCompile("(?s)`(.*?)`").ReplaceAllString(text, "``")
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
				l = regexp.MustCompile("(?s)`(.*?)`").ReplaceAllString(l, "``")
				l = regexp.MustCompile(`//.*`).ReplaceAllString(l, `//`)
				lines[filePath] = append(lines[filePath], l)
			}
		}
	}
	for _, filePath := range os.Args {
		if strings.HasSuffix(filePath, ".go") {
			for i, line := range lines[filePath] {
				if regexp.MustCompile(`( |\t)go `).FindAllString(line, -1) != nil {
					if regexp.MustCompile(`( |\t)go func\b`).FindAllString(line, -1) != nil {
						if strings.HasSuffix(line, "{") {
							if !strings.HasPrefix(strings.TrimLeft(lines[filePath][i+1], "\t"), "defer") {
								fmt.Println("missing defer anon func multiliner:     ", filePath+":"+fmt.Sprint(i+1), line)
								fail = true
							}
						} else {
							if regexp.MustCompile(`\bdefer\b`).FindAllString(line, -1) == nil {
								fmt.Println("missing defer anon func oneliner:       ", filePath+":"+fmt.Sprint(i+1), line)
								fail = true
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
						// TODO index source with a map instead of walking repeatedly
						for fp, lns := range lines {
							for j, l := range lns {
								if strings.HasPrefix(l, "func ") && regexp.MustCompile(`\b`+funcName+`\b`).FindAllString(l, -1) != nil {
									found = true
									if strings.HasSuffix(l, "{") {
										if !strings.HasPrefix(strings.TrimLeft(lines[fp][j+1], "\t"), "defer") {
											fmt.Println("missing defer top level func multiliner:", fp+":"+fmt.Sprint(j+1), l)
											fail = true
										}
									} else {
										if regexp.MustCompile(`\bdefer\b`).FindAllString(l, -1) == nil {
											fmt.Println("missing defer top level func oneliner:  ", fp+":"+fmt.Sprint(j+1), l)
											fail = true
										}
									}
								} else if strings.HasPrefix(strings.TrimLeft(l, "\t"), funcName+" := func(") {
									found = true
									if strings.HasSuffix(l, "{") {
										if !strings.HasPrefix(strings.TrimLeft(lines[fp][j+1], "\t"), "defer") {
											fmt.Println("missing defer named func multiliner:    ", filePath+":"+fmt.Sprint(j+1), l)
											fail = true
										}
									} else {
										if regexp.MustCompile(`\bdefer\b`).FindAllString(l, -1) == nil {
											fmt.Println("missing defer named func oneliner:      ", filePath+":"+fmt.Sprint(j+1), l)
											fail = true
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
	if fail {
		os.Exit(1)
	}
}
