// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tar_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Example_minimal() {
	// Create and add some files to the archive.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	var files = []struct {
		Name, Body string
	}{
		{"readme.txt", "This archive contains some text files."},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	// Open and iterate through the files in the archive.
	tr := tar.NewReader(&buf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, tr); err != nil {
			log.Fatal(err)
		}
		fmt.Println()
	}

	// Output:
	// Contents of readme.txt:
	// This archive contains some text files.
	// Contents of gopher.txt:
	// Gopher names:
	// George
	// Geoffrey
	// Gonzo
	// Contents of todo.txt:
	// Get animal handling license.
}

func TestTarOneFile(t *testing.T) {
	srcFile := "test1.txt"
	destFile := "test.tar"

	fw, err := os.Create(destFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	tw := tar.NewWriter(fw)
	defer tw.Close()

	fi, err := os.Stat(srcFile)
	if err != nil {
		log.Fatal(err)
	}

	// 将源文件的文件信息写入tar.*Header，并写入tw中
	hdr, err := tar.FileInfoHeader(fi, "")
	err = tw.WriteHeader(hdr)
	if err != nil {
		log.Fatal(err)
	}

	fr, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fr.Close()

	written, err := io.Copy(fw, fr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("共写入了%d个字符数", written)
}

func TestUnTar2SingleFile(t *testing.T) {
	srcFile := "test.tar"

	fr, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fr.Close()

	tr := tar.NewReader(fr)

	//for hdr, err := tr.Next(); err != io.EOF; hdr, err = tr.Next() {
	//
	//}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fi := hdr.FileInfo()
		fw, err := os.Create(fi.Name())
		if err != nil {
			log.Fatal(err)
		}

		written, err := io.Copy(fw, tr)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("解压 %s to %s ，共%d个字符数", srcFile, fi.Name(), written)

		os.Chmod(fi.Name(), fi.Mode().Perm())

		fw.Close()
	}
}

func TestTarFiles(t *testing.T) {
	src := "test"
	dst := src + ".tar"

	err := Tar(src, dst)
	if err != nil {
		log.Fatal(err)
	}
}

func Tar(src, dst string) error {
	fw, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	//gw := gzip.NewWriter(fw)
	//defer gw.Close()

	tw := tar.NewWriter(fw)
	defer tw.Close()

	return filepath.Walk(src, func(fileName string, fi fs.FileInfo, err error) error {
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		hdr.Name = strings.TrimPrefix(fileName, string(filepath.Separator))

		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}

		// 判断下文件是否是标准文件，如果不是就不处理了，
		if !fi.Mode().IsRegular() {
			return nil
		}

		fr, err := os.Open(fileName)
		if err != nil {
			return err
		}

		written, err := io.Copy(tw, fr)
		if err != nil {
			return err
		}

		log.Printf("成功打包 %s ，共写入了 %d 字节的数据\n", fileName, written)

		return nil
	})
}

func TestUnTar(t *testing.T) {
	src := "test.tar"
	dst := ""

	err := UnTar(dst, src)
	if err != nil {
		log.Fatal(err)
	}
}

func UnTar(dst, src string) error {
	fr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fr.Close()

	tr := tar.NewReader(fr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if hdr == nil {
			continue
		}

		dstFileDir := filepath.Join(dst, hdr.Name)
		fmt.Println("dstFileDir:", dstFileDir)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(dstFileDir, 0775); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

			written, err := io.Copy(file, tr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("解压 %s，共%d个字符数", src, written)

			file.Close()
		}
	}
	return nil
}
