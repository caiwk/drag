// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: multiraft.proto

/*
	Package multiraftpb is a generated protocol buffer package.

	It is generated from these files:
		multiraft.proto

	It has these top-level messages:
		Session
		SessionRequest
		SessionResponse
		RaftProposal
		RaftReadIndex
		RaftResponse
*/
package multiraftpb

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"
import client "github.com/lni/dragonboat/client"

import context "golang.org/x/net/context"
import grpc "google.golang.org/grpc"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// Session is the session object used to track proposals for the
// specified raft cluster. SeriesID is a sequential id used to identify
// proposals, RespondedTo is a sequential id used to track the last
// responded proposal.
type Session struct {
	ClusterID   uint64 `protobuf:"varint,1,opt,name=ClusterID" json:"ClusterID"`
	ClientID    uint64 `protobuf:"varint,2,opt,name=ClientID" json:"ClientID"`
	SeriesID    uint64 `protobuf:"varint,3,opt,name=SeriesID" json:"SeriesID"`
	RespondedTo uint64 `protobuf:"varint,4,opt,name=RespondedTo" json:"RespondedTo"`
}

func (m *Session) Reset()                    { *m = Session{} }
func (m *Session) String() string            { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()               {}
func (*Session) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{0} }

func (m *Session) GetClusterID() uint64 {
	if m != nil {
		return m.ClusterID
	}
	return 0
}

func (m *Session) GetClientID() uint64 {
	if m != nil {
		return m.ClientID
	}
	return 0
}

func (m *Session) GetSeriesID() uint64 {
	if m != nil {
		return m.SeriesID
	}
	return 0
}

func (m *Session) GetRespondedTo() uint64 {
	if m != nil {
		return m.RespondedTo
	}
	return 0
}

// SessionRequest is the message used to specified the interested raft
// cluster.
type SessionRequest struct {
	ClusterId uint64 `protobuf:"varint,1,req,name=cluster_id,json=clusterId" json:"cluster_id"`
}

func (m *SessionRequest) Reset()                    { *m = SessionRequest{} }
func (m *SessionRequest) String() string            { return proto.CompactTextString(m) }
func (*SessionRequest) ProtoMessage()               {}
func (*SessionRequest) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{1} }

func (m *SessionRequest) GetClusterId() uint64 {
	if m != nil {
		return m.ClusterId
	}
	return 0
}

// SessionResponse is the message used to indicate whether the
// Session object is successfully closed.
type SessionResponse struct {
	Completed bool `protobuf:"varint,1,req,name=completed" json:"completed"`
}

func (m *SessionResponse) Reset()                    { *m = SessionResponse{} }
func (m *SessionResponse) String() string            { return proto.CompactTextString(m) }
func (*SessionResponse) ProtoMessage()               {}
func (*SessionResponse) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{2} }

func (m *SessionResponse) GetCompleted() bool {
	if m != nil {
		return m.Completed
	}
	return false
}

// RaftProposal is the message used to describe the proposal to be made on the
// selected raft cluster.
type RaftProposal struct {
	Session client.Session `protobuf:"bytes,1,opt,name=session" json:"session"`
	Data    []byte         `protobuf:"bytes,2,opt,name=data" json:"data"`
}

func (m *RaftProposal) Reset()                    { *m = RaftProposal{} }
func (m *RaftProposal) String() string            { return proto.CompactTextString(m) }
func (*RaftProposal) ProtoMessage()               {}
func (*RaftProposal) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{3} }

func (m *RaftProposal) GetSession() client.Session {
	if m != nil {
		return m.Session
	}
	return client.Session{}
}

func (m *RaftProposal) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// RaftReadIndex is the message used to describe the input to the ReadIndex
// protocol. The ReadIndex protocol is used for making linearizable read on
// the selected raft cluster.
type RaftReadIndex struct {
	ClusterId uint64 `protobuf:"varint,1,opt,name=cluster_id,json=clusterId" json:"cluster_id"`
	Data      []byte `protobuf:"bytes,2,opt,name=data" json:"data"`
}

