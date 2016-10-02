package sandbox

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
)

const (
	MimeDetectLen      = 512 // bytes
	MimeTypeZip        = "application/zip"
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
	io.Closer
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
		defer f.Close()
		contents, err := ioutil.ReadAll(f.ReadCloser)
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
	Name       string
	ReadCloser io.ReadCloser
}

// Close an archive file when finished.
func (f *ArchiveFile) Close() error {
	return f.ReadCloser.Close()
}

// A concrete implementation of the Archive interface for zip archives. It
// is backed by the golang zip.Reader struct from the standard lib.
type ZipArchive struct {
	reader *zip.Reader
}

// Get a list of files in the zip archive. Don't forget to close the file
// readers when you are done with them.
func (z *ZipArchive) Files() ([]*ArchiveFile, error) {
	output := []*ArchiveFile{}
	for _, f := range z.reader.File {
		rc, err := f.Open()
		if err != nil && err != io.EOF {
			return nil, err
		}
		output = append(output, &ArchiveFile{f.Name, rc})
	}
	return output, nil
}

// Close the zip archive when you're done with it.
func (z *ZipArchive) Close() error {
	return nil
}

// Open a zip archive.
func OpenZip(r *bytes.Reader) (*ZipArchive, error) {
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, err
	}
	return &ZipArchive{zr}, nil
}

// Open an arbitrary archive. The particular type is determined using
// http.DetectContentType on the first 512 bytes.
func OpenArchive(r io.Reader) (Archive, error) {
	// TODO: can this be done without reading all of the bytes into memory? How
	// does this affect performance/memory usage.
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(contents)
	return OpenZip(b)

	//mimetype := http.DetectContentType(header)
	//header := make([]byte, 512)
	//_, err = r.Read(header)
	//if err != nil && err != io.EOF {
	//return nil, err
	//}

	//	fmt.Println(mimetype)
	//	switch mimetype {
	//	case MimeTypeZip:
	//		return OpenZip(b)
	//	case MimeTypeGZip:
	//		return nil, errors.New("GZip currently unsupported.")
	//	case MimeTypeRar:
	//		return nil, errors.New("Rar currently unsupported.")
	//	default:
	//		return nil, errors.New("Please use a valid archive type.")
	//	}

}
