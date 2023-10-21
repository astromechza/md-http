package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/russross/blackfriday"
)

const (
	DefaultListenAddr  = "0.0.0.0:8080"
	DefaultPageTitle   = "Landing page"
	DefaultCssUrl      = ""
	DefaultDebug       = false
	DefaultUsagePrefix = "Usage: md-http [options...] <filepath>\n"
	DefaultUsageSuffix = "\nAll options also have an environment variable counterpart: MDHTTP_<option>=<value>\n"
)

// main is the entrypoint from the command line to capture the args it is running with
func main() {
	if err := mainInner(os.Args, os.Stdout); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Exiting with error: " + err.Error())
		os.Exit(1)
	}
}

// mainInner is the real interface entrypoint, but testable.
// This defines the flags, validation, and parsing options.
func mainInner(args []string, output io.Writer) error {
	fs := flag.NewFlagSet(filepath.Base(args[0]), flag.ContinueOnError)

	fs.SetOutput(output)
	var listenAddr string
	fs.StringVar(&listenAddr, "listen", DefaultListenAddr, "The socket address to listen on")
	var pageTitle string
	fs.StringVar(&pageTitle, "title", DefaultPageTitle, "The HTML title of the page")
	var cssUrl string
	fs.StringVar(&cssUrl, "css", DefaultCssUrl, "An optional css file path or url (http:// or https://) to serve in the output")
	var debugLevel bool
	fs.BoolVar(&debugLevel, "debug", DefaultDebug, "Enable debug logging")

	fs.Usage = func() {
		_, _ = fs.Output().Write([]byte(DefaultUsagePrefix))
		fs.PrintDefaults()
		_, _ = fs.Output().Write([]byte(DefaultUsageSuffix))
	}
	var err error
	fs.VisitAll(func(f *flag.Flag) {
		if value, ok := os.LookupEnv(f.Name); ok {
			if fErr := fs.Set(f.Name, value); fErr != nil {
				err = fErr
			}
		}
	})
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 1 {
		_, _ = fs.Output().Write([]byte("Expected a single argument as the markdown filepath!\n\n"))
		fs.Usage()
		return http.ErrServerClosed
	}

	addrPort, err := netip.ParseAddrPort(listenAddr)
	if err != nil {
		_, _ = fmt.Fprintf(fs.Output(), "Invalid value for 'listen' '%s'\n\n", listenAddr)
		fs.Usage()
		return http.ErrServerClosed
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		AddSource: debugLevel,
		Level:     map[bool]slog.Level{false: slog.LevelInfo, true: slog.LevelDebug}[debugLevel],
	})))

	return run(addrPort, fs.Arg(0), cssUrl, pageTitle)
}

// run does the real logic of reading the file and running the server
func run(listenAddr netip.AddrPort, markdownFile string, cssUrl string, pageTitle string) error {
	slog.Debug("reading markdown file", "path", markdownFile)
	raw, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("failed to open the file: %w", err)
	}

	if cssUrl != "" && !strings.HasPrefix(cssUrl, "http://") && !strings.HasPrefix(cssUrl, "https://") {
		cssUrl = strings.TrimPrefix(cssUrl, "file://")
		slog.Debug("reading css file", "path", cssUrl)
		rawCss, err := os.ReadFile(cssUrl)
		if err != nil {
			return fmt.Errorf("failed to read the css file: %v", err)
		}
		http.HandleFunc("/default.css", func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/css")
			_, _ = writer.Write(rawCss)
		})
		cssUrl = "./default.css"
	}

	slog.Debug("converting markdown to html")
	htmlContent := blackfriday.Markdown(
		raw,
		blackfriday.HtmlRenderer(
			// common defaults
			blackfriday.HTML_USE_XHTML|
				blackfriday.HTML_USE_SMARTYPANTS|
				blackfriday.HTML_SMARTYPANTS_FRACTIONS|
				blackfriday.HTML_SMARTYPANTS_DASHES|
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES|
				// extras
				blackfriday.HTML_COMPLETE_PAGE|
				blackfriday.HTML_FOOTNOTE_RETURN_LINKS|
				blackfriday.HTML_HREF_TARGET_BLANK,
			pageTitle,
			cssUrl,
		),
		// defaults
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS|
			blackfriday.EXTENSION_TABLES|
			blackfriday.EXTENSION_FENCED_CODE|
			blackfriday.EXTENSION_AUTOLINK|
			blackfriday.EXTENSION_STRIKETHROUGH|
			blackfriday.EXTENSION_SPACE_HEADERS|
			blackfriday.EXTENSION_HEADER_IDS|
			blackfriday.EXTENSION_BACKSLASH_LINE_BREAK|
			blackfriday.EXTENSION_DEFINITION_LISTS|
			// extras
			blackfriday.EXTENSION_FOOTNOTES|
			blackfriday.EXTENSION_AUTO_HEADER_IDS,
	)

	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		_, _ = writer.Write([]byte("healthz check passed"))
	})

	http.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write(htmlContent)
	})

	server := &http.Server{Addr: listenAddr.String(), Handler: http.DefaultServeMux}
	go func() {
		exit := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

		// Receive output from signalChan.
		sig := <-exit
		slog.Info("Signal caught, stopping http server", "signal", sig.String())
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("Failure during shutdown", "err", err)
		}
	}()

	slog.Info("Starting http server", "listen", "http://"+listenAddr.String())
	return server.ListenAndServe()
}
