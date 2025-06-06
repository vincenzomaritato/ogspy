// ogspy - Lightweight CLI tool to inspect, validate and monitor Open Graph metadata.
// Build:  go build -trimpath -ldflags "-s -w" -o ogspy
// License: MIT © 2025 Vincenzo Maritato
//
// Description
// ============================================================================
//   ogcli is a single‑binary command‑line utility aimed at developers, SEO
//   specialists and content creators who need to ensure that their links render
//   perfectly on social platforms. The program fetches a URL, extracts its Open
//   Graph (OG) tags and offers three primary commands:
//
//     • inspect   – Print every OG tag as a coloured table or raw JSON
//     • validate  – Exit with a non‑zero status code when required tags are missing
//     • monitor   – Watch a URL at a configurable interval and report tag diffs
//
//   The binary embeds an explicit user‑agent string, performs HTTP requests with
//   timeouts and relies on minimal external dependencies (goquery for HTML
//   parsing and cobra for the CLI). ogcli exits with non‑zero codes on HTTP
//   errors, network timeouts or missing tags, making it ideal for automation
//   within CI/CD pipelines.
//
//   Author:  Vincenzo Maritato
// ============================================================================

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// ------------------------------------------------------------------------------------------------
// Constants & Global Variables
// ------------------------------------------------------------------------------------------------

const (
	version        = "1.0.0"
	defaultTimeout = 10 * time.Second // Default HTTP request timeout
	userAgent      = "OGSPY/" + version + " (https://github.com/vincenzomaritato/ogspy)"
)

var (
	// essentialTags must be present for a shareable preview to work correctly.
	essentialTags = []string{"title", "type", "image", "url", "description"}

	// recommendedTags is the superset checked by the default validation command.
	recommendedTags = append(append([]string{}, essentialTags...), "site_name", "locale", "video", "audio", "article:author", "article:publisher", "article:section", "article:tag")
)

// logger is populated in newRootCmd().PersistentPreRun.
var logger *slog.Logger

// isTerminal reports whether stdout is a terminal; colour output should be disabled otherwise.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// checkImage downloads the image and validates size, dimensions and aspect‑ratio.
func checkImage(imgURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// HEAD first to check size
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, imgURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot HEAD og:image: %w", err)
	}
	resp.Body.Close()
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		if size, _ := strconv.ParseInt(cl, 10, 64); size > 5*1024*1024 {
			return fmt.Errorf("og:image is larger than 5 MB")
		}
	}

	// Download full image (limit 5 MB)
	resp, err = http.Get(imgURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return err
	}
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("cannot decode og:image: %w", err)
	}
	if img.Width < 1200 || img.Height < 630 {
		return fmt.Errorf("og:image resolution too small (%dx%d)", img.Width, img.Height)
	}
	ratio := float64(img.Width) / float64(img.Height)
	if math.Abs(ratio-1.91) > 0.1 {
		return fmt.Errorf("og:image aspect ratio %.2f deviates from 1.91:1", ratio)
	}
	return nil
}

// semanticValidate returns warnings about advanced semantic rules.
func semanticValidate(og map[string]string) []string {
	var warns []string

	if imgURL, ok := og["image"]; ok && imgURL != "" {
		if !strings.HasPrefix(imgURL, "https://") {
			warns = append(warns, "og:image should use HTTPS")
		}
		if err := checkImage(imgURL); err != nil {
			warns = append(warns, err.Error())
		}
	}

	switch og["type"] {
	case "article":
		if og["article:author"] == "" {
			warns = append(warns, "article:author is missing")
		}
		if og["article:section"] == "" {
			warns = append(warns, "article:section is missing")
		}
	}

	return warns
}

// ------------------------------------------------------------------------------------------------
// HTTP Layer
// ------------------------------------------------------------------------------------------------

