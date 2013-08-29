// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// AVP Data conversions between Diameter and Go.  Part of go-diameter.
// Based on database/sql types.

package base

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// OctetString Diameter Type.
type OctetString struct {
	Value   string
	Padding uint32 // Extra bytes to make the Value a multiple of 4 octets
}

// Data implements the Data interface.
func (os *OctetString) Data() Data {
	return os.Value
}

// Put implements the Codec interface. It updates internal Value and Padding.
func (os *OctetString) Put(d Data) {
	b := d.([]byte)
	l := uint32(len(b))
	os.Padding = pad4(l) - l
	os.Value = string(b)
}

// Bytes implement the Codec interface. Padding is always recalculated from
// the internal Value.
func (os *OctetString) Bytes() []byte {
	os.updatePadding() // Do this every time? Geez.
	l := uint32(len(os.Value))
	b := make([]byte, l+os.Padding)
	copy(b, os.Value)
	return b
}

// Length implements the Codec interface. Returns length without padding.
func (os *OctetString) Length() uint32 {
	return uint32(len(os.Value)) - os.Padding
}

// update internal padding value.
func (os *OctetString) updatePadding() {
	if os.Padding == 0 {
		l := uint32(len(os.Value))
		os.Padding = pad4(l) - l
	}
}

// String returns a human readable version of the AVP.
func (os *OctetString) String() string {
	os.updatePadding() // Update padding
	return fmt.Sprintf("OctetString{Value:'%s',Padding:%d}",
		os.Value, os.Padding)
}

// UTF8String Diameter Type.
type UTF8String struct {
	OctetString
}

// String returns a human readable version of the AVP.
func (p *UTF8String) String() string {
	p.updatePadding() // Update padding
	return fmt.Sprintf("UTF8String{Value:'%s',Padding:%d}",
		p.Value, p.Padding)
}

// DiameterIdentity Diameter Type.
type DiameterIdentity struct {
	OctetString
}

// String returns a human readable version of the AVP.
func (p *DiameterIdentity) String() string {
	p.updatePadding() // Update padding
	return fmt.Sprintf("DiameterIdentity{Value:'%s',Padding:%d}",
		p.Value, p.Padding)
}

// IPFilterRule Diameter Type.
type IPFilterRule struct {
	OctetString
}

// String returns a human readable version of the AVP.
func (p *IPFilterRule) String() string {
	p.updatePadding() // Update padding
	return fmt.Sprintf("IPFilterRule{Value:'%s',Padding:%d}",
		p.Value, p.Padding)
}

// Time Diameter Type.
type Time struct {
	Value time.Time
}

// Data implements the Data interface.
func (t *Time) Data() Data {
	return t.Value
}

// Put implements the Codec interface. It updates internal Value.
func (t *Time) Put(d Data) {
	b := d.([]byte)
	if len(b) == 4 {
		t.Value = time.Unix(int64(binary.BigEndian.Uint32(b)), 0)
	}
}

// Bytes implement the Codec interface.
func (t *Time) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(t.Value.Unix()))
	return buf.Bytes()
}

// Length implements the Codec interface. Returns length without padding.
func (t *Time) Length() uint32 {
	return 4
}

// String returns a human readable version of the AVP.
func (t *Time) String() string {
	return fmt.Sprintf("Time{Value:'%s'}", t.Value.String())
}

// Address Diameter Type.
type Address struct {
	// Address Family (e.g. AF_INET=1)
	// http://www.iana.org/assignments/address-family-numbers/address-family-numbers.xhtml
	Family []byte

	// Parsed IP address
	IP net.IP

	// Padding to 4 octets
	Padding int
}

// Data implements the Data interface.
func (addr *Address) Data() Data {
	return addr.IP
}

// Put implements the Coded interface. It updates internal Family and IP.
func (addr *Address) Put(d Data) {
	b := d.([]byte)
	// TODO: Support IPv6
	if len(b) >= 6 && b[1] == 1 { // AF_INET=1 IPv4 only.
		addr.Family = []byte{b[0], b[1]}
		addr.IP = net.IPv4(b[2], b[3], b[4], b[5])
		addr.Padding = 2
	}
}

// Bytes implement the Codec interface.
func (addr *Address) Bytes() []byte {
	if ip := addr.IP.To4(); ip != nil {
		if addr.Family == nil {
			addr.Family = []byte{0, 1}
		}
		// IPv4 always need 2 byte padding. (derived from OctetString)
		addr.Padding = 2 // TODO: Fix this
		b := []byte{
			addr.Family[0],
			addr.Family[1],
			ip[0],
			ip[1],
			ip[2],
			ip[3],
			0, // Padding
			0, // Padding
		}
		return b
	}
	return []byte{}
}

// Length implements the Codec interface. Returns length without padding.
func (addr *Address) Length() uint32 {
	if addr.IP.To4() != nil {
		return 6 // TODO: Fix this
	}
	return 0
}

// String returns a human readable version of the AVP.
func (addr *Address) String() string {
	addr.Bytes() // Update family and padding
	return fmt.Sprintf("Address{Value:'%s',Padding:%d}",
		addr.IP.String(), addr.Padding)
}

