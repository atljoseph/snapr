package cli

import (
	"snapr/util"
	"text/template"
)

// Template descrbes all of our in memory templates
type Template struct {
	Name   string
	Markup string
}

// this holds all the templates after parsing
var serveCmdTempl *template.Template

// we define our templates here
var serveCmdTmpls = []Template{
	Template{
		Name: `page-start`,
		Markup: `<html>
		<head>
			<link rel="icon" href="data:,">
		</head>
		<body>
		`,
	},
	Template{
		Name: `page-end`,
		Markup: `
		</body></html>`,
	},
	Template{
		Name: `browse`,
		Markup: `
		{{ template "page-start" }}
			{{range .Folders}}
			<p>
				<a href="browse?dir={{.Key}}">{{.DisplayKey}}</a>
			</p>
			{{end}}
			{{range .Files}}
			<p>
				<p>
					{{.DisplayKey}}
					&nbsp;<a href="download?key={{.DisplayKey}}">Download</a>
				</p>
			</p>
			{{end}}
			{{range .Images}}
			<span>
				<p>
					{{.DisplayKey}}
					&nbsp;<a href="download?key={{.DisplayKey}}">Download</a>
				</p>
				<img src="data:image/jpg;base64,{{.Base64}}">
			</span>
			{{end}}
		{{ template "page-end" }}`,
	},
	Template{
		Name: `download`,
		Markup: `
		{{ template "page-start" }}
			<span>
				<p>{{.Message}}</p>
			</span>
		{{ template "page-end" }}`,
	},
}

// ParseTemplates parses and gets the templates for the server
func ParseTemplates() (*template.Template, error) {
	funcTag := "ParseTemplates"

	// parse a dumy template to get a *template.Template object
	t, err := template.New("dummy-template").Parse("not used anywhere")
	if err != nil {
		return t, util.WrapError(err, funcTag, "parsing initial template")
	}

	// add the ones we have defined
	for _, tmpl := range serveCmdTmpls {
		t, err = t.New(tmpl.Name).Parse(tmpl.Markup)
		if err != nil {
			return t, util.WrapError(err, funcTag, "parsing configured template")
		}
	}
	return t, nil
}

// // ParseTemplateFiles parses and gets the templates for the server
// func ParseTemplateFiles(templateDir string, templateFiles []string) (*template.Template, error) {
// 	funcTag := "ParseTemplates"

// 	// get abs path
// 	absPath, err := filepath.Abs(templateDir)
// 	if err != nil {
// 		return nil, util.WrapError(err, funcTag, "getting absolute template path")
// 	}

// 	// add the dir to all the files
// 	var filesWithDir []string
// 	for _, templateFile := range templateFiles {
// 		// logrus.Infof("%s, %s", absPath, templateFile)
// 		filesWithDir = append(filesWithDir, filepath.Join(absPath, templateFile))
// 	}

// 	// parse templates
// 	templateObj, err := template.ParseFiles(filesWithDir...)
// 	if err != nil {
// 		return nil, util.WrapError(err, funcTag, "parse files")
// 	}

// 	return templateObj, nil
// }
