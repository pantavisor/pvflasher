package image

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		shouldExist bool
		wantErr     bool
	}{
		{
			name: "Open existing file",
			setup: func() string {
				path := filepath.Join(tmpDir, "existing.txt")
				os.WriteFile(path, []byte("test content"), 0644)
				return path
			},
			shouldExist: true,
			wantErr:     false,
		},
		{
			name: "Open non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.txt")
			},
			shouldExist: false,
			wantErr:     true,
		},
		{
			name: "Open directory",
			setup: func() string {
				dir := filepath.Join(tmpDir, "testdir")
				os.Mkdir(dir, 0755)
				return dir
			},
			shouldExist: true,
			wantErr:     false, // directories can be opened for reading on Unix
		},
		{
			name: "Open empty file",
			setup: func() string {
				path := filepath.Join(tmpDir, "empty.txt")
				os.WriteFile(path, []byte{}, 0644)
				return path
			},
			shouldExist: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			f, err := OpenFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if f != nil {
				defer f.Close()
			}

			if tt.shouldExist && f != nil {
				stat, err := f.Stat()
				if err != nil {
					t.Errorf("Stat() error = %v", err)
					return
				}

				if stat.Name() != filepath.Base(path) {
					t.Errorf("Stat().Name() = %v, want %v", stat.Name(), filepath.Base(path))
				}
			}
		})
	}
}

func TestOpenFile_Content(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "content.txt")
	content := []byte("Hello, World!")

	os.WriteFile(path, content, 0644)

	f, err := OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	// Read content
	buf := make([]byte, len(content))
	n, err := f.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if n != len(content) {
		t.Errorf("Read() = %v, want %v", n, len(content))
	}

	if string(buf) != string(content) {
		t.Errorf("Content = %q, want %q", string(buf), string(content))
	}
}

func TestOpenFile_Permissions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with specific permissions
	path := filepath.Join(tmpDir, "readonly.txt")
	os.WriteFile(path, []byte("content"), 0444) // read-only

	f, err := OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	defer f.Close()

	// Should be able to read
	buf := make([]byte, 7)
	_, err = f.Read(buf)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}

	// Try to write (should fail on read-only file)
	_, err = f.Write([]byte("test"))
	if err == nil {
		// Some systems may allow this, so just log
		t.Log("Write succeeded on read-only file (may be expected on some systems)")
	}
}
