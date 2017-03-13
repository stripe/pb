package diff

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Changing a protofile Name should be fine. The package name is never determined
// by the filename.
// Backwards incompatible changes:
// - Removing a field
// - Changing the type of a field
// - Removing a RPC endpoint
// - Removing an Enum type
// - Changing the input or output message type
// - Nesting / Unnesting a message or enum type
// - Looking at options is important too

// Things that would require code changes
// - What if they change the java package name?
// - Renaming a field? (if using the JSON output)
// - Renaming an enum field?
// - Marking a field as repeated

// There are two types of changes: ones that will break existing clients, and
// ones that will require new code changes

type Problem interface {
	String() string
}

type Report struct {
	Problems []Problem
}

func (r *Report) Add(prob Problem) {
	r.Problems = append(r.Problems, prob)
}

func Diff(previous, current *plugin.CodeGeneratorRequest) (*Report, error) {
	curr := map[string]*descriptor.FileDescriptorProto{}
	report := &Report{Problems: []Problem{}}

	for _, protoFile := range current.ProtoFile {
		curr[*protoFile.Name] = protoFile
	}

	for _, protoFile := range previous.ProtoFile {
		next, exists := curr[*protoFile.Name]
		if !exists {
			report.Add(ProblemRemovedFile{*protoFile.Name})
			continue
		}
		diffFile(report, protoFile, next)
	}

	var err error
	if len(report.Problems) > 0 {
		err = fmt.Errorf("found %d problems: %s", len(report.Problems), report.Problems)
	}

	return report, err
}

func diffFile(report *Report, previous, current *descriptor.FileDescriptorProto) {
	{ // EnumType
		curr := map[string]*descriptor.EnumDescriptorProto{}
		for _, enum := range current.EnumType {
			curr[*enum.Name] = enum
		}
		for _, enum := range previous.EnumType {
			next, exists := curr[*enum.Name]
			if !exists {
				report.Add(ProblemRemovedEnum{*enum.Name})
				continue
			}
			diffEnum(report, enum, next)
		}
	}

	{ // Service
	}

	{ // MessageType
		curr := map[string]*descriptor.DescriptorProto{}
		for _, msg := range current.MessageType {
			curr[*msg.Name] = msg
		}
		for _, msg := range previous.MessageType {
			next, exists := curr[*msg.Name]
			if !exists {
				report.Add(ProblemRemovedMessage{*msg.Name})
				continue
			}
			diffMsg(report, msg, next)
		}
	}
}

func diffMsg(report *Report, previous, current *descriptor.DescriptorProto) {
	curr := map[int32]*descriptor.FieldDescriptorProto{}

	for _, field := range current.Field {
		curr[*field.Number] = field
	}

	for _, field := range previous.Field {
		next, exists := curr[*field.Number]
		if !exists {
			report.Add(ProblemRemovedField{*field.Name})
			continue
		}
		if *field.Type != *next.Type {
			report.Add(ProblemChangedFieldType{
				Field:   *field.Name,
				OldType: field.Type,
				NewType: next.Type,
			})
		}
	}
}

func diffEnum(report *Report, previous, current *descriptor.EnumDescriptorProto) {
}