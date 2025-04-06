// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.12.4
// source: proto/token/payment/payment.proto

package token

import (
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Payment struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TokenData     string                 `protobuf:"bytes,1,opt,name=token_data,proto3" json:"token_data,omitempty"`
	CardType      string                 `protobuf:"bytes,2,opt,name=card_type,proto3" json:"card_type,omitempty"`
	CardModel     string                 `protobuf:"bytes,3,opt,name=card_model,proto3" json:"card_model,omitempty"`
	CardAtc       uint32                 `protobuf:"varint,4,opt,name=card_atc,proto3" json:"card_atc,omitempty"`
	Currency      string                 `protobuf:"bytes,5,opt,name=currency,proto3" json:"currency,omitempty"`
	Amount        float64                `protobuf:"fixed64,6,opt,name=amount,proto3" json:"amount,omitempty"`
	Terminal      string                 `protobuf:"bytes,7,opt,name=terminal,proto3" json:"terminal,omitempty"`
	Status        string                 `protobuf:"bytes,8,opt,name=status,proto3" json:"status,omitempty"`
	Mcc           string                 `protobuf:"bytes,9,opt,name=mcc,proto3" json:"mcc,omitempty"`
	PaymentAt     *timestamp.Timestamp   `protobuf:"bytes,10,opt,name=payment_at,proto3" json:"payment_at,omitempty"`
	TransactionId string                 `protobuf:"bytes,11,opt,name=transaction_id,proto3" json:"transaction_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Payment) Reset() {
	*x = Payment{}
	mi := &file_proto_token_payment_payment_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Payment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payment) ProtoMessage() {}

func (x *Payment) ProtoReflect() protoreflect.Message {
	mi := &file_proto_token_payment_payment_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payment.ProtoReflect.Descriptor instead.
func (*Payment) Descriptor() ([]byte, []int) {
	return file_proto_token_payment_payment_proto_rawDescGZIP(), []int{0}
}

func (x *Payment) GetTokenData() string {
	if x != nil {
		return x.TokenData
	}
	return ""
}

func (x *Payment) GetCardType() string {
	if x != nil {
		return x.CardType
	}
	return ""
}

func (x *Payment) GetCardModel() string {
	if x != nil {
		return x.CardModel
	}
	return ""
}

func (x *Payment) GetCardAtc() uint32 {
	if x != nil {
		return x.CardAtc
	}
	return 0
}

func (x *Payment) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *Payment) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Payment) GetTerminal() string {
	if x != nil {
		return x.Terminal
	}
	return ""
}

func (x *Payment) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Payment) GetMcc() string {
	if x != nil {
		return x.Mcc
	}
	return ""
}

func (x *Payment) GetPaymentAt() *timestamp.Timestamp {
	if x != nil {
		return x.PaymentAt
	}
	return nil
}

func (x *Payment) GetTransactionId() string {
	if x != nil {
		return x.TransactionId
	}
	return ""
}

type Step struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	StepProcess   string                 `protobuf:"bytes,1,opt,name=step_process,proto3" json:"step_process,omitempty"`
	ProcessedAt   *timestamp.Timestamp   `protobuf:"bytes,2,opt,name=processed_at,proto3" json:"processed_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Step) Reset() {
	*x = Step{}
	mi := &file_proto_token_payment_payment_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Step) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Step) ProtoMessage() {}

func (x *Step) ProtoReflect() protoreflect.Message {
	mi := &file_proto_token_payment_payment_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Step.ProtoReflect.Descriptor instead.
func (*Step) Descriptor() ([]byte, []int) {
	return file_proto_token_payment_payment_proto_rawDescGZIP(), []int{1}
}

func (x *Step) GetStepProcess() string {
	if x != nil {
		return x.StepProcess
	}
	return ""
}

func (x *Step) GetProcessedAt() *timestamp.Timestamp {
	if x != nil {
		return x.ProcessedAt
	}
	return nil
}

type PaymentTokenRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Payment       *Payment               `protobuf:"bytes,1,opt,name=payment,proto3" json:"payment,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PaymentTokenRequest) Reset() {
	*x = PaymentTokenRequest{}
	mi := &file_proto_token_payment_payment_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PaymentTokenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaymentTokenRequest) ProtoMessage() {}

func (x *PaymentTokenRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_token_payment_payment_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaymentTokenRequest.ProtoReflect.Descriptor instead.
func (*PaymentTokenRequest) Descriptor() ([]byte, []int) {
	return file_proto_token_payment_payment_proto_rawDescGZIP(), []int{2}
}

func (x *PaymentTokenRequest) GetPayment() *Payment {
	if x != nil {
		return x.Payment
	}
	return nil
}

type PaymentTokenResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Payment       *Payment               `protobuf:"bytes,1,opt,name=payment,proto3" json:"payment,omitempty"`
	Steps         []*Step                `protobuf:"bytes,2,rep,name=steps,proto3" json:"steps,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PaymentTokenResponse) Reset() {
	*x = PaymentTokenResponse{}
	mi := &file_proto_token_payment_payment_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PaymentTokenResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PaymentTokenResponse) ProtoMessage() {}

