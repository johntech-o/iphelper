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
	// get all the area info of the ip databases
	table := store.GetMetaTable()
	for typ, areas := range table {
		fmt.Println(typ, len(areas), areas)
		fmt.Println("--------------------")
	}

	// get geo info of ip address
	geo, e := store.GetIpGeo("43.240.79.255")
	fmt.Println(geo, e)

	//  get location areacode of ip address
	code, e := store.GetIpAreacode("43.240.79.255")
	fmt.Println(code, e)

	// get the location info of areacode
	// u can save the areacode to user`s session
	// get the location info by areacode is more fast than by ip address
	codeGeo := store.GetAreacodeGeo(code)
	fmt.Println(codeGeo)
	for typ, area := range codeGeo {
		if geo[typ] != area {
			t.Error("meta data and ip store not match")
		}
		t.Log(typ, area)
	}

}
