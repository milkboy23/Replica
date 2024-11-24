// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.12.4
// source: auction.proto

package __

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

// A bid on the auction, from a client and with a given amount
type AuctionBid struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`         // ID of client who made the bid
	Amount int32 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"` // Amount which was bid
}

func (x *AuctionBid) Reset() {
	*x = AuctionBid{}
	mi := &file_auction_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuctionBid) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuctionBid) ProtoMessage() {}

func (x *AuctionBid) ProtoReflect() protoreflect.Message {
	mi := &file_auction_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuctionBid.ProtoReflect.Descriptor instead.
func (*AuctionBid) Descriptor() ([]byte, []int) {
	return file_auction_proto_rawDescGZIP(), []int{0}
}

func (x *AuctionBid) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *AuctionBid) GetAmount() int32 {
	if x != nil {
		return x.Amount
	}
	return 0
}

// A response/result of a bid as to whether or not the bid succeeded or failed
type BidAcknowledge struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status int32 `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"` // Status/result of bid | 0 = success, 1 = fail, 2 = exception/error
}

func (x *BidAcknowledge) Reset() {
	*x = BidAcknowledge{}
	mi := &file_auction_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BidAcknowledge) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BidAcknowledge) ProtoMessage() {}

func (x *BidAcknowledge) ProtoReflect() protoreflect.Message {
	mi := &file_auction_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BidAcknowledge.ProtoReflect.Descriptor instead.
func (*BidAcknowledge) Descriptor() ([]byte, []int) {
	return file_auction_proto_rawDescGZIP(), []int{1}
}

func (x *BidAcknowledge) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

// The outcome/current state of an auction
type AuctionOutcome struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsFinished bool  `protobuf:"varint,1,opt,name=isFinished,proto3" json:"isFinished,omitempty"` // Is the auction finished?
	HighestBid int32 `protobuf:"varint,2,opt,name=highestBid,proto3" json:"highestBid,omitempty"` // Current highest bid
	LeaderId   int32 `protobuf:"varint,3,opt,name=leaderId,proto3" json:"leaderId,omitempty"`     // ID of current auction leader
}

func (x *AuctionOutcome) Reset() {
	*x = AuctionOutcome{}
	mi := &file_auction_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuctionOutcome) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuctionOutcome) ProtoMessage() {}

func (x *AuctionOutcome) ProtoReflect() protoreflect.Message {
	mi := &file_auction_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuctionOutcome.ProtoReflect.Descriptor instead.
func (*AuctionOutcome) Descriptor() ([]byte, []int) {
	return file_auction_proto_rawDescGZIP(), []int{2}
}

func (x *AuctionOutcome) GetIsFinished() bool {
	if x != nil {
		return x.IsFinished
	}
	return false
}

func (x *AuctionOutcome) GetHighestBid() int32 {
	if x != nil {
		return x.HighestBid
	}
	return 0
}

func (x *AuctionOutcome) GetLeaderId() int32 {
	if x != nil {
		return x.LeaderId
	}
	return 0
}

// A package containing all relevant data of the auction
type AuctionData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HighestBid     int32                `protobuf:"varint,1,opt,name=highestBid,proto3" json:"highestBid,omitempty"`        // Current highest bid amount
	HighestBidder  int32                `protobuf:"varint,2,opt,name=highestBidder,proto3" json:"highestBidder,omitempty"`  // ID of the highest bidder
	AuctionEndTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=auctionEndTime,proto3" json:"auctionEndTime,omitempty"` // Timestamp of when the auction ends
}

func (x *AuctionData) Reset() {
	*x = AuctionData{}
	mi := &file_auction_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AuctionData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuctionData) ProtoMessage() {}

