package app

import "strings"

const renameString = "0.gr."

func makeCrunchFilePath(bucketName, filePath string) string {
	ret := strings.TrimPrefix(filePath, "/")
	ret = strings.TrimPrefix(ret, bucketName)
	ret = strings.TrimPrefix(ret, "/")

	if strings.Contains(ret, ".") {
		// Replace the last occurrence of "." with ".0.gr."
		lastDotIndex := strings.LastIndex(ret, ".")
		ret = ret[:lastDotIndex] + "." + renameString + ret[lastDotIndex+1:]
	} else if strings.Contains(ret, "/") {
		// Replace the last occurrence of "/" with "/0.gr."
		lastSlashIndex := strings.LastIndex(ret, "/")
		ret = ret[:lastSlashIndex] + "/" + renameString + ret[lastSlashIndex+1:]
	} else {
		ret = renameString + ret
	}

	return ret
}

func isCrunchedFile(filePath string) bool {
	return strings.Contains(filePath, renameString)
}
