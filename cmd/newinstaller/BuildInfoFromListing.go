package main

import (
	"regexp"
)

// A design note:
//  we *could* have made some kind of "getFileType" function
//  *but* that means either making all the regexes global
//  or recreating them every function call (or, at least,
//  pushing them back onto the stack)
//
//  it still might be the thing to do, but, it feels wrong
//
//  overall I don't feel great about the design of this
//  functionality, but I'm not sure I can think of a better
//  way to build it

const (
	cmakelistType int = iota
	configureType
	confACType
	automakefileType
	cfileType
	cppfileType
	fortranfileType
	pythonfileType
	pysetupType
)

type FileOfInterest struct {
	sourceFilename string // Which archive the file comes from
	internalPath   string // The path of the file inside the archive
	fileType       int    // What kind of file the target file is
}

func GetFoIsFromListing(listing []string, sourceFilename string) []FileOfInterest {
	var regexes []*regexp.Regexp
	regexes = []*regexp.Regexp{
		cmakelistType:    regexp.MustCompile("(?:^|/)CMakeLists\\.txt$"),
		configureType:    regexp.MustCompile("(?:^|/)configure$"),
		confACType:       regexp.MustCompile("(?:^|/)configure\\.ac$"),
		automakefileType: regexp.MustCompile("(?:^|/)[Mm]akefile\\.am$"),
		cfileType:        regexp.MustCompile("(?:^|/)[^/]+\\.[ch]$"),
		cppfileType:      regexp.MustCompile("(?:^|/)[^/]+\\.[ch]pp$"),
		fortranfileType:  regexp.MustCompile("(?:^|/)[^/]+\\.[fF](90)?$"),
		// These can ping on the same file
		//  so put the more specific one first.
		pysetupType:    regexp.MustCompile("(?:^|/)setup\\.py?$"),
		pythonfileType: regexp.MustCompile("(?:^|/)[^/]+\\.py$"),
	}

	var FoIs []FileOfInterest

	for _, filename := range listing {
		// Check against all the regexps -- this... could take a file
		for fileType, one_regex := range regexes {
			if one_regex.MatchString(filename) {
				FoIs = append(
					FoIs,
					FileOfInterest{sourceFilename: sourceFilename, internalPath: filename, fileType: fileType})
				break
			}
		}
	}

	return FoIs

}