// IPv4 Type for Framed-IP-Address and alike.
type IPv4 struct {
	// Parsed IP address
	IP net.IP
}

// Data implements the Data interface.
func (addr *IPv4) Data() Data {
	return addr.IP
}

// Put implements the Coded interface. It updates internal Family and IP.
func (addr *IPv4) Put(d Data) {
	b := d.([]byte)
	if len(b) == 4 {
		addr.IP = net.IPv4(b[0], b[1], b[2], b[3])
	}
}

// Bytes implement the Codec interface.
func (addr *IPv4) Bytes() []byte {
	if ip := addr.IP.To4(); ip != nil {
		return ip
	}
	return []byte{}
}

// Length implements the Codec interface. Returns length without padding.
func (addr *IPv4) Length() uint32 {
	if addr.IP.To4() != nil {
		return 4 // TODO: Fix this
	}
	return 0
}

// String returns a human readable version of the AVP.
func (addr *IPv4) String() string {
	return fmt.Sprintf("IPv4{Value:'%s'}", addr.IP.String())
}

// DiameterURI Diameter Type.
type DiameterURI struct {
	Value string
}

// Data implements the Data interface.
func (du *DiameterURI) Data() Data {
	return du.Value
}

// Put implements the Codec interface.
func (du *DiameterURI) Put(d Data) {
	du.Value = string(d.([]byte))
}

// Bytes implement the Codec interface.
func (du *DiameterURI) Bytes() []byte {
	return []byte(du.Value)
}

// Length implements the Codec interface.
func (du *DiameterURI) Length() uint32 {
	return uint32(len(du.Value))
}

// String returns a human readable version of the AVP.
func (du *DiameterURI) String() string {
	return fmt.Sprintf("DiameterURI{Value:'%s'}", du.Value)
}

// Integer32 Diameter Type
type Integer32 struct {
	Value  int32
	Buffer []byte
}

// Data *implements the Data interface.
func (n Integer32) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Int32.
func (n *Integer32) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Int32 and stored on Buffer.
func (n *Integer32) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Integer32) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Integer32) String() string {
	return fmt.Sprintf("Integer32{Value:%d}", n.Value)
}

// Integer64 Diameter Type
type Integer64 struct {
	Value  int64
	Buffer []byte
}

// Data implements the Data interface.
func (n *Integer64) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Int64.
func (n *Integer64) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Int64 and stored on Buffer.
func (n *Integer64) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Integer64) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Integer64) String() string {
	return fmt.Sprintf("Integer64{Value:%d}", n.Value)
}

// Unsigned32 Diameter Type
type Unsigned32 struct {
	Value  uint32
	Buffer []byte
}

// Data implements the Data interface.
func (n *Unsigned32) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Uint32.
func (n *Unsigned32) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Uint32 and stored on Buffer.
func (n *Unsigned32) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Unsigned32) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Unsigned32) String() string {
	return fmt.Sprintf("Unsigned32{Value:%d}", n.Value)
}

// Unsigned64 Diameter Type
type Unsigned64 struct {
	Value  uint64
	Buffer []byte
}

// Data implements the Data interface.
func (n *Unsigned64) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Uint64.
func (n *Unsigned64) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Uint64 and stored on Buffer.
func (n *Unsigned64) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Unsigned64) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Unsigned64) String() string {
	return fmt.Sprintf("Unsigned64{Value:%d}", n.Value)
}

// Float32 Diameter Type
type Float32 struct {
	Value  float32
	Buffer []byte
}

// Data implements the Data interface.
func (n *Float32) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Float32.
func (n *Float32) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Float32 and stored on Buffer.
func (n *Float32) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Float32) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Float32) String() string {
	return fmt.Sprintf("Float32{Value:%d}", n.Value)
}

// Float64 Diameter Type
type Float64 struct {
	Value  float64
	Buffer []byte
}

// Data implements the Data interface.
func (n *Float64) Data() Data {
	return n.Value
}

// Put implements the Codec interface. It updates internal Buffer and Float64.
func (n *Float64) Put(d Data) {
	n.Buffer = d.([]byte)
	binary.Read(bytes.NewBuffer(n.Buffer), binary.BigEndian, &n.Value)
}

// Bytes implement the Codec interface. Bytes are always rewritten from
// the internal Float64 and stored on Buffer.
func (n *Float64) Bytes() []byte {
	b := new(bytes.Buffer)
	binary.Write(b, binary.BigEndian, n.Value)
	n.Buffer = b.Bytes()
	return n.Buffer
}

// Length implements the Codec interface.
func (n *Float64) Length() uint32 {
	if n.Buffer == nil {
		n.Bytes()
	}
	return uint32(len(n.Buffer))
}

// String returns a human readable version of the AVP.
func (n *Float64) String() string {
	return fmt.Sprintf("Float64{Value:%d}", n.Value)
}

// Enumerated Diameter Type
type Enumerated struct {
	Integer32
}

// String returns a human readable version of the AVP.
func (p *Enumerated) String() string {
	return fmt.Sprintf("Enumerated{Value:%d}", p.Value)
}
