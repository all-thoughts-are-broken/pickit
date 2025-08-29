package utils

import "fmt"

func ImageUrlBuilder(aid int, cdn string, count int) []string {
	urls := make([]string, count)
	for i := 0; i < count; i++ {
		urls[i] = fmt.Sprintf("%s/media/photos/%d/%05d.webp", cdn, aid, i+1)
	}
	return urls
}
