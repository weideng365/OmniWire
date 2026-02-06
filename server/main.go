package main

import (
	_ "omniwire/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"omniwire/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
