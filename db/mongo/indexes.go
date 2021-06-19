package mongo

var Indexes []Index

// 将要创建索引占存
func CreateIndex(indexes []Index) {
	Indexes = append(Indexes, indexes...)
}
