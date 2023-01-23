/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kh4n/rpcgen/parsergenerator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates Go and JS code from rpcgen.yml",
	Run: func(cmd *cobra.Command, args []string) {
		optsBytes, err := os.ReadFile("rpcgen.yml")
		if err != nil {
			log.Fatalln("unable to read rpcgen.yml:", err)
		}
		opts := parsergenerator.Options{}
		err = yaml.Unmarshal(optsBytes, &opts)
		if err != nil {
			log.Fatalln("unable to parse:", err)
		}
		defs := make([]parsergenerator.GenDef, 0)
		filepath.Walk(opts.Folder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, ".yml") {
				return nil
			}
			def, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			defs = append(defs, parsergenerator.GenDef{})
			err = yaml.Unmarshal(def, &defs[len(defs)-1])
			if err != nil {
				return err
			}
			return nil
		})
		goGen := parsergenerator.GoGenerator{}
		ret, err := parsergenerator.ParseAll(&opts, defs, &goGen)
		if err != nil {
			log.Fatalln("unable to parse:", err)
		}
		outPath := path.Join(opts.OutputFolder, opts.OutputFolder+".go")
		fmt.Println("writing to:", outPath)
		err = os.WriteFile(outPath, []byte(*ret), 0644)
		if err != nil {
			log.Fatalln("unable to write file", outPath, ":", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
