package mode

import (
	"path"
	"path/filepath"
	"pickit/internal/utils"
)

func DownloadAlbum(cdn, output, proxy string, aid, count, concurrency int) {
	// 构建下载 url 切片
	urls := utils.ImageUrlBuilder(aid, cdn, count)

	// 构建下载任务
	task := make([]utils.DownloadTask, len(urls))
	for i, url := range urls {
		task[i] = utils.DownloadTask{
			Url:  url,
			Dist: path.Join(output, filepath.Base(url)),
		}
	}

	_ = utils.BatchDownload(task, proxy, concurrency, 6)
}
