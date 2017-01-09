package main

import (
	"fmt"
	"os"

	"github.com/nethack42/gokv"
	"gopkg.in/urfave/cli.v2"
)

type Output interface {
	Node(kv.Node) []byte
}

type output struct {
	Short string
	Usage string

	Output Output
}

var defaultOutput Output

var outputs map[string]output

func registerOutput(name, short, usage string, o Output) {
	if outputs == nil {
		outputs = make(map[string]output)
	}

	if _, ok := outputs[name]; ok {
		panic("output processor with name " + name + " already registered")
	}

	outputs[name] = output{
		Short:  short,
		Usage:  usage,
		Output: o,
	}
}

func outputFlags() []cli.Flag {
	var res []cli.Flag

	for name, out := range outputs {
		res = append(res, &cli.BoolFlag{
			Name:    name,
			Aliases: []string{out.Short},
			Usage:   out.Usage,
		})
	}

	return res
}

func getOutput(c *cli.Context) (Output, error) {
	var output string

	var flags string
	for name := range outputs {
		flags += ", --" + name
	}
	flags = flags[1:]

	for name, _ := range outputs {
		if c.Bool(name) {
			if output != "" {
				return nil, fmt.Errorf("Only one of %s can be set", flags)
			}
			output = name
		}
	}

	if output == "" {
		if defaultOutput != nil {
			return defaultOutput, nil
		}
		return nil, fmt.Errorf("no output provider avaiable")
	}

	return outputs[output].Output, nil
}

func routeOutput(c *cli.Context, res interface{}) error {
	output, err := getOutput(c)
	if err != nil {
		panic(err)
		return err
	}

	var data []byte

	switch v := res.(type) {
	case kv.Node:
		data = output.Node(v)
	default:
		panic("cannot handle result data")
	}

	os.Stdout.Write(data)

	return nil
}
