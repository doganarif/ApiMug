package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/doganarif/ApiMug/internal/api"
	"github.com/doganarif/ApiMug/internal/server"
	"github.com/doganarif/ApiMug/internal/tui"
	"github.com/doganarif/ApiMug/pkg/spec"
	"github.com/spf13/cobra"
)

var (
	port    int
	baseURL string
	rootCmd = &cobra.Command{
		Use:   "apimug [spec-file-or-url]",
		Short: "ApiMug - Beautiful OpenAPI/Swagger viewer and server",
		Long:  `ApiMug is a CLI tool to browse and serve OpenAPI/Swagger specifications with a beautiful TUI interface.`,
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}
)

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the Swagger UI server on")
	rootCmd.Flags().StringVarP(&baseURL, "base-url", "b", "", "Base URL for API requests (default: from spec)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	source := args[0]
	ctx := context.Background()

	loader := spec.NewLoader()

	var doc *api.Spec

	if isURL(source) {
		fmt.Printf("Loading spec from URL: %s\n", source)
		d, err := loader.LoadFromURL(ctx, source)
		if err != nil {
			return fmt.Errorf("failed to load spec from URL: %w", err)
		}
		doc = &api.Spec{Doc: d, Source: source}
	} else {
		fmt.Printf("Loading spec from file: %s\n", source)
		d, err := loader.LoadFromFile(ctx, source)
		if err != nil {
			return fmt.Errorf("failed to load spec from file: %w", err)
		}
		doc = &api.Spec{Doc: d, Source: source}
	}

	title, version, _ := doc.GetInfo()
	fmt.Printf("Loaded: %s (v%s)\n", title, version)
	fmt.Printf("Endpoints: %d\n\n", len(doc.GetEndpoints()))

	if baseURL == "" && doc.Doc != nil && len(doc.Doc.Servers) > 0 {
		baseURL = doc.Doc.Servers[0].URL
		fmt.Printf("Using base URL from spec: %s\n", baseURL)
	} else if baseURL != "" {
		fmt.Printf("Using base URL: %s\n", baseURL)
	} else {
		fmt.Println("Warning: No base URL configured. You won't be able to send requests.")
	}

	var srv *server.Server
	restartCh := make(chan struct{}, 1)
	currentPort := port

	srv = server.New(doc, currentPort)
	go func() {
		for {
			fmt.Printf("Starting Swagger UI server on http://localhost%s\n", srv.Addr())
			if err := srv.Start(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			}
			<-restartCh
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if srv != nil {
			srv.Shutdown(ctx)
		}
		os.Exit(0)
	}()

	time.Sleep(500 * time.Millisecond)

	onSettingsChange := func(newBaseURL string, newPort int) {
		if newPort != currentPort {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Shutdown(ctx)

			currentPort = newPort
			srv = server.New(doc, currentPort)
			restartCh <- struct{}{}

			time.Sleep(500 * time.Millisecond)
			fmt.Printf("\nServer restarted on port %d\n", currentPort)
		}
		if newBaseURL != baseURL {
			baseURL = newBaseURL
			fmt.Printf("\nBase URL updated to: %s\n", baseURL)
		}
	}

	p := tea.NewProgram(tui.NewModel(doc, baseURL, currentPort, onSettingsChange), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

func isURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}
