package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type DownloadTask struct {
	Url  string
	Dist string
}

type BatchDownloadResult struct {
	Url string
	Err error
}

// NewHTTPClient 创建带代理和超时的HTTP客户端
func NewHTTPClient(proxy string, timeout time.Duration) *http.Client {
	// 创建带代理的Transport
	transport := &http.Transport{}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
			Logger.Debug("配置HTTP代理", Str("proxy", proxy))
		} else {
			Logger.Warn("代理URL解析失败",
				Str("proxy", proxy),
				Err(err))
		}
	}

	// 创建带超时的客户端
	Logger.Debug("创建HTTP客户端",
		Float64("timeout_seconds", timeout.Seconds()))
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// Download 下载函数
func Download(url, dist, proxy string) (err error) {
	Logger.Info("开始下载文件",
		Str("url", url),
		Str("dist", dist),
		Str("proxy", proxy))

	// 创建HTTP请求
	client := NewHTTPClient(proxy, 15*time.Second)
	resp, err := client.Get(url)
	if err != nil {
		err = fmt.Errorf("HTTP请求失败: %w (url=%s)", err, url)
		Logger.Error("下载失败", Err(err))
		return err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("无效状态码: %d", resp.StatusCode)
		Logger.Error("下载失败",
			Err(err),
			Int("status_code", resp.StatusCode))
		return err
	}

	// 创建目标文件
	out, err := os.Create(dist)
	if err != nil {
		err = fmt.Errorf("文件创建失败: %w", err)
		Logger.Error("下载失败", Err(err))
		return err
	}

	// 错误时删除文件
	defer func() {
		if err != nil {
			Logger.Warn("删除不完整文件",
				Str("file", dist),
				Err(err))
			os.Remove(dist) // 删除空白或不完整文件
		}
	}()

	defer out.Close()

	// 复制数据
	Logger.Debug("开始复制文件内容",
		Str("url", url),
		Str("dist", dist))
	if _, err := io.Copy(out, resp.Body); err != nil {
		err = fmt.Errorf("文件复制失败: %w", err)
		Logger.Error("下载失败", Err(err))
		return err
	}

	Logger.Info("文件下载成功",
		Str("url", url),
		Str("dist", dist))
	return nil
}

// DownloadWithRetry 带有重试机制的下载函数
func DownloadWithRetry(url, dist, proxy string, maxRetries int) error {
	Logger.Info("开始带重试的下载",
		Str("url", url),
		Str("dist", dist),
		Int("max_retries", maxRetries))

	var err error
	retryDelay := 1 * time.Second // 初始重试延迟
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 说明第一次下载失败了，需要等待一段时间后重试
			Logger.Warn("准备重试下载",
				Str("url", url),
				Int("attempt", attempt),
				Int("max_retries", maxRetries),
				Float64("delay_seconds", retryDelay.Seconds()))
			time.Sleep(retryDelay)
			retryDelay *= 2 // 每次重试延迟加倍
		}

		// 执行下载
		err = Download(url, dist, proxy)
		if err == nil {
			// 下载成功
			Logger.Info("重试下载成功",
				Str("url", url),
				Str("dist", dist),
				Int("attempts", attempt+1))
			return nil
		}

		// 记录错误
		Logger.Warn("下载尝试失败",
			Str("url", url),
			Int("attempt", attempt+1),
			Int("max_retries", maxRetries),
			Err(err))
	}

	// 所有重试失败后返回错误
	err = fmt.Errorf("下载失败，已重试 %d 次: %w", maxRetries, err)
	Logger.Error("所有重试均失败",
		Str("url", url),
		Int("attempts", maxRetries),
		Err(err))
	return err
}

// BatchDownload 多线程下载
func BatchDownload(tasks []DownloadTask, proxy string, workers int, maxRetries int) []BatchDownloadResult {
	// 处理空任务列表
	if len(tasks) == 0 {
		Logger.Warn("批量下载接收到空任务列表")
		return []BatchDownloadResult{}
	}

	Logger.Info("开始批量下载任务",
		Int("total_tasks", len(tasks)),
		Int("workers", workers),
		Int("max_retries", maxRetries))

	// 限制worker数量不超过任务数
	if workers > len(tasks) {
		workers = len(tasks)
		Logger.Debug("调整worker数量",
			Int("new_workers", workers))
	}

	var wg sync.WaitGroup
	taskCh := make(chan DownloadTask, len(tasks))
	resultCh := make(chan BatchDownloadResult, len(tasks))
	results := make([]BatchDownloadResult, len(tasks))

	// 准备任务通道
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	// 启动工作协程池
	Logger.Debug("启动下载工作协程",
		Int("worker_count", workers))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			Logger.Debug("工作协程启动",
				Int("worker_id", workerID))
			for task := range taskCh {
				Logger.Debug("工作协程处理任务",
					Int("worker_id", workerID),
					Str("url", task.Url),
					Str("dist", task.Dist))

				err := DownloadWithRetry(task.Url, task.Dist, proxy, maxRetries)
				resultCh <- BatchDownloadResult{
					Url: task.Url,
					Err: err,
				}
			}
			Logger.Debug("工作协程退出",
				Int("worker_id", workerID))
		}(i)
	}

	// 收集结果协程
	go func() {
		wg.Wait()
		close(resultCh)
		Logger.Debug("所有工作协程已完成")
	}()

	// 按任务顺序重组结果（通过URL匹配）
	successCount := 0
	failureCount := 0
	for res := range resultCh {
		for i, task := range tasks {
			if task.Url == res.Url {
				results[i] = res
				if res.Err == nil {
					successCount++
				} else {
					failureCount++
				}
				break
			}
		}
	}

	Logger.Info("批量下载任务完成",
		Int("total_tasks", len(tasks)),
		Int("success_count", successCount),
		Int("failure_count", failureCount))

	return results
}
