# iphelper
A ip-database library  for golang. Find a information of an IP address including country,province,city,region,street and network provider.

u can find example in the iphelper_test.go 


// init use the ip.dat file

	var store = NewIpStore("ip.dat")


// get location info of ip address

	geo, e := store.GetIpGeo("43.240.79.255")
	
	fmt.Println(geo, e)


//  get location areacode of ip address

	code, e := store.GetIpAreacode("43.240.79.255")

//  output:map[country:中国 province:上海市 city:上海市 zone:未知 location:未知 operator:未知 areacode:20017009000000100] <nil>

	fmt.Println(code, e)

// get the location info of areacode
// u can save the areacode to user`s session
// get the location info by areacode is more fast than by ip address

	codeGeo := store.GetAreacodeGeo(code)
	
// 	output: 20017009000000100 <nil>

	fmt.Println(codeGeo)


// get all the area info of the ip databases
	
	table := store.GetMetaTable()
	
	for typ, areas := range table {
		fmt.Println(typ, len(areas), areas)
		fmt.Println("--------------------")
	}
