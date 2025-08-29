package mode

import "pickit/internal/utils"

func CreatePDF(input, output, password string) {
	err := utils.ConvertImagesToPDF(input, output, password)
	if err != nil {
		utils.LogFatal("Failed to convert images to pdf")
	}
}
