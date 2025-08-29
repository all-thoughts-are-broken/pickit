package utils

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"os"
	"path/filepath"
)

type imageInfo struct {
	width, height float64
	path          string
}

func ConvertImagesToPDF(dir, output, password string) error {
	// 记录转换开始
	LogInfo("开始图像转PDF转换",
		Str("input_dir", dir),
		Str("output_file", output),
	)

	files, err := GetDirInfo(dir)
	if err != nil {
		LogError("获取目录信息失败", Err(err))
		return err
	}

	LogDebug("目录扫描完成",
		Int("目录数量", len(files)),
		Int("总文件数", totalFileCount(files)),
	)

	// 确保输出目录存在
	if err := ensureOutputDir(output); err != nil {
		LogError("创建输出目录失败", Err(err))
		return err
	}

	// 单层目录处理
	if len(files) == 1 {
		LogDebug("处理单层目录结构")
		info := files[0]
		imageList := info.Files

		LogDebug("目录内容",
			Str("目录名", info.Name),
			Int("图片数量", len(imageList)),
		)

		pdf := gofpdf.NewCustom(&gofpdf.InitType{
			UnitStr: "pt",
		})

		// 收集图片信息
		imageInfos := make([]imageInfo, 0, len(imageList))
		for _, file := range imageList {
			imgInfo := pdf.RegisterImage(file, "")
			if imgInfo == nil {
				LogWarn("图片注册失败，可能不是有效图像文件", Str("file", file))
				continue
			}
			imageInfos = append(imageInfos, imageInfo{
				width:  imgInfo.Width(),
				height: imgInfo.Height(),
				path:   file,
			})
		}

		LogDebug("图片信息收集完成", Int("有效图片数", len(imageInfos)))

		if len(imageInfos) == 0 {
			LogError("目录中没有有效的图片文件", Str("dir", info.Name))
			return fmt.Errorf("目录中没有有效的图片文件: %s", info.Name)
		}

		// 计算最终尺寸
		imgWs := make([]float64, len(imageInfos))
		for i, imgInfo := range imageInfos {
			imgWs[i] = imgInfo.width
		}
		finalWidth := mostFrequent(imgWs)
		adjustedImages := adjustImagesToWidth(imageInfos, finalWidth)

		var finalHeight float64
		for _, img := range adjustedImages {
			finalHeight += img.height
		}

		LogDebug("页面尺寸计算完成",
			Float64("width", finalWidth),
			Float64("height", finalHeight),
		)

		// 创建页面
		pdf.AddPageFormat("", gofpdf.SizeType{
			Wd: finalWidth,
			Ht: finalHeight,
		})

		// 添加图片
		var currentY float64
		for i, img := range adjustedImages {
			LogDebug("添加图片到PDF",
				Str("file", img.path),
				Int("index", i),
				Float64("width", img.width),
				Float64("height", img.height),
			)

			// 添加书签
			pdf.Bookmark(fmt.Sprintf("Page %d", i+1), 0, currentY)

			pdf.ImageOptions(
				img.path,
				0,          // x坐标
				currentY,   // y坐标
				img.width,  // 宽度
				img.height, // 高度
				false,      // 不使用流
				gofpdf.ImageOptions{
					ImageType: "JPEG",
				},
				0,
				"",
			)
			currentY += img.height
		}

		// 设置基础加密
		if password != "" {
			LogInfo("设置PDF密码保护")
			pdf.SetProtection(gofpdf.CnProtectPrint, password, password)
		}

		// 生成PDF
		LogInfo("正在生成PDF文件", Str("output", output))
		if err := pdf.OutputFileAndClose(output); err != nil {
			LogError("PDF生成失败", Err(err))
			return fmt.Errorf("合成 PDF 失败: %s", err)
		}

		LogInfo("PDF生成成功",
			Str("output", output),
			Int("图片数量", len(adjustedImages)),
			Float64("文件大小(MB)", getFileSizeMB(output)),
		)
		return nil
	} else {
		// 多层目录处理
		LogDebug("处理多层目录结构", Int("章节数", len(files)))
		pdf := gofpdf.NewCustom(&gofpdf.InitType{
			UnitStr: "pt",
		})

		for chapterIdx, chapter := range files {
			LogDebug("处理章节",
				Int("chapter", chapterIdx+1),
				Str("name", chapter.Name),
				Int("图片数量", len(chapter.Files)),
			)

			imageInfos := make([]imageInfo, 0, len(chapter.Files))
			for _, file := range chapter.Files {
				imgInfo := pdf.RegisterImage(file, "")
				if imgInfo == nil {
					LogWarn("图片注册失败，跳过", Str("file", file))
					continue
				}
				imageInfos = append(imageInfos, imageInfo{
					width:  imgInfo.Width(),
					height: imgInfo.Height(),
					path:   file,
				})
			}

			if len(imageInfos) == 0 {
				LogWarn("章节中没有有效图片，跳过",
					Int("chapter", chapterIdx+1),
					Str("name", chapter.Name),
				)
				continue
			}

			// 计算尺寸
			imgWs := make([]float64, len(imageInfos))
			for i, imgInfo := range imageInfos {
				imgWs[i] = imgInfo.width
			}
			finalWidth := mostFrequent(imgWs)
			adjustedImages := adjustImagesToWidth(imageInfos, finalWidth)

			var finalHeight float64
			for _, img := range adjustedImages {
				finalHeight += img.height
			}

			LogDebug("章节页面尺寸计算完成",
				Int("chapter", chapterIdx+1),
				Float64("width", finalWidth),
				Float64("height", finalHeight),
			)

			// 创建页面
			pdf.AddPageFormat("", gofpdf.SizeType{
				Wd: finalWidth,
				Ht: finalHeight,
			})

			// 添加章节书签
			pdf.Bookmark(fmt.Sprintf("Chapter %d", chapterIdx+1), 0, 0)

			// 添加图片
			var currentY float64
			for pageIdx, img := range adjustedImages {
				LogDebug("添加图片到章节",
					Int("chapter", chapterIdx+1),
					Int("page", pageIdx+1),
					Str("file", img.path),
				)

				pdf.Bookmark(fmt.Sprintf("Page %d", pageIdx+1), 1, currentY)
				pdf.ImageOptions(
					img.path,
					0,          // x坐标
					currentY,   // y坐标
					img.width,  // 宽度
					img.height, // 高度
					false,      // 不使用流
					gofpdf.ImageOptions{
						ImageType: "JPEG",
					},
					0,
					"",
				)
				currentY += img.height
			}
		}
		// 设置基础加密
		if password != "" {
			LogInfo("设置PDF密码保护")
			pdf.SetProtection(gofpdf.CnProtectPrint, password, password)
		}
		// 生成PDF
		LogInfo("正在生成多章节PDF文件", Str("output", output))
		if err := pdf.OutputFileAndClose(output); err != nil {
			LogError("PDF生成失败", Err(err))
			return fmt.Errorf("合成 PDF 失败: %s", err)
		}

		LogInfo("多章节PDF生成成功",
			Str("output", output),
			Int("章节数", len(files)),
			Int("总图片数", totalFileCount(files)),
			Float64("文件大小(MB)", getFileSizeMB(output)),
		)
		return nil
	}
}

