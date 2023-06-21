package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Arapak/sio-tool/codeforces_client"
	"github.com/Arapak/sio-tool/util"

	"github.com/fatih/color"
	"github.com/k0kubun/go-ansi"
	"github.com/mitchellh/go-homedir"
)

func (c *Config) AddTemplate() (err error) {
	color.Cyan("Add a template")
	color.Cyan("Language list:")
	type kv struct {
		K, V string
	}
	var langs []kv
	for k, v := range codeforces_client.Langs {
		langs = append(langs, kv{k, v})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].V < langs[j].V })
	for _, t := range langs {
		fmt.Printf("%5v: %v\n", t.K, t.V)
	}
	color.Cyan(`Select a language (e.g. "54"): `)
	lang := ""
	for {
		lang = util.ScanlineTrim()
		if val, ok := codeforces_client.Langs[lang]; ok {
			color.Green(val)
			break
		}
		color.Red("Invalid index. Please input again")
	}

	note := `Template:
  You can insert some placeholders into your template code. When generate a code from the
  template, st will replace all placeholders by following rules:

  $%U%$   Handle (e.g. Arapak)
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)`
	_, _ = ansi.Println(note)
	color.Cyan(`Template absolute path(e.g. "~/template/io.cpp"): `)
	path := ""
	for {
		path = util.ScanlineTrim()
		path, err = homedir.Expand(path)
		if err == nil {
			if _, err := os.Stat(path); err == nil {
				break
			}
		}
		color.Red("%v is invalid. Please input again: ", path)
	}

	color.Cyan(`The suffix of template above will be added by default.`)
	color.Cyan(`Other suffix? (e.g. "cxx cc"), empty is ok: `)
	tmpSuffix := strings.Fields(util.ScanlineTrim())
	tmpSuffix = append(tmpSuffix, strings.Replace(filepath.Ext(path), ".", "", 1))
	suffixMap := map[string]bool{}
	var suffix []string
	for _, s := range tmpSuffix {
		if _, ok := suffixMap[s]; !ok {
			suffixMap[s] = true
			suffix = append(suffix, s)
		}
	}

	color.Cyan(`Template's alias (e.g. "cpp" "py"): `)
	alias := ""
	for {
		alias = util.ScanlineTrim()
		if len(alias) > 0 {
			break
		}
		color.Red("Alias can not be empty. Please input again: ")
	}

	color.Green("Script in template:")
	note = `Template will run 3 scripts in sequence when you run "st test":
    - before_script   (execute once)
    - script          (execute the number of samples times)
    - after_script    (execute once)
  You could set "before_script" or "after_script" to empty string, meaning not executing.
  You have to run your program in "script" with standard input/output (no need to redirect).

  You can insert some placeholders in your scripts. When execute a script,
  st will replace all placeholders by following rules:

  $%path%$   Path to source file (Excluding $%full%$, e.g. "/home/arapak/")
  $%full%$   Full name of source file (e.g. "a.cpp")
  $%file%$   Name of source file (Excluding suffix, e.g. "a")
  $%rand%$   Random string with 8 character (including "a-z" "0-9")`
	_, _ = ansi.Println(note)

	color.Cyan(`Before script (e.g. "g++ $%full%$ -o $%file%$.e -std=c++17"), empty is ok: `)
	beforeScript := util.ScanlineTrim()

	color.Cyan(`Script (e.g. "./$%file%$.e" "python3 $%full%$"): `)
	script := ""
	for {
		script = util.ScanlineTrim()
		if len(script) > 0 {
			break
		}
		color.Red("Script can not be empty. Please input again: ")
	}

	color.Cyan(`After script (e.g. "rm $%file%$.e"), empty is ok: `)
	afterScript := util.ScanlineTrim()

	c.Template = append(c.Template, CodeTemplate{
		alias, lang, path, suffix,
		beforeScript, script, afterScript,
	})

	if util.YesOrNo("Make it default (y/n)? ") {
		c.Default = len(c.Template) - 1
	}
	return c.save()
}

func (c *Config) RemoveTemplate() (err error) {
	color.Cyan("Remove a template")
	if len(c.Template) == 0 {
		color.Red("There is no template. Please add one")
		return nil
	}
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = color.New(color.FgGreen).Sprint("*")
		}
		_, _ = ansi.Printf(`%v%2v: "%v" "%v"`, star, i, template.Alias, template.Path)
		_, _ = ansi.Println()
	}
	idx := util.ChooseIndex(len(c.Template))
	c.Template = append(c.Template[:idx], c.Template[idx+1:]...)
	if idx == c.Default {
		c.Default = 0
	} else if idx < c.Default {
		c.Default--
	}
	return c.save()
}

func (c *Config) SetDefaultTemplate() error {
	color.Cyan("Set default template")
	if len(c.Template) == 0 {
		color.Red("There is no template. Please add one")
		return nil
	}
	for i, template := range c.Template {
		star := " "
		if i == c.Default {
			star = color.New(color.FgGreen).Sprint("*")
		}
		_, _ = ansi.Printf(`%v%2v: "%v" "%v"`, star, i, template.Alias, template.Path)
		_, _ = ansi.Println()
	}
	c.Default = util.ChooseIndex(len(c.Template))
	return c.save()
}

func (c *Config) TemplateByAlias(alias string) []CodeTemplate {
	var ret []CodeTemplate
	for _, template := range c.Template {
		if template.Alias == alias {
			ret = append(ret, template)
		}
	}
	return ret
}
