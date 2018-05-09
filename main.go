package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type BotConfig struct {
	Key string `json:"bot_key"`
}

type Bot struct {
	bot            *tb.Bot
	storage        *TaskStorage
	currentCommand string
}

var currentCommand string
var currentTask int

func readConfigFromFile(path string) (BotConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return BotConfig{}, err
	} else {
		result := BotConfig{}
		err := json.Unmarshal(data, &result)
		return result, err
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	path := "./config.json"
	botConfig, err := readConfigFromFile(path)
	if err != nil {
		log.Fatal(err)
	}
	tbot, err := tb.NewBot(tb.Settings{
		Token:  botConfig.Key,
		Poller: &tb.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		log.Fatalf("Cannot initiate new bot: %s", err.Error())
	}
	storage, err := NewStorage()
	mybot := Bot{
		bot:     tbot,
		storage: storage,
	}
	if err != nil {
		log.Panic(err)
		return
	}

	mybot.bot.Handle("/start", func(m *tb.Message) {
		mybot.bot.Send(m.Chat, fmt.Sprintf(`This is a bot for manage tasks.`))
	})

	mybot.bot.Handle("/create_task", func(m *tb.Message) {
		mybot.createTask(m)
	})

	mybot.bot.Handle("/create_project", func(m *tb.Message) {
		mybot.createProject(m)
	})

	mybot.bot.Handle("/list_task", func(m *tb.Message) {
		mybot.handleListTask(m)
	})

	mybot.bot.Handle("/list_projects", func(m *tb.Message) {
		mybot.handleListProjects(m)
	})

	mybot.bot.Handle("/set_default_project", func(m *tb.Message) {
		mybot.handleSetDefaultProject(m)
	})

	mybot.bot.Handle("/current_project", func(m *tb.Message) {
		mybot.handleCurrentProject(m)
	})

	mybot.bot.Handle(tb.OnText, func(m *tb.Message) {
		mybot.handleText(m)
	})

	mybot.bot.Handle("/assignTask", func(m *tb.Message) {
		// TODO:
	})

	mybot.bot.Handle("/setDeadline", func(m *tb.Message) {
		// TODO:
	})

	mybot.bot.Handle("/setStatus", func(m *tb.Message) {
		// TODO:
	})

	mybot.bot.Handle("/discuss", func(m *tb.Message) {
		// TODO: complete this function
	})

	mybot.bot.Handle("/pin", func(m *tb.Message) {
		mybot.handlePin(m)
	})

	// mybot.bot.Handle("/listTaskByStatus", func(m *tb.Message) {
	// 	mybot.handleListTaskByStatus(m)
	// })

	mybot.bot.Handle("/listTaskByAssignee", func(m *tb.Message) {
		mybot.handleListTaskByAssignee(m)
	})

	mybot.bot.Handle("/mine", func(m *tb.Message) {
		mybot.handleMyList(m)
	})

	mybot.bot.Start()
}

func (self Bot) saveProject(projectTitle string, m *tb.Message) {
	newProject := Project{
		Title:   projectTitle,
		Creator: m.Sender.Username,
	}
	err := self.storage.StoreProject(newProject)
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot create project: %s", err.Error()))
	} else {
		self.bot.Send(m.Chat, fmt.Sprintf("Create project *%s* successfully", projectTitle), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) saveTask(taskTitle string, m *tb.Message) {
	newTask := Task{
		Title: taskTitle,
	}
	defaultProject, _ := self.storage.GetDefaultProject(m.Chat.ID)
	err := self.storage.StoreTask(newTask, defaultProject.ProjectID)
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot create task: %s", err.Error()))
	} else {
		self.bot.Send(m.Chat, fmt.Sprintf("Create *%s* successfully", taskTitle), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) createProject(m *tb.Message) {
	messages := strings.Split(strings.TrimSpace(m.Text), " ")
	if len(messages) > 1 {
		self.saveProject(strings.Join(messages[1:], " "), m)
	} else {
		currentCommand = "create_project"
		self.bot.Send(m.Chat, fmt.Sprintf("Project name: "))
	}
}

func (self Bot) createTask(m *tb.Message) {
	currentCommand = "create_task"
	defaultProject, _ := self.storage.GetDefaultProject(m.Chat.ID)
	if defaultProject.ProjectID == 0 {
		projects, err := self.storage.GetAllProjects()
		if err != nil {
			self.bot.Send(m.Chat, fmt.Sprintf("Cannot get list project to set: %s", err.Error()))
			return
		}
		inlineKeys := [][]tb.InlineButton{}
		for _, project := range projects {
			inlineBtn := tb.InlineButton{
				Unique: strconv.Itoa(project.ID),
				Text:   project.Title,
			}
			self.bot.Handle(&inlineBtn, func(c *tb.Callback) {
				id, _ := strconv.Atoi(inlineBtn.Unique)
				self.setDefaultProject(m.Chat.ID, id, m)
				self.bot.Respond(c, &tb.CallbackResponse{})
			})

			inlineKeys = append(inlineKeys, []tb.InlineButton{inlineBtn})
		}
		self.bot.Send(m.Chat, "Which project you want to create task for? \n", &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		})
	} else {
		project, _ := self.storage.GetProject(defaultProject.ProjectID)
		log.Printf(currentCommand)
		self.bot.Send(m.Chat, fmt.Sprintf("Create task for *%s*. Task title: ", project.Title), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) handleListTask(m *tb.Message) {
	message := "Which task do you want to list? \n"
	inlineKeys := [][]tb.InlineButton{}
	all := tb.InlineButton{
		Unique: "all",
		Text:   "All",
	}
	self.bot.Handle(&all, func(c *tb.Callback) {
		self.handleListAllTasks(m)
		self.bot.Respond(c, &tb.CallbackResponse{})
	})
	inlineKeys = append(inlineKeys, []tb.InlineButton{all})
	notStart := tb.InlineButton{
		Unique: "not_start",
		Text:   "Not Started Yet",
	}
	inlineKeys = append(inlineKeys, []tb.InlineButton{notStart})
	doing := tb.InlineButton{
		Unique: "doing",
		Text:   "Doing",
	}
	self.bot.Handle(&notStart, func(c *tb.Callback) {
		self.handleListTaskByStatus(m, notStart.Unique)
	})
	inlineKeys = append(inlineKeys, []tb.InlineButton{doing})
	done := tb.InlineButton{
		Unique: "done",
		Text:   "Done",
	}
	inlineKeys = append(inlineKeys, []tb.InlineButton{done})
	byAssignee := tb.InlineButton{
		Unique: "by_assignee",
		Text:   "By Assignee",
	}
	inlineKeys = append(inlineKeys, []tb.InlineButton{byAssignee})
	self.bot.Reply(m, message, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		},
	})
}

func (self Bot) handleListProjects(m *tb.Message) {
	projects, err := self.storage.GetAllProjects()
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get list projects: %s", err.Error()))
	} else {
		if len(projects) == 0 {
			self.bot.Reply(m, "There is not a project yet.")
		}
		message := "Project list: \n"
		for _, project := range projects {
			message += fmt.Sprintf("*%s* created by _@%s_ \n", project.Title, project.Creator)
		}
		self.bot.Reply(m, message, &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) handleSetDefaultProject(m *tb.Message) {
	projects, err := self.storage.GetAllProjects()
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get list project to set: %s", err.Error()))
		return
	}
	inlineKeys := [][]tb.InlineButton{}
	for _, project := range projects {
		inlineBtn := tb.InlineButton{
			Unique: strconv.Itoa(project.ID),
			Text:   project.Title,
		}
		self.bot.Handle(&inlineBtn, func(c *tb.Callback) {
			id, _ := strconv.Atoi(inlineBtn.Unique)
			self.setDefaultProject(m.Chat.ID, id, m)
			self.bot.Respond(c, &tb.CallbackResponse{})
		})

		inlineKeys = append(inlineKeys, []tb.InlineButton{inlineBtn})
	}
	self.bot.Send(m.Chat, "Which project you want to set default? \n", &tb.ReplyMarkup{
		InlineKeyboard: inlineKeys,
	})
}

func (self Bot) setDefaultProject(chatID int64, projectID int, m *tb.Message) {
	err := self.storage.StoreDefaultProject(chatID, projectID)
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot set default project for this chat: %s", err.Error()))
	} else {
		defaultProject, _ := self.storage.GetDefaultProject(chatID)
		project, _ := self.storage.GetProject(defaultProject.ProjectID)
		if currentCommand != "create_task" {
			self.bot.Send(m.Chat, fmt.Sprintf("Default project for this chat now is: %s", project.Title))
		} else {
			self.bot.Send(m.Chat, fmt.Sprintf("Create task for *%s*. Task title: ", project.Title))
		}
	}
}

func (self Bot) handleCurrentProject(m *tb.Message) {
	defaultProject, err := self.storage.GetDefaultProject(m.Chat.ID)
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get current project for this chat: %s", err.Error()))
	} else {
		project, _ := self.storage.GetProject(defaultProject.ProjectID)
		self.bot.Reply(m, fmt.Sprintf("Current project for this chat: %s", project.Title))
	}
}

