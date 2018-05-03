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
		mybot.bot.Send(m.Chat, "Thank you for trying")
	})

	mybot.bot.Handle("/createTask", func(m *tb.Message) {
		mybot.createTask(m)
	})

	mybot.bot.Handle("/createProject", func(m *tb.Message) {
		mybot.createProject(m)
	})

	mybot.bot.Handle("/listTask", func(m *tb.Message) {
		mybot.handleListTask(m)
	})

	mybot.bot.Handle("/listProjects", func(m *tb.Message) {
		mybot.handleListProjects(m)
	})

	mybot.bot.Handle("/setDefaultProject", func(m *tb.Message) {
		mybot.handleSetDefaultProject(m)
	})

	mybot.bot.Handle("/currentProject", func(m *tb.Message) {
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
		// TODO:
	})

	mybot.bot.Handle("/listTaskByStatus", func(m *tb.Message) {
		// TODO:
	})

	mybot.bot.Handle("/listTaskByAssignee", func(m *tb.Message) {
		// TODO:
	})

	mybot.bot.Handle("/mine", func(m *tb.Message) {
		// TODO:
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
	log.Printf("Saving task")
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
	tasks, err := self.storage.GetAllTasks()
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot get task list: %s", err.Error()))
	} else {
		message := "Task list: \n"
		for _, task := range tasks {
			message += fmt.Sprintf("%s", task.Title)
		}
		self.bot.Send(m.Chat, message)
	}
}

func (self Bot) handleListProjects(m *tb.Message) {
	projects, err := self.storage.GetAllProjects()
	if err != nil {
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot get list projects: %s", err.Error()))
	} else {
		if len(projects) == 0 {
			self.bot.Send(m.Chat, "There is not a project yet.")
		}
		message := "Project list: \n"
		for _, project := range projects {
			message += fmt.Sprintf("*%s* created by _@%s_ \n", project.Title, project.Creator)
		}
		self.bot.Send(m.Chat, message, &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (self Bot) handleSetDefaultProject(m *tb.Message) {
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
		self.bot.Send(m.Chat, fmt.Sprintf("Cannot get current project for this chat: %s", err.Error()))
	} else {
		project, _ := self.storage.GetProject(defaultProject.ProjectID)
		self.bot.Send(m.Chat, fmt.Sprintf("Current project for this chat: %s", project.Title))
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
	}
}
