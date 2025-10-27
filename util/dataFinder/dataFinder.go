package datafinder

import (
	"bufio"
	"errors"
	"fmt"
	"grepkvorum/util/circ"
	"grepkvorum/util/entities"
	"regexp"
	"strings"
)

var (
	ErrWrongRegExpPatter = errors.New("wrong reg exp pattern")
)

// flags
type AppFlags struct {
	AfterMatch       int
	BeforeMatch      int
	CountStrings     bool
	IgnoreRegistr    bool
	UseStringCompare bool
	InvertMatch      bool
	IgnoreVector     bool
	NumeredOutput    bool
}

func Find(pattern string, f *AppFlags, sc *bufio.Scanner) (*entities.ResultVector, error) {
	// custom array for data
	result := entities.NewResultVector()
	buf := circ.NewCirc(f.BeforeMatch)

	var matcher *regexp.Regexp
	var needAfter int
	var err error
	var needToLower bool

	// if we dont use fixed string compare, we will try to compile regexp always
	if !f.UseStringCompare {
		if f.IgnoreRegistr { // if ignore regist we will add this instruction for regexp
			pattern = "(?i)" + pattern
		}

		matcher, err = regexp.Compile(pattern)
		if err != nil {
			return result, fmt.Errorf("%w: %w", ErrWrongRegExpPatter, err)
		}
	}

	// if ignore registr and use fixed strings compare
	// we will make pattern and data line to lower
	if f.UseStringCompare && f.IgnoreRegistr {
		pattern = strings.ToLower(pattern)

		needToLower = true
	}

	i := 0
	for sc.Scan() {
		i++
		v := sc.Text()
		// cmp string and pattern
		var match bool
		if needToLower {
			match = matchString(strings.ToLower(v), pattern, matcher)
		} else {
			match = matchString(v, pattern, matcher)
		}

		added := false
		// only if substr(regexp) dont founded in line and Inverted result is on
		if !match && f.InvertMatch || match && !f.InvertMatch {
			added = true
		}

		if added {
			if f.BeforeMatch != 0 { // if we need lines before matchings (-B)
				j := i - f.BeforeMatch
				if j < 0 {
					j = 0
				}

				for _, v := range buf.Read() {
					result.Add(v.S, v.I, f.IgnoreVector, f.NumeredOutput)
				}
			}
		}

		if added {
			result.Add(v, i, f.IgnoreVector, f.NumeredOutput)
		}

		if !added {
			buf.Add(v, i)
			if needAfter != 0 {
				result.Add(v, i, f.IgnoreVector, f.NumeredOutput)
				needAfter--
			}
		} else {
			if f.CountStrings {
				result.AddCounter()
			}
			needAfter = f.AfterMatch
		}
	}

	return result, nil
}

func matchString(v string, pattern string, matcher *regexp.Regexp) bool {
	if matcher != nil && matcher.Match([]byte(v)) {
		return true
	} else if strings.Contains(v, pattern) {
		return true
	}

	return false
}