// fetchHTML performs a GET request with context/timeout management and returns
// the retrieved HTML document as a string.
func fetchHTML(ctx context.Context, url string) (string, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	html, err := doc.Html()
	if logger != nil {
		logger.Debug("http.fetch",
			slog.String("url", url),
			slog.Int("status", resp.StatusCode),
			slog.Duration("elapsed", time.Since(start)),
		)
	}
	return html, err
}

// ------------------------------------------------------------------------------------------------
// OG Parsing Utilities
// ------------------------------------------------------------------------------------------------

// parseOG walks the HTML document and extracts every meta tag whose name or
// property attribute starts with "og:"; the returned map is keyed without the
// "og:" prefix (e.g. "og:title" becomes "title").
func parseOG(html string) map[string]string {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	og := make(map[string]string)

	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		// Handle <meta property="og:..." content="...">
		if prop, ok := s.Attr("property"); ok && strings.HasPrefix(prop, "og:") {
			if content, ok := s.Attr("content"); ok {
				og[strings.TrimPrefix(prop, "og:")] = content
			}
		}

		// Handle <meta name="og:..." content="...">
		if name, ok := s.Attr("name"); ok && strings.HasPrefix(name, "og:") {
			if content, ok := s.Attr("content"); ok {
				og[strings.TrimPrefix(name, "og:")] = content
			}
		}
	})
	return og
}

// diffMaps returns the set of keys that differ between two OG maps; for each
// differing key the tuple (old, new) is stored.
func diffMaps(old, new map[string]string) map[string][2]string {
	diff := make(map[string][2]string)

	// Build a unified keyset
	keys := make(map[string]struct{})
	for k := range old {
		keys[k] = struct{}{}
	}
	for k := range new {
		keys[k] = struct{}{}
	}

	// Compare values for each key
	for k := range keys {
		if old[k] != new[k] {
			diff[k] = [2]string{old[k], new[k]}
		}
	}
	return diff
}

// ------------------------------------------------------------------------------------------------
// Presentation Helpers
// ------------------------------------------------------------------------------------------------

// printTable renders the OG map as a compact, colourised table with a header.
func printTable(og map[string]string) {
	keys := make([]string, 0, len(og))
	for k := range og {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	header := color.New(color.FgHiWhite, color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", header("Property            Value"))
	fmt.Println(strings.Repeat("─", 40))

	cyan := color.New(color.FgCyan, color.Bold)
	for _, k := range keys {
		cyan.Printf("og:%-15s", k)
		fmt.Printf(" %s\n", og[k])
	}
}

// printMissing highlights absent tags and returns an exit‑code‑style integer
// (0 when all tags are present, 1 otherwise).
func printMissing(og map[string]string, essentialsOnly bool) int {
	required := recommendedTags
	if essentialsOnly {
		required = essentialTags
	}

	missing := make([]string, 0)
	for _, k := range required {
		if og[k] == "" {
			missing = append(missing, "og:"+k)
		}
	}

	if len(missing) > 0 {
		color.New(color.FgRed, color.Bold).Printf("\n✘ Missing Open Graph tags (%d):\n", len(missing))
		for _, tag := range missing {
			fmt.Printf("  • %s\n", tag)
		}
		return 1
	}

	color.New(color.FgGreen, color.Bold).Println("✔ All required tags are present.")
	return 0
}

// printUnified renders a unified diff (à la git) for a given OG diff map.
func printUnified(diff map[string][2]string) {
	for k, v := range diff {
		fmt.Printf("@@ og:%s @@\n", k)
		if v[0] != "" {
			fmt.Printf("- %s\n", v[0])
		}
		if v[1] != "" {
			fmt.Printf("+ %s\n", v[1])
		}
	}
}

// ------------------------------------------------------------------------------------------------
// Cobra Command Definitions
// ------------------------------------------------------------------------------------------------

func newRootCmd() *cobra.Command {
	var noColor bool
	var logJSON bool
	var logLevel string

	cmd := &cobra.Command{
		Use:     "ogspy",
		Short:   "Lightweight CLI tool to inspect, validate and monitor Open Graph metadata.",
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Auto-disable colour when requested or when stdout is not a TTY.
			if noColor || !isTerminal() || os.Getenv("NO_COLOR") != "" {
				color.NoColor = true
			}

			// Configure structured logging ---------------------------------
			lvl := slog.LevelInfo
			switch strings.ToLower(logLevel) {
			case "debug":
				lvl = slog.LevelDebug
			case "warn", "warning":
				lvl = slog.LevelWarn
			case "error":
				lvl = slog.LevelError
			}
			var handler slog.Handler
			if logJSON || !isTerminal() {
				handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
			} else {
				handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
			}
			logger = slog.New(handler)
			slog.SetDefault(logger)
		},
	}

	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable coloured output")
	cmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "Emit logs as newline-delimited JSON")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	cmd.AddCommand(newInspectCmd(), newValidateCmd(), newMonitorCmd())
	return cmd
}

