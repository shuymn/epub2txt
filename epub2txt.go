package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/k3a/html2text"
	"github.com/kapmahc/epub"
)

const (
	ExitOK = iota
	ExitErr
)

func main() {
	// error if there are too few arguments
	if len(os.Args) != 2 {
		usage()
		os.Exit(ExitErr)
	}

	pattern := os.Args[1]
	paths, err := filepath.Glob(pattern)
	if err != nil {
		usage()
		os.Exit(ExitErr)
	}

	// convert the files matching the pattern one by one
	for _, path := range paths {
		// open epub
		book, err := epub.Open(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(ExitErr)
		}

		// create txt file and convert
		ext := filepath.Ext(path)
		dest := fmt.Sprintf("%s.%s", path[:len(path)-len(ext)], "txt")
		err = func() error {
			txt, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer txt.Close()

			return convert(book, book.Ncx.Points, txt)
		}()
		if err != nil {
			fmt.Println(err)
			os.Exit(ExitErr)
		}
	}
	os.Exit(ExitOK)
}

func usage() {
	fmt.Println("Usage: epub2txt <path-to-epub>")
}

func convert(book *epub.Book, nps []epub.NavPoint, w io.Writer) error {
	for _, np := range nps {
		// get file name
		src := np.Content.Src
		txt, err := func() (string, error) {
			// open xhtml
			f, err := book.Open(src)
			if err != nil {
				return "", err
			}
			defer f.Close()

			html, err := io.ReadAll(f)
			if err != nil {
				return "", err
			}

			// convert to txt
			return html2text.HTML2Text(string(html)), nil
		}()
		if err != nil {
			return err
		}

		// write to txt file
		_, err = w.Write([]byte(txt))
		if err != nil {
			return err
		}

		// convert recursively
		if len(np.Points) > 0 {
			if err = convert(book, np.Points, w); err != nil {
				return err
			}
		}
	}
	return nil
}
