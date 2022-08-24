package game

import (
	"compress/bzip2"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/kralicky/ttr/pkg/api"
)

func DataDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "ttr-cli-data"), nil
}

func UpsertDataDir() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return dir, os.MkdirAll(dir, 0755)
}

func SyncGameData(ctx context.Context, client api.Client) error {
	patchManifest, err := client.DownloadPatchManifest(ctx)
	if err != nil {
		return err
	}

	dataDir, err := DataDir()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for filename, spec := range patchManifest {
		if !ShouldDownload(spec) {
			continue
		}
		wg.Add(1)
		go func(filename string, spec *api.PatchSpec) {
			defer wg.Done()
			f, err := os.OpenFile(filepath.Join(dataDir, filename), os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening file %s: %v\n", filename, err)
				return
			}
			defer f.Close()

			hash := sha1.New()
			io.Copy(hash, f)
			f.Seek(0, io.SeekStart)
			if fmt.Sprintf("%x", hash.Sum(nil)) == spec.Hash {
				// file is up to date
				return
			}

			wc, err := client.DownloadFile(ctx, spec.Download)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error downloading file %s: %v\n", filename, err)
				return
			}
			defer wc.Close()
			// copy wc to f, and compute the hash of the downloaded contents as we go
			dlHash := sha1.New()
			bz2 := bzip2.NewReader(io.TeeReader(wc, dlHash))
			if _, err := io.Copy(f, bz2); err != nil {
				fmt.Fprintf(os.Stderr, "error while writing file %s: %v\n", filename, err)
				return
			}

			// compare the hash of the compressed contents with the expected hash
			sum := fmt.Sprintf("%x", dlHash.Sum(nil))
			if sum != spec.CompressedHash {
				fmt.Println("got:", sum)
				fmt.Println("exp:", spec.CompressedHash)
				panic("hash mismatch: downloaded contents of file " + filename + " do not match the expected hash")
			}
		}(filename, spec)
	}
	wg.Wait()
	return nil
}