func (m *RaftReadIndex) Reset()                    { *m = RaftReadIndex{} }
func (m *RaftReadIndex) String() string            { return proto.CompactTextString(m) }
func (*RaftReadIndex) ProtoMessage()               {}
func (*RaftReadIndex) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{4} }

func (m *RaftReadIndex) GetClusterId() uint64 {
	if m != nil {
		return m.ClusterId
	}
	return 0
}

func (m *RaftReadIndex) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// RaftResponse is the message used to describe the response produced by
// the Update or Lookup function of the IDataStore instance.
type RaftResponse struct {
	Result uint64 `protobuf:"varint,1,opt,name=result" json:"result"`
	Data   []byte `protobuf:"bytes,2,opt,name=data" json:"data"`
}

func (m *RaftResponse) Reset()                    { *m = RaftResponse{} }
func (m *RaftResponse) String() string            { return proto.CompactTextString(m) }
func (*RaftResponse) ProtoMessage()               {}
func (*RaftResponse) Descriptor() ([]byte, []int) { return fileDescriptorMultiraft, []int{5} }

func (m *RaftResponse) GetResult() uint64 {
	if m != nil {
		return m.Result
	}
	return 0
}

func (m *RaftResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Session)(nil), "multiraftpb.Session")
	proto.RegisterType((*SessionRequest)(nil), "multiraftpb.SessionRequest")
	proto.RegisterType((*SessionResponse)(nil), "multiraftpb.SessionResponse")
	proto.RegisterType((*RaftProposal)(nil), "multiraftpb.RaftProposal")
	proto.RegisterType((*RaftReadIndex)(nil), "multiraftpb.RaftReadIndex")
	proto.RegisterType((*RaftResponse)(nil), "multiraftpb.RaftResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for NodehostAPI service

type NodehostAPIClient interface {
	// GetSession returns a new Session object ready to be used for
	// making new proposals.
	GetSession(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*client.Session, error)
	// CloseSession closes the specified Session object and removes
	// it from the associated Raft cluster. The Completed boolean field in
	// the returned SessionResponse object indicates whether the
	// Session object is successfully closed.
	CloseSession(ctx context.Context, in *client.Session, opts ...grpc.CallOption) (*SessionResponse, error)
	// Propose makes a proposal. When there is no error, the Result field of the
	// returned RaftResponse is the uint64 value returned by the Update function
	// of the associated IDataStore instance.
	Propose(ctx context.Context, in *RaftProposal, opts ...grpc.CallOption) (*RaftResponse, error)
	// Read makes a new linearizable read on the specified cluster. When there is
	// no error, the Data field of the returned RaftResponse is the returned
	// query result generated by the Lookup function of the associated IDataStore
	// instance.
	Read(ctx context.Context, in *RaftReadIndex, opts ...grpc.CallOption) (*RaftResponse, error)
}

type nodehostAPIClient struct {
	cc *grpc.ClientConn
}

func NewNodehostAPIClient(cc *grpc.ClientConn) NodehostAPIClient {
	return &nodehostAPIClient{cc}
}

func (c *nodehostAPIClient) GetSession(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*client.Session, error) {
	out := new(client.Session)
	err := grpc.Invoke(ctx, "/multiraftpb.NodehostAPI/GetSession", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodehostAPIClient) CloseSession(ctx context.Context, in *client.Session, opts ...grpc.CallOption) (*SessionResponse, error) {
	out := new(SessionResponse)
	err := grpc.Invoke(ctx, "/multiraftpb.NodehostAPI/CloseSession", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodehostAPIClient) Propose(ctx context.Context, in *RaftProposal, opts ...grpc.CallOption) (*RaftResponse, error) {
	out := new(RaftResponse)
	err := grpc.Invoke(ctx, "/multiraftpb.NodehostAPI/Propose", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodehostAPIClient) Read(ctx context.Context, in *RaftReadIndex, opts ...grpc.CallOption) (*RaftResponse, error) {
	out := new(RaftResponse)
	err := grpc.Invoke(ctx, "/multiraftpb.NodehostAPI/Read", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for NodehostAPI service

type NodehostAPIServer interface {
	// GetSession returns a new Session object ready to be used for
	// making new proposals.
	GetSession(context.Context, *SessionRequest) (*client.Session, error)
	// CloseSession closes the specified Session object and removes
	// it from the associated Raft cluster. The Completed boolean field in
	// the returned SessionResponse object indicates whether the
	// Session object is successfully closed.
	CloseSession(context.Context, *client.Session) (*SessionResponse, error)
	// Propose makes a proposal. When there is no error, the Result field of the
	// returned RaftResponse is the uint64 value returned by the Update function
	// of the associated IDataStore instance.
	Propose(context.Context, *RaftProposal) (*RaftResponse, error)
	// Read makes a new linearizable read on the specified cluster. When there is
	// no error, the Data field of the returned RaftResponse is the returned
	// query result generated by the Lookup function of the associated IDataStore
	// instance.
	Read(context.Context, *RaftReadIndex) (*RaftResponse, error)
}

func RegisterNodehostAPIServer(s *grpc.Server, srv NodehostAPIServer) {
	s.RegisterService(&_NodehostAPI_serviceDesc, srv)
}

func _NodehostAPI_GetSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodehostAPIServer).GetSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/multiraftpb.NodehostAPI/GetSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodehostAPIServer).GetSession(ctx, req.(*SessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodehostAPI_CloseSession_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(client.Session)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodehostAPIServer).CloseSession(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/multiraftpb.NodehostAPI/CloseSession",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodehostAPIServer).CloseSession(ctx, req.(*client.Session))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodehostAPI_Propose_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RaftProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodehostAPIServer).Propose(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/multiraftpb.NodehostAPI/Propose",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodehostAPIServer).Propose(ctx, req.(*RaftProposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodehostAPI_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RaftReadIndex)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodehostAPIServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/multiraftpb.NodehostAPI/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodehostAPIServer).Read(ctx, req.(*RaftReadIndex))
	}
	return interceptor(ctx, in, info, handler)
}

var _NodehostAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "multiraftpb.NodehostAPI",
	HandlerType: (*NodehostAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSession",
			Handler:    _NodehostAPI_GetSession_Handler,
		},
		{
			MethodName: "CloseSession",
			Handler:    _NodehostAPI_CloseSession_Handler,
		},
		{
			MethodName: "Propose",
			Handler:    _NodehostAPI_Propose_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _NodehostAPI_Read_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "multiraft.proto",
}

func (m *Session) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Session) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.ClusterID))
	dAtA[i] = 0x10
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.ClientID))
	dAtA[i] = 0x18
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.SeriesID))
	dAtA[i] = 0x20
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.RespondedTo))
	return i, nil
}

