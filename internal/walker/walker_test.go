package walker

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectFiles(t *testing.T) {
	r := require.New(t)

	exportDir, files := createFileStructure(t, "export 2024 25",
		"2/a.txt",
		"2/b.txt",
		"1/c.txt",
		"1/c.xmp",
		"4/darktable_exported/whatever.jpeg", // should be ignored
	)

	result, err := SelectFiles(exportDir, []string{"*.xmp", "*/darktable_exported/*"})
	r.NoError(err)
	r.NotNil(result)
	r.Equal("export-2024-25", result.Name, "replace all spaces with dashes")

	r.Equal([]ExportFile{
		{
			Path:    files[2],
			RelPath: "1/c.txt",
			Size:    3,
		},
		{
			Path:    files[0],
			RelPath: "2/a.txt",
			Size:    3,
		},
		{
			Path:    files[1],
			RelPath: "2/b.txt",
			Size:    3,
		},
	}, result.Files)

}

func createFileStructure(t testing.TB, baseDirName string, files ...string) (string, []string) {
	r := require.New(t)

	d := t.TempDir()
	baseDir := filepath.Join(d, baseDirName)

	r.NoError(os.Mkdir(baseDir, os.ModePerm))

	fullFileNames := lo.Map(files, func(item string, _ int) string {
		return filepath.Join(baseDir, strings.ReplaceAll(item, "/", string(os.PathSeparator)))
	})

	for _, fileName := range fullFileNames {
		r.NoError(os.MkdirAll(filepath.Dir(fileName), os.ModePerm))
		r.NoError(os.WriteFile(fileName, []byte{0, 0, 0}, os.ModePerm))
	}
	return baseDir, fullFileNames
}
