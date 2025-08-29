package mode

import (
	"path"
	"path/filepath"
	"pickit/internal/utils"
	"strings"
)

func RestoreImages(input, output string, aid, concurrency int) {
	dirInfo, err := utils.GetDirInfo(input)
	if err != nil {
		utils.LogFatal(err.Error())
	}
	if len(dirInfo) > 1 {
		utils.LogFatal("不支持非单层目录")
	}

	task := make([]utils.DecodeAndSaveTask, 0)
	for _, file := range dirInfo[0].Files {
		filename := filepath.Base(file)

		// 获取不带扩展名的文件名
		fileNameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		// 构建新的文件名，使用 .jpeg 后缀
		newFileName := fileNameWithoutExt + ".jpeg"

		task = append(task, utils.DecodeAndSaveTask{
			ImgSrcPath:      file,
			DecodedSavePath: path.Join(output, newFileName),
		})
	}

	_ = utils.BatchDecodeAndSave(220980, aid, task, concurrency)
}
