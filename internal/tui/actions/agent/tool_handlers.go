package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/biisal/godo/internal/bus"
	"github.com/biisal/godo/internal/config"
	"github.com/biisal/godo/internal/tui/actions/todo"
	"github.com/gocolly/colly/v2"
	"github.com/openai/openai-go"
)

func runPerformSql(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	result, err := todo.PerformSqlQuery(args.Query)
	if err != nil {
		return "", false, err
	}
	return result, true, nil
}

func runShellCommand(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}

	bus.EmitShell("$ " + args.Command + "\n")

	const timeout = 60 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", args.Command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", false, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", false, err
	}

	if err := cmd.Start(); err != nil {
		return "", false, err
	}

	var fullOutput strings.Builder
	outputChan := make(chan string)
	var wg sync.WaitGroup

	readOutput := func(r io.Reader) {
		defer wg.Done()
		buf := make([]byte, 512)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				outputChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}

	wg.Add(2)
	go readOutput(stdout)
	go readOutput(stderr)

	go func() {
		_ = cmd.Wait()
		wg.Wait()
		close(outputChan)
	}()

	for out := range outputChan {
		fullOutput.WriteString(out)
		bus.EmitShell(out)
	}

	output := fullOutput.String()

	// If the context timed out, return an error so the agent knows.
	if ctx.Err() == context.DeadlineExceeded {
		msg := fmt.Sprintf("command timed out after %s", timeout)
		fullOutput.WriteString("\n" + msg + "\n")
		bus.EmitShell(msg + "\n")
		slog.Warn("shell command timed out", "command", args.Command, "timeout", timeout)
		return fullOutput.String(), false, errors.New(msg)
	}

	slog.Debug("command output completed", "command", args.Command)
	return output, false, nil
}

