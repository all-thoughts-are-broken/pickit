package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "pickit",
	Short: "pickit 是一个命令行工具，提供了图片下载、还原、合成 PDF 的一些功能。",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