func (x *AuctionData) ProtoReflect() protoreflect.Message {
	mi := &file_auction_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuctionData.ProtoReflect.Descriptor instead.
func (*AuctionData) Descriptor() ([]byte, []int) {
	return file_auction_proto_rawDescGZIP(), []int{3}
}

func (x *AuctionData) GetHighestBid() int32 {
	if x != nil {
		return x.HighestBid
	}
	return 0
}

func (x *AuctionData) GetHighestBidder() int32 {
	if x != nil {
		return x.HighestBidder
	}
	return 0
}

func (x *AuctionData) GetAuctionEndTime() *timestamp.Timestamp {
	if x != nil {
		return x.AuctionEndTime
	}
	return nil
}

// Empty message used to represent void return-types
type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_auction_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_auction_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_auction_proto_rawDescGZIP(), []int{4}
}

var File_auction_proto protoreflect.FileDescriptor

var file_auction_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x61, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x34, 0x0a, 0x0a, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x69, 0x64, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x28, 0x0a, 0x0e, 0x42, 0x69, 0x64, 0x41, 0x63, 0x6b,
	0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x22, 0x6c, 0x0a, 0x0e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x63, 0x6f,
	0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x73, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x69, 0x73, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68,
	0x65, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42,
	0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x49, 0x64, 0x22, 0x97,
	0x01, 0x0a, 0x0b, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1e,
	0x0a, 0x0a, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0a, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42, 0x69, 0x64, 0x12, 0x24,
	0x0a, 0x0d, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42, 0x69, 0x64, 0x64, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x68, 0x69, 0x67, 0x68, 0x65, 0x73, 0x74, 0x42, 0x69,
	0x64, 0x64, 0x65, 0x72, 0x12, 0x42, 0x0a, 0x0e, 0x61, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x45,
	0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x61, 0x75, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x45, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x32, 0x78, 0x0a, 0x04, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x23, 0x0a, 0x03, 0x42, 0x69, 0x64,
	0x12, 0x0b, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x69, 0x64, 0x1a, 0x0f, 0x2e,
	0x42, 0x69, 0x64, 0x41, 0x63, 0x6b, 0x6e, 0x6f, 0x77, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x12, 0x21,
	0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x06, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x1a, 0x0f, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x75, 0x74, 0x63, 0x6f, 0x6d,
	0x65, 0x12, 0x28, 0x0a, 0x10, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x41, 0x75,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0c, 0x2e, 0x41, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x44,
	0x61, 0x74, 0x61, 0x1a, 0x06, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x03, 0x5a, 0x01, 0x2f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_auction_proto_rawDescOnce sync.Once
	file_auction_proto_rawDescData = file_auction_proto_rawDesc
)

func file_auction_proto_rawDescGZIP() []byte {
	file_auction_proto_rawDescOnce.Do(func() {
		file_auction_proto_rawDescData = protoimpl.X.CompressGZIP(file_auction_proto_rawDescData)
	})
	return file_auction_proto_rawDescData
}

var file_auction_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_auction_proto_goTypes = []any{
	(*AuctionBid)(nil),          // 0: AuctionBid
	(*BidAcknowledge)(nil),      // 1: BidAcknowledge
	(*AuctionOutcome)(nil),      // 2: AuctionOutcome
	(*AuctionData)(nil),         // 3: AuctionData
	(*Empty)(nil),               // 4: Empty
	(*timestamp.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_auction_proto_depIdxs = []int32{
	5, // 0: AuctionData.auctionEndTime:type_name -> google.protobuf.Timestamp
	0, // 1: Node.Bid:input_type -> AuctionBid
	4, // 2: Node.Result:input_type -> Empty
	3, // 3: Node.ReplicateAuction:input_type -> AuctionData
	1, // 4: Node.Bid:output_type -> BidAcknowledge
	2, // 5: Node.Result:output_type -> AuctionOutcome
	4, // 6: Node.ReplicateAuction:output_type -> Empty
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_auction_proto_init() }
func file_auction_proto_init() {
	if File_auction_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_auction_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_auction_proto_goTypes,
		DependencyIndexes: file_auction_proto_depIdxs,
		MessageInfos:      file_auction_proto_msgTypes,
	}.Build()
	File_auction_proto = out.File
	file_auction_proto_rawDesc = nil
	file_auction_proto_goTypes = nil
	file_auction_proto_depIdxs = nil
}
