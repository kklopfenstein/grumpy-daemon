package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
)

func RandomJoke() string {
	file, err := os.Open("data/jokes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var jokes []string

	for scanner.Scan() {
		jokes = append(jokes, scanner.Text())
	}

	joke := jokes[rand.Intn(len(jokes))]
	joke = strings.Replace(joke, "<>", "\n", -1)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return joke
}
