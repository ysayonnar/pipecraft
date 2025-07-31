package main

import "pipecraft/internal/config"

func main() {
	cfg := config.MustParse()

	_ = cfg
}