func runReadSkill(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		SkillName string `json:"skillName"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("Could not determine user home directory for skills", "err", err)
		homeDir, _ = os.Getwd()
	}
	skillPath := filepath.Join(homeDir, config.AppDIR, "content", "skills", args.SkillName+".md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read skill %s: %w", args.SkillName, err)
	}
	return string(content), false, nil
}

func runGlobSearch(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Pattern string `json:"pattern"`
		Root    string `json:"root"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	args.Pattern = strings.TrimSpace(args.Pattern)
	if args.Pattern == "" {
		return "", false, fmt.Errorf("pattern is required")
	}
	if args.Root == "" {
		args.Root = "."
	}

	rootAbs, err := filepath.Abs(args.Root)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve root %q: %w", args.Root, err)
	}

	pattern := filepath.ToSlash(filepath.Clean(args.Pattern))
	const maxMatches = 500
	matches := make([]string, 0, 64)

	err = filepath.WalkDir(rootAbs, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootAbs, p)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if globMatch(pattern, rel) {
			matches = append(matches, rel)
			if len(matches) >= maxMatches {
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil && err != fs.SkipAll {
		return "", false, fmt.Errorf("glob search failed: %w", err)
	}

	return map[string]any{
		"root":      rootAbs,
		"pattern":   args.Pattern,
		"count":     len(matches),
		"truncated": len(matches) >= maxMatches,
		"matches":   matches,
	}, false, nil
}

func runReadFiles(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Paths           []string `json:"paths"`
		MaxBytesPerFile int      `json:"maxBytesPerFile"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	if len(args.Paths) == 0 {
		return "", false, fmt.Errorf("paths is required")
	}
	if args.MaxBytesPerFile <= 0 {
		args.MaxBytesPerFile = 64 * 1024
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get current directory: %w", err)
	}

	results := make([]map[string]any, 0, len(args.Paths))
	for _, rawPath := range args.Paths {
		pathArg := strings.TrimSpace(rawPath)
		if pathArg != "" {
			bus.EmitShell("reading " + pathArg + "\n")
		}
		if pathArg == "" {
			bus.EmitShell("error: empty path\n")
			results = append(results, map[string]any{
				"path":  rawPath,
				"error": "empty path",
			})
			continue
		}

		targetPath := pathArg
		if !filepath.IsAbs(targetPath) {
			targetPath = filepath.Join(cwd, targetPath)
		}
		targetPath = filepath.Clean(targetPath)

		info, statErr := os.Stat(targetPath)
		if statErr != nil {
			bus.EmitShell("error: " + statErr.Error() + "\n")
			results = append(results, map[string]any{
				"path":  pathArg,
				"error": statErr.Error(),
			})
			continue
		}
		if info.IsDir() {
			bus.EmitShell("error: path is a directory\n")
			results = append(results, map[string]any{
				"path":  pathArg,
				"error": "path is a directory",
			})
			continue
		}

		data, readErr := os.ReadFile(targetPath)
		if readErr != nil {
			bus.EmitShell("error: " + readErr.Error() + "\n")
			results = append(results, map[string]any{
				"path":  pathArg,
				"error": readErr.Error(),
			})
			continue
		}

		truncated := false
		if len(data) > args.MaxBytesPerFile {
			data = data[:args.MaxBytesPerFile]
			truncated = true
		}

		results = append(results, map[string]any{
			"path":      pathArg,
			"sizeBytes": info.Size(),
			"truncated": truncated,
			"content":   string(data),
		})
		bus.EmitShell(string(data) + "\n")
	}

	return map[string]any{
		"count":   len(results),
		"results": results,
	}, false, nil
}

func runWriteFile(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Path          string `json:"path"`
		Content       string `json:"content"`
		CreateParents *bool  `json:"createParents"`
		Append        *bool  `json:"append"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}

	pathArg := strings.TrimSpace(args.Path)
	if pathArg == "" {
		return "", false, fmt.Errorf("path is required")
	}
	createParents := args.CreateParents != nil && *args.CreateParents
	appendMode := args.Append != nil && *args.Append

	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get current directory: %w", err)
	}

	targetPath := pathArg
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(cwd, targetPath)
	}
	targetPath = filepath.Clean(targetPath)

	if createParents {
		parentDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			return "", false, fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	flags := os.O_WRONLY | os.O_CREATE
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(targetPath, flags, 0o644)
	if err != nil {
		return "", false, fmt.Errorf("failed to open file for writing: %w", err)
	}

	bytesWritten, writeErr := file.WriteString(args.Content)
	closeErr := file.Close()
	if writeErr != nil {
		return "", false, fmt.Errorf("failed to write file: %w", writeErr)
	}
	if closeErr != nil {
		return "", false, fmt.Errorf("failed to close file: %w", closeErr)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to stat written file: %w", err)
	}

	return map[string]any{
		"success":        true,
		"path":           targetPath,
		"bytesWritten":   bytesWritten,
		"fileSizeBytes":  info.Size(),
		"appended":       appendMode,
		"createdParents": createParents,
	}, false, nil
}

