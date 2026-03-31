package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. 构建 MongoDB 连接 URI（和你的 docker 完全匹配）
	uri := "mongodb://admin:123456@localhost:27017/"

	// 2. 创建连接客户端
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 检查连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB 连接失败：", err)
	}

	fmt.Println("✅ MongoDB 连接成功！")

	// 4. 选择数据库 + 集合
	db := client.Database("testdb")         // 数据库名
	userCollection := db.Collection("user") // 集合名（=表）

	// -------------------
	// 插入一条数据
	// -------------------
	user := User{
		Name: "小明",
		Age:  20,
		City: "北京",
	}
	insertResult, err := userCollection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("插入数据ID：", insertResult.InsertedID)

	// -------------------
	// 查询一条数据
	// -------------------
	var result User
	err = userCollection.FindOne(ctx, User{Name: "小明"}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("查询到的数据：", result)
}
