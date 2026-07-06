package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	coreUtils "github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/web"
)

func main() {
	addr := flag.String("addr", getEnvOrDefault("PORT", "8080"), "listen address for the web server")
	flag.Parse()
	listenAddr := normalizeAddr(*addr)
	fmt.Println("Yatori web server starting on", listenAddr)
	coreUtils.YatoriCoreInit()
	server := web.NewServer()
	log.Fatal(server.Run(listenAddr))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ":8080"
	}
	if strings.HasPrefix(addr, ":") {
		return addr
	}
	if strings.Contains(addr, ":") {
		return addr
	}
	return ":" + addr
}
