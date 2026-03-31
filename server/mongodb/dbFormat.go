package mongodb

// 定义结构体（对应 MongoDB 文档）
type User struct {
	Name string `bson:"name"`
	Age  int    `bson:"age"`
	City string `bson:"city"`
}
