package main

import (
	"fmt"
	"sync"
	"time"
)

// Структура Task описывает задачу
type Task struct {
	id         int       // Уникальный идентификатор задачи
	cT         time.Time // Время создания задачи
	fT         time.Time // Время завершения обработки задачи
	taskRESULT []byte    // Результат выполнения задачи
}

func main() {
	// cоздание такска

	taskCreator := func(taskChan chan<- Task) {
		go func() {
			for {
				cT := time.Now()
				if cT.Nanosecond()%2 > 0 {
					taskChan <- Task{cT: cT, taskRESULT: []byte("Some error occurred")}
				} else {
					taskChan <- Task{cT: cT}
				}
				time.Sleep(time.Second * 10) // Генерация задач каждые 10 секунд
			}
		}()
	}

	//Обработка задачи
	taskWorker := func(task Task) Task {
		if len(task.taskRESULT) == 0 {
			time.Sleep(time.Millisecond * 150)
			task.fT = time.Now()
			task.taskRESULT = []byte("Task has been succeeded")
		}
		return task
	}

	processTasks := func(taskChan <-chan Task, doneTasks chan<- Task, undoneTasks chan<- Task) {
		for task := range taskChan {
			processedTask := taskWorker(task)
			if string(processedTask.taskRESULT) == "Task has been succeeded" {
				doneTasks <- processedTask
			} else {
				undoneTasks <- processedTask
			}
		}
		close(doneTasks)
		close(undoneTasks)
	}

	/*
		C помощью пакета sync сделаем мультипоточную обработку заданий,
		будем использовать метод sync.WaitGroup для ожидания завершения всех горутин
	*/

	var wg sync.WaitGroup
	taskChan := make(chan Task)
	doneTasks := make(chan Task)
	undoneTasks := make(chan Task)

	wg.Add(1)
	go func() {
		defer wg.Done()
		taskCreator(taskChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		processTasks(taskChan, doneTasks, undoneTasks)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		printResults(doneTasks, undoneTasks)
	}()

	wg.Wait()
}

// Выведем результаты в консоль
func printResults(doneTasks <-chan Task, undoneTasks <-chan Task) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case doneTask, ok := <-doneTasks:
			if !ok {
				doneTasks = nil
				continue
			}
			fmt.Printf("Task Completed: ID %d, Created at %s, Completed в %s\n", doneTask.id, doneTask.cT.Format(time.RFC3339), doneTask.fT.Format(time.RFC3339))
		case undoneTask, ok := <-undoneTasks:
			if !ok {
				undoneTasks = nil
				continue
			}
			fmt.Printf("Unfinished task: ID %d, Created at %s, Error: %s\n", undoneTask.id, undoneTask.cT.Format(time.RFC3339), string(undoneTask.taskRESULT))
		case <-ticker.C:
			fmt.Println("Print Results..")
		}

		if doneTasks == nil && undoneTasks == nil {
			break
		}
	}
}
