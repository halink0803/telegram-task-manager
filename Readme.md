##Telegram Task Management Bot

[![Go Report Card](https://goreportcard.com/badge/github.com/halink0803/telegram-task-manager)](https://goreportcard.com/report/github.com/halink0803/telegram-task-manager)

[![Build Status](https://travis-ci.org/halink0803/telegram-task-manager.svg?branch=master)](https://travis-ci.org/halink0803/telegram-task-manager)


### Available commands
    list_projects - show all project available
    create_project - create a new project  
    set_default_project - set a default project for a conversation  
    current_project - show current project
    add_task - add new to a project  
    list_task - list tasks (all, not start, doing, done or by assignee)  
    mine - list your tasks  
    pin - Reply to a message to pin that message, not reply to show the pinned message
    assign - Reply to a task and mention a user to assign a task for that user (eg: /assign @halink0803)
    set_status - Reply to a task and provide status you want to set (eg: /set_status done)
    set_deadline - Reply to a task and provide a deadline to set deadline (eg: /set_dealine 12/04)
    detail - Reply to a task to show detail of that task
    discussion - Reply to a message