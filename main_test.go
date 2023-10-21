package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func freePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func TestRunNominal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := freePort()
	require.NoError(t, err)
	addrPort, err := netip.ParseAddrPort(fmt.Sprintf("127.0.0.1:%d", port))
	require.NoError(t, err)

	mdPath := filepath.Join(t.TempDir(), "example.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# example header\n\n"), 0400))

	cssPath := filepath.Join(t.TempDir(), "some.css")
	require.NoError(t, os.WriteFile(cssPath, []byte("body { color: red; }"), 0400))

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		assert.EqualError(t, run(ctx, argsStruct{AddrPort: addrPort, MarkdownFile: mdPath, CssUrl: cssPath, PageTitle: "some title"}), http.ErrServerClosed.Error())
	}()

	for {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err != nil {
			time.Sleep(time.Second)
		} else {
			defer resp.Body.Close()
			break
		}
	}

	t.Run("test main", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

		data, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(data), `<!DOCTYPE html PUBLIC`)
		assert.Contains(t, string(data), `<title>some title</title>`)
		assert.Contains(t, string(data), `<h1 id="example-header">example header</h1>`)
		assert.Contains(t, string(data), `<link rel="stylesheet" type="text/css" href="default.css" />`)
	})

	t.Run("test healthz", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

		data, _ := io.ReadAll(resp.Body)
		assert.Equal(t, `healthz check passed`, string(data))
	})

	t.Run("test css", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/default.css", port))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/css; charset=utf-8", resp.Header.Get("Content-Type"))

		data, _ := io.ReadAll(resp.Body)
		assert.Equal(t, `body { color: red; }`, string(data))
	})

	cancel()
	wg.Done()
}

func TestParse_minimum(t *testing.T) {
	mdPath := filepath.Join(t.TempDir(), "example.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# example header\n\n"), 0400))

	buff := new(bytes.Buffer)
	args, err := parse([]string{"binary", mdPath}, buff)
	assert.NoError(t, err)
	assert.Equal(t, argsStruct{
		PageTitle:    "Landing page",
		MarkdownFile: mdPath,
		AddrPort:     netip.AddrPortFrom(netip.AddrFrom4([4]byte{0, 0, 0, 0}), 8080),
	}, args)
}

func TestParse_all(t *testing.T) {
	mdPath := filepath.Join(t.TempDir(), "example.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# example header\n\n"), 0400))

	cssPath := filepath.Join(t.TempDir(), "example.css")
	require.NoError(t, os.WriteFile(cssPath, []byte(""), 0400))

	buff := new(bytes.Buffer)
	args, err := parse([]string{"binary", "-css", cssPath, "-debug", "-title", "Thing", "-listen", "127.0.0.1:8090", "-jsonlog", mdPath}, buff)
	assert.NoError(t, err)
	assert.Equal(t, argsStruct{
		PageTitle:    "Thing",
		MarkdownFile: mdPath,
		CssUrl:       cssPath,
		AddrPort:     netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), 8090),
		LogDebug:     true,
		LogJson:      true,
	}, args)
}

func TestParse_env(t *testing.T) {
	mdPath := filepath.Join(t.TempDir(), "example.md")
	require.NoError(t, os.WriteFile(mdPath, []byte("# example header\n\n"), 0400))

	cssPath := filepath.Join(t.TempDir(), "example.css")
	require.NoError(t, os.WriteFile(cssPath, []byte(""), 0400))

	defer os.Clearenv()
	require.NoError(t, os.Setenv("MDHTTP_css", cssPath))
	require.NoError(t, os.Setenv("MDHTTP_title", "Thing"))
	require.NoError(t, os.Setenv("MDHTTP_listen", "127.0.0.1:8090"))
	require.NoError(t, os.Setenv("MDHTTP_debug", "true"))
	require.NoError(t, os.Setenv("MDHTTP_jsonlog", "true"))

	buff := new(bytes.Buffer)
	args, err := parse([]string{"binary", mdPath}, buff)
	assert.NoError(t, err)
	assert.Equal(t, argsStruct{
		PageTitle:    "Thing",
		MarkdownFile: mdPath,
		CssUrl:       cssPath,
		AddrPort:     netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), 8090),
		LogDebug:     true,
		LogJson:      true,
	}, args)
}
