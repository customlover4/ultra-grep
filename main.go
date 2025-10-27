package main

import (
	"fmt"
	"grepkvorum/util"
	datareader "grepkvorum/util/dataReader"
	"io"
	"os"

	"github.com/spf13/pflag"
)

var (
	// Util flags
	filePathFlag         = pflag.StringP("file", "f", "-", "path to file with data")
	afterMatchFlag       = pflag.IntP("after", "A", 0, "after match strings to output")
	beforeMatchFlag      = pflag.IntP("before", "B", 0, "before match strings to output")
	aroundMatchFlag      = pflag.IntP("context", "C", 0, "around match strings to output")
	countStringsFlag     = pflag.BoolP("count", "c", false, "output only matches strings count")
	ignoreRegistrFlag    = pflag.BoolP("ignore", "i", false, "ignore regist for mathcing")
	fixStringCmpFlag     = pflag.BoolP("fixed", "F", false, "fixed substring compare for pattern")
	numberForStringsFlag = pflag.BoolP("number", "n", false, "output string number before value")
	invertMatchFlag      = pflag.BoolP("invert", "v", false, "output only not matches strings")

	// mode flags
	IgnoreVector = pflag.BoolP("vector", "V", false, "if ignoring vector, output may be strange (with duplicates or wrong sequence), but memory usage will be low")

	// Cli flags
	multiServerFlag       = pflag.BoolP("multi", "m", false, "multi-server mode")
	portFlag              = pflag.IntP("port", "p", 8080, "port value for coordinator")
	countServersFlag      = pflag.IntP("servers", "s", 3, "servers number")
	replicationFactorFlag = pflag.IntP("replications", "r", 3, "replications count for every shard")
)

func main() {
	pflag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "can't find pattern in arguments")
		os.Exit(1)
	}

	pattern := os.Args[1]

	var stream io.Reader
	if *filePathFlag == "-" {
		stream = os.Stdin
	} else {
		f, err := os.Open(*filePathFlag)
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = f.Close()
		}()

		stream = f
	}

	a := util.Args{
		InStream:             stream,
		AfterMatch:           *afterMatchFlag,
		BeforeMatch:          *beforeMatchFlag,
		AroundMatch:          *aroundMatchFlag,
		CountStrings:         *countStringsFlag,
		IgnoreRegistr:        *ignoreRegistrFlag,
		UseStringCompare:     *fixStringCmpFlag,
		NumberForStringsFlag: *numberForStringsFlag,
		InvertMatch:          *invertMatchFlag,
		MultiServer:          *multiServerFlag,
		IgnoreVector:         *IgnoreVector,
	}

	if !(*multiServerFlag) {
		_, err := util.Do(a, true, pattern)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		fpath := *filePathFlag
		if fpath == "-" {
			fpath = datareader.MoveStdinToFile()
		}

		shards, err := datareader.SplitFileByLines(fpath, *countServersFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		err = util.CoordinatorProcess(
			*countServersFlag, *replicationFactorFlag, *portFlag,
			fpath, pattern, a, shards,
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		if *filePathFlag == "-" {
			err = os.Remove(fpath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
	}
}
