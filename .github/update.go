package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// no external libraries were used specifically

const (
	staffDirectory  = ".github"
	defaultMetadata = "metadata.json"
)

type fileMap struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type rootMetadata struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref,omitempty"`
}

type hwMetadata struct {
	Ref         string    `json:"ref,omitempty"`
	VersionPath string    `json:"version_path"`
	Files       []fileMap `json:"files"`
}

type repoSpec struct {
	Owner string
	Repo  string
	Ref   string
}

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 180 * time.Second,
}

var transport = &http.Transport{
	DialContext:           dialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	MaxConnsPerHost:       100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   15 * time.Second,
	ExpectContinueTimeout: 2 * time.Second,
	MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
}

var client = http.Client{
	Transport: transport,
	Timeout:   30 * time.Second,
}

var hwPattern = regexp.MustCompile(`^hw[1-9][0-9]*$`)

const courseCITokenEnv = "COURSE_CI_TOKEN"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	metadataPath := filepath.Join(staffDirectory, defaultMetadata)

	force := flag.Bool("force", false, "force update")
	hw := flag.String("hw", "", "homework directory in remote repo")
	flag.Parse()

	if strings.TrimSpace(*hw) == "" {
		fmt.Fprintln(os.Stderr, "error: flag -hw is required")
		os.Exit(1)
	}

	localVersionPath := filepath.Join(staffDirectory, *hw+".version.txt")

	hwDir := strings.TrimSpace(*hw)

	if !hwPattern.MatchString(hwDir) {
		fmt.Fprintf(os.Stderr, "error: invalid -hw value: %q (expected hwN)\n", *hw)
		os.Exit(1)
	}

	rootMd, spec, err := parseRootMetadata(metadataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("> Source: %s/%s", spec.Owner, spec.Repo)
	if spec.Ref != "" {
		fmt.Printf(" @ %s", spec.Ref)
	}
	fmt.Printf("\n> Homework: %s\n", hwDir)

	token := os.Getenv(courseCITokenEnv)

	if token != "" {
		fmt.Println("✓ The github token has been processed")
	}

	cloneURL := sshRepoURL(spec.Owner, spec.Repo)
	var gitToken string

	if token != "" {
		cloneURL = httpsRepoURL(spec.Owner, spec.Repo)
		gitToken = token
		fmt.Println("✓ Clone repo via token")
	} else {
		fmt.Println("✓ Clone repo via SSH")
	}

	remoteMetadataPath := filepath.ToSlash(filepath.Join(hwDir, "metadata.json"))
	metadataRepo, err := cloneSparse(ctx, cloneURL, spec.Ref, remoteMetadataPath, nil, gitToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clone error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(metadataRepo)

	hwMetaBytes, err := os.ReadFile(filepath.Join(metadataRepo, remoteMetadataPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not read homework metadata (%s): %v\n", remoteMetadataPath, err)
		os.Exit(1)
	}

	hwMd, err := parseHWMetadata(hwMetaBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not parse homework metadata (%s): %v\n", remoteMetadataPath, err)
		os.Exit(1)
	}

	effectiveSpec := spec
	if hwMd.Ref != "" {
		effectiveSpec.Ref = hwMd.Ref
	} else if rootMd.Ref != "" {
		effectiveSpec.Ref = rootMd.Ref
	}

	remoteVersionPath := filepath.ToSlash(filepath.Join(hwDir, hwMd.VersionPath))
	remoteFiles := make([]fileMap, 0, len(hwMd.Files))
	for _, fm := range hwMd.Files {
		if strings.TrimSpace(fm.From) == "" || strings.TrimSpace(fm.To) == "" {
			fmt.Fprintln(os.Stderr, "from/to missed in files")
			os.Exit(1)
		}

		remoteFiles = append(remoteFiles, fileMap{
			From: filepath.ToSlash(filepath.Join(hwDir, fm.From)),
			To:   fm.To,
		})
	}

	tmpRepo, err := cloneSparse(ctx, cloneURL, effectiveSpec.Ref, remoteVersionPath, remoteFiles, gitToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "clone error: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpRepo)

	remoteVerBytes, err := os.ReadFile(filepath.Join(tmpRepo, remoteVersionPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not read tests repository version (%s): %v\n", remoteVersionPath, err)
		os.Exit(1)
	}
	remoteVer := strings.TrimSpace(string(remoteVerBytes))

	localVer := ""
	b, err := os.ReadFile(localVersionPath)
	if err == nil {
		localVer = strings.TrimSpace(string(b))
	} else if os.IsNotExist(err) {
		f, createErr := os.Create(localVersionPath)
		if createErr != nil {
			fmt.Fprintf(os.Stderr, "can not create version file (%s): %v\n", localVersionPath, createErr)
			os.Exit(1)
		}
		_ = f.Close()
		localVer = ""
	} else {
		fmt.Fprintf(os.Stderr, "can not read version file (%s): %v\n", localVersionPath, err)
		os.Exit(1)
	}

	if !(*force) && localVer == remoteVer {
		fmt.Printf("✓ The version is up to date (%s)\n", remoteVer)
		return
	}

	if *force {
		fmt.Printf("↺ Force update, ignore local version %q, remote version %q)\n", localVer, remoteVer)
	} else {
		fmt.Printf("↺ Fetch tests: local version %q → remote version %q\n", localVer, remoteVer)
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not get working directory: %v\n", err)
		os.Exit(1)
	}

	for _, fm := range hwMd.Files {
		if strings.TrimSpace(fm.From) == "" || strings.TrimSpace(fm.To) == "" {
			fmt.Fprintln(os.Stderr, "from/to missed in files")
			os.Exit(1)
		}

		remoteFilePath := filepath.ToSlash(filepath.Join(hwDir, fm.From))
		fmt.Printf("→ Fetch: %s\n", remoteFilePath)

		content, err := os.ReadFile(filepath.Join(tmpRepo, remoteFilePath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "  read cloned file %s: %v\n", remoteFilePath, err)
			os.Exit(1)
		}

		dst, err := safeJoin(root, fm.To)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  invalid dst path %q: %v\n", fm.To, err)
			os.Exit(1)
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "  mkdir %s: %v\n", filepath.Dir(dst), err)
			os.Exit(1)
		}

		if err := os.WriteFile(dst, content, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "  write %s: %v\n", dst, err)
			os.Exit(1)
		}

		fmt.Printf("  ✓ Saved: %s\n", relOrSame(root, dst))
	}

	if err := os.MkdirAll(filepath.Dir(localVersionPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", filepath.Dir(localVersionPath), err)
		os.Exit(1)
	}

	if err := os.WriteFile(localVersionPath, []byte(remoteVer+"\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", localVersionPath, err)
		os.Exit(1)
	}

	fmt.Printf("✓ Tests updated: %s → %s\n", localVer, remoteVer)
	fmt.Println("Success!")
}

func parseRootMetadata(path string) (rootMetadata, repoSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return rootMetadata{}, repoSpec{}, err
	}

	var md rootMetadata
	if err = json.Unmarshal(data, &md); err != nil {
		return rootMetadata{}, repoSpec{}, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if md.Repo == "" {
		return rootMetadata{}, repoSpec{}, errors.New("missed repo in metadata.json")
	}

	owner, repo := splitOwnerRepo(md.Repo)
	if owner == "" || repo == "" {
		return rootMetadata{}, repoSpec{}, fmt.Errorf("find empty repository path: %s", md.Repo)
	}

	return md, repoSpec{Owner: owner, Repo: repo, Ref: md.Ref}, nil
}

func parseHWMetadata(data []byte) (hwMetadata, error) {
	var md hwMetadata
	if err := json.Unmarshal(data, &md); err != nil {
		return hwMetadata{}, fmt.Errorf("json.Unmarshal: %w", err)
	}

	if md.VersionPath == "" {
		return hwMetadata{}, errors.New("missed version_path in homework metadata")
	}

	if len(md.Files) == 0 {
		return hwMetadata{}, errors.New("missed files in homework metadata")
	}

	return md, nil
}

func splitOwnerRepo(s string) (string, string) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "git@github.com:")
	s = strings.TrimPrefix(s, "https://github.com/")
	s = strings.TrimPrefix(s, "http://github.com/")

	if i := strings.Index(s, "#"); i >= 0 {
		s = s[:i]
	}

	parts := strings.Split(s, "/")
	if len(parts) >= 2 {
		return parts[0], strings.TrimSuffix(parts[1], ".git")
	}

	return "", ""
}

