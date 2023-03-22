// Copyright 2021-2023 The sacloud/autoscaler Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.12
// source: request.proto

package request

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ジョブのステータス
type ScalingJobStatus int32

const (
	ScalingJobStatus_JOB_UNKNOWN   ScalingJobStatus = 0 // 不明
	ScalingJobStatus_JOB_ACCEPTED  ScalingJobStatus = 1 // 受付済み
	ScalingJobStatus_JOB_RUNNING   ScalingJobStatus = 2 // 実行中
	ScalingJobStatus_JOB_DONE      ScalingJobStatus = 3 // 完了(ハンドラが処理を行った)
	ScalingJobStatus_JOB_CANCELED  ScalingJobStatus = 4 // 開始前に中断
	ScalingJobStatus_JOB_IGNORED   ScalingJobStatus = 5 // 無視(受け入れなかった)
	ScalingJobStatus_JOB_FAILED    ScalingJobStatus = 6 // 失敗/エラー
	ScalingJobStatus_JOB_DONE_NOOP ScalingJobStatus = 7 // 完了(ハンドラが何も処理しなかった)
)

// Enum value maps for ScalingJobStatus.
var (
	ScalingJobStatus_name = map[int32]string{
		0: "JOB_UNKNOWN",
		1: "JOB_ACCEPTED",
		2: "JOB_RUNNING",
		3: "JOB_DONE",
		4: "JOB_CANCELED",
		5: "JOB_IGNORED",
		6: "JOB_FAILED",
		7: "JOB_DONE_NOOP",
	}
	ScalingJobStatus_value = map[string]int32{
		"JOB_UNKNOWN":   0,
		"JOB_ACCEPTED":  1,
		"JOB_RUNNING":   2,
		"JOB_DONE":      3,
		"JOB_CANCELED":  4,
		"JOB_IGNORED":   5,
		"JOB_FAILED":    6,
		"JOB_DONE_NOOP": 7,
	}
)

func (x ScalingJobStatus) Enum() *ScalingJobStatus {
	p := new(ScalingJobStatus)
	*p = x
	return p
}

func (x ScalingJobStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ScalingJobStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_request_proto_enumTypes[0].Descriptor()
}

func (ScalingJobStatus) Type() protoreflect.EnumType {
	return &file_request_proto_enumTypes[0]
}

func (x ScalingJobStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ScalingJobStatus.Descriptor instead.
func (ScalingJobStatus) EnumDescriptor() ([]byte, []int) {
	return file_request_proto_rawDescGZIP(), []int{0}
}

// Scalingサービスのリクエストパラメータ
type ScalingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 呼び出し元を示すラベル値、Coreでの処理には影響しない。デフォルト値:
	// "default"
	Source string `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	// 操作対象のリソース名。リソース名にはCoreのコンフィギュレーションの中で定義した名前を指定する
	// 対応するリソース名がCoreで見つけられなかった場合はエラーを返す
	//
	// デフォルト値: "default"
	// デフォルト値を指定した場合はCoreのコンフィギュレーションで定義された先頭のリソースが操作対象となる
	ResourceName string `protobuf:"bytes,2,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// 希望するスケール(プランなど)につけた名前
	// 特定のスケールに一気にスケールを変更したい場合に指定する
	// 指定する名前はCoreのコンフィギュレーションで定義しておく必要がある
	DesiredStateName string `protobuf:"bytes,3,opt,name=desired_state_name,json=desiredStateName,proto3" json:"desired_state_name,omitempty"`
	// 同期的に処理を行うか
	Sync bool `protobuf:"varint,4,opt,name=sync,proto3" json:"sync,omitempty"`
}

func (x *ScalingRequest) Reset() {
	*x = ScalingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_request_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScalingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScalingRequest) ProtoMessage() {}

func (x *ScalingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_request_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScalingRequest.ProtoReflect.Descriptor instead.
func (*ScalingRequest) Descriptor() ([]byte, []int) {
	return file_request_proto_rawDescGZIP(), []int{0}
}

func (x *ScalingRequest) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *ScalingRequest) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *ScalingRequest) GetDesiredStateName() string {
	if x != nil {
		return x.DesiredStateName
	}
	return ""
}

func (x *ScalingRequest) GetSync() bool {
	if x != nil {
		return x.Sync
	}
	return false
}

// Scalingサービスのレスポンス
type ScalingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// スケールジョブのID
	// リクエストパラメータに応じてCoreがジョブを起動しIDを割り当てたもの
	ScalingJobId string `protobuf:"bytes,1,opt,name=scaling_job_id,json=scalingJobId,proto3" json:"scaling_job_id,omitempty"`
	// スケールジョブのステータス
	// Coreがリクエストを処理した段階のステータスを返す
	Status ScalingJobStatus `protobuf:"varint,2,opt,name=status,proto3,enum=autoscaler.ScalingJobStatus" json:"status,omitempty"`
	// Coreからのメッセージ
	// 何らかの事情でリクエストを受け付けられなかった場合の理由が記載される
	Message string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *ScalingResponse) Reset() {
	*x = ScalingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_request_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScalingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScalingResponse) ProtoMessage() {}

