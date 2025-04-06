package tar_utils

import (
	"archive/tar"
	"io"
	"os"
)

func AddStringToTar(tw *tar.Writer, name string, content string) {
	AddBytesToTar(tw, name, []byte(content))
}

func AddBytesToTar(tw *tar.Writer, name string, content []byte) {
	hdr := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		panic(err)
	}
	if _, err := tw.Write(content); err != nil {
		panic(err)
	}
}

func AddFileToTar(tw *tar.Writer, srcPath string, tarPath string) {
	file, err := os.Open(srcPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		panic(err)
	}

	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		panic(err)
	}
	hdr.Name = tarPath // e.g. files/foo.wav

	if err := tw.WriteHeader(hdr); err != nil {
		panic(err)
	}
	if _, err := io.Copy(tw, file); err != nil {
		panic(err)
	}
}
