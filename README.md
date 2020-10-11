# Temporal AWS SDK Generator

| :warning: **This code is an experiment**: Absolutely no guarantee of backwards compatibility |
| --- |

This repository contains generator that generates code from template files passing
to them a list of AWS Service definitions parsed from AWS Go SDK.

## Template File Format

Templates must be in [Go template format](https://golang.org/pkg/text/template/).
The data structure passed to the templates is [AWSSDKDefinition](internal/definitions.go).
A template file must call `SetFileName` function at the beginning to specify the output file name.
It is allowed to call `SetFileName` multiple times to generate multiple files from the same template.

### Template Functions

The following functions are available inside templates:

* `SetFileName(fileName string)` sets output file name. It can contain directories. So "foo/bar/baz.go" is a valid name.
* `ToUpper(s string) string` maps all string letters to upper case
* `ToLower(s string) string` maps all string letters to lower case
* `HasPrefix(s, previx string) bool` returns true if string has prefix
* `CapitalizeFirstLetter(s string) string` capitalizes the first letter of the string
* `IsNil(value) boolean` returns if the value is nil
* `IsDuplicate(collection, value string) bool` used to deduplicate values

## Generator Options

* `template-dir` directory with template files. Only files with `.tmpl` extension are processed. Required.
* `output-dir` base directory for generated files. Required.
* `service` name of AWS service to generate code for. Optional. Default is all the services.