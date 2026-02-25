package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nhosoya/claude-code-share/internal/server"
)

func main() {
	port := flag.Int("port", 3333, "HTTP server port")
	host := flag.String("host", "0.0.0.0", "HTTP server host")
	logDir := flag.String("log-dir", defaultLogDir(), "Path to Claude Code projects directory")
	flag.Parse()

	srv := server.New(*logDir)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	printStartupInfo(addr, *port, *logDir)

	slog.Info("starting server", "addr", addr, "log-dir", *logDir)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func defaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".claude", "projects")
	}
	return filepath.Join(home, ".claude", "projects")
}

func printStartupInfo(addr string, port int, logDir string) {
	fmt.Printf("claude-code-share\n")
	fmt.Printf("  Log directory: %s\n", logDir)
	fmt.Printf("  Local:         http://localhost:%d\n", port)

	addrs := lanAddresses()
	for _, a := range addrs {
		fmt.Printf("  Network:       http://%s:%d\n", a, port)
	}
	fmt.Println()
}

func lanAddresses() []string {
	var addrs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range ifAddrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}
			addrs = append(addrs, ip.String())
		}
	}
	return addrs
}
