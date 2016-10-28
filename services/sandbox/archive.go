package sandbox

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	MimeDetectLen      = 512 // bytes
	MimeTypeZip        = "application/zip"
	MimeTypeTar        = "application/x-tar"
	MimeTypeGZip       = "application/x-gzip"
	MimeTypeRar        = "application/x-rar-compressed"
	ArchivePermissions = 0555
)

// A generic archive interface that generalizes archive file formats: .zip,
// .tar.gz, etc. So that they can all be interacted with identically. For the
// most part, Botbox doesn't care what particular archive is used, so we can
// abstract the implementation details away.
type Archive interface {
	Files() ([]*ArchiveFile, error)
}

// Convert an opened archive to a tar stream. Handy since Docker likes
// everything to be a Tar stream.
func ArchiveToTar(a Archive) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(buf)
	files, err := a.Files()
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		contents, err := ioutil.ReadAll(f.Reader)
		if err != nil {
			return nil, err
		}
		tr.WriteHeader(&tar.Header{
			Name: f.Name,
			Size: int64(len(contents)),
			Mode: ArchivePermissions,
		})
		tr.Write(contents)
	}
	tr.Close()
	return bytes.NewReader(buf.Bytes()), nil
}

// A file in an archive. It will have a name (which is actually a full URI
// with directories in the path), and a reader to the file contents.
type ArchiveFile struct {
	Name   string
	Reader io.Reader
}

// A concrete implementation of the Archive interface for zip archives. It
// is backed by the golang zip.Reader struct from the standard lib.
type ZipArchive struct {
	reader *zip.Reader
}

// Get a list of files in the zip archive.
func (z *ZipArchive) Files() ([]*ArchiveFile, error) {
	output := []*ArchiveFile{}
	for _, f := range z.reader.File {
		rc, err := f.Open()
		if err != nil && err != io.EOF {
			return nil, err
		}
		defer rc.Close()
		contents, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(contents)
		output = append(output, &ArchiveFile{f.Name, r})
	}
	return output, nil
}

// Open a zip archive.
func OpenZip(r *bytes.Reader) (*ZipArchive, error) {
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, err
	}
	return &ZipArchive{zr}, nil
}

// An archive built from a tar file. It is backed by the golang tar.Reader
// type from the standard lib.
type TarArchive struct {
	reader *tar.Reader
}

// Get a list of files in the tar archive.
func (t *TarArchive) Files() ([]*ArchiveFile, error) {
	output := []*ArchiveFile{}
	for {
		h, err := t.reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		contents, err := ioutil.ReadAll(t.reader)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(contents)
		f := &ArchiveFile{h.Name, r}
		output = append(output, f)
	}

	return output, nil
}

// Open a tar archive.
func OpenTar(r *bytes.Reader) (*TarArchive, error) {
	tr := tar.NewReader(r)
	return &TarArchive{tr}, nil
}

// Open an arbitrary archive. The particular type is determined using
// http.DetectContentType on the first 512 bytes.
func OpenArchive(r io.Reader) (Archive, error) {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(contents)

	header := contents[0:512]
	mimetype := http.DetectContentType(header)

	switch mimetype {
	case MimeTypeZip:
		return OpenZip(b)
	case MimeTypeGZip:
		return nil, errors.New("GZip currently unsupported.")
	case MimeTypeRar:
		return nil, errors.New("Rar currently unsupported.")
	case MimeTypeTar:
		return OpenTar(b)
	default:
		return nil, errors.New("Please use a valid archive type.")
	}

}
