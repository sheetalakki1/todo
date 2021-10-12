package redis

import "go_todo/model"

type PostCache interface {
	Set(key string, value map[string]interface{})
	Get(key string) *model.Todo
}
