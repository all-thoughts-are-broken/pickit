package cmd

import (
	"log"
	"pickit/internal/mode"
)

import (
	"github.com/spf13/cobra"
)

type downloadFlags struct {
	cdn         string // 图片域名地址
	output      string // 输出路径
	aid         int    // 车牌号
	count       int    // 图片数量
	concurrency int    // 并发数
	proxy       string // 魔法
}

var downloadOpts downloadFlags

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "下载图片",
	Run: func(cmd *cobra.Command, args []string) {
		// 下载
		mode.DownloadAlbum(downloadOpts.cdn, downloadOpts.output, downloadOpts.proxy, downloadOpts.aid, downloadOpts.count, downloadOpts.concurrency)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// 本地标志
	downloadCmd.Flags().StringVarP(&downloadOpts.cdn, "cdn", "u", "", "图片 cdn 域名（必传）")
	downloadCmd.Flags().StringVarP(&downloadOpts.output, "output", "o", "", "保存文件夹路径（必传）")
	downloadCmd.Flags().IntVarP(&downloadOpts.aid, "aid", "a", 0, "车牌号（必传）")
	downloadCmd.Flags().IntVarP(&downloadOpts.count, "count", "n", 0, "图片数量（必传）")
	downloadCmd.Flags().IntVarP(&downloadOpts.concurrency, "concurrency", "c", 8, "并发数（可选）")
	downloadCmd.Flags().StringVarP(&downloadOpts.proxy, "proxy", "p", "", "魔法（可选）")

	// cdn、output、aid、count 这四个是必传的
	requiredFlags := []string{"cdn", "output", "aid", "count"}
	for _, flag := range requiredFlags {
		if err := downloadCmd.MarkFlagRequired(flag); err != nil {
			log.Fatalf("初始化失败: 无法标记 %s 为必需参数: %v", flag, err)
		}
	}
}
