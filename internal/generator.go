package internal

import (
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

const TEMPLATE_SUFFIX = ".tmpl"

func capitalizeFirstLetter(packageName string) string {
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
	deduper := &deduper{collections: make(map[string]map[string]struct{})}
	funcMap := template.FuncMap{
		"SetFileName": func(name string) (string, error) {
			err := writer.SetCurrentFile(name)
			if err != nil {
				return "", err
			}
			return "", nil
		},
		"ToUpper":               strings.ToUpper,
		"ToLower":               strings.ToLower,
		"HasPrefix":             strings.HasPrefix,
		"CapitalizeFirstLetter": capitalizeFirstLetter,
		"IsNil": func(value interface{}) bool {
			return value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
		},
		"IsDuplicate": deduper.IsDuplicate,
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
	awsSDK := &AWSSDKDefinition{Version: aws.SDKVersion, Services: definitions}
	return templates.Execute(writer, awsSDK)
}

type deduper struct {
	collections map[string]map[string]struct{}
}

// IsDuplicate returns false for the first call and then it returns true for all other calls for the
// same collection and value.
func (d *deduper) IsDuplicate(collectionName, value string) bool {
	collection, ok := d.collections[collectionName]
	if !ok {
		collection = make(map[string]struct{})
		d.collections[collectionName] = collection
	}
	_, ok = collection[value]
	if ok {
		return true
	}
	collection[value] = struct{}{}
	return false
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
