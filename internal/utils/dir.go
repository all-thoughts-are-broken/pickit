package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type DirInfo struct {
	Name  string
	Files []string
}

// ProcessPaths 处理路径参数
func ProcessPaths(input, output string) (string, string) {
	// 统一为POSIX格式处理
	normalizedInput := filepath.ToSlash(filepath.Clean(input))
	normalizedOutput := filepath.ToSlash(filepath.Clean(output))

	// 处理用户目录扩展
	if strings.HasPrefix(normalizedInput, "~/") {
		home, _ := os.UserHomeDir()
		normalizedInput = filepath.ToSlash(filepath.Join(home, normalizedInput[2:]))
	}

	// 根据当前操作系统调整路径格式
	if runtime.GOOS == "windows" {
		input = filepath.FromSlash(normalizedInput)
		output = filepath.FromSlash(normalizedOutput)
	} else {
		input = normalizedInput
		output = normalizedOutput
	}

	LogInfo("路径处理完成",
		Str("input", input),
		Str("output", output))
	return input, output
}

func GetDirInfo(dir string) ([]DirInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	// 判断是否为扁平结构
	hasSubDir := false
	for _, entry := range entries {
		if entry.IsDir() {
			hasSubDir = true
			break
		}
	}

	result := make([]DirInfo, 0)

	if !hasSubDir {
		// 单层目录处理
		files := make([]string, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
		SortNumericPaths(files)
		result = append(result, DirInfo{"", files})
	} else {
		// 多层目录处理
		folders := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() {
				folders = append(folders, entry.Name())
			}
		}

		SortNumericDirs(folders)

		for _, folder := range folders {
			subEntries, err := os.ReadDir(filepath.Join(dir, folder))
			if err != nil {
				return nil, fmt.Errorf("读取子目录 %s 失败: %w", folder, err)
			}

			files := make([]string, 0, len(subEntries))
			for _, entry := range subEntries {
				if !entry.IsDir() {
					files = append(files, filepath.Join(dir, folder, entry.Name()))
				}
			}

			SortNumericPaths(files)
			result = append(result, DirInfo{folder, files})
		}
	}

	return result, nil
}

// SortNumericPaths 对路径切片进行数字优先排序
func SortNumericPaths(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		nameI := strings.TrimSuffix(filepath.Base(paths[i]), filepath.Ext(paths[i]))
		nameJ := strings.TrimSuffix(filepath.Base(paths[j]), filepath.Ext(paths[j]))

		numI, errI := strconv.Atoi(nameI)
		numJ, errJ := strconv.Atoi(nameJ)

		if errI == nil && errJ == nil {
			return numI < numJ
		}

		return nameI < nameJ
	})
}

// SortNumericDirs 对目录名切片进行数字优先排序
func SortNumericDirs(dirs []string) {
	sort.Slice(dirs, func(i, j int) bool {
		numI, errI := strconv.Atoi(dirs[i])
		numJ, errJ := strconv.Atoi(dirs[j])

		if errI == nil && errJ == nil {
			return numI < numJ
		}

		return dirs[i] < dirs[j]
	})
}
