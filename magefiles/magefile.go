//go:build mage
// +build mage

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/mqzabin/blog/magefiles/extensions"
)

var cmdHugo = extensions.NewCommand("hugo", "--source", "./hugo")

type Setup mg.Namespace

func (Setup) LinkContent() error {
	const (
		symlinkPath                   = "./hugo/content"
		relativeOriginalDirectoryPath = "../blog-content"
	)

	fmt.Println("Ensuring Hugo site content symlink")
	fmt.Printf("Creating a symlink at %q pointing to %q...\n", symlinkPath, relativeOriginalDirectoryPath)

	if err := extensions.EnsureSymlink(symlinkPath, relativeOriginalDirectoryPath); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}

		fmt.Printf("Content symbolic link already exists at %q, skipping creation...\n", symlinkPath)

		return nil
	}

	fmt.Println("Content symbolic link successfully created!")

	return nil
}

type Prod mg.Namespace

func (Prod) Build(ctx context.Context, baseURL string) error {
	mg.Deps(Setup{}.LinkContent)

	fmt.Printf("Running production build process for base URL %q...\n", baseURL)

	return cmdHugo.Args("--gc", "--minify", "--baseURL", baseURL).
		Env("HUGO_ENVIRONMENT", "production").
		Env("HUGO_ENV", "production").
		Run(ctx)
}

type Dev mg.Namespace

func (Dev) Server(ctx context.Context) error {
	mg.Deps(Setup{}.LinkContent)

	const port = "1313"

	fmt.Printf("Running development server on http://localhost:%s\n", port)

	return cmdHugo.Args("server", "-p", "1313").Run(ctx)
}

func (Dev) Tidy(ctx context.Context) error {
	return cmdHugo.Args("mod", "tidy").Run(ctx)
}
