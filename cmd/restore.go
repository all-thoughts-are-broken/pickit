package cmd

import (
	"log"
	"pickit/internal/mode"
)

import (
	"github.com/spf13/cobra"
)

type restoreFlags struct {
	input       string // 输入路径
	output      string // 输出路径
	aid         int    // 车牌号
	concurrency int    // 并发数
}

var restoreOpts restoreFlags

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "还原图片",
	Run: func(cmd *cobra.Command, args []string) {
		// 还原图片
		mode.RestoreImages(restoreOpts.input, restoreOpts.output, restoreOpts.aid, restoreOpts.concurrency)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// 本地标志
	restoreCmd.Flags().StringVarP(&restoreOpts.input, "input", "i", "", "需要还原的图片文件夹路径（必传）")
	restoreCmd.Flags().StringVarP(&restoreOpts.output, "output", "o", "", "还原后的图片输出文件夹路径（必传）")
	restoreCmd.Flags().IntVarP(&restoreOpts.aid, "aid", "a", 0, "车牌号（必传）")
	restoreCmd.Flags().IntVarP(&restoreOpts.concurrency, "concurrency", "c", 8, "并发数（可选）")

	// input、output、aid 这三个是必传的
	requiredFlags := []string{"input", "output", "aid"}
	for _, flag := range requiredFlags {
		if err := restoreCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("初始化失败: 无法标记 %s 为必需参数: %v", flag, err)
		}
	}
}