func runEditFile(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Path       string `json:"path"`
		OldString  string `json:"oldString"`
		NewString  string `json:"newString"`
		LineNumber int    `json:"lineNumber"`
		NewContent string `json:"newContent"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	pathArg := strings.TrimSpace(args.Path)
	if pathArg == "" {
		return "", false, fmt.Errorf("path is required")
	}

	targetPath, err := resolvePath(pathArg)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(data)

	useStringMode := args.OldString != "" || args.NewString != ""
	useLineMode := args.LineNumber > 0 || args.NewContent != ""

	if useStringMode && useLineMode {
		return "", false, fmt.Errorf("provide either oldString/newString or lineNumber/newContent, not both")
	}
	if !useStringMode && !useLineMode {
		return "", false, fmt.Errorf("provide either oldString/newString or lineNumber/newContent")
	}

	var newFileContent string
	mode := ""
	if useStringMode {
		if args.OldString == "" {
			return "", false, fmt.Errorf("oldString is required for string replacement mode")
		}
		if !strings.Contains(content, args.OldString) {
			return "", false, fmt.Errorf("oldString not found in file")
		}
		newFileContent = strings.Replace(content, args.OldString, args.NewString, 1)
		mode = "string_replace"
	} else {
		if args.LineNumber <= 0 {
			return "", false, fmt.Errorf("lineNumber must be >= 1 for line replacement mode")
		}
		lines := strings.Split(content, "\n")
		if args.LineNumber > len(lines) {
			return "", false, fmt.Errorf("lineNumber out of range: %d (file has %d lines)", args.LineNumber, len(lines))
		}
		lines[args.LineNumber-1] = args.NewContent
		newFileContent = strings.Join(lines, "\n")
		mode = "line_replace"
	}

	if err := os.WriteFile(targetPath, []byte(newFileContent), 0o644); err != nil {
		return "", false, fmt.Errorf("failed to write edited file: %w", err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to stat edited file: %w", err)
	}
	return map[string]any{
		"success":       true,
		"path":          targetPath,
		"mode":          mode,
		"fileSizeBytes": info.Size(),
	}, false, nil
}

func runPatchFile(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Path  string `json:"path"`
		Patch string `json:"patch"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	pathArg := strings.TrimSpace(args.Path)
	if pathArg == "" {
		return "", false, fmt.Errorf("path is required")
	}
	if strings.TrimSpace(args.Patch) == "" {
		return "", false, fmt.Errorf("patch is required")
	}

	targetPath, err := resolvePath(pathArg)
	if err != nil {
		return "", false, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("failed to get current directory: %w", err)
	}

	cmd := exec.Command("git", "apply", "--recount", "--whitespace=nowarn", "-")
	cmd.Dir = cwd
	cmd.Stdin = strings.NewReader(args.Patch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", false, fmt.Errorf("failed to apply patch: %w; output: %s", err, strings.TrimSpace(string(out)))
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("patch applied but stat failed for %s: %w", targetPath, err)
	}
	return map[string]any{
		"success":       true,
		"path":          targetPath,
		"fileSizeBytes": info.Size(),
	}, false, nil
}

func runInsertAtLine(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Path       string `json:"path"`
		LineNumber int    `json:"lineNumber"`
		Content    string `json:"content"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	pathArg := strings.TrimSpace(args.Path)
	if pathArg == "" {
		return "", false, fmt.Errorf("path is required")
	}
	if args.LineNumber <= 0 {
		return "", false, fmt.Errorf("lineNumber must be >= 1")
	}

	targetPath, err := resolvePath(pathArg)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read file: %w", err)
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	if args.LineNumber > len(lines)+1 {
		return "", false, fmt.Errorf("lineNumber out of range: %d (file has %d lines)", args.LineNumber, len(lines))
	}

	insertIdx := args.LineNumber - 1
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, args.Content)
	newLines = append(newLines, lines[insertIdx:]...)
	newFileContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(targetPath, []byte(newFileContent), 0o644); err != nil {
		return "", false, fmt.Errorf("failed to write updated file: %w", err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to stat updated file: %w", err)
	}
	return map[string]any{
		"success":       true,
		"path":          targetPath,
		"lineNumber":    args.LineNumber,
		"fileSizeBytes": info.Size(),
	}, false, nil
}

func resolvePath(pathArg string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	targetPath := pathArg
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(cwd, targetPath)
	}
	return filepath.Clean(targetPath), nil
}

func runProjectTree(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Root         string `json:"root"`
		MaxDepth     int    `json:"maxDepth"`
		IncludeFiles *bool  `json:"includeFiles"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	if args.Root == "" {
		args.Root = "."
	}
	if args.MaxDepth <= 0 {
		args.MaxDepth = 4
	}
	includeFiles := true
	if args.IncludeFiles != nil {
		includeFiles = *args.IncludeFiles
	}

	rootAbs, err := filepath.Abs(args.Root)
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve root %q: %w", args.Root, err)
	}
	bus.EmitShell("building tree for " + rootAbs + "\n")

	const maxEntries = 2000
	lines := make([]string, 0, 256)

	err = filepath.WalkDir(rootAbs, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == rootAbs {
			return nil
		}

		rel, err := filepath.Rel(rootAbs, p)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		depth := strings.Count(rel, "/") + 1
		if depth > args.MaxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !includeFiles && !d.IsDir() {
			return nil
		}

		name := path.Base(rel)
		prefix := strings.Repeat("  ", depth-1) + "- "
		if d.IsDir() {
			lines = append(lines, prefix+name+"/")
		} else {
			lines = append(lines, prefix+name)
		}

		if len(lines) >= maxEntries {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil && err != fs.SkipAll {
		return "", false, fmt.Errorf("tree scan failed: %w", err)
	}

	treeText := strings.Join(lines, "\n")
	bus.EmitShell(treeText + "\n")

	return map[string]any{
		"root":      rootAbs,
		"maxDepth":  args.MaxDepth,
		"entries":   len(lines),
		"truncated": len(lines) >= maxEntries,
		"tree":      treeText,
	}, false, nil
}

func globMatch(pattern, relPath string) bool {
	pattern = strings.Trim(filepath.ToSlash(filepath.Clean(pattern)), "/")
	relPath = strings.Trim(filepath.ToSlash(filepath.Clean(relPath)), "/")
	patternParts := splitPathParts(pattern)
	relParts := splitPathParts(relPath)
	return matchGlobParts(patternParts, relParts)
}

func splitPathParts(p string) []string {
	if p == "" || p == "." {
		return nil
	}
	raw := strings.Split(p, "/")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if part == "" || part == "." {
			continue
		}
		parts = append(parts, part)
	}
	return parts
}