func (self Bot) handleListAllTasks(m *tb.Message) {
	tasks, err := self.storage.GetAllTasks()
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get task list: %s", err.Error()))
	} else {
		self.bot.Reply(m, "Task list:")
		for _, task := range tasks {
			message := ""
			message += fmt.Sprintf("*%s* - *%s* - *%s*", task.Assigned, task.Deadline, task.Title)

			inlineKeys := [][]tb.InlineButton{}
			assignButton := tb.InlineButton{
				Unique: strconv.Itoa(task.ID),
				Text:   "👤",
			}
			self.bot.Handle(&assignButton, func(c *tb.Callback) {
				taskID, _ := strconv.Atoi(assignButton.Unique)
				self.handleAssignTask(m, taskID)
			})
			inlineKeys = append(inlineKeys, []tb.InlineButton{assignButton})
			self.bot.Reply(m, message, &tb.SendOptions{
				ParseMode: tb.ModeMarkdown,
				ReplyMarkup: &tb.ReplyMarkup{
					InlineKeyboard: inlineKeys,
				},
			})
		}
	}
}

func (self Bot) sendTasks(taskType string, m *tb.Message, tasks []TaskDB) {
	if len(tasks) == 0 {
		self.bot.Reply(m, fmt.Sprintf("There is no *%s* task for show", taskType), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) handleListTaskByStatus(m *tb.Message, status string) {
	tasks, err := self.storage.GetTaskByStatus(status)
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get *%s* tasks: %s", status, err.Error()), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
	self.sendTasks(status, m, tasks)
}

func (self Bot) handleListTaskByAssignee(m *tb.Message) {

}

func (self Bot) handleMyList(m *tb.Message) {
	telegramID := m.Sender.Username
	tasks, err := self.storage.GetTaskByAssignee("@" + telegramID)
	if err != nil {
		self.bot.Reply(m, fmt.Sprintf("Cannot get your task list: %s", err.Error()))
	} else {
		message := "Your task list: \n"
		for _, task := range tasks {
			message += fmt.Sprintf("*%s*", task.Title)
		}
		self.bot.Reply(m, message, &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) handleAssignTask(m *tb.Message, taskID int) {
	currentCommand = "assign_task"
	currentTask = taskID
	self.bot.Send(m.Chat, fmt.Sprintf("Who do you want to assign?"))
}

func (self Bot) assignTask(assignee string, currentTask int, m *tb.Message) {
	currentCommand = ""
	task, err := self.storage.GetTask(currentTask)
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot assign task: %s", err.Error))
		return
	}
	task.Assigned = assignee
	err = self.storage.UpdateTask(task)
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot assigntask: %s", err.Error()))
		return
	}
	self.bot.Send(m.Chat, fmt.Sprintf("Task *%s* is assigned to *%s* successfully", task.Title, assignee), &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	})
}

func (self Bot) handlePin(m *tb.Message) {
	if m.IsReply() {
		// update pin message
		pinMessage := m.ReplyTo
		err := self.storage.UpdatePinMessage(pinMessage.Text)
		log.Printf("Error: %+v", err)
		if err != nil {
			self.bot.Reply(m, fmt.Sprintf("Cannot pin message: %s", err.Error()))
		} else {
			self.bot.Reply(m, fmt.Sprintf("Update pin message successfully"))
		}
	} else {
		// show pin message
		pinMessage, err := self.storage.GetPinMessage()
		if err != nil {
			self.bot.Reply(m, fmt.Sprintf("Cannot show pin message: %s", err.Error()))
		} else {
			self.bot.Reply(m, fmt.Sprintf("Pined message: %s", pinMessage))
		}
	}
}

func (self Bot) handleText(m *tb.Message) {
	log.Printf(currentCommand)
	switch currentCommand {
	case "create_task":
		self.saveTask(m.Text, m)
		currentCommand = ""
	case "create_project":
		self.saveProject(m.Text, m)
		currentCommand = ""
	case "assign_task":
		self.assignTask(m.Text, currentTask, m)
	}
}
