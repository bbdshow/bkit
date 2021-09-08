package mongo

var Indexes []Index

// CreateIndex add indexes to Indexes, waiting create
func CreateIndex(indexes []Index) {
	Indexes = append(Indexes, indexes...)
}
