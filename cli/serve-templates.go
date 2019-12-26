package cli

import (
	"fmt"
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
		Name: `js-util`,
		Markup: `
		<script>
			const message = (msg, elemId) => {
				const msgElem = document.getElementById(elemId)
				if (msgElem) { msgElem.innerHTML = msg }
			}
			const removeElem = (elemId) => {
				const elem = document.getElementById(elemId)
				if (elem) { elem.parentNode.removeChild(elem) } 
				else { console.error("Expecting element named " + elemId + ", but did not find one") }
			}
			const disableItem = (elemId) => {
				console.log(elemId)
				const elem = document.getElementById(elemId)
				if (elem) { 
					const inputs = elem.querySelectorAll('button, input') 
					if (inputs && inputs.length) {
						for (i=0; i < inputs.length; i++) { 
							if (inputs[i]) { inputs[i].remove() } 
						}
					}
					elem.style.background = 'lightpink'
					elem.disabled = true
				} 
				else { console.error("Expecting element named " + elemId + ", but did not find one") }
			}
			const post = async (url, data) => {
				const options = {
					method: 'POST',
					body: JSON.stringify(data),
					headers: { 'Content-Type': 'application/json' }
				}
				let res
				try {
					res = await fetch(url, options)
					console.log(res)
					if (res && res.ok) { 
						return await res.json() 
					} else {
						throw await res.text()	
					}
				} catch (err) {
					console.error(err)
					message(err, "message")
					throw err
				}
			}
			// post('http://localhost:8080/download?key=photo-albums.json', { p1: 1, p2: 'Hello World' }).then(res => console.log(res)).catch(err => console.log(err));
		</script>`,
	},
	Template{
		Name: `browse`,
		Markup: `
		{{ template "page-start" }}
			{{ template "js-util" }}
			<script>
				const msgElemId = 'message'
				const downloadKey = (key) => {
					post('download', { key })
						.then(res => { message(res.message, msgElemId) })
						.catch(err => { message(err, msgElemId) })
				}
				const deleteKey = (type, key) => {
					const body = { key }
					if (type == 'dir') {
						console.log('deleting directory')
						body.is_dir = true
					}
					post('delete', body)
						.then(res => {
							message(res.message, msgElemId)
							removeElem(key + '-' + type)
						})
						.catch(err => { message(err, msgElemId) })
				}
				const renameKey = (type, src_key) => {
					const elemId = src_key + '-' + type + '-input'
					const elemInput = document.getElementById(elemId);
					console.log(elemInput)
					if (elemInput) {
						const body = { src_key, dest_key: elemInput.value }
						console.log(body)
						if (type == 'dir') {
							console.log('renaming directory')
							body.is_dir = true
						}
						post('rename', body)
							.then(res => { 
								message(res.message, msgElemId) 
								disableItem(src_key + '-' + type)
							})
							.catch(err => { message(err, msgElemId) })
					}
					else { console.error("Expecting element named " + elemId + ", but did not find one") }
				}
			</script>
			<div>
				<span id="message"><span>
			</div>
			{{range .Folders}}
			<div id="{{.DisplayKey}}-dir">
				<a href="browse?dir={{.Key}}">{{.DisplayKey}}</a>
				&nbsp;<button onclick="deleteKey('dir', '{{.DisplayKey}}')">Delete</button>
				&nbsp;<button onclick="renameKey('dir', '{{.DisplayKey}}')">Rename</button>
				&nbsp;<input id="{{.DisplayKey}}-dir-input" value="{{.DisplayKey}}"></input>
			</div>
			{{end}}
			{{range .Files}}
			<div id="{{.DisplayKey}}-file">
				<p>
					{{.DisplayKey}}
					&nbsp;<button onclick="downloadKey('{{.DisplayKey}}')">Download</button>
					&nbsp;<button onclick="deleteKey('file', '{{.DisplayKey}}')">Delete</button>
					&nbsp;<button onclick="renameKey('file', '{{.DisplayKey}}')">Rename</button>
					&nbsp;<input id="{{.DisplayKey}}-file-input" value="{{.DisplayKey}}"></input>
				</p>
			</div>
			{{end}}
			{{range .Images}}
			<div id="{{.DisplayKey}}-image">
				<p>
					{{.DisplayKey}}
					&nbsp;<button onclick="downloadKey('{{.DisplayKey}}')">Download</button>
					&nbsp;<button onclick="deleteKey('image', '{{.DisplayKey}}')">Delete</button>
					&nbsp;<button onclick="renameKey('image', '{{.DisplayKey}}')">Rename</button>
					&nbsp;<input id="{{.DisplayKey}}-image-input" value="{{.DisplayKey}}"></input>
				</p>
				<img src="data:image/jpg;base64,{{.Base64}}">
			</div>
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
			return t, util.WrapError(err, funcTag, fmt.Sprintf("parsing configured template: %s", tmpl.Name))
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
