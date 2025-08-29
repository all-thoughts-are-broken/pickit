package utils

import (
	"fmt"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
	"image"
	"path/filepath"
	"strings"
	"sync"
)

type DecodeAndSaveTask struct {
	ImgSrcPath      string
	DecodedSavePath string
}

type DecodeAndSaveResult struct {
	ImgSrcPath string
	err        error
}

func DecodeAndSave(scrambleId, aid int, imgSrcPath, decodedSavePath string) error {
	LogDebug("开始处理图片",
		Str("source", imgSrcPath),
		Str("destination", decodedSavePath),
		Int("scrambleId", scrambleId),
		Int("aid", aid))
	// 从路径中提取文件名（去除扩展名）
	filename := filepath.Base(imgSrcPath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	// 获取图片分割数
	LogDebug("计算图片分割数量",
		Str("filename", filename),
		Int("scrambleId", scrambleId),
		Int("aid", aid))
	num := GetNum(scrambleId, aid, filename)
	LogInfo("图片分割计算结果",
		Str("source", imgSrcPath),
		Int("segments", num))

	LogDebug("打开原始图像", Str("path", imgSrcPath))
	// 打开原始图像
	srcImg, err := imaging.Open(imgSrcPath)
	if err != nil {
		LogError("无法打开图像文件", Str("path", imgSrcPath), Err(err))
		return fmt.Errorf("打开图片失败: %w", err)
	}
	// 无需解密，直接保存为JPEG
	if num == 0 {
		LogInfo("图片无需处理，直接保存",
			Str("source", imgSrcPath),
			Str("destination", decodedSavePath))
		return imaging.Save(srcImg, decodedSavePath)
	}

	// 获取图片尺寸
	bounds := srcImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	LogDebug("图片尺寸信息",
		Int("width", width),
		Int("height", height),
		Int("segments", num))
	// 计算每段高度
	segmentHeight := height / num
	remainder := height % num
	LogDebug("计算分段高度",
		Int("segmentHeight", segmentHeight),
		Int("remainder", remainder))
	// 创建一个新的透明背景画布
	dstImg := imaging.New(width, height, image.Transparent)
	LogDebug("创建透明背景画布",
		Int("width", width),
		Int("height", height))
	// 当前粘贴位置Y坐标
	dstY := 0

	// 处理每个分段
	for i := 0; i < num; i++ {
		// 计算当前分段高度
		currentSegmentHeight := segmentHeight
		if i == 0 {
			currentSegmentHeight += remainder
		}

		// 计算源图中的Y坐标（从底部开始）
		srcY := height - (segmentHeight*(i+1) + remainder)

		// 如果源图坐标过低，限制在范围内
		if srcY < 0 {
			srcY = 0
			LogDebug("调整源图Y坐标为0",
				Int("segment", i+1),
				Int("originalY", height-(segmentHeight*(i+1)+remainder)))
		}
		LogDebug("处理图像分段",
			Int("segment", i+1),
			Int("height", currentSegmentHeight),
			Int("sourceY", srcY))
		// 提取图像分段
		srcSegment := imaging.Crop(
			srcImg,
			image.Rect(0, srcY, width, srcY+currentSegmentHeight),
		)

		// 将分段粘贴到新图像
		dstImg = imaging.Paste(
			dstImg,
			srcSegment,
			image.Pt(0, dstY),
		)
		LogDebug("更新粘贴位置",
			Int("segment", i+1),
			Int("currentY", dstY))
		// 更新粘贴位置
		dstY += currentSegmentHeight
	}
	LogInfo("保存最终结果图像",
		Str("path", decodedSavePath))
	// 保存结果图像为JPEG
	return imaging.Save(dstImg, decodedSavePath)
}

func BatchDecodeAndSave(scrambleId, aid int, items []DecodeAndSaveTask, workers int) []DecodeAndSaveResult {
	LogInfo("开始批量处理图片",
		Int("total", len(items)),
		Int("workers", workers),
		Int("scrambleId", scrambleId),
		Int("aid", aid))

	if len(items) == 0 {
		LogWarn("没有需要处理的图片，批量处理终止")
		return nil
	}
	if workers <= 0 {
		workers = 1
		LogWarn("无效的工作线程数量，使用默认值",
			Int("provided", workers),
			Int("actual", 1))
	}

	tasks := make(chan DecodeAndSaveTask, len(items))
	results := make(chan DecodeAndSaveResult, len(items))

	var wg sync.WaitGroup

	LogDebug("启动工作线程", Int("count", workers))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			LogDebug("工作线程开始处理任务", Int("worker", workerID))
			for task := range tasks {
				LogDebug("工作线程处理新任务",
					Int("worker", workerID),
					Str("source", task.ImgSrcPath))

				err := DecodeAndSave(scrambleId, aid, task.ImgSrcPath, task.DecodedSavePath)

				if err != nil {
					LogError("图片处理失败",
						Int("worker", workerID),
						Str("source", task.ImgSrcPath),
						Err(err))
				} else {
					LogDebug("图片处理成功",
						Int("worker", workerID),
						Str("source", task.ImgSrcPath))
				}

				results <- DecodeAndSaveResult{
					ImgSrcPath: task.ImgSrcPath,
					err:        err,
				}
			}
			LogDebug("工作线程结束", Int("worker", workerID))
		}(i + 1)
	}

	LogDebug("分发任务到工作线程")
	for _, item := range items {
		tasks <- item
	}
	close(tasks)
	LogInfo("所有任务已分发到工作队列")

	LogDebug("等待所有工作线程完成")
	wg.Wait()
	LogInfo("所有工作线程已完成处理")

	close(results)

	res := make([]DecodeAndSaveResult, 0, len(items))
	for result := range results {
		res = append(res, result)
	}

	successCount := 0
	for _, r := range res {
		if r.err == nil {
			successCount++
		}
	}

	LogInfo("批量处理完成",
		Int("total", len(res)),
		Int("success", successCount),
		Int("failed", len(res)-successCount))

	return res
}