func matchGlobParts(patternParts, pathParts []string) bool {
	if len(patternParts) == 0 {
		return len(pathParts) == 0
	}
	if patternParts[0] == "**" {
		if matchGlobParts(patternParts[1:], pathParts) {
			return true
		}
		if len(pathParts) > 0 {
			return matchGlobParts(patternParts, pathParts[1:])
		}
		return false
	}
	if len(pathParts) == 0 {
		return false
	}
	ok, err := path.Match(patternParts[0], pathParts[0])
	if err != nil || !ok {
		return false
	}
	return matchGlobParts(patternParts[1:], pathParts[1:])
}

func runDuckDuckGoSearch(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		Query      string `json:"query"`
		MaxResults int    `json:"maxResults"`
		Page       int    `json:"page"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	args.Query = strings.TrimSpace(args.Query)
	if args.Query == "" {
		return "", false, fmt.Errorf("query is required")
	}
	if args.MaxResults <= 0 {
		args.MaxResults = 10
	}
	if args.MaxResults > 50 {
		args.MaxResults = 50
	}
	if args.Page <= 0 {
		args.Page = 1
	}
	if args.Page > 20 {
		args.Page = 20
	}

	c := colly.NewCollector(
		colly.UserAgent("godo-agent/1.0 (+https://github.com/biisal/godo)"),
	)
	c.SetRequestTimeout(20 * time.Second)

	statusCode := 0
	var reqErr error
	results := make([]map[string]string, 0, args.MaxResults)

	c.OnResponse(func(r *colly.Response) {
		statusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		reqErr = err
		if r != nil {
			statusCode = r.StatusCode
		}
	})
	c.OnHTML(".result", func(e *colly.HTMLElement) {
		if len(results) >= args.MaxResults {
			return
		}
		title := normalizeSpace(e.ChildText("a.result__a"))
		if title == "" {
			return
		}
		link := cleanDuckDuckGoResultURL(html.UnescapeString(strings.TrimSpace(e.ChildAttr("a.result__a", "href"))))
		desc := normalizeSpace(e.ChildText(".result__snippet"))
		results = append(results, map[string]string{
			"title":       title,
			"url":         link,
			"description": desc,
		})
	})

	formData := map[string]string{
		"q":  args.Query,
		"kl": "us-en",
	}
	if args.Page > 1 {
		offset := (args.Page - 1) * 30
		formData["s"] = fmt.Sprintf("%d", offset)
		formData["dc"] = fmt.Sprintf("%d", offset+1)
	}
	err := c.Post("https://html.duckduckgo.com/html/", formData)
	if err != nil {
		return "", false, fmt.Errorf("duckduckgo request failed: %w", err)
	}
	if reqErr != nil {
		return "", false, fmt.Errorf("duckduckgo request failed: %w", reqErr)
	}

	return map[string]any{
		"query":      args.Query,
		"page":       args.Page,
		"count":      len(results),
		"results":    results,
		"statusCode": statusCode,
	}, false, nil
}

func runScrapePage(tc openai.ChatCompletionMessageToolCall) (any, bool, error) {
	var args struct {
		URL      string `json:"url"`
		MaxChars int    `json:"maxChars"`
	}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("invalid tool arguments: %w", err)
	}
	args.URL = strings.TrimSpace(args.URL)
	if args.URL == "" {
		return "", false, fmt.Errorf("url is required")
	}
	if args.MaxChars <= 0 {
		args.MaxChars = 8000
	}
	if args.MaxChars > 50000 {
		args.MaxChars = 50000
	}

	c := colly.NewCollector(
		colly.UserAgent("godo-agent/1.0 (+https://github.com/biisal/godo)"),
	)
	c.SetRequestTimeout(20 * time.Second)

	statusCode := 0
	finalURL := ""
	var reqErr error
	title := ""
	desc := ""
	text := ""

	c.OnResponse(func(r *colly.Response) {
		statusCode = r.StatusCode
		finalURL = r.Request.URL.String()
	})
	c.OnError(func(r *colly.Response, err error) {
		reqErr = err
		if r != nil {
			statusCode = r.StatusCode
			if r.Request != nil && r.Request.URL != nil {
				finalURL = r.Request.URL.String()
			}
		}
	})
	c.OnHTML("title", func(e *colly.HTMLElement) {
		if title == "" {
			title = normalizeSpace(e.Text)
		}
	})
	c.OnHTML(`meta[name="description"]`, func(e *colly.HTMLElement) {
		if desc == "" {
			desc = normalizeSpace(e.Attr("content"))
		}
	})
	c.OnHTML(`meta[property="og:description"]`, func(e *colly.HTMLElement) {
		if desc == "" {
			desc = normalizeSpace(e.Attr("content"))
		}
	})
	c.OnHTML("body", func(e *colly.HTMLElement) {
		if text == "" {
			text = normalizeSpace(e.Text)
		}
	})

	err := c.Visit(args.URL)
	if err != nil {
		return "", false, fmt.Errorf("page request failed: %w", err)
	}
	if reqErr != nil {
		return "", false, fmt.Errorf("page request failed: %w", reqErr)
	}

	truncated := false
	if len(text) > args.MaxChars {
		text = text[:args.MaxChars]
		truncated = true
	}

	var out strings.Builder
	out.WriteString("URL: " + args.URL + "\n")
	if finalURL != "" && finalURL != args.URL {
		out.WriteString("FinalURL: " + finalURL + "\n")
	}
	fmt.Fprintf(&out, "StatusCode: %d\n", statusCode)
	if title != "" {
		out.WriteString("Title: " + title + "\n")
	}
	if desc != "" {
		out.WriteString("Description: " + desc + "\n")
	}
	if truncated {
		out.WriteString("Truncated: true\n")
	}
	out.WriteString("\n")
	out.WriteString(text)

	return out.String(), false, nil
}

func cleanDuckDuckGoResultURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if strings.HasPrefix(u.Path, "/l/") {
		redirect := u.Query().Get("uddg")
		if redirect != "" {
			decoded, err := url.QueryUnescape(redirect)
			if err == nil {
				return decoded
			}
			return redirect
		}
	}
	return raw
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}
