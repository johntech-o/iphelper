package iphelper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	HEADER_LENGTH   = 8
	BODYLINE_LENGTH = 20
)

// 获取ip段信息
type IpRow struct {
	Start    uint32
	End      uint32
	Country  uint16
	Province uint16
	City     uint16
	Zone     uint16
	Provider uint16
	Idc      uint16
}

type IpStore struct {
	bodyLength   uint32
	metaLength   uint32
	headerBuffer []byte
	bodyBuffer   []byte
	metaBuffer   []byte
	IpTable      []IpRow // ip信息表 按范围自增
	metaTable    map[string][]string
}

func NewIpStore(filename string) *IpStore {
	store := IpStore{headerBuffer: make([]byte, HEADER_LENGTH), metaTable: make(map[string][]string)}
	store.parseStore(filename)
	return &store
}

// 获取ip的区域编码
func (this *IpStore) GetIpAreacode(ipSearch string) (uint64, error) {
	row, err := this.getIpRangeInfo(ipSearch)
	if err != nil {
		return 0, err
	}
	areacode := this.getLocationCodeByRow(row)
	codeUint64, err := strconv.ParseUint(areacode, 10, 64)
	if err != nil {
		return 0, err
	}
	return codeUint64, nil

}

// 获取ip的位置信息
func (this *IpStore) GetIpLocation(ipSearch string) (location map[string]string, err error) {
	row, err := this.getIpRangeInfo(ipSearch)
	if err != nil {
		return location, err
	}
	location, err = this.parseIpLocation(row)
	return location, err
}

// 获取ip的区域信息列表
func (this *IpStore) GetMetaTable() map[string][]string {
	return this.metaTable
}

func (this *IpStore) GetAreacodeLocation(areacode uint64) map[string]string {
	result := map[string]string{}
	result["idc"] = this.metaTable["idc"][areacode%100]
	areacode /= 100
	result["[rovider"] = this.metaTable["provider"][areacode%100]
	areacode /= 100
	result["zone"] = this.metaTable["zone"][areacode%10000]
	areacode /= 10000
	result["city"] = this.metaTable["city"][areacode%10000]
	areacode /= 10000
	result["province"] = this.metaTable["province"][areacode%10000]
	areacode /= 10000
	result["country"] = this.metaTable["country"][areacode%10000]
	return result
}

// 获取ip所在ip段的信息
func (this *IpStore) getIpRangeInfo(ipSearch string) (row IpRow, err error) {
	search := uint32(IP2Num(ipSearch))
	// fmt.Println(search)
	var start uint32 = 0
	var end uint32 = uint32(len(this.IpTable) - 1)
	var offset uint32 = 0
	for start <= end {
		mid := uint32(math.Floor(float64((end - start) / 2)))
		offset = start + mid
		IpRow := this.IpTable[offset]
		// fmt.Println(IpRow)
		if search >= IpRow.Start {
			if search <= IpRow.End {
				return IpRow, nil
			} else {
				start = offset + 1
				continue
			}
		} else {
			end = offset - 1
			continue
		}
	}
	return row, errors.New("fail to find")
}

func (this *IpStore) parseStore(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic("error opening file: %v\n" + err.Error())
	}
	if _, err := file.Read(this.headerBuffer); err != nil {
		panic("error read header" + err.Error())
	}
	if err := this.parseHeader(); err != nil {
		panic("parse header error:" + err.Error())
	}
	this.bodyBuffer = make([]byte, this.bodyLength)
	if _, err := file.ReadAt(this.bodyBuffer, HEADER_LENGTH); err != nil {
		panic("read body error")
	}
	this.metaBuffer = make([]byte, this.metaLength)
	if _, err := file.ReadAt(this.metaBuffer, int64(HEADER_LENGTH+this.bodyLength)); err != nil {
		panic("read meta error")
	}
	if err := this.paseBody(); err != nil {
		panic("parse body  failed")
	}
	if err := this.parseMeta(); err != nil {
		panic("pase meta failed")
	}
}

func (this *IpStore) parseIpLocation(row IpRow) (map[string]string, error) {
	location := make(map[string]string)
	location["country"] = this.metaTable["country"][row.Country]
	location["province"] = this.metaTable["province"][row.Province]
	location["city"] = this.metaTable["city"][row.City]
	location["zone"] = this.metaTable["zone"][row.Zone]
	location["provider"] = this.metaTable["provider"][row.Provider]
	location["idc"] = this.metaTable["idc"][row.Idc]
	location["areacode"] = this.getLocationCodeByRow(row)
	return location, nil

}

func (this *IpStore) getLocationCodeByRow(row IpRow) string {
	countryCode := strconv.Itoa(int(row.Country))
	provinceCode := fmt.Sprintf("%04d", row.Province)
	cityCode := fmt.Sprintf("%04d", row.City)
	zoneCode := fmt.Sprintf("%04d", row.Zone)
	provoderCode := fmt.Sprintf("%02d", row.Provider)
	idcCode := fmt.Sprintf("%02d", row.Idc)
	return countryCode + provinceCode + cityCode + zoneCode + provoderCode + idcCode

}