// ------------------------------------------------------------------------------------------------
// Inspect Command (concurrent worker‑pool)
// ------------------------------------------------------------------------------------------------
func newInspectCmd() *cobra.Command {
	var jsonOut bool
	var timeout int
	var workers int

	c := &cobra.Command{
		Use:   "inspect URL [URL...]",
		Short: "Inspect Open Graph metadata for one or many URLs (use “-” to read from STDIN)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Collect URLs from args / STDIN
			var urls []string
			for _, a := range args {
				if a == "-" {
					scanner := bufio.NewScanner(os.Stdin)
					for scanner.Scan() {
						line := strings.TrimSpace(scanner.Text())
						if line != "" {
							urls = append(urls, line)
						}
					}
					if err := scanner.Err(); err != nil {
						return err
					}
				} else {
					urls = append(urls, a)
				}
			}
			if len(urls) == 0 {
				return errors.New("no URLs provided")
			}

			// Worker‑pool setup
			type result struct {
				url string
				og  map[string]string
				err error
			}
			tasks := make(chan string)
			results := make(chan result)
			var wg sync.WaitGroup

			if workers <= 0 {
				workers = runtime.NumCPU()
			}
			if workers > len(urls) {
				workers = len(urls)
			}

			// Spawn workers
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for u := range tasks {
						ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
						html, err := fetchHTML(ctx, u)
						cancel()
						if err != nil {
							results <- result{url: u, err: err}
							continue
						}
						results <- result{url: u, og: parseOG(html)}
					}
				}()
			}

			// Feed tasks
			go func() {
				for _, u := range urls {
					tasks <- u
				}
				close(tasks)
			}()

			// Close results when all workers return
			go func() {
				wg.Wait()
				close(results)
			}()

			exitCode := 0
			aggregated := make(map[string]map[string]string)

			for r := range results {
				if r.err != nil {
					color.Red("Error fetching %s: %v", r.url, r.err)
					exitCode = 1
					continue
				}
				if jsonOut {
					aggregated[r.url] = r.og
				} else {
					color.New(color.FgMagenta, color.Bold).Printf("\n[%s]\n", r.url)
					printTable(r.og)
					fmt.Println()
					printMissing(r.og, false)
				}
			}

			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(aggregated); err != nil {
					return err
				}
			}
			if exitCode != 0 {
				return errors.New("one or more URLs failed inspection")
			}
			return nil
		},
	}

	c.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output raw JSON instead of a table")
	c.Flags().IntVarP(&timeout, "timeout", "t", int(defaultTimeout.Seconds()), "HTTP timeout in seconds")
	c.Flags().IntVarP(&workers, "workers", "w", runtime.NumCPU(), "Number of concurrent workers")
	return c
}

// ------------------------------------------------------------------------------------------------
// Validate Command
// ------------------------------------------------------------------------------------------------

