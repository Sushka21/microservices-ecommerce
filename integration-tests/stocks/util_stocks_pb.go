package stocks

import (
	"context"
	reflect "reflect"
	sync "sync"

	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/descriptorpb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SetStockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sku   uint32 `protobuf:"varint,1,opt,name=sku,proto3" json:"sku,omitempty"`
	Count uint64 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *SetStockRequest) Reset() {
	*x = SetStockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetStockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetStockRequest) ProtoMessage() {}

func (x *SetStockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetStockRequest.ProtoReflect.Descriptor instead.
func (*SetStockRequest) Descriptor() ([]byte, []int) {
	return file_loms_api_stocks_v1_stocks_proto_rawDescGZIP(), []int{0}
}

func (x *SetStockRequest) GetSku() uint32 {
	if x != nil {
		return x.Sku
	}
	return 0
}

func (x *SetStockRequest) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

type GetStockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sku uint32 `protobuf:"varint,1,opt,name=sku,proto3" json:"sku,omitempty"`
}

func (x *GetStockRequest) Reset() {
	*x = GetStockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStockRequest) ProtoMessage() {}

func (x *GetStockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStockRequest.ProtoReflect.Descriptor instead.
func (*GetStockRequest) Descriptor() ([]byte, []int) {
	return file_loms_api_stocks_v1_stocks_proto_rawDescGZIP(), []int{1}
}

func (x *GetStockRequest) GetSku() uint32 {
	if x != nil {
		return x.Sku
	}
	return 0
}

type GetStockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count uint64 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *GetStockResponse) Reset() {
	*x = GetStockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStockResponse) ProtoMessage() {}

func (x *GetStockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_loms_api_stocks_v1_stocks_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStockResponse.ProtoReflect.Descriptor instead.
func (*GetStockResponse) Descriptor() ([]byte, []int) {
	return file_loms_api_stocks_v1_stocks_proto_rawDescGZIP(), []int{2}
}

func (x *GetStockResponse) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

var File_loms_api_stocks_v1_stocks_proto protoreflect.FileDescriptor

var file_loms_api_stocks_v1_stocks_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6c, 0x6f, 0x6d, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x74, 0x6f, 0x63, 0x6b,
	0x73, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x67,
	0x6f, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73,
	0x2e, 0x6d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x65,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x72, 0x63, 0x65, 0x2e, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x2e,
	0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a, 0x0f, 0x53, 0x65, 0x74, 0x53,
	0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x03, 0x73,
	0x6b, 0x75, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x28,
	0x01, 0x52, 0x03, 0x73, 0x6b, 0x75, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x2c, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x19, 0x0a, 0x03, 0x73, 0x6b, 0x75, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x2a, 0x02, 0x28, 0x01, 0x52, 0x03, 0x73, 0x6b, 0x75, 0x22, 0x28, 0x0a, 0x10, 0x47, 0x65,
	0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x32, 0xe6, 0x02, 0x0a, 0x06, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x12,
	0xca, 0x01, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12, 0x50, 0x2e, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x67, 0x6f, 0x72, 0x6f, 0x75,
	0x74, 0x69, 0x6e, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x2e, 0x6d, 0x69, 0x63,
	0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x65, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x72, 0x63, 0x65, 0x2e, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x51,
	0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x67, 0x6f, 0x72,
	0x6f, 0x75, 0x74, 0x69, 0x6e, 0x65, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x2e, 0x6d,
	0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x65, 0x63, 0x6f,
	0x6d, 0x6d, 0x65, 0x72, 0x63, 0x65, 0x2e, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x19, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x13, 0x22, 0x0e, 0x2f, 0x76, 0x31, 0x2f, 0x73,
	0x74, 0x6f, 0x63, 0x6b, 0x2f, 0x69, 0x6e, 0x66, 0x6f, 0x3a, 0x01, 0x2a, 0x12, 0x8e, 0x01, 0x0a,
	0x08, 0x53, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x63, 0x6b, 0x12, 0x50, 0x2e, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x67, 0x6f, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e,
	0x65, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x2e, 0x6d, 0x69, 0x63, 0x72, 0x6f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x72, 0x63,
	0x65, 0x2e, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x74, 0x53,
	0x74, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x3a, 0x01, 0x2a, 0x22, 0x0d,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x2f, 0x73, 0x65, 0x74, 0x42, 0x5f, 0x5a,
	0x5d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69, 0x67, 0x6f, 0x72,
	0x6f, 0x75, 0x74, 0x69, 0x6e, 0x65, 0x2d, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x2f, 0x6d,
	0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x65, 0x63, 0x6f,
	0x6d, 0x6d, 0x65, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x64, 0x2f, 0x6c, 0x6f, 0x6d, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x74,
	0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x73, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_loms_api_stocks_v1_stocks_proto_rawDescOnce sync.Once
	file_loms_api_stocks_v1_stocks_proto_rawDescData = file_loms_api_stocks_v1_stocks_proto_rawDesc
)

