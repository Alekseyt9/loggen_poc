package main

func main() {
	cfg := defaultConfig()
	g := Generate(cfg)
	printInfo(g)
}
