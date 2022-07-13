package cli

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

// cobraUsageTemplate is cobra.Command.UsageTemplate() with flags extracted
const cobraUsageTemplate = `Usage:{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
{{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
{{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
{{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

%s{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
{{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// flagTemplate describes the flag
const flagTemplate = "  --{{.Name}}{{if not .IsBoolFlag}}=<value>{{end}}\n{{.Description}}\n"

// UsageTmpl returns a cobra-compatible usage template that will be printed
// during the help output.
// This template prints help like:
//   --name=<value>
//    <description>
// We use it over the default template so that the output it easier to read.
func UsageTmpl(flags map[string]Flag) string {
	type _flag struct {
		Name        string
		Description string
		IsBoolFlag  bool
	}

	var fs []_flag
	for name, f := range flags {
		if f.IsHidden() {
			continue
		}

		desc := to80CharCols(f.Format())
		_, isBool := f.(*BoolFlag)

		fs = append(fs, _flag{
			Name:        name,
			Description: desc,
			IsBoolFlag:  isBool,
		})
	}

	sort.Slice(fs, func(i, j int) bool {
		return fs[i].Name < fs[j].Name
	})

	tmpl := template.Must(template.New("").Parse(flagTemplate))
	var flagHelpOutput string
	for _, f := range fs {
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, f); err != nil {
			panic(err)
		}
		flagHelpOutput += buf.String()
	}

	return fmt.Sprintf(cobraUsageTemplate, flagHelpOutput)
}

func to80CharCols(s string) string {
	var splitAt80 string

	splitSpaces := strings.Split(s, " ")

	var nextLine string
	for i, spaceSplit := range splitSpaces {
		if len(nextLine)+len(spaceSplit)+1 > 80 {
			splitAt80 += fmt.Sprintf("      %s\n", strings.TrimSuffix(nextLine, " "))
			nextLine = ""
		}

		if i == len(splitSpaces)-1 {
			nextLine += spaceSplit + " "
			splitAt80 += fmt.Sprintf("      %s\n", strings.TrimSuffix(nextLine, " "))
			break
		}

		nextLine += spaceSplit + " "
	}

	return splitAt80
}
