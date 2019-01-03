/**
 *  MIT License
 *
 *  Copyright (c) 2018-2018 Ji-Young Park(jiyoung.park.dev@gmail.com)
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *      The above copyright notice and this permission notice shall be included in all
 *      copies or substantial portions of the Software.
 *
 *      THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *      IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *      FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *      AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *      LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *      OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *      SOFTWARE.
*/

package main

import (
	"os"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	git "gopkg.in/src-d/go-git.v4"

	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/jparklab/gitcui/ui"
)

type RunOptions struct {
	URL string
	Path string
	DoClone bool
}

func (o *RunOptions) addFlags(cmd *cobra.Command, flags *pflag.FlagSet) {
	flags.BoolVar(&o.DoClone, "clone", false, "If specified, clone from the given url")
}

func main() {
	runOptions := RunOptions{}

	rootCmd := &cobra.Command{
		Use: "gitcui <url or path>",
		Long: "CLI to play with git library",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			path := args[0]

			var repo *git.Repository
			var err error

			if runOptions.DoClone {
				log.Printf("Clone %s\n", path)

				// Git object storer
				storer := memory.NewStorage()
				// clone repository into memory
				repo, err = git.Clone(storer, nil, &git.CloneOptions{
					URL: path,
				})
				if err != nil {
					log.Printf("Failed to clone: %v\n", err)
					os.Exit(1)
				}
			} else {
				log.Printf("Open %s\n", path)

				repo, err = git.PlainOpen(path)
				if err != nil {
					log.Printf("Failed to open: %v\n", err)
					os.Exit(1)
				}
			}

			ui.Run(repo)
		},
	}

	runOptions.addFlags(rootCmd, rootCmd.PersistentFlags())

	if err := rootCmd.Execute(); err != nil {
		log.Printf("%v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
