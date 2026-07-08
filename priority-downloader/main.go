package main

import (
	"fmt"
	"sync"
	"time"
)

type File struct {
	Name     string
	Size     int
	Priority string
}

type Stats struct {
	mu    sync.Mutex
	size  int
	files []File
}

func Downloader(id int, wg *sync.WaitGroup, highPriorityChan <-chan File, lowPriorityChan <-chan File, stats *Stats) {
	defer wg.Done()
	highClosed := false
	lowClosed := false
	for {
		if highClosed && lowClosed {
			fmt.Printf("[Воркер %d] Все каналы закрыты. Завершаю работу.\n", id)
			return
		}

		select {
		case file, ok := <-highPriorityChan:
			if !ok {
				highClosed = true
				continue
			}
			time.Sleep(3 * time.Second)

			stats.mu.Lock()
			stats.files = append(stats.files, file)
			stats.size += file.Size
			stats.mu.Unlock()
			fmt.Printf("[Воркер %d] Обработал %s в высоком приоритете\n", id, file.Name)
		default:
			select {
			case file, ok := <-lowPriorityChan:
				if !ok {
					lowClosed = true
					continue
				}

				time.Sleep(3 * time.Second)
				stats.mu.Lock()
				stats.files = append(stats.files, file)
				stats.size += file.Size
				stats.mu.Unlock()
				fmt.Printf("[Воркер %d] Обработал %s в низком приоритете\n", id, file.Name)

			default:
				time.Sleep(20 * time.Millisecond)
			}
		}
	}
}

func main() {
	highPriorityChan := make(chan File, 20)
	lowPriorityChan := make(chan File, 20)
	var wg sync.WaitGroup

	stats := Stats{
		files: make([]File, 0),
	}

	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go Downloader(i, &wg, highPriorityChan, lowPriorityChan, &stats)
	}

	for i := 1; i <= 5; i++ {
		lowPriorityChan <- File{Name: fmt.Sprintf("фоновые-обои-%d.jpg", i), Size: 300, Priority: "low"}
	}

	for i := 1; i <= 3; i++ {
		highPriorityChan <- File{Name: fmt.Sprintf("критический-апдейт-%d.msi", i), Size: 500, Priority: "high"}
	}

	for i := 6; i <= 8; i++ {
		lowPriorityChan <- File{Name: fmt.Sprintf("песня-%d.mp3", i), Size: 200, Priority: "low"}
	}
	close(highPriorityChan)
	close(lowPriorityChan)

	wg.Wait()

	fmt.Println("\n==========================================")
	fmt.Printf("Всего скачано файлов: %d\n", len(stats.files))
	fmt.Printf("Общий объем данных: %d уб\n", stats.size)
	fmt.Println("==========================================")
}
