package walker

import (
	"fmt"
	"github.com/samber/lo"
)

type Progress struct {
	destinationObject string
	completed         int64
	totalSize         int64
}

func NewProgress(destinationObject string, files []ExportFile) *Progress {

	return &Progress{
		destinationObject: destinationObject,
		totalSize: lo.SumBy(files, func(item ExportFile) int64 {
			return item.Size
		}),
	}
}

func (p *Progress) Starting(f ExportFile) {
	fmt.Printf("\r%s - %s %d%% (%s / %s)",
		p.destinationObject,
		f.RelPath,
		p.completed*100/p.totalSize,
		toReadableSize(p.completed),
		toReadableSize(p.totalSize),
	)
}

func (p *Progress) Finished(f ExportFile) {
	p.completed += f.Size
}

func (p *Progress) Done() {
	fmt.Printf("\r%s - finished, %s\n", p.destinationObject, toReadableSize(p.totalSize))
}

func toReadableSize(size int64) string {
	const mb = 1024 * 1024
	const gb = mb * 1024
	if size < gb {
		return fmt.Sprintf("%.2f MB", float64(size)/float64(mb))
	}

	return fmt.Sprintf("%.2f GB", float64(size)/float64(gb))
}
