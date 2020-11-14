// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0-devel
// 	protoc        v3.12.3
// source: basic.proto

package ss

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SS_COMMON_RESULT int32

const (
	SS_COMMON_RESULT_SUCCESS SS_COMMON_RESULT = 0
	SS_COMMON_RESULT_FAILED  SS_COMMON_RESULT = 1
	SS_COMMON_RESULT_NOEXIST SS_COMMON_RESULT = 2
)

// Enum value maps for SS_COMMON_RESULT.
var (
	SS_COMMON_RESULT_name = map[int32]string{
		0: "SUCCESS",
		1: "FAILED",
		2: "NOEXIST",
	}
	SS_COMMON_RESULT_value = map[string]int32{
		"SUCCESS": 0,
		"FAILED":  1,
		"NOEXIST": 2,
	}
)

func (x SS_COMMON_RESULT) Enum() *SS_COMMON_RESULT {
	p := new(SS_COMMON_RESULT)
	*p = x
	return p
}

func (x SS_COMMON_RESULT) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SS_COMMON_RESULT) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[0].Descriptor()
}

func (SS_COMMON_RESULT) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[0]
}

func (x SS_COMMON_RESULT) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SS_COMMON_RESULT.Descriptor instead.
func (SS_COMMON_RESULT) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{0}
}

type SS_COMMON_REASON int32

const (
	SS_COMMON_REASON_REASON_TICK   SS_COMMON_REASON = 0 //ticker save
	SS_COMMON_REASON_REASON_UPDATE SS_COMMON_REASON = 1 //activate save
	SS_COMMON_REASON_REASON_EXIT   SS_COMMON_REASON = 2 //server exit
)

// Enum value maps for SS_COMMON_REASON.
var (
	SS_COMMON_REASON_name = map[int32]string{
		0: "REASON_TICK",
		1: "REASON_UPDATE",
		2: "REASON_EXIT",
	}
	SS_COMMON_REASON_value = map[string]int32{
		"REASON_TICK":   0,
		"REASON_UPDATE": 1,
		"REASON_EXIT":   2,
	}
)

func (x SS_COMMON_REASON) Enum() *SS_COMMON_REASON {
	p := new(SS_COMMON_REASON)
	*p = x
	return p
}

func (x SS_COMMON_REASON) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SS_COMMON_REASON) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[1].Descriptor()
}

func (SS_COMMON_REASON) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[1]
}

func (x SS_COMMON_REASON) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SS_COMMON_REASON.Descriptor instead.
func (SS_COMMON_REASON) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{1}
}

type SS_COMMON_TYPE int32

const (
	SS_COMMON_TYPE_COMM_TYPE_NORMAL  SS_COMMON_TYPE = 0
	SS_COMMON_TYPE_COMM_TYPE_HISTORY SS_COMMON_TYPE = 1
)

// Enum value maps for SS_COMMON_TYPE.
var (
	SS_COMMON_TYPE_name = map[int32]string{
		0: "COMM_TYPE_NORMAL",
		1: "COMM_TYPE_HISTORY",
	}
	SS_COMMON_TYPE_value = map[string]int32{
		"COMM_TYPE_NORMAL":  0,
		"COMM_TYPE_HISTORY": 1,
	}
)

func (x SS_COMMON_TYPE) Enum() *SS_COMMON_TYPE {
	p := new(SS_COMMON_TYPE)
	*p = x
	return p
}

func (x SS_COMMON_TYPE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SS_COMMON_TYPE) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[2].Descriptor()
}

func (SS_COMMON_TYPE) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[2]
}

func (x SS_COMMON_TYPE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SS_COMMON_TYPE.Descriptor instead.
func (SS_COMMON_TYPE) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{2}
}

type SS_OFFLINE_INFO_TYPE int32

const (
	SS_OFFLINE_INFO_TYPE_OFT_KICK_GROUP SS_OFFLINE_INFO_TYPE = 0 //kick out group by master <type|grp_id|grp_name|kick_ts>
)

// Enum value maps for SS_OFFLINE_INFO_TYPE.
var (
	SS_OFFLINE_INFO_TYPE_name = map[int32]string{
		0: "OFT_KICK_GROUP",
	}
	SS_OFFLINE_INFO_TYPE_value = map[string]int32{
		"OFT_KICK_GROUP": 0,
	}
)

func (x SS_OFFLINE_INFO_TYPE) Enum() *SS_OFFLINE_INFO_TYPE {
	p := new(SS_OFFLINE_INFO_TYPE)
	*p = x
	return p
}

func (x SS_OFFLINE_INFO_TYPE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SS_OFFLINE_INFO_TYPE) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[3].Descriptor()
}

