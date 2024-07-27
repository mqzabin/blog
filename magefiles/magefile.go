//go:build mage
// +build mage

package main

import (
	"context"

	"github.com/magefile/mage/mg"
	"github.com/mqzabin/blog/magefiles/extensions"
)

var cmdHugo = extensions.NewCommand("hugo", "--source", "./hugo")

type Prod mg.Namespace

func (Prod) Build(ctx context.Context, baseURL string) error {
	return cmdHugo.Args("--gc", "--minify", "--baseURL", baseURL).
		Env("HUGO_ENVIRONMENT", "production").
		Env("HUGO_ENV", "production").
		Run(ctx)
}

type Dev mg.Namespace

func (Dev) Server(ctx context.Context) error {
	return cmdHugo.Args("server", "-p", "1313").Run(ctx)
}

func (Dev) Tidy(ctx context.Context) error {
	return cmdHugo.Args("mod", "tidy").Run(ctx)
}
