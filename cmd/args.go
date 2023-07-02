package cmd

import (
	"os"
	"strings"

	"github.com/Arapak/sio-tool/codeforces_client"
	"github.com/Arapak/sio-tool/config"
	"github.com/Arapak/sio-tool/sio_client"
	"github.com/Arapak/sio-tool/szkopul_client"

	"github.com/docopt/docopt-go"
)

type ParsedArgs struct {
	CodeforcesInfo codeforces_client.Info
	SzkopulInfo    szkopul_client.Info
	SioInfo        sio_client.Info
	File           string
	Generator      string
	Solve          string
	Brute          string
	Source         string
	Name           string
	Path           string
	Link           string
	Shortname      string
	Contest        string
	Stage          string
	Specifier      []string `docopt:"<specifier>"`
	Alias          string   `docopt:"<alias>"`
	Accepted       bool     `docopt:"ac"`
	All            bool     `docopt:"all"`
	Handle         string   `docopt:"<handle>"`
	Version        string   `docopt:"{version}"`
	Config         bool     `docopt:"config"`
	Submit         bool     `docopt:"submit"`
	List           bool     `docopt:"list"`
	Parse          bool     `docopt:"parse"`
	Gen            bool     `docopt:"gen"`
	Test           bool     `docopt:"test"`
	Watch          bool     `docopt:"watch"`
	Open           bool     `docopt:"open"`
	Stand          bool     `docopt:"stand"`
	Sid            bool     `docopt:"sid"`
	Race           bool     `docopt:"race"`
	Pull           bool     `docopt:"pull"`
	Clone          bool     `docopt:"clone"`
	Upgrade        bool     `docopt:"upgrade"`
	StressTest     bool     `docopt:"stress-test"`
	Database       bool     `docopt:"db"`
	Add            bool     `docopt:"add"`
	Find           bool     `docopt:"find"`
	Goto           bool     `docopt:"goto"`
	Codeforces     bool
	Szkopul        bool
	SioStaszic     bool
	SioMimuw       bool
}

var Args *ParsedArgs

func SubPath(parent, sub string) bool {
	return strings.HasPrefix(parent, sub)
}

func determineClient() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg := config.Instance
	codeforcesDir := SubPath(path, cfg.FolderName["codeforces-root"])
	if codeforcesDir {
		Args.Codeforces = true
		return nil
	}
	sioStaszicDir := SubPath(path, cfg.FolderName["sio-staszic-root"])
	if sioStaszicDir {
		Args.SioStaszic = true
		return nil
	}
	sioMimuwDir := SubPath(path, cfg.FolderName["sio-mimuw-root"])
	if sioMimuwDir {
		Args.SioMimuw = true
		return nil
	}
	szkopulDir := SubPath(path, cfg.FolderName["szkopul-root"])
	if szkopulDir {
		Args.Szkopul = true
		return nil
	}
	return nil
}

func parseArgs(opts docopt.Opts) error {
	if file, ok := opts["--file"].(string); ok {
		Args.File = file
	} else if file, ok := opts["<file>"].(string); ok {
		Args.File = file
	}
	err := determineClient()
	if err != nil {
		return err
	}
	cfg := config.Instance
	if Args.Codeforces {
		return parseArgsCodeforces()
	}
	if Args.SioStaszic {
		if Args.Handle == "" {
			Args.Handle = sio_client.StaszicInstance.Username
		}
		return parseArgsSio(cfg.FolderName["sio-staszic-root"])
	}
	if Args.SioMimuw {
		if Args.Handle == "" {
			Args.Handle = sio_client.MimuwInstance.Username
		}
		return parseArgsSio(cfg.FolderName["sio-mimuw-root"])
	}
	if Args.Szkopul {
		return parseArgsSzkopul()
	}
	return nil
}