func (SS_OFFLINE_INFO_TYPE) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[3]
}

func (x SS_OFFLINE_INFO_TYPE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SS_OFFLINE_INFO_TYPE.Descriptor instead.
func (SS_OFFLINE_INFO_TYPE) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{3}
}

type SS_GROUP_INFO_FIELD int32

const (
	SS_GROUP_INFO_FIELD_GRP_FIELD_ALL  SS_GROUP_INFO_FIELD = 0
	SS_GROUP_INFO_FIELD_GRP_FIELD_SNAP SS_GROUP_INFO_FIELD = 1
)

// Enum value maps for SS_GROUP_INFO_FIELD.
var (
	SS_GROUP_INFO_FIELD_name = map[int32]string{
		0: "GRP_FIELD_ALL",
		1: "GRP_FIELD_SNAP",
	}
	SS_GROUP_INFO_FIELD_value = map[string]int32{
		"GRP_FIELD_ALL":  0,
		"GRP_FIELD_SNAP": 1,
	}
)

func (x SS_GROUP_INFO_FIELD) Enum() *SS_GROUP_INFO_FIELD {
	p := new(SS_GROUP_INFO_FIELD)
	*p = x
	return p
}

func (x SS_GROUP_INFO_FIELD) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SS_GROUP_INFO_FIELD) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[4].Descriptor()
}

func (SS_GROUP_INFO_FIELD) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[4]
}

func (x SS_GROUP_INFO_FIELD) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SS_GROUP_INFO_FIELD.Descriptor instead.
func (SS_GROUP_INFO_FIELD) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{4}
}

type GROUP_ATTR_TYPE int32

const (
	GROUP_ATTR_TYPE_GRP_ATTR_VISIBLE   GROUP_ATTR_TYPE = 0 //cound be shown on panel
	GROUP_ATTR_TYPE_GRP_ATTR_INVISIBLE GROUP_ATTR_TYPE = 1 //only be searched by group id
	GROUP_ATTR_TYPE_GRP_ATTR_DESC      GROUP_ATTR_TYPE = 2 //change group desc
)

// Enum value maps for GROUP_ATTR_TYPE.
var (
	GROUP_ATTR_TYPE_name = map[int32]string{
		0: "GRP_ATTR_VISIBLE",
		1: "GRP_ATTR_INVISIBLE",
		2: "GRP_ATTR_DESC",
	}
	GROUP_ATTR_TYPE_value = map[string]int32{
		"GRP_ATTR_VISIBLE":   0,
		"GRP_ATTR_INVISIBLE": 1,
		"GRP_ATTR_DESC":      2,
	}
)

func (x GROUP_ATTR_TYPE) Enum() *GROUP_ATTR_TYPE {
	p := new(GROUP_ATTR_TYPE)
	*p = x
	return p
}

func (x GROUP_ATTR_TYPE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GROUP_ATTR_TYPE) Descriptor() protoreflect.EnumDescriptor {
	return file_basic_proto_enumTypes[5].Descriptor()
}

func (GROUP_ATTR_TYPE) Type() protoreflect.EnumType {
	return &file_basic_proto_enumTypes[5]
}

