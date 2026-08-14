package download

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shafqat-a/iplegence/internal/config"
)

type Result struct {
	Source config.SourceSpec
	Path   string
	Err    error
}

var azureServiceTagsRe = regexp.MustCompile(`https://download\.microsoft\.com/download/[^"' \s]+ServiceTags_Public_[^"' \s]+\.json`)

func Fetch(ctx context.Context, spec config.SourceSpec) error {
	if !spec.Enabled {
		return nil
	}
	if missing := spec.MissingEnv(); len(missing) > 0 {
		if spec.Required {
			return fmt.Errorf("%s missing required env: %s", spec.ID, strings.Join(missing, ", "))
		}
		return nil
	}
	url := spec.ExpandURL()
	timeout := 300 * time.Second
	if dl := deadlineFrom(ctx); dl > 0 {
		timeout = dl
	}
	if spec.Kind == "azure_servicetags" {
		inner, err := discoverAzureURL(ctx, url, timeout)
		if err != nil {
			if !spec.Required {
				return nil
			}
			return err
		}
		url = inner
	}
	if err := os.MkdirAll(filepath.Dir(spec.Path), 0o755); err != nil {
		return err
	}
	if err := getToFile(ctx, url, spec.Path, timeout); err != nil {
		if strings.Contains(spec.URL, "${YYYYMM}") {
			prev := spec.ExpandURLAt(time.Now().UTC().AddDate(0, -1, 0))
			if prev != url {
				if err2 := getToFile(ctx, prev, spec.Path, timeout); err2 == nil {
					err = nil
				} else {
					err = fmt.Errorf("%v; previous month: %w", err, err2)
				}
			}
		}
		if err != nil {
			return err
		}
	}
	if spec.Kind == "maxmind_tar_gz" {
		if spec.ExtractGlob == "" || spec.ExtractTo == "" {
			return fmt.Errorf("%s: extract_glob and extract_to required", spec.ID)
		}
		if err := extractFirstMatch(spec.Path, spec.ExtractGlob, spec.ExtractTo); err != nil {
			return err
		}
	}
	if spec.Kind == "gzip" {
		if spec.ExtractTo == "" {
			return fmt.Errorf("%s: extract_to required", spec.ID)
		}
		if err := gunzipFile(spec.Path, spec.ExtractTo); err != nil {
			return err
		}
	}
	return nil
}

func FetchAll(ctx context.Context, cfg config.File) []Result {
	n := cfg.DownloadConcurrency
	if n < 1 {
		n = 1
	}
	sem := make(chan struct{}, n)
	out := make([]Result, len(cfg.Sources))
	var wg sync.WaitGroup
	for i, spec := range cfg.Sources {
		if !spec.Enabled {
			out[i] = Result{Source: spec}
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, spec config.SourceSpec) {
			defer wg.Done()
			defer func() { <-sem }()
			srcCtx := ctx
			if cfg.DownloadTimeoutSeconds > 0 {
				var cancel context.CancelFunc
				srcCtx, cancel = context.WithTimeout(ctx, time.Duration(cfg.DownloadTimeoutSeconds)*time.Second)
				defer cancel()
			}
			err := Fetch(srcCtx, spec)
			path := spec.LoadPath()
			if err == nil && missingOptional(spec) {
				path = ""
			}
			out[i] = Result{Source: spec, Path: path, Err: err}
		}(i, spec)
	}
	wg.Wait()
	return out
}

func RequiredError(results []Result) error {
	var errs []error
	for _, r := range results {
		if r.Err != nil && r.Source.Required {
			errs = append(errs, fmt.Errorf("%s: %w", r.Source.ID, r.Err))
		}
	}
	return errors.Join(errs...)
}

func missingOptional(spec config.SourceSpec) bool {
	return !spec.Required && len(spec.MissingEnv()) > 0
}

func deadlineFrom(ctx context.Context) time.Duration {
	if dl, ok := ctx.Deadline(); ok {
		return time.Until(dl)
	}
	return 0
}

func getToFile(ctx context.Context, url, dest string, timeout time.Duration) error {
	tmp := dest + ".tmp"
	defer os.Remove(tmp)
	var last error
	backoff := time.Second
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}
		err := oneGet(ctx, url, tmp, timeout)
		if err == nil {
			info, statErr := os.Stat(tmp)
			if statErr != nil {
				return statErr
			}
			if info.Size() == 0 {
				return fmt.Errorf("downloaded empty file from %s", url)
			}
			if err := os.Rename(tmp, dest); err != nil {
				return err
			}
			return nil
		}
		last = err
		if !retryable(err) {
			return err
		}
	}
	return last
}

type statusError struct {
	code int
	msg  string
}

func (e statusError) Error() string { return e.msg }

func retryable(err error) bool {
	var se statusError
	if errors.As(err, &se) {
		return se.code == 429 || se.code >= 500
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return err != nil
}

func oneGet(ctx context.Context, url, dest string, timeout time.Duration) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		io.Copy(io.Discard, resp.Body)
		return statusError{code: resp.StatusCode, msg: fmt.Sprintf("GET %s: %s", url, resp.Status)}
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, resp.Body)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func discoverAzureURL(ctx context.Context, pageURL string, timeout time.Duration) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", statusError{code: resp.StatusCode, msg: fmt.Sprintf("GET %s: %s", pageURL, resp.Status)}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", err
	}
	m := azureServiceTagsRe.Find(body)
	if m == nil {
		return "", fmt.Errorf("azure service tags download URL not found")
	}
	return string(m), nil
}

func extractFirstMatch(archive, glob, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		ok, err := filepath.Match(filepath.ToSlash(glob), filepath.ToSlash(hdr.Name))
		if err != nil {
			return err
		}
		if !ok {
			// also match basenames against **/Name
			base := filepath.Base(hdr.Name)
			if gok, _ := filepath.Match(filepath.Base(glob), base); !gok {
				continue
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		tmp := dest + ".tmp"
		out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			os.Remove(tmp)
			return copyErr
		}
		if closeErr != nil {
			os.Remove(tmp)
			return closeErr
		}
		info, err := os.Stat(tmp)
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			os.Remove(tmp)
			return fmt.Errorf("extracted empty file from %s", archive)
		}
		return os.Rename(tmp, dest)
	}
	return fmt.Errorf("no file matching %q in %s", glob, archive)
}

func gunzipFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, gz)
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(tmp)
		return closeErr
	}
	info, err := os.Stat(tmp)
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		os.Remove(tmp)
		return fmt.Errorf("gunzip produced empty file from %s", src)
	}
	return os.Rename(tmp, dest)
}
