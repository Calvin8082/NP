package main

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const disabled uint8 = 0
const enabledTxOnly uint8 = 1
const enabledRxOnly uint8 = 2
const enabledRxTx uint8 = 3

type tlv struct {
	tlvType   uint8
	tlvLength int
	tlvValue  interface{}
}
type chassisIDValue struct {
	subType int
	value   []byte
}

type portIDValue struct {
	subType int
	value   string
}
type capabilitiesValue struct {
	system  int
	enabled int
}

type managementAddressValue struct {
	addressStringLength int
	addressSubtype      int
	managementAddress   []byte
	interfaceSubType    int
	interfaceNumber     int
	oidStringLength     int
}

var (
	txTTR *time.Ticker

	txMacAddr             string
	txIPAddr              string
	msgTxHold             uint64
	msgTxInterval         uint64
	reinitDelay           uint64
	txDelay               uint64
	adminStatus           uint8
	somethingChangedLocal bool
	txTTL                 uint64
	txDelayWhile          uint64
	portNum               string
	ttlValue              uint16
	portDsr               string
	systemName            string
	systemDsr             string
	etherType             = []byte{0x88, 0xcc}
	lldpDA1               = []byte{0x01, 0x80, 0xc2, 0x00, 0x00, 0x00}
	lldpDA2               = []byte{0x01, 0x80, 0xc2, 0x00, 0x00, 0x03}
	lldpDA3               = []byte{0x01, 0x80, 0xc2, 0x00, 0x00, 0x0e}
	// HeaderContent dfsdfsdf
	HeaderContent  = make([]byte, 0, 14)
	tlvContent     = make(map[string]tlv)
	tlvByteContent = make(map[string][]byte)
)

func txinit() {

	msgTxHold = 4      // default 4
	reinitDelay = 2    // default 2 seconds
	msgTxInterval = 30 // default 30 seconds
	txDelay = 2        // default 2 seconds
	txMacAddr = getMyMacAddr()
	txIPAddr = "192.168.16.2"
	portNum = "15"
	ttlValue = 15
	portDsr = "Port 15"
	systemName = "Calvin"
	systemDsr = "Calvin's virtualBox"
	fmt.Println("================================================")
	fmt.Println("tx_Init complete.")
	fmt.Println("================================================")
	txfunc()

}

func txIdle() {

	txTTL = min(65535, (msgTxInterval * msgTxHold))
	txTTR = time.NewTicker(time.Second * time.Duration(msgTxInterval))
	somethingChangedLocal = false
	txDelayWhile = txDelay

}

func getMyMacAddr() (addr string) {

	interf, err := net.Interfaces()
	if err == nil {
		for _, i := range interf {
			if i.Flags&net.FlagUp != 0 && (bytes.Compare(i.HardwareAddr, nil) != 0) {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}

	return addr

}

func txfunc() {

	txfillTlv()
	TxBuildPktContent()
}

func txfillTlv() {

	chassisIDArray := []byte(txMacAddr)
	tlvContent["chassisID"] = tlv{1, 1 + len(chassisIDArray), chassisIDValue{4, chassisIDArray}}
	tlvContent["portID"] = tlv{2, len(portNum), portIDValue{7, portNum}}
	tlvContent["ttl"] = tlv{3, 2, ttlValue}
	tlvContent["portDescription"] = tlv{4, len(portDsr), portDsr}
	tlvContent["systemName"] = tlv{5, len(systemName), systemName}
	tlvContent["systemDescription"] = tlv{6, len(systemDsr), systemDsr}
	tlvContent["systemCapabilities"] = tlv{7, 4, capabilitiesValue{4, 4}}
	tlvContent["mgmtAddrV4"] = tlv{8, 12, managementAddressValue{5, 1, []byte(txIPAddr), 2, 0, 0}}

	tlvContent["end"] = tlv{0, 0, nil}
}

// TxBuildPktContent sdfdsfs
func TxBuildPktContent() {

	HeaderContent = HeaderContent[:0]
	fillheader(txMacAddr)
	headerlength := len(txMacAddr)
	datalenth := TxbuildByteContent()
	totallength := headerlength + datalenth

	buf := bytes.NewBuffer(make([]byte, 0, totallength))

	buf.Write(HeaderContent)
	for _, n := range tlvByteContent {
		buf.Write(n)
		//fmt.Println("BYTE content:\n", n)
	}
	//mt.Println("PKT content:\n", buf)
}

// TxbuildByteContent sdfdsfs
func TxbuildByteContent() (length int) {
	length = 0
	for i, v := range tlvContent {
		length += 2 + v.tlvLength
		buf := make([]byte, 0, 2+v.tlvLength)
		buf = append(buf, (byte)(v.tlvType<<1), (byte)(v.tlvLength))
		switch val := v.tlvValue.(type) {
		case chassisIDValue:
			buf = append(buf, (byte)(val.subType))
			buf = append(buf, val.value...)
		case portIDValue:
			buf = append(buf, (byte)(val.subType))
			buf = append(buf, val.value...)
		case uint16:
			buf = append(buf, (byte)(val))
		case string:
			buf = append(buf, []byte(val)...)
		case capabilitiesValue:
			buf = append(buf, (byte)(val.system))
			buf = append(buf, (byte)(val.enabled))
		case managementAddressValue:
			buf = append(buf, (byte)(val.addressStringLength), (byte)(val.addressSubtype))
			buf = append(buf, val.managementAddress...)
			buf = append(buf, (byte)(val.interfaceSubType))
			buf = append(buf, (byte)((uint32)(val.interfaceNumber)))
			buf = append(buf, (byte)(val.oidStringLength))
		case nil:

		default:
			fmt.Println("Value type error")
		}
		tlvByteContent[i] = buf
		//fmt.Println(tlvByteContent[i])
	}
	return length
}

func fillheader(txAddr string) {

	HeaderContent = append(HeaderContent, lldpDA1...)
	HeaderContent = append(HeaderContent, []byte(txAddr)...)
	HeaderContent = append(etherType)
}
func min(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

/*func main() {

	txinit()
	_, err := net.Dial("lldp", "192.168.16.1")
	if err != nil {
		fmt.Println(err)
	}
}*/
