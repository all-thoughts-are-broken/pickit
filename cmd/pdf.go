package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"pickit/internal/mode"
)

type pdfFlags struct {
	input    string // 输入路径
	output   string // 输出路径
	password string // 密码
}

var pdfOpts pdfFlags

var cmdPdf = &cobra.Command{
	Use:   "pdf",
	Short: "合成 PDF",
	Run: func(cmd *cobra.Command, args []string) {
		mode.CreatePDF(pdfOpts.input, pdfOpts.output, pdfOpts.password)
	},
}

func init() {
	rootCmd.AddCommand(cmdPdf)

	// 本地标志
	cmdPdf.Flags().StringVarP(&pdfOpts.input, "input", "i", "", "需要还原的图片文件夹路径（必传）")
	cmdPdf.Flags().StringVarP(&pdfOpts.output, "output", "o", "", "还原后的图片输出文件夹路径（必传）")
	cmdPdf.Flags().StringVarP(&pdfOpts.password, "password", "p", "", "密码（可选）")

	// input、output 这两个是必传的
	requiredFlags := []string{"input", "output"}
	for _, flag := range requiredFlags {
		if err := cmdPdf.MarkFlagRequired(flag); err != nil {
			log.Fatalf("初始化失败: 无法标记 %s 为必需参数: %v", flag, err)
		}
	}
}
