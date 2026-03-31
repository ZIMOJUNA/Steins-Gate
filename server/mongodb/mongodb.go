package mongodb

import (
	"context"
	"log"
	"server/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 全局客户端（整个项目共用一个）
var client *mongo.Client
var db *mongo.Database

// init 自动初始化
func init() {

	// 创建连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(config.Conf.MongoDB.Url)
	var err error
	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("❌ MongoDB 连接失败：", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB ping 失败：", err)
	}

	// 选择数据库
	db = client.Database(config.Conf.MongoDB.DBName)
	log.Println("✅ MongoDB 全局初始化完成！")
}

// GetClient 获取全局客户端
func GetClient() *mongo.Client {
	return client
}

// GetCollection 获取集合（表）
func GetCollection(collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

// ====================== 通用查询方法 ======================

// FindOne 查询单条数据
// collectionName: 集合名
// filter: 查询条件
// result: 接收结果的结构体指针
func FindOne(collectionName string, filter interface{}, result interface{}) error {
	coll := GetCollection(collectionName)
	return coll.FindOne(context.TODO(), filter).Decode(result)
}

// Find 查询多条数据
func Find(collectionName string, filter interface{}, results interface{}) error {
	coll := GetCollection(collectionName)
	ctx := context.TODO()

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, results)
}

// InsertOne 插入单条
func InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	coll := GetCollection(collectionName)
	return coll.InsertOne(context.TODO(), document)
}
