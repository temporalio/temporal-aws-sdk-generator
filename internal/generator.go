package internal

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

const TEMPLATE_SUFFIX = ".tmpl"

func toPrefix(packageName string) string {
	if packageName == "" {
		return packageName
	}
	return strings.ToUpper(packageName[0:1]) + packageName[1:]
}

type TemporalAWSGenerator struct {
	TemplateDir string
	// Some of the output structures are duplicated.
	// Used to to dedupe types definitions based on them.
	outputStructs map[string]bool
}

func NewGenerator(templateDir string) *TemporalAWSGenerator {
	return &TemporalAWSGenerator{TemplateDir: templateDir, outputStructs: make(map[string]bool)}
}

func (g *TemporalAWSGenerator) GenerateCode(outputDir string, definitions []*InterfaceDefinition) error {
	var templateFiles []string
	files, err := ioutil.ReadDir(g.TemplateDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !f.IsDir() {
			templateName := f.Name()
			if strings.HasSuffix(templateName, TEMPLATE_SUFFIX) {
				templateFiles = append(templateFiles, strings.TrimSuffix(templateName, TEMPLATE_SUFFIX))
			}
		}
	}
	for _, templateFile := range templateFiles {
		err := g.generateFromSingleTemplate(templateFile+".tmpl", outputDir, definitions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *TemporalAWSGenerator) generateFromSingleTemplate(templateFile string, outputDir string, definitions []*InterfaceDefinition) error {
	writer := &MultiFileWriter{outputDir: outputDir}
	defer writer.Close()
	funcMap := template.FuncMap{
		"SetFileName": func(name string) (string, error) {
			err := writer.SetCurrentFile(name)
			if err != nil {
				return "", err
			}
			return "", nil
		},
		"ToUpper":   strings.ToUpper,
		"ToLower":   strings.ToLower,
		"HasPrefix": strings.HasPrefix,
		"ToPrefix":  toPrefix,
		"IsNil": func(value interface{}) bool {
			return value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
		},
	}

	templates, err := template.New(templateFile).Funcs(funcMap).ParseFiles(g.TemplateDir + "/" + templateFile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0700)
		if err != nil {
			return err
		}
	}
	return templates.Execute(writer, definitions)
}

type MultiFileWriter struct {
	outputDir   string
	currentFile *os.File
}

func (m *MultiFileWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 || m.currentFile == nil {
		return len(p), nil
	}
	return m.currentFile.Write(p)
}

func (m *MultiFileWriter) SetCurrentFile(name string) error {
	outputFile := m.outputDir + "/" + name
	if err := os.MkdirAll(filepath.Dir(outputFile), 0770); err != nil {
		return err
	}
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	if m.currentFile != nil {
		err = m.currentFile.Close()
		if err != nil {
			return err
		}
	}
	m.currentFile = f
	return nil
}

func (m *MultiFileWriter) Close() {
	if m.currentFile != nil {
		err := m.currentFile.Close()
		if err != nil {
			log.Printf("Failure closing file: %v", err)
		}
	}
}
