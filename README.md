# zipar

From [Wiktionary](https://en.wiktionary.org), adapted:
> /ziˈpa(ʁ)/ - (Portuguese) verb
> 1. (transitive, computing) to
> [zip](https://en.wiktionary.org/wiki/zip#English) --- in the sense of converting
> a computer file into a smaller package.
>
> Etymology: From English zip + [-ar](https://en.wiktionary.org/wiki/-ar#Etymology_1_7)

``zipar`` is a program that extracts, lists and creates
[PKZIP](https://support.pkware.com/pkzip/appnote)-compatible files, but using
a ``tar``(1)-like interface inspired on
[Schily's tar](https://cdrtools.sourceforge.net/old/private/man/star/star.1.html)
and [Heirloom (UNIX v7) tar](http://heirloom-ng.pindorama.net.br/manual/man1/tar.1.html).  
It is the first program to ever making use of the
[libcmon](https://pindorama.net.br/libcmon), which is still on testing and with
features being gradually implemented.  
**__Not__** to be confunded with Ryotaro Banno's
([``ushitora-anqou``](https://github.com/ushitora-anqou))
[ZipAr](https://github.com/ushitora-anqou/zipar), which just archives files
with no compression, using multi-thread parallelism for velocity, and that
appears to be sort of niche compared to this project considering it does
nothing besides that. It is also written entirely in OCaml.

## Some history (because why not?)

This project originally was thought as some sort of shell script that would work
as a boilerplate command to Info-ZIP's
[``unzip``](https://infozip.sourceforge.net/UnZip.html)/
[``zip``](https://infozip.sourceforge.net/Zip.html) programs so that both could
be used in a saner<sup><a href="https://xkcd.com/1168/"
target="_blank">(or maybe not)</a></sup> way. Some time passed and, in
mid-2023, I decided to start doing ~~and ended up not finishing~~ a [cbr to pdf
converter](https://github.com/takusuman/cbr2pdf), and saw that dealing with zip
files using Go's ``archive/zip`` library was actually pretty good, so it would
be better to have ``zipar`` as an independent program than as a boilerplate to
two binaries. Then, ≃1.7 years in the future (now), I decided to recycle that
old code from cbr2pdf into libcmon and develop zipar over it.  
Of course, it is not as far as portable as Info-ZIP's unzip/zip programs, nor it
has __all__ of the functionality of these (yet), but I can say it is pretty much
reliable as far as I have tested, and that it will be at Copacabana's base
system. Testing is welcome.

## Features

**NOTE**: Features that have to be implemented via the
[libcmon/zhip](https://pkg.go.dev/pindorama.net.br/libcmon/zhip)
library first will be highlighted as such with the "👊" symbol.

- [X] ``tar``'s "tripartite": create, extract and list files;
- [] Be able to select the compression level (similar to Info-ZIP's zip '``-Z``'
  option);
- [] Be able to create an archive using the std.in./an file (similar to UNIX
  v7/Heirloom ``tar``'s ``-I`` option) as the list of files to be added;
- [] Print information with the ``--json`` option that will be actually useful
  (needs a new struct type at zhip 👊);
- [] "Explode"/junk or ignore directories when creating archives 👊;
- [] Extract files with passwords 👊;
- [] Amend/update/exclude entries from zip files 👊;
- [] Interpret the [``Extra`` section of the zip.FileHeader
  struct](https://pkg.go.dev/archive/zip#FileHeader) for information such as
  Info-ZIP's extensions for UNIX permissions, NTFS info., etc. 👊

## Licence
The
[MIT licence](https://github.com/Projeto-Pindorama/libcmon?tab=License-1-ov-file).

### Who can I blame for it?
Me, Luiz Antonio (a.k.a takusuman).
