package main

import (
	"bufio"
	"fmt"
	"github.com/haccer/available"
	"log"
	"os"
)

func main() {
	fmt.Println("Loading wordlist and looking for available domains")
	file, err := os.Open("wtzl.txt")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	file.Close()
	ext := []string{".us"}
	for _, word := range txtlines {
		for _, tld := range ext {
			domain := word + tld
			available := available.Domain(domain)
			fmt.Println(domain + " looking...")
			if available {
				fmt.Println(domain + " is EMPTY")
			}
		}
	}
}