func escapePathSegments(p string) string {
	parts := strings.Split(p, "/")

	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}

	return strings.Join(parts, "/")
}

func cloneSparse(ctx context.Context, repo, ref, versionPath string, files []fileMap, token string) (string, error) {
	tmp, err := os.MkdirTemp("", "course-tests-*")
	if err != nil {
		return "", err
	}

	var cleanup func()
	env := os.Environ()

	if token != "" {
		env, cleanup, err = gitHTTPSAuthEnv(token)
		if err != nil {
			os.RemoveAll(tmp)
			return "", err
		}
		defer cleanup()
	}

	args := []string{
		"clone",
		"--depth", "1",
		"--filter=blob:none",
		"--no-checkout",
	}

	if ref != "" {
		args = append(args, "--branch", ref)
	}

	args = append(args, repo, tmp)

	if out, err := runGit(ctx, env, args...); err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("git clone failed: %w\n%s", err, out)
	}

	if out, err := runGit(ctx, env, "-C", tmp, "sparse-checkout", "init", "--no-cone"); err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("git sparse-checkout init failed: %w\n%s", err, out)
	}

	sparseArgs := []string{"-C", tmp, "sparse-checkout", "set", "--no-cone"}

	for _, f := range files {
		if from := strings.TrimSpace(f.From); from != "" {
			sparseArgs = append(sparseArgs, from)
		}
	}

	if versionPath = strings.TrimSpace(versionPath); versionPath != "" {
		sparseArgs = append(sparseArgs, versionPath)
	}

	if out, err := runGit(ctx, env, sparseArgs...); err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("git sparse-checkout set failed: %w\n%s", err, out)
	}

	if out, err := runGit(ctx, env, "-C", tmp, "checkout"); err != nil {
		os.RemoveAll(tmp)
		return "", fmt.Errorf("git checkout failed: %w\n%s", err, out)
	}

	return tmp, nil
}

