package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/urfave/cli/v2"
)

var isDebug bool
var repositoryVariableLookup = map[string]string{}

func readInputFile(inputFileName string) (*hclwrite.File, error) {
	fmt.Printf("Reading configuration from file\n")

	inputSource, err := os.ReadFile(inputFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}
	fmt.Printf("TF configuration read from %s\n", inputFileName)

	tfFile, diags := hclwrite.ParseConfig(inputSource, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags)
	}

	if isDebug {
		fmt.Printf("Input configuration: %v\n", string(tfFile.Bytes()))
	}

	return tfFile, nil
}

func getAttributeValueAsString(attr *hclwrite.Attribute) (string, error) {
	tokens := attr.Expr().BuildTokens(nil)

	for _, token := range tokens {
		if isDebug {
			fmt.Printf("token: %v\n", token)
		}

		if token.Type == hclsyntax.TokenQuotedLit {
			return string(token.Bytes), nil
		}
	}

	return "", fmt.Errorf("No string token found")
}

func migrateConfiguration(file *hclwrite.File) error {
	fmt.Printf("Renaming resource type\n")

	// loop through every block to replace generic resource type with package specific resource type
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "resource" {
			fmt.Printf("resource type: %s\n", strings.Join(labels, "."))
			resourceType := labels[0]

			re := regexp.MustCompile(`artifactory_(local|remote|virtual)_repository`)
			matches := re.FindStringSubmatch(resourceType)
			if len(matches) > 1 {
				repoType := matches[1]

				packageType, err := getAttributeValueAsString(block.Body().GetAttribute("package_type"))
				if err != nil {
					fmt.Printf("Unable to find attribute 'package_type'. Skipping...\n")
					continue
				}
				newResourceType := fmt.Sprintf("artifactory_%s_%s_repository", repoType, packageType)
				block.SetLabels([]string{newResourceType, labels[1]})
				block.Body().RemoveAttribute("package_type")

				// Insert old and new resource type into lookup map
				key, err := getAttributeValueAsString(block.Body().GetAttribute("key"))
				repositoryVariableLookup[strings.Join(labels, ".")] = strings.Join([]string{newResourceType, labels[1], key}, ".")
			}
		}
	}

	if isDebug {
		fmt.Printf("repositoryVariableLookup: %v\n", repositoryVariableLookup)
	}

	// loop through everything to rename attribute prefix
	fmt.Printf("Renaming variable references\n")
	for _, block := range file.Body().Blocks() {
		for name, attr := range block.Body().Attributes() {
			if isDebug {
				tokens := attr.Expr().BuildTokens(nil)
				fmt.Printf("Attribute name: %s, value: %s\n", name, string(tokens.Bytes()))
			}

			for old, new := range repositoryVariableLookup {
				attr.Expr().RenameVariablePrefix(strings.Split(old, "."), strings.Split(new, ".")[:2])
			}
		}
	}

	return nil
}

func writeOutputFile(outputFileName string, tfFile *hclwrite.File) error {
	fmt.Printf("Writing configuration to file\n")

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return err
	}

	if isDebug {
		fmt.Printf("Output configuration: %v\n", string(tfFile.Bytes()))
	}

	outputFile.Write(tfFile.Bytes())
	fmt.Printf("Configuration saved to %s", outputFileName)
	return nil
}

func main() {
	var inputFileName string
	var outputFileName string
	var outputImport bool

	app := &cli.App{
		Name:  "v5-v6-migrator",
		Usage: "Artifactory Terraform V5-V6 HCL migrator - Migrate generic repository resources to package specific repository resources",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Alex Hung",
				Email: "alexh@jfrog.com",
			},
		},
		Version:              "0.1.0",
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Value:       false,
				Destination: &isDebug,
			},
			&cli.StringFlag{
				Name:        "input",
				Usage:       ".tf `FILE` to migrate",
				Aliases:     []string{"i"},
				Destination: &inputFileName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "output",
				Usage:       "Output .tf `FILE`",
				Aliases:     []string{"o"},
				Destination: &outputFileName,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "import",
				Usage:       "Output TF import statements",
				Value:       false,
				Destination: &outputImport,
			},
		},
		Action: func(ctx *cli.Context) error {
			if isDebug {
				fmt.Printf("Input: %s\n", inputFileName)
				fmt.Printf("Output: %s\n", outputFileName)
			}

			tfFile, err := readInputFile(inputFileName)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to read configuration: %s", err), 1)
			}

			err = migrateConfiguration(tfFile)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to migrate configuration: %s", err), 1)
			}

			err = writeOutputFile(outputFileName, tfFile)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to write configuration: %s", err), 2)
			}

			if outputImport {
				for _, new := range repositoryVariableLookup {
					components := strings.Split(new, ".")
					fmt.Printf("terraform import %s.%s %s\n", components[0], components[1], components[2])
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
