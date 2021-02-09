package utils

type Result struct {
	SuccessCount []int
	TotalCount   []int
}

func NewResult(threads int) *Result {
	return &Result{SuccessCount: make([]int, threads),
		TotalCount: make([]int, threads)}
}

type UploadResult struct {
	*Result
	FileInfo [][]FileInfo
}

func NewUploadResult(threads int) *UploadResult {
	uploadResult := new(UploadResult)
	uploadResult.Result = NewResult(threads)
	uploadResult.FileInfo = make([][]FileInfo, threads)
	return uploadResult
}

func SumIntArray(arr []int) int {
	sum := 0
	for _, i := range arr {
		sum += i
	}
	return sum
}
