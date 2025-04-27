/*
 * zipar - tar(1)-esque program for Zip files.
 *
 * Copyright (C) 2025 Luiz Antônio Rangel (takusuman)
 *
 * SPDX-Licence-Identifier: MIT 
 *
 */

package main

import (
	"archive/zip"
	"fmt"
	"os"
	"pindorama.net.br/libcmon/zhip"
)

func main() {
	arc, err := zip.OpenReader(os.Args[1])
	if err != nil {
		os.Exit(1)
	}
	defer arc.Close()

	file := zhip.GetZipEntries(arc)
	for ; file != nil; file = zhip.GetZipEntries(arc) {
		fmt.Printf("%#v\n", file)
	}

	// verbose: "x %s, %lld bytes, ""%lld%s\n"
}
