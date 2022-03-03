package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

var configFile string
var onlyExpand bool
var showLists bool

func init() {
	flag.StringVar(&configFile, "config", "./update-acls.conf", "path to config file")
	flag.BoolVar(&onlyExpand, "no-targets", false, "skip the output phase, only expand lists")
	flag.BoolVar(&showLists, "show-lists", false, "print the built lists after expansion")
}

func main() {
	flag.Parse()

	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not parse config file: %w", err))
	}

	if len(config.Lists) == 0 {
		log.Fatalln("no list configurations were found in the config")
	}

	if showLists {
		log.Println("read list configs:")
		for _, v := range config.Lists {
			log.Printf("%s: %s\n", v.Name, v.Description)
		}
	}

	// Expansion step: expand lists of sources into a list of users
	lists, errs := config.ExpandAEULists()

	if len(errs) != 0 {
		for _, v := range errs {
			log.Println(v)
		}
		log.Fatalf("%d errors produced during list expansion, will not continue to output step\n", len(errs))
	}

	// Might make this a cli opt
	//configDump, err := yaml.Marshal(config)
	//fmt.Printf("%+v\n", string(configDump))

	if showLists {
		for _, v := range lists {
			log.Printf("expanded %s to: %s\n", v.Name, strings.Join(v.BuiltList, ","))
		}
	}

	if onlyExpand {
		return
	}

	// Output step: modify destinations with our lists of users
	for _, v := range lists {
		errs = WriteListDestinations(v)
	}

	if len(errs) != 0 {
		for _, v := range errs {
			log.Println(v)
		}
		log.Fatalf("%d errors produced during list writing\n", len(errs))
	}

	log.Println("update complete")
}
