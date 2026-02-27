package main

import (
    "fmt"
    "github.com/yourusername/event-analytics-service/internal/config"
)

func main() {
    cfg := config.LoadConfig()
    fmt.Printf("ClickHouseHost: %s\n", cfg.ClickHouseHost)
}