// 辅助函数 =========================================

// 确保输出目录存在
func ensureOutputDir(output string) error {
	dir := filepath.Dir(output)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		LogDebug("创建输出目录", Str("path", dir))
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// 获取文件大小(MB)
func getFileSizeMB(path string) float64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return float64(info.Size()) / (1024 * 1024)
}

// 计算总文件数
func totalFileCount(files []DirInfo) int {
	total := 0
	for _, f := range files {
		total += len(f.Files)
	}
	return total
}

func mostFrequent(arr []float64) float64 {
	frequency := make(map[float64]int)
	maxCount := 0
	var mostFrequentElement float64

	for _, num := range arr {
		frequency[num]++
		if frequency[num] > maxCount {
			maxCount = frequency[num]
			mostFrequentElement = num
		}
	}

	return mostFrequentElement
}

// adjustImagesToWidth 调整所有图片到目标宽度
func adjustImagesToWidth(imageInfos []imageInfo, targetWidth float64) []imageInfo {
	adjusted := make([]imageInfo, len(imageInfos))

	for i, img := range imageInfos {
		// 如果宽度匹配，直接使用原始尺寸
		if img.width == targetWidth {
			adjusted[i] = imageInfo{
				width:  img.width,
				height: img.height,
				path:   img.path,
			}
			continue
		}

		// 计算缩放比例和新尺寸
		scale := targetWidth / img.width
		adjusted[i] = imageInfo{
			width:  targetWidth,
			height: img.height * scale,
			path:   img.path,
		}
	}

	return adjusted
}
