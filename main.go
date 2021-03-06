/*-
 * Copyright (c) 2017, 2018
 *	Benny Siegert <bsiegert@gmail.com>
 *
 * Provided that these terms and disclaimer and all copyright notices
 * are retained or reproduced in an accompanying document, permission
 * is granted to deal in this work without restriction, including un-
 * limited rights to use, publicly perform, distribute, sell, modify,
 * merge, give away, or sublicence.
 *
 * This work is provided "AS IS" and WITHOUT WARRANTY of any kind, to
 * the utmost extent permitted by applicable law, neither express nor
 * implied; without malicious intent or gross negligence. In no event
 * may a licensor, author or contributor be held liable for indirect,
 * direct, other damage, loss, or other issues arising in any way out
 * of dealing in the work, even if advised of the possibility of such
 * damage or existence of a defect, except proven that it results out
 * of said person's immediate fault when using the work as intended.
 */

// gourl2pkg is a tool to add or update Go packages in pkgsrc.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/bsiegert/gourl2pkg/dep"
	"github.com/bsiegert/gourl2pkg/pkgsrc"
)

var InfoLog = log.New(ioutil.Discard, "", log.LstdFlags)

var (
	index     = flag.Bool("index", false, "Print the reverse index of Go packages instead of adding any ports.")
	verbose   = flag.Bool("v", true, "Print verbose messages about what is happening.")
	pkgsrcdir = flag.String("pkgsrc", "", "Path to the top-level pkgsrc directory, will be taken from the PKGSRCDIR environment variable if not given.")
	local     = flag.Bool("local", false, "Use local GOPATH (mainly useful for testing)")
	force     = flag.Bool("force", false, "If true, do not skip packages on the command line that are already in pkgsrc.")
)

func init() {
	if *verbose {
		InfoLog = log.New(os.Stderr, "", log.LstdFlags)
	}
	if *pkgsrcdir == "" {
		*pkgsrcdir = os.Getenv("PKGSRC")
	}
	if *pkgsrcdir == "" {
		*pkgsrcdir = "/usr/pkgsrc"
	}
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 && !*index {
		log.Fatal("Need at least one argument")
	}

	var err error
	if _, err = os.Stat(*pkgsrcdir); err != nil {
		log.Fatalf("There is a problem with the pkgsrc directory (%v).\n"+
			"Please set the '-pkgsrc' option.", err)
	}
	revIndex, err = FullScan(*pkgsrcdir)
	if err != nil {
		log.Fatal(err)
	}
	if *index {
		revIndex.WriteTo(os.Stdout)
		return
	}

	var (
		ToPackage []string
		tmpdir    string
	)
	if *local {
		ToPackage = flag.Args()
	} else {
		tmpdir, err = ioutil.TempDir("", "gourl2pkg")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(tmpdir)
		InfoLog.Printf("Initial code download (%s)", tmpdir)
		ToPackage, err = GoGetResolve(flag.Args(), tmpdir)
		if err != nil {
			log.Fatal(err)
		}
	}

	for len(ToPackage) > 0 {
		InfoLog.Printf("Remaining to package: %s", ToPackage)
		l := len(ToPackage) - 1
		p := ToPackage[l]
		ToPackage = ToPackage[:l]
		if err := HandleURL(tmpdir, p); err != nil {
			log.Fatal(err)
		}
	}
}

func HandleURL(basedir string, srcpath string) error {
	if pkg, ok := revIndex.PrefixMatch(srcpath); ok {
		log.Printf("%s is already part of a pkgsrc package (%s)", srcpath, pkg)
		// TODO(bsiegert) some kind of upgrade flow
		if !*force {
			return nil
		}
	}

	meta := pkgsrc.PkgMeta{GoSrcpath: srcpath}
	dep.FindImportsRecursive(basedir, revIndex, &meta)
	fmt.Println(meta) // Placeholder
	return nil
}
