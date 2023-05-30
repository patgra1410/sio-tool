package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Arapak/sio-tool/codeforces_client"
	"github.com/Arapak/sio-tool/config"
)

// Parse command
func CodeforcesParse() (err error) {
	cfg := config.Instance
	cln := codeforces_client.Instance
	info := Args.CodeforcesInfo
	source := ""
	ext := ""
	if cfg.GenAfterParse {
		if len(cfg.Template) == 0 {
			return errors.New("you have to add at least one code template by `st config`")
		}
		path := cfg.Template[cfg.Default].Path
		ext = filepath.Ext(path)
		if source, err = readTemplateSource(path, cln); err != nil {
			return
		}
	}

	db, err := sql.Open("sqlite", cfg.DbPath)
	if err != nil {
		fmt.Printf("failed to open database connection: %v\n", err)
		return
	}
	defer db.Close()

	work := func() error {
		_, paths, err := cln.Parse(info, db)
		if err != nil {
			return err
		}
		if cfg.GenAfterParse {
			for _, path := range paths {
				GenFiles(source, path, ext)
			}
		}
		return nil
	}
	if err = work(); err != nil {
		if err = loginAgainCodeforces(cln, err); err == nil {
			err = work()
		}
	}
	return
}
