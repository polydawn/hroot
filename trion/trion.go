package main

func main() {
	config := FindConfig(".")
	prepRun(config)()
}
