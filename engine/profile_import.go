package engine

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/importer"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
)

func importProfileFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
	switch {
	case strings.Contains(filename, "Streaming_History_Audio"):
		return importer.ImportSpotifyFile(ctx, store, mbzc, filename)
	case strings.Contains(filename, "maloja"):
		return importer.ImportMalojaFile(ctx, store, mbzc, filename)
	case strings.Contains(filename, "recenttracks"):
		return importer.ImportLastFMFile(ctx, store, mbzc, filename)
	case strings.Contains(filename, "listenbrainz"):
		return importer.ImportListenBrainzExport(ctx, store, mbzc, filename)
	case strings.Contains(filename, "koito"):
		return importer.ImportKoitoFile(ctx, store, filename)
	default:
		return fmt.Errorf("unrecognized import format")
	}
}
func profileImportHandler(store db.DB, mbzc mbz.MusicBrainzCaller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if contentType == "multipart/form-data" {
			uploadProfileImport(w, r, store, mbzc)
			return
		}
		body, err := utils.DecodeBody[struct {
			Filename string `json:"filename"`
		}](r)
		if err != nil || body.Filename == "" || path.Base(body.Filename) != body.Filename || strings.ContainsAny(body.Filename, `\\/`) || body.Filename == "." || body.Filename == ".." {
			utils.WriteError(w, "a simple filename is required", 400)
			return
		}
		filename := path.Join(fmt.Sprint(db.ProfileID(r.Context())), body.Filename)
		// Do not follow symlinks supplied in the import directory.
		dirInfo, dirErr := os.Lstat(path.Join(cfg.ConfigDir(), "import", fmt.Sprint(db.ProfileID(r.Context()))))
		if dirErr != nil || !dirInfo.IsDir() {
			utils.WriteError(w, "profile import directory not found", 404)
			return
		}
		info, err := os.Lstat(path.Join(cfg.ConfigDir(), "import", filename))
		if err != nil || !info.Mode().IsRegular() {
			utils.WriteError(w, "import file not found", 404)
			return
		}
		if err := importProfileFile(r.Context(), store, mbzc, filename); err != nil {
			utils.WriteError(w, "import failed", 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Uploads have server-generated filenames; the selected format selects the
// existing parser, while the URL selects the destination independently of file content.
func uploadProfileImport(w http.ResponseWriter, r *http.Request, store db.DB, mbzc mbz.MusicBrainzCaller) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		utils.WriteError(w, "import upload exceeds 100 MiB or is invalid", 400)
		return
	}
	defer r.MultipartForm.RemoveAll()
	formats := map[string]string{"koito": "koito.json", "listenbrainz": "listenbrainz.zip", "spotify": "Streaming_History_Audio.json", "lastfm": "recenttracks.json", "maloja": "maloja.json"}
	suffix, ok := formats[r.FormValue("format")]
	if !ok {
		utils.WriteError(w, "select a supported import format", 400)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, "import file required", 400)
		return
	}
	defer file.Close()
	dir := path.Join(cfg.ConfigDir(), "import", fmt.Sprint(db.ProfileID(r.Context())))
	if err = os.MkdirAll(dir, 0700); err != nil {
		utils.WriteError(w, "could not prepare import", 500)
		return
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		utils.WriteError(w, "invalid profile import directory", 400)
		return
	}
	name := uuid.NewString() + "-" + suffix
	destination := path.Join(dir, name)
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		utils.WriteError(w, "could not save import", 500)
		return
	}
	_, err = io.Copy(out, file)
	closeErr := out.Close()
	if err != nil || closeErr != nil {
		_ = os.Remove(destination)
		utils.WriteError(w, "could not save import", 500)
		return
	}
	// Failed parsers can have saved a prefix of the input. Remove the pending
	// upload so a later startup cannot retry it without an explicit admin action.
	defer os.Remove(destination)
	if err = importProfileFile(r.Context(), store, mbzc, path.Join(fmt.Sprint(db.ProfileID(r.Context())), name)); err != nil {
		utils.WriteError(w, "import failed; some records may have been saved. Check the file before retrying", 400)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
