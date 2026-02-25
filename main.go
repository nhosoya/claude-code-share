package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

type lanAddr struct {
	IP    string
	Iface string
}

func printStartupInfo(addr string, port int, logDir string) {
	fmt.Printf("claude-code-share\n")
	fmt.Printf("  Log directory: %s\n", logDir)
	fmt.Printf("  Local:         http://localhost:%d\n", port)

	addrs := lanAddresses()
	for _, a := range addrs {
		fmt.Printf("  Network:       http://%s:%d (%s)\n", a.IP, port, a.Iface)
	}
	fmt.Println()
}

// isPhysicalInterface returns true for likely physical network interfaces
// (Wi-Fi, Ethernet) and false for virtual ones (VPN, Docker, etc.).
func isPhysicalInterface(name string) bool {
	// macOS: en0 = Wi-Fi, en1-enN = Ethernet/Thunderbolt
	// Linux: eth0, wlan0, enpXsY, wlpXsY
	prefixes := []string{"en", "eth", "wlan", "enp", "wlp"}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func lanAddresses() []lanAddr {
	var addrs []lanAddr
	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if !isPhysicalInterface(iface.Name) {
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
			addrs = append(addrs, lanAddr{IP: ip.String(), Iface: iface.Name})
		}
	}
	return addrs
}
