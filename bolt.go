package main

import (
	"log"

	"github.com/asdine/storm"
)

//TaskStorage db object
type TaskStorage struct {
	db *storm.DB
}

//TaskDB db object
type TaskDB struct {
	ID          int `storm:"id,increment"`
	ProjectID   int `storm:"index"`
	Title       string
	Deadline    string
	Status      string `storm:"index"` // init,doing,done
	Assigned    string `storm:"index"`
	Description string
}

//ProjectDB db object
type ProjectDB struct {
	ID      int `storm:"id,increment"`
	Title   string
	Creator string `storm:"index"`
	Status  string `storm:"index"`
}

//DefaultProject db object
type DefaultProject struct {
	ID        int   `storm:"id,increment"`
	ChatID    int64 `storm:"index"`
	ProjectID int
}

//PinMessage db object
type PinMessage struct {
	ID      int `storm:"id,increment"`
	Message string
	ChatID  int64 `storm:"index"`
}

//NewStorage return new storage object
func NewStorage() (*TaskStorage, error) {
	db, err := storm.Open("task.db")

	if err != nil {
		log.Printf("Cannot open db: %s", err.Error())
	}
	storage := &TaskStorage{
		db: db,
	}
	return storage, nil
}

//StoreTask save new task to db
func (t *TaskStorage) StoreTask(task Task, projectID int) error {
	data := TaskDB{
		ProjectID: projectID,
		Title:     task.Title,
		Deadline:  task.Deadline,
		Assigned:  task.Assigned,
		Status:    task.Status,
	}
	err := t.db.Save(&data)
	if err != nil {
		log.Printf("Cannot save task: %s", err.Error())
	}
	return err
}

//UpdateTask update a task
//A task can be update assignee, deadline, status, etc.
func (t *TaskStorage) UpdateTask(task TaskDB) error {
	err := t.db.Update(&task)
	if err != nil {
		log.Printf("Cannot update task %s: %s", task.Title, err.Error())
	}
	return err
}

//GetAllTasks return all task available
func (t *TaskStorage) GetAllTasks() ([]TaskDB, error) {
	var tasks []TaskDB
	err := t.db.All(&tasks)
	if err != nil {
		log.Printf("Cannot get all tasks: %s", err.Error())
	}
	return tasks, err
}

//GetTaskByStatus get task by its status
func (t *TaskStorage) GetTaskByStatus(status string) ([]TaskDB, error) {
	var tasks []TaskDB
	err := t.db.Find("Status", status, &tasks)
	if err != nil {
		log.Printf("Cannot get task with status %s: %s", status, err.Error())
	}
	return tasks, err
}

//GetTaskByAssignee get task by its assignee
func (t *TaskStorage) GetTaskByAssignee(telegramID string) ([]TaskDB, error) {
	var tasks []TaskDB
	err := t.db.Find("Assigned", telegramID, &tasks)
	if err != nil {
		log.Printf("Cannot get task by assignee %s: %s", telegramID, err.Error())
	}
	return tasks, err
}

//GetTask by task ID
func (t *TaskStorage) GetTask(taskID int) (TaskDB, error) {
	var task TaskDB
	err := t.db.One("ID", taskID, &task)
	if err != nil {
		log.Printf("Cannot get task %d: %s", taskID, err.Error())
	}
	return task, err
}

//StoreProject store a project
func (t *TaskStorage) StoreProject(project Project) error {
	data := ProjectDB{
		Title:   project.Title,
		Creator: project.Creator,
	}
	err := t.db.Save(&data)
	if err != nil {
		log.Printf("Cannot save project: %s", err.Error())
	}
	return err
}

//GetAllProjects get all projects
func (t *TaskStorage) GetAllProjects() ([]ProjectDB, error) {
	var result []ProjectDB
	err := t.db.All(&result)
	if err != nil {
		log.Printf("Cannot get projects: %s", err.Error())
	}
	return result, err
}

//GetProject get a project by its id
func (t *TaskStorage) GetProject(projectID int) (ProjectDB, error) {
	var project ProjectDB
	err := t.db.One("ID", projectID, &project)
	if err != nil {
		log.Printf("Cannot get project with id %d: %s", projectID, err.Error())
	}
	return project, err
}

//StoreDefaultProject save default project
func (t *TaskStorage) StoreDefaultProject(chatID int64, projectID int) error {
	defaultProject, err := t.GetDefaultProject(chatID)
	if err != nil {
		return err
	}
	if defaultProject.ProjectID != 0 {
		defaultProject.ProjectID = projectID
		err = t.db.Update(&defaultProject)
		if err != nil {
			log.Printf("Cannot save default project: %s", err.Error())
		}
	} else {
		defaultProject = DefaultProject{
			ChatID:    chatID,
			ProjectID: projectID,
		}
		err = t.db.Save(&defaultProject)
		if err != nil {
			log.Printf("Cannot save default project: %s", err.Error())
		}
	}
	return err
}

//GetDefaultProject get default project
func (t *TaskStorage) GetDefaultProject(chatID int64) (DefaultProject, error) {
	var defaultProject DefaultProject
	err := t.db.One("ChatID", chatID, &defaultProject)
	if err != nil && err != storm.ErrNotFound {
		log.Printf("Cannot get default project of chat id %d: %s", chatID, err.Error())
		return defaultProject, err
	}
	return defaultProject, nil
}

//UpdatePinMessage update pin message for not super group
func (t *TaskStorage) UpdatePinMessage(message string, chatID int64) error {
	var messages []PinMessage
	err := t.db.All(&messages)
	if err != nil {
		log.Printf("Cannot update pin message: %s", err.Error())
		return err
	}
	pinMessage := PinMessage{}
	if len(messages) != 0 {
		pinMessage = messages[0]
		pinMessage.Message = message
		pinMessage.ChatID = chatID
		err = t.db.Update(&pinMessage)
		if err != nil {
			log.Printf("Cannot update pin message: %s", err.Error())
		}
	} else {
		pinMessage.Message = message
		pinMessage.ChatID = chatID
		err = t.db.Save(&pinMessage)
		if err != nil {
			log.Printf("Cannot update pin message: %s", err.Error())
		}
	}
	return err
}

//GetPinMessage get pinned message
func (t *TaskStorage) GetPinMessage(chatID int64) (string, error) {
	var messages []PinMessage
	err := t.db.Find("ChatID", chatID, &messages)
	if err != nil && err != storm.ErrNotFound {
		log.Printf("Cannot update pin message: %s", err.Error())
		return "", err
	}
	pinMessage := PinMessage{}
	if len(messages) != 0 {
		pinMessage = messages[0]
	}
	return pinMessage.Message, nil
}
