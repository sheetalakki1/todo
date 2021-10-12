package model

type Todo struct {
	TYPE string `bson:"type"`
}

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error,omitempty"`
}
type UpdateResult struct {
	ID   string `json:"id"`
	TYPE string `json:"title"`
}
type DeleteResult struct {
	ID string `json:"id"`
}
