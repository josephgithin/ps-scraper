package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	scraper "alienz.in/ps-scraper/scraper"
	"github.com/akamensky/argparse"
)

func main() {

	parser := argparse.NewParser("ps-scrapper", "scraps pluralsight courses")
	user := parser.String("u", "username", &argparse.Options{Required: true, Help: "pluralsight username"})
	password := parser.String("p", "password", &argparse.Options{Required: true, Help: "pluralsight password"})
	courses := parser.List("c", "course", &argparse.Options{Required: true, Help: "course names to scrap"})
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	log.Println(*user)
	log.Println(*password)
	log.Println(*courses)

	for i, course := range *courses {
		log.Printf("Scrapping course %v : %v\n", i, course)
		s1 := strings.Split(course, "-")
		firstModule := s1[len(s1)-1]
		s1 = s1[:len(s1)-1]
		course := strings.Join(s1[:], "-")
		scraper.PsScrap(*user, *password, course, firstModule)
	}

}
