package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/AlecAivazis/survey/v2"

	"github.com/docopt/docopt-go"

	"github.com/Arapak/sio-tool/codeforces_client"
	"github.com/Arapak/sio-tool/config"
	"github.com/fatih/color"

    "strings"
)

func Eval(opts docopt.Opts) error {
	Args = &ParsedArgs{}
	err := opts.Bind(Args)
	if err != nil {
		return err
	}
	if err := parseArgs(opts); err != nil {
		return err
	}
	if Args.Config {
		return Config()
	} else if Args.Gen {
		return Gen()
	} else if Args.Test {
		return Test()
	} else if Args.StressTest {
		return StressTest()
	} else if Args.Upgrade {
		return Upgrade()
	} else if Args.PackageTest {
		return PackageTest()
	} else if Args.AddPackage {
		return AddPackage()
	} else if Args.Database {
		if Args.Add {
			return DatabaseAdd()
		} else if Args.Find {
			return DatabaseFind()
		} else if Args.Goto {
			return DatabaseGoto()
		}
	} else {
		if Args.Codeforces {
			if Args.Submit {
				return CodeforcesSubmit()
			} else if Args.List {
				return CodeforcesList()
			} else if Args.Parse {
				return CodeforcesParse()
			} else if Args.Watch {
				return CodeforcesWatch()
			} else if Args.Open {
				return CodeforcesOpen()
			} else if Args.Stand {
				return CodeforcesStand()
			} else if Args.Sid {
				return CodeforcesSid()
			} else if Args.Race {
				return CodeforcesRace()
			} else if Args.Pull {
				return CodeforcesPull()
			}
		} else if Args.Szkopul {
			if Args.Submit {
				return SzkopulSubmit()
			} else if Args.Watch {
				return SzkopulWatch()
			} else if Args.Parse {
				return SzkopulParse()
			} else if Args.Sid {
				return SzkopulSid()
			} else if Args.Open {
				return SzkopulOpen()
			} else if Args.List {
				return SzkopulList()
			}
		} else if Args.SioStaszic || Args.SioMimuw || Args.SioTalent {
			if Args.Submit {
				return SioSubmit()
			} else if Args.Watch {
				return SioWatch()
			} else if Args.List {
				return SioList()
			} else if Args.Sid {
				return SioSid()
			} else if Args.Open {
				return SioOpen()
			} else if Args.Parse {
				return SioParse()
			} else if Args.Race {
				return SioRace()
			} else if Args.Stand {
				return SioStand()
			} else if Args.DownloadPackages {
				return SioDownloadPackages()
			} else if Args.UploadPackage {
				return SioUploadPackage()
			}
		}
	}
	color.Red("This function is not available here. Maybe you are in the wrong folder?")
	return nil
}

func getSampleID() (samples []string) {
	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := os.ReadDir(path)
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`in(\w+).txt`)
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("out%v.txt", idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

func getSampleByName(filename string) (samples []string) {
	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := os.ReadDir(path)
	if err != nil {
		return
	}
	reg := regexp.MustCompile(fmt.Sprintf("%s(\\w+).in", regexp.QuoteMeta(filename)))
	for _, path := range paths {
		name := path.Name()
		tmp := reg.FindSubmatch([]byte(name))
		if tmp != nil {
			idx := string(tmp[1])
			ans := fmt.Sprintf("%s%v.out", filename, idx)
			if _, err := os.Stat(ans); err == nil {
				samples = append(samples, idx)
			}
		}
	}
	return
}

// CodeList Name matches some template suffix, index are template array indexes
type CodeList struct {
	Name  string
	Index []int
}

func getCode(filename string, templates []config.CodeTemplate, fileExtensions map[string]struct{}) (codes []CodeList, err error) {
	mp := make(map[string][]int)
	for i, temp := range templates {
		suffixMap := map[string]bool{}
		for _, suffix := range temp.Suffix {
			if _, ok := fileExtensions[suffix]; len(fileExtensions) != 0 && !ok {
				continue
			}
			if _, ok := suffixMap[suffix]; !ok {
				suffixMap[suffix] = true
				sf := "." + suffix
				mp[sf] = append(mp[sf], i)
			}
		}
	}

	if filename != "" {
		ext := filepath.Ext(filename)
		if idx, ok := mp[ext]; ok {
			return []CodeList{{filename, idx}}, nil
		}
		return nil, fmt.Errorf("%v can not match any template. You could add a new template by `st config`", filename)
	}

	path, err := os.Getwd()
	if err != nil {
		return
	}
	paths, err := os.ReadDir(path)
	if err != nil {
		return
	}

	for _, path := range paths {
		name := path.Name()
		ext := filepath.Ext(name)
		if idx, ok := mp[ext]; ok {
			codes = append(codes, CodeList{name, idx})
		}
	}

	return codes, nil
}

func getOneCode(filename string, templates []config.CodeTemplate, fileExtensions map[string]struct{}) (name string, index int, err error) {
    cfg := config.Instance
    tempName := ""
    if filename != "" && !strings.Contains(filename, ".") {
        tempName = filename
        filename = ""
    }

	codes, err := getCode(filename, templates, fileExtensions)
	if err != nil {
		return
	}
	if len(codes) < 1 {
		return "", 0, errors.New("cannot find any code,\nmaybe you should add a new template by `st config`")
	}
	if len(codes) > 1 {
		codeNames := make([]string, len(codes))
		for i, code := range codes {
			codeNames[i] = code.Name
		}
		codeIndex := 0
		prompt := &survey.Select{
			Message: "Multiple files can be selected.",
			Options: codeNames,
		}
		if err = survey.AskOne(prompt, &codeIndex); err != nil {
			return
		}
		codes[0] = codes[codeIndex]
	}
	if len(codes[0].Index) > 1 {
		langs := make([]string, len(codes[0].Index))
        gotten := -1
		for i, idx := range codes[0].Index {
			langs[i] = codeforces_client.Langs[templates[idx].Lang] + " from " + templates[idx].Alias
            if tempName != "" && templates[idx].Alias == tempName {
                gotten = idx
            }
		}
        if tempName == "" {
            gotten = cfg.Default
        }
        contains := false
        for _, idx := range codes[0].Index {
            if idx == gotten {
                contains = true
            }
        }
        if !contains {
            prompt := &survey.Select{
                Message: "Multiple languages match the file.",
                Options: langs,
            }
            if err = survey.AskOne(prompt, &codes[0].Index[0]); err != nil {
                return
            }
        } else {
            codes[0].Index[0] = gotten
        }
	}
	return codes[0].Name, codes[0].Index[0], nil
}
