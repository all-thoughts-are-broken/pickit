package main

import (
	"pickit/cmd"
	"pickit/internal/utils"
)

func main() {
	utils.InitLogger()
	defer utils.SyncLogger()

	cmd.Execute()
}
