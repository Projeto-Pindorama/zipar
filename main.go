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
	"encoding/json"
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
	fJSON            bool
	fExplode         bool
	fVerbose         bool
	fCreate          bool
	fTableOfContents bool
	fExtract         bool
	destdir          string
	archive          string
	largest_file     int
)

func main() {
	/* Options. */
	flag.BoolVar(&fJSON, "json", false,
		"Print archive information as JSON.")
	flag.BoolVar(&fExplode, "explode", false,
		"Explode the archive into the disk, ignoring directory hierarchy. "+
			"Can also be used when creating archives. (TODO)")
	flag.BoolVar(&fVerbose, "verbose", false,
		"Enable verbose output.")
	flag.BoolVar(&fCreate, "create", false,
		"Create a new zipfile and write named files into it.")
	flag.BoolVar(&fTableOfContents, "toc", false,
		"List the contents of the zipfile.")
	flag.BoolVar(&fExtract, "extract", false,
		"The named files are extracted from the zipfile.")
	flag.StringVar(&destdir, "chdir", ".",
		"Use the next argument as the directory to place the files into.")
	flag.StringVar(&archive, "file", "",
		"Use the next argument as the name of the archive.")
	getopt.Aliases(
		"d", "explode",
		"v", "verbose",
		"c", "create",
		"t", "toc",
		"x", "extract",
		"f", "file",
		"C", "chdir",
	)
	getopt.Parse()
	flag.Usage = usage

	/*
	 * Extra arguments; possibly specific
	 * files to be extracted and, for sure,
	 * files to be included in a zipfile
	 * --- in case of the '-c' option.
	 */
	extra := flag.Args()
	nextra := flag.NArg()

	if len(os.Args) < 2 || archive == "" {
		flag.Usage()
	}

	/*
	 * "In America, there's plenty of light beer and you can always
	 * find a party! In Russia, the Party always finds you."
	 */
	switch true {
		case fCreate:
			newfile, err := os.Create(archive)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Failed to create %s: %s\n", archive, err)
			}
			defer newfile.Close()
			awriter := zip.NewWriter(newfile)
			/* Best compression per default. */
			zhip.CompressionMethod = zip.Deflate
			/* Assuming extra arguments as files to be added. */
			for f := 0; f < len(extra); f++ {
				record_entry(awriter, extra[f])
			}
			_ = awriter.Close()
		case fTableOfContents, fExtract:
			areader, err := zip.OpenReader(archive)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"failed to open %s: %s\n",
					archive, err)
			}
			defer areader.Close()

			switch true {
				case (fTableOfContents && fJSON):
					/*
					 * Obtain the entire *zip.FileHeader slice,
					 * Marshal it and print as JSON.
					 */
					eslice := zhip.GetZipESlice(areader)
					/*
					 * TODO: Instead of using
					 * zip.FileHeader, create a new struct at
					 * libcmon that contains more useful
					 * information for command-line usage.
					 */
					jsoninfo, err := json.MarshalIndent(eslice, "", "  ")
					if err != nil {
						fmt.Fprintf(os.Stderr,
							"Error marshaling JSON: %s\n",
							err)
						os.Exit(1)
					}
					fmt.Print(string(jsoninfo))
					os.Exit(0)
				case fTableOfContents:
					/*
					 * Obtain the largest file size integer
					 * length for the '-t' option formatting.
					 */
					largest_file = len(strconv.FormatUint(
						uint64(zhip.GetZipLargestEntry(areader)), 10))
				case (fExtract && fJSON): /* For '-x' with --json. */
					/* Open and close JSON object in
					 * case of fJSON being true. */
					fmt.Println("[")
					defer fmt.Print("]")
			}
			zipwalk:
			for ;; {
				file := zhip.GetZipEntries(areader)
				if file == nil {
					break
				}
				/*
				 * Check if the user specified files to be extracted.
				 * Perhaps this could go into libcmon too.
				 */
				for f := 0; nextra != 0 && f < nextra; f++ {
					if !strings.HasPrefix(file.Name, extra[f]) {
						continue zipwalk
					}
				}
				switch true {
					case fTableOfContents:
						print_entry_info(file)
					case fExtract:
						extract_entry(areader, file)
				}
			}
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

func record_entry(awriter *zip.Writer, newent string) {
	nentinfo, err := os.Stat(newent)

	if fVerbose {
		fmt.Printf("a %s ", newent)
	}

	wbytes, err := zhip.RecordNewEntry(awriter, newent)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Failed to record entry '%s' into zipfile: %s\n",
			newent, err)
	}

	/*
	 * Recursive directory logic.
	 */
	if nentinfo.IsDir() {
		if fVerbose {
			fmt.Println("directory")
		}
		direntries, err := os.ReadDir(newent)
		if err != nil {
			fmt.Fprintln(os.Stderr,
				"Failed to record directory %s to zipfile: %s\n",
				newent, err)
		}
		for e := 0; e < len(direntries); e++ {
			dentry := direntries[e].Name()
			dentry_fpath := filepath.Join(newent, dentry)
			record_entry(awriter, dentry_fpath)
		}

	} else {
		if fVerbose {
			fmt.Printf("%d bytes\n", wbytes)
		}
	}
}

func extract_entry(areader *zip.ReadCloser, file *zip.FileHeader) {
	var err error
	var dest_path string
	var dest *os.File

	/* Business as usual. */
	if fVerbose && !fJSON {
		fmt.Printf("x %s ", file.Name)
	} else if fJSON {
		jsoninfo, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to parse %T to JSON: %s\n",
				file, err)
			return
		}
		fmt.Printf("%s,\n", string(jsoninfo))
	}

	/* Check if we are extracting the entire directory hierarchy.
	 * For some reason, this is an option that has been requisited
	 * per users of Info-ZIP's unzip command for some years, so we
	 * will be having it implemented here as well. */
	if !fExplode { /* Default. */
		dest_path = filepath.Join(destdir, file.Name)
	} else {
		dest_path = filepath.Join(destdir, filepath.Base(file.Name))
	}

	if file.FileInfo().IsDir() && !fExplode {
		err = os.MkdirAll(dest_path, file.Mode())
		if fVerbose && !fJSON {
			fmt.Println("directory")
		}
	} else {
		var err_creat error /* So 'dest' isn't also a new variable. */
		/* Not to be preoccupied with os.MkdirAll() and fExplode, since
		 * we've already basename()'d the file.Name and the dirname() of
		 * it will be something as the destination directory --- informed
		 * per '-C' or just the current working directory. */
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
		if fVerbose && !fJSON {
			fmt.Printf("%d bytes\n", wbytes)
		}
	}
}

func usage() {
	/* getopt.PrintDefaults() will already be executed. */
	fmt.Fprintf(flag.CommandLine.Output(),
		"%s: Missing command, must specify -x, -c or -t.\n",
		os.Args[0])
	os.Exit(1)
}
