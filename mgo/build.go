package mgo

import "go.mongodb.org/mongo-driver/bson"

func BuildUpdateFields(in interface{}, selects, omits []string) (bson.M, error) {
	// 将结构体转换为 bson.M
	data, err := bson.Marshal(in)
	if err != nil {
		return nil, err
	}
	var updateFields bson.M
	err = bson.Unmarshal(data, &updateFields)
	if err != nil {
		return nil, err
	}
	if selects != nil {
		// 选择需要更新的字段
		for k := range updateFields {
			hit := false
			for _, v := range selects {
				if k == v {
					hit = true
					break
				}
			}
			if !hit {
				delete(updateFields, k)
			}
		}
	}
	// 删除不需要更新的字段
	for _, v := range omits {
		delete(updateFields, v)
	}
	return updateFields, nil
}
