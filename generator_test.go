package main

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

const testTemplate = `
`

const expected = `
`

func TestGenerator(t *testing.T) {
	// Create temporary directory with a template file in it
	fs := afero.NewOsFs()
	tempDir, err := afero.TempDir(fs, "", "io.temporal.generator")
	if err != nil {
		t.Fatal("Failed to create the temporary directory", tempDir, err)
	}
	defer fs.RemoveAll(tempDir)
	templateFileName := tempDir + "/test.tmpl"
	templateFile, err := fs.Create(templateFileName)
	if err != nil {
		t.Fatal("Failed to create the template file", templateFileName, err)
	}
	defer templateFile.Close()
	_, err = templateFile.WriteString(testTemplate)
	if err != nil {
		t.Fatal("Failed to write the template file", templateFileName, err)
	}

	// Run the generator
	args := []string{
		os.Args[0], // program name
		"--template-dir=" + tempDir,
		"--output-dir=" + tempDir,
		"--service=s3",
	}
	run(args)

	// Compare output
	generatedFileName := tempDir + "/s3_test.go"
	generated, err := ioutil.ReadFile(generatedFileName)
	if err != nil {
		t.Fatal("Failed to open the generated file", generatedFileName, err)
	}
	assert.Equal(t, expected, generated)
}
