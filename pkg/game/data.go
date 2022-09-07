package game

import (
	"bytes"
	"compress/bzip2"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gabstv/go-bsdiff/pkg/bspatch"
	"github.com/kralicky/ttr/pkg/api"
	log "github.com/sirupsen/logrus"
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
	log.WithField("dir", dir).Debug("using data directory")
	return dir, os.MkdirAll(dir, 0755)
}

func SyncGameData(ctx context.Context, client api.DownloadClient) error {
	log.Debug("syncing game data")
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
			log.WithField("filename", filename).Debug("skipping download")
			continue
		}
		filename := filename
		spec := spec
		eg.Go(func() error {
			f, err := os.OpenFile(filepath.Join(dataDir, filename), os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", filename, err)
			}
			defer f.Close()

			hash := sha1.New()
			if _, err := io.Copy(hash, f); err != nil {
				return fmt.Errorf("error reading file %s: %w", filename, err)
			}
			sum := hex.EncodeToString(hash.Sum(nil))
			if sum == spec.Hash {
				// file is up to date
				log.WithField("filename", filename).Debug("file is up to date")
				return nil
			}

			f.Seek(0, io.SeekStart)

			// check if there is a known patch available for the file we have
			if p, ok := spec.Patches[sum]; ok {
				return fetchAndPatchFile(ctx, client, filename, spec, p, f)
			}

			return fetchAndUpdateFile(ctx, client, filename, spec, f)
		})
	}

	return eg.Wait()
}

func fetchAndUpdateFile(
	ctx context.Context,
	client api.DownloadClient,
	filename string,
	spec *api.ManifestEntry,
	f *os.File,
) error {
	log.WithField("filename", filename).Debug("updating file")

	wc, err := client.DownloadFile(ctx, spec.Download)
	if err != nil {
		return fmt.Errorf("error downloading file %s: %w", filename, err)
	}
	defer wc.Close()
	// copy wc to f, and compute the hash of the downloaded contents as we go
	compHash := sha1.New()
	decompHash := sha1.New()
	bz2 := io.TeeReader(bzip2.NewReader(io.TeeReader(wc, compHash)), decompHash)
	f.Truncate(0)
	if _, err := io.Copy(f, bz2); err != nil {
		return fmt.Errorf("error while writing file %s: %w", filename, err)
	}

	// compare the hash of the compressed and decompressed contents with the expected hash
	if sum := hex.EncodeToString(compHash.Sum(nil)); sum != spec.CompressedHash {
		return fmt.Errorf("hash mismatch: downloaded contents of file %s do not match the expected hash", filename)
	}
	if sum := hex.EncodeToString(decompHash.Sum(nil)); sum != spec.Hash {
		return fmt.Errorf("hash mismatch: decompressed contents of file %s do not match the expected hash", filename)
	}

	return nil
}

func fetchAndPatchFile(
	ctx context.Context,
	client api.DownloadClient,
	filename string,
	spec *api.ManifestEntry,
	patch *api.PatchSpec,
	f *os.File,
) error {
	log.WithField("filename", filename).Debug("patching file")

	wc, err := client.DownloadFile(ctx, patch.Filename)
	if err != nil {
		return fmt.Errorf("error downloading patch %s: %w", filename, err)
	}
	defer wc.Close()

	patchBuf := new(bytes.Buffer)
	// compute the hash of the downloaded contents as we go
	compHash := sha1.New()
	decompHash := sha1.New()
	bz2 := io.TeeReader(bzip2.NewReader(io.TeeReader(wc, compHash)), decompHash)
	if _, err := io.Copy(patchBuf, bz2); err != nil {
		return fmt.Errorf("error while downloading patch %s: %w", filename, err)
	}

	// compare the hash of the compressed and decompressed contents with the expected hash
	if sum := hex.EncodeToString(compHash.Sum(nil)); sum != patch.CompressedPatchHash {
		return fmt.Errorf("hash mismatch: downloaded contents of patch %s do not match the expected hash", filename)
	}
	if sum := hex.EncodeToString(decompHash.Sum(nil)); sum != patch.PatchHash {
		return fmt.Errorf("hash mismatch: decompressed contents of patch %s do not match the expected hash", filename)
	}

	// apply the bsdiff patch

	patchedFileBuf := new(bytes.Buffer)
	patchedSum := sha1.New()
	if err := bspatch.Reader(f, io.MultiWriter(patchedFileBuf, patchedSum), patchBuf); err != nil {
		return fmt.Errorf("error while applying patch %s: %w", filename, err)
	}

	// compare the hash of the patched file with the expected hash
	if sum := hex.EncodeToString(patchedSum.Sum(nil)); sum != spec.Hash {
		return fmt.Errorf("hash mismatch: patched file %s does not match the expected hash", filename)
	}

	// write the patched file to disk
	f.Truncate(0)
	if _, err := io.Copy(f, patchedFileBuf); err != nil {
		return fmt.Errorf("error while writing file %s: %w", filename, err)
	}

	return nil
}