func (this *IpStore) paseBody() error {
	buf := bytes.NewBuffer(this.bodyBuffer)
	var offset uint32 = 0
	for offset < this.bodyLength {
		line := buf.Next(BODYLINE_LENGTH)
		row, err := this.parseBodyLine(line)
		if err != nil {
			return err
		}
		this.IpTable = append(this.IpTable, row)
		offset += BODYLINE_LENGTH
	}
	return nil
}

func (this *IpStore) parseMeta() (err error) {
	var countryLength, provinceLength, cityLength, zoneLength, providerLength, idcLength uint32 = 0, 0, 0, 0, 0, 0
	var offset uint32 = 4
	buf := bytes.NewBuffer(this.metaBuffer[0:offset])
	if err = binary.Read(buf, binary.BigEndian, &countryLength); err != nil {
		return err
	}
	countryMeta := this.metaBuffer[offset : offset+countryLength]
	this.metaTable["country"] = strings.Split(string(countryMeta), "|")

	offset = 4 + countryLength
	buf = bytes.NewBuffer(this.metaBuffer[offset : offset+4])
	if err = binary.Read(buf, binary.BigEndian, &provinceLength); err != nil {
		return err
	}
	offset += 4
	provinceMeta := this.metaBuffer[offset : offset+provinceLength]
	this.metaTable["province"] = strings.Split(string(provinceMeta), "|")

	offset += provinceLength
	buf = bytes.NewBuffer(this.metaBuffer[offset : offset+4])
	if err = binary.Read(buf, binary.BigEndian, &cityLength); err != nil {
		return err
	}
	offset += 4
	cityMeta := this.metaBuffer[offset : offset+cityLength]
	this.metaTable["city"] = strings.Split(string(cityMeta), "|")

	offset += cityLength
	buf = bytes.NewBuffer(this.metaBuffer[offset : offset+4])
	if err = binary.Read(buf, binary.BigEndian, &zoneLength); err != nil {
		return err
	}
	offset += 4
	zoneMeta := this.metaBuffer[offset : offset+zoneLength]
	this.metaTable["zone"] = strings.Split(string(zoneMeta), "|")

	offset += zoneLength
	buf = bytes.NewBuffer(this.metaBuffer[offset : offset+4])
	if err = binary.Read(buf, binary.BigEndian, &providerLength); err != nil {
		return err
	}
	offset += 4
	providerMeta := this.metaBuffer[offset : offset+providerLength]
	this.metaTable["provider"] = strings.Split(string(providerMeta), "|")

	offset += providerLength
	buf = bytes.NewBuffer(this.metaBuffer[offset : offset+4])
	if err = binary.Read(buf, binary.BigEndian, &idcLength); err != nil {
		return err
	}
	offset += 4
	idcMeta := this.metaBuffer[offset : offset+idcLength]
	this.metaTable["idc"] = strings.Split(string(idcMeta), "|")
	return nil
}

func (this *IpStore) parseBodyLine(buffer []byte) (row IpRow, err error) {
	buf := bytes.NewBuffer(buffer[0:4])
	if err = binary.Read(buf, binary.BigEndian, &row.Start); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[4:8])
	if err = binary.Read(buf, binary.BigEndian, &row.End); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[8:10])
	if err = binary.Read(buf, binary.BigEndian, &row.Country); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[10:12])
	if err = binary.Read(buf, binary.BigEndian, &row.Province); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[12:14])
	if err = binary.Read(buf, binary.BigEndian, &row.City); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[14:16])
	if err = binary.Read(buf, binary.BigEndian, &row.Zone); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[16:18])
	if err = binary.Read(buf, binary.BigEndian, &row.Provider); err != nil {
		goto fail
	}
	buf = bytes.NewBuffer(buffer[18:20])
	if err = binary.Read(buf, binary.BigEndian, &row.Idc); err != nil {
		goto fail
	}
	// fmt.Println(row)
	return row, err

fail:
	return row, err

}

func (this *IpStore) parseHeader() error {
	buf := bytes.NewBuffer(this.headerBuffer[0:4])
	if err := binary.Read(buf, binary.BigEndian, &this.bodyLength); err != nil {
		return err
	}
	buf = bytes.NewBuffer(this.headerBuffer[4:])
	if err := binary.Read(buf, binary.BigEndian, &this.metaLength); err != nil {
		return err
	}
	return nil
}

func IP2Num(requestip string) uint64 {
	//获取客户端地址的long
	nowip := strings.Split(requestip, ".")
	if len(nowip) != 4 {
		return 0
	}
	a, _ := strconv.ParseUint(nowip[0], 10, 64)
	b, _ := strconv.ParseUint(nowip[1], 10, 64)
	c, _ := strconv.ParseUint(nowip[2], 10, 64)
	d, _ := strconv.ParseUint(nowip[3], 10, 64)
	ipNum := a<<24 | b<<16 | c<<8 | d
	return ipNum
}

func Num2IP(ipnum uint64) string {
	byte1 := ipnum & 0xff
	byte2 := (ipnum & 0xff00)
	byte2 >>= 8
	byte3 := (ipnum & 0xff0000)
	byte3 >>= 16
	byte4 := (ipnum & 0xff000000)
	byte4 >>= 24
	result := strconv.FormatUint(byte4, 10) + "." +
		strconv.FormatUint(byte3, 10) + "." +
		strconv.FormatUint(byte2, 10) + "." +
		strconv.FormatUint(byte1, 10)
	return result
}
