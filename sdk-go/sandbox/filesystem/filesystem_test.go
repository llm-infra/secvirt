package filesystem

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesystem(t *testing.T) {
	fs := NewFileSystem(
		"http://192.168.134.142:48008",
		"",
		"root",
	)

	entries, err := fs.List(t.Context(), "/app", 0)
	assert.NoError(t, err)
	fmt.Println(entries)

	_, err = fs.Mkdir(t.Context(), "/app/ttt")
	assert.NoError(t, err)
	_, err = fs.Mkdir(t.Context(), "/app/ttt")
	assert.NoError(t, err)

	exist, err := fs.Exist(t.Context(), "/app")
	assert.NoError(t, err)
	fmt.Println(exist)

	err = fs.Write(t.Context(), "readme", []byte("hello"))
	assert.NoError(t, err)

	data, err := fs.Read(t.Context(), "readme")
	assert.NoError(t, err)
	fmt.Println(data)

	r, err := fs.ReadStream(t.Context(), "readme")
	assert.NoError(t, err)

	buf := make([]byte, 100)
	n, err := r.Read(buf)
	assert.NoError(t, err)
	fmt.Println(buf, n)
}
