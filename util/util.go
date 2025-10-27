package util

import (
	"bufio"
	"fmt"
	datafinder "grepkvorum/util/dataFinder"
	datareader "grepkvorum/util/dataReader"
	"grepkvorum/util/entities"
	"io"
	"os"
)

type Args struct {
	CountStrings         bool
	IgnoreRegistr        bool
	UseStringCompare     bool
	InvertMatch          bool
	NumberForStringsFlag bool
	MultiServer          bool
	IgnoreVector         bool
	AfterMatch           int
	BeforeMatch          int
	AroundMatch          int
	InStream             io.Reader
}

func initialize(a Args, pattern string) *bufio.Scanner {
	// parse out pattern
	if len(os.Args) < 2 {
		fmt.Println("pattern for find doesn't set")
		os.Exit(1)
	}
	if rune(pattern[0]) == '-' {
		fmt.Println("flag can't be before pattern")
		os.Exit(1)
	}

	// get new scanner
	sc, err := datareader.GetScanner(a.InStream)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return sc
}

func Do(a Args, print bool, pattern string) (*entities.ResultVector, error) {
	sc := initialize(a, pattern)

	f := &datafinder.AppFlags{
		AfterMatch:       a.AfterMatch,
		BeforeMatch:      a.BeforeMatch,
		CountStrings:     a.CountStrings,
		IgnoreRegistr:    a.IgnoreRegistr,
		UseStringCompare: a.UseStringCompare,
		InvertMatch:      a.InvertMatch,
		IgnoreVector:     !a.MultiServer && a.IgnoreVector,
		NumeredOutput:    a.NumberForStringsFlag,
	}

	// if we have -C, we will just put his value in AfterMatch(-A) and BeforeMathc(-B)
	if a.AroundMatch != 0 {
		f.AfterMatch = a.AroundMatch
		f.BeforeMatch = a.AroundMatch
	}

	// if we only count result strings, we will not find after lines and before lines
	if a.CountStrings {
		f.AfterMatch = 0
		f.BeforeMatch = 0
	}
	r, err := datafinder.Find(pattern, f, sc)
	if err != nil {
		return nil, err
	}

	if print {
		if a.CountStrings {
			fmt.Println(r.GetCounter())
		} else {
			r.PrettyPrint(a.NumberForStringsFlag)
		}
		return nil, nil
	}

	return r, nil
}
