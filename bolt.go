package main

import (
	"log"

	"github.com/asdine/storm"
)

type TaskStorage struct {
	db *storm.DB
}

type TaskDB struct {
	ID        int `storm:"id,increment"`
	ProjectID int `storm:"index"`
	Title     string
	Deadline  uint64
	Status    string `storm:"index"`
	Assigned  string `storm:"index"`
}

type ProjectDB struct {
	ID      int `storm:"id,increment"`
	Title   string
	Creator string `storm:"index"`
	Status  string `storm:"index"`
}

type DefaultProject struct {
	ID        int   `storm:"id,increment"`
	ChatID    int64 `storm:"index"`
	ProjectID int
}

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

func (self *TaskStorage) StoreTask(task Task, projectID int) error {
	data := TaskDB{
		ProjectID: projectID,
		Title:     task.Title,
		Deadline:  task.Deadline,
		Assigned:  task.Assigned,
		Status:    task.Status,
	}
	err := self.db.Save(&data)
	if err != nil {
		log.Printf("Cannot save task: %s", err.Error())
	}
	return err
}

func (self *TaskStorage) UpdateTask(task TaskDB) error {
	err := self.db.Update(task)
	if err != nil {
		log.Printf("Cannot update task %s: %s", task.Title, err.Error())
	}
	return err
}

func (self *TaskStorage) GetAllTasks() ([]TaskDB, error) {
	var tasks []TaskDB
	err := self.db.All(&tasks)
	if err != nil {
		log.Printf("Cannot get all tasks: %s", err.Error())
	}
	return tasks, err
}

func (self *TaskStorage) GetTaskByStatus(status string) ([]TaskDB, error) {
	var tasks []TaskDB
	err := self.db.Find("Status", status, &tasks)
	if err != nil {
		log.Printf("Cannot get task with status %s: %s", status, err.Error())
	}
	return tasks, err
}

func (self *TaskStorage) GetTaskByAssignee(telegramID string) ([]TaskDB, error) {
	var tasks []TaskDB
	err := self.db.Find("Assigned", telegramID, &tasks)
	if err != nil {
		log.Printf("Cannot get task by assignee %s: %s", telegramID, err.Error())
	}
	return tasks, err
}

func (self *TaskStorage) StoreProject(project Project) error {
	data := ProjectDB{
		Title:   project.Title,
		Creator: project.Creator,
	}
	err := self.db.Save(&data)
	if err != nil {
		log.Printf("Cannot save project: %s", err.Error())
	}
	return err
}

func (self *TaskStorage) GetAllProjects() ([]ProjectDB, error) {
	var result []ProjectDB
	err := self.db.All(&result)
	if err != nil {
		log.Printf("Cannot get projects: %s", err.Error())
	}
	return result, err
}

func (self *TaskStorage) GetProject(projectID int) (ProjectDB, error) {
	var project ProjectDB
	err := self.db.One("ID", projectID, &project)
	if err != nil {
		log.Printf("Cannot get project with id %d: %s", projectID, err.Error())
	}
	return project, err
}

func (self *TaskStorage) StoreDefaultProject(chatID int64, projectID int) error {
	defaultProject, err := self.GetDefaultProject(chatID)
	if err != nil {
		return err
	}
	if defaultProject.ProjectID != 0 {
		defaultProject.ProjectID = projectID
		err = self.db.Update(&defaultProject)
		if err != nil {
			log.Printf("Cannot save default project: %s", err.Error())
		}
	} else {
		defaultProject = DefaultProject{
			ChatID:    chatID,
			ProjectID: projectID,
		}
		err = self.db.Save(&defaultProject)
		if err != nil {
			log.Printf("Cannot save default project: %s", err.Error())
		}
	}
	return err
}

func (self *TaskStorage) GetDefaultProject(chatID int64) (DefaultProject, error) {
	var defaultProject DefaultProject
	err := self.db.One("ChatID", chatID, &defaultProject)
	if err != nil && err != storm.ErrNotFound {
		log.Printf("Cannot get default project of chat id %d: %s", chatID, err.Error())
		return defaultProject, err
	}
	return defaultProject, nil
}
