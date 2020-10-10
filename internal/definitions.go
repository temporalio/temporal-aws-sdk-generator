package internal

import (
	"fmt"
	"sort"
)

type (
	AWSSDKDefinition struct {
		Version  string
		Services []*InterfaceDefinition
	}

	InterfaceDefinition struct {
		// Service name (lower-cased, URL compatible service name)
		ID string
		// Service name (friendly name)
		Name string
		// Service methods
		Methods []*MethodDefinition
	}

	MethodDefinition struct {
		// Method name
		Name string
		// Input structure
		Input *StructDefinition
		// Output structure
		Output *StructDefinition
	}

	StructDefinition struct {
		Package string
		Name    string
		Fields  map[string]*FieldDefinition
	}

	FieldDefinition struct {
		Name string
		Type string
	}
)

func NewStructDefinition(pkg, name string) *StructDefinition {
	return &StructDefinition{Package: pkg, Name: name, Fields: make(map[string]*FieldDefinition)}
}

func (m *MethodDefinition) String() string {
	return fmt.Sprintf("%v(%v) %v", m.Name, m.Input, m.Output)
}

func (i InterfaceDefinition) Imports() []string {
	packages := make(map[string]bool)
	for _, method := range i.Methods {
		packages[method.Input.Package] = true
		if method.Output != nil {
			packages[method.Output.Package] = true
		}
	}
	delete(packages, "")
	result := make([]string, 0, len(packages))
	for p := range packages {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

func (i InterfaceDefinition) String() string {
	result := i.Name + ": "
	for i, method := range i.Methods {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%v", method.Name)
	}
	return result
}
