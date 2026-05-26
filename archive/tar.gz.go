package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TarGz handles tar.gz extraction operations.
type TarGz struct{}

// NewTarGz creates a new TarGz extractor.
func NewTarGz() *TarGz {
	return &TarGz{}
}

// UntarFile 从 tar.gz 中找到 basename 匹配 filename 的条目，写入 dest（权限 0755）。
func (t *TarGz) UntarFile(tarGzPath, filename, dest string) error {
	if tarGzPath == "" {
		return fmt.Errorf("tar.gz path is required")
	}
	if filename == "" {
		return fmt.Errorf("filename is required")
	}
	if dest == "" {
		return fmt.Errorf("destination path is required")
	}

	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != filename {
			continue
		}

		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, tr); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("archive: %s not found in %s", filename, tarGzPath)
}