func (m *SessionRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SessionRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.ClusterId))
	return i, nil
}

func (m *SessionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SessionResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	if m.Completed {
		dAtA[i] = 1
	} else {
		dAtA[i] = 0
	}
	i++
	return i, nil
}

func (m *RaftProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RaftProposal) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0xa
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.Session.Size()))
	n1, err := m.Session.MarshalTo(dAtA[i:])
	if err != nil {
		return 0, err
	}
	i += n1
	if m.Data != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMultiraft(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *RaftReadIndex) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RaftReadIndex) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.ClusterId))
	if m.Data != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMultiraft(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *RaftResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RaftResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	dAtA[i] = 0x8
	i++
	i = encodeVarintMultiraft(dAtA, i, uint64(m.Result))
	if m.Data != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintMultiraft(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func encodeVarintMultiraft(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Session) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovMultiraft(uint64(m.ClusterID))
	n += 1 + sovMultiraft(uint64(m.ClientID))
	n += 1 + sovMultiraft(uint64(m.SeriesID))
	n += 1 + sovMultiraft(uint64(m.RespondedTo))
	return n
}

func (m *SessionRequest) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovMultiraft(uint64(m.ClusterId))
	return n
}

func (m *SessionResponse) Size() (n int) {
	var l int
	_ = l
	n += 2
	return n
}

