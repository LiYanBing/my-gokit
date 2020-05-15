package api_interface

import "time"

type InterfaceStatus string

const (
	InterfaceStatusDone   InterfaceStatus = "done"
	InterfaceStatusUndone InterfaceStatus = "undone"
)

type InterfaceType string

const (
	InterfaceTypeStatic InterfaceType = "static"
	InterfaceTypeVar    InterfaceType = "var"
)

type InterfaceMethod string

const (
	InterfaceMethodGet     InterfaceMethod = "GET"
	InterfaceMethodPut     InterfaceMethod = "PUT"
	InterfaceMethodPost    InterfaceMethod = "POST"
	InterfaceMethodDelete  InterfaceMethod = "DELETE"
	InterfaceMethodOptions InterfaceMethod = "OPTIONS"
	InterfaceMethodHead    InterfaceMethod = "HEAD"
	InterfaceMethodPatch   InterfaceMethod = "PATCH"
)

type Interface struct {
	ID      bson.id         `json:"_id" bson:"_id"`
	EditUID int64           `json:"edit_uid" bson:"edit_uid"`
	Status  InterfaceStatus `json:"status" bson:"status"`
	Type    InterfaceType   `json:"type" bson:"type"`
	Tags    []string        `json:"tags" bson:"tags"`
	Method  InterfaceMethod `json:"method" bson:"method"`
	Title   string          `json:"title" bson:"title"`
	Path    string          `json:"path" bson:"path"`
	UserID  int64           `json:"user_id" bson:"user_id"`
	AddTime time.Time       `json:"add_time" bson:"add_time"`
	UpTime  time.Time       `json:"up_time" bson:"up_time"`
}

/**
"_id" : 101,
"edit_uid" : 0,
"status" : "undone",
"type" : "static",
"req_body_is_json_schema" : true,
"res_body_is_json_schema" : true,
"api_opened" : false,
"index" : 0,
"tag" : [
	"用户管理"
],
"method" : "POST",
"title" : "批量为用户取消标签",
"path" : "/tags/members/batchuntagging",
"req_params" : [ ],
"req_body_form" : [ ],
"req_headers" : [
	{
		"required" : "1",
		"_id" : ObjectId("5ebcb35b36d3051742ad0c5e"),
		"name" : "Content-Type",
		"value" : "application/json"
	}
],
"req_query" : [ ],
"req_body_type" : "json",
"res_body_type" : "json",
"res_body" : "{\n  \"type\": \"object\",\n  \"$$ref\": \"#/definitions/wechatCGIUserUntaggingResponse\"\n}",
"req_body_other" : "{\n  \"type\": \"object\",\n  \"properties\": {\n    \"openid_list\": {\n      \"type\": \"array\",\n      \"items\": {\n        \"type\": \"string\"\n      }\n    },\n    \"tagid\": {\n      \"type\": \"string\",\n      \"format\": \"int64\"\n    }\n  },\n  \"title\": \"CGIUserUntagging\",\n  \"$$ref\": \"#/definitions/wechatCGIUserUntaggingRequest\"\n}",
"project_id" : 18,
"catid" : 25,
"query_path" : {
	"path" : "/tags/members/batchuntagging",
	"params" : [ ]
},
"uid" : 11,
"add_time" : 1589424987,
"up_time" : 1589424987,
"__v" : 0
*/
