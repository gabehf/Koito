package importer

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
)

// runs after every importer
func finishImport(ctx context.Context, filename string, numImported int) error {
	l := logger.FromContext(ctx)
	_, err := os.Stat(path.Join(cfg.ConfigDir(), "import_complete"))
	if err != nil {
		err = os.MkdirAll(path.Join(cfg.ConfigDir(), "import_complete"), 0744)
		if err != nil {
			l.Err(err).Msg("Failed to create import_complete dir! Import files must be removed from the import directory manually, or else the importer will run on every app start")
		}
	}
	if err := os.MkdirAll(path.Dir(path.Join(cfg.ConfigDir(), "import_complete", filename)), 0744); err != nil {
		return err
	}
	err = os.Rename(path.Join(cfg.ConfigDir(), "import", filename), path.Join(cfg.ConfigDir(), "import_complete", filename))
	if err != nil {
		l.Err(err).Msg("Failed to move file to import_complete dir! Import files must be removed from the import directory manually, or else the importer will run on every app start")
	}
	if numImported != 0 {
		l.Info().Msgf("Finished importing %s; imported %d items", filename, numImported)
	}
	return nil
}

// from https://stackoverflow.com/a/55093788 with modification to use cfg and check for zero values
func inImportTimeWindow(check time.Time) bool {
	end, start := cfg.ImportWindow()
	if start.IsZero() && end.IsZero() {
		return true
	}
	if !start.IsZero() && end.IsZero() {
		return !check.Before(start)
	}
	if start.IsZero() && !end.IsZero() {
		return !check.After(end)
	}
	return !check.Before(start) && !check.After(end)
}
