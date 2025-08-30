package main

import (
	"flag"
	"log"

	"ticker-forge/internal/cli"
)

func main() {
	mode := flag.String("mode", "tui", "tui|serve")
	port := flag.String("port", "8080", "port to listen on")
	symbol := flag.String("symbol", "AAPL", "default ticker")
	rng := flag.String("range", "1d", "default range (1d,5d,1mo...)")
	interval := flag.String("interval", "1m", "default interval (1m,5m,15m...)")
	flag.Parse()

	opts := cli.Options{
		Port:            *port,
		DefaultSymbol:   *symbol,
		DefaultRange:    *rng,
		DefaultInterval: *interval,
	}
	switch *mode {
	case "serve":
		opts.Mode = cli.ModeServe
	default:
		opts.Mode = cli.ModeTUI
	}

	if err := cli.Run(opts); err != nil {
		log.Fatal(err)
	}
}
