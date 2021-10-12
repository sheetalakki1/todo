package repository

import (
	"encoding/json"
	"go_todo/constant"
	"go_todo/logger"
	"go_todo/model"
	"go_todo/redis"
	"net/http"
	"os"

	kaleyra_bson "bitbucket.org/kaleyra/mongo-sdk/bson"
	"bitbucket.org/kaleyra/mongo-sdk/mongo"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongo_bson "gopkg.in/mgo.v2/bson"
)

var (
	collection  *mongo.Collection
	postCache   redis.PostCache = redis.NewRedisCache(os.Getenv("REDIS_URL"), 0, 10)
	sugarLogger                 = logger.InitLogger()
)

func init() {
	err := godotenv.Load(constant.Envfile)
	if err != nil {
		sugarLogger.Errorf("Failed to load the env file : Error = %s", err.Error())
		return
	}

	db := mongo.URI{
		Username: "",
		Password: "",
		Host:     os.Getenv("MONGO_HOST"),
		DB:       os.Getenv("DATABASE_NAME"),
		Port:     os.Getenv("MONGO_PORT"),
	}

	sugarLogger = logger.InitLogger()
	client, err := mongo.NewClient(db)
	if err != nil {
		sugarLogger.Errorf("Failed to connect to mongodb = %s", err.Error())
		return
	}

	zap.L().Info("Connected to MongoDB!")
	collection = client.Collection(os.Getenv("COLLECTION_NAME"))
	zap.L().Info("Collection instance created!")
}

func GetTodoById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var result *model.Todo
	if !mongo_bson.IsObjectIdHex(params[constant.ID]) {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to retrieve todo",
			Error:   "Invalid ID provided",
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}

	id, _ := primitive.ObjectIDFromHex(params[constant.ID])
	post := postCache.Get(id.String())
	if post == nil {
		filter := kaleyra_bson.D{{Key: constant.Key, Value: id}}
		result, err := collection.FindOne(filter)
		if err != nil {
			responseData := &model.Response{
				Status:  http.StatusInternalServerError,
				Message: "Failed to retrieve todo",
				Error:   err,
			}
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(responseData)
			if err != nil {
				sugarLogger.Error(err)
				return
			}
		}
		postCache.Set(id.String(), result)
	}

	result = postCache.Get(id.String())
	responseData := &model.Response{
		Status:  http.StatusOK,
		Message: "Retrieved todo",
		Data:    result,
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(responseData)
	if err != nil {
		sugarLogger.Error(err)
		return
	}
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	if !mongo_bson.IsObjectIdHex(params[constant.ID]) {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to delete todo",
			Error:   "Invalid ID provided",
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}

	id, _ := primitive.ObjectIDFromHex(params[constant.ID])
	filter := kaleyra_bson.D{{Key: constant.Key, Value: id}}
	res, err := collection.DeleteOne(filter)
	if err != nil {
		responseData := &model.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to delete todo",
			Error:   err,
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}

	if res == 0 {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "ID not found",
			Error:   err,
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}

	deleteData := &model.DeleteResult{
		ID: id.Hex(),
	}

	responseData := &model.Response{
		Status:  http.StatusOK,
		Message: "Todo deleted successfully",
		Data:    deleteData,
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		sugarLogger.Error(err)
		return
	}
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todos model.Todo
	er := json.NewDecoder(r.Body).Decode(&todos)
	if er != nil {
		sugarLogger.Error(er)
	}

	filter := kaleyra_bson.D{{Key: "type", Value: todos.TYPE}}
	result, err := collection.FindOne(filter)

	if result != nil {
		responseData := &model.Response{
			Status: http.StatusBadRequest,
			Error:  "Todo already exists",
		}
		json.NewEncoder(w).Encode(responseData)
		return
	}
	_, err = collection.InsertOne(todos)
	if err != nil {
		responseData := &model.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to insert todo",
			Error:   err,
		}
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
	}

	responseData := &model.Response{
		Status:  http.StatusCreated,
		Message: "Todo inserted successfully",
		Data:    todos,
	}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		sugarLogger.Error(err)
		return
	}
}

func FetchTodos(w http.ResponseWriter, r *http.Request) {
	res, err := collection.FindAll()
	if err != nil {
		responseData := &model.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to retrieve todos",
			Error:   err,
		}
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
	}

	responseData := &model.Response{
		Status:  http.StatusOK,
		Message: "Retrieved todos",
		Data:    res,
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		sugarLogger.Error(err)
		return
	}

}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	if !mongo_bson.IsObjectIdHex(params[constant.ID]) {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to update todo",
			Error:   "Invalid ID provided",
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}
	id, _ := primitive.ObjectIDFromHex(params[constant.ID])
	var t *model.Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to update todo",
			Error:   err,
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}
	if t.TYPE == "" {
		responseData := &model.Response{
			Status:  http.StatusBadRequest,
			Message: "Failed to update todo",
			Error:   "Empty title provided",
		}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}
	filter := kaleyra_bson.D{{Key: constant.Key, Value: id}}
	update := model.Todo{
		TYPE: t.TYPE,
	}
	_, err = collection.UpsertOne(filter, update)
	if err != nil {
		responseData := &model.Response{
			Status:  http.StatusInternalServerError,
			Message: "Failed to update todo",
			Error:   err,
		}
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(responseData)
		if err != nil {
			sugarLogger.Error(err)
			return
		}
		return
	}
	insertResult := &model.UpdateResult{
		ID:   id.Hex(),
		TYPE: t.TYPE,
	}
	responseData := &model.Response{
		Status:  http.StatusOK,
		Message: "Todo updated successfully",
		Data:    insertResult,
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		sugarLogger.Error(err)
		return
	}

}