func (x *PaymentTokenResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_token_payment_payment_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PaymentTokenResponse.ProtoReflect.Descriptor instead.
func (*PaymentTokenResponse) Descriptor() ([]byte, []int) {
	return file_proto_token_payment_payment_proto_rawDescGZIP(), []int{3}
}

func (x *PaymentTokenResponse) GetPayment() *Payment {
	if x != nil {
		return x.Payment
	}
	return nil
}

func (x *PaymentTokenResponse) GetSteps() []*Step {
	if x != nil {
		return x.Steps
	}
	return nil
}

var File_proto_token_payment_payment_proto protoreflect.FileDescriptor

const file_proto_token_payment_payment_proto_rawDesc = "" +
	"\n" +
	"!proto/token/payment/payment.proto\x12\x05token\x1a\x1fgoogle/protobuf/timestamp.proto\"\xe1\x02\n" +
	"\aPayment\x12\x1e\n" +
	"\n" +
	"token_data\x18\x01 \x01(\tR\n" +
	"token_data\x12\x1c\n" +
	"\tcard_type\x18\x02 \x01(\tR\tcard_type\x12\x1e\n" +
	"\n" +
	"card_model\x18\x03 \x01(\tR\n" +
	"card_model\x12\x1a\n" +
	"\bcard_atc\x18\x04 \x01(\rR\bcard_atc\x12\x1a\n" +
	"\bcurrency\x18\x05 \x01(\tR\bcurrency\x12\x16\n" +
	"\x06amount\x18\x06 \x01(\x01R\x06amount\x12\x1a\n" +
	"\bterminal\x18\a \x01(\tR\bterminal\x12\x16\n" +
	"\x06status\x18\b \x01(\tR\x06status\x12\x10\n" +
	"\x03mcc\x18\t \x01(\tR\x03mcc\x12:\n" +
	"\n" +
	"payment_at\x18\n" +
	" \x01(\v2\x1a.google.protobuf.TimestampR\n" +
	"payment_at\x12&\n" +
	"\x0etransaction_id\x18\v \x01(\tR\x0etransaction_id\"j\n" +
	"\x04Step\x12\"\n" +
	"\fstep_process\x18\x01 \x01(\tR\fstep_process\x12>\n" +
	"\fprocessed_at\x18\x02 \x01(\v2\x1a.google.protobuf.TimestampR\fprocessed_at\"?\n" +
	"\x13PaymentTokenRequest\x12(\n" +
	"\apayment\x18\x01 \x01(\v2\x0e.token.PaymentR\apayment\"c\n" +
	"\x14PaymentTokenResponse\x12(\n" +
	"\apayment\x18\x01 \x01(\v2\x0e.token.PaymentR\apayment\x12!\n" +
	"\x05steps\x18\x02 \x03(\v2\v.token.StepR\x05stepsB\x11Z\x0f/protogen/tokenb\x06proto3"

var (
	file_proto_token_payment_payment_proto_rawDescOnce sync.Once
	file_proto_token_payment_payment_proto_rawDescData []byte
)

func file_proto_token_payment_payment_proto_rawDescGZIP() []byte {
	file_proto_token_payment_payment_proto_rawDescOnce.Do(func() {
		file_proto_token_payment_payment_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_token_payment_payment_proto_rawDesc), len(file_proto_token_payment_payment_proto_rawDesc)))
	})
	return file_proto_token_payment_payment_proto_rawDescData
}

var file_proto_token_payment_payment_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_token_payment_payment_proto_goTypes = []any{
	(*Payment)(nil),              // 0: token.Payment
	(*Step)(nil),                 // 1: token.Step
	(*PaymentTokenRequest)(nil),  // 2: token.PaymentTokenRequest
	(*PaymentTokenResponse)(nil), // 3: token.PaymentTokenResponse
	(*timestamp.Timestamp)(nil),  // 4: google.protobuf.Timestamp
}
var file_proto_token_payment_payment_proto_depIdxs = []int32{
	4, // 0: token.Payment.payment_at:type_name -> google.protobuf.Timestamp
	4, // 1: token.Step.processed_at:type_name -> google.protobuf.Timestamp
	0, // 2: token.PaymentTokenRequest.payment:type_name -> token.Payment
	0, // 3: token.PaymentTokenResponse.payment:type_name -> token.Payment
	1, // 4: token.PaymentTokenResponse.steps:type_name -> token.Step
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_proto_token_payment_payment_proto_init() }
func file_proto_token_payment_payment_proto_init() {
	if File_proto_token_payment_payment_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_token_payment_payment_proto_rawDesc), len(file_proto_token_payment_payment_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_token_payment_payment_proto_goTypes,
		DependencyIndexes: file_proto_token_payment_payment_proto_depIdxs,
		MessageInfos:      file_proto_token_payment_payment_proto_msgTypes,
	}.Build()
	File_proto_token_payment_payment_proto = out.File
	file_proto_token_payment_payment_proto_goTypes = nil
	file_proto_token_payment_payment_proto_depIdxs = nil
}
