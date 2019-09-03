package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"log"
	"os"
	"regexp"
)

type tarCompressionType int

const (
	UncompressedType tarCompressionType = iota
	GZType
	XZType
	BZType
)

// List the contents of an archive file, returning a slice of strings containing the filenames.
//
// It attempts to work out the format based on file extension, and currently implements a few common types of tar file and also zip files.
func GetArchiveListing(filename string) []string {
	tarFormatRegex := regexp.MustCompile("\\.tar$")
	tarGZFormatRegex := regexp.MustCompile("(?:\\.tar\\.gz|\\.tgz)$")
	tarXZFormatRegex := regexp.MustCompile("(?:\\.tar\\.xz|\\.txz)$") // I've never seen a file with a txz extension but apparently they exist
	tarBZFormatRegex := regexp.MustCompile("(?:\\.tar\\.bz2|\\.tbz)$")

	zipFormatRegex := regexp.MustCompile("\\.zip$")

	if tarFormatRegex.MatchString(filename) {
		return getTarFileListing(filename, UncompressedType)
	}
	if tarGZFormatRegex.MatchString(filename) {
		return getTarFileListing(filename, GZType)
	}
	if tarXZFormatRegex.MatchString(filename) {
		return getTarFileListing(filename, XZType)
	}
	if tarBZFormatRegex.MatchString(filename) {
		return getTarFileListing(filename, BZType)
	}
	if zipFormatRegex.MatchString(filename) {
		return getZipListing(filename)
	}
	log.Fatal("Error: archive format not recognised.")
	return nil
}

func getTarReaderFromFile(file *os.File, tarType tarCompressionType) *tar.Reader {
	// We basically need to stack a bunch of reader interfaces to get our tar data, but which types
	//  depends on what compression
	var tarReader *tar.Reader
	switch tarType {
	case UncompressedType:
		tarReader = tar.NewReader(file)
	case XZType:
		log.Fatal("Error: XZ type not yet implemented")
		// Maybe use: https://godoc.org/github.com/xi2/xz
	case GZType:
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal("Error: could not open file " + file.Name() + " as gzip file: " + err.Error())
		}
		tarReader = tar.NewReader(gzReader)
	case BZType:
		bz2Reader := bzip2.NewReader(file)
		tarReader = tar.NewReader(bz2Reader)
	default:
		log.Fatal("Error: invalid tar compression type specified")
	}
	return tarReader
}

func getTarFileListing(filename string, tarType tarCompressionType) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error: could not open file " + filename + ": " + err.Error())
	}
	defer file.Close()

	tarReader := getTarReaderFromFile(file, tarType)

	var listing []string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		listing = append(listing, header.Name)
	}

	return listing
}

func getZipListing(filename string) []string {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer zipReader.Close()

	var listing []string
	for _, archivedFilename := range zipReader.File {
		listing = append(listing, archivedFilename.Name)
	}
	return listing
}

// Types of file we have to handle:
// tar.gz / tgz
// tar.xz
// tar.bz2 / tbz
// zip
//
// Maybe handle later:
// tar.Z
// 7z

// Don't handle?
// rpm
// deb

// This is because if we have an RPM, we don't need to get a listing from it: we just need to unpack it, and unless it's a source rpm (wtf) we won't need to build anything.
// Getting a listing of the file serves to let us identify the build system, and we don't need to do that for these.

// We could also add directories as a case here to let us use e.g. cloned git repositories