func newValidateCmd() *cobra.Command {
	var essentialsOnly bool
	var timeout int
	var semantic bool

	c := &cobra.Command{
		Use:   "validate URL",
		Short: "Exit with status 1 if required OG tags are missing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			html, err := fetchHTML(ctx, args[0])
			if err != nil {
				return err
			}
			og := parseOG(html)
			if semantic {
				warns := semanticValidate(og)
				for _, w := range warns {
					color.New(color.FgYellow).Printf("⚠ %s\n", w)
				}
			}

			if code := printMissing(og, essentialsOnly); code != 0 {
				return errors.New("required tags are missing")
			}
			return nil
		},
	}

	c.Flags().BoolVarP(&essentialsOnly, "essentials", "e", false, "Validate only essential tags (title, type, image, url, description)")
	c.Flags().IntVarP(&timeout, "timeout", "t", int(defaultTimeout.Seconds()), "HTTP timeout in seconds")
	c.Flags().BoolVarP(&semantic, "semantic", "s", false, "Enable advanced semantic validation")
	return c
}

// ------------------------------------------------------------------------------------------------
// Monitor Command (non‑blocking fetch/render pipeline)
// ------------------------------------------------------------------------------------------------
func newMonitorCmd() *cobra.Command {
	var interval int
	var timeout int
	var jsonDiff bool
	var unified bool

	c := &cobra.Command{
		Use:   "monitor URL",
		Short: "Watch the URL and report any OG tag changes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			type event struct {
				ts string
				og map[string]string
			}
			diffChan := make(chan event)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			color.New(color.FgYellow, color.Bold).Printf("Monitoring %s every %d seconds… (Ctrl+C to stop)\n", url, interval)

			// Fetch + diff worker loop
			go func() {
				var prev map[string]string
				for {
					select {
					case <-ctx.Done():
						close(diffChan)
						return
					case <-ticker.C:
						go func(p map[string]string) {
							fetchCtx, cancelFetch := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
							html, err := fetchHTML(fetchCtx, url)
							cancelFetch()
							if err != nil {
								color.Red("Error: %v", err)
								return
							}
							og := parseOG(html)
							diffChan <- event{
								ts: time.Now().UTC().Format(time.RFC3339),
								og: og,
							}
						}(prev)
						// prev is updated once the event is processed in main goroutine
					}
				}
			}()

			// Render loop (non‑blocking)
			var prev map[string]string
			for ev := range diffChan {
				diff := diffMaps(prev, ev.og)
				if len(diff) > 0 {
					switch {
					case jsonDiff:
						payload := map[string]interface{}{"timestamp": ev.ts, "diff": diff}
						enc := json.NewEncoder(os.Stdout)
						enc.SetIndent("", "  ")
						_ = enc.Encode(payload)
					case unified:
						printUnified(diff)
					default:
						color.New(color.FgYellow, color.Bold).Printf("\n🕒 %s – %d change(s) detected\n", ev.ts, len(diff))
						for k, v := range diff {
							color.New(color.FgCyan, color.Bold).Printf("og:%s", k)
							fmt.Print(" ")
							color.Red(v[0])
							fmt.Print(" → ")
							color.Green(v[1])
							fmt.Println()
						}
					}
				}
				prev = ev.og
			}
			return nil
		},
	}

	c.Flags().IntVarP(&interval, "interval", "i", 300, "Seconds between successive checks")
	c.Flags().IntVarP(&timeout, "timeout", "t", int(defaultTimeout.Seconds()), "HTTP timeout in seconds")
	c.Flags().BoolVarP(&jsonDiff, "json-diff", "j", false, "Print the diff as JSON instead of coloured text")
	c.Flags().BoolVarP(&unified, "unified", "u", false, "Print diff in unified format")
	return c
}

// ------------------------------------------------------------------------------------------------
// Program Entry Point
// ------------------------------------------------------------------------------------------------

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
