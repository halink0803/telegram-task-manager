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

// BotConfig object
type BotConfig struct {
	Key string `json:"bot_key"`
}

//Bot object
type Bot struct {
	bot     *tb.Bot
	storage *TaskStorage
}

var currentCommand map[string]string
var currentTask int

func readConfigFromFile(path string) (BotConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return BotConfig{}, err
	}
	result := BotConfig{}
	err = json.Unmarshal(data, &result)
	return result, err
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

	mybot.bot.Handle("/list_tasks", func(m *tb.Message) {
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

	mybot.bot.Handle("/assign", func(m *tb.Message) {
		mybot.handleAssignTask(m)
	})

	mybot.bot.Handle("/set_deadline", func(m *tb.Message) {
		mybot.handleSetDeadline(m)
	})

	mybot.bot.Handle("/set_status", func(m *tb.Message) {
		mybot.handleSetStatus(m)
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

func (b Bot) saveProject(projectTitle string, m *tb.Message) {
	newProject := Project{
		Title:   projectTitle,
		Creator: m.Sender.Username,
	}
	err := b.storage.StoreProject(newProject)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot create project: %s", err.Error()))
	} else {
		b.bot.Send(m.Chat, fmt.Sprintf("Create project *%s* successfully", projectTitle), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) saveTask(m *tb.Message) {
	defaultProject, _ := b.storage.GetDefaultProject(m.Chat.ID)
	entities := m.Entities
	taskFactors := strings.Split(m.Text, "-")
	assignee := ""
	for _, entity := range entities {
		if entity.Type == tb.EntityMention {
			if entity.User != nil {
				assignee = entity.User.Username
			} else {
				assignee = taskFactors[1]
			}
		}
	}
	title := taskFactors[0]
	deadline := ""
	if len(taskFactors) > 2 {
		deadline = taskFactors[2]
	}
	log.Printf("deadline: %s", deadline)
	newTask := Task{
		Assigned: assignee,
		Title:    title,
		Deadline: deadline,
	}
	err := b.storage.StoreTask(newTask, defaultProject.ProjectID)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot create task: %s", err.Error()))
	} else {
		b.bot.Send(m.Chat, fmt.Sprintf("Created *%s* for *%s*", title, assignee), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) createProject(m *tb.Message) {
	messages := strings.Split(strings.TrimSpace(m.Text), " ")
	if len(messages) > 1 {
		b.saveProject(strings.Join(messages[1:], " "), m)
	} else {
		if len(currentCommand) == 0 {
			currentCommand = map[string]string{}
		}
		currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = "create_project"
		b.bot.Send(m.Chat, fmt.Sprintf("Project name: "))
	}
}

func (b Bot) createTask(m *tb.Message) {
	currentCommand = map[string]string{
		fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID): "create_task",
	}
	defaultProject, _ := b.storage.GetDefaultProject(m.Chat.ID)
	if defaultProject.ProjectID == 0 {
		projects, err := b.storage.GetAllProjects()
		if err != nil {
			b.bot.Send(m.Chat, fmt.Sprintf("Cannot get list project to set: %s", err.Error()))
			return
		}
		inlineKeys := [][]tb.InlineButton{}
		for _, project := range projects {
			inlineBtn := tb.InlineButton{
				Unique: strconv.Itoa(project.ID),
				Text:   project.Title,
			}
			b.bot.Handle(&inlineBtn, func(c *tb.Callback) {
				id, _ := strconv.Atoi(inlineBtn.Unique)
				b.setDefaultProject(m.Chat.ID, id, m)
				b.bot.Respond(c, &tb.CallbackResponse{})
			})

			inlineKeys = append(inlineKeys, []tb.InlineButton{inlineBtn})
		}
		b.bot.Send(m.Chat, "Which project you want to create task for? \n", &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		})
	} else {
		project, _ := b.storage.GetProject(defaultProject.ProjectID)
		b.bot.Send(m.Chat, fmt.Sprintf(`Create task for *%s*. Follow this structure:
			Task Title (required) - @username (optional) - Deadline (optional) - Description (optional)`, project.Title), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) handleListTask(m *tb.Message) {
	message := "Which task do you want to list? \n"
	inlineKeys := [][]tb.InlineButton{}
	all := tb.InlineButton{
		Unique: "all",
		Text:   "All",
	}
	b.bot.Handle(&all, func(c *tb.Callback) {
		b.handleListAllTasks(m)
		b.bot.Respond(c, &tb.CallbackResponse{})
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
	b.bot.Handle(&notStart, func(c *tb.Callback) {
		b.handleListTaskByStatus(m, notStart.Unique)
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
	b.bot.Reply(m, message, &tb.SendOptions{
		ReplyMarkup: &tb.ReplyMarkup{
			InlineKeyboard: inlineKeys,
		},
	})
}

func (b Bot) handleListProjects(m *tb.Message) {
	projects, err := b.storage.GetAllProjects()
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get list projects: %s", err.Error()))
	} else {
		if len(projects) == 0 {
			b.bot.Reply(m, "There is not a project yet.")
		}
		message := "Project list: \n"
		for _, project := range projects {
			message += fmt.Sprintf("*%s* created by _@%s_ \n", project.Title, project.Creator)
		}
		b.bot.Reply(m, message, &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) handleSetDefaultProject(m *tb.Message) {
	projects, err := b.storage.GetAllProjects()
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get list project to set: %s", err.Error()))
		return
	}
	inlineKeys := [][]tb.InlineButton{}
	for _, project := range projects {
		inlineBtn := tb.InlineButton{
			Unique: strconv.Itoa(project.ID),
			Text:   project.Title,
		}
		b.bot.Handle(&inlineBtn, func(c *tb.Callback) {
			id, _ := strconv.Atoi(inlineBtn.Unique)
			b.setDefaultProject(m.Chat.ID, id, m)
			b.bot.Respond(c, &tb.CallbackResponse{})
		})

		inlineKeys = append(inlineKeys, []tb.InlineButton{inlineBtn})
	}
	b.bot.Send(m.Chat, "Which project you want to set default? \n", &tb.ReplyMarkup{
		InlineKeyboard: inlineKeys,
	})
}

func (b Bot) setDefaultProject(chatID int64, projectID int, m *tb.Message) {
	err := b.storage.StoreDefaultProject(chatID, projectID)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot set default project for this chat: %s", err.Error()))
	} else {
		defaultProject, _ := b.storage.GetDefaultProject(chatID)
		project, _ := b.storage.GetProject(defaultProject.ProjectID)
		command, exist := currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)]
		if exist && command != "create_task" {
			b.bot.Send(m.Chat, fmt.Sprintf("Default project for this chat now is: %s", project.Title))
		} else {
			b.bot.Send(m.Chat, fmt.Sprintf("Create task for *%s*. Task title: ", project.Title))
		}
	}
}

func (b Bot) handleCurrentProject(m *tb.Message) {
	defaultProject, err := b.storage.GetDefaultProject(m.Chat.ID)
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get current project for this chat: %s", err.Error()))
	} else {
		project, _ := b.storage.GetProject(defaultProject.ProjectID)
		b.bot.Reply(m, fmt.Sprintf("Current project for this chat: *%s*", project.Title), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) handleListAllTasks(m *tb.Message) {
	tasks, err := b.storage.GetAllTasks()
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get task list: %s", err.Error()))
	} else {
		b.bot.Reply(m, "Task list:")
		for _, task := range tasks {
			message := ""
			message += fmt.Sprintf("*%d* *%s* - *%s* - *%s* - %s", task.ID,
				task.Assigned, task.Deadline, task.Title, task.Status)

			// inlineKeys := [][]tb.InlineButton{}
			// assignButton := tb.InlineButton{
			// 	Unique: strconv.Itoa(task.ID),
			// 	Text:   "👤",
			// }
			// b.bot.Handle(&assignButton, func(c *tb.Callback) {
			// 	taskID, _ := strconv.Atoi(assignButton.Unique)
			// 	b.handleAssignTask(m, taskID)
			// })
			// inlineKeys = append(inlineKeys, []tb.InlineButton{assignButton})
			b.bot.Send(m.Chat, message, &tb.SendOptions{
				ParseMode: tb.ModeMarkdown,
				// ReplyMarkup: &tb.ReplyMarkup{
				// 	InlineKeyboard: inlineKeys,
				// },
			})
		}
	}
}

func (b Bot) sendTasks(taskType string, m *tb.Message, tasks []TaskDB) {
	if len(tasks) == 0 {
		b.bot.Reply(m, fmt.Sprintf("There is no *%s* task for show", taskType), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) handleListTaskByStatus(m *tb.Message, status string) {
	tasks, err := b.storage.GetTaskByStatus(status)
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get *%s* tasks: %s", status, err.Error()), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
	b.sendTasks(status, m, tasks)
}

func (b Bot) handleListTaskByAssignee(m *tb.Message) {

}

func (b Bot) handleMyList(m *tb.Message) {
	telegramID := m.Sender.Username
	log.Printf(m.Sender.Username)
	tasks, err := b.storage.GetTaskByAssignee("@" + telegramID)
	if err != nil {
		b.bot.Reply(m, fmt.Sprintf("Cannot get your task list: %s", err.Error()))
	} else {
		message := "Your task list: \n"
		for _, task := range tasks {
			message += fmt.Sprintf("*%s* - *%s* \n", task.Title, task.Deadline)
		}
		b.bot.Reply(m, message, &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
	}
}

func (b Bot) handleAssignTask(m *tb.Message) {
	log.Printf("Assign")
	if len(currentCommand) == 0 {
		currentCommand = map[string]string{}
	}
	currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = "assign_task"
	if !m.IsReply() {
		log.Printf("Not reply anything")
		b.bot.Reply(m, "Which task you want to assign task for? Send taskID and @mention an user to assign.")
	} else {
		task := m.ReplyTo
		taskID, err := strconv.Atoi(strings.Split(task.Text, " ")[0])
		if err != nil {
			b.bot.Reply(m, "Cannot get task to assign")
		}
		b.assignTask(taskID, m)
	}
	// self.bot.Send(m.Chat, fmt.Sprintf("Who do you want to assign? @mention an user to assign."))
}

func (b Bot) assignTask(currentTask int, m *tb.Message) {
	_, exist := currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)]
	if exist {
		currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = ""
	}
	task, err := b.storage.GetTask(currentTask)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot assign task: %s", err.Error()))
		return
	}
	entities := m.Entities
	assignee := ""
	for _, entity := range entities {
		if entity.Type == tb.EntityMention || entity.Type == tb.EntityTMention {
			if entity.User != nil {
				assignee = entity.User.Username
			} else {
				assignee = m.Payload
			}
		}
	}
	task.Assigned = assignee
	err = b.storage.UpdateTask(task)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot assigntask: %s", err.Error()))
		return
	}
	b.bot.Send(m.Chat, fmt.Sprintf("Task *%s* is assigned to *%s* successfully", task.Title, assignee), &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	})
}

func (b Bot) handlePin(m *tb.Message) {
	if m.IsReply() {
		// update pin message
		pinMessage := m.ReplyTo
		err := b.storage.UpdatePinMessage(pinMessage.Text, m.Chat.ID)
		log.Printf("Error: %+v", err)
		if err != nil {
			b.bot.Reply(m, fmt.Sprintf("Cannot pin message: %s", err.Error()))
		} else {
			b.bot.Reply(m, fmt.Sprintf("Update pin message successfully"))
		}
	} else {
		// show pin message
		pinMessage, err := b.storage.GetPinMessage(m.Chat.ID)
		log.Printf("Pin message: %s", pinMessage)
		if err != nil {
			b.bot.Reply(m, fmt.Sprintf("Cannot show pin message: %s", err.Error()))
		} else {
			if pinMessage != "" {
				b.bot.Reply(m, fmt.Sprintf("Pined message: %s", pinMessage))
			} else {
				b.bot.Reply(m, fmt.Sprintf("There is no pinned message yet"))
			}
		}
	}
}

func (b Bot) setDeadline(taskID int, deadline string, m *tb.Message) {
	task, err := b.storage.GetTask(taskID)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot set task deadline: %s", err.Error()))
		return
	}
	task.Deadline = deadline
	err = b.storage.UpdateTask(task)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot set task deadline: %s", err.Error()))
		return
	}
	b.bot.Send(m.Chat, fmt.Sprintf("Task *%s* deadline set to *%s* successfully", task.Title, deadline), &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	})
}

func (b Bot) setStatus(taskID int, status string, m *tb.Message) {
	task, err := b.storage.GetTask(currentTask)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot set status task: %s", err.Error()))
		return
	}
	task.Status = status
	err = b.storage.UpdateTask(task)
	if err != nil {
		b.bot.Send(m.Chat, fmt.Sprintf("Cannot set status task: %s", err.Error()))
		return
	}
	b.bot.Send(m.Chat, fmt.Sprintf("Task *%s* status set to *%s* successfully", task.Title, status), &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	})
}