func (x *ScalingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_request_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScalingResponse.ProtoReflect.Descriptor instead.
func (*ScalingResponse) Descriptor() ([]byte, []int) {
	return file_request_proto_rawDescGZIP(), []int{1}
}

func (x *ScalingResponse) GetScalingJobId() string {
	if x != nil {
		return x.ScalingJobId
	}
	return ""
}

func (x *ScalingResponse) GetStatus() ScalingJobStatus {
	if x != nil {
		return x.Status
	}
	return ScalingJobStatus_JOB_UNKNOWN
}

func (x *ScalingResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_request_proto protoreflect.FileDescriptor

var file_request_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0a, 0x61, 0x75, 0x74, 0x6f, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x22, 0x8f, 0x01, 0x0a, 0x0e,
	0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x12, 0x64,
	0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x79, 0x6e,
	0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x73, 0x79, 0x6e, 0x63, 0x22, 0x87, 0x01,
	0x0a, 0x0f, 0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x24, 0x0a, 0x0e, 0x73, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x6a, 0x6f, 0x62,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x63, 0x61, 0x6c, 0x69,
	0x6e, 0x67, 0x4a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x34, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x61, 0x75, 0x74, 0x6f, 0x73, 0x63,
	0x61, 0x6c, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x4a, 0x6f, 0x62, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2a, 0x9a, 0x01, 0x0a, 0x10, 0x53, 0x63, 0x61, 0x6c,
	0x69, 0x6e, 0x67, 0x4a, 0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0f, 0x0a, 0x0b,
	0x4a, 0x4f, 0x42, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x10, 0x0a,
	0x0c, 0x4a, 0x4f, 0x42, 0x5f, 0x41, 0x43, 0x43, 0x45, 0x50, 0x54, 0x45, 0x44, 0x10, 0x01, 0x12,
	0x0f, 0x0a, 0x0b, 0x4a, 0x4f, 0x42, 0x5f, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02,
	0x12, 0x0c, 0x0a, 0x08, 0x4a, 0x4f, 0x42, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x03, 0x12, 0x10,
	0x0a, 0x0c, 0x4a, 0x4f, 0x42, 0x5f, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x45, 0x44, 0x10, 0x04,
	0x12, 0x0f, 0x0a, 0x0b, 0x4a, 0x4f, 0x42, 0x5f, 0x49, 0x47, 0x4e, 0x4f, 0x52, 0x45, 0x44, 0x10,
	0x05, 0x12, 0x0e, 0x0a, 0x0a, 0x4a, 0x4f, 0x42, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10,
	0x06, 0x12, 0x11, 0x0a, 0x0d, 0x4a, 0x4f, 0x42, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x5f, 0x4e, 0x4f,
	0x4f, 0x50, 0x10, 0x07, 0x32, 0x90, 0x01, 0x0a, 0x0e, 0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x02, 0x55, 0x70, 0x12, 0x1a, 0x2e,
	0x61, 0x75, 0x74, 0x6f, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6c, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x61, 0x75, 0x74, 0x6f,
	0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3f, 0x0a, 0x04, 0x44, 0x6f, 0x77, 0x6e, 0x12, 0x1a,
	0x2e, 0x61, 0x75, 0x74, 0x6f, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6c,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x61, 0x75, 0x74,
	0x6f, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x61, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x61, 0x75,
	0x74, 0x6f, 0x73, 0x63, 0x61, 0x6c, 0x65, 0x72, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_request_proto_rawDescOnce sync.Once
	file_request_proto_rawDescData = file_request_proto_rawDesc
)

func file_request_proto_rawDescGZIP() []byte {
	file_request_proto_rawDescOnce.Do(func() {
		file_request_proto_rawDescData = protoimpl.X.CompressGZIP(file_request_proto_rawDescData)
	})
	return file_request_proto_rawDescData
}

var file_request_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_request_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_request_proto_goTypes = []interface{}{
	(ScalingJobStatus)(0),   // 0: autoscaler.ScalingJobStatus
	(*ScalingRequest)(nil),  // 1: autoscaler.ScalingRequest
	(*ScalingResponse)(nil), // 2: autoscaler.ScalingResponse
}
var file_request_proto_depIdxs = []int32{
	0, // 0: autoscaler.ScalingResponse.status:type_name -> autoscaler.ScalingJobStatus
	1, // 1: autoscaler.ScalingService.Up:input_type -> autoscaler.ScalingRequest
	1, // 2: autoscaler.ScalingService.Down:input_type -> autoscaler.ScalingRequest
	2, // 3: autoscaler.ScalingService.Up:output_type -> autoscaler.ScalingResponse
	2, // 4: autoscaler.ScalingService.Down:output_type -> autoscaler.ScalingResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_request_proto_init() }
func file_request_proto_init() {
	if File_request_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_request_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScalingRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_request_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScalingResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_request_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_request_proto_goTypes,
		DependencyIndexes: file_request_proto_depIdxs,
		EnumInfos:         file_request_proto_enumTypes,
		MessageInfos:      file_request_proto_msgTypes,
	}.Build()
	File_request_proto = out.File
	file_request_proto_rawDesc = nil
	file_request_proto_goTypes = nil
	file_request_proto_depIdxs = nil
}
