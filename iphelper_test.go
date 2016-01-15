package iphelper

import (
	"fmt"
	"testing"
)

func TestCreate(t *testing.T) {
	err := MakeDat("ipquery.txt", "ip.dat")
	if err != nil {
		t.Error(err)
	}
}

func TestQuery(t *testing.T) {

	// init use the ip.dat file
	var store = NewIpStore("ip.dat")

	// get all the geo info  of the ip databases
	table := store.GetMetaTable()
	for typ, areas := range table {
		fmt.Println(typ, len(areas), areas)
		fmt.Println("--------------------")
	}
	// 1697906688	1697972223	中国	北京市	北京市	朝阳区	未知	电信
	fmt.Println(Num2IP(1697906688), Num2IP(1697972223))

	// get geo info and areacode of ip address
	geo, e := store.GetGeoByIp("101.52.255.200")
	fmt.Println(geo, e)

	//  get geo code of ip address
	code, e := store.GetGeocodeByIp("101.52.255.200")
	fmt.Println(code, e)

	// get the geo info of areacode
	// u can save the areacode to user`s session
	// get the location info by areacode is more fast than by ip address
	codeGeo := store.GetGeoByGeocode(code)
	fmt.Println(codeGeo)
	for typ, area := range codeGeo {
		if geo[typ] != area {
			t.Error("meta data and ip store not match")
		}
		t.Log(typ, area)
	}

}
