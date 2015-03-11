# iphelper
A ip-database library  for golang. Find a information of an IP address including country,province,city,region,street and network provider.

u can find example in the iphelper_test.go 


package iphelper

import (
	"fmt"
	"testing"
)

// init use the ip.dat file
var store = NewIpStore("ip.dat")


// get location info of ip address
l, e := store.GetIpLocation("43.240.79.255")
fmt.Println(l, e)


//  get location areacode of ip address
code, e := store.GetIpAreacode("43.240.79.255")
fmt.Println(code, e)

// get the location info of areacode
// u can save the areacode to user`s session
// get the location info by areacode is more fast than by ip address
l = store.GetAreacodeLocation(code)
fmt.Println(l)


// get all the area info of the ip databases
table := store.GetMetaTable()
fmt.Println(table["country"], table["province"], table["city"], table["zone"], table["provider"], table["idc"])