func file_loms_api_stocks_v1_stocks_proto_rawDescGZIP() []byte {
	file_loms_api_stocks_v1_stocks_proto_rawDescOnce.Do(func() {
		file_loms_api_stocks_v1_stocks_proto_rawDescData = protoimpl.X.CompressGZIP(file_loms_api_stocks_v1_stocks_proto_rawDescData)
	})
	return file_loms_api_stocks_v1_stocks_proto_rawDescData
}

var file_loms_api_stocks_v1_stocks_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_loms_api_stocks_v1_stocks_proto_goTypes = []interface{}{
	(*SetStockRequest)(nil),  // 0: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.SetStockRequest
	(*GetStockRequest)(nil),  // 1: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.GetStockRequest
	(*GetStockResponse)(nil), // 2: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.GetStockResponse
	(*emptypb.Empty)(nil),    // 3: google.protobuf.Empty
}
var file_loms_api_stocks_v1_stocks_proto_depIdxs = []int32{
	1, // 0: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks.GetStock:input_type -> github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.GetStockRequest
	0, // 1: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks.SetStock:input_type -> github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.SetStockRequest
	2, // 2: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks.GetStock:output_type -> github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.GetStockResponse
	3, // 3: github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks.SetStock:output_type -> google.protobuf.Empty
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_loms_api_stocks_v1_stocks_proto_init() }
func file_loms_api_stocks_v1_stocks_proto_init() {
	if File_loms_api_stocks_v1_stocks_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_loms_api_stocks_v1_stocks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetStockRequest); i {
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
		file_loms_api_stocks_v1_stocks_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStockRequest); i {
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
		file_loms_api_stocks_v1_stocks_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStockResponse); i {
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
			RawDescriptor: file_loms_api_stocks_v1_stocks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_loms_api_stocks_v1_stocks_proto_goTypes,
		DependencyIndexes: file_loms_api_stocks_v1_stocks_proto_depIdxs,
		MessageInfos:      file_loms_api_stocks_v1_stocks_proto_msgTypes,
	}.Build()
	File_loms_api_stocks_v1_stocks_proto = out.File
	file_loms_api_stocks_v1_stocks_proto_rawDesc = nil
	file_loms_api_stocks_v1_stocks_proto_goTypes = nil
	file_loms_api_stocks_v1_stocks_proto_depIdxs = nil
}

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StocksClient is the client API for Stocks service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StocksClient interface {
	GetStock(ctx context.Context, in *GetStockRequest, opts ...grpc.CallOption) (*GetStockResponse, error)
	SetStock(ctx context.Context, in *SetStockRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type stocksClient struct {
	cc grpc.ClientConnInterface
}

func NewStocksClient(cc grpc.ClientConnInterface) StocksClient {
	return &stocksClient{cc}
}

func (c *stocksClient) GetStock(ctx context.Context, in *GetStockRequest, opts ...grpc.CallOption) (*GetStockResponse, error) {
	out := new(GetStockResponse)
	err := c.cc.Invoke(ctx, "/github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks/GetStock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stocksClient) SetStock(ctx context.Context, in *SetStockRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks/SetStock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StocksServer is the server API for Stocks service.
// All implementations should embed UnimplementedStocksServer
// for forward compatibility
type StocksServer interface {
	GetStock(context.Context, *GetStockRequest) (*GetStockResponse, error)
	SetStock(context.Context, *SetStockRequest) (*emptypb.Empty, error)
}

// UnimplementedStocksServer should be embedded to have forward compatible implementations.
type UnimplementedStocksServer struct {
}

func (UnimplementedStocksServer) GetStock(context.Context, *GetStockRequest) (*GetStockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStock not implemented")
}
func (UnimplementedStocksServer) SetStock(context.Context, *SetStockRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetStock not implemented")
}

// UnsafeStocksServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StocksServer will
// result in compilation errors.
type UnsafeStocksServer interface {
	mustEmbedUnimplementedStocksServer()
}

func RegisterStocksServer(s grpc.ServiceRegistrar, srv StocksServer) {
	s.RegisterService(&Stocks_ServiceDesc, srv)
}

func _Stocks_GetStock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StocksServer).GetStock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks/GetStock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StocksServer).GetStock(ctx, req.(*GetStockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Stocks_SetStock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetStockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StocksServer).SetStock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks/SetStock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StocksServer).SetStock(ctx, req.(*SetStockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Stocks_ServiceDesc is the grpc.ServiceDesc for Stocks service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Stocks_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "github.com.igoroutine_courses.microservices.ecommerce.stocks.v1.Stocks",
	HandlerType: (*StocksServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetStock",
			Handler:    _Stocks_GetStock_Handler,
		},
		{
			MethodName: "SetStock",
			Handler:    _Stocks_SetStock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "loms/api/stocks/v1/stocks.proto",
}



