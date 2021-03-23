# bmpinvert: Invert 32 bpp BMP files

## Requirements
* Golang (<https://golang.org/doc/install>)
## Installation
### From go
```bash
go get github.com/rdilip/bmpinvert
go install github.com/rdilip/bmpinvert
```
### From source
Clone the repository, then run `go build`.

## Usage
`bmpinvert` takes two command line flags.
```bash
-savedir: Specify the directory in which to save the inverted images.
-inpdir: Specify a directory from which to extract inverted images.
```
`bmpinvert` also takes a `-h` flag that prints the above summary. To use `bmpinvert`, run

```bash
./bmpinvert -savedir=SAVEDIR -inpdir=INPDIR file1.bmp file2.bmp...
```

All files must be 32 bpp, and the program assumes no compression was used on the images. You can list any number of bmp files in the current directory, and one additional input directory. All files will be saved to savedir. 

## Notes
* Relatively untested on non-Unix systems -- it should work, but there might be sharp edges. 
* The resources consulted were the Golang documentation and the BMP Wikipedia page. 
