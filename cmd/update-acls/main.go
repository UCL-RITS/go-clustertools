package main

import (
	"flag"
	"fmt"
	"log"
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
			fmt.Printf("%s: %+v\n", v.Name, v.BuiltList)
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
