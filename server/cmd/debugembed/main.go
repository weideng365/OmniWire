//go:build ignore

package main

import (
	"fmt"
	"io/fs"
	"os"

	"omniwire/internal/packed"
)

func main() {
	if !packed.Enabled {
		fmt.Println("EMBED NOT ENABLED!")
		os.Exit(1)
	}
	sub, err := fs.Sub(packed.FS, "public")
	if err != nil {
		fmt.Println("fs.Sub error:", err)
		os.Exit(1)
	}
	fmt.Println("=== Embed FS contents ===")
	fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if !d.IsDir() {
			fmt.Println(path)
		}
		return nil
	})
}