func runGit(ctx context.Context, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

func gitHTTPSAuthEnv(token string) ([]string, func(), error) {
	dir, err := os.MkdirTemp("", "git-askpass-*")
	if err != nil {
		return nil, nil, err
	}

	askpassPath := filepath.Join(dir, "askpass.sh")
	script := `#!/bin/sh
case "$1" in
	*Username*) echo "x-access-token" ;;
	*Password*) echo "$COURSE_CI_TOKEN" ;;
	*) echo "" ;;
esac
`

	if err := os.WriteFile(askpassPath, []byte(script), 0o700); err != nil {
		os.RemoveAll(dir)
		return nil, nil, err
	}

	env := append(os.Environ(),
		"GIT_ASKPASS="+askpassPath,
		"GIT_TERMINAL_PROMPT=0",
		courseCITokenEnv+"="+token,
	)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return env, cleanup, nil
}

func sshRepoURL(owner, repo string) string {
	return fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
}

func httpsRepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func fetchGitHubContentRaw(ctx context.Context, spec repoSpec, pathInRepo, token string) ([]byte, error) {
	owner := url.PathEscape(spec.Owner)
	repo := url.PathEscape(spec.Repo)
	esc := escapePathSegments(pathInRepo)

	base := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, esc)

	u, _ := url.Parse(base)
	q := u.Query()

	if spec.Ref != "" {
		q.Set("ref", spec.Ref)
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("User-Agent", "course-tests-fetcher/4.0")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository path not found: %s", pathInRepo)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("GitHub API %s: %s", resp.Status, strings.TrimSpace(string(slurp)))
	}

	return io.ReadAll(resp.Body)
}

func safeJoin(root, p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", errors.New("absolute paths is restricted")
	}

	clean := filepath.Clean(p)
	dst := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, dst)
	if err != nil {
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("the path goes beyond the repository")
	}

	return dst, nil
}

func relOrSame(base, p string) string {
	if r, err := filepath.Rel(base, p); err == nil {
		return r
	}
	return p
}