func (m *RaftProposal) Size() (n int) {
	var l int
	_ = l
	l = m.Session.Size()
	n += 1 + l + sovMultiraft(uint64(l))
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sovMultiraft(uint64(l))
	}
	return n
}

func (m *RaftReadIndex) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovMultiraft(uint64(m.ClusterId))
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sovMultiraft(uint64(l))
	}
	return n
}

func (m *RaftResponse) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovMultiraft(uint64(m.Result))
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sovMultiraft(uint64(l))
	}
	return n
}

func sovMultiraft(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMultiraft(x uint64) (n int) {
	return sovMultiraft(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Session) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Session: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Session: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterID", wireType)
			}
			m.ClusterID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ClusterID |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClientID", wireType)
			}
			m.ClientID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ClientID |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SeriesID", wireType)
			}
			m.SeriesID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SeriesID |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RespondedTo", wireType)
			}
			m.RespondedTo = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RespondedTo |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SessionRequest) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SessionRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SessionRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterId", wireType)
			}
			m.ClusterId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ClusterId |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return proto.NewRequiredNotSetError("cluster_id")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SessionResponse) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SessionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SessionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Completed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Completed = bool(v != 0)
			hasFields[0] |= uint64(0x00000001)
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return proto.NewRequiredNotSetError("completed")
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RaftProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RaftProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RaftProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Session", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMultiraft
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Session.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMultiraft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RaftReadIndex) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RaftReadIndex: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RaftReadIndex: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterId", wireType)
			}
			m.ClusterId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ClusterId |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMultiraft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RaftResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RaftResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RaftResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Result", wireType)
			}
			m.Result = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Result |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMultiraft
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMultiraft(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMultiraft
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMultiraft(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMultiraft
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMultiraft
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMultiraft
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMultiraft
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMultiraft(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMultiraft = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMultiraft   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("multiraft.proto", fileDescriptorMultiraft) }

var fileDescriptorMultiraft = []byte{
	// 446 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x92, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x86, 0xe3, 0x60, 0x91, 0x32, 0x09, 0x04, 0xed, 0xc9, 0x98, 0x2a, 0x44, 0x46, 0x42, 0x5c,
	0xb0, 0xa5, 0x4a, 0xbd, 0x21, 0x55, 0x6d, 0x22, 0x90, 0x2f, 0x55, 0xe5, 0x72, 0xe1, 0x84, 0x9c,
	0xec, 0x24, 0xb5, 0xe4, 0x78, 0x8c, 0x77, 0x2d, 0xf1, 0x18, 0xbd, 0xf2, 0x46, 0x3d, 0xf2, 0x04,
	0x08, 0x85, 0x17, 0x41, 0xde, 0x5d, 0xbb, 0x9b, 0x36, 0xe4, 0xb6, 0xfb, 0xcf, 0x3f, 0x9f, 0x66,
	0xfe, 0x5d, 0x18, 0x6f, 0xea, 0x5c, 0x66, 0x55, 0xba, 0x92, 0x61, 0x59, 0x91, 0x24, 0x36, 0xec,
	0x84, 0x72, 0xe1, 0x7f, 0x58, 0x67, 0xf2, 0xa6, 0x5e, 0x84, 0x4b, 0xda, 0x44, 0x6b, 0x5a, 0x53,
	0xa4, 0x3c, 0x8b, 0x7a, 0xa5, 0x6e, 0xea, 0xa2, 0x4e, 0xba, 0xd7, 0x0f, 0x2d, 0x7b, 0x5e, 0x64,
	0x11, 0xaf, 0xd2, 0x35, 0x15, 0x0b, 0x4a, 0x65, 0xb4, 0xcc, 0x33, 0x2c, 0x64, 0x24, 0x50, 0x88,
	0x8c, 0x0a, 0xed, 0x0f, 0x7e, 0x3a, 0x30, 0xb8, 0xd6, 0x0a, 0x0b, 0xe0, 0xd9, 0x2c, 0xaf, 0x85,
	0xc4, 0x2a, 0x9e, 0x7b, 0xce, 0xd4, 0x79, 0xef, 0x5e, 0xb8, 0x77, 0xbf, 0xdf, 0xf4, 0x92, 0x7b,
	0x99, 0x4d, 0xe1, 0x68, 0xa6, 0x38, 0xf1, 0xdc, 0xeb, 0x5b, 0x96, 0x4e, 0x6d, 0x1c, 0xd7, 0x58,
	0x65, 0x28, 0xe2, 0xb9, 0xf7, 0xc4, 0x76, 0xb4, 0x2a, 0x7b, 0x07, 0xc3, 0x04, 0x45, 0x49, 0x05,
	0x47, 0xfe, 0x85, 0x3c, 0xd7, 0x32, 0xd9, 0x85, 0xe0, 0x14, 0x5e, 0x98, 0xd1, 0x12, 0xfc, 0x5e,
	0xa3, 0x90, 0xec, 0x2d, 0xc0, 0x52, 0x8f, 0xf2, 0x2d, 0xe3, 0x9e, 0x33, 0xed, 0xdf, 0x8f, 0x68,
	0xf4, 0x98, 0x07, 0xa7, 0x30, 0xee, 0xda, 0x1a, 0x98, 0xc0, 0x66, 0xb3, 0x25, 0x6d, 0xca, 0x1c,
	0x25, 0xea, 0xb6, 0xa3, 0xae, 0xad, 0x95, 0x83, 0xaf, 0x30, 0x4a, 0xd2, 0x95, 0xbc, 0xaa, 0xa8,
	0x24, 0x91, 0xe6, 0x2c, 0x82, 0x81, 0x89, 0x4a, 0x65, 0x31, 0x3c, 0x19, 0x87, 0x3a, 0xc1, 0xd0,
	0xd0, 0x0d, 0xa2, 0x75, 0x31, 0x0f, 0x5c, 0x9e, 0xca, 0x54, 0xc5, 0x32, 0x32, 0x45, 0xa5, 0x04,
	0x97, 0xf0, 0xbc, 0x41, 0x27, 0x98, 0xf2, 0xb8, 0xe0, 0xf8, 0xe3, 0xd1, 0x1e, 0xce, 0x9e, 0x3d,
	0x0e, 0xf0, 0x3e, 0xe9, 0x51, 0xbb, 0xf5, 0x8e, 0xe1, 0x69, 0x85, 0xa2, 0xce, 0xe5, 0x0e, 0xca,
	0x68, 0xff, 0xe7, 0x9c, 0xdc, 0xf6, 0x61, 0x78, 0x49, 0x1c, 0x6f, 0x48, 0xc8, 0xf3, 0xab, 0x98,
	0x7d, 0x04, 0xf8, 0x8c, 0xb2, 0xfd, 0x0e, 0xaf, 0x43, 0xeb, 0x1f, 0x86, 0xbb, 0x2f, 0xe1, 0x3f,
	0x0c, 0x23, 0xe8, 0xb1, 0x33, 0x18, 0xcd, 0x72, 0x12, 0xd8, 0xf6, 0x3f, 0xb4, 0xf8, 0xc7, 0xfb,
	0x81, 0x7a, 0x89, 0xa0, 0xc7, 0xce, 0x61, 0xa0, 0xd3, 0x47, 0xf6, 0x6a, 0xc7, 0x6a, 0xbf, 0x8b,
	0xff, 0xb8, 0x64, 0x21, 0xce, 0xc0, 0x6d, 0x52, 0x66, 0xfe, 0x1e, 0x93, 0x09, 0xff, 0x20, 0xe0,
	0xe2, 0xe5, 0xdd, 0x76, 0xe2, 0xfc, 0xda, 0x4e, 0x9c, 0x3f, 0xdb, 0x89, 0x73, 0xfb, 0x77, 0xd2,
	0xfb, 0x17, 0x00, 0x00, 0xff, 0xff, 0xd4, 0x82, 0x0f, 0x5f, 0x9f, 0x03, 0x00, 0x00,
}