func (x GROUP_ATTR_TYPE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GROUP_ATTR_TYPE.Descriptor instead.
func (GROUP_ATTR_TYPE) EnumDescriptor() ([]byte, []int) {
	return file_basic_proto_rawDescGZIP(), []int{5}
}

var File_basic_proto protoreflect.FileDescriptor

var file_basic_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x62, 0x61, 0x73, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x73,
	0x73, 0x2a, 0x38, 0x0a, 0x10, 0x53, 0x53, 0x5f, 0x43, 0x4f, 0x4d, 0x4d, 0x4f, 0x4e, 0x5f, 0x52,
	0x45, 0x53, 0x55, 0x4c, 0x54, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53,
	0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0b,
	0x0a, 0x07, 0x4e, 0x4f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x10, 0x02, 0x2a, 0x47, 0x0a, 0x10, 0x53,
	0x53, 0x5f, 0x43, 0x4f, 0x4d, 0x4d, 0x4f, 0x4e, 0x5f, 0x52, 0x45, 0x41, 0x53, 0x4f, 0x4e, 0x12,
	0x0f, 0x0a, 0x0b, 0x52, 0x45, 0x41, 0x53, 0x4f, 0x4e, 0x5f, 0x54, 0x49, 0x43, 0x4b, 0x10, 0x00,
	0x12, 0x11, 0x0a, 0x0d, 0x52, 0x45, 0x41, 0x53, 0x4f, 0x4e, 0x5f, 0x55, 0x50, 0x44, 0x41, 0x54,
	0x45, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x52, 0x45, 0x41, 0x53, 0x4f, 0x4e, 0x5f, 0x45, 0x58,
	0x49, 0x54, 0x10, 0x02, 0x2a, 0x3d, 0x0a, 0x0e, 0x53, 0x53, 0x5f, 0x43, 0x4f, 0x4d, 0x4d, 0x4f,
	0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x4d, 0x4d, 0x5f, 0x54,
	0x59, 0x50, 0x45, 0x5f, 0x4e, 0x4f, 0x52, 0x4d, 0x41, 0x4c, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11,
	0x43, 0x4f, 0x4d, 0x4d, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x48, 0x49, 0x53, 0x54, 0x4f, 0x52,
	0x59, 0x10, 0x01, 0x2a, 0x2a, 0x0a, 0x14, 0x53, 0x53, 0x5f, 0x4f, 0x46, 0x46, 0x4c, 0x49, 0x4e,
	0x45, 0x5f, 0x49, 0x4e, 0x46, 0x4f, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x12, 0x12, 0x0a, 0x0e, 0x4f,
	0x46, 0x54, 0x5f, 0x4b, 0x49, 0x43, 0x4b, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x10, 0x00, 0x2a,
	0x3c, 0x0a, 0x13, 0x53, 0x53, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x49, 0x4e, 0x46, 0x4f,
	0x5f, 0x46, 0x49, 0x45, 0x4c, 0x44, 0x12, 0x11, 0x0a, 0x0d, 0x47, 0x52, 0x50, 0x5f, 0x46, 0x49,
	0x45, 0x4c, 0x44, 0x5f, 0x41, 0x4c, 0x4c, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x47, 0x52, 0x50,
	0x5f, 0x46, 0x49, 0x45, 0x4c, 0x44, 0x5f, 0x53, 0x4e, 0x41, 0x50, 0x10, 0x01, 0x2a, 0x52, 0x0a,
	0x0f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x41, 0x54, 0x54, 0x52, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x12, 0x14, 0x0a, 0x10, 0x47, 0x52, 0x50, 0x5f, 0x41, 0x54, 0x54, 0x52, 0x5f, 0x56, 0x49, 0x53,
	0x49, 0x42, 0x4c, 0x45, 0x10, 0x00, 0x12, 0x16, 0x0a, 0x12, 0x47, 0x52, 0x50, 0x5f, 0x41, 0x54,
	0x54, 0x52, 0x5f, 0x49, 0x4e, 0x56, 0x49, 0x53, 0x49, 0x42, 0x4c, 0x45, 0x10, 0x01, 0x12, 0x11,
	0x0a, 0x0d, 0x47, 0x52, 0x50, 0x5f, 0x41, 0x54, 0x54, 0x52, 0x5f, 0x44, 0x45, 0x53, 0x43, 0x10,
	0x02, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_basic_proto_rawDescOnce sync.Once
	file_basic_proto_rawDescData = file_basic_proto_rawDesc
)

func file_basic_proto_rawDescGZIP() []byte {
	file_basic_proto_rawDescOnce.Do(func() {
		file_basic_proto_rawDescData = protoimpl.X.CompressGZIP(file_basic_proto_rawDescData)
	})
	return file_basic_proto_rawDescData
}

var file_basic_proto_enumTypes = make([]protoimpl.EnumInfo, 6)
var file_basic_proto_goTypes = []interface{}{
	(SS_COMMON_RESULT)(0),     // 0: ss.SS_COMMON_RESULT
	(SS_COMMON_REASON)(0),     // 1: ss.SS_COMMON_REASON
	(SS_COMMON_TYPE)(0),       // 2: ss.SS_COMMON_TYPE
	(SS_OFFLINE_INFO_TYPE)(0), // 3: ss.SS_OFFLINE_INFO_TYPE
	(SS_GROUP_INFO_FIELD)(0),  // 4: ss.SS_GROUP_INFO_FIELD
	(GROUP_ATTR_TYPE)(0),      // 5: ss.GROUP_ATTR_TYPE
}
var file_basic_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_basic_proto_init() }
func file_basic_proto_init() {
	if File_basic_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_basic_proto_rawDesc,
			NumEnums:      6,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_basic_proto_goTypes,
		DependencyIndexes: file_basic_proto_depIdxs,
		EnumInfos:         file_basic_proto_enumTypes,
	}.Build()
	File_basic_proto = out.File
	file_basic_proto_rawDesc = nil
	file_basic_proto_goTypes = nil
	file_basic_proto_depIdxs = nil
}
