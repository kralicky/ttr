package game

import (
	"compress/bzip2"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kralicky/ttr/pkg/api"
	"golang.org/x/sync/errgroup"
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

	eg, ctx := errgroup.WithContext(ctx)
	for filename, spec := range patchManifest {
		if !ShouldDownload(spec) {
			continue
		}
		filename := filename
		spec := spec
		eg.Go(func() error {
			f, err := os.OpenFile(filepath.Join(dataDir, filename), os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return fmt.Errorf("error opening file %s: %v", filename, err)
			}
			defer f.Close()

			hash := sha1.New()
			if _, err := io.Copy(hash, f); err != nil {
				return fmt.Errorf("error reading file %s: %v", filename, err)
			}
			if fmt.Sprintf("%x", hash.Sum(nil)) == spec.Hash {
				// file is up to date
				return nil
			}
			f.Seek(0, io.SeekStart)
			f.Truncate(0)

			wc, err := client.DownloadFile(ctx, spec.Download)
			if err != nil {
				return fmt.Errorf("error downloading file %s: %v", filename, err)
			}
			defer wc.Close()
			// copy wc to f, and compute the hash of the downloaded contents as we go
			compHash := sha1.New()
			decompHash := sha1.New()
			bz2 := io.TeeReader(bzip2.NewReader(io.TeeReader(wc, compHash)), decompHash)
			if _, err := io.Copy(f, bz2); err != nil {
				return fmt.Errorf("error while writing file %s: %v", filename, err)
			}

			// compare the hash of the compressed contents with the expected hash

			if sum := fmt.Sprintf("%x", compHash.Sum(nil)); sum != spec.CompressedHash {
				fmt.Println("got:", sum)
				fmt.Println("exp:", spec.CompressedHash)
				return fmt.Errorf("hash mismatch: downloaded contents of file %s do not match the expected hash", filename)
			}
			if sum := fmt.Sprintf("%x", decompHash.Sum(nil)); sum != spec.Hash {
				fmt.Println("got:", sum)
				fmt.Println("exp:", spec.Hash)
				return fmt.Errorf("hash mismatch: decompressed contents of file %s do not match the expected hash", filename)
			}
			return nil
		})
	}

	return eg.Wait()
}
