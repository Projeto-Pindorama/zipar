/*
 * zipar - tar(1)-esque program for Zip files.
 *
 * Copyright (C) 2025 Luiz Antônio Rangel (takusuman)
 *
 * SPDX-Licence-Identifier: MIT
 */

package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pindorama.net.br/libcmon/zhip"
	"rsc.io/getopt"
	"strconv"
	"strings"
)

var (
	fVerbose         bool
	fTableOfContents bool
	fExtract         bool
	destdir          string
	archive          string
	areader          *zip.ReadCloser
	largest_file     int
)

func main() {
	var err error

	/* Options. */
	flag.BoolVar(&fVerbose, "verbose", false,
		"Enable verbose output.")
	flag.BoolVar(&fTableOfContents, "toc", false,
		"List the contents of the zipfile.")
	flag.BoolVar(&fExtract, "extract", false,
		"The named files are extracted from the zipfile.")
	flag.StringVar(&destdir, "chdir", ".",
		"Use the next argument as the directory to place the files into.")
	flag.StringVar(&archive, "file", "",
		"Use the next argument as the name of the archive.")
	getopt.Aliases(
		"v", "verbose",
		"t", "toc",
		"x", "extract",
		"f", "file",
		"C", "chdir",
	)
	getopt.Parse()

	/*
	 * Extra arguments; possibly specific
	 * files to be extracted.
	 */
	extra := flag.Args()
	nextra := flag.NArg()
	areader, err = zip.OpenReader(archive)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"failed to open %s: %s\n",
			archive, err)
	}
	defer areader.Close()
	if fTableOfContents || fExtract {
		/*
		 * Obtain the largest file size integer
		 * length for the '-t' option formatting.
		 */
		if fTableOfContents {
			largest_file = len(strconv.FormatUint(
				uint64(zhip.GetZipLargestEntry(areader)), 10))
		}
	zipwalk:
		for ;; {
			file := zhip.GetZipEntries(areader)
			if file == nil {
				break
			}
			/* Check if the user specified files to be extracted. */
			for f := 0; nextra != 0 && f < nextra; f++ {
				if !strings.HasPrefix(file.Name, extra[f]) {
					continue zipwalk
				}
			}
			if fTableOfContents {
				print_entry_info(file)
			} else if fExtract {
				extract_entry(file)
			}
		}
	} else {
		os.Exit(1) // TODO: Usage()
	}
}

func print_entry_info(file *zip.FileHeader) {
	if fVerbose {
		fmt.Printf("%*d %s:%02.0f%% %10s %s ",
			largest_file,
			file.UncompressedSize,
			zhip.GetCompressionMethod(file),
			zhip.GetCompressionRatio(file),
			file.Mode().String(),
			file.Modified.Format("2006-01-02 15:04:05"),
		)
	}
	fmt.Println(file.Name)
}

func extract_entry(file *zip.FileHeader) {
	var err error
	var dest *os.File

	if fVerbose {
		fmt.Printf("x %s ", file.Name)
		if file.FileInfo().IsDir() {
			fmt.Println("directory")
		} else {
			fmt.Printf("%d bytes\n",
				file.UncompressedSize)
		}
	}

	dest_path := filepath.Join(destdir, file.Name)
	if file.FileInfo().IsDir() {
		err = os.MkdirAll(dest_path, file.Mode())
	} else {
		var err_creat error /* So 'dest' isn't also a new variable. */
		err_mkdir := os.MkdirAll(filepath.Dir(dest_path), 0755)
		dest, err_creat = os.Create(dest_path)
		err = errors.Join(err_mkdir, err_creat)
		defer dest.Close()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"failed to create %s: %s\n",
			dest_path, err)
		os.Exit(1)
	}

	if !file.FileInfo().IsDir() {
		zfile, err := areader.File[zhip.EntryNo[file.Name]].Open()
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"failed to open %s from %s: %s\n",
				file.Name, archive, err)
			os.Exit(1)
		}
		defer zfile.Close()
		wbytes, err := io.Copy(dest, zfile)
		if uint64(wbytes) != uint64(file.UncompressedSize) {
			fmt.Fprintf(os.Stderr,
				"failed to write %d bytes to the disk; wrote just %d: %s\n",
				file.UncompressedSize, wbytes, err)
			os.Exit(1)
		}
		err = os.Chmod(dest.Name(), file.Mode())
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"failed to restore permissions %04o for file %s: %s\n",
				file.Mode(), dest.Name(), err)
		}
	}
}