func (b Bot) handleSetDeadline(m *tb.Message) {
	if !m.IsReply() {
		b.bot.Reply(m, fmt.Sprintf("You should reply to a task to set deadline"))
	} else {
		task := m.ReplyTo
		taskID, err := strconv.Atoi(strings.Split(task.Text, " ")[0])
		if err != nil {
			b.bot.Reply(m, fmt.Sprintf("Cannot get task ID to set deadline to"))
		}
		deadline := strings.Split(m.Text, " ")[1]
		b.setDeadline(taskID, deadline, m)
	}
}

func (b Bot) handleSetStatus(m *tb.Message) {
	if !m.IsReply() {
		b.bot.Reply(m, fmt.Sprintf("You should reply to a task to set status"))
	} else {
		task := m.ReplyTo
		taskID, err := strconv.Atoi(strings.Split(task.Text, " ")[0])
		if err != nil {
			b.bot.Reply(m, fmt.Sprintf("Cannot get task ID to set status to"))
		}
		status := strings.Split(m.Text, " ")[1]
		b.setStatus(taskID, status, m)
	}
}

func (b Bot) handleText(m *tb.Message) {
	command, exist := currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)]
	if !exist {
		return
	}
	log.Printf("Command: %s", command)
	switch command {
	case "create_task":
		b.saveTask(m)
		currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = ""
	case "create_project":
		b.saveProject(m.Text, m)
		currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = ""
	case "assign_task":
		log.Printf("Current task: %+v", currentTask)
		b.assignTask(currentTask, m)
		currentCommand[fmt.Sprintf("%d_%d", m.Sender.ID, m.Chat.ID)] = ""
	}
}
