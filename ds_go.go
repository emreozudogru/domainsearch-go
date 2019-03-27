package main

import (
	"bufio"
	"fmt"
	"github.com/haccer/available"
	"log"
	"os"
)

func main() {
	fmt.Println("Türkçe sözlükteki kelimelerin, com ve net uzantılı alan adlarının boşta olup olmadığı kontrol ediliyor...")
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
		//		for i := 0; i < len(ext); i++ {
		for _, tld := range ext {
			domain := word + tld
			available := available.Domain(domain)
			fmt.Println(domain + " looking...")
			if available {
				fmt.Println(domain + " alan adı şu an için boşta.")
			}
		}
	}
}
