package scaffold

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	txttmpl "text/template"

	"github.com/mykelswitzer/progenitor/internal/filesys"
	rp "github.com/mykelswitzer/progenitor/internal/repo"
	"github.com/mykelswitzer/progenitor/pkg/config"
	str "github.com/mykelswitzer/progenitor/pkg/strings"
	"github.com/pkg/errors"
	"github.com/posener/gitfs"
	"github.com/posener/gitfs/fsutil"
	"github.com/spf13/afero"
)

type TemplateData interface {
	Init(config *config.Config) TemplateData
}

const TMPLSFX string = ".tmpl"

func trimSuffix(path string) string {
	return strings.TrimSuffix(path, TMPLSFX)
}

func contains(a []string, x string) bool {
	if len(a) == 0 {
		return false
	}

	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func templateFunctions() txttmpl.FuncMap {
	return txttmpl.FuncMap{
		"tolower":     strings.ToLower,
		"tocamel":     str.ToCamel,
		"topascal":    str.ToPascal,
		"toplural":    str.ToPlural,
		"topackage":   str.ToPackage,
		"tosnakecase": str.ToSnakeCase,
		"toupper":     strings.ToUpper,
	}
}

func getScaffoldTemplatePath(projectType string, withVersion bool) string {

	var (
		repoName string = "progenitor-tmpl-" + projectType
		path     string = "github.com/mykelswitzer/" + repoName + "/template"
		version  string
	)

	if withVersion {
		if mf, ok := debug.ReadBuildInfo(); ok {
			for _, m := range mf.Deps {
				if strings.HasSuffix(m.Path, repoName) {
					version = m.Version
					break
				}
			}
		}
		path += "@tags/" + version
	}

	log.Println("reading scaffolding template files from:" + path)

	return path

}

func getLatestTemplates(token string, templatePath string, skipTemplates []string, basePath afero.Fs) (map[string]*txttmpl.Template, error) {

	var templates = map[string]*txttmpl.Template{}

	ctx := context.Background()
	oauth := rp.GithubAuth(token, ctx)
	fs, err := gitfs.New(ctx,
		templatePath,
		gitfs.OptClient(oauth),
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed initializing git filesystem")
	}

	walker := fsutil.Walk(fs, "")
	for walker.Step() {

		if err := walker.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		if !walker.Stat().IsDir() {
			// get && check path
			//dirpath := filepath.Dir(walker.Path())
			file := walker.Path()
			ex, err := filesys.DirExists(filepath.Dir(file), basePath)
			if err != nil {
				return nil, errors.Wrap(err, "Failed reading base path while parsing templates")
			}
			// if the path exists, parse the templates
			if ex && contains(skipTemplates, file) == false {
				log.Println("Fetching template: ", trimSuffix(file))

				tmpl, err := filesys.TmplParse(fs, templateFunctions(), nil, file)
				if err != nil {
					werr := errors.Wrapf(err, "Unable to parse template %s", file)
					log.Println(werr)
				}
				templates[file] = tmpl
			}
		}
	}

	return templates, nil
}
